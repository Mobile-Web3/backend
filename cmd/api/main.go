package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/Mobile-Web3/backend/docs/api"
	"github.com/Mobile-Web3/backend/internal/balance"
	"github.com/Mobile-Web3/backend/internal/chain"
	"github.com/Mobile-Web3/backend/internal/transaction"
	"github.com/Mobile-Web3/backend/pkg/cosmos"
	"github.com/Mobile-Web3/backend/pkg/env"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Swagger UI
// @version         1.0
// @description     API
// @BasePath  /api
func main() {
	ctx := context.Background()
	logger := log.NewStandard()

	if err := env.Parse(); err != nil {
		logger.Error(err)
		return
	}

	chainClient, err := cosmos.NewChainClient()
	if err != nil {
		logger.Error(err)
		return
	}

	balanceController := balance.NewController(logger, balance.NewService(chainClient))
	transactionController := transaction.NewController(logger, transaction.NewService(chainClient))
	chainsController := chain.NewController(chain.NewService(chainClient))

	gin.SetMode("release")
	router := gin.New()
	router.Use(gin.Recovery())
	swaggerHandler := gin.WrapH(httpSwagger.Handler(httpSwagger.URL("doc.json")))
	router.GET("/api/swagger/*any", swaggerHandler)

	api := router.Group("/api")
	{
		api.POST("/balance/check", balanceController.GetBalance)
		api.POST("/transaction/send", transactionController.Send)
		api.POST("/transaction/simulate", transactionController.Simulate)
		api.POST("/chains/all", chainsController.GetAllChains)
	}

	port := os.Getenv("PORT")
	if port == "" {
		logger.Error("empty PORT")
		return
	}
	server := http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		if err = server.Shutdown(ctx); err != nil {
			logger.Error(err)
		}
	}()

	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error(err)
	}
}
