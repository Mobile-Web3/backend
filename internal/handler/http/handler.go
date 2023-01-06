package http

import (
	"context"
	"net/http"

	_ "github.com/Mobile-Web3/backend/docs/api"
	"github.com/Mobile-Web3/backend/internal/domain/account"
	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/internal/domain/transaction"
	"github.com/gin-gonic/gin"
	swagger "github.com/swaggo/http-swagger"
)

type requestHandler[TRequest any, TResponse any] func(ctx context.Context, request TRequest) (TResponse, error)

func newRequestHandler[TRequest any, TResponse any](handler requestHandler[TRequest, TResponse]) func(context *gin.Context) {
	return func(context *gin.Context) {
		var request TRequest
		if err := context.BindJSON(&request); err != nil {
			context.JSON(http.StatusOK, newErrorResponse(err.Error()))
			return
		}

		response, err := handler(context.Request.Context(), request)
		if err != nil {
			context.JSON(http.StatusOK, newErrorResponse(err.Error()))
			return
		}

		context.JSON(http.StatusOK, newSuccessResponse(response))
	}
}

type emptyHandler[TResponse any] func(ctx context.Context) (TResponse, error)

func newEmptyHandler[TResponse any](handler emptyHandler[TResponse]) func(context *gin.Context) {
	return func(context *gin.Context) {
		response, err := handler(context.Request.Context())
		if err != nil {
			context.JSON(http.StatusOK, newErrorResponse(err.Error()))
			return
		}

		context.JSON(http.StatusOK, newSuccessResponse(response))
	}
}

func New(chainRepository chain.Repository, accountService *account.Service, transactionService *transaction.Service) http.Handler {
	gin.SetMode("release")
	router := gin.New()

	router.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		c.JSON(http.StatusOK, newErrorResponse("internal error"))
	}))

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
