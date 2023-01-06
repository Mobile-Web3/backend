package http

import (
	"context"
	"net/http"

	_ "github.com/Mobile-Web3/backend/docs/api"
	"github.com/Mobile-Web3/backend/internal/chain"
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

func New(repository chain.Repository, service *chain.Service) http.Handler {
	gin.SetMode("release")
	router := gin.New()

	router.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		c.JSON(http.StatusOK, newErrorResponse("internal error"))
	}))

	swaggerHandler := gin.WrapH(swagger.Handler(swagger.URL("doc.json")))
	router.GET("/api/swagger/*any", swaggerHandler)

	api := router.Group("/api")
	{
		account := api.Group("account")
		{
			account.POST("mnemonic", newRequestHandler(service.CreateMnemonic))
			account.POST("create", newRequestHandler(service.CreateAccount))
			account.POST("restore", newRequestHandler(service.RestoreAccount))
			account.POST("balance", newRequestHandler(service.CheckBalance))
		}

		chains := api.Group("chains")
		{
			chains.POST("all", newEmptyHandler(repository.GetAllChains))
		}

		transaction := api.Group("transaction")
		{
			transaction.POST("send", newRequestHandler(service.SendTransaction))
			transaction.POST("simulate", newRequestHandler(service.SimulateTransaction))
		}
	}

	return router
}
