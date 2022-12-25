package cosmos

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/Mobile-Web3/backend/internal/chain"
	"github.com/Mobile-Web3/backend/pkg/cosmos"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	lens "github.com/strangelove-ventures/lens/client"
	abci "github.com/tendermint/tendermint/abci/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"go.uber.org/zap"
)

type LensCosmosClient struct {
	errorLogger *log.Logger
	client      *lens.ChainClient
}

type rpcConfig struct {
	ChainID     string
	ChainPrefix string
	RpcURL      string
	HomePath    string
	Key         string
	CoinType    uint32
}

func newLensClient(errorLogger *log.Logger, zapLogger *zap.Logger, config rpcConfig) (*LensCosmosClient, error) {
	if config.Key == "" {
		chainConfig := lens.ChainClientConfig{
			Key:            "default",
			KeyringBackend: "memory",
			RPCAddr:        config.RpcURL,
			AccountPrefix:  config.ChainPrefix,
			ChainID:        config.ChainID,
			Timeout:        "5s",
		}

		chainClient, err := lens.NewChainClient(zapLogger, &chainConfig, config.HomePath, os.Stdin, os.Stdout)
		if err != nil {
			errorLogger.Println(err)
			return nil, err
		}

		return &LensCosmosClient{
			errorLogger: errorLogger,
			client:      chainClient,
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
		Modules:        lens.ModuleBasics,
	}

	chainClient, err := lens.NewChainClient(zapLogger, &chainConfig, config.HomePath, os.Stdin, os.Stdout)
	if err != nil {
		errorLogger.Println(err)
		return nil, err
	}

	_, err = chainClient.RestoreKey("source_key", config.Key, config.CoinType)
	if err != nil {
		errorLogger.Println(err)
		return nil, err
	}

	chainConfig.Key = "source_key"

	return &LensCosmosClient{
		errorLogger: errorLogger,
		client:      chainClient,
	}, nil
}

func (c *LensCosmosClient) GetBalance(ctx context.Context, address string) (chain.Balance, error) {
	bankClient := bank.NewQueryClient(c.client)
	bankResponse, err := bankClient.AllBalances(ctx, &bank.QueryAllBalancesRequest{
		Address: address,
	})
	if err != nil {
		c.errorLogger.Println(err)
		return chain.Balance{}, err
	}

	stackingClient := staking.NewQueryClient(c.client)
	stakingResponse, err := stackingClient.DelegatorDelegations(ctx, &staking.QueryDelegatorDelegationsRequest{
		DelegatorAddr: address,
	})
	if err != nil {
		c.errorLogger.Println(err)
		return chain.Balance{}, err
	}

	response := chain.Balance{
		AvailableAmount: big.NewInt(0),
		StakedAmount:    big.NewInt(0),
	}
	if len(bankResponse.Balances) > 0 {
		response.AvailableAmount = bankResponse.Balances[0].Amount.BigInt()
	}
	if len(stakingResponse.DelegationResponses) > 0 {
		response.StakedAmount = stakingResponse.DelegationResponses[0].Balance.Amount.BigInt()
	}

	return response, nil
}

func (c *LensCosmosClient) SendTransaction(ctx context.Context, txData chain.SendTxData) (chain.SendTxResponse, error) {
	coins, err := sdk.ParseCoinNormalized(txData.Amount)
	if err != nil {
		c.errorLogger.Println(err)
		return chain.SendTxResponse{}, err
	}

	msgSend := &bank.MsgSend{
		FromAddress: txData.From,
		ToAddress:   txData.To,
		Amount:      sdk.Coins{coins},
	}

	txf, err := c.client.PrepareFactory(c.createFactory())
	if err != nil {
		c.errorLogger.Println(err)
		return chain.SendTxResponse{}, err
	}

	if txData.Memo != "" {
		txf = txf.WithMemo(txData.Memo)
	}

	adjusted, err := strconv.ParseUint(txData.GasAdjusted, 0, 64)
	if err != nil {
		c.errorLogger.Println(err)
		return chain.SendTxResponse{}, err
	}

	txf = txf.WithGas(adjusted)
	txf = txf.WithFees(txData.GasPrice)

	txb, err := txf.BuildUnsignedTx(msgSend)
	if err != nil {
		c.errorLogger.Println(err)
		return chain.SendTxResponse{}, err
	}

	c.client.Codec.Marshaler.MustMarshalJSON(msgSend)

	if err = tx.Sign(txf, c.client.Config.Key, txb, false); err != nil {
		c.errorLogger.Println(err)
		return chain.SendTxResponse{}, err
	}

	txBytes, err := c.client.Codec.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		c.errorLogger.Println(err)
		return chain.SendTxResponse{}, err
	}

	response, err := c.client.BroadcastTx(ctx, txBytes)
	if err != nil {
		c.errorLogger.Println(err)
		return chain.SendTxResponse{}, err
	}

	if response.Code != 0 {
		err = fmt.Errorf("transaction failed with code: %d; TxHash: %s", response.Code, response.TxHash)
		c.errorLogger.Println(err)
		return chain.SendTxResponse{}, err
	}

	return chain.SendTxResponse{
		Height:    response.Height,
		TxHash:    response.TxHash,
		Data:      response.Data,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
		RawLog:    response.RawLog,
	}, nil
}

func (c *LensCosmosClient) SimulateTransaction(ctx context.Context, txData chain.SimulateTxData) (chain.SimulateTxResult, error) {
	coins, err := sdk.ParseCoinNormalized(txData.Amount)
	if err != nil {
		c.errorLogger.Println(err)
		return chain.SimulateTxResult{}, err
	}

	msgSend := &bank.MsgSend{
		FromAddress: txData.From,
		ToAddress:   txData.To,
		Amount:      sdk.Coins{coins},
	}

	rpcClient, err := client.NewClientFromNode(c.client.Config.RPCAddr)
	if err != nil {
		c.errorLogger.Println(err)
		return chain.SimulateTxResult{}, err
	}

	txf := cosmos.CreateTxFactory(cosmos.CreateFactoryContext{
		ChainID:          c.client.Config.ChainID,
		TxConfig:         c.client.Codec.TxConfig,
		AccountRetriever: c.client,
		KeyBase:          c.client.Keybase,
		SignMode:         c.client.Config.SignMode(),
	})

	if txf, err = cosmos.PrepareTxFactory(cosmos.PrepareFactoryContext{
		Key:               c.client.Config.Key,
		KeyBase:           c.client.Keybase,
		RPCClient:         rpcClient,
		Factory:           txf,
		InterfaceRegistry: c.client.Codec.InterfaceRegistry,
		Codec:             c.client.Codec.Marshaler,
	}); err != nil {
		c.errorLogger.Println(err)
		return chain.SimulateTxResult{}, err
	}

	keyRecord, err := c.client.Keybase.Key(c.client.Config.Key)
	if err != nil {
		c.errorLogger.Println(err)
		return chain.SimulateTxResult{}, err
	}

	txBytes, err := cosmos.BuildTransaction(txf, keyRecord, msgSend)
	if err != nil {
		c.errorLogger.Println(err)
		return chain.SimulateTxResult{}, err
	}
	simQuery := abci.RequestQuery{
		Path: "/cosmos.tx.v1beta1.Service/Simulate",
		Data: txBytes,
	}

	opts := rpcclient.ABCIQueryOptions{
		Height: simQuery.Height,
		Prove:  simQuery.Prove,
	}
	response, err := rpcClient.ABCIQueryWithOptions(ctx, simQuery.Path, simQuery.Data, opts)
	if err != nil {
		c.errorLogger.Println(err)
		return chain.SimulateTxResult{}, err
	}

	if response.Response.Code != 0 {

	}

	var result txtypes.SimulateResponse
	if err = result.Unmarshal(response.Response.Value); err != nil {
		c.errorLogger.Println(err)
		return chain.SimulateTxResult{}, err
	}

	return chain.SimulateTxResult{
		GasUsed: float64(result.GasInfo.GasUsed),
	}, nil
}

func (c *LensCosmosClient) createFactory() tx.Factory {
	return tx.Factory{}.
		WithAccountRetriever(c.client).
		WithChainID(c.client.Config.ChainID).
		WithTxConfig(c.client.Codec.TxConfig).
		WithGasPrices(c.client.Config.GasPrices).
		WithKeybase(c.client.Keybase).
		WithSignMode(c.client.Config.SignMode())
}
