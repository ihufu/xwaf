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
// 处理跨域请求，设置CORS相关响应头
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		// 设置CORS响应头
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

		// 处理OPTIONS请求
		if method == "OPTIONS" {
			if err := c.AbortWithError(http.StatusNoContent, errors.NewError(errors.Success, "预检请求成功")); err != nil {
				logger.Errorf("处理OPTIONS请求失败: %v", err)
			}
			return
		}

		// 记录跨域请求日志
		logger.Infof("收到跨域请求: Method=%s, Origin=%s", method, origin)
		c.Next()
	}
}

// RequestID 生成请求ID中间件
// 为每个请求生成唯一的请求ID，用于请求追踪
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			logger.Infof("生成新的请求ID: %s", requestID)
		} else {
			logger.Infof("使用已有请求ID: %s", requestID)
		}

		// 设置请求ID到上下文和响应头
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// ErrorHandler 错误处理中间件
// 统一处理请求过程中的错误，包括panic和普通错误
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := c.GetString("request_id")
				logger.Errorf("请求处理发生panic: RequestID=%s, Error=%v", requestID, err)

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

				// 设置请求ID并返回错误响应
				wafErr.RequestID = requestID
				handler.Error(c, wafErr)
				c.Abort()
			}
		}()

		// 继续处理请求
		c.Next()

		// 检查请求处理过程中是否有错误
		if len(c.Errors) > 0 {
			requestID := c.GetString("request_id")
			for _, e := range c.Errors {
				if err, ok := e.Err.(*errors.Error); ok {
					logger.Errorf("请求处理发生错误: RequestID=%s, Code=%d, Message=%s",
						requestID, err.Code, err.Message)
				} else {
					logger.Errorf("请求处理发生未知错误: RequestID=%s, Error=%v",
						requestID, e.Err)
				}
			}
		}
	}
}

// RequestLogger 请求日志中间件
// 记录请求的详细信息，包括请求方法、路径、状态码和处理时间
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method
		requestID := c.GetString("request_id")

		// 记录请求开始
		logger.Infof("开始处理请求: RequestID=%s, Method=%s, Path=%s",
			requestID, method, path)

		c.Next()

		// 构建完整的请求路径
		if raw != "" {
			path = path + "?" + raw
		}

		// 获取请求处理结果
		statusCode := c.Writer.Status()
		latency := time.Since(start)

		// 根据状态码选择日志级别
		if statusCode >= 400 {
			logger.Errorf("请求处理失败: [%s] %s %s %d %s",
				requestID, method, path, statusCode, latency)
		} else {
			logger.Infof("请求处理完成: [%s] %s %s %d %s",
				requestID, method, path, statusCode, latency)
		}
	}
}
