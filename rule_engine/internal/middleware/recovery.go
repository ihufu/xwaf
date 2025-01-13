package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// Recovery 恢复中间件
// 捕获并处理请求处理过程中的panic，确保服务不会因为panic而中断
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取请求信息
				requestID := c.GetString("request_id")
				path := c.Request.URL.Path
				method := c.Request.Method
				clientIP := c.ClientIP()

				// 获取堆栈信息
				stack := debug.Stack()

				// 记录详细的错误日志
				logger.Errorf("请求处理发生panic: RequestID=%s, Method=%s, Path=%s, ClientIP=%s, Error=%v\n堆栈信息:\n%s",
					requestID, method, path, clientIP, err, stack)

				// 创建WAF错误
				var wafErr *errors.Error
				switch e := err.(type) {
				case *errors.Error:
					wafErr = e
				case error:
					wafErr = errors.NewError(errors.ErrSystem, e.Error())
				default:
					wafErr = errors.NewError(errors.ErrSystem, "系统内部错误")
				}

				// 设置请求ID
				wafErr.RequestID = requestID

				// 返回错误响应
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":       wafErr.Code,
					"message":    wafErr.Message,
					"request_id": requestID,
					"timestamp":  wafErr.Timestamp,
				})
			}
		}()

		// 继续处理请求
		c.Next()
	}
}
