package connection

import (
	"context"
	"fmt"
	"sync"

	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/google/uuid"
	"github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

type WebsocketClient interface {
	SubscribeToTx(ctx context.Context, address string, params map[string]interface{}) error
}

type tendermintWebsocketClient struct {
	chainID       string
	logger        log.Logger
	isInit        bool
	endpoint      string
	mutex         sync.RWMutex
	client        *http.HTTP
	getRpc        GetRPCEndpointsHandler
	handleTxEvent TxEventHandler
}

func NewTendermintWebsocketClient(chainID string, logger log.Logger, getRpcHandler GetRPCEndpointsHandler, txEventHandler TxEventHandler) WebsocketClient {
	return &tendermintWebsocketClient{
		chainID:       chainID,
		logger:        logger,
		getRpc:        getRpcHandler,
		handleTxEvent: txEventHandler,
	}
}

func (c *tendermintWebsocketClient) invalidate() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.isInit = false
	_ = c.client.Stop()
}

func (c *tendermintWebsocketClient) init(ctx context.Context) (*http.HTTP, error) {
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

	var rpcEndpoint string
	var client *http.HTTP
	for _, endpoint := range endpoints {
		client, err = newNodeClient(endpoint)
		if err != nil {
			err = fmt.Errorf("creating node client with endpoint: %s; %s", endpoint, err.Error())
			c.logger.Error(err)
			return nil, err
		}

		if err = client.Start(); err != nil {
			continue
		}

		rpcEndpoint = endpoint
		break
	}
	if err != nil {
		return nil, ErrNoAvailableRPC
	}

	c.client = client
	c.endpoint = rpcEndpoint
	c.isInit = true
	return c.client, nil
}

func (c *tendermintWebsocketClient) getActiveClient(ctx context.Context) (*http.HTTP, error) {
	c.mutex.RLock()
	if c.isInit {
		c.mutex.RUnlock()
		return c.client, nil
	}

	c.mutex.RUnlock()
	return c.init(ctx)
}

type unsubscribeCallback func(ctx context.Context, subscriber, query string) error

func (c *tendermintWebsocketClient) listenTxEvents(ctx context.Context,
	subscriber string,
	query string,
	params map[string]interface{},
	channel <-chan ctypes.ResultEvent,
	unsubscribe unsubscribeCallback) {
	defer unsubscribe(ctx, subscriber, query)
	event, ok := <-channel
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

	_ = c.handleTxEvent(ctx, txEvent, params)
}

func (c *tendermintWebsocketClient) SubscribeToTx(ctx context.Context, address string, params map[string]interface{}) error {
	client, err := c.getActiveClient(ctx)
	if err != nil {
		return err
	}

	subscriber := fmt.Sprintf("%s-%s-%s", c.chainID, address, uuid.New().String())
	query := fmt.Sprintf("tm.event = 'Tx' AND transfer.sender = '%s'", address)
	eventsChannel, err := client.Subscribe(ctx, subscriber, query)
	if err != nil {
		c.invalidate()
		err = fmt.Errorf("subscribing for tx events to %s; %s", c.endpoint, err.Error())
		c.logger.Error(err)
		return err
	}

	go c.listenTxEvents(ctx, subscriber, query, params, eventsChannel, client.Unsubscribe)
	return nil
}
