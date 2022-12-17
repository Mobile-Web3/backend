package chain

import (
	"net/http"

	"github.com/Mobile-Web3/backend/pkg/api"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
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
	chains := c.service.GetChains()
	context.JSON(http.StatusOK, api.NewSuccessResponse(chains))
}
