package middleware

import (
	"time"

	"labelpro-server/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("request_id", c.GetString("request_id")),
		}

		if statusCode >= 500 {
			logger.Error("request completed", fields...)
		} else if statusCode >= 400 {
			logger.Warn("request completed", fields...)
		} else {
			logger.Info("request completed", fields...)
		}
	}
}
