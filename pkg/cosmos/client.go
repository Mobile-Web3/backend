package cosmos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	lens "github.com/strangelove-ventures/lens/client"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"
)

var (
	ErrInvalidGitURL    = errors.New("invalid git url")
	ErrEmptyRegistryURL = errors.New("empty CHAIN_REGISTRY_URL")
	ErrEmptyRegistryDir = errors.New("empty REGISTRY_DIR")
	ErrChainNotFound    = errors.New("chain not found")
)

type ChainClient struct {
	registryURL string
	registryDir string
	chains      map[string]Chain
	mu          sync.RWMutex
	zapLogger   *zap.Logger
	homePath    string
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

	chainClient := &ChainClient{
		registryURL: url,
		registryDir: registryDir,
		mu:          sync.RWMutex{},
		zapLogger:   zap.L(),
		homePath:    os.Getenv("HOME"),
	}

	if err := chainClient.UploadChainInfo(); err != nil {
		return nil, err
	}

	return chainClient, nil
}

func (cc *ChainClient) UploadChainInfo() error {
	dirs, err := os.ReadDir(cc.registryDir)
	if err != nil {
		return err
	}

	chains := make(map[string]Chain)
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
			if readFileErr := cc.readChainFile(chainFileName, assetFileName, chains); readFileErr != nil {
				return readFileErr
			}
		}
	}

	cc.mu.Lock()
	cc.chains = chains
	cc.mu.Unlock()
	return nil
}

func (cc *ChainClient) GetChainByWallet(walletAddress string) (Chain, error) {
	var chain Chain
	cc.mu.RLock()
	for key := range cc.chains {
		before, _, ok := strings.Cut(walletAddress, key)
		if ok && before == "" {
			chain = cc.chains[key]
			break
		}
	}
	cc.mu.RUnlock()

	if chain.ID == "" {
		return chain, ErrChainNotFound
	}

	return chain, nil
}

func (cc *ChainClient) GetRPCConnection(ctx context.Context, chain Chain) (RPCConnection, error) {
	rpc, err := chain.Info.GetRandomRPCEndpoint(ctx)
	if err != nil {
		return nil, err
	}

	chainConfig := lens.ChainClientConfig{
		Key:            "default",
		KeyringBackend: "memory",
		RPCAddr:        rpc,
		AccountPrefix:  chain.Prefix,
		ChainID:        chain.ID,
		Timeout:        "5s",
	}

	chainClient, err := lens.NewChainClient(cc.zapLogger, &chainConfig, cc.homePath, os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}

	return chainClient, nil
}

func (cc *ChainClient) GetRPCConnectionWithMnemonic(ctx context.Context, mnemonic string, chain Chain) (RPCConnection, error) {
	rpc, err := chain.Info.GetRandomRPCEndpoint(ctx)
	if err != nil {
		return nil, err
	}

	chainConfig := lens.ChainClientConfig{
		Key:            "default",
		KeyringBackend: "memory",
		RPCAddr:        rpc,
		AccountPrefix:  chain.Prefix,
		ChainID:        chain.ID,
		GasAdjustment:  1.3,
		GasPrices:      chain.GetLowGasPrice(),
		Timeout:        "30s",
		OutputFormat:   "json",
		SignModeStr:    "direct",
		Modules:        lens.ModuleBasics,
	}

	chainClient, err := lens.NewChainClient(cc.zapLogger, &chainConfig, cc.homePath, os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}

	_, err = chainClient.RestoreKey("source_key", mnemonic, chain.Slip44)
	if err != nil {
		return nil, err
	}

	chainConfig.Key = "source_key"
	return chainClient, nil
}

type assets struct {
	Assets []Asset `json:"assets"`
}

func (cc *ChainClient) readChainFile(chainFileName string, assetFileName string, chains map[string]Chain) error {
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
	chains[chain.Prefix] = chain
	return nil
}
