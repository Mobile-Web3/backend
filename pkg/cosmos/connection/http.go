package connection

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/tendermint/tendermint/libs/bytes"
	tendermint "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

type HttpABCIClient interface {
	tendermint.ABCIClient
}

type httpABCIClient struct {
	chainID     string
	logger      log.Logger
	isInit      bool
	endpoint    string
	mutex       sync.RWMutex
	queryClient tendermint.ABCIClient
	getRpc      GetRPCEndpointsHandler
}

func NewHttpABCIClient(chainID string, logger log.Logger, getRpcHandler GetRPCEndpointsHandler) HttpABCIClient {
	return &httpABCIClient{
		chainID: chainID,
		logger:  logger,
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

	rpcClient, err := newNodeClient(endpoint)
	if err != nil {
		err = fmt.Errorf("creating node client with endpoint: %s; %s", endpoint, err.Error())
		c.logger.Error(err)
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
