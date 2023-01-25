package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Mobile-Web3/backend/internal/metrics"
	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/gin-gonic/gin"
)

func MetricsHandler(context *gin.Context) {
	result := fmt.Sprintf(`<h3>Metrics</h3>
<div>average requests   per second for the last hour: %d</div>
<div>average errors     per second for the last hour: %d</div>
<div>average exceptions per second for the last hour: %d</div>`,
		metrics.RpsCounter.Rate()/60/60, metrics.ErrorsCounter.Rate()/60/60, metrics.PanicsCounter.Rate()/60/60)
	context.Writer.WriteHeader(http.StatusOK)
	context.Writer.Header().Set("Content-Type", "text/html")
	_, err := context.Writer.Write([]byte(result))
	if err != nil {
		context.String(http.StatusInternalServerError, err.Error())
		return
	}
}

type apiResponse struct {
	IsSuccess bool        `json:"isSuccess"`
	Error     string      `json:"error"`
	Result    interface{} `json:"result"`
}

func successResponse(context *gin.Context, result interface{}) {
	context.JSON(http.StatusOK, apiResponse{
		IsSuccess: true,
		Result:    result,
	})
}

func ErrorResponse(context *gin.Context, err error) {
	context.JSON(http.StatusOK, apiResponse{
		IsSuccess: false,
		Error:     err.Error(),
	})
}

type validatedRequest interface {
	Validate() error
}

type emptyHandle[TResponse any] func(ctx context.Context) (TResponse, error)
type requestHandler[TRequest validatedRequest, TResponse any] func(ctx context.Context, request TRequest) (TResponse, error)

func newEmptyHandler[TResponse any](handle emptyHandle[TResponse]) gin.HandlerFunc {
	return func(context *gin.Context) {
		response, err := handle(context.Request.Context())
		if err != nil {
			metrics.ErrorsCounter.Incr(1)
			ErrorResponse(context, err)
			return
		}

		successResponse(context, response)
	}
}

func newRequestHandler[TRequest validatedRequest, TResponse any](handleRequest requestHandler[TRequest, TResponse], logger log.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		var request TRequest
		if err := context.BindJSON(&request); err != nil {
			logger.Error(err)
			metrics.ErrorsCounter.Incr(1)
			ErrorResponse(context, err)
			return
		}

		if err := request.Validate(); err != nil {
			metrics.ErrorsCounter.Incr(1)
			ErrorResponse(context, err)
			return
		}

		response, err := handleRequest(context.Request.Context(), request)
		if err != nil {
			metrics.ErrorsCounter.Incr(1)
			ErrorResponse(context, err)
			return
		}

		successResponse(context, response)
	}
}

func handleRequest[TRequest validatedRequest, TResponse any](request TRequest, context *gin.Context, handle requestHandler[TRequest, TResponse]) {
	if err := request.Validate(); err != nil {
		metrics.ErrorsCounter.Incr(1)
		ErrorResponse(context, err)
		return
	}

	response, err := handle(context.Request.Context(), request)
	if err != nil {
		metrics.ErrorsCounter.Incr(1)
		ErrorResponse(context, err)
		return
	}

	successResponse(context, response)
}
