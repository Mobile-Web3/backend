package connection

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/client"
	"github.com/tendermint/tendermint/libs/bytes"
	tendermint "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"github.com/tendermint/tendermint/types"
)

var ErrNoAvailableRPC = errors.New("no available rpc")

type HttpABCIClient interface {
	tendermint.ABCIClient
}

type GetRpcHandler func(ctx context.Context, chainID string) ([]string, error)

type httpABCIClient struct {
	chainID     string
	isInit      bool
	endpoint    string
	mutex       sync.RWMutex
	queryClient tendermint.ABCIClient
	getRpc      GetRpcHandler
}

func NewHttpABCIClient(chainID string, getRpcHandler GetRpcHandler) HttpABCIClient {
	return &httpABCIClient{
		chainID: chainID,
		getRpc:  getRpcHandler,
	}
}

func (c *httpABCIClient) init(ctx context.Context) (tendermint.ABCIClient, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	rpcEndpoints, err := c.getRpc(ctx, c.chainID)
	if err != nil {
		return nil, err
	}

	c.isInit = false
	endpoints, err := getHealthEndpoints(ctx, rpcEndpoints)
	if err != nil {
		return nil, err
	}

	var endpoint string
	n := len(endpoints) - 1
	switch n {
	case 0:
		endpoint = endpoints[0]
	default:
		endpoint = endpoints[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(n)]
	}

	rpcClient, err := sdk.NewClientFromNode(endpoint)
	if err != nil {
		err = fmt.Errorf("creating node client with endpoint: %s; %s", endpoint, err.Error())
		return nil, err
	}

	c.endpoint = endpoint
	c.queryClient = rpcClient
	c.isInit = true
	return c.queryClient, nil
}

func (c *httpABCIClient) invalidate() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.isInit = false
}

func (c *httpABCIClient) getActiveClient(ctx context.Context) (tendermint.ABCIClient, error) {
	c.mutex.RLock()
	if c.isInit {
		c.mutex.RUnlock()
		return c.queryClient, nil
	}

	c.mutex.RUnlock()
	return c.init(ctx)
}

func (c *httpABCIClient) ABCIInfo(ctx context.Context) (*ctypes.ResultABCIInfo, error) {
	queryClient, err := c.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := queryClient.ABCIInfo(ctx)
	if err != nil {
		c.invalidate()
		return nil, err
	}
	return result, nil
}
func (c *httpABCIClient) ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*ctypes.ResultABCIQuery, error) {
	queryClient, err := c.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := queryClient.ABCIQuery(ctx, path, data)
	if err != nil {
		c.invalidate()
		return nil, err
	}
	return result, nil
}
func (c *httpABCIClient) ABCIQueryWithOptions(ctx context.Context, path string, data bytes.HexBytes,
	opts tendermint.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	queryClient, err := c.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := queryClient.ABCIQueryWithOptions(ctx, path, data, opts)
	if err != nil {
		c.invalidate()
		return nil, err
	}
	return result, nil
}

func (c *httpABCIClient) BroadcastTxCommit(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	queryClient, err := c.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := queryClient.BroadcastTxCommit(ctx, tx)
	if err != nil {
		c.invalidate()
		return nil, err
	}
	return result, nil
}
func (c *httpABCIClient) BroadcastTxAsync(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	queryClient, err := c.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := queryClient.BroadcastTxAsync(ctx, tx)
	if err != nil {
		c.invalidate()
		return nil, err
	}
	return result, nil
}
func (c *httpABCIClient) BroadcastTxSync(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	queryClient, err := c.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := queryClient.BroadcastTxSync(ctx, tx)
	if err != nil {
		c.invalidate()
		return nil, err
	}
	return result, nil
}

func newRpcClient(endpoint string, timeout time.Duration) (*http.HTTP, error) {
	httpClient, err := client.DefaultHTTPClient(endpoint)
	if err != nil {
		return nil, err
	}

	httpClient.Timeout = timeout
	rpcClient, err := http.NewWithClient(endpoint, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}

	return rpcClient, nil
}

func checkRpcClient(ctx context.Context, endpoint string, output chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	rpcClient, err := newRpcClient(endpoint, 5*time.Second)
	if err != nil {
		return
	}

	result, err := rpcClient.Status(ctx)
	if err != nil {
		return
	}

	if result.SyncInfo.CatchingUp {
		return
	}

	output <- endpoint
}

func getHealthEndpoints(ctx context.Context, endpoints []string) ([]string, error) {
	wg := &sync.WaitGroup{}
	output := make(chan string)
	final := make(chan []string)

	go func() {
		var result []string
		for endpoint := range output {
			result = append(result, endpoint)
		}
		final <- result
	}()

	for _, endpoint := range endpoints {
		wg.Add(1)
		go checkRpcClient(ctx, endpoint, output, wg)
	}

	wg.Wait()
	close(output)
	result := <-final
	close(final)

	if len(result) == 0 {
		return nil, ErrNoAvailableRPC
	}

	return result, nil
}
