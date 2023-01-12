package client

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var ErrSignModeUnknown = errors.New("unknown sign mode")

type GetRPCEndpointHandler func(ctx context.Context, chainID string) ([]string, error)

type Client struct {
	logger log.Logger

	interfaceRegistry types.InterfaceRegistry
	codec             codec.Codec
	txConfig          client.TxConfig
	amino             *codec.LegacyAmino

	rpcLifetime time.Duration
	chainMutex  sync.RWMutex
	chains      map[string]*chainState

	signMode signing.SignMode

	getRPCEndpointHandler GetRPCEndpointHandler
}

func NewClient(signMode string, rpcLifetime time.Duration, logger log.Logger, getRPCEndpointHandler GetRPCEndpointHandler) (*Client, error) {
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
		logger: logger,

		interfaceRegistry: codecData.InterfaceRegistry,
		codec:             codecData.Marshaler,
		txConfig:          codecData.TxConfig,
		amino:             codecData.Amino,

		rpcLifetime: rpcLifetime,
		chainMutex:  sync.RWMutex{},
		chains:      make(map[string]*chainState),
		signMode:    mode,

		getRPCEndpointHandler: getRPCEndpointHandler,
	}, nil
}
