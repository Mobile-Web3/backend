package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ethhd "github.com/evmos/ethermint/crypto/hd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var ErrUnsupportedCoinType = errors.New("unsupported coin type")

func (c *Client) newTxFactory(chainID string, keyBase keyring.Keyring) tx.Factory {
	return tx.Factory{}.
		WithChainID(chainID).
		WithTxConfig(c.txConfig).
		WithKeybase(keyBase).
		WithSignMode(c.signMode)
}

func (c *Client) getAccount(ctx context.Context, address string, chainID string) (authtypes.AccountI, error) {
	var header metadata.MD

	grpcConn, err := c.GetGrpcConnection(ctx, chainID)
	if err != nil {
		return nil, err
	}

	queryClient := authtypes.NewQueryClient(grpcConn)
	res, err := queryClient.Account(ctx, &authtypes.QueryAccountRequest{Address: address}, grpc.Header(&header))
	if err != nil {
		return nil, err
	}
	blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	if l := len(blockHeight); l != 1 {
		return nil, fmt.Errorf("unexpected '%s' header length; got %d, expected: %d", grpctypes.GRPCBlockHeightHeader, l, 1)
	}

	var acc authtypes.AccountI
	if err = c.interfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return nil, err
	}

	return acc, nil
}

func (c *Client) prepareTxFactory(ctx context.Context, chainID string, chainPrefix string, factory tx.Factory, keyRecord *keyring.Record) (tx.Factory, error) {
	address, err := keyRecord.GetAddress()
	if err != nil {
		return tx.Factory{}, err
	}

	accNumber := factory.AccountNumber()
	accSequence := factory.Sequence()

	if accNumber == 0 || accSequence == 0 {
		addr, bechErr := sdk.Bech32ifyAddressBytes(chainPrefix, address)
		if bechErr != nil {
			return tx.Factory{}, bechErr
		}

		accountInfo, accErr := c.getAccount(ctx, addr, chainID)
		if accErr != nil {
			return tx.Factory{}, accErr
		}

		number, sequence := accountInfo.GetAccountNumber(), accountInfo.GetSequence()

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

func (c *Client) createTxFactory(ctx context.Context, chainID string, chainPrefix string, coinType uint32, mnemonic string) (TxContext, error) {
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
	txf, err = c.prepareTxFactory(ctx, chainID, chainPrefix, txf, info)
	if err != nil {
		return TxContext{}, err
	}

	return TxContext{
		Factory:   txf,
		KeyRecord: info,
	}, nil
}
