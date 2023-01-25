package v1

import (
	"github.com/Mobile-Web3/backend/internal/domain/transaction"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/gin-gonic/gin"
)

type TransactionsController struct {
	logger  log.Logger
	service *transaction.Service
}

func NewTransactionsController(logger log.Logger, service *transaction.Service) *TransactionsController {
	return &TransactionsController{
		logger:  logger,
		service: service,
	}
}

// SendTransaction godoc
// @Summary      Отправить транзакцию
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body transaction.SendInput true "body"
// @Success      200 {object} apiResponse{result=transaction.SendResponse}
// @Router       /v1/transactions/send [post]
func (c *TransactionsController) SendTransaction() gin.HandlerFunc {
	return newRequestHandler(c.service.SendTransaction, c.logger)
}

// SendTransactionFirebase godoc
// @Summary      Отправить транзакцию с подпиской на события тендерминта с пушами в firebase
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body transaction.SendInputFirebase true "body"
// @Success      200 {object} apiResponse{result=transaction.SendResponseFirebase}
// @Router       /v1/transactions/send/firebase [post]
func (c *TransactionsController) SendTransactionFirebase() gin.HandlerFunc {
	return newRequestHandler(c.service.SendTransactionWithEvents, c.logger)
}

// SimulateTransaction godoc
// @Summary      Симуляция транзакции для расчета параметров
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Content-Type application/json
// @param        request body transaction.SimulateInput true "body"
// @Success      200 {object} apiResponse{result=transaction.SimulateResponse}
// @Router       /v1/transactions/simulate [post]
func (c *TransactionsController) SimulateTransaction() gin.HandlerFunc {
	return newRequestHandler(c.service.SimulateTransaction, c.logger)
}
