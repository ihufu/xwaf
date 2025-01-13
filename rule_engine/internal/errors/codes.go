package errors

// ErrorCode 错误码类型
type ErrorCode int

const (
	// Success 成功
	Success ErrorCode = 0

	// 系统错误 (1000-1999)
	ErrInit       ErrorCode = 1000 // 初始化错误：组件初始化失败
	ErrConfig     ErrorCode = 1001 // 配置错误：配置文件格式错误或缺少必要配置
	ErrRuntime    ErrorCode = 1002 // 运行时错误：程序运行时发生的未预期错误
	ErrSystem     ErrorCode = 1003 // 系统错误：操作系统或底层服务错误
	ErrValidation ErrorCode = 1004 // 验证错误：输入参数验证失败

	// 请求错误 (2000-2999)
	ErrInvalidRequest   ErrorCode = 2000 // 无效请求：请求格式错误或缺少必要字段
	ErrMethodNotAllowed ErrorCode = 2001 // 方法不允许：不支持的HTTP方法
	ErrRequestTooLarge  ErrorCode = 2002 // 请求过大：请求体超过允许的大小限制
	ErrInvalidParams    ErrorCode = 2003 // 无效参数：请求参数格式错误或取值范围错误
	ErrRateLimit        ErrorCode = 2004 // 请求频率限制：超过接口调用频率限制

	// 规则错误 (3000-3999)
	ErrRuleEngine     ErrorCode = 3000 // 规则引擎错误：规则引擎内部错误
	ErrRuleSync       ErrorCode = 3001 // 规则同步错误：规则同步到节点失败
	ErrRuleCheck      ErrorCode = 3002 // 规则检查错误：规则语法或逻辑检查失败
	ErrRuleMatch      ErrorCode = 3003 // 规则匹配错误：规则匹配过程发生错误
	ErrRuleValidation ErrorCode = 3004 // 规则验证错误：规则格式或内容验证失败
	ErrRuleNotFound   ErrorCode = 3005 // 规则不存在：指定的规则未找到
	ErrRuleConflict   ErrorCode = 3006 // 规则冲突：规则之间存在冲突

	// 缓存错误 (4000-4999)
	ErrCache        ErrorCode = 4000 // 缓存错误：缓存操作失败
	ErrCacheMiss    ErrorCode = 4001 // 缓存未命中：缓存中不存在指定的数据
	ErrCacheExpired ErrorCode = 4002 // 缓存过期：缓存数据已过期
	ErrCacheInvalid ErrorCode = 4003 // 缓存无效：缓存数据格式错误或已损坏

	// 安全错误 (5000-5999)
	ErrSecurity     ErrorCode = 5000 // 安全错误：通用安全错误
	ErrIPBlocked    ErrorCode = 5001 // IP封禁：IP地址已被封禁
	ErrCCAttack     ErrorCode = 5002 // CC攻击：检测到CC攻击行为
	ErrXSSAttack    ErrorCode = 5003 // XSS攻击：检测到XSS攻击行为
	ErrSQLInjection ErrorCode = 5004 // SQL注入：检测到SQL注入攻击
	ErrAuthFailed   ErrorCode = 5005 // 认证失败：用户认证失败
	ErrPermDenied   ErrorCode = 5006 // 权限不足：用户没有执行该操作的权限
)

// ErrorMessage 错误码对应的消息
var ErrorMessage = map[ErrorCode]string{
	Success: "操作成功",

	// 系统错误
	ErrInit:       "系统初始化失败，请检查配置和依赖服务",
	ErrConfig:     "配置错误，请检查配置文件格式和必要参数",
	ErrRuntime:    "系统运行时错误，请联系技术支持",
	ErrSystem:     "系统内部错误，请稍后重试",
	ErrValidation: "输入参数验证失败，请检查输入内容",

	// 请求错误
	ErrInvalidRequest:   "无效的请求格式，请检查请求内容",
	ErrMethodNotAllowed: "不支持的请求方法，请使用正确的HTTP方法",
	ErrRequestTooLarge:  "请求内容过大，请减小请求体积",
	ErrInvalidParams:    "无效的请求参数，请检查参数格式和取值范围",
	ErrRateLimit:        "请求频率超限，请稍后重试",

	// 规则错误
	ErrRuleEngine:     "规则引擎处理失败，请重试或联系技术支持",
	ErrRuleSync:       "规则同步失败，请检查节点状态",
	ErrRuleCheck:      "规则检查失败，请检查规则语法和逻辑",
	ErrRuleMatch:      "规则匹配失败，请检查规则配置",
	ErrRuleValidation: "规则验证失败，请检查规则格式和内容",
	ErrRuleNotFound:   "规则不存在，请检查规则ID",
	ErrRuleConflict:   "规则存在冲突，请检查规则优先级和作用范围",

	// 缓存错误
	ErrCache:        "缓存操作失败，请重试",
	ErrCacheMiss:    "缓存数据不存在，请检查数据是否已过期",
	ErrCacheExpired: "缓存数据已过期，请重新获取",
	ErrCacheInvalid: "缓存数据无效，请重新获取",

	// 安全错误
	ErrSecurity:     "安全检查失败，请检查请求内容",
	ErrIPBlocked:    "IP地址已被封禁，请联系管理员",
	ErrCCAttack:     "检测到疑似CC攻击，请降低访问频率",
	ErrXSSAttack:    "检测到疑似XSS攻击，请检查输入内容",
	ErrSQLInjection: "检测到疑似SQL注入攻击，请检查输入内容",
	ErrAuthFailed:   "用户认证失败，请检查认证信息",
	ErrPermDenied:   "权限不足，请确认是否有执行该操作的权限",
}

// ErrorMessages 错误码对应的消息
var ErrorMessages = map[ErrorCode]string{
	Success:          "成功",
	ErrSystem:        "系统错误",
	ErrInvalidParams: "无效的参数",
	ErrRuleEngine:    "规则引擎错误",
	ErrValidation:    "验证错误",
	// ... existing code ...
}
