package api

import (
	"net/http"

	"github.com/Mobile-Web3/backend/internal/chain"
	"github.com/Mobile-Web3/backend/pkg/cosmos/client"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	chainRepository chain.Repository
	chainService    *chain.Service
	cosmosClient    *client.Client
}

func NewController(chainRepository chain.Repository, chainService *chain.Service, cosmosClient *client.Client) *Controller {
	return &Controller{
		chainRepository: chainRepository,
		chainService:    chainService,
		cosmosClient:    cosmosClient,
	}
}

type checkRequest struct {
	WalletAddress string `json:"walletAddress"`
}

// CheckBalance godoc
// @Summary      Получить инфу о балансе
// @Tags         balance
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body checkRequest true "body"
// @Success      200 {object} api.Response{result=chain.CheckResponse}
// @Router       /balance/check [post]
func (c *Controller) CheckBalance(context *gin.Context) {
	request := checkRequest{}
	if err := context.BindJSON(&request); err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	response, err := c.chainService.CheckBalance(context.Request.Context(), request.WalletAddress)
	if err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, newSuccessResponse(response))
}

// SendTransaction godoc
// @Summary      Отправить транзакцию
// @Tags         transaction
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body chain.SendTxInput true "body"
// @Success      200 {object} api.Response{result=chain.SendTxResponse}
// @Router       /transaction/send [post]
func (c *Controller) SendTransaction(context *gin.Context) {
	request := chain.SendTxInput{}
	if err := context.BindJSON(&request); err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	response, err := c.chainService.SendTransaction(context.Request.Context(), request)
	if err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, newSuccessResponse(response))
}

// SimulateTransaction godoc
// @Summary      Симуляция транзакции для расчета параметров
// @Tags         transaction
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body chain.SimulateTxInput true "body"
// @Success      200 {object} api.Response{result=chain.SimulateTxResponse}
// @Router       /transaction/simulate [post]
func (c *Controller) SimulateTransaction(context *gin.Context) {
	request := chain.SimulateTxInput{}
	if err := context.BindJSON(&request); err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	response, err := c.chainService.SimulateTransaction(context.Request.Context(), request)
	if err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, newSuccessResponse(response))
}

// GetAllChains godoc
// @Summary      Получение данных о сетях
// @Tags         chain
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @Success      200 {object} api.Response{result=[]chain.ShortResponse}
// @Router       /chains/all [post]
func (c *Controller) GetAllChains(context *gin.Context) {
	chains, err := c.chainRepository.GetAllChains(context.Request.Context())
	if err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, newSuccessResponse(chains))
}

type createMnemonicInput struct {
	MnemonicSize uint8 `json:"mnemonicSize"`
}

// CreateMnemonic godoc
// @Summary      Создание мнемоника
// @Tags         account
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body createMnemonicInput true "body"
// @Success      200 {object} api.Response{result=string}
// @Router       /account/mnemonic [post]
func (c *Controller) CreateMnemonic(context *gin.Context) {
	request := createMnemonicInput{}
	if err := context.BindJSON(&request); err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	mnemonic, err := c.cosmosClient.CreateMnemonic(request.MnemonicSize)
	if err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, newSuccessResponse(mnemonic))
}

// CreateAccount godoc
// @Summary      Получение аккаунта по мнемонику
// @Tags         account
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body chain.CreateAccountInput true "body"
// @Success      200 {object} api.Response{result=chain.AccountResponse}
// @Router       /account/create [post]
func (c *Controller) CreateAccount(context *gin.Context) {
	request := chain.CreateAccountInput{}
	if err := context.BindJSON(&request); err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	response, err := c.chainService.CreateAccount(context.Request.Context(), request)
	if err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, newSuccessResponse(response))
}

// RestoreAccount godoc
// @Summary      Получение аккаунта по ключу
// @Tags         account
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body chain.RestoreAccountInput true "body"
// @Success      200 {object} api.Response{result=chain.AccountResponse}
// @Router       /account/restore [post]
func (c *Controller) RestoreAccount(context *gin.Context) {
	request := chain.RestoreAccountInput{}
	if err := context.BindJSON(&request); err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	response, err := c.chainService.RestoreAccount(context.Request.Context(), request)
	if err != nil {
		context.JSON(http.StatusOK, newErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, newSuccessResponse(response))
}
