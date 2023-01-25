package http

import (
	"errors"
	"net"
	"os"
	"strings"

	v1 "github.com/Mobile-Web3/backend/internal/handler/http/v1"
	"github.com/Mobile-Web3/backend/internal/metrics"
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

				metrics.PanicsCounter.Incr(1)
				logger.Panic(value)

				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							return
						}
					}
				}

				v1.ErrorResponse(context, value)
			}
		}()

		context.Next()
	}
}

func metricsMiddleware(_ *gin.Context) {
	metrics.RpsCounter.Incr(1)
}
