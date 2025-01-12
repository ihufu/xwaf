package errors

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

// Error WAF错误结构
type Error struct {
	Code      ErrorCode   `json:"code"`       // 错误码
	Message   string      `json:"message"`    // 错误描述
	Details   interface{} `json:"details"`    // 详细信息
	RequestID string      `json:"request_id"` // 请求ID
	Timestamp int64       `json:"timestamp"`  // 时间戳
	Stack     string      `json:"stack"`      // 错误堆栈(仅在开发环境显示)
}

// NewError 创建新的错误
func NewError(code ErrorCode, details interface{}) *Error {
	err := &Error{
		Code:      code,
		Message:   ErrorMessage[code],
		Details:   details,
		Timestamp: time.Now().Unix(),
	}

	// 在开发环境添加堆栈信息
	if IsDebug() {
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		err.Stack = string(buf[:n])
	}

	return err
}

// WithRequestID 设置请求ID
func (e *Error) WithRequestID(requestID string) *Error {
	e.RequestID = requestID
	return e
}

// Error 实现error接口
func (e *Error) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// ToJSON 将错误转换为JSON
func (e *Error) ToJSON() ([]byte, error) {
	// 在非开发环境下移除堆栈信息
	if !IsDebug() {
		e.Stack = ""
	}
	return json.Marshal(e)
}

// FromJSON 从JSON解析错误
func FromJSON(data []byte) (*Error, error) {
	var err Error
	if e := json.Unmarshal(data, &err); e != nil {
		return nil, e
	}
	return &err, nil
}

// IsDebug 判断是否为开发环境
func IsDebug() bool {
	return os.Getenv("WAF_ENV") == "development"
}

// IsNotFound 判断是否为资源不存在错误
func (e *Error) IsNotFound() bool {
	return e.Code == ErrRuleNotFound
}

// IsValidationError 判断是否为验证错误
func (e *Error) IsValidationError() bool {
	return e.Code == ErrRuleValidation || e.Code == ErrInvalidParams
}

// IsSecurityError 判断是否为安全错误
func (e *Error) IsSecurityError() bool {
	return e.Code >= ErrSecurity && e.Code <= ErrPermDenied
}

// IsCacheError 判断是否为缓存错误
func (e *Error) IsCacheError() bool {
	return e.Code >= ErrCache && e.Code <= ErrCacheInvalid
}

// ShouldRetry 判断是否应该重试
func (e *Error) ShouldRetry() bool {
	switch e.Code {
	case ErrRuleSync, ErrCache, ErrCacheMiss, ErrRuleEngine:
		return true
	default:
		return false
	}
}
