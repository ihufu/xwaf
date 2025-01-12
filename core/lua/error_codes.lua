-- 错误码定义模块
-- 注意：此错误码与 rule_engine/internal/errors/codes.go 保持一致
-- Go代码中使用驼峰命名(如 ErrInit)，Lua中使用WAF_前缀和下划线命名(如 WAF_INIT_ERROR)
local cjson = require "cjson"

local _M = {}

-- 错误码定义
_M.codes = {
    -- 成功 (对应 Go: Success)
    WAF_SUCCESS = 0,

    -- 系统错误 (1000-1999) (对应 Go: ErrXxx)
    WAF_INIT_ERROR = 1000,           -- 初始化错误
    WAF_CONFIG_ERROR = 1001,         -- 配置错误
    WAF_RUNTIME_ERROR = 1002,        -- 运行时错误
    WAF_SYSTEM_ERROR = 1003,         -- 系统错误
    WAF_TEMPLATE_ERROR = 1004,       -- 模板错误
    
    -- 请求错误 (2000-2999) (对应 Go: ErrXxx)
    WAF_INVALID_REQUEST = 2000,      -- 无效请求
    WAF_METHOD_NOT_ALLOWED = 2001,   -- 方法不允许
    WAF_REQUEST_TOO_LARGE = 2002,    -- 请求过大
    WAF_INVALID_PARAMS = 2003,       -- 无效参数
    WAF_RATE_LIMIT = 2004,          -- 请求频率限制
    WAF_RATE_LIMIT_ERROR = 2005,    -- 速率限制错误
    
    -- 规则错误 (3000-3999) (对应 Go: ErrRuleXxx)
    WAF_RULE_ENGINE_ERROR = 3000,    -- 规则引擎错误
    WAF_RULE_SYNC_ERROR = 3001,      -- 规则同步错误
    WAF_RULE_CHECK_ERROR = 3002,     -- 规则检查错误
    WAF_RULE_MATCH_ERROR = 3003,     -- 规则匹配错误
    WAF_RULE_VALIDATION_ERROR = 3004, -- 规则验证错误
    WAF_RULE_NOT_FOUND = 3005,       -- 规则不存在
    WAF_RULE_CONFLICT = 3006,        -- 规则冲突
    
    -- 缓存错误 (4000-4999) (对应 Go: ErrCacheXxx)
    WAF_CACHE_ERROR = 4000,          -- 缓存错误
    WAF_CACHE_MISS = 4001,           -- 缓存未命中
    WAF_CACHE_EXPIRED = 4002,        -- 缓存过期
    WAF_CACHE_INVALID = 4003,        -- 缓存无效
    
    -- 安全错误 (5000-5999) (对应 Go: ErrXxx)
    WAF_SECURITY_ERROR = 5000,       -- 安全错误
    WAF_IP_BLOCKED = 5001,           -- IP封禁
    WAF_CC_ATTACK = 5002,            -- CC攻击
    WAF_XSS_ATTACK = 5003,           -- XSS攻击
    WAF_SQL_INJECTION = 5004,        -- SQL注入
    WAF_AUTH_FAILED = 5005,          -- 认证失败
    WAF_PERM_DENIED = 5006           -- 权限不足
}

-- 错误码对应的消息
local error_messages = {
    [_M.codes.WAF_SUCCESS] = "成功",
    
    -- 系统错误
    [_M.codes.WAF_INIT_ERROR] = "初始化错误",
    [_M.codes.WAF_CONFIG_ERROR] = "配置错误",
    [_M.codes.WAF_RUNTIME_ERROR] = "运行时错误",
    [_M.codes.WAF_SYSTEM_ERROR] = "系统错误",
    [_M.codes.WAF_TEMPLATE_ERROR] = "模板错误",
    
    -- 请求错误
    [_M.codes.WAF_INVALID_REQUEST] = "无效的请求",
    [_M.codes.WAF_METHOD_NOT_ALLOWED] = "不支持的请求方法",
    [_M.codes.WAF_REQUEST_TOO_LARGE] = "请求体过大",
    [_M.codes.WAF_INVALID_PARAMS] = "无效的参数",
    [_M.codes.WAF_RATE_LIMIT] = "请求频率超限",
    [_M.codes.WAF_RATE_LIMIT_ERROR] = "速率限制错误",
    
    -- 规则错误
    [_M.codes.WAF_RULE_ENGINE_ERROR] = "规则引擎错误",
    [_M.codes.WAF_RULE_SYNC_ERROR] = "规则同步失败",
    [_M.codes.WAF_RULE_CHECK_ERROR] = "规则检查失败",
    [_M.codes.WAF_RULE_MATCH_ERROR] = "规则匹配失败",
    [_M.codes.WAF_RULE_VALIDATION_ERROR] = "规则验证失败",
    [_M.codes.WAF_RULE_NOT_FOUND] = "规则不存在",
    [_M.codes.WAF_RULE_CONFLICT] = "规则冲突",
    
    -- 缓存错误
    [_M.codes.WAF_CACHE_ERROR] = "缓存错误",
    [_M.codes.WAF_CACHE_MISS] = "缓存未命中",
    [_M.codes.WAF_CACHE_EXPIRED] = "缓存已过期",
    [_M.codes.WAF_CACHE_INVALID] = "缓存数据无效",
    
    -- 安全错误
    [_M.codes.WAF_SECURITY_ERROR] = "安全错误",
    [_M.codes.WAF_IP_BLOCKED] = "IP已被封禁",
    [_M.codes.WAF_CC_ATTACK] = "检测到CC攻击",
    [_M.codes.WAF_XSS_ATTACK] = "检测到XSS攻击",
    [_M.codes.WAF_SQL_INJECTION] = "检测到SQL注入攻击",
    [_M.codes.WAF_AUTH_FAILED] = "认证失败",
    [_M.codes.WAF_PERM_DENIED] = "权限不足"
}

-- 创建新的错误对象
function _M.new_error(code, message, details)
    local error = {
        code = code,
        message = message or error_messages[code],
        details = details,
        request_id = ngx.var.request_id,
        timestamp = ngx.now(),
        stack = nil  -- 仅在开发环境显示堆栈
    }
    
    -- 在开发环境添加堆栈信息
    if ngx.var.waf_env == "development" then
        error.stack = debug.traceback()
    end
    
    return error
end

-- 将错误对象转换为JSON
function _M.to_json(error)
    -- 在非开发环境下移除堆栈信息
    if ngx.var.waf_env ~= "development" then
        error.stack = nil
    end
    return cjson.encode(error)
end

-- 从JSON解析错误对象
function _M.from_json(json)
    local ok, error = pcall(cjson.decode, json)
    if not ok then
        return nil, "解析JSON失败: " .. error
    end
    return error
end

-- 判断是否为资源不存在错误
function _M.is_not_found(error)
    return error.code == _M.codes.WAF_RULE_NOT_FOUND
end

-- 判断是否为验证错误
function _M.is_validation_error(error)
    return error.code == _M.codes.WAF_RULE_VALIDATION_ERROR or 
           error.code == _M.codes.WAF_INVALID_PARAMS
end

-- 判断是否为安全错误
function _M.is_security_error(error)
    local code = error and error.code
    if not code then
        return false
    end
    return code >= _M.codes.WAF_SECURITY_ERROR and 
           code <= _M.codes.WAF_PERM_DENIED
end

-- 判断是否为缓存错误
function _M.is_cache_error(error)
    return error.code >= _M.codes.WAF_CACHE_ERROR and 
           error.code <= _M.codes.WAF_CACHE_INVALID
end

-- 判断是否应该重试
function _M.should_retry(error)
    local retry_codes = {
        _M.codes.WAF_RULE_SYNC_ERROR,
        _M.codes.WAF_CACHE_ERROR,
        _M.codes.WAF_CACHE_MISS,
        _M.codes.WAF_RULE_ENGINE_ERROR
    }
    
    for _, code in ipairs(retry_codes) do
        if error.code == code then
            return true
        end
    end
    
    return false
end

return _M 