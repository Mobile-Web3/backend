package client

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/tendermint/tendermint/rpc/client/http"
)

var errNoAvailableRPC = errors.New("no available rpc")

type chainState struct {
	chainID          string
	mutex            sync.RWMutex
	rpc              []string
	rpcClient        *http.HTTP
	isConnectionInit bool
	rpcLifetime      time.Duration
	rpcTimer         *time.Timer
}

func newChainState(chainID string, lifetime time.Duration, rpc []string) *chainState {
	state := &chainState{
		chainID:     chainID,
		mutex:       sync.RWMutex{},
		rpc:         rpc,
		rpcLifetime: lifetime,
	}

	state.rpcTimer = time.AfterFunc(lifetime, state.invalidateRPC)
	state.rpcTimer.Stop()
	return state
}

func (s *chainState) GetActiveRPC(ctx context.Context) (*http.HTTP, error) {
	s.mutex.RLock()
	if s.isConnectionInit {
		s.mutex.RUnlock()
		return s.rpcClient, nil
	}

	s.mutex.RUnlock()
	return s.initHealthRPC(ctx)
}

func (s *chainState) invalidateRPC() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.isConnectionInit = false
}

type endpointStorage struct {
	mutex     sync.Mutex
	endpoints []string
}

func (s *endpointStorage) addEndpoint(endpoint string) {
	s.mutex.Lock()
	s.endpoints = append(s.endpoints, endpoint)
	s.mutex.Unlock()
}

func (s *chainState) initHealthRPC(ctx context.Context) (*http.HTTP, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.rpcTimer.Stop()
	wg := &sync.WaitGroup{}
	storage := &endpointStorage{}

	for _, endpoint := range s.rpc {
		wg.Add(1)
		go checkHealthRPC(ctx, endpoint, storage, wg)
	}

	wg.Wait()
	if len(storage.endpoints) == 0 {
		return nil, errNoAvailableRPC
	}

	var endpoint string
	n := len(storage.endpoints) - 1
	switch n {
	case 0:
		endpoint = storage.endpoints[0]
	default:
		endpoint = storage.endpoints[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(n)]
	}

	rpcClient, err := client.NewClientFromNode(endpoint)
	if err != nil {
		return nil, err
	}

	s.rpcClient = rpcClient
	s.rpcTimer.Reset(s.rpcLifetime)
	s.isConnectionInit = true
	return s.rpcClient, nil
}

func checkHealthRPC(ctx context.Context, endpoint string, storage *endpointStorage, wg *sync.WaitGroup) {
	defer wg.Done()
	err := healthcheckRPC(ctx, endpoint)
	if err != nil {
		return
	}

	storage.addEndpoint(endpoint)
}
