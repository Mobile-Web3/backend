package cosmos

import (
	"context"
	"errors"
	"sync"

	"github.com/Mobile-Web3/backend/internal/chain"
)

var ErrChainNotFound = errors.New("chain not found")

type ChainRepository struct {
	prefixes  []string
	chains    map[string]chain.Chain
	responses []chain.ShortResponse
	mutex     sync.RWMutex
}

func NewChainRepository() *ChainRepository {
	return &ChainRepository{
		mutex: sync.RWMutex{},
	}
}

func (r *ChainRepository) GetAllChains(ctx context.Context) ([]chain.ShortResponse, error) {
	return r.responses, nil
}

func (r *ChainRepository) GetAllPrefixes(ctx context.Context) ([]string, error) {
	return r.prefixes, nil
}

func (r *ChainRepository) GetChainByPrefix(ctx context.Context, prefix string) (chain.Chain, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	chainData, ok := r.chains[prefix]
	if !ok {
		return chain.Chain{}, ErrChainNotFound
	}

	return chainData, nil
}

func (r *ChainRepository) UpdateChains(ctx context.Context, chains []chain.Chain) error {
	var prefixes []string
	var responses []chain.ShortResponse
	chainsMap := make(map[string]chain.Chain)

	for _, chainData := range chains {
		prefixes = append(prefixes, chainData.Prefix)
		chainsMap[chainData.Prefix] = chainData
		responses = append(responses, chain.ShortResponse{
			ID:          chainData.ID,
			Name:        chainData.Name,
			PrettyName:  chainData.PrettyName,
			Prefix:      chainData.Prefix,
			Slip44:      chainData.Slip44,
			Description: chainData.Asset.Description,
			Base:        chainData.Asset.Base,
			Symbol:      chainData.Asset.Symbol,
			Display:     chainData.Asset.Display,
			LogoPngURL:  chainData.Asset.Logo.Png,
			LogoSvgURL:  chainData.Asset.Logo.Svg,
			KeyAlgos:    chainData.KeyAlgos,
		})
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.chains = chainsMap
	r.prefixes = prefixes
	r.responses = responses
	return nil
}

func (r *ChainRepository) GetRPCEndpoints(ctx context.Context, chainID string) ([]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var chainData chain.Chain
	isFound := false
	for _, item := range r.chains {
		if item.ID == chainID {
			chainData = item
			isFound = true
			break
		}
	}

	if !isFound {
		return nil, ErrChainNotFound
	}

	var endpoints []string
	for _, endpoint := range chainData.Api.Rpc {
		endpoints = append(endpoints, endpoint.Address)
	}

	return endpoints, nil
}
