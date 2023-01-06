package transaction

import (
	"context"
	"fmt"
	"math"

	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/pkg/cosmos/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	abci "github.com/tendermint/tendermint/abci/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

type Service struct {
	gasAdjustment   float64
	chainRepository chain.Repository
	cosmosClient    *client.Client
}

func NewService(gasAdjustment float64, chainRepository chain.Repository, cosmosClient *client.Client) *Service {
	return &Service{
		gasAdjustment:   gasAdjustment,
		chainRepository: chainRepository,
		cosmosClient:    cosmosClient,
	}
}

type SendInput struct {
	ChainID     string `json:"chainId"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      string `json:"amount"`
	Key         string `json:"key"`
	Memo        string `json:"memo"`
	GasAdjusted string `json:"gasAdjusted"`
	GasPrice    string `json:"gasPrice"`
}

type SendResponse struct {
	TxHash string `json:"txHash"`
}

func (s *Service) SendTransaction(ctx context.Context, input SendInput) (SendResponse, error) {
	fromChain, err := s.chainRepository.GetByID(ctx, input.ChainID)
	if err != nil {
		return SendResponse{}, err
	}

	amount, err := fromChain.FromDisplayToBase(input.Amount)
	if err != nil {
		return SendResponse{}, err
	}

	gasPrice, err := fromChain.FromDisplayToBase(input.GasPrice)
	if err != nil {
		return SendResponse{}, err
	}

	coins, err := sdk.ParseCoinNormalized(amount)
	if err != nil {
		return SendResponse{}, err
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
		ChainPrefix: fromChain.Prefix,
		Key:         input.Key,
		Message:     msgSend,
	})
	if err != nil {
		return SendResponse{}, err
	}

	rpcClient, err := s.cosmosClient.GetChainRPC(ctx, fromChain.ID)
	if err != nil {
		return SendResponse{}, err
	}

	response, err := rpcClient.BroadcastTxSync(ctx, txBytes)
	if err != nil {
		return SendResponse{}, err
	}

	if response.Code != 0 {
		err = fmt.Errorf("transaction failed with code: %d; TxHash: %s", response.Code, response.Hash.String())
		return SendResponse{}, err
	}

	return SendResponse{
		TxHash: response.Hash.String(),
	}, nil
}

type SimulateInput struct {
	ChainID string `json:"chainId"`
	From    string `json:"from"`
	To      string `json:"to"`
	Amount  string `json:"amount"`
	Key     string `json:"key"`
	Memo    string `json:"memo"`
}

type SimulateResponse struct {
	GasAdjusted     string `json:"gasAdjusted"`
	LowGasPrice     string `json:"lowGasPrice"`
	AverageGasPrice string `json:"averageGasPrice"`
	HighGasPrice    string `json:"highGasPrice"`
}

func (s *Service) SimulateTransaction(ctx context.Context, input SimulateInput) (SimulateResponse, error) {
	fromChain, err := s.chainRepository.GetByID(ctx, input.ChainID)
	if err != nil {
		return SimulateResponse{}, err
	}

	amount, err := fromChain.FromDisplayToBase(input.Amount)
	if err != nil {
		return SimulateResponse{}, err
	}

	coins, err := sdk.ParseCoinNormalized(amount)
	if err != nil {
		return SimulateResponse{}, err
	}

	msgSend := &bank.MsgSend{
		FromAddress: input.From,
		ToAddress:   input.To,
		Amount:      sdk.Coins{coins},
	}

	txBytes, err := s.cosmosClient.CreateSimulateTransaction(ctx, client.SimulateTransactionData{
		ChainID:     fromChain.ID,
		Memo:        input.Memo,
		ChainPrefix: fromChain.Prefix,
		Key:         input.Key,
		Message:     msgSend,
	})
	if err != nil {
		return SimulateResponse{}, err
	}

	simQuery := abci.RequestQuery{
		Path: "/cosmos.tx.v1beta1.Service/Simulate",
		Data: txBytes,
	}

	opts := rpcclient.ABCIQueryOptions{
		Height: simQuery.Height,
		Prove:  simQuery.Prove,
	}

	rpcClient, err := s.cosmosClient.GetChainRPC(ctx, fromChain.ID)
	if err != nil {
		return SimulateResponse{}, err
	}

	response, err := rpcClient.ABCIQueryWithOptions(ctx, simQuery.Path, simQuery.Data, opts)
	if err != nil {
		return SimulateResponse{}, err
	}

	if response.Response.Code != 0 {
		return SimulateResponse{}, fmt.Errorf("transaction failed with code %d. log: %s",
			response.Response.Code,
			response.Response.Log)
	}

	var result txtypes.SimulateResponse
	if err = result.Unmarshal(response.Response.Value); err != nil {
		return SimulateResponse{}, err
	}

	gasAdjusted := math.Round(float64(result.GasInfo.GasUsed) * s.gasAdjustment)
	_, exponent, err := fromChain.GetBaseDenom()
	if err != nil {
		return SimulateResponse{}, err
	}

	divider := math.Pow(10, float64(exponent))
	lowGasPrice := math.Round(gasAdjusted*fromChain.LowGasPrice) / divider
	averageGasPrice := math.Round(gasAdjusted*fromChain.AverageGasPrice) / divider
	highGasPrice := math.Round(gasAdjusted*fromChain.HighGasPrice) / divider

	return SimulateResponse{
		GasAdjusted:     fmt.Sprintf("%.0f", gasAdjusted),
		LowGasPrice:     fmt.Sprintf("%f", lowGasPrice),
		AverageGasPrice: fmt.Sprintf("%f", averageGasPrice),
		HighGasPrice:    fmt.Sprintf("%f", highGasPrice),
	}, nil
}
