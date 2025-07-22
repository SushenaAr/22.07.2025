package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"log/slog"
)

const RequestLoggerKey = "logger" // key for request context

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//scoped logger
		logger := slog.Default().With("request_id",
			c.Request.Context().Value(RequestIDKey))

		//pushing logger into context
		ctx := context.WithValue(c.Request.Context(), RequestLoggerKey, logger)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		//logging
		logger.Info("request completed", "status", c.Writer.Status(),
			"method", c.Request.Method, "path", c.FullPath())
	}
}
