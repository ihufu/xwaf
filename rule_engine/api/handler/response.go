package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
)

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
	c.JSON(http.StatusOK, gin.H{
		"code":    err.Code,
		"message": err.Message,
	})
}
