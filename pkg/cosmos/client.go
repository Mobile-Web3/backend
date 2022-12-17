package cosmos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"
)

var (
	ErrInvalidGitURL      = errors.New("invalid git url")
	ErrEmptyRegistryURL   = errors.New("empty CHAIN_REGISTRY_URL env")
	ErrEmptyRegistryDir   = errors.New("empty REGISTRY_DIR env")
	ErrEmptyGasAdjustment = errors.New("empty GAS_ADJUSTMENT env")
	ErrChainNotFound      = errors.New("chain not found")
)

type ChainClient struct {
	registryURL   string
	registryDir   string
	chains        map[string]*Chain
	mu            sync.RWMutex
	zapLogger     *zap.Logger
	homePath      string
	gasAdjustment float64
}

func NewChainClient() (*ChainClient, error) {
	url := os.Getenv("CHAIN_REGISTRY_URL")
	if url == "" {
		return nil, ErrEmptyRegistryURL
	}

	registryDir := os.Getenv("REGISTRY_DIR")
	if registryDir == "" {
		return nil, ErrEmptyRegistryDir
	}

	_, _, ok := strings.Cut(url, "http://")
	if !ok {
		_, _, ok = strings.Cut(url, "https://")
		if !ok {
			return nil, ErrInvalidGitURL
		}
	}

	_, _, ok = strings.Cut(url, ".git")
	if !ok {
		return nil, ErrInvalidGitURL
	}

	gasAdjustmentStr := os.Getenv("GAS_ADJUSTMENT")
	if gasAdjustmentStr == "" {
		return nil, ErrEmptyGasAdjustment
	}
	gasAdjustment, err := strconv.ParseFloat(gasAdjustmentStr, 64)
	if err != nil {
		return nil, err
	}

	chainClient := &ChainClient{
		registryURL:   url,
		registryDir:   registryDir,
		mu:            sync.RWMutex{},
		zapLogger:     zap.L(),
		homePath:      os.Getenv("HOME"),
		gasAdjustment: gasAdjustment,
	}

	if err = chainClient.UploadChainInfo(); err != nil {
		return nil, err
	}

	return chainClient, nil
}

func (cc *ChainClient) UploadChainInfo() error {
	dirs, err := os.ReadDir(cc.registryDir)
	if err != nil {
		return err
	}

	chainRPCLifetimeENV := os.Getenv("CHAIN_RPC_LIFETIME")
	if chainRPCLifetimeENV == "" {
		chainRPCLifetimeENV = "10m"
	}

	rpcLifetime, err := time.ParseDuration(chainRPCLifetimeENV)
	if err != nil {
		return err
	}

	chains := make(map[string]*Chain)
	for _, dir := range dirs {
		isNotSystemDir := true
		for _, excludingDirName := range excludingDirs {
			if dir.Name() == excludingDirName {
				isNotSystemDir = false
				break
			}
		}
		if dir.Type() == os.ModeDir && isNotSystemDir {
			chainFileName := fmt.Sprintf("%s/%s/chain.json", cc.registryDir, dir.Name())
			assetFileName := fmt.Sprintf("%s/%s/assetlist.json", cc.registryDir, dir.Name())
			if readFileErr := cc.readChainFile(chainFileName, assetFileName, rpcLifetime, chains); readFileErr != nil {
				return readFileErr
			}
		}
	}

	cc.mu.Lock()
	cc.chains = chains
	cc.mu.Unlock()
	return nil
}

func (cc *ChainClient) GetAllChains() []*Chain {
	var result []*Chain
	for _, chain := range cc.chains {
		result = append(result, chain)
	}
	return result
}

func (cc *ChainClient) GetChainByWallet(walletAddress string) (*Chain, error) {
	var chain *Chain
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	for key := range cc.chains {
		before, _, ok := strings.Cut(walletAddress, key)
		if ok && before == "" {
			chain = cc.chains[key]
			break
		}
	}

	if chain == nil {
		return nil, ErrChainNotFound
	}

	return chain, nil
}

func (cc *ChainClient) GetRPCConnection(ctx context.Context, chain *Chain) (RPCConnection, error) {
	rpc, err := chain.getRpc(ctx)
	if err != nil {
		return nil, err
	}

	config := rpcConfig{
		RpcURL:        rpc,
		ChainID:       chain.ID,
		ChainPrefix:   chain.Prefix,
		HomePath:      cc.homePath,
		GasAdjustment: cc.gasAdjustment,
	}

	lensClient, err := newLensClient(cc.zapLogger, config)
	if err != nil {
		return nil, err
	}

	return lensClient, nil
}

func (cc *ChainClient) GetRPCConnectionWithMnemonic(ctx context.Context, mnemonic string, chain *Chain) (RPCConnection, error) {
	rpc, err := chain.getRpc(ctx)
	if err != nil {
		return nil, err
	}

	config := rpcConfig{
		RpcURL:        rpc,
		ChainID:       chain.ID,
		ChainPrefix:   chain.Prefix,
		HomePath:      cc.homePath,
		Slip44:        chain.Slip44,
		Mnemonic:      mnemonic,
		GasAdjustment: cc.gasAdjustment,
	}

	lensClient, err := newLensClient(cc.zapLogger, config)
	if err != nil {
		return nil, err
	}

	return lensClient, nil
}

type assets struct {
	Assets []Asset `json:"assets"`
}

func (cc *ChainClient) readChainFile(chainFileName string, assetFileName string, rpcLifetime time.Duration, chains map[string]*Chain) error {
	chainFile, err := os.ReadFile(chainFileName)
	if err != nil {
		return err
	}

	chain := Chain{}
	if err = json.Unmarshal(chainFile, &chain); err != nil {
		return err
	}

	chainInfo := registry.NewChainInfo(cc.zapLogger)
	if err = json.Unmarshal(chainFile, &chainInfo); err != nil {
		return err
	}

	assetFile, err := os.ReadFile(assetFileName)
	if err != nil {
		return err
	}

	asset := assets{}
	if err = json.Unmarshal(assetFile, &asset); err != nil {
		return err
	}

	chain.Info = chainInfo
	chain.Asset = asset.Assets[0]
	chain.rpcLifetime = rpcLifetime
	chain.rpcMutex = sync.RWMutex{}

	for _, feeToken := range chain.Fees.FeeTokens {
		if feeToken.Denom == chain.Asset.Base {
			if feeToken.LowGasPrice <= 0 && feeToken.AverageGasPrice <= 0 && feeToken.HighGasPrice <= 0 {
				chain.LowGasPrice = gasPriceLow + feeToken.MinGasPrice
				chain.AverageGasPrice = gasPriceAverage + feeToken.MinGasPrice
				chain.HighGasPrice = gasPriceHigh + feeToken.MinGasPrice
				break
			}

			chain.LowGasPrice = feeToken.LowGasPrice + feeToken.MinGasPrice
			chain.AverageGasPrice = feeToken.AverageGasPrice + feeToken.MinGasPrice
			chain.HighGasPrice = feeToken.HighGasPrice + feeToken.MinGasPrice
			break
		}
	}

	chains[chain.Prefix] = &chain
	return nil
}
