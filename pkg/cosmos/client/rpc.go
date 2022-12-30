package client

import (
	"context"

	"github.com/Mobile-Web3/backend/pkg/cosmos/chain"
	"github.com/tendermint/tendermint/rpc/client/http"
)

func (c *Client) getChainState(chainID string) *chain.State {
	c.chainMutex.RLock()
	defer c.chainMutex.RUnlock()
	state := c.chains[chainID]
	return state
}

func (c *Client) initChainState(chainID string, rpcEndpoints []string) *chain.State {
	c.chainMutex.Lock()
	defer c.chainMutex.Unlock()
	state := chain.NewState(chainID, c.rpcLifetime, rpcEndpoints)
	c.chains[chainID] = state
	return state
}

func (c *Client) GetChainRPC(ctx context.Context, chainID string) (*http.HTTP, error) {
	chainState := c.getChainState(chainID)
	if chainState == nil {
		endpoints, err := c.getRPCEndpointHandler(ctx, chainID)
		if err != nil {
			return nil, err
		}

		chainState = c.initChainState(chainID, endpoints)
	}

	return chainState.GetActiveRPC(ctx)
}
