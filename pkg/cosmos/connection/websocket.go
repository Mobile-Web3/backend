package connection

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/client"
	"github.com/google/uuid"
	"github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

type TxEvent struct {
	TxHash    string
	Code      uint32
	Log       string
	Info      string
	GasUsed   int64
	GasWanted int64
}

type TxEventHandler func(ctx context.Context, event TxEvent, params map[string]interface{}) error

type WebsocketClient struct {
	chainID       string
	getRpc        GetRpcHandler
	handleTxEvent TxEventHandler
	cancelChannel chan string
}

func NewWebsocketClient(chainID string, getRpcHandler GetRpcHandler, txEventHandler TxEventHandler) *WebsocketClient {
	return &WebsocketClient{
		chainID:       chainID,
		getRpc:        getRpcHandler,
		handleTxEvent: txEventHandler,
		cancelChannel: make(chan string),
	}
}

func (c *WebsocketClient) listenTxEvents(
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
			txHash := ""
			if len(event.Events["tx.hash"]) > 0 {
				txHash = event.Events["tx.hash"][0]
			}

			txEvent := TxEvent{
				TxHash:    txHash,
				Code:      txData.Result.Code,
				Log:       txData.Result.Log,
				Info:      txData.Result.Info,
				GasUsed:   txData.Result.GasUsed,
				GasWanted: txData.Result.GasWanted,
			}

			_ = c.handleTxEvent(ctx, txEvent, params)
			return
		}
	}
}

func (c *WebsocketClient) subscribeToTx(
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
		return "", err
	}

	go c.listenTxEvents(subscriber, query, params, eventsChannel, client)
	return subscriber, nil
}

func (c *WebsocketClient) SubscribeToTx(ctx context.Context, address string, params map[string]interface{}) (string, error) {
	endpoints, err := c.getRpc(ctx, c.chainID)
	if err != nil {
		return "", err
	}

	var client *http.HTTP
	for _, endpoint := range endpoints {
		client, err = sdk.NewClientFromNode(endpoint)
		if err != nil {
			continue
		}

		if err = client.Start(); err != nil {
			continue
		}

		return c.subscribeToTx(ctx, address, endpoint, params, client)
	}

	err = fmt.Errorf("tendermint websocket connecting; %s", err.Error())
	return "", err
}

func (c *WebsocketClient) UnsubscribeFromTx(subscriber string) {
	c.cancelChannel <- subscriber
}
