package transaction

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/Mobile-Web3/backend/pkg/cosmos"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var ErrInvalidChainPair = errors.New("invalid chain pair from to")

type Service struct {
	chainClient *cosmos.ChainClient
}

func NewService(chainClient *cosmos.ChainClient) *Service {
	return &Service{
		chainClient: chainClient,
	}
}

type SendInput struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      string `json:"amount"`
	Mnemonic    string `json:"mnemonic"`
	Memo        string `json:"memo"`
	GasAdjusted string `json:"gasAdjusted"`
	GasPrice    string `json:"gasPrice"`
}

type SendResponse struct {
	Height    int64  `json:"height"`
	TxHash    string `json:"txHash"`
	Data      string `json:"data"`
	GasWanted int64  `json:"gasWanted"`
	GasUsed   int64  `json:"gasUsed"`
	RawLog    string `json:"rawLog"`
}

func (s *Service) Send(ctx context.Context, input SendInput) (SendResponse, error) {
	fromChain, err := s.chainClient.GetChainByWallet(input.From)
	if err != nil {
		return SendResponse{}, err
	}

	toChain, err := s.chainClient.GetChainByWallet(input.To)
	if err != nil {
		return SendResponse{}, err
	}

	if fromChain.ID != toChain.ID {
		return SendResponse{}, ErrInvalidChainPair
	}

	amount, err := fromChain.FromDisplayToBase(input.Amount)
	if err != nil {
		return SendResponse{}, err
	}

	gasPrice, err := fromChain.FromDisplayToBase(input.GasPrice)
	if err != nil {
		return SendResponse{}, err
	}

	rpcConnection, err := s.chainClient.GetRPCConnectionWithMnemonic(ctx, input.Mnemonic, fromChain)
	if err != nil {
		return SendResponse{}, err
	}

	coins, err := sdk.ParseCoinNormalized(amount)
	if err != nil {
		return SendResponse{}, err
	}

	msgSend := &banktypes.MsgSend{
		FromAddress: input.From,
		ToAddress:   input.To,
		Amount:      sdk.Coins{coins},
	}

	msgData := cosmos.MsgData{
		Memo:          input.Memo,
		GasAdjustment: input.GasAdjusted,
		GasPrice:      gasPrice,
		Msg:           msgSend,
	}

	response, err := rpcConnection.SendMsg(ctx, msgData)
	if err != nil {
		return SendResponse{}, err
	}

	return SendResponse{
		Height:    response.Height,
		TxHash:    response.TxHash,
		Data:      response.Data,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
		RawLog:    response.RawLog,
	}, nil
}

type SimulateInput struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
	Mnemonic string `json:"mnemonic"`
	Memo     string `json:"memo"`
}

type SimulateResponse struct {
	GasAdjusted     string `json:"gasAdjusted"`
	LowGasPrice     string `json:"lowGasPrice"`
	AverageGasPrice string `json:"averageGasPrice"`
	HighGasPrice    string `json:"highGasPrice"`
}

func (s *Service) Simulate(ctx context.Context, input SimulateInput) (SimulateResponse, error) {
	fromChain, err := s.chainClient.GetChainByWallet(input.From)
	if err != nil {
		return SimulateResponse{}, err
	}

	toChain, err := s.chainClient.GetChainByWallet(input.To)
	if err != nil {
		return SimulateResponse{}, err
	}

	if fromChain.ID != toChain.ID {
		return SimulateResponse{}, ErrInvalidChainPair
	}

	amount, err := fromChain.FromDisplayToBase(input.Amount)
	if err != nil {
		return SimulateResponse{}, err
	}

	rpcConnection, err := s.chainClient.GetRPCConnectionWithMnemonic(ctx, input.Mnemonic, fromChain)
	if err != nil {
		return SimulateResponse{}, err
	}

	coins, err := sdk.ParseCoinNormalized(amount)
	if err != nil {
		return SimulateResponse{}, err
	}

	msgSend := &banktypes.MsgSend{
		FromAddress: input.From,
		ToAddress:   input.To,
		Amount:      sdk.Coins{coins},
	}

	_, gasAdjustment, err := rpcConnection.CalculateGas(ctx, msgSend)
	if err != nil {
		return SimulateResponse{}, err
	}

	_, exponent, err := toChain.GetBaseDenom()
	if err != nil {
		return SimulateResponse{}, err
	}

	divider := math.Pow(10, float64(exponent))
	gasAdjFloat := float64(gasAdjustment)
	lowGasPrice := math.Round(gasAdjFloat*toChain.LowGasPrice) / divider
	averageGasPrice := math.Round(gasAdjFloat*toChain.AverageGasPrice) / divider
	highGasPrice := math.Round(gasAdjFloat*toChain.HighGasPrice) / divider

	response := SimulateResponse{
		GasAdjusted:     fmt.Sprintf("%d", gasAdjustment),
		LowGasPrice:     fmt.Sprintf("%f", lowGasPrice),
		AverageGasPrice: fmt.Sprintf("%f", averageGasPrice),
		HighGasPrice:    fmt.Sprintf("%f", highGasPrice),
	}
	return response, nil
}
