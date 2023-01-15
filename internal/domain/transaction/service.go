package transaction

import (
	"context"
	"fmt"
	"math"

	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/pkg/cosmos"
	"github.com/Mobile-Web3/backend/pkg/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	abci "github.com/tendermint/tendermint/abci/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

type Service struct {
	gasAdjustment   float64
	logger          log.Logger
	chainRepository chain.Repository
	cosmosClient    *cosmos.Client
}

func NewService(gasAdjustment float64, logger log.Logger, chainRepository chain.Repository, cosmosClient *cosmos.Client) *Service {
	return &Service{
		gasAdjustment:   gasAdjustment,
		logger:          logger,
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

	denom, exponent, err := chain.GetBaseDenom(fromChain.Asset.Base, fromChain.Asset.Display, fromChain.Asset.DenomUnits)
	if err != nil {
		err = fmt.Errorf("chain: %s; %s", fromChain.Name, err.Error())
		s.logger.Error(err)
		return SendResponse{}, err
	}

	amount, err := chain.FromDisplayToBase(input.Amount, denom, exponent)
	if err != nil {
		err = fmt.Errorf("denom converting; chain: %s; amount: %s; denom: %s; %s", fromChain.Name, input.Amount, denom, err.Error())
		s.logger.Error(err)
		return SendResponse{}, err
	}

	gasPrice, err := chain.FromDisplayToBase(input.GasPrice, denom, exponent)
	if err != nil {
		err = fmt.Errorf("denom converting; chain: %s; amount: %s; denom: %s; %s", fromChain.Name, input.GasPrice, denom, err.Error())
		s.logger.Error(err)
		return SendResponse{}, err
	}

	coins, err := sdk.ParseCoinNormalized(amount)
	if err != nil {
		s.logger.Error(err)
		return SendResponse{}, err
	}

	msgSend := &bank.MsgSend{
		FromAddress: input.From,
		ToAddress:   input.To,
		Amount:      sdk.Coins{coins},
	}

	txBytes, err := s.cosmosClient.CreateSignedTransaction(ctx, cosmos.SendTransactionData{
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

	rpcClient := s.cosmosClient.GetChainHttpClient(fromChain.ID)
	response, err := rpcClient.BroadcastTxSync(ctx, txBytes)
	if err != nil {
		s.logger.Error(err)
		return SendResponse{}, err
	}

	if response.Code != 0 {
		s.logger.Error(err)
		return SendResponse{}, err
	}

	return SendResponse{
		TxHash: response.Hash.String(),
	}, nil
}

type SendInputFirebase struct {
	ChainID       string `json:"chainId"`
	From          string `json:"from"`
	To            string `json:"to"`
	Amount        string `json:"amount"`
	Key           string `json:"key"`
	Memo          string `json:"memo"`
	GasAdjusted   string `json:"gasAdjusted"`
	GasPrice      string `json:"gasPrice"`
	FirebaseToken string `json:"firebaseToken"`
}

type SendResponseFirebase struct {
	TxHash     string `json:"txHash"`
	WithEvents bool   `json:"withEvents"`
}

func (s *Service) SendTransactionWithEvents(ctx context.Context, input SendInputFirebase) (SendResponseFirebase, error) {
	fromChain, err := s.chainRepository.GetByID(ctx, input.ChainID)
	if err != nil {
		return SendResponseFirebase{}, err
	}

	denom, exponent, err := chain.GetBaseDenom(fromChain.Asset.Base, fromChain.Asset.Display, fromChain.Asset.DenomUnits)
	if err != nil {
		err = fmt.Errorf("chain: %s; %s", fromChain.Name, err.Error())
		s.logger.Error(err)
		return SendResponseFirebase{}, err
	}

	amount, err := chain.FromDisplayToBase(input.Amount, denom, exponent)
	if err != nil {
		err = fmt.Errorf("denom converting; chain: %s; amount: %s; denom: %s; %s", fromChain.Name, input.Amount, denom, err.Error())
		s.logger.Error(err)
		return SendResponseFirebase{}, err
	}

	gasPrice, err := chain.FromDisplayToBase(input.GasPrice, denom, exponent)
	if err != nil {
		err = fmt.Errorf("denom converting; chain: %s; amount: %s; denom: %s; %s", fromChain.Name, input.GasPrice, denom, err.Error())
		s.logger.Error(err)
		return SendResponseFirebase{}, err
	}

	coins, err := sdk.ParseCoinNormalized(amount)
	if err != nil {
		s.logger.Error(err)
		return SendResponseFirebase{}, err
	}

	msgSend := &bank.MsgSend{
		FromAddress: input.From,
		ToAddress:   input.To,
		Amount:      sdk.Coins{coins},
	}

	txBytes, err := s.cosmosClient.CreateSignedTransaction(ctx, cosmos.SendTransactionData{
		ChainID:     fromChain.ID,
		Memo:        input.Memo,
		GasAdjusted: input.GasAdjusted,
		GasPrice:    gasPrice,
		ChainPrefix: fromChain.Prefix,
		Key:         input.Key,
		Message:     msgSend,
	})
	if err != nil {
		return SendResponseFirebase{}, err
	}

	withEvents := true
	websocketClient := s.cosmosClient.GetChainWebsocketClient(fromChain.ID)
	subscriber, err := websocketClient.SubscribeToTx(ctx, input.From, map[string]interface{}{
		"token": input.FirebaseToken,
	})
	if err != nil {
		withEvents = false
	}

	rpcClient := s.cosmosClient.GetChainHttpClient(fromChain.ID)
	response, err := rpcClient.BroadcastTxSync(ctx, txBytes)
	if err != nil {
		websocketClient.UnsubscribeFromTx(subscriber)
		s.logger.Error(err)
		return SendResponseFirebase{}, err
	}

	if response.Code != 0 {
		websocketClient.UnsubscribeFromTx(subscriber)
		err = fmt.Errorf("transaction failed with code: %d; TxHash: %s; log: %s", response.Code, response.Hash.String(), response.Log)
		s.logger.Error(err)
		return SendResponseFirebase{}, err
	}

	return SendResponseFirebase{
		TxHash:     response.Hash.String(),
		WithEvents: withEvents,
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

	denom, exponent, err := chain.GetBaseDenom(fromChain.Asset.Base, fromChain.Asset.Display, fromChain.Asset.DenomUnits)
	if err != nil {
		err = fmt.Errorf("chain: %s; %s", fromChain.Name, err.Error())
		s.logger.Error(err)
		return SimulateResponse{}, err
	}

	amount, err := chain.FromDisplayToBase(input.Amount, denom, exponent)
	if err != nil {
		err = fmt.Errorf("denom converting; chain: %s; amount: %s; denom: %s; %s", fromChain.Name, input.Amount, denom, err.Error())
		s.logger.Error(err)
		return SimulateResponse{}, err
	}

	coins, err := sdk.ParseCoinNormalized(amount)
	if err != nil {
		s.logger.Error(err)
		return SimulateResponse{}, err
	}

	msgSend := &bank.MsgSend{
		FromAddress: input.From,
		ToAddress:   input.To,
		Amount:      sdk.Coins{coins},
	}

	txBytes, err := s.cosmosClient.CreateSimulateTransaction(ctx, cosmos.SimulateTransactionData{
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

	rpcClient := s.cosmosClient.GetChainHttpClient(fromChain.ID)
	response, err := rpcClient.ABCIQueryWithOptions(ctx, simQuery.Path, simQuery.Data, opts)
	if err != nil {
		s.logger.Error(err)
		return SimulateResponse{}, err
	}

	if response.Response.Code != 0 {
		err = fmt.Errorf("transaction failed with code %d. log: %s",
			response.Response.Code,
			response.Response.Log)
		s.logger.Error(err)
		return SimulateResponse{}, err
	}

	var result txtypes.SimulateResponse
	if err = result.Unmarshal(response.Response.Value); err != nil {
		s.logger.Error(err)
		return SimulateResponse{}, err
	}

	gasAdjusted := math.Round(float64(result.GasInfo.GasUsed) * s.gasAdjustment)
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
