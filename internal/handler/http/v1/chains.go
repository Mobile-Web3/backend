package v1

import (
	"strconv"

	"github.com/Mobile-Web3/backend/internal/domain/chain"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/gin-gonic/gin"
)

type ChainsController struct {
	logger     log.Logger
	repository chain.Repository
	service    *chain.Service
}

func NewChainsController(logger log.Logger, repository chain.Repository, service *chain.Service) *ChainsController {
	return &ChainsController{
		logger:     logger,
		repository: repository,
		service:    service,
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

// GetPagedValidators godoc
// @Summary      Получение данных о валидаторах
// @Tags         chains
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @Param        id     path  string true  "chainId"
// @Param        limit  query int    false "кол-во валидаторов для запроса"
// @Param        offset query int    false "кол-во валидаторов для пропуска"
// @Success      200 {object} apiResponse{result=chain.PagedValidatorsResponse}
// @Router       /v1/chains/{id}/validators [get]
func (c *ChainsController) GetPagedValidators(context *gin.Context) {
	request := chain.PagedValidatorsInput{
		ChainID: context.Param("id"),
	}

	limit, _ := strconv.ParseUint(context.Query("limit"), 0, 64)
	if limit <= 0 {
		limit = 10
	}

	offset, _ := strconv.ParseUint(context.Query("offset"), 0, 64)
	if offset < 0 {
		offset = 0
	}

	request.Limit = limit
	request.Offset = offset
	handleRequest(request, context, c.service.GetPagedValidators)
}
