package chain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/google/go-github/v49/github"
)

var errBadResponse = errors.New("chain-registry repository respond with bad status")

type Registry interface {
	UploadChainInfo(ctx context.Context) error
}

type registry struct {
	repository Repository
}

func NewRegistry(repository Repository) Registry {
	return &registry{
		repository: repository,
	}
}

func (r *registry) UploadChainInfo(ctx context.Context) error {
	client := github.NewClient(http.DefaultClient)
	tree, res, err := client.Git.GetTree(
		ctx,
		"cosmos",
		"chain-registry",
		"master",
		false)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errBadResponse
	}

	var chainNames []string
	for _, entry := range tree.Entries {
		if *entry.Type == "tree" &&
			!strings.HasPrefix(*entry.Path, ".") &&
			*entry.Path != "_non-cosmos" &&
			*entry.Path != "_IBC" &&
			*entry.Path != "thorchain" &&
			*entry.Path != "testnets" {
			chainNames = append(chainNames, *entry.Path)
		}
	}

	wg := &sync.WaitGroup{}
	storage := &registryStorage{
		mutex: sync.Mutex{},
	}
	for _, chainName := range chainNames {
		chainURL := fmt.Sprintf("https://raw.githubusercontent.com/cosmos/chain-registry/master/%s/chain.json", chainName)
		assetURL := fmt.Sprintf("https://raw.githubusercontent.com/cosmos/chain-registry/master/%s/assetlist.json", chainName)
		wg.Add(1)
		go r.uploadChain(chainURL, assetURL, storage, wg)
	}

	wg.Wait()
	err = storage.getError()
	if err != nil {
		return err
	}

	index := 0
	chains := make([]Chain, len(storage.chains))
	for _, chainName := range chainNames {
		chainData, ok := storage.getByName(chainName)
		if !ok {
			continue
		}
		chains[index] = chainData
		index++
	}

	return r.repository.UpdateChains(ctx, chains)
}

type assets struct {
	Assets []Asset `json:"assets"`
}

func (r *registry) uploadChain(chainURL string, assetURL string, storage *registryStorage, wg *sync.WaitGroup) {
	defer wg.Done()
	response, err := http.Get(chainURL)
	if err != nil {
		storage.addChain(Chain{}, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		storage.addChain(Chain{}, errBadResponse)
		return
	}

	chainData := Chain{}
	decoder := json.NewDecoder(response.Body)
	if err = decoder.Decode(&chainData); err != nil {
		storage.addChain(Chain{}, err)
		return
	}

	if chainData.Slip44 == 118 || chainData.Slip44 == 60 {
		assetResponse, err := http.Get(assetURL)
		if err != nil {
			storage.addChain(Chain{}, err)
			return
		}
		defer assetResponse.Body.Close()

		if assetResponse.StatusCode != http.StatusOK {
			storage.addChain(Chain{}, errBadResponse)
			return
		}

		asset := assets{}
		assetDecoder := json.NewDecoder(assetResponse.Body)
		if err = assetDecoder.Decode(&asset); err != nil {
			storage.addChain(Chain{}, err)
			return
		}

		if len(asset.Assets) == 0 {
			storage.addChain(Chain{}, fmt.Errorf("chain: %s, asset not found", chainData.Name))
			return
		}

		chainData.Asset = asset.Assets[0]
		chainData.InitGasPrice()

		if err = chainData.InitRPCUrls(); err != nil {
			storage.addChain(Chain{}, err)
			return
		}

		storage.addChain(chainData, nil)
	}
}

type registryStorage struct {
	chains []Chain
	errs   []error
	mutex  sync.Mutex
}

func (s *registryStorage) addChain(chain Chain, err error) {
	s.mutex.Lock()
	s.chains = append(s.chains, chain)
	if err != nil {
		s.errs = append(s.errs, err)
	}
	s.mutex.Unlock()
}

func (s *registryStorage) getByName(name string) (Chain, bool) {
	for _, chainData := range s.chains {
		if chainData.Name == name {
			return chainData, true
		}
	}

	return Chain{}, false
}

func (s *registryStorage) getError() error {
	if len(s.errs) == 0 {
		return nil
	}

	var sb strings.Builder
	_, sbErr := sb.WriteString("Errors while downloading registry data: ")
	if sbErr != nil {
		return sbErr
	}

	for _, err := range s.errs {
		_, sbErr = sb.WriteString(err.Error())
		if sbErr != nil {
			return sbErr
		}

		_, sbErr = sb.WriteString("; ")
		if sbErr != nil {
			return sbErr
		}
	}

	return errors.New(sb.String())
}
