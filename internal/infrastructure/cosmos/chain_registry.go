package cosmos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/Mobile-Web3/backend/internal/domain/chain"
)

var ErrChainNotFound = errors.New("chain not found")
var excludingDirs = []string{".git", ".github", "_IBC", "_non-cosmos", "testnets", "thorchain"}

type ChainRegistry struct {
	registryURL    string
	registryDir    string
	mutex          sync.RWMutex
	prefixes       []string
	chains         map[string]chain.Chain
	chainResponses []chain.ShortResponse
}

func NewChainRegistry(registryURL string, registryDir string) (*ChainRegistry, error) {
	chainRegistry := &ChainRegistry{
		registryDir: registryDir,
		registryURL: registryURL,
		mutex:       sync.RWMutex{},
		chains:      make(map[string]chain.Chain),
	}

	err := chainRegistry.UploadChainInfo()
	if err != nil {
		return nil, err
	}

	return chainRegistry, nil
}

func (cr *ChainRegistry) GetAllChains(ctx context.Context) ([]chain.ShortResponse, error) {
	return cr.chainResponses, nil
}

func (cr *ChainRegistry) GetAllPrefixes(ctx context.Context) ([]string, error) {
	return cr.prefixes, nil
}

func (cr *ChainRegistry) GetChainByPrefix(ctx context.Context, prefix string) (chain.Chain, error) {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()

	result, ok := cr.chains[prefix]
	if !ok {
		return chain.Chain{}, ErrChainNotFound
	}

	return result, nil
}

func (cr *ChainRegistry) UploadChainInfo() error {
	dirs, err := os.ReadDir(cr.registryDir)
	if err != nil {
		return err
	}

	cr.prefixes = nil
	chains := make(map[string]chain.Chain)
	var chainResponses []chain.ShortResponse
	for _, dir := range dirs {
		isNotSystemDir := true
		for _, excludingDirName := range excludingDirs {
			if dir.Name() == excludingDirName {
				isNotSystemDir = false
				break
			}
		}
		if dir.Type() == os.ModeDir && isNotSystemDir {
			chainFileName := fmt.Sprintf("%s/%s/chain.json", cr.registryDir, dir.Name())
			assetFileName := fmt.Sprintf("%s/%s/assetlist.json", cr.registryDir, dir.Name())
			chainResponse, readFileErr := cr.readChainFile(chainFileName, assetFileName, chains)
			if readFileErr != nil {
				return readFileErr
			}
			chainResponses = append(chainResponses, chainResponse)
		}
	}

	cr.mutex.Lock()
	defer cr.mutex.Unlock()
	cr.chains = chains
	cr.chainResponses = chainResponses
	return nil
}

type assets struct {
	Assets []chain.Asset `json:"assets"`
}

func (cr *ChainRegistry) readChainFile(chainFileName string, assetFileName string, chains map[string]chain.Chain) (chain.ShortResponse, error) {
	chainFile, err := os.ReadFile(chainFileName)
	if err != nil {
		return chain.ShortResponse{}, err
	}

	chainData := chain.Chain{}
	if err = json.Unmarshal(chainFile, &chainData); err != nil {
		return chain.ShortResponse{}, err
	}

	assetFile, err := os.ReadFile(assetFileName)
	if err != nil {
		return chain.ShortResponse{}, err
	}

	asset := assets{}
	if err = json.Unmarshal(assetFile, &asset); err != nil {
		return chain.ShortResponse{}, err
	}

	chainData.Asset = asset.Assets[0]
	chainData.InitGasPrice()

	if err = chainData.InitRPCUrls(); err != nil {
		return chain.ShortResponse{}, err
	}

	chains[chainData.Prefix] = chainData
	cr.prefixes = append(cr.prefixes, chainData.Prefix)
	return chain.ShortResponse{
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
	}, nil
}
