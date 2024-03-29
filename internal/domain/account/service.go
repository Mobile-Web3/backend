package account

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/pkg/cosmos"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Service struct {
	logger          log.Logger
	chainRepository chain.Repository
	cosmosClient    *cosmos.Client
}

func NewService(logger log.Logger, chainRepository chain.Repository, cosmosClient *cosmos.Client) *Service {
	return &Service{
		logger:          logger,
		chainRepository: chainRepository,
		cosmosClient:    cosmosClient,
	}
}

type CreateMnemonicInput struct {
	MnemonicSize uint8 `json:"mnemonicSize"`
}

func (input CreateMnemonicInput) Validate() error {
	if input.MnemonicSize < 12 || input.MnemonicSize > 24 {
		return fmt.Errorf("invalid mnemonic size, available values: 12, 24; provided size %d", input.MnemonicSize)
	}

	return nil
}

func (s *Service) CreateMnemonic(ctx context.Context, input CreateMnemonicInput) (string, error) {
	var entropySize int
	switch input.MnemonicSize {
	case 12:
		entropySize = 128
	case 24:
		entropySize = 256
	}
	return s.cosmosClient.CreateMnemonic(entropySize)
}

func (s *Service) getAddresses(key types.PrivKey, prefixes []string) ([]string, error) {
	addresses := make([]string, len(prefixes))
	address := key.PubKey().Address()
	for index, prefix := range prefixes {
		addr, err := s.cosmosClient.ConvertAddressPrefix(prefix, address)
		if err != nil {
			return nil, err
		}

		addresses[index] = addr
	}

	return addresses, nil
}

type KeyResponse struct {
	Key       string   `json:"key"`
	Addresses []string `json:"addresses"`
}

func formatErrors(errs []string) error {
	result := errs[0]
	for i := 1; i < len(errs); i++ {
		result = fmt.Sprintf("%s; %s", result, errs[i])
	}
	return errors.New(result)
}

type CreateAccountInput struct {
	Mnemonic      string   `json:"mnemonic"`
	CoinType      uint32   `json:"coinType"`
	AccountPath   uint32   `json:"accountPath"`
	IndexPath     uint32   `json:"indexPath"`
	ChainPrefixes []string `json:"chainPrefixes"`
}

func (input CreateAccountInput) Validate() error {
	var errs []string
	if input.Mnemonic == "" {
		errs = append(errs, "invalid mnemonic")
	}

	if input.CoinType != 118 && input.CoinType != 60 {
		errs = append(errs, fmt.Sprintf("coin type %d is not supported, supported coins - 60, 118", input.CoinType))
	}

	if len(input.ChainPrefixes) == 0 {
		errs = append(errs, "at least one chain is needed")
	}

	if len(errs) > 0 {
		return formatErrors(errs)
	}

	return nil
}

func (s *Service) CreateAccount(ctx context.Context, input CreateAccountInput) (KeyResponse, error) {
	privateKey, err := s.cosmosClient.CreateAccountFromMnemonic(input.Mnemonic, "", input.CoinType, input.AccountPath, input.IndexPath)
	if err != nil {
		return KeyResponse{}, err
	}

	addresses, err := s.getAddresses(privateKey, input.ChainPrefixes)
	if err != nil {
		return KeyResponse{}, err
	}

	return KeyResponse{
		Key:       hex.EncodeToString(privateKey.Bytes()),
		Addresses: addresses,
	}, nil
}

type RestoreAccountInput struct {
	Key           string   `json:"key"`
	ChainPrefixes []string `json:"chainPrefixes"`
}

func (input RestoreAccountInput) Validate() error {
	var errs []string
	if input.Key == "" {
		errs = append(errs, "invalid key")
	}

	if len(input.ChainPrefixes) == 0 {
		errs = append(errs, "at least one chain is needed")
	}

	if len(errs) > 0 {
		return formatErrors(errs)
	}

	return nil
}

func (s *Service) RestoreAccount(ctx context.Context, input RestoreAccountInput) (KeyResponse, error) {
	privateKey, err := s.cosmosClient.CreateAccountFromHexKey(input.Key)
	if err != nil {
		return KeyResponse{}, err
	}

	addresses, err := s.getAddresses(privateKey, input.ChainPrefixes)
	if err != nil {
		return KeyResponse{}, err
	}

	return KeyResponse{
		Key:       hex.EncodeToString(privateKey.Bytes()),
		Addresses: addresses,
	}, nil
}

type BalanceInput struct {
	ChainID string `json:"chainId"`
	Address string `json:"address"`
}

func (input BalanceInput) Validate() error {
	var errs []string
	if input.ChainID == "" {
		errs = append(errs, "invalid chainId")
	}

	if input.Address == "" {
		errs = append(errs, "invalid address")
	}

	if len(errs) > 0 {
		return formatErrors(errs)
	}

	return nil
}

type BalanceResponse struct {
	TotalAmount     string `json:"totalAmount"`
	AvailableAmount string `json:"availableAmount"`
	StakedAmount    string `json:"stakedAmount"`
}

func (s *Service) CheckBalance(ctx context.Context, input BalanceInput) (BalanceResponse, error) {
	chainInfo, err := s.chainRepository.GetByID(ctx, input.ChainID)
	if err != nil {
		return BalanceResponse{}, err
	}

	connection := s.cosmosClient.GetChainGrpcClient(chainInfo.ID)
	bankClient := bank.NewQueryClient(connection)
	bankResponse, err := bankClient.AllBalances(ctx, &bank.QueryAllBalancesRequest{
		Address: input.Address,
	})
	if err != nil {
		s.logger.Error(err)
		return BalanceResponse{}, err
	}

	stakingClient := staking.NewQueryClient(connection)
	stakingResponse, err := stakingClient.DelegatorDelegations(ctx, &staking.QueryDelegatorDelegationsRequest{
		DelegatorAddr: input.Address,
	})
	if err != nil {
		s.logger.Error(err)
		return BalanceResponse{}, err
	}

	response := BalanceResponse{}
	for _, denomUnit := range chainInfo.Asset.DenomUnits {
		if denomUnit.Denom == chainInfo.Asset.Display {
			total := sdkmath.NewInt(0)

			availableAmount := "0"
			if len(bankResponse.Balances) > 0 {
				availableAmount = chain.FromBaseToDisplay(bankResponse.Balances[0].Amount.String(), denomUnit.Exponent)
				total = total.Add(bankResponse.Balances[0].Amount)
			}

			stakedAmount := "0"
			if len(stakingResponse.DelegationResponses) > 0 {
				stakedAmount = chain.FromBaseToDisplay(stakingResponse.DelegationResponses[0].Balance.Amount.String(), denomUnit.Exponent)
				total = total.Add(stakingResponse.DelegationResponses[0].Balance.Amount)
			}

			response.AvailableAmount = availableAmount
			response.StakedAmount = stakedAmount
			response.TotalAmount = chain.FromBaseToDisplay(total.String(), denomUnit.Exponent)
			break
		}
	}

	return response, nil
}
