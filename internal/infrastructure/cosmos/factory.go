package cosmos

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"go.uber.org/zap"
)

type ClientFactory struct {
	homePath    string
	errorLogger *log.Logger
	zapLogger   *zap.Logger
	mutex       sync.RWMutex
	rpcLifetime time.Duration
	chains      map[string]*chainState
}

func NewClientFactory(rpcLifetime time.Duration, errorLogger *log.Logger) *ClientFactory {
	return &ClientFactory{
		homePath:    os.Getenv("HOME"),
		errorLogger: errorLogger,
		zapLogger:   zap.L(),
		mutex:       sync.RWMutex{},
		rpcLifetime: rpcLifetime,
		chains:      make(map[string]*chainState),
	}
}

func (f *ClientFactory) GetRPCConnection(ctx context.Context, config chain.RPCConfig) (chain.RPCConnection, error) {
	state := f.getChainState(config.ChainID)

	if state == nil {
		state = f.initChainState(config.ChainID, config.RPC)
	}

	rpc, err := state.getActiveRPC(ctx)
	if err != nil {
		if !errors.Is(err, errNoAvailableRPC) {
			f.errorLogger.Println(err)
		}
		return nil, err
	}

	client, err := newLensClient(f.errorLogger, f.zapLogger, rpcConfig{
		ChainID:     config.ChainID,
		ChainPrefix: config.ChainPrefix,
		RpcURL:      rpc,
		HomePath:    f.homePath,
		CoinType:    config.CoinType,
		Key:         config.Key,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (f *ClientFactory) getChainState(chainID string) *chainState {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	state := f.chains[chainID]
	return state
}

func (f *ClientFactory) initChainState(chainID string, rpcData []chain.Rpc) *chainState {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	var rpc []string
	for _, rpcInfo := range rpcData {
		rpc = append(rpc, rpcInfo.Address)
	}

	state := newChainState(chainID, f.rpcLifetime, rpc)
	f.chains[chainID] = state
	return state
}
