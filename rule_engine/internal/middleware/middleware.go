package middleware

import (
	"net/http"
	"time"

	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// RequestID 生成请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 处理panic错误
				var wafErr *errors.Error
				switch e := err.(type) {
				case *errors.Error:
					wafErr = e
				case error:
					wafErr = errors.NewError(errors.ErrSystem, e.Error())
				default:
					wafErr = errors.NewError(errors.ErrSystem, "系统内部错误")
				}
				wafErr.RequestID = c.GetString("request_id")
				handler.Error(c, wafErr)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		// 记录请求日志
		logger.Infof("[%s] %s %s %d %s",
			c.GetString("request_id"),
			c.Request.Method,
			path,
			c.Writer.Status(),
			time.Since(start),
		)
	}
}
