package chain

import (
	"context"
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/crypto/types"
)

type AccountResponse struct {
	Key       string   `json:"key"`
	Addresses []string `json:"addresses"`
}

type CreateAccountInput struct {
	Mnemonic      string   `json:"mnemonic"`
	CoinType      uint32   `json:"coinType"`
	AccountPath   uint32   `json:"accountPath"`
	IndexPath     uint32   `json:"indexPath"`
	ChainPrefixes []string `json:"chainPrefixes"`
}

func (s *Service) CreateAccount(ctx context.Context, input CreateAccountInput) (AccountResponse, error) {
	privateKey, err := s.cosmosClient.CreateAccountFromMnemonic(input.Mnemonic, "", input.CoinType, input.AccountPath, input.IndexPath)
	if err != nil {
		return AccountResponse{}, err
	}

	addresses, err := s.getAddresses(privateKey, input.ChainPrefixes)
	if err != nil {
		return AccountResponse{}, err
	}

	return AccountResponse{
		Key:       hex.EncodeToString(privateKey.Bytes()),
		Addresses: addresses,
	}, nil
}

type RestoreAccountInput struct {
	Key           string   `json:"key"`
	ChainPrefixes []string `json:"chainPrefixes"`
}

func (s *Service) RestoreAccount(ctx context.Context, input RestoreAccountInput) (AccountResponse, error) {
	privateKey, err := s.cosmosClient.CreateAccountFromHexKey(input.Key)
	if err != nil {
		return AccountResponse{}, err
	}

	addresses, err := s.getAddresses(privateKey, input.ChainPrefixes)
	if err != nil {
		return AccountResponse{}, err
	}

	return AccountResponse{
		Key:       hex.EncodeToString(privateKey.Bytes()),
		Addresses: addresses,
	}, nil
}

func (s *Service) getAddresses(key types.PrivKey, prefixes []string) ([]string, error) {
	var addresses []string
	address := key.PubKey().Address()
	for _, prefix := range prefixes {
		addr, err := s.cosmosClient.ConvertAddressPrefix(prefix, address)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, addr)
	}

	return addresses, nil
}
