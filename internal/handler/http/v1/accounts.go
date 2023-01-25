package v1

import (
	"github.com/Mobile-Web3/backend/internal/domain/account"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/gin-gonic/gin"
)

type AccountsController struct {
	logger  log.Logger
	service *account.Service
}

func NewAccountsController(logger log.Logger, service *account.Service) *AccountsController {
	return &AccountsController{
		logger:  logger,
		service: service,
	}
}

// CreateMnemonic godoc
// @Summary      Создание мнемоника
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body account.CreateMnemonicInput true "body"
// @Success      200 {object} apiResponse{result=string}
// @Router       /v1/accounts/mnemonic [post]
func (c *AccountsController) CreateMnemonic() gin.HandlerFunc {
	return newRequestHandler(c.service.CreateMnemonic, c.logger)
}

// CreateAccount godoc
// @Summary      Получение аккаунта по мнемонику
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body account.CreateAccountInput true "body"
// @Success      200 {object} apiResponse{result=account.KeyResponse}
// @Router       /v1/accounts/create [post]
func (c *AccountsController) CreateAccount() gin.HandlerFunc {
	return newRequestHandler(c.service.CreateAccount, c.logger)
}

// RestoreAccount godoc
// @Summary      Получение аккаунта по ключу
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body account.RestoreAccountInput true "body"
// @Success      200 {object} apiResponse{result=account.KeyResponse}
// @Router       /v1/accounts/restore [post]
func (c *AccountsController) RestoreAccount() gin.HandlerFunc {
	return newRequestHandler(c.service.RestoreAccount, c.logger)
}

// GetBalance godoc
// @Summary      Получить инфу о балансе
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @Param        chainId query string  true "id сети"
// @Param        address query string  true "адрес кошелька"
// @Success      200 {object} apiResponse{result=account.BalanceResponse}
// @Router       /v1/accounts/balance [get]
func (c *AccountsController) GetBalance(context *gin.Context) {
	request := account.BalanceInput{
		ChainID: context.Query("chainId"),
		Address: context.Query("address"),
	}

	handleRequest(request, context, c.service.CheckBalance)
}
