package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	headerRequestID = "X-Request-ID"
	ctxRequestIDKey = "request_id"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(ctxRequestIDKey, rid)
		c.Writer.Header().Set(headerRequestID, rid)
		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(ctxRequestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
