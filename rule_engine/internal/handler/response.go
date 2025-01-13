package handler

import (
	"net/http"
	"time"

	"github.com/xwaf/rule_engine/internal/errors"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	resp := Response{
		Code:      int(errors.Success),
		Message:   errors.ErrorMessages[errors.Success],
		Data:      data,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	}
	c.JSON(http.StatusOK, resp)
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	var resp Response

	if e, ok := err.(*errors.Error); ok {
		resp = Response{
			Code:      int(e.Code),
			Message:   e.Message,
			Data:      e.Details,
			RequestID: e.RequestID,
			Timestamp: e.Timestamp,
		}
	} else {
		// 未知错误处理为系统错误
		resp = Response{
			Code:      int(errors.ErrSystem),
			Message:   err.Error(),
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now().Unix(),
		}
	}

	// 根据错误码确定HTTP状态码
	httpStatus := getHTTPStatus(resp.Code)
	c.JSON(httpStatus, resp)
}

// getHTTPStatus 根据错误码获取对应的HTTP状态码
func getHTTPStatus(code int) int {
	switch {
	case code == 0:
		return http.StatusOK
	case code < 1000:
		return http.StatusInternalServerError
	case code < 2000:
		return http.StatusBadRequest
	case code < 3000:
		return http.StatusServiceUnavailable
	case code < 4000:
		return http.StatusNotFound
	default:
		return http.StatusForbidden
	}
}
