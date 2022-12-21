package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/Mobile-Web3/backend/docs/api"
	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/internal/infrastructure/cosmos"
	response "github.com/Mobile-Web3/backend/pkg/api"
	"github.com/Mobile-Web3/backend/pkg/env"
	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"
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

	url := os.Getenv("CHAIN_REGISTRY_URL")
	if url == "" {
		errorLogger.Println("empty CHAIN_REGISTRY_URL env")
		return
	}

	registryDir := os.Getenv("REGISTRY_DIR")
	if registryDir == "" {
		errorLogger.Println("empty REGISTRY_DIR env")
		return
	}

	_, _, ok := strings.Cut(url, "http://")
	if !ok {
		_, _, ok = strings.Cut(url, "https://")
		if !ok {
			errorLogger.Println("invalid git url")
			return
		}
	}

	_, _, ok = strings.Cut(url, ".git")
	if !ok {
		errorLogger.Println("invalid git url")
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

	chainRegistry, err := cosmos.NewChainRegistry(url, registryDir)
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

	chainClientFactory := cosmos.NewClientFactory(rpcLifetime, errorLogger)
	chainService := chain.NewService(gasAdjustment, chainRegistry, chainClientFactory)
	controller := NewController(chainRegistry, chainService)

	gin.SetMode("release")
	router := gin.New()

	router.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		c.JSON(http.StatusOK, response.NewErrorResponse("internal error"))
	}))

	swaggerHandler := gin.WrapH(httpSwagger.Handler(httpSwagger.URL("doc.json")))
	router.GET("/api/swagger/*any", swaggerHandler)

	api := router.Group("/api")
	{
		api.POST("/balance/check", controller.CheckBalance)
		api.POST("/transaction/send", controller.SendTransaction)
		api.POST("/transaction/simulate", controller.SimulateTransaction)
		api.POST("/chains/all", controller.GetAllChains)
	}

	port := os.Getenv("PORT")
	if port == "" {
		errorLogger.Println("empty PORT env")
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

		infoLogger.Println("server shutting down")
		if err = server.Shutdown(context.Background()); err != nil {
			errorLogger.Println(err)
		}
	}()

	infoLogger.Println("server started")
	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errorLogger.Println(err)
	}
}
