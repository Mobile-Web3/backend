package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/Mobile-Web3/backend/cmd/gateway/docs"
	"github.com/Mobile-Web3/backend/internal/controller"
	"github.com/Mobile-Web3/backend/internal/service"
	"github.com/Mobile-Web3/backend/pkg/cosmos"
	"github.com/Mobile-Web3/backend/pkg/env"
	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Swagger UI
// @version         1.0
// @description     API
// @BasePath  /api
func main() {
	ctx := context.Background()
	logger := log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	if err := env.Parse(); err != nil {
		logger.Fatal(err)
	}

	chainClient, err := cosmos.NewChainClient()
	if err != nil {
		logger.Fatal(err)
	}

	balanceService := service.NewBalanceService(chainClient)
	balanceController := controller.NewBalanceController(balanceService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	swaggerHandler := gin.WrapH(httpSwagger.Handler(httpSwagger.URL("doc.json")))
	router.GET("/swagger/*any", swaggerHandler)

	api := router.Group("/api")
	{
		api.POST("/balance/check", balanceController.GetBalance)
	}

	port := os.Getenv("PORT")
	if port == "" {
		logger.Fatal("empty PORT")
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
			logger.Fatal(err)
		}
	}()

	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal()
	}
}
