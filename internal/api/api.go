package api

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "github.com/Mobile-Web3/backend/docs/api"
	"github.com/Mobile-Web3/backend/internal/db/memory"
	"github.com/Mobile-Web3/backend/internal/domain/account"
	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/internal/domain/transaction"
	"github.com/Mobile-Web3/backend/internal/firebase"
	"github.com/Mobile-Web3/backend/internal/github"
	httphandler "github.com/Mobile-Web3/backend/internal/handler/http"
	"github.com/Mobile-Web3/backend/internal/server/http"
	"github.com/Mobile-Web3/backend/pkg/cosmos"
	"github.com/Mobile-Web3/backend/pkg/env"
	"github.com/Mobile-Web3/backend/pkg/log"
)

var (
	errEmptyGasAdjustment  = errors.New("empty GAS_ADJUSTMENT env")
	errEmptyPort           = errors.New("empty PORT env")
	errFirebaseEmptyConfig = errors.New("empty FIREBASE_KEY_PATH env")
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

	chainRepository := memory.NewChainRepository()
	chainRegistryClient := github.NewChainRegistryClient(logger)
	chainService := chain.NewService(chainRegistryClient, chainRepository)
	if err = chainService.UpdateChainInfo(context.Background()); err != nil {
		logger.Error(err)
		return
	}

	firebaseConfigPath := os.Getenv("FIREBASE_KEY_PATH")
	if firebaseConfigPath == "" {
		logger.Error(errFirebaseEmptyConfig)
		return
	}

	firebaseCloudMessaging, err := firebase.NewCloudMessagingClient(firebaseConfigPath, logger)
	if err != nil {
		logger.Error(err)
		return
	}

	cosmosClient, err := cosmos.NewClient("direct", logger, firebaseCloudMessaging.SendTxResult, chainRepository.GetRPCEndpoints)
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

	accounts := account.NewService(logger, chainRepository, cosmosClient)
	transactions := transaction.NewService(gasAdjustment, logger, chainRepository, cosmosClient)
	handler, err := httphandler.NewHandler(&httphandler.Dependencies{
		Logger:             logger,
		Repository:         chainRepository,
		AccountService:     accounts,
		TransactionService: transactions,
	})
	if err != nil {
		logger.Error(err)
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		logger.Error(errEmptyPort)
		return
	}

	worker := NewWorker(logger, chainService)
	if err = worker.Start(); err != nil {
		logger.Error(err)
		return
	}

	server := http.New(port, logger, handler)
	go server.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	server.Stop()
	worker.Stop()
}
