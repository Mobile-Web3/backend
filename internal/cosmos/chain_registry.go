package cosmos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Mobile-Web3/backend/internal/chain"
)

var excludingDirs = []string{".git", ".github", "_IBC", "_non-cosmos", "testnets", "thorchain"}

type ChainRegistry struct {
	registryURL string
	registryDir string
	repository  chain.Repository
}

func NewChainRegistry(registryURL string, registryDir string, repository chain.Repository) (*ChainRegistry, error) {
	chainRegistry := &ChainRegistry{
		registryDir: registryDir,
		registryURL: registryURL,
		repository:  repository,
	}

	return chainRegistry, nil
}

func (cr *ChainRegistry) UploadChainInfo(ctx context.Context) error {
	dirs, err := os.ReadDir(cr.registryDir)
	if err != nil {
		return err
	}

	var chains []chain.Chain
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
			chainData, readFileErr := cr.readChainFile(chainFileName, assetFileName)
			if chainData.Slip44 == 118 || chainData.Slip44 == 60 {
				if readFileErr != nil {
					return readFileErr
				}
				chains = append(chains, chainData)
			}
		}
	}

	return cr.repository.UpdateChains(ctx, chains)
}

type assets struct {
	Assets []chain.Asset `json:"assets"`
}

func (cr *ChainRegistry) readChainFile(chainFileName string, assetFileName string) (chain.Chain, error) {
	chainFile, err := os.ReadFile(chainFileName)
	if err != nil {
		return chain.Chain{}, err
	}

	chainData := chain.Chain{}
	if err = json.Unmarshal(chainFile, &chainData); err != nil {
		return chain.Chain{}, err
	}

	assetFile, err := os.ReadFile(assetFileName)
	if err != nil {
		return chain.Chain{}, err
	}

	asset := assets{}
	if err = json.Unmarshal(assetFile, &asset); err != nil {
		return chain.Chain{}, err
	}

	chainData.Asset = asset.Assets[0]
	chainData.InitGasPrice()

	if err = chainData.InitRPCUrls(); err != nil {
		return chain.Chain{}, err
	}

	return chainData, nil
}
