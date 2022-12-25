package cosmos

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	tendermintRPC "github.com/tendermint/tendermint/rpc/client"
)

type CreateFactoryContext struct {
	ChainID          string
	TxConfig         client.TxConfig
	AccountRetriever client.AccountRetriever
	KeyBase          keyring.Keyring
	SignMode         signing.SignMode
}

func CreateTxFactory(ctx CreateFactoryContext) tx.Factory {
	return tx.Factory{}.
		WithAccountRetriever(ctx.AccountRetriever).
		WithChainID(ctx.ChainID).
		WithTxConfig(ctx.TxConfig).
		WithKeybase(ctx.KeyBase).
		WithSignMode(ctx.SignMode)
}

type PrepareFactoryContext struct {
	Key               string
	KeyBase           keyring.Keyring
	RPCClient         tendermintRPC.Client
	Factory           tx.Factory
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
}

func PrepareTxFactory(ctx PrepareFactoryContext) (tx.Factory, error) {
	record, err := ctx.KeyBase.Key(ctx.Key)
	if err != nil {
		return tx.Factory{}, err
	}

	address, err := record.GetAddress()
	if err != nil {
		return tx.Factory{}, err
	}

	clientCtx := client.Context{}.
		WithClient(ctx.RPCClient).
		WithInterfaceRegistry(ctx.InterfaceRegistry).
		WithCodec(ctx.Codec)

	if err = ctx.Factory.AccountRetriever().EnsureExists(clientCtx, address); err != nil {
		return tx.Factory{}, err
	}

	accNumber := ctx.Factory.AccountNumber()
	accSequence := ctx.Factory.Sequence()

	if accNumber == 0 || accSequence == 0 {
		number, sequence, getNumErr := ctx.Factory.AccountRetriever().GetAccountNumberSequence(clientCtx, address)
		if getNumErr != nil {
			return tx.Factory{}, getNumErr
		}

		if accNumber == 0 {
			ctx.Factory = ctx.Factory.WithAccountNumber(number)
		}

		if accSequence == 0 {
			ctx.Factory = ctx.Factory.WithSequence(sequence)
		}
	}

	return ctx.Factory, nil
}
