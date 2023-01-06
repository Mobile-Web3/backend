package chain

import (
	"context"
	"encoding/hex"
	"math"
	"math/big"

	"github.com/cosmos/cosmos-sdk/crypto/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type CreateMnemonicInput struct {
	MnemonicSize uint8 `json:"mnemonicSize"`
}

func (s *Service) CreateMnemonic(ctx context.Context, input CreateMnemonicInput) (string, error) {
	return s.cosmosClient.CreateMnemonic(input.MnemonicSize)
}

type AccountResponse struct {
	Key       string   `json:"key"`
	Addresses []string `json:"addresses"`
}

func (s *Service) getAddresses(key types.PrivKey, prefixes []string) ([]string, error) {
	var addresses []string
	address := key.PubKey().Address()
	for _, prefix := range prefixes {
		addr, err := s.cosmosClient.ConvertAddressPrefix(prefix, address)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, addr)
	}

	return addresses, nil
}

type CreateAccountInput struct {
	Mnemonic      string   `json:"mnemonic"`
	CoinType      uint32   `json:"coinType"`
	AccountPath   uint32   `json:"accountPath"`
	IndexPath     uint32   `json:"indexPath"`
	ChainPrefixes []string `json:"chainPrefixes"`
}

func (s *Service) CreateAccount(ctx context.Context, input CreateAccountInput) (AccountResponse, error) {
	privateKey, err := s.cosmosClient.CreateAccountFromMnemonic(input.Mnemonic, "", input.CoinType, input.AccountPath, input.IndexPath)
	if err != nil {
		return AccountResponse{}, err
	}

	addresses, err := s.getAddresses(privateKey, input.ChainPrefixes)
	if err != nil {
		return AccountResponse{}, err
	}

	return AccountResponse{
		Key:       hex.EncodeToString(privateKey.Bytes()),
		Addresses: addresses,
	}, nil
}

type RestoreAccountInput struct {
	Key           string   `json:"key"`
	ChainPrefixes []string `json:"chainPrefixes"`
}

func (s *Service) RestoreAccount(ctx context.Context, input RestoreAccountInput) (AccountResponse, error) {
	privateKey, err := s.cosmosClient.CreateAccountFromHexKey(input.Key)
	if err != nil {
		return AccountResponse{}, err
	}

	addresses, err := s.getAddresses(privateKey, input.ChainPrefixes)
	if err != nil {
		return AccountResponse{}, err
	}

	return AccountResponse{
		Key:       hex.EncodeToString(privateKey.Bytes()),
		Addresses: addresses,
	}, nil
}

type BalanceInput struct {
	Address string `json:"address"`
}

type BalanceResponse struct {
	TotalAmount     string `json:"totalAmount"`
	AvailableAmount string `json:"availableAmount"`
	StakedAmount    string `json:"stakedAmount"`
}

func (s *Service) CheckBalance(ctx context.Context, input BalanceInput) (BalanceResponse, error) {
	chain, err := s.getChainByWallet(ctx, input.Address)
	if err != nil {
		return BalanceResponse{}, err
	}

	connection, err := s.cosmosClient.GetGrpcConnection(ctx, chain.ID)
	if err != nil {
		return BalanceResponse{}, err
	}

	bankClient := bank.NewQueryClient(connection)
	bankResponse, err := bankClient.AllBalances(ctx, &bank.QueryAllBalancesRequest{
		Address: input.Address,
	})
	if err != nil {
		return BalanceResponse{}, err
	}

	stackingClient := staking.NewQueryClient(connection)
	stakingResponse, err := stackingClient.DelegatorDelegations(ctx, &staking.QueryDelegatorDelegationsRequest{
		DelegatorAddr: input.Address,
	})
	if err != nil {
		return BalanceResponse{}, err
	}

	response := BalanceResponse{}
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
