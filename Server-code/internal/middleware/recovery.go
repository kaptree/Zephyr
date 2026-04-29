package middleware

import (
	"net"
	"net/http"
	"runtime/debug"

	"labelpro-server/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {

				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					brokenPipe = true
					_ = ne
				}

				stack := string(debug.Stack())

				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("request_id", c.GetString("request_id")),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("stack", stack),
				)

				if brokenPipe {
					_ = c.Error(err.(error))
					c.Abort()
					return
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "服务器内部错误",
					"data":    nil,
				})
			}
		}()
		c.Next()
	}
}
