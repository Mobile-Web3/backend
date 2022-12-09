package cosmos

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/grpc"
)

type RPCConnection interface {
	grpc.ClientConn
	SendMsg(ctx context.Context, msg sdk.Msg, memo string) (*sdk.TxResponse, error)
}
