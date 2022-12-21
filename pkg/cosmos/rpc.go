package cosmos

import (
	"context"
	"errors"
	"time"

	tendermintrpc "github.com/tendermint/tendermint/rpc/client/http"
	tenderminthttp "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

var ErrCatchingUp = errors.New("still catching up")

func NewRPCClient(addr string, timeout time.Duration) (*tendermintrpc.HTTP, error) {
	httpClient, err := tenderminthttp.DefaultHTTPClient(addr)
	if err != nil {
		return nil, err
	}

	httpClient.Timeout = timeout
	rpcClient, err := tendermintrpc.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}

	return rpcClient, nil
}

func HealthcheckRPC(ctx context.Context, endpoint string) error {
	client, err := NewRPCClient(endpoint, 5*time.Second)
	if err != nil {
		return err
	}

	result, err := client.Status(ctx)
	if err != nil {
		return err
	}

	if result.SyncInfo.CatchingUp {
		return ErrCatchingUp
	}

	return nil
}
