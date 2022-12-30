package chain

import (
	"context"
	"math"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc"
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

	connection, err := grpc.Dial("grpc-cosmoshub.blockapsis.com:9090",
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())))
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

			availableAmount := big.NewFloat(0).SetInt(bankResponse.Balances[0].Amount.BigInt())
			stakedAmount := big.NewFloat(0).SetInt(stakingResponse.DelegationResponses[0].Balance.Amount.BigInt())

			totalAmount := big.NewFloat(0).Add(availableAmount, stakedAmount)
			response.AvailableAmount = availableAmount.Quo(availableAmount, multiplier).String()
			response.StakedAmount = stakedAmount.Quo(stakedAmount, multiplier).String()
			response.TotalAmount = totalAmount.Quo(totalAmount, multiplier).String()
			break
		}
	}

	return response, nil
}
