package connection

import (
	"context"
	"errors"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/client"
	"github.com/tendermint/tendermint/rpc/client/http"
	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

var (
	ErrCatchingUp     = errors.New("still catching up")
	ErrNoAvailableRPC = errors.New("no available rpc")
)

type TxEvent struct {
	TxHash    string
	Code      uint32
	Log       string
	Info      string
	GasUsed   int64
	GasWanted int64
}

type GetRPCEndpointsHandler func(ctx context.Context, chainID string) ([]string, error)
type TxEventHandler func(ctx context.Context, event TxEvent, params map[string]interface{}) error

func newNodeClient(endpoint string) (*http.HTTP, error) {
	return sdk.NewClientFromNode(endpoint)
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

func checkRpcClient(ctx context.Context, endpoint string) error {
	rpcClient, err := newRpcClient(endpoint, 5*time.Second)
	if err != nil {
		return err
	}

	result, err := rpcClient.Status(ctx)
	if err != nil {
		return err
	}

	if result.SyncInfo.CatchingUp {
		return ErrCatchingUp
	}

	return nil
}

type endpointStorage struct {
	mutex     sync.Mutex
	endpoints []string
}

func (s *endpointStorage) addEndpoint(endpoint string) {
	s.mutex.Lock()
	s.endpoints = append(s.endpoints, endpoint)
	s.mutex.Unlock()
}

func healthcheckRPC(ctx context.Context, endpoint string, storage *endpointStorage, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := checkRpcClient(ctx, endpoint); err != nil {
		return
	}
	storage.addEndpoint(endpoint)
}

func getHealthEndpoints(ctx context.Context, endpoints []string) ([]string, error) {
	wg := &sync.WaitGroup{}
	storage := &endpointStorage{}

	for _, endpoint := range endpoints {
		wg.Add(1)
		go healthcheckRPC(ctx, endpoint, storage, wg)
	}

	wg.Wait()

	if len(storage.endpoints) == 0 {
		return nil, ErrNoAvailableRPC
	}

	return storage.endpoints, nil
}
