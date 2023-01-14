package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/client"
	tendermint "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"github.com/tendermint/tendermint/types"
)

var (
	ErrCatchingUp        = errors.New("still catching up")
	ErrNilTxEventHandler = errors.New("tx event handler is nil")
)

func newNodeRPCClient(endpoint string) (*http.HTTP, error) {
	return sdk.NewClientFromNode(endpoint)
}

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

func (c *Client) getChainState(chainID string) *chainState {
	c.chainMutex.RLock()
	defer c.chainMutex.RUnlock()
	state := c.chains[chainID]
	return state
}

func (c *Client) initChainState(chainID string) *chainState {
	c.chainMutex.Lock()
	defer c.chainMutex.Unlock()
	state := newChainState(chainID, c.rpcLifetime, c.logger, c.getRPCEndpointHandler)
	c.chains[chainID] = state
	return state
}

func (c *Client) GetChainRPC(ctx context.Context, chainID string) (tendermint.Client, string, error) {
	state := c.getChainState(chainID)
	if state == nil {
		state = c.initChainState(chainID)
	}

	return state.GetActiveRPC(ctx)
}

func (c *Client) InvalidateChainClient(chainID string) {
	state := c.getChainState(chainID)
	if state != nil {
		state.invalidateRPC()
	}
}

type TxEvent struct {
	Code      uint32
	Log       string
	Info      string
	GasUsed   int64
	GasWanted int64
}

func (c *Client) listenTxEvents(ctx context.Context, token string, chainClient tendermint.Client, ch <-chan ctypes.ResultEvent) {
	defer chainClient.Stop()
	event, ok := <-ch
	if !ok {
		return
	}

	txData := event.Data.(types.EventDataTx)
	txEvent := TxEvent{
		Code:      txData.Result.Code,
		Log:       txData.Result.Log,
		Info:      txData.Result.Info,
		GasUsed:   txData.Result.GasUsed,
		GasWanted: txData.Result.GasWanted,
	}
	_ = c.txEventHandler(ctx, token, txEvent)
}

func (c *Client) SubscribeForTx(ctx context.Context, token string, chainID string, address string) error {
	if c.txEventHandler == nil {
		return ErrNilTxEventHandler
	}
	_, endpoint, err := c.GetChainRPC(ctx, chainID)
	if err != nil {
		return err
	}

	chainClient, err := newNodeRPCClient(endpoint)
	if err != nil {
		c.logger.Error(err)
		return err
	}

	if err = chainClient.Start(); err != nil {
		err = fmt.Errorf("websocket connection to %s is not available; %s", endpoint, err.Error())
		c.logger.Error(err)
		return err
	}

	subscriber := fmt.Sprintf("mobileweb3-%s-%s", chainID, address)
	query := fmt.Sprintf("tm.event = 'Tx' AND transfer.sender = '%s'", address)
	eventsChannel, err := chainClient.Subscribe(ctx, subscriber, query)
	if err != nil {
		err = fmt.Errorf("subscribing for tx events to %s; %s", endpoint, err.Error())
		c.logger.Error(err)
		return err
	}

	go c.listenTxEvents(ctx, token, chainClient, eventsChannel)
	return nil
}
