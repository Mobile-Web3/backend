package http

import (
	"context"
	"net/http"

	_ "github.com/Mobile-Web3/backend/docs/api"
	"github.com/Mobile-Web3/backend/internal/domain/account"
	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/internal/domain/transaction"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/gin-gonic/gin"
	swagger "github.com/swaggo/http-swagger"
)

var logger log.Logger

type requestHandler[TRequest any, TResponse any] func(ctx context.Context, request TRequest) (TResponse, error)
type emptyHandler[TResponse any] func(ctx context.Context) (TResponse, error)

func newRequestHandler[TRequest any, TResponse any](handler requestHandler[TRequest, TResponse]) func(context *gin.Context) {
	return func(context *gin.Context) {
		var request TRequest
		if err := context.BindJSON(&request); err != nil {
			context.JSON(http.StatusOK, newErrorResponse(err.Error()))
			return
		}

		response, err := handler(context.Request.Context(), request)
		if err != nil {
			logger.Error(err)
			context.JSON(http.StatusOK, newErrorResponse(err.Error()))
			return
		}

		context.JSON(http.StatusOK, newSuccessResponse(response))
	}
}

func newEmptyHandler[TResponse any](handler emptyHandler[TResponse]) func(context *gin.Context) {
	return func(context *gin.Context) {
		response, err := handler(context.Request.Context())
		if err != nil {
			logger.Error(err)
			context.JSON(http.StatusOK, newErrorResponse(err.Error()))
			return
		}

		context.JSON(http.StatusOK, newSuccessResponse(response))
	}
}

type Dependencies struct {
	Logger             log.Logger
	Repository         chain.Repository
	AccountService     *account.Service
	TransactionService *transaction.Service
}

func New(dependencies *Dependencies) http.Handler {
	accountService := dependencies.AccountService
	transactionService := dependencies.TransactionService
	chainRepository := dependencies.Repository
	logger = dependencies.Logger

	gin.SetMode("release")
	router := gin.New()
	router.Use(recoverMiddleware(logger))
	swaggerHandler := gin.WrapH(swagger.Handler(swagger.URL("doc.json")))
	router.GET("/api/swagger/*any", swaggerHandler)

	api := router.Group("/api")
	{
		accounts := api.Group("account")
		{
			accounts.POST("mnemonic", newRequestHandler(accountService.CreateMnemonic))
			accounts.POST("create", newRequestHandler(accountService.CreateAccount))
			accounts.POST("restore", newRequestHandler(accountService.RestoreAccount))
			accounts.POST("balance", newRequestHandler(accountService.CheckBalance))
		}

		chains := api.Group("chains")
		{
			chains.POST("all", newEmptyHandler(chainRepository.GetAllChains))
		}

		transactions := api.Group("transaction")
		{
			transactions.POST("send", newRequestHandler(transactionService.SendTransaction))
			transactions.POST("simulate", newRequestHandler(transactionService.SimulateTransaction))
		}
	}

	return router
}
