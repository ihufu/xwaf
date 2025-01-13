-- 日志管理模块
local cjson = require "cjson"
local error_codes = require "error_codes"

local _M = {}

-- 日志级别定义
local LOG_LEVELS = {
    DEBUG = ngx.DEBUG,
    INFO = ngx.INFO,
    WARN = ngx.WARN,
    ERROR = ngx.ERR
}

-- 当前日志级别
local current_level = LOG_LEVELS.INFO

-- 检查日志级别是否有效
function _M.is_valid_level(level)
    return LOG_LEVELS[string.upper(level)] ~= nil
end

-- 设置日志级别
function _M.set_level(level)
    level = string.upper(level)
    if not LOG_LEVELS[level] then
        return false, error_codes.new_error(
            error_codes.codes.CONFIG_ERROR,
            "无效的日志级别",
            level
        )
    end
    current_level = LOG_LEVELS[level]
    return true
end

-- 获取当前日志级别
function _M.get_level()
    for name, value in pairs(LOG_LEVELS) do
        if value == current_level then
            return name
        end
    end
    return "UNKNOWN"
end

-- 构建日志上下文
local function build_log_context()
    local context = {
        timestamp = ngx.now(),
        hostname = ngx.var.hostname,
        pid = ngx.worker.pid()
    }
    
    -- 如果在请求上下文中,添加请求相关信息
    if ngx.get_phase() ~= "init" then
        context.request_id = ngx.var.request_id
        context.client_ip = ngx.var.remote_addr
        context.uri = ngx.var.uri
        context.method = ngx.var.request_method
        context.server_name = ngx.var.server_name
        context.trace_id = ngx.ctx.trace_id
    end
    
    return context
end

-- 格式化日志消息
local function format_log_message(level, message, context)
    -- 确保message是字符串
    if type(message) ~= "string" then
        message = tostring(message)
    end
    
    -- 构建日志条目
    local log_entry = {
        level = level,
        message = message,
        context = context or build_log_context()
    }
    
    -- 安全地序列化日志条目
    local ok, json = pcall(cjson.encode, log_entry)
    if not ok then
        -- 如果序列化失败，返回一个简单的格式
        return string.format("[%s] %s (序列化失败: %s)", level, message, json)
    end
    
    return json
end

-- 检查日志文件大小
local function check_log_size(file_path, max_size)
    local f = io.open(file_path, "r")
    if not f then
        return true
    end
    
    -- 获取文件大小
    local size = f:seek("end")
    f:close()
    
    -- 如果超过最大大小，进行轮转
    if size >= max_size then
        -- 重命名当前日志文件
        local backup = file_path .. "." .. os.date("%Y%m%d%H%M%S")
        os.rename(file_path, backup)
        
        -- 创建新的日志文件
        f = io.open(file_path, "w")
        if f then
            f:close()
        end
        
        return true
    end
    
    return true
end

-- 写入日志文件
local function write_log(level, message)
    -- 获取日志文件路径
    local log_file = ngx.config.prefix() .. "logs/waf/" .. level:lower() .. ".log"
    
    -- 检查日志文件大小（默认最大100MB）
    local max_size = 100 * 1024 * 1024
    if not check_log_size(log_file, max_size) then
        ngx.log(ngx.ERR, "日志文件大小检查失败")
        return false
    end
    
    -- 打开日志文件
    local f = io.open(log_file, "a")
    if not f then
        ngx.log(ngx.ERR, "打开日志文件失败: " .. log_file)
        return false
    end
    
    -- 写入日志
    local ok, err = f:write(message .. "\n")
    f:close()
    
    if not ok then
        ngx.log(ngx.ERR, "写入日志失败: " .. tostring(err))
        return false
    end
    
    return true
end

-- 记录调试日志
function _M.debug(message)
    if current_level > LOG_LEVELS.DEBUG then
        return
    end
    local log_message = format_log_message("DEBUG", message)
    write_log("DEBUG", log_message)
    ngx.log(ngx.DEBUG, log_message)
end

-- 记录信息日志
function _M.info(message)
    if current_level > LOG_LEVELS.INFO then
        return
    end
    local log_message = format_log_message("INFO", message)
    write_log("INFO", log_message)
    ngx.log(ngx.INFO, log_message)
end

-- 记录警告日志
function _M.warn(message)
    if current_level > LOG_LEVELS.WARN then
        return
    end
    local log_message = format_log_message("WARN", message)
    write_log("WARN", log_message)
    ngx.log(ngx.WARN, log_message)
end

-- 记录错误日志
function _M.error(message)
    if current_level > LOG_LEVELS.ERROR then
        return
    end
    local log_message = format_log_message("ERROR", message)
    write_log("ERROR", log_message)
    ngx.log(ngx.ERR, log_message)
end

-- 记录带上下文的日志
function _M.log_with_context(level, message, context)
    level = string.upper(level)
    if not LOG_LEVELS[level] or current_level > LOG_LEVELS[level] then
        return
    end
    
    -- 合并默认上下文和自定义上下文
    local default_context = build_log_context()
    if context then
        for k, v in pairs(context) do
            default_context[k] = v
        end
    end
    
    local log_message = format_log_message(level, message, default_context)
    write_log(level, log_message)
    ngx.log(LOG_LEVELS[level], log_message)
end

return _M 