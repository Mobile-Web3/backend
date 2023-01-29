package chain

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Mobile-Web3/backend/pkg/cosmos"
	"github.com/cosmos/cosmos-sdk/types/query"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Service struct {
	registry     Registry
	repository   Repository
	cosmosClient *cosmos.Client
}

func NewService(registry Registry, repository Repository, cosmosClient *cosmos.Client) *Service {
	return &Service{
		registry:     registry,
		repository:   repository,
		cosmosClient: cosmosClient,
	}
}

func (s *Service) UpdateChainInfo(ctx context.Context) error {
	chains, err := s.registry.UploadChainInfo(ctx)
	if err != nil {
		return err
	}

	return s.repository.UpdateChains(ctx, chains)
}

type Validator struct {
	Address     string `json:"address"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Identity    string `json:"identity"`
	Tokens      string `json:"tokens"`
	Commission  string `json:"commission"`
	Website     string `json:"website"`
}

type PagedValidatorsInput struct {
	ChainID string
	Limit   uint64
	Offset  uint64
}

func (input PagedValidatorsInput) Validate() error {
	if input.ChainID == "" {
		return fmt.Errorf("invalid chainId")
	}

	return nil
}

type PagedValidatorsResponse struct {
	Limit  uint64      `json:"limit"`
	Offset uint64      `json:"offset"`
	Data   []Validator `json:"data"`
}

func (s *Service) GetPagedValidators(ctx context.Context, input PagedValidatorsInput) (PagedValidatorsResponse, error) {
	chainData, err := s.repository.GetByID(ctx, input.ChainID)
	if err != nil {
		return PagedValidatorsResponse{}, err
	}

	_, exponent, err := GetBaseDenom(chainData.Asset.Base, chainData.Asset.Display, chainData.Asset.DenomUnits)
	if err != nil {
		return PagedValidatorsResponse{}, err
	}

	connection := s.cosmosClient.GetChainGrpcClient(input.ChainID)
	client := staking.NewQueryClient(connection)
	response, err := client.Validators(ctx, &staking.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
		Pagination: &query.PageRequest{
			Limit:  input.Limit,
			Offset: input.Offset,
		},
	})
	if err != nil {
		return PagedValidatorsResponse{}, err
	}

	n := len(response.Validators)
	result := make([]Validator, n)
	for index, validator := range response.Validators {
		rateFloat, floatErr := strconv.ParseFloat(validator.Commission.Rate.String(), 64)
		if floatErr != nil {
			return PagedValidatorsResponse{}, floatErr
		}
		result[index] = Validator{
			Address:     validator.OperatorAddress,
			Name:        validator.Description.Moniker,
			Description: validator.Description.Details,
			Identity:    validator.Description.Identity,
			Tokens:      FromBaseToDisplay(validator.Tokens.String(), exponent),
			Commission:  fmt.Sprintf("%.1f", rateFloat*100),
			Website:     validator.Description.Website,
		}
	}

	return PagedValidatorsResponse{
		Limit:  input.Limit,
		Offset: input.Offset,
		Data:   result,
	}, nil
}
