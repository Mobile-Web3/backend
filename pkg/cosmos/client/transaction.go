package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

func (c *Client) sign(key types.PrivKey, txf tx.Factory, txBuilder client.TxBuilder, overwriteSig bool) error {
	signMode := txf.SignMode()
	if signMode == signing.SignMode_SIGN_MODE_UNSPECIFIED {
		signMode = c.txConfig.SignModeHandler().DefaultMode()
	}

	pubKey := key.PubKey()

	signerData := authsigning.SignerData{
		ChainID:       txf.ChainID(),
		AccountNumber: txf.AccountNumber(),
		Sequence:      txf.Sequence(),
		PubKey:        pubKey,
		Address:       sdk.AccAddress(pubKey.Address()).String(),
	}

	// For SIGN_MODE_DIRECT, calling SetSignatures calls setSignerInfos on
	// TxBuilder under the hood, and SignerInfos is needed to generated the
	// sign bytes. This is the reason for setting SetSignatures here, with a
	// nil signature.
	//
	// Note: this line is not needed for SIGN_MODE_LEGACY_AMINO, but putting it
	// also doesn't affect its generated sign bytes, so for code's simplicity
	// sake, we put it here.
	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: txf.Sequence(),
	}

	var err error
	var prevSignatures []signing.SignatureV2
	if !overwriteSig {
		prevSignatures, err = txBuilder.GetTx().GetSignaturesV2()
		if err != nil {
			return err
		}
	}

	// Overwrite or append signer infos.
	var sigs []signing.SignatureV2
	if overwriteSig {
		sigs = []signing.SignatureV2{sig}
	} else {
		sigs = append(prevSignatures, sig)
	}
	if err := txBuilder.SetSignatures(sigs...); err != nil {
		return err
	}

	// Generate the bytes to be signed.
	bytesToSign, err := c.txConfig.SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return err
	}

	sigBytes, err := key.Sign(bytesToSign)
	if err != nil {
		return err
	}

	// Construct the SignatureV2 struct
	sigData = signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: sigBytes,
	}
	sig = signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: txf.Sequence(),
	}

	if overwriteSig {
		return txBuilder.SetSignatures(sig)
	}
	prevSignatures = append(prevSignatures, sig)
	return txBuilder.SetSignatures(prevSignatures...)
}

type SendTransactionData struct {
	ChainID     string
	Memo        string
	GasAdjusted string
	GasPrice    string
	ChainPrefix string
	Key         string
	Message     sdk.Msg
}

func (c *Client) CreateSignedTransaction(ctx context.Context, input SendTransactionData) ([]byte, error) {
	txContext, err := c.createTxFactory(ctx, input.ChainID, input.ChainPrefix, input.Key)
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

	if err = c.sign(txContext.PrivateKey, txFactory, builder, false); err != nil {
		return nil, err
	}

	return c.txConfig.TxEncoder()(builder.GetTx())
}

type protoTxProvider interface {
	GetProtoTx() *txtypes.Tx
}

type SimulateTransactionData struct {
	ChainID     string
	Memo        string
	ChainPrefix string
	Key         string
	Message     sdk.Msg
}

func (c *Client) CreateSimulateTransaction(ctx context.Context, input SimulateTransactionData) ([]byte, error) {
	txContext, err := c.createTxFactory(ctx, input.ChainID, input.ChainPrefix, input.Key)
	if err != nil {
		return nil, err
	}
	factory := txContext.Factory

	builder, err := factory.BuildUnsignedTx(input.Message)
	if err != nil {
		return nil, err
	}

	sig := signing.SignatureV2{
		PubKey: txContext.PrivateKey.PubKey(),
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
