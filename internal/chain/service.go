package chain

import (
	"context"
	"errors"
	"strings"

	"github.com/Mobile-Web3/backend/pkg/cosmos/client"
)

var ErrChainNotFound = errors.New("chain not found")

type Service struct {
	gasAdjustment float64
	repository    Repository
	cosmosClient  *client.Client
}

func NewService(gasAdjustment float64, repository Repository, cosmosClient *client.Client) *Service {
	return &Service{
		gasAdjustment: gasAdjustment,
		repository:    repository,
		cosmosClient:  cosmosClient,
	}
}

func (s *Service) getChainByWallet(ctx context.Context, walletAddress string) (Chain, error) {
	prefixes, err := s.repository.GetAllPrefixes(ctx)
	if err != nil {
		return Chain{}, err
	}

	chain := Chain{}
	isFound := false
	for _, prefix := range prefixes {
		before, _, ok := strings.Cut(walletAddress, prefix)
		if ok && before == "" {
			chain, err = s.repository.GetChainByPrefix(ctx, prefix)
			if err != nil {
				return Chain{}, err
			}
			isFound = true
			break
		}
	}

	if !isFound {
		return Chain{}, ErrChainNotFound
	}

	return chain, nil
}
