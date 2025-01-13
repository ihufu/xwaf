package errors

// ErrorCode 错误码类型
type ErrorCode int

const (
	// Success 成功
	Success ErrorCode = 0

	// 系统错误 (1000-1999)
	ErrInit    ErrorCode = 1000 // 初始化错误
	ErrConfig  ErrorCode = 1001 // 配置错误
	ErrRuntime ErrorCode = 1002 // 运行时错误
	ErrSystem  ErrorCode = 1003 // 系统错误

	// 请求错误 (2000-2999)
	ErrInvalidRequest   ErrorCode = 2000 // 无效请求
	ErrMethodNotAllowed ErrorCode = 2001 // 方法不允许
	ErrRequestTooLarge  ErrorCode = 2002 // 请求过大
	ErrInvalidParams    ErrorCode = 2003 // 无效参数
	ErrRateLimit        ErrorCode = 2004 // 请求频率限制

	// 规则错误 (3000-3999)
	ErrRuleEngine     ErrorCode = 3000 // 规则引擎错误
	ErrRuleSync       ErrorCode = 3001 // 规则同步错误
	ErrRuleCheck      ErrorCode = 3002 // 规则检查错误
	ErrRuleMatch      ErrorCode = 3003 // 规则匹配错误
	ErrRuleValidation ErrorCode = 3004 // 规则验证错误
	ErrRuleNotFound   ErrorCode = 3005 // 规则不存在
	ErrRuleConflict   ErrorCode = 3006 // 规则冲突

	// 缓存错误 (4000-4999)
	ErrCache        ErrorCode = 4000 // 缓存错误
	ErrCacheMiss    ErrorCode = 4001 // 缓存未命中
	ErrCacheExpired ErrorCode = 4002 // 缓存过期
	ErrCacheInvalid ErrorCode = 4003 // 缓存无效

	// 安全错误 (5000-5999)
	ErrSecurity     ErrorCode = 5000 // 安全错误
	ErrIPBlocked    ErrorCode = 5001 // IP封禁
	ErrCCAttack     ErrorCode = 5002 // CC攻击
	ErrXSSAttack    ErrorCode = 5003 // XSS攻击
	ErrSQLInjection ErrorCode = 5004 // SQL注入
	ErrAuthFailed   ErrorCode = 5005 // 认证失败
	ErrPermDenied   ErrorCode = 5006 // 权限不足
)

// ErrorMessage 错误码对应的消息
var ErrorMessage = map[ErrorCode]string{
	Success: "成功",

	// 系统错误
	ErrInit:    "初始化错误",
	ErrConfig:  "配置错误",
	ErrRuntime: "运行时错误",
	ErrSystem:  "系统错误",

	// 请求错误
	ErrInvalidRequest:   "无效的请求",
	ErrMethodNotAllowed: "不支持的请求方法",
	ErrRequestTooLarge:  "请求体过大",
	ErrInvalidParams:    "无效的参数",
	ErrRateLimit:        "请求频率超限",

	// 规则错误
	ErrRuleEngine:     "规则引擎错误",
	ErrRuleSync:       "规则同步失败",
	ErrRuleCheck:      "规则检查失败",
	ErrRuleMatch:      "规则匹配失败",
	ErrRuleValidation: "规则验证失败",
	ErrRuleNotFound:   "规则不存在",
	ErrRuleConflict:   "规则冲突",

	// 缓存错误
	ErrCache:        "缓存错误",
	ErrCacheMiss:    "缓存未命中",
	ErrCacheExpired: "缓存已过期",
	ErrCacheInvalid: "缓存数据无效",

	// 安全错误
	ErrSecurity:     "安全错误",
	ErrIPBlocked:    "IP已被封禁",
	ErrCCAttack:     "检测到CC攻击",
	ErrXSSAttack:    "检测到XSS攻击",
	ErrSQLInjection: "检测到SQL注入攻击",
	ErrAuthFailed:   "认证失败",
	ErrPermDenied:   "权限不足",
}

// ErrorMessages 错误码对应的消息
var ErrorMessages = map[ErrorCode]string{
	Success:          "成功",
	ErrSystem:        "系统错误",
	ErrInvalidParams: "无效的参数",
	ErrRuleEngine:    "规则引擎错误",
	// ... existing code ...
}
