package chain

import (
	"context"
	"errors"
	"time"

	"github.com/tendermint/tendermint/rpc/client/http"
	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

var ErrCatchingUp = errors.New("still catching up")

func newRPCClient(addr string, timeout time.Duration) (*http.HTTP, error) {
	httpClient, err := client.DefaultHTTPClient(addr)
	if err != nil {
		return nil, err
	}

	httpClient.Timeout = timeout
	rpcClient, err := http.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}

	return rpcClient, nil
}

func healthcheckRPC(ctx context.Context, endpoint string) error {
	rpcClient, err := newRPCClient(endpoint, 5*time.Second)
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
