package chain

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/Mobile-Web3/backend/pkg/cosmos/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	abci "github.com/tendermint/tendermint/abci/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

var ErrInvalidChainPair = errors.New("invalid chain pair from to")

type SendTxInput struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      string `json:"amount"`
	Mnemonic    string `json:"mnemonic"`
	Memo        string `json:"memo"`
	GasAdjusted string `json:"gasAdjusted"`
	GasPrice    string `json:"gasPrice"`
}

type SendTxResponse struct {
	Height    int64  `json:"height"`
	TxHash    string `json:"txHash"`
	Data      string `json:"data"`
	GasWanted int64  `json:"gasWanted"`
	GasUsed   int64  `json:"gasUsed"`
	RawLog    string `json:"rawLog"`
}

func (s *Service) SendTransaction(ctx context.Context, input SendTxInput) (SendTxResponse, error) {
	fromChain, err := s.getChainByWallet(ctx, input.From)
	if err != nil {
		return SendTxResponse{}, err
	}

	toChain, err := s.getChainByWallet(ctx, input.To)
	if err != nil {
		return SendTxResponse{}, err
	}

	if fromChain.ID != toChain.ID {
		return SendTxResponse{}, ErrInvalidChainPair
	}

	amount, err := fromChain.FromDisplayToBase(input.Amount)
	if err != nil {
		return SendTxResponse{}, err
	}

	gasPrice, err := fromChain.FromDisplayToBase(input.GasPrice)
	if err != nil {
		return SendTxResponse{}, err
	}

	coins, err := sdk.ParseCoinNormalized(amount)
	if err != nil {
		return SendTxResponse{}, err
	}

	msgSend := &bank.MsgSend{
		FromAddress: input.From,
		ToAddress:   input.To,
		Amount:      sdk.Coins{coins},
	}

	txBytes, err := s.cosmosClient.CreateSignedTransaction(ctx, client.SendTransactionData{
		ChainID:     fromChain.ID,
		Memo:        input.Memo,
		GasAdjusted: input.GasAdjusted,
		GasPrice:    gasPrice,
		CoinType:    fromChain.Slip44,
		ChainPrefix: toChain.Prefix,
		Mnemonic:    input.Mnemonic,
		Message:     msgSend,
	})
	if err != nil {
		return SendTxResponse{}, err
	}

	rpcClient, err := s.cosmosClient.GetChainRPC(ctx, toChain.ID)
	if err != nil {
		return SendTxResponse{}, err
	}

	response, err := rpcClient.BroadcastTxSync(ctx, txBytes)
	if err != nil {
		return SendTxResponse{}, err
	}

	if response.Code != 0 {
		err = fmt.Errorf("transaction failed with code: %d; TxHash: %s", response.Code, response.Hash.String())
		return SendTxResponse{}, err
	}

	return SendTxResponse{
		Height:    0,
		TxHash:    response.Hash.String(),
		Data:      response.Data.String(),
		GasWanted: 0,
		GasUsed:   0,
		RawLog:    response.Log,
	}, nil
}

type SimulateTxInput struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
	Mnemonic string `json:"mnemonic"`
	Memo     string `json:"memo"`
}

type SimulateTxResponse struct {
	GasAdjusted     string `json:"gasAdjusted"`
	LowGasPrice     string `json:"lowGasPrice"`
	AverageGasPrice string `json:"averageGasPrice"`
	HighGasPrice    string `json:"highGasPrice"`
}

func (s *Service) SimulateTransaction(ctx context.Context, input SimulateTxInput) (SimulateTxResponse, error) {
	fromChain, err := s.getChainByWallet(ctx, input.From)
	if err != nil {
		return SimulateTxResponse{}, err
	}

	toChain, err := s.getChainByWallet(ctx, input.To)
	if err != nil {
		return SimulateTxResponse{}, err
	}

	if fromChain.ID != toChain.ID {
		return SimulateTxResponse{}, ErrInvalidChainPair
	}

	amount, err := fromChain.FromDisplayToBase(input.Amount)
	if err != nil {
		return SimulateTxResponse{}, err
	}

	coins, err := sdk.ParseCoinNormalized(amount)
	if err != nil {
		return SimulateTxResponse{}, err
	}

	msgSend := &bank.MsgSend{
		FromAddress: input.From,
		ToAddress:   input.To,
		Amount:      sdk.Coins{coins},
	}

	txBytes, err := s.cosmosClient.CreateSimulateTransaction(ctx, client.SimulateTransactionData{
		ChainID:     toChain.ID,
		Memo:        input.Memo,
		CoinType:    toChain.Slip44,
		ChainPrefix: toChain.Prefix,
		Mnemonic:    input.Mnemonic,
		Message:     msgSend,
	})
	if err != nil {
		return SimulateTxResponse{}, err
	}

	simQuery := abci.RequestQuery{
		Path: "/cosmos.tx.v1beta1.Service/Simulate",
		Data: txBytes,
	}

	opts := rpcclient.ABCIQueryOptions{
		Height: simQuery.Height,
		Prove:  simQuery.Prove,
	}

	rpcClient, err := s.cosmosClient.GetChainRPC(ctx, toChain.ID)
	if err != nil {
		return SimulateTxResponse{}, err
	}

	response, err := rpcClient.ABCIQueryWithOptions(ctx, simQuery.Path, simQuery.Data, opts)
	if err != nil {
		return SimulateTxResponse{}, err
	}

	if response.Response.Code != 0 {

	}

	var result txtypes.SimulateResponse
	if err = result.Unmarshal(response.Response.Value); err != nil {
		return SimulateTxResponse{}, err
	}

	gasAdjusted := math.Round(float64(result.GasInfo.GasUsed) * s.gasAdjustment)
	_, exponent, err := toChain.GetBaseDenom()
	if err != nil {
		return SimulateTxResponse{}, err
	}

	divider := math.Pow(10, float64(exponent))
	lowGasPrice := math.Round(gasAdjusted*toChain.LowGasPrice) / divider
	averageGasPrice := math.Round(gasAdjusted*toChain.AverageGasPrice) / divider
	highGasPrice := math.Round(gasAdjusted*toChain.HighGasPrice) / divider

	return SimulateTxResponse{
		GasAdjusted:     fmt.Sprintf("%.0f", gasAdjusted),
		LowGasPrice:     fmt.Sprintf("%f", lowGasPrice),
		AverageGasPrice: fmt.Sprintf("%f", averageGasPrice),
		HighGasPrice:    fmt.Sprintf("%f", highGasPrice),
	}, nil
}
