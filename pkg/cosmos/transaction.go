package cosmos

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type protoTxProvider interface {
	GetProtoTx() *txtypes.Tx
}

func BuildTransaction(factory tx.Factory, key *keyring.Record, msg sdk.Msg) ([]byte, error) {
	builder, err := factory.BuildUnsignedTx(msg)
	if err != nil {
		return nil, err
	}

	var pk cryptotypes.PubKey = &secp256k1.PubKey{}

	pk, err = key.GetPubKey()
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
