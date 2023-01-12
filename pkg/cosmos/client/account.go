package client

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	ethhd "github.com/evmos/ethermint/crypto/hd"
)

func (c *Client) CreateMnemonic(mnemonicSize uint8) (string, error) {
	var entropySize int
	switch mnemonicSize {
	case 12:
		entropySize = 128
	case 24:
		entropySize = 256
	default:
		err := fmt.Errorf("invalid mnemonic size, available values: 12, 24; provided size %d", mnemonicSize)
		c.logger.Error(err)
		return "", err
	}

	entropy, err := bip39.NewEntropy(entropySize)
	if err != nil {
		err = fmt.Errorf("creating entropy with size %d; %s", entropySize, err.Error())
		c.logger.Error(err)
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		err = fmt.Errorf("creating mnemonic; %s", err.Error())
		c.logger.Error(err)
		return "", err
	}

	return mnemonic, nil
}

func (c *Client) CreateAccountFromMnemonic(mnemonic string, passphrase string, coinType uint32, account uint32, index uint32) (types.PrivKey, error) {
	var algo keyring.SignatureAlgo
	switch coinType {
	case 118:
		algo = keyring.SignatureAlgo(hd.Secp256k1)
	case 60:
		algo = keyring.SignatureAlgo(ethhd.EthSecp256k1)
	default:
		err := fmt.Errorf("unsupported coin type; provided coin type %d", coinType)
		c.logger.Error(err)
		return nil, err
	}

	path := hd.CreateHDPath(coinType, account, index)
	derivedKey, err := algo.Derive()(mnemonic, passphrase, path.String())
	if err != nil {
		err = fmt.Errorf("deriving key; %s", err.Error())
		c.logger.Error(err)
		return nil, err
	}

	return algo.Generate()(derivedKey), nil
}

func (c *Client) CreateAccountFromHexKey(key string) (types.PrivKey, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		err = fmt.Errorf("decoding hexstring key; %s", err.Error())
		c.logger.Error(err)
		return nil, err
	}

	return &secp256k1.PrivKey{
		Key: keyBytes,
	}, nil
}

func (c *Client) ConvertAddressPrefix(chainPrefix string, address types.Address) (string, error) {
	result, err := sdk.Bech32ifyAddressBytes(chainPrefix, address)
	if err != nil {
		err = fmt.Errorf("converting cosmos address %s with prefix %s; %s", address, chainPrefix, err.Error())
		c.logger.Error(err)
		return "", err
	}

	return result, nil
}
