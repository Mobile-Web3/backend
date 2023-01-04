package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var ErrUnsupportedCoinType = errors.New("unsupported coin type")

func (c *Client) newTxFactory(chainID string) tx.Factory {
	return tx.Factory{}.
		WithChainID(chainID).
		WithTxConfig(c.txConfig).
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

func (c *Client) prepareTxFactory(ctx context.Context, chainID string, chainPrefix string, factory tx.Factory, address types.Address) (tx.Factory, error) {
	accNumber := factory.AccountNumber()
	accSequence := factory.Sequence()

	if accNumber == 0 || accSequence == 0 {
		addr, bechErr := c.ConvertAddressPrefix(chainPrefix, address)
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
	Factory    tx.Factory
	PrivateKey types.PrivKey
}

func (c *Client) createTxFactory(ctx context.Context, chainID string, chainPrefix string, key string) (TxContext, error) {
	privateKey, err := c.CreateAccountFromHexKey(key)
	if err != nil {
		return TxContext{}, err
	}

	txf := c.newTxFactory(chainID)
	txf, err = c.prepareTxFactory(ctx, chainID, chainPrefix, txf, privateKey.PubKey().Address())
	if err != nil {
		return TxContext{}, err
	}

	return TxContext{
		Factory:    txf,
		PrivateKey: privateKey,
	}, nil
}
