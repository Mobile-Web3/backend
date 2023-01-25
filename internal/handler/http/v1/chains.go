package v1

import (
	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/gin-gonic/gin"
)

type ChainsController struct {
	logger     log.Logger
	repository chain.Repository
}

func NewChainsController(logger log.Logger, repository chain.Repository) *ChainsController {
	return &ChainsController{
		logger:     logger,
		repository: repository,
	}
}

// GetChains godoc
// @Summary      Получение данных о сетях
// @Tags         chains
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @Success      200 {object} apiResponse{result=[]chain.ShortResponse}
// @Router       /v1/chains [get]
func (c *ChainsController) GetChains() gin.HandlerFunc {
	return newEmptyHandler(c.repository.GetAllChains)
}
