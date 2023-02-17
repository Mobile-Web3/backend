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
	"github.com/tendermint/tendermint/types"
)

var ErrNoAvailableRPC = errors.New("no available rpc")

type GetRpcHandler func(ctx context.Context, chainID string) ([]string, error)

type HttpClient struct {
	chainID     string
	isInit      bool
	endpoint    string
	mutex       sync.RWMutex
	queryClient tendermint.ABCIClient
	getRpc      GetRpcHandler
}

func NewHttpClient(chainID string, getRpcHandler GetRpcHandler) *HttpClient {
	return &HttpClient{
		chainID: chainID,
		getRpc:  getRpcHandler,
	}
}

func checkRpcClient(ctx context.Context, endpoint string, output chan<- *http.HTTP, wg *sync.WaitGroup) {
	defer wg.Done()
	rpcClient, err := sdk.NewClientFromNode(endpoint)
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

	output <- rpcClient
}

func getHealthEndpoint(ctx context.Context, endpoints []string) (*http.HTTP, error) {
	wg := &sync.WaitGroup{}
	output := make(chan *http.HTTP)
	final := make(chan []*http.HTTP)

	go func() {
		var result []*http.HTTP
		for rpcClient := range output {
			result = append(result, rpcClient)
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

	var rpcClient *http.HTTP
	n := len(result) - 1
	switch n {
	case 0:
		rpcClient = result[0]
	default:
		rpcClient = result[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(n)]
	}

	return rpcClient, nil
}

func (c *HttpClient) init(ctx context.Context) (tendermint.ABCIClient, string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	rpcEndpoints, err := c.getRpc(ctx, c.chainID)
	if err != nil {
		return nil, "", err
	}

	c.isInit = false
	rpcClient, err := getHealthEndpoint(ctx, rpcEndpoints)
	if err != nil {
		return nil, "", err
	}

	c.endpoint = rpcClient.Remote()
	c.queryClient = rpcClient
	c.isInit = true
	return c.queryClient, c.endpoint, nil
}

func (c *HttpClient) invalidate() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.isInit = false
}

func (c *HttpClient) getActiveClient(ctx context.Context) (tendermint.ABCIClient, string, error) {
	c.mutex.RLock()
	if !c.isInit {
		c.mutex.RUnlock()
		return c.init(ctx)
	}
	defer c.mutex.RUnlock()
	return c.queryClient, c.endpoint, nil
}

func (c *HttpClient) ABCIInfo(ctx context.Context) (*ctypes.ResultABCIInfo, error) {
	queryClient, endpoint, err := c.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := queryClient.ABCIInfo(ctx)
	if err != nil {
		c.invalidate()
		return nil, fmt.Errorf("error while ABCIInfo request with endpoint %s; %s", endpoint, err.Error())
	}
	return result, nil
}
func (c *HttpClient) ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*ctypes.ResultABCIQuery, error) {
	queryClient, endpoint, err := c.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := queryClient.ABCIQuery(ctx, path, data)
	if err != nil {
		c.invalidate()
		return nil, fmt.Errorf("error while ABCIQuery request with endpoint %s; %s", endpoint, err.Error())
	}
	return result, nil
}
func (c *HttpClient) ABCIQueryWithOptions(ctx context.Context, path string, data bytes.HexBytes,
	opts tendermint.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	queryClient, endpoint, err := c.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := queryClient.ABCIQueryWithOptions(ctx, path, data, opts)
	if err != nil {
		c.invalidate()
		return nil, fmt.Errorf("error while ABCIQueryWithOptions request with endpoint %s; %s", endpoint, err.Error())
	}
	return result, nil
}

func (c *HttpClient) BroadcastTxCommit(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	queryClient, endpoint, err := c.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := queryClient.BroadcastTxCommit(ctx, tx)
	if err != nil {
		c.invalidate()
		return nil, fmt.Errorf("error while BroadcastTxCommit request with endpoint %s; %s", endpoint, err.Error())
	}
	return result, nil
}
func (c *HttpClient) BroadcastTxAsync(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	queryClient, endpoint, err := c.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := queryClient.BroadcastTxAsync(ctx, tx)
	if err != nil {
		c.invalidate()
		return nil, fmt.Errorf("error while BroadcastTxAsync request with endpoint %s; %s", endpoint, err.Error())
	}
	return result, nil
}
func (c *HttpClient) BroadcastTxSync(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	queryClient, endpoint, err := c.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	result, err := queryClient.BroadcastTxSync(ctx, tx)
	if err != nil {
		c.invalidate()
		return nil, fmt.Errorf("error while BroadcastTxSync request with endpoint %s; %s", endpoint, err.Error())
	}
	return result, nil
}
