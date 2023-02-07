package memory

import (
	"context"

	"github.com/Mobile-Web3/backend/internal/domain/chain"
)

type ChainsLavaRepository struct {
	lavaEndpoints map[string][]string
	repository    chain.Repository
}

func NewChainLavaRepository(repository chain.Repository) *ChainsLavaRepository {
	lavaEndpoints := map[string][]string{
		"cosmoshub-4": {"https://endpoints-testnet-1.lavanet.xyz:443/gateway/cos5/rpc-http/a60943bcfd533d305df0818fc2b0e028"},
		"osmosis-1":   {"https://endpoints-testnet-1.lavanet.xyz:443/gateway/cos3/rpc-http/a60943bcfd533d305df0818fc2b0e028"},
		"juno-1":      {"https://endpoints-testnet-1.lavanet.xyz:443/gateway/jun1/rpc-http/a60943bcfd533d305df0818fc2b0e028"},
	}

	return &ChainsLavaRepository{
		lavaEndpoints: lavaEndpoints,
		repository:    repository,
	}
}

func (r *ChainsLavaRepository) GetAllChains(ctx context.Context) ([]chain.ShortResponse, error) {
	return r.repository.GetAllChains(ctx)
}

func (r *ChainsLavaRepository) GetByID(ctx context.Context, chainID string) (chain.Chain, error) {
	return r.repository.GetByID(ctx, chainID)
}

func (r *ChainsLavaRepository) UpdateChains(ctx context.Context, chains []chain.Chain) error {
	return r.repository.UpdateChains(ctx, chains)
}

func (r *ChainsLavaRepository) GetRPCEndpoints(ctx context.Context, chainID string) ([]string, error) {
	for key, endpoints := range r.lavaEndpoints {
		if key == chainID {
			return endpoints, nil
		}
	}

	return r.repository.GetRPCEndpoints(ctx, chainID)
}
