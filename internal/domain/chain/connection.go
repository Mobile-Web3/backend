package chain

import (
	"context"
	"math/big"
)

type Balance struct {
	AvailableAmount *big.Int
	StakedAmount    *big.Int
}

type SendTxData struct {
	From        string
	To          string
	Amount      string
	Memo        string
	GasAdjusted string
	GasPrice    string
}

type SendTxResponse struct {
	Height    int64  `json:"height"`
	TxHash    string `json:"txHash"`
	Data      string `json:"data"`
	GasWanted int64  `json:"gasWanted"`
	GasUsed   int64  `json:"gasUsed"`
	RawLog    string `json:"rawLog"`
}

type SimulateTxData struct {
	From   string
	To     string
	Amount string
	Memo   string
}

type SimulateTxResult struct {
	GasUsed float64
}

type RPCConnection interface {
	GetBalance(ctx context.Context, address string) (Balance, error)
	SendTransaction(ctx context.Context, txData SendTxData) (SendTxResponse, error)
	SimulateTransaction(ctx context.Context, txData SimulateTxData) (SimulateTxResult, error)
}

type RPCConfig struct {
	ChainID     string
	ChainPrefix string
	CoinType    uint32
	Key         string
	RPC         []Rpc
}

type ConnectionFactory interface {
	GetRPCConnection(ctx context.Context, config RPCConfig) (RPCConnection, error)
}
