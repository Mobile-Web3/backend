package chain

import (
	"context"
	"math"
	"math/big"

	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
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

	connection, err := s.cosmosClient.GetGrpcConnection(ctx, chain.ID)
	if err != nil {
		return CheckResponse{}, err
	}

	bankClient := bank.NewQueryClient(connection)
	bankResponse, err := bankClient.AllBalances(ctx, &bank.QueryAllBalancesRequest{
		Address: walletAddress,
	})
	if err != nil {
		return CheckResponse{}, err
	}

	stackingClient := staking.NewQueryClient(connection)
	stakingResponse, err := stackingClient.DelegatorDelegations(ctx, &staking.QueryDelegatorDelegationsRequest{
		DelegatorAddr: walletAddress,
	})
	if err != nil {
		return CheckResponse{}, err
	}

	response := CheckResponse{}
	for _, denomUnit := range chain.Asset.DenomUnits {
		if denomUnit.Denom == chain.Asset.Display {
			multiplier := big.NewFloat(0).SetFloat64(math.Pow(10, float64(denomUnit.Exponent)))

			availableAmount := big.NewFloat(0)
			if len(bankResponse.Balances) > 0 {
				availableAmount = availableAmount.SetInt(bankResponse.Balances[0].Amount.BigInt())
			}

			stakedAmount := big.NewFloat(0)
			if len(stakingResponse.DelegationResponses) > 0 {
				stakedAmount = stakedAmount.SetInt(stakingResponse.DelegationResponses[0].Balance.Amount.BigInt())
			}

			totalAmount := big.NewFloat(0).Add(availableAmount, stakedAmount)
			response.AvailableAmount = availableAmount.Quo(availableAmount, multiplier).String()
			response.StakedAmount = stakedAmount.Quo(stakedAmount, multiplier).String()
			response.TotalAmount = totalAmount.Quo(totalAmount, multiplier).String()
			break
		}
	}

	return response, nil
}
