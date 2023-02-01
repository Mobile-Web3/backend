package http

import (
	"fmt"
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

	router.GET("/api/page", func(context *gin.Context) {
		dependencies.Logger.Info(fmt.Sprintf("Referer: %s", context.Request.Header.Get("Referer")))
		context.Writer.WriteHeader(http.StatusOK)
		context.Writer.Header().Set("Content-Type", "text/html")

		if _, err := context.Writer.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>visits test</title>
</head>
<body>
<main>
    <h1>Визиты</h1>
    <p>
        Каждый раз, когда на одну из страниц сайта с подключённой Метрикой заходит посетитель,
        серверы Метрики получают информацию об этом событии — оно называется просмотром.
        Несколько просмотров, совершённых одним пользователем,
        Метрика объединяет в один визит.
        Именно визиты являются основными данными для построения отчётов в Метрике
    </p>
    <p>
        Визит начинается с перехода на сайт из какого-либо внешнего источника.
        Это может быть ссылка на стороннем сайте, рекламное объявление, поисковая система или социальная сеть.
        Пользователи также могут попадать на ваш сайт,
        открыв закладку или напечатав URL в адресной строке браузера — такие переходы называются прямыми.
    </p>
    <p>
        Визит заканчивается, когда в течение определённого времени от посетителя не поступает новых событий — по умолчанию это 30 минут.
        Если в течение 30 минут от последнего события тот же пользователь снова перейдет на сайт из внешнего источника,
        Метрика не будет считать такое посещение новым визитом — все новые просмотры страниц будут добавлены к предыдущему визиту.
    </p>
    <div>
        <a href="http://localhost:8080">link</a>
    </div>
</main>
</body>
</html>`)); err != nil {
			context.Writer.WriteHeader(http.StatusInternalServerError)
		}
	})

	return router
}
