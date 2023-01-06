package api

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/Mobile-Web3/backend/docs/api"
	"github.com/Mobile-Web3/backend/internal/cosmos"
	"github.com/Mobile-Web3/backend/internal/domain/account"
	"github.com/Mobile-Web3/backend/internal/domain/transaction"
	httphandler "github.com/Mobile-Web3/backend/internal/handler/http"
	"github.com/Mobile-Web3/backend/internal/server/http"
	"github.com/Mobile-Web3/backend/pkg/cosmos/client"
	"github.com/Mobile-Web3/backend/pkg/env"
)

// @title           Swagger UI
// @version         1.0
// @description     API
// @BasePath  /api

func Run() {
	infoLogger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errorLogger := log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	err := env.Parse()
	if err != nil {
		errorLogger.Println(err)
		return
	}

	registryURL := os.Getenv("CHAIN_REGISTRY_URL")
	if registryURL == "" {
		errorLogger.Println("empty CHAIN_REGISTRY_URL env")
		return
	}

	registryDir := os.Getenv("REGISTRY_DIR")
	if registryDir == "" {
		errorLogger.Println("empty REGISTRY_DIR env")
		return
	}

	gasAdjustmentStr := os.Getenv("GAS_ADJUSTMENT")
	if gasAdjustmentStr == "" {
		errorLogger.Println("empty GAS_ADJUSTMENT env")
		return
	}
	gasAdjustment, err := strconv.ParseFloat(gasAdjustmentStr, 64)
	if err != nil {
		errorLogger.Println(err)
		return
	}

	chainRepository := cosmos.NewChainRepository()
	chainRegistry, err := cosmos.NewChainRegistry(registryURL, registryDir, chainRepository)
	if err != nil {
		errorLogger.Println(err)
		return
	}

	err = chainRegistry.UploadChainInfo(context.Background())
	if err != nil {
		errorLogger.Println(err)
		return
	}

	chainRPCLifetimeENV := os.Getenv("CHAIN_RPC_LIFETIME")
	if chainRPCLifetimeENV == "" {
		chainRPCLifetimeENV = "10m"
	}
	rpcLifetime, err := time.ParseDuration(chainRPCLifetimeENV)
	if err != nil {
		errorLogger.Println(err)
		return
	}

	cosmosClient, err := client.NewClient("direct", rpcLifetime, chainRepository.GetRPCEndpoints)
	if err != nil {
		errorLogger.Println(err)
		return
	}

	accounts := account.NewService(chainRepository, cosmosClient)
	transactions := transaction.NewService(gasAdjustment, chainRepository, cosmosClient)
	handler := httphandler.New(chainRepository, accounts, transactions)

	port := os.Getenv("PORT")
	if port == "" {
		errorLogger.Println("empty PORT env")
		return
	}

	server := http.New(port, handler, infoLogger, errorLogger)
	go server.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	server.Stop()
}
