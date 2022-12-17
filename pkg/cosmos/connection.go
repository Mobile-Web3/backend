package cosmos

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/gogoproto/grpc"
	lens "github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap"
)

type MsgData struct {
	Memo          string
	GasAdjustment string
	GasPrice      string
	Msg           sdk.Msg
}

type RPCConnection interface {
	grpc.ClientConn
	CalculateGas(ctx context.Context, msg sdk.Msg) (txtypes.SimulateResponse, uint64, error)
	SendMsg(ctx context.Context, msgData MsgData) (*sdk.TxResponse, error)
}

type LensClientRPCConnection struct {
	*lens.ChainClient
}

type rpcConfig struct {
	ChainID       string
	ChainPrefix   string
	RpcURL        string
	HomePath      string
	Mnemonic      string
	Slip44        uint32
	GasAdjustment float64
}

func newLensClient(logger *zap.Logger, config rpcConfig) (*LensClientRPCConnection, error) {
	if config.Mnemonic == "" {
		chainConfig := lens.ChainClientConfig{
			Key:            "default",
			KeyringBackend: "memory",
			RPCAddr:        config.RpcURL,
			AccountPrefix:  config.ChainPrefix,
			ChainID:        config.ChainID,
			Timeout:        "5s",
			GasAdjustment:  config.GasAdjustment,
		}

		chainClient, err := lens.NewChainClient(logger, &chainConfig, config.HomePath, os.Stdin, os.Stdout)
		if err != nil {
			return nil, err
		}

		return &LensClientRPCConnection{
			ChainClient: chainClient,
		}, nil
	}

	chainConfig := lens.ChainClientConfig{
		Key:            "default",
		KeyringBackend: "memory",
		RPCAddr:        config.RpcURL,
		AccountPrefix:  config.ChainPrefix,
		ChainID:        config.ChainID,
		Timeout:        "30s",
		OutputFormat:   "json",
		SignModeStr:    "direct",
		GasAdjustment:  config.GasAdjustment,
		Modules:        lens.ModuleBasics,
	}

	chainClient, err := lens.NewChainClient(logger, &chainConfig, config.HomePath, os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}

	_, err = chainClient.RestoreKey("source_key", config.Mnemonic, config.Slip44)
	if err != nil {
		return nil, err
	}

	chainConfig.Key = "source_key"

	return &LensClientRPCConnection{
		ChainClient: chainClient,
	}, nil
}

func (c *LensClientRPCConnection) CalculateGas(ctx context.Context, msg sdk.Msg) (txtypes.SimulateResponse, uint64, error) {
	txf, err := c.ChainClient.PrepareFactory(c.ChainClient.TxFactory())
	if err != nil {
		return txtypes.SimulateResponse{}, 0, err
	}

	response, adjusted, err := c.ChainClient.CalculateGas(ctx, txf, msg)
	if err != nil {
		return txtypes.SimulateResponse{}, 0, err
	}

	return response, adjusted, nil
}

func (c *LensClientRPCConnection) SendMsg(ctx context.Context, msgData MsgData) (*sdk.TxResponse, error) {
	txf, err := c.ChainClient.PrepareFactory(c.ChainClient.TxFactory())
	if err != nil {
		return nil, err
	}

	if msgData.Memo != "" {
		txf = txf.WithMemo(msgData.Memo)
	}

	adjusted, err := strconv.ParseUint(msgData.GasAdjustment, 0, 64)
	if err != nil {
		return nil, err
	}

	txf = txf.WithGas(adjusted)
	txf = txf.WithFees(msgData.GasPrice)

	txb, err := txf.BuildUnsignedTx(msgData.Msg)
	if err != nil {
		return nil, err
	}

	c.ChainClient.Codec.Marshaler.MustMarshalJSON(msgData.Msg)

	err = func() error {
		done := c.ChainClient.SetSDKContext()
		defer done()
		if err = tx.Sign(txf, c.ChainClient.Config.Key, txb, false); err != nil {
			return err
		}
		return nil
	}()

	if err != nil {
		return nil, err
	}

	txBytes, err := c.ChainClient.Codec.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		return nil, err
	}

	res, err := c.ChainClient.BroadcastTx(ctx, txBytes)
	if err != nil {
		return nil, err
	}

	if res.Code != 0 {
		return res, fmt.Errorf("transaction failed with code: %d", res.Code)
	}

	return res, nil
}
