-- 配置管理模块
local cjson = require "cjson"
local logger = require "logger"
local error_codes = require "error_codes"

local _M = {}

-- 配置缓存
local config_cache = {}

-- 记录日志
local function log_error(msg)
    if logger then
        logger.error(msg)
    elseif ngx and ngx.log then
        ngx.log(ngx.ERR, msg)
    end
end

-- 加载配置文件
local function load_config_file()
    -- 获取nginx配置前缀
    local prefix = ngx.config.prefix()
    if not prefix then
        return nil, "无法获取nginx配置前缀"
    end
    
    -- 配置文件路径
    local config_path = os.getenv("XWAF_CONFIG_PATH")
    if config_path then
        -- 验证环境变量配置的路径
        local f = io.open(config_path, "r")
        if not f then
            return nil, "环境变量配置的路径无效: " .. config_path
        end
        f:close()
    end
    
    local config_file = config_path or (prefix .. "xwaf/core/lua/config.json")
    
    -- 读取配置文件
    local f, err = io.open(config_file, "r")
    if not f then
        return nil, "打开配置文件失败: " .. (err or "未知错误")
    end
    
    local content = f:read("*a")
    f:close()
    
    -- 解析配置文件
    local ok, config = pcall(cjson.decode, content)
    if not ok then
        return nil, error_codes.new_error(
            error_codes.codes.CONFIG_ERROR,
            "解析配置文件失败",
            config
        )
    end
    
    return config
end

-- 验证配置
local function validate_config(config)
    -- 检查必要的配置项
    if not config.rule_engine or not config.rule_engine.host then
        return false, error_codes.new_error(
            error_codes.codes.CONFIG_ERROR,
            "缺少规则引擎配置"
        )
    end
    
    -- 检查日志配置
    if config.log then
        if config.log.level and not logger.is_valid_level(config.log.level) then
            return false, error_codes.new_error(
                error_codes.codes.CONFIG_ERROR,
                "无效的日志级别",
                config.log.level
            )
        end
        if config.log.max_size and type(config.log.max_size) ~= "number" then
            return false, error_codes.new_error(
                error_codes.codes.CONFIG_ERROR,
                "日志大小限制必须是数字"
            )
        end
    end
    
    -- 检查请求配置
    if config.request then
        if config.request.max_body_size and type(config.request.max_body_size) ~= "number" then
            return false, error_codes.new_error(
                error_codes.codes.CONFIG_ERROR,
                "请求体大小限制必须是数字"
            )
        end
        if config.request.max_uri_length and type(config.request.max_uri_length) ~= "number" then
            return false, error_codes.new_error(
                error_codes.codes.CONFIG_ERROR,
                "URI长度限制必须是数字"
            )
        end
    end
    
    -- 检查速率限制配置
    if config.rate_limit then
        if config.rate_limit.enable ~= nil and type(config.rate_limit.enable) ~= "boolean" then
            return false, error_codes.new_error(
                error_codes.codes.CONFIG_ERROR,
                "速率限制开关必须是布尔值"
            )
        end
        if config.rate_limit.rate and type(config.rate_limit.rate) ~= "number" then
            return false, error_codes.new_error(
                error_codes.codes.CONFIG_ERROR,
                "速率限制值必须是数字"
            )
        end
    end
    
    -- 检查Redis配置
    if config.redis then
        if config.redis.pool_size and type(config.redis.pool_size) ~= "number" then
            return false, error_codes.new_error(
                error_codes.codes.CONFIG_ERROR,
                "Redis连接池大小必须是数字"
            )
        end
        if config.redis.timeout and type(config.redis.timeout) ~= "number" then
            return false, error_codes.new_error(
                error_codes.codes.CONFIG_ERROR,
                "Redis超时时间必须是数字"
            )
        end
    end
    
    return true
end

-- 加载配置
function _M.load()
    -- 获取锁
    local lock = ngx.shared.waf_locks
    local ok, err = lock:add("config_lock", true, 30)  -- 30秒超时
    if not ok then
        if err == "exists" then
            -- 等待其他进程完成配置加载
            ngx.sleep(0.1)
            return _M.get()
        end
        return nil, "获取配置锁失败: " .. err
    end
    
    -- 确保最后释放锁
    local function cleanup()
        lock:delete("config_lock")
    end
    
    -- 先尝试从共享内存中获取配置
    local config_str = ngx.shared.waf_cache:get("config")
    if config_str then
        local ok, config = pcall(cjson.decode, config_str)
        if ok then
            config_cache = config
            cleanup()
            return config_cache
        else
            log_error("解析共享内存配置失败: " .. tostring(config))
        end
    end
    
    -- 加载配置文件
    local config, err = load_config_file()
    if err then
        log_error("加载配置文件失败: " .. error_codes.to_json(err))
        cleanup()
        return nil, err
    end
    
    -- 验证配置
    local ok, err = validate_config(config)
    if not ok then
        log_error("配置验证失败: " .. error_codes.to_json(err))
        cleanup()
        return nil, err
    end
    
    -- 保存到共享内存
    local ok, err = ngx.shared.waf_cache:set("config", cjson.encode(config))
    if not ok then
        log_error("保存配置到共享内存失败: " .. tostring(err))
        cleanup()
        return nil, err
    end
    
    -- 更新配置缓存
    config_cache = config
    
    -- 释放锁
    cleanup()
    
    return config
end

-- 获取配置项
function _M.get(key)
    if not key then
        return config_cache
    end
    
    -- 获取锁
    local lock = ngx.shared.waf_locks
    local ok, err = lock:add("config_lock", true, 30)  -- 30秒超时
    if ok then
        -- 尝试从共享内存更新配置缓存
        local config_str = ngx.shared.waf_cache:get("config")
        if config_str then
            local ok, config = pcall(cjson.decode, config_str)
            if ok then
                config_cache = config
            else
                log_error("解析配置缓存失败: " .. tostring(config))
            end
        end
        
        -- 释放锁
        lock:delete("config_lock")
    end
    
    -- 如果缓存为空则加载配置
    if not next(config_cache) then
        local config, err = _M.load()
        if err then
            log_error("加载配置失败: " .. error_codes.to_json(err))
            return nil, err
        end
        config_cache = config
    end
    
    -- 支持多级key访问
    local value = config_cache
    for k in string.gmatch(key, "[^.]+") do
        if type(value) ~= "table" then
            return nil, "配置路径无效"
        end
        value = value[k]
        if value == nil then
            return nil, "配置项不存在: " .. key
        end
    end
    
    return value
end

-- 设置配置项
function _M.set(key, value)
    if not key then
        return false, "配置键不能为空"
    end
    
    -- 检查value类型
    if value == nil then
        return false, "配置值不能为nil"
    end
    
    -- 获取锁
    local lock = ngx.shared.waf_locks
    local ok, err = lock:add("config_lock", true, 30)  -- 30秒超时
    if not ok then
        return false, "获取配置锁失败: " .. err
    end
    
    -- 确保最后释放锁
    local function cleanup()
        lock:delete("config_lock")
    end
    
    local current = config_cache
    local keys = {}
    for k in string.gmatch(key, "[^.]+") do
        table.insert(keys, k)
    end
    
    -- 检查原有值的类型
    local old_value = _M.get(key)
    if old_value ~= nil and type(old_value) ~= type(value) then
        cleanup()
        return false, "配置值类型不匹配"
    end
    
    -- 构建新配置
    local new_config = table.clone(config_cache)  -- 深度复制
    local current = new_config
    for i = 1, #keys - 1 do
        local k = keys[i]
        if type(current[k]) ~= "table" then
            current[k] = {}
        end
        current = current[k]
    end
    current[keys[#keys]] = value
    
    -- 验证新配置
    local ok, err = validate_config(new_config)
    if not ok then
        cleanup()
        return false, err
    end
    
    -- 更新共享内存
    local ok, err = ngx.shared.waf_cache:set("config", cjson.encode(new_config))
    if not ok then
        log_error("保存配置到共享内存失败: " .. tostring(err))
        cleanup()
        return false, err
    end
    
    -- 更新配置缓存
    config_cache = new_config
    
    -- 释放锁
    cleanup()
    
    return true
end

-- 重新加载配置
function _M.reload()
    -- 获取锁
    local lock = ngx.shared.waf_locks
    local ok, err = lock:add("config_lock", true, 30)  -- 30秒超时
    if not ok then
        return nil, "获取配置锁失败: " .. err
    end
    
    -- 确保最后释放锁
    local function cleanup()
        lock:delete("config_lock")
    end
    
    -- 清除共享内存中的配置
    ngx.shared.waf_cache:delete("config")
    
    -- 重新加载配置
    local config, err = _M.load()
    if err then
        cleanup()
        return nil, err
    end
    
    -- 释放锁
    cleanup()
    
    return config
end

return _M