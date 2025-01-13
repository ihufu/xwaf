package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
)

// 错误码到HTTP状态码的映射
var errorCodeToHTTPStatus = map[errors.ErrorCode]int{
	errors.ErrValidation:     http.StatusBadRequest,          // 400 - 参数验证失败
	errors.ErrRuleNotFound:   http.StatusNotFound,            // 404 - 规则不存在
	errors.ErrRuleConflict:   http.StatusConflict,            // 409 - 规则冲突
	errors.ErrConfig:         http.StatusBadRequest,          // 400 - 配置错误
	errors.ErrCache:          http.StatusServiceUnavailable,  // 503 - 缓存服务不可用
	errors.ErrCacheMiss:      http.StatusNotFound,            // 404 - 缓存未命中
	errors.ErrRuleMatch:      http.StatusBadRequest,          // 400 - 规则匹配错误
	errors.ErrRuleEngine:     http.StatusInternalServerError, // 500 - 规则引擎错误
	errors.ErrRuleValidation: http.StatusBadRequest,          // 400 - 规则验证失败
	errors.ErrSQLInjection:   http.StatusBadRequest,          // 400 - SQL注入检测
	errors.ErrInit:           http.StatusServiceUnavailable,  // 503 - 初始化错误
	errors.ErrSystem:         http.StatusInternalServerError, // 500 - 系统错误
}

// getHTTPStatus 根据错误码获取对应的HTTP状态码
func getHTTPStatus(code errors.ErrorCode) int {
	if status, ok := errorCodeToHTTPStatus[code]; ok {
		return status
	}
	return http.StatusInternalServerError // 默认返回500
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, err *errors.Error) {
	if err == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    errors.ErrSystem,
			"message": "未知错误",
		})
		return
	}

	status := getHTTPStatus(err.Code)
	c.JSON(status, gin.H{
		"code":    err.Code,
		"message": err.Message,
	})
}

// ValidationError 返回参数验证错误响应
func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    errors.ErrValidation,
		"message": message,
	})
}

// NotFoundError 返回资源不存在错误响应
func NotFoundError(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{
		"code":    errors.ErrRuleNotFound,
		"message": message,
	})
}

// SystemError 返回系统错误响应
func SystemError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    errors.ErrSystem,
		"message": message,
	})
}

// ConflictError 返回资源冲突错误响应
func ConflictError(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, gin.H{
		"code":    errors.ErrRuleConflict,
		"message": message,
	})
}

// ServiceUnavailableError 返回服务不可用错误响应
func ServiceUnavailableError(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"code":    errors.ErrSystem,
		"message": message,
	})
}
