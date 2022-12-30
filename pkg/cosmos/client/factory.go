package client

import (
	"context"
	"errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	ethhd "github.com/evmos/ethermint/crypto/hd"
)

var ErrUnsupportedCoinType = errors.New("unsupported coin type")

func (c *Client) newTxFactory(chainID string, keyBase keyring.Keyring) tx.Factory {
	return tx.Factory{}.
		WithAccountRetriever(nil). // TODO acc retriever
		WithChainID(chainID).
		WithTxConfig(c.txConfig).
		WithKeybase(keyBase).
		WithSignMode(c.signMode)
}

func (c *Client) prepareTxFactory(ctx context.Context, chainID string, factory tx.Factory, keyRecord *keyring.Record) (tx.Factory, error) {
	address, err := keyRecord.GetAddress()
	if err != nil {
		return tx.Factory{}, err
	}

	chainRPC, err := c.GetChainRPC(ctx, chainID)
	if err != nil {
		return tx.Factory{}, err
	}

	clientCtx := client.Context{}.
		WithClient(chainRPC).
		WithInterfaceRegistry(c.interfaceRegistry).
		WithCodec(c.codec)

	if err = factory.AccountRetriever().EnsureExists(clientCtx, address); err != nil {
		return tx.Factory{}, err
	}

	accNumber := factory.AccountNumber()
	accSequence := factory.Sequence()

	if accNumber == 0 || accSequence == 0 {
		number, sequence, getNumErr := factory.AccountRetriever().GetAccountNumberSequence(clientCtx, address)
		if getNumErr != nil {
			return tx.Factory{}, getNumErr
		}

		if accNumber == 0 {
			factory = factory.WithAccountNumber(number)
		}

		if accSequence == 0 {
			factory = factory.WithSequence(sequence)
		}
	}

	return factory, nil
}

type TxContext struct {
	Factory   tx.Factory
	KeyRecord *keyring.Record
}

func (c *Client) createTxFactory(ctx context.Context, chainID string, coinType uint32, mnemonic string) (TxContext, error) {
	keyBase := keyring.NewInMemory(c.codec)
	var algo keyring.SignatureAlgo

	switch coinType {
	case 118:
		algo = keyring.SignatureAlgo(hd.Secp256k1)
	case 60:
		algo = keyring.SignatureAlgo(ethhd.EthSecp256k1)
	default:
		return TxContext{}, ErrUnsupportedCoinType
	}

	info, err := keyBase.NewAccount(c.keyName, mnemonic, "", hd.CreateHDPath(coinType, 0, 0).String(), algo)
	if err != nil {
		return TxContext{}, err
	}

	txf := c.newTxFactory(chainID, keyBase)
	txf, err = c.prepareTxFactory(ctx, chainID, txf, info)

	return TxContext{
		Factory:   txf,
		KeyRecord: info,
	}, nil
}
