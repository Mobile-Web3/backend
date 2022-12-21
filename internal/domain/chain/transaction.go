package chain

import (
	"context"
	"errors"
	"fmt"
	"math"
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

	rpcConnection, err := s.connectionFactory.GetRPCConnection(ctx, RPCConfig{
		ChainID:     fromChain.ID,
		ChainPrefix: fromChain.Prefix,
		CoinType:    fromChain.Slip44,
		Key:         input.Mnemonic,
		RPC:         fromChain.Api.Rpc,
	})

	return rpcConnection.SendTransaction(ctx, SendTxData{
		From:        input.From,
		To:          input.To,
		Amount:      amount,
		Memo:        input.Memo,
		GasPrice:    gasPrice,
		GasAdjusted: input.GasAdjusted,
	})
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

	rpcConnection, err := s.connectionFactory.GetRPCConnection(ctx, RPCConfig{
		ChainID:     fromChain.ID,
		ChainPrefix: fromChain.Prefix,
		CoinType:    fromChain.Slip44,
		Key:         input.Mnemonic,
		RPC:         fromChain.Api.Rpc,
	})
	if err != nil {
		return SimulateTxResponse{}, err
	}

	result, err := rpcConnection.SimulateTransaction(ctx, SimulateTxData{
		From:   input.From,
		To:     input.To,
		Memo:   input.Memo,
		Amount: amount,
	})
	if err != nil {
		return SimulateTxResponse{}, err
	}

	gasAdjusted := math.Round(result.GasUsed * s.gasAdjustment)
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
