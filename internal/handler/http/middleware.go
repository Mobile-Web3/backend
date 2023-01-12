package http

import (
	"errors"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/Mobile-Web3/backend/pkg/log"
	"github.com/gin-gonic/gin"
)

func recoverMiddleware(logger log.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				context.Abort()

				var value error
				switch err.(type) {
				case string:
					value = errors.New(err.(string))
				case error:
					value = err.(error)
				default:
					value = errors.New("unknown error")
				}

				logger.Panic(value)
				recoveredPanics.WithLabelValues(value.Error()).Inc()

				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							return
						}
					}
				}

				context.JSON(http.StatusOK, newErrorResponse(value.Error()))
			}
		}()

		context.Next()
	}
}

func requestMetricsMiddleware(context *gin.Context) {
	totalRequests.WithLabelValues(context.Request.URL.String()).Inc()
}
