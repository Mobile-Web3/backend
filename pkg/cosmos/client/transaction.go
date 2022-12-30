package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type SendTransactionData struct {
	ChainID     string
	Memo        string
	GasAdjusted string
	GasPrice    string
	CoinType    uint32
	Mnemonic    string
	Message     sdk.Msg
}

func (c *Client) CreateSignedTransaction(ctx context.Context, input SendTransactionData) ([]byte, error) {
	txContext, err := c.createTxFactory(ctx, input.ChainID, input.CoinType, input.Mnemonic)
	if err != nil {
		return nil, err
	}

	txFactory := txContext.Factory

	if input.Memo != "" {
		txFactory = txFactory.WithMemo(input.Memo)
	}

	adjusted, err := strconv.ParseUint(input.GasAdjusted, 0, 64)
	if err != nil {
		return nil, err
	}

	txFactory = txFactory.WithGas(adjusted)
	txFactory = txFactory.WithFees(input.GasPrice)

	builder, err := txFactory.BuildUnsignedTx(input.Message)
	if err != nil {
		return nil, err
	}

	c.codec.MustMarshalJSON(input.Message)

	if err = tx.Sign(txFactory, c.keyName, builder, false); err != nil {
		return nil, err
	}

	return c.txConfig.TxEncoder()(builder.GetTx())
}

type protoTxProvider interface {
	GetProtoTx() *txtypes.Tx
}

type SimulateTransactionData struct {
	ChainID  string
	Memo     string
	CoinType uint32
	Mnemonic string
	Message  sdk.Msg
}

func (c *Client) CreateSimulateTransaction(ctx context.Context, input SimulateTransactionData) ([]byte, error) {
	txContext, err := c.createTxFactory(ctx, input.ChainID, input.CoinType, input.Mnemonic)
	if err != nil {
		return nil, err
	}
	factory := txContext.Factory

	builder, err := factory.BuildUnsignedTx(input.Message)
	if err != nil {
		return nil, err
	}

	var pk cryptotypes.PubKey = &secp256k1.PubKey{}

	pk, err = txContext.KeyRecord.GetPubKey()
	if err != nil {
		return nil, err
	}

	sig := signing.SignatureV2{
		PubKey: pk,
		Data: &signing.SingleSignatureData{
			SignMode: factory.SignMode(),
		},
		Sequence: factory.Sequence(),
	}
	if err = builder.SetSignatures(sig); err != nil {
		return nil, err
	}

	protoProvider, ok := builder.(protoTxProvider)
	if !ok {
		return nil, fmt.Errorf("cannot simulate amino tx")
	}

	simReq := txtypes.SimulateRequest{Tx: protoProvider.GetProtoTx()}
	return simReq.Marshal()
}
