package api

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/Mobile-Web3/backend/docs/api"
	"github.com/Mobile-Web3/backend/internal/db/memory"
	"github.com/Mobile-Web3/backend/internal/domain/account"
	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/internal/domain/transaction"
	httphandler "github.com/Mobile-Web3/backend/internal/handler/http"
	"github.com/Mobile-Web3/backend/internal/server/http"
	"github.com/Mobile-Web3/backend/pkg/cosmos/client"
	"github.com/Mobile-Web3/backend/pkg/env"
	"github.com/Mobile-Web3/backend/pkg/log"
)

var (
	errEmptyGasAdjustment = errors.New("empty GAS_ADJUSTMENT env")
	errEmptyPort          = errors.New("empty PORT env")
)

// @title           Swagger UI
// @version         1.0
// @description     API
// @BasePath  /api

func Run() {
	logger := log.NewFmt("02.01.2006 15:04:05")

	err := env.Parse()
	if err != nil {
		logger.Error(err)
		return
	}

	chainRPCLifetimeENV := os.Getenv("CHAIN_RPC_LIFETIME")
	if chainRPCLifetimeENV == "" {
		chainRPCLifetimeENV = "10m"
	}
	rpcLifetime, err := time.ParseDuration(chainRPCLifetimeENV)
	if err != nil {
		logger.Error(err)
		return
	}

	chainRepository := memory.NewChainRepository()
	chainRegistry := chain.NewRegistry(chainRepository)
	if err = chainRegistry.UploadChainInfo(context.Background()); err != nil {
		logger.Error(err)
		return
	}

	cosmosClient, err := client.NewClient("direct", rpcLifetime, chainRepository.GetRPCEndpoints)
	if err != nil {
		logger.Error(err)
		return
	}

	gasAdjustmentStr := os.Getenv("GAS_ADJUSTMENT")
	if gasAdjustmentStr == "" {
		logger.Error(errEmptyGasAdjustment)
		return
	}
	gasAdjustment, err := strconv.ParseFloat(gasAdjustmentStr, 64)
	if err != nil {
		logger.Error(err)
		return
	}

	accounts := account.NewService(chainRepository, cosmosClient)
	transactions := transaction.NewService(gasAdjustment, chainRepository, cosmosClient)
	handler := httphandler.New(&httphandler.Dependencies{
		Logger:             logger,
		Repository:         chainRepository,
		AccountService:     accounts,
		TransactionService: transactions,
	})

	port := os.Getenv("PORT")
	if port == "" {
		logger.Error(errEmptyPort)
		return
	}

	worker := NewWorker(time.Hour*12, logger, chainRegistry)
	worker.Start()
	server := http.New(port, logger, handler)
	go server.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	server.Stop()
	worker.Stop()
}
