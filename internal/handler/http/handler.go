package http

import (
	"net/http"

	_ "github.com/Mobile-Web3/backend/docs/api"
	"github.com/Mobile-Web3/backend/internal/domain/account"
	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/internal/domain/transaction"
	v1 "github.com/Mobile-Web3/backend/internal/handler/http/v1"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/gin-gonic/gin"
	swagger "github.com/swaggo/http-swagger"
)

type Dependencies struct {
	Logger             log.Logger
	Repository         chain.Repository
	ChainService       *chain.Service
	AccountService     *account.Service
	TransactionService *transaction.Service
}

func NewHandler(dependencies *Dependencies) http.Handler {
	accountsController := v1.NewAccountsController(dependencies.Logger, dependencies.AccountService)
	chainsController := v1.NewChainsController(dependencies.Logger, dependencies.Repository, dependencies.ChainService)
	transactionsController := v1.NewTransactionsController(dependencies.Logger, dependencies.TransactionService)

	gin.SetMode("release")
	router := gin.New()
	router.Use(recoverMiddleware(dependencies.Logger))

	swaggerHandler := gin.WrapH(swagger.Handler(swagger.URL("doc.json")))
	router.GET("/api/swagger/*any", swaggerHandler)
	router.GET("/api/metrics", v1.MetricsHandler)

	api := router.Group("/api/v1", metricsMiddleware)
	{
		accounts := api.Group("accounts")
		{
			accounts.POST("mnemonic", accountsController.CreateMnemonic())
			accounts.POST("create", accountsController.CreateAccount())
			accounts.POST("restore", accountsController.RestoreAccount())
			accounts.GET("balance", accountsController.GetBalance)
		}

		chains := api.Group("chains")
		{
			chains.GET("", chainsController.GetChains())
			chains.GET(":id/validators", chainsController.GetPagedValidators)
		}

		transactions := api.Group("transactions")
		{
			transactions.POST("send", transactionsController.SendTransaction())
			transactions.POST("send/firebase", transactionsController.SendTransactionFirebase())
			transactions.POST("simulate", transactionsController.SimulateTransaction())
		}
	}

	return router
}
