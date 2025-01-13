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
	File      string      `json:"file"`       // 错误发生的文件
	Line      int         `json:"line"`       // 错误发生的行号
}

// NewError 创建新的错误
func NewError(code ErrorCode, details interface{}) *Error {
	err := &Error{
		Code:      code,
		Message:   ErrorMessage[code],
		Details:   details,
		Timestamp: time.Now().Unix(),
	}

	// 在开发环境添加堆栈和位置信息
	if IsDebug() {
		// 获取调用栈
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		err.Stack = string(buf[:n])

		// 获取错误发生的文件和行号
		_, file, line, ok := runtime.Caller(1)
		if ok {
			err.File = file
			err.Line = line
		}
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
		e.File = ""
		e.Line = 0
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
	env := os.Getenv("WAF_ENV")
	return env == "development" || env == "debug"
}

// IsNotFound 判断是否为资源不存在错误
func (e *Error) IsNotFound() bool {
	return e.Code == ErrRuleNotFound
}

// IsValidationError 判断是否为验证错误
func (e *Error) IsValidationError() bool {
	return e.Code == ErrValidation || e.Code == ErrInvalidParams || e.Code == ErrRuleValidation
}

// IsSecurityError 判断是否为安全错误
func (e *Error) IsSecurityError() bool {
	return e.Code >= ErrSecurity && e.Code <= ErrPermDenied
}

// IsCacheError 判断是否为缓存错误
func (e *Error) IsCacheError() bool {
	return e.Code >= ErrCache && e.Code <= ErrCacheInvalid
}

// IsRuleError 判断是否为规则相关错误
func (e *Error) IsRuleError() bool {
	return e.Code >= ErrRuleEngine && e.Code <= ErrRuleConflict
}

// IsSystemError 判断是否为系统错误
func (e *Error) IsSystemError() bool {
	return e.Code >= ErrInit && e.Code <= ErrValidation
}

// IsRequestError 判断是否为请求错误
func (e *Error) IsRequestError() bool {
	return e.Code >= ErrInvalidRequest && e.Code <= ErrRateLimit
}

// ShouldRetry 判断是否应该重试
func (e *Error) ShouldRetry() bool {
	switch e.Code {
	case ErrRuleSync, ErrCache, ErrCacheMiss, ErrRuleEngine, ErrSystem:
		return true
	default:
		return false
	}
}

// GetHTTPStatus 获取对应的HTTP状态码
func (e *Error) GetHTTPStatus() int {
	switch {
	case e.IsValidationError():
		return 400 // Bad Request
	case e.IsNotFound():
		return 404 // Not Found
	case e.IsSecurityError():
		return 403 // Forbidden
	case e.IsRequestError():
		if e.Code == ErrMethodNotAllowed {
			return 405 // Method Not Allowed
		}
		if e.Code == ErrRateLimit {
			return 429 // Too Many Requests
		}
		return 400 // Bad Request
	case e.IsSystemError():
		return 500 // Internal Server Error
	default:
		return 500 // Internal Server Error
	}
}

// WithStack 添加堆栈信息
func (e *Error) WithStack() *Error {
	if IsDebug() {
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		e.Stack = string(buf[:n])

		_, file, line, ok := runtime.Caller(1)
		if ok {
			e.File = file
			e.Line = line
		}
	}
	return e
}

// WithMessage 添加额外的错误信息
func (e *Error) WithMessage(format string, args ...interface{}) *Error {
	if len(args) > 0 {
		e.Message = fmt.Sprintf(format, args...)
	} else {
		e.Message = format
	}
	return e
}

// WithDetails 添加详细信息
func (e *Error) WithDetails(details interface{}) *Error {
	e.Details = details
	return e
}
