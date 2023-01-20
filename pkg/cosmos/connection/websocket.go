package connection

import (
	"context"
	"fmt"

	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/google/uuid"
	"github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

type WebsocketClient interface {
	SubscribeToTx(ctx context.Context, address string, params map[string]interface{}) (string, error)
	UnsubscribeFromTx(subscriber string)
}

type tendermintWebsocketClient struct {
	chainID       string
	logger        log.Logger
	getRpc        GetRPCEndpointsHandler
	handleTxEvent TxEventHandler
	cancelChannel chan string
}

func NewTendermintWebsocketClient(chainID string, logger log.Logger, getRpcHandler GetRPCEndpointsHandler, txEventHandler TxEventHandler) WebsocketClient {
	return &tendermintWebsocketClient{
		chainID:       chainID,
		logger:        logger,
		getRpc:        getRpcHandler,
		handleTxEvent: txEventHandler,
		cancelChannel: make(chan string),
	}
}

func (c *tendermintWebsocketClient) listenTxEvents(
	subscriber string,
	query string,
	params map[string]interface{},
	channel <-chan ctypes.ResultEvent,
	client *http.HTTP) {
	ctx := context.Background()
	defer client.Unsubscribe(ctx, subscriber, query)
	defer client.Stop()
	for {
		select {
		case sub := <-c.cancelChannel:
			if sub == subscriber {
				return
			}
		case event, ok := <-channel:
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
	}
}

func (c *tendermintWebsocketClient) subscribeToTx(
	ctx context.Context,
	address string,
	endpoint string,
	params map[string]interface{},
	client *http.HTTP) (string, error) {
	subscriber := fmt.Sprintf("%s-%s-%s", c.chainID, address, uuid.New().String())
	query := fmt.Sprintf("tm.event = 'Tx' AND transfer.sender = '%s'", address)
	eventsChannel, err := client.Subscribe(ctx, subscriber, query)
	if err != nil {
		err = fmt.Errorf("subscribing for tx events to %s; %s", endpoint, err.Error())
		c.logger.Error(err)
		return "", err
	}

	go c.listenTxEvents(subscriber, query, params, eventsChannel, client)
	return subscriber, nil
}

func (c *tendermintWebsocketClient) SubscribeToTx(ctx context.Context, address string, params map[string]interface{}) (string, error) {
	endpoints, err := c.getRpc(ctx, c.chainID)
	if err != nil {
		return "", err
	}

	var client *http.HTTP
	for _, endpoint := range endpoints {
		client, err = newNodeClient(endpoint)
		if err != nil {
			continue
		}

		if err = client.Start(); err != nil {
			continue
		}

		return c.subscribeToTx(ctx, address, endpoint, params, client)
	}

	err = fmt.Errorf("tendermint websocket connecting; %s", err.Error())
	c.logger.Error(err)
	return "", err
}

func (c *tendermintWebsocketClient) UnsubscribeFromTx(subscriber string) {
	c.cancelChannel <- subscriber
}
