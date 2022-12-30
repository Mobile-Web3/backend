package client

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Mobile-Web3/backend/pkg/cosmos/chain"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var ErrSignModeUnknown = errors.New("unknown sign mode")

type GetRPCEndpointHandler func(ctx context.Context, chainID string) ([]string, error)

type Client struct {
	interfaceRegistry types.InterfaceRegistry
	codec             codec.Codec
	txConfig          client.TxConfig
	amino             *codec.LegacyAmino

	rpcLifetime time.Duration
	chainMutex  sync.RWMutex
	chains      map[string]*chain.State

	keyName  string
	signMode signing.SignMode

	getRPCEndpointHandler GetRPCEndpointHandler
}

func NewClient(signMode string, rpcLifetime time.Duration, getRPCEndpointHandler GetRPCEndpointHandler) (*Client, error) {
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

		rpcLifetime: rpcLifetime,
		chainMutex:  sync.RWMutex{},
		chains:      make(map[string]*chain.State),
		keyName:     "source_key",
		signMode:    mode,

		getRPCEndpointHandler: getRPCEndpointHandler,
	}, nil
}
