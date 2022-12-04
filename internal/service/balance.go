package service

import (
	"context"
	"math"
	"math/big"

	"github.com/Mobile-Web3/backend/pkg/cosmos"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type BalanceService struct {
	chainClient *cosmos.ChainClient
}

func NewBalanceService(chainClient *cosmos.ChainClient) *BalanceService {
	return &BalanceService{
		chainClient: chainClient,
	}
}

type BalanceResponse struct {
	TotalAmount     string `json:"totalAmount"`
	AvailableAmount string `json:"availableAmount"`
	StakedAmount    string `json:"stakedAmount"`
}

func (s *BalanceService) GetBalance(ctx context.Context, walletAddress string) (BalanceResponse, error) {
	chain, err := s.chainClient.GetChainByWallet(walletAddress)
	if err != nil {
		return BalanceResponse{}, err
	}

	rpcConnection, err := s.chainClient.GetRPCConnection(ctx, chain)
	if err != nil {
		return BalanceResponse{}, err
	}

	bankClient := bank.NewQueryClient(rpcConnection)
	bankResponse, err := bankClient.AllBalances(ctx, &bank.QueryAllBalancesRequest{
		Address: walletAddress,
	})
	if err != nil {
		return BalanceResponse{}, err
	}

	stackingClient := staking.NewQueryClient(rpcConnection)
	stakingResponse, err := stackingClient.DelegatorDelegations(context.Background(), &staking.QueryDelegatorDelegationsRequest{
		DelegatorAddr: walletAddress,
	})
	if err != nil {
		return BalanceResponse{}, err
	}

	response := BalanceResponse{}
	for _, denomUnit := range chain.Asset.DenomUnits {
		if denomUnit.Denom == chain.Asset.Display {
			multiplier := big.NewFloat(0).SetFloat64(math.Pow(10, float64(denomUnit.Exponent)))

			availableAmount := big.NewFloat(0)
			if len(bankResponse.Balances) != 0 {
				availableAmount.SetInt(bankResponse.Balances[0].Amount.BigInt())
			}

			stakedAmount := big.NewFloat(0)
			if len(stakingResponse.DelegationResponses) != 0 {
				stakedAmount.SetInt(stakingResponse.DelegationResponses[0].Balance.Amount.BigInt())
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
