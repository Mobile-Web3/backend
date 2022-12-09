package balance

import (
	"net/http"

	"github.com/Mobile-Web3/backend/pkg/api"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	logger  log.Logger
	service *Service
}

func NewController(logger log.Logger, service *Service) *Controller {
	return &Controller{
		logger:  logger,
		service: service,
	}
}

type checkRequest struct {
	WalletAddress string `json:"walletAddress"`
}

// GetBalance godoc
// @Summary      Получить инфу о балансе
// @Tags         balance
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body checkRequest true "body"
// @Success      200 {object} api.Response{result=balance.CheckResponse}
// @Router       /balance/check [post]
func (c *Controller) GetBalance(context *gin.Context) {
	body := checkRequest{}
	if err := context.BindJSON(&body); err != nil {
		c.logger.Error(err)
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	if body.WalletAddress == "" {
		context.JSON(http.StatusOK, api.NewErrorResponse("invalid wallet address"))
		return
	}

	response, err := c.service.GetBalance(context.Request.Context(), body.WalletAddress)
	if err != nil {
		c.logger.Error(err)
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, api.NewSuccessResponse(response))
}
