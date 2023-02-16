package cosmos

import (
	"errors"
	"sync"

	"github.com/Mobile-Web3/backend/pkg/cosmos/connection"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var ErrSignModeUnknown = errors.New("unknown sign mode")

type chain struct {
	ID              string
	HttpClient      connection.HttpABCIClient
	GrpcClient      connection.GrpcConnection
	WebsocketClient connection.WebsocketClient
}

type Client struct {
	interfaceRegistry types.InterfaceRegistry
	codec             codec.Codec
	txConfig          client.TxConfig
	amino             *codec.LegacyAmino

	mutex  sync.RWMutex
	chains map[string]chain

	signMode signing.SignMode

	getRpcHandler  connection.GetRpcHandler
	txEventHandler connection.TxEventHandler
}

func NewClient(
	signMode string,
	txEventHandler connection.TxEventHandler,
	getRpcHandler connection.GetRpcHandler) (*Client, error) {
	codecData := makeCodec()

	mode := signing.SignMode_SIGN_MODE_UNSPECIFIED
	switch signMode {
	case "direct":
		mode = signing.SignMode_SIGN_MODE_DIRECT
	case "amino-json":
		mode = signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	default:
		return nil, ErrSignModeUnknown
	}

	return &Client{
		interfaceRegistry: codecData.InterfaceRegistry,
		codec:             codecData.Marshaler,
		txConfig:          codecData.TxConfig,
		amino:             codecData.Amino,

		mutex:  sync.RWMutex{},
		chains: make(map[string]chain),

		signMode: mode,

		txEventHandler: txEventHandler,
		getRpcHandler:  getRpcHandler,
	}, nil
}

func (c *Client) getChainData(chainID string) chain {
	c.mutex.RLock()
	chainData, ok := c.chains[chainID]
	if ok {
		c.mutex.RUnlock()
		return chainData
	}

	c.mutex.RUnlock()
	c.mutex.Lock()
	defer c.mutex.Unlock()

	chainData.ID = chainID
	chainData.HttpClient = connection.NewHttpABCIClient(chainID, c.getRpcHandler)
	chainData.GrpcClient = connection.NewGrpcConnectionFromRPC(chainData.HttpClient, c.interfaceRegistry, c.codec)
	chainData.WebsocketClient = connection.NewTendermintWebsocketClient(chainID, c.getRpcHandler, c.txEventHandler)
	c.chains[chainID] = chainData
	return chainData
}

func (c *Client) GetChainHttpClient(chainID string) connection.HttpABCIClient {
	chainData := c.getChainData(chainID)
	return chainData.HttpClient
}

func (c *Client) GetChainGrpcClient(chainID string) connection.GrpcConnection {
	chainData := c.getChainData(chainID)
	return chainData.GrpcClient
}

func (c *Client) GetChainWebsocketClient(chainID string) connection.WebsocketClient {
	chainData := c.getChainData(chainID)
	return chainData.WebsocketClient
}
