package client

import (
	"encoding/hex"
	"errors"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	ethhd "github.com/evmos/ethermint/crypto/hd"
)

var ErrInvalidMnemonicSize = errors.New("invalid mnemonic size, available values: 12, 24")

func (c *Client) CreateMnemonic(mnemonicSize uint8) (string, error) {
	var entropySize int
	switch mnemonicSize {
	case 12:
		entropySize = 128
	case 24:
		entropySize = 256
	default:
		return "", ErrInvalidMnemonicSize
	}

	entropy, err := bip39.NewEntropy(entropySize)
	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
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
		return nil, ErrUnsupportedCoinType
	}

	path := hd.CreateHDPath(coinType, account, index)
	derivedKey, err := algo.Derive()(mnemonic, passphrase, path.String())
	if err != nil {
		return nil, err
	}

	return algo.Generate()(derivedKey), nil
}

func (c *Client) CreateAccountFromHexKey(key string) (types.PrivKey, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}

	return &secp256k1.PrivKey{
		Key: keyBytes,
	}, nil
}

func (c *Client) ConvertAddressPrefix(chainPrefix string, address types.Address) (string, error) {
	return sdk.Bech32ifyAddressBytes(chainPrefix, address)
}
