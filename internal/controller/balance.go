package controller

import (
	"net/http"

	"github.com/Mobile-Web3/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type BalanceController struct {
	service *service.BalanceService
}

func NewBalanceController(service *service.BalanceService) *BalanceController {
	return &BalanceController{
		service: service,
	}
}

type checkBalanceRequest struct {
	WalletAddress string `json:"walletAddress"`
}

// GetBalance godoc
// @Summary      Получить инфу о балансе
// @Tags         balance
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body checkBalanceRequest true "body"
// @Success      200 {object} apiResponse{result=service.BalanceResponse}
// @Router       /balance/check [post]
func (c *BalanceController) GetBalance(context *gin.Context) {
	body := checkBalanceRequest{}
	if err := context.BindJSON(&body); err != nil {
		context.JSON(http.StatusOK, createErrorResponse(err.Error()))
		return
	}

	if body.WalletAddress == "" {
		context.JSON(http.StatusOK, createErrorResponse("invalid wallet address"))
		return
	}

	response, err := c.service.GetBalance(context.Request.Context(), body.WalletAddress)
	if err != nil {
		context.JSON(http.StatusOK, createErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, createSuccessResponse(response))
}
