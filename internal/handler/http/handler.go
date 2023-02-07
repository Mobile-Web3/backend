package http

import (
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/Mobile-Web3/backend/docs/api"
	"github.com/Mobile-Web3/backend/internal/domain/account"
	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/internal/domain/transaction"
	v1 "github.com/Mobile-Web3/backend/internal/handler/http/v1"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	router.GET("/api/test", func(context *gin.Context) {
		context.Writer.WriteHeader(http.StatusOK)
		context.Writer.Header().Set("Content-Type", "text/html")
		_, _ = context.Writer.Write([]byte(test))
	})
	router.GET("/api/page", func(context *gin.Context) {
		query := context.Request.URL.Query()
		userId := query.Get("user_id")
		if userId == "" {
			userId = uuid.New().String()
			query.Set("user_id", userId)
			fmt.Println("visit")
		}

		data := pageData{
			UserID: userId,
		}
		tmpl, _ := template.ParseFiles("web/index.html")
		err := tmpl.Execute(context.Writer, data)
		if err != nil {
			fmt.Println(err)
		}
	})
	return router
}

type pageData struct {
	UserID string
}

const test = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>visits test</title>
</head>
<body>
<main>
    <div>
        <a href="https://mobileweb3.tech/api/page">link</a>
    </div>
</main>
</body>
</html>`
