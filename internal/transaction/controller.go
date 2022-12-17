package transaction

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

// Send godoc
// @Summary      Отправить транзакцию
// @Tags         transaction
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body transaction.SendInput true "body"
// @Success      200 {object} api.Response{result=transaction.SendResponse}
// @Router       /transaction/send [post]
func (c *Controller) Send(context *gin.Context) {
	input := SendInput{}
	if err := context.BindJSON(&input); err != nil {
		c.logger.Error(err)
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	response, err := c.service.Send(context.Request.Context(), input)
	if err != nil {
		c.logger.Error(err)
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, api.NewSuccessResponse(response))
}

// Simulate godoc
// @Summary      Симуляция транзакции для расчета параметров
// @Tags         transaction
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body transaction.SimulateInput true "body"
// @Success      200 {object} api.Response{result=transaction.SimulateResponse}
// @Router       /transaction/simulate [post]
func (c *Controller) Simulate(context *gin.Context) {
	input := SimulateInput{}
	if err := context.BindJSON(&input); err != nil {
		c.logger.Error(err)
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	response, err := c.service.Simulate(context.Request.Context(), input)
	if err != nil {
		c.logger.Error(err)
		context.JSON(http.StatusOK, api.NewErrorResponse(err.Error()))
		return
	}

	context.JSON(http.StatusOK, api.NewSuccessResponse(response))
}
