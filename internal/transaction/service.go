package transaction

import (
	"context"
	"errors"

	"github.com/Mobile-Web3/backend/pkg/cosmos"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var ErrInvalidChainPair = errors.New("invalid chain pair from to")

type Service struct {
	chainClient *cosmos.ChainClient
}

func NewService(chainClient *cosmos.ChainClient) *Service {
	return &Service{
		chainClient: chainClient,
	}
}

type SendInput struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
	Mnemonic string `json:"mnemonic"`
	Memo     string `json:"memo"`
}

type SendResponse struct {
	Height    int64  `json:"height"`
	TxHash    string `json:"txHash"`
	Data      string `json:"data"`
	GasWanted int64  `json:"gasWanted"`
	GasUsed   int64  `json:"gasUsed"`
	RawLog    string `json:"rawLog"`
}

func (s *Service) Send(ctx context.Context, input SendInput) (SendResponse, error) {
	fromChain, err := s.chainClient.GetChainByWallet(input.From)
	if err != nil {
		return SendResponse{}, err
	}

	toChain, err := s.chainClient.GetChainByWallet(input.To)
	if err != nil {
		return SendResponse{}, err
	}

	if fromChain.ID != toChain.ID {
		return SendResponse{}, ErrInvalidChainPair
	}

	amount, err := fromChain.FromDisplayToBase(input.Amount)
	if err != nil {
		return SendResponse{}, err
	}

	rpcConnection, err := s.chainClient.GetRPCConnectionWithMnemonic(ctx, input.Mnemonic, fromChain)
	if err != nil {
		return SendResponse{}, err
	}

	coins, err := sdk.ParseCoinNormalized(amount)
	if err != nil {
		return SendResponse{}, err
	}

	msgSend := &banktypes.MsgSend{
		FromAddress: input.From,
		ToAddress:   input.To,
		Amount:      sdk.Coins{coins},
	}

	response, err := rpcConnection.SendMsg(ctx, msgSend, input.Memo)
	if err != nil {
		return SendResponse{}, err
	}

	return SendResponse{
		Height:    response.Height,
		TxHash:    response.TxHash,
		Data:      response.Data,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
		RawLog:    response.RawLog,
	}, nil
}
