package http

import (
	"context"
	"fmt"
	"net/http"

	_ "github.com/Mobile-Web3/backend/docs/api"
	"github.com/Mobile-Web3/backend/internal/domain/account"
	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/internal/domain/transaction"
	"github.com/Mobile-Web3/backend/internal/metrics"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/gin-gonic/gin"
	swagger "github.com/swaggo/http-swagger"
)

type handler[TResponse any] func(ctx context.Context) (TResponse, error)
type requestHandler[TRequest any, TResponse any] func(ctx context.Context, request TRequest) (TResponse, error)

func newHandler[TResponse any](handler handler[TResponse]) gin.HandlerFunc {
	return func(context *gin.Context) {
		response, err := handler(context.Request.Context())
		if err != nil {
			metrics.ErrorsCounter.Incr(1)
			context.JSON(http.StatusOK, newErrorResponse(err.Error()))
			return
		}

		context.JSON(http.StatusOK, newSuccessResponse(response))
	}
}

func newRequestHandler[TRequest any, TResponse any](handler requestHandler[TRequest, TResponse], logger log.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		var request TRequest
		if err := context.BindJSON(&request); err != nil {
			logger.Error(err)
			metrics.ErrorsCounter.Incr(1)
			context.JSON(http.StatusOK, newErrorResponse(err.Error()))
			return
		}

		response, err := handler(context.Request.Context(), request)
		if err != nil {
			metrics.ErrorsCounter.Incr(1)
			context.JSON(http.StatusOK, newErrorResponse(err.Error()))
			return
		}

		context.JSON(http.StatusOK, newSuccessResponse(response))
	}
}

func metricsHandler(context *gin.Context) {
	result := fmt.Sprintf(`<h3>Metrics</h3>
<div>average requests   per second for the last hour: %d</div>
<div>average errors     per second for the last hour: %d</div>
<div>average exceptions per second for the last hour: %d</div>`,
		metrics.RpsCounter.Rate()/60/60, metrics.ErrorsCounter.Rate()/60/60, metrics.PanicsCounter.Rate()/60/60)
	context.Writer.WriteHeader(http.StatusOK)
	context.Writer.Header().Set("Content-Type", "text/html")
	_, err := context.Writer.Write([]byte(result))
	if err != nil {
		context.String(http.StatusInternalServerError, err.Error())
		return
	}
}

type Dependencies struct {
	Logger             log.Logger
	Repository         chain.Repository
	AccountService     *account.Service
	TransactionService *transaction.Service
}

func NewHandler(dependencies *Dependencies) (http.Handler, error) {
	accountService := dependencies.AccountService
	transactionService := dependencies.TransactionService
	chainRepository := dependencies.Repository
	logger := dependencies.Logger

	gin.SetMode("release")
	router := gin.New()
	router.Use(recoverMiddleware(logger))

	swaggerHandler := gin.WrapH(swagger.Handler(swagger.URL("doc.json")))
	router.GET("/api/swagger/*any", swaggerHandler)
	router.GET("/api/metrics", metricsHandler)

	api := router.Group("/api", metricsMiddleware)
	{
		accounts := api.Group("account")
		{
			accounts.POST("mnemonic", newRequestHandler(accountService.CreateMnemonic, logger))
			accounts.POST("create", newRequestHandler(accountService.CreateAccount, logger))
			accounts.POST("restore", newRequestHandler(accountService.RestoreAccount, logger))
			accounts.POST("balance", newRequestHandler(accountService.CheckBalance, logger))
		}

		chains := api.Group("chains")
		{
			chains.POST("all", newHandler(chainRepository.GetAllChains))
		}

		transactions := api.Group("transaction")
		{
			transactions.POST("send", newRequestHandler(transactionService.SendTransaction, logger))
			transactions.POST("simulate", newRequestHandler(transactionService.SimulateTransaction, logger))
		}
	}

	return router, nil
}
