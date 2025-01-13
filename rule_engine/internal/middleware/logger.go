package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// Logger 日志中间件
// 记录请求的处理时间、状态码、客户端IP和请求路径
// 同时处理日志记录过程中可能出现的错误
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		// 使用defer确保即使发生panic也能记录日志
		defer func() {
			latency := time.Since(start)
			statusCode := c.Writer.Status()
			clientIP := c.ClientIP()

			// 检查是否存在错误
			if len(c.Errors) > 0 {
				// 记录错误日志
				for _, e := range c.Errors {
					if err, ok := e.Err.(*errors.Error); ok {
						logger.Errorf("请求处理发生错误: %s | 错误码: %d | URL: %s | IP: %s",
							err.Message,
							err.Code,
							path,
							clientIP,
						)
					} else {
						logger.Errorf("请求处理发生未知错误: %v | URL: %s | IP: %s",
							e.Err,
							path,
							clientIP,
						)
					}
				}
			}

			// 记录访问日志
			logger.Infof("| %3d | %13v | %15s | %s",
				statusCode,
				latency,
				clientIP,
				path,
			)
		}()

		c.Next()
	}
}
