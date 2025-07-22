package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strings"
)

const RequestIDKey = "ID"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//get or create request ID
		requestID := c.GetHeader("X-Request-ID")
		if strings.TrimSpace(requestID) == "" {
			requestID = uuid.New().String()
		}

		//writing X-Request-ID into context
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		//write X-Request-ID into Response
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}
