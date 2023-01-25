package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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

func errorResponse(context *gin.Context, err error) {
	context.JSON(http.StatusOK, apiResponse{
		IsSuccess: false,
		Error:     err.Error(),
	})
}
