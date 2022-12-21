package chain

import (
	"context"
	"math"
	"math/big"
)

type CheckResponse struct {
	TotalAmount     string `json:"totalAmount"`
	AvailableAmount string `json:"availableAmount"`
	StakedAmount    string `json:"stakedAmount"`
}

func (s *Service) CheckBalance(ctx context.Context, walletAddress string) (CheckResponse, error) {
	chain, err := s.getChainByWallet(ctx, walletAddress)
	if err != nil {
		return CheckResponse{}, err
	}

	rpcConnection, err := s.connectionFactory.GetRPCConnection(ctx, RPCConfig{
		ChainID:     chain.ID,
		ChainPrefix: chain.Prefix,
		CoinType:    chain.Slip44,
		RPC:         chain.Api.Rpc,
	})
	if err != nil {
		return CheckResponse{}, err
	}

	balance, err := rpcConnection.GetBalance(ctx, walletAddress)
	if err != nil {
		return CheckResponse{}, err
	}

	response := CheckResponse{}
	for _, denomUnit := range chain.Asset.DenomUnits {
		if denomUnit.Denom == chain.Asset.Display {
			multiplier := big.NewFloat(0).SetFloat64(math.Pow(10, float64(denomUnit.Exponent)))
			availableAmount := big.NewFloat(0).SetInt(balance.AvailableAmount)
			stakedAmount := big.NewFloat(0).SetInt(balance.StakedAmount)
			totalAmount := big.NewFloat(0).Add(availableAmount, stakedAmount)
			response.AvailableAmount = availableAmount.Quo(availableAmount, multiplier).String()
			response.StakedAmount = stakedAmount.Quo(stakedAmount, multiplier).String()
			response.TotalAmount = totalAmount.Quo(totalAmount, multiplier).String()
			break
		}
	}

	return response, nil
}
