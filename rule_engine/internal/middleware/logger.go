package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		latency := time.Since(start)
		logger.Infof("| %3d | %13v | %15s | %s",
			c.Writer.Status(),
			latency,
			c.ClientIP(),
			path,
		)
	}
}
