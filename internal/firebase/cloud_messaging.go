package firebase

import (
	"context"
	"fmt"

	fcm "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	cosmos "github.com/Mobile-Web3/backend/pkg/cosmos/client"
	"github.com/Mobile-Web3/backend/pkg/log"
	"google.golang.org/api/option"
)

type CloudMessagingClient struct {
	logger log.Logger
	client *messaging.Client
}

func NewCloudMessagingClient(keyPath string, logger log.Logger) (*CloudMessagingClient, error) {
	ctx := context.Background()
	opt := option.WithCredentialsFile(keyPath)
	app, err := fcm.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return &CloudMessagingClient{
		logger: logger,
		client: client,
	}, nil
}

func (c *CloudMessagingClient) SendTxResult(ctx context.Context, token string, event cosmos.TxEvent) error {
	message := &messaging.Message{
		Token: token,
		Data: map[string]string{
			"isSuccess": fmt.Sprintf("%t", event.Code == 0),
			"info":      event.Info,
			"gasUsed":   fmt.Sprintf("%d", event.GasUsed),
			"gasWanted": fmt.Sprintf("%d", event.GasWanted),
			"log":       event.Log,
		},
	}

	_, err := c.client.Send(ctx, message)
	if err != nil {
		err = fmt.Errorf("firebase cloud messaging; %s", err.Error())
		c.logger.Error(err)
		return err
	}

	return nil
}
