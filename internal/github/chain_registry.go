package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/google/go-github/v49/github"
)

var errBadResponse = errors.New("chain-registry repository respond with bad status")

type ChainRegistryClient struct {
	logger log.Logger
	client *github.Client
}

func NewChainRegistryClient(logger log.Logger) *ChainRegistryClient {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	return &ChainRegistryClient{
		logger: logger,
		client: github.NewClient(httpClient),
	}
}

func (c *ChainRegistryClient) UploadChainInfo(ctx context.Context) ([]chain.Chain, error) {
	tree, res, err := c.client.Git.GetTree(
		ctx,
		"cosmos",
		"chain-registry",
		"master",
		false)
	if err != nil {
		c.logger.Error(err)
		return nil, err
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("bad request from github; status code: %d", res.StatusCode)
		c.logger.Error(err)
		return nil, errBadResponse
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
		go c.uploadChain(chainURL, assetURL, storage, wg)
	}

	wg.Wait()
	err = storage.getError()
	if err != nil {
		return nil, err
	}

	index := 0
	chains := make([]chain.Chain, len(storage.chains))
	for _, chainName := range chainNames {
		chainData, ok := storage.getByName(chainName)
		if !ok {
			continue
		}
		chains[index] = chainData
		index++
	}

	return chains, nil
}

type assets struct {
	Assets []chain.Asset `json:"assets"`
}

func (c *ChainRegistryClient) uploadChain(chainURL string, assetURL string, storage *registryStorage, wg *sync.WaitGroup) {
	defer wg.Done()
	response, err := http.Get(chainURL)
	if err != nil {
		c.logger.Error(err)
		storage.addChain(chain.Chain{}, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		storage.addChain(chain.Chain{}, errBadResponse)
		return
	}

	chainData := chain.Chain{}
	decoder := json.NewDecoder(response.Body)
	if err = decoder.Decode(&chainData); err != nil {
		c.logger.Error(err)
		storage.addChain(chain.Chain{}, err)
		return
	}

	if chainData.Slip44 == 118 || chainData.Slip44 == 60 {
		assetResponse, err := http.Get(assetURL)
		if err != nil {
			c.logger.Error(err)
			storage.addChain(chain.Chain{}, err)
			return
		}
		defer assetResponse.Body.Close()

		if assetResponse.StatusCode != http.StatusOK {
			storage.addChain(chain.Chain{}, errBadResponse)
			return
		}

		asset := assets{}
		assetDecoder := json.NewDecoder(assetResponse.Body)
		if err = assetDecoder.Decode(&asset); err != nil {
			c.logger.Error(err)
			storage.addChain(chain.Chain{}, err)
			return
		}

		if len(asset.Assets) == 0 {
			storage.addChain(chain.Chain{}, fmt.Errorf("chain: %s, asset not found", chainData.Name))
			return
		}

		chainData.Asset = asset.Assets[0]
		lowPrice, averagePrice, highPrice := chain.GetGasPrices(chainData.Asset.Base, chainData.Fees.FeeTokens)
		chainData.LowGasPrice = lowPrice
		chainData.AverageGasPrice = averagePrice
		chainData.HighGasPrice = highPrice

		rpc, err := chain.ValidateRPCUrls(chainData.Api.Rpc)
		if err != nil {
			c.logger.Error(err)
			storage.addChain(chain.Chain{}, err)
			return
		}

		chainData.Api.Rpc = rpc
		storage.addChain(chainData, nil)
	}
}

type registryStorage struct {
	chains []chain.Chain
	errs   []error
	mutex  sync.Mutex
}

func (s *registryStorage) addChain(chain chain.Chain, err error) {
	s.mutex.Lock()
	s.chains = append(s.chains, chain)
	if err != nil {
		s.errs = append(s.errs, err)
	}
	s.mutex.Unlock()
}

func (s *registryStorage) getByName(name string) (chain.Chain, bool) {
	for _, chainData := range s.chains {
		if chainData.Name == name {
			return chainData, true
		}
	}

	return chain.Chain{}, false
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
