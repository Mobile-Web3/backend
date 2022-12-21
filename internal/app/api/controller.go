package api

import (
	"net/http"

	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/pkg/api"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	chainRepository chain.Repository
	chainService    *chain.Service
}

func NewController(chainRepository chain.Repository, chainService *chain.Service) *Controller {
	return &Controller{
		chainRepository: chainRepository,
		chainService:    chainService,
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
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	response, err := c.chainService.CheckBalance(context.Request.Context(), request.WalletAddress)
	if err != nil {
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, api.NewSuccessResponse(response))
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
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	response, err := c.chainService.SendTransaction(context.Request.Context(), request)
	if err != nil {
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, api.NewSuccessResponse(response))
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
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	response, err := c.chainService.SimulateTransaction(context.Request.Context(), request)
	if err != nil {
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, api.NewSuccessResponse(response))
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
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, api.NewSuccessResponse(chains))
}
