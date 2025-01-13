-- WAF 速率限制模块
local error_codes = require "error_codes"
local logger = require "logger"

local _M = {}

-- 默认配置
local default_config = {
    window = 60,           -- 时间窗口(秒)
    limit = 100,          -- 请求限制
    block_time = 300      -- 阻断时间(秒)
}

-- 验证IP地址格式
local function is_valid_ip(ip)
    if not ip then
        return false
    end
    
    -- IPv4格式验证
    local chunks = {ip:match("^(%d+)%.(%d+)%.(%d+)%.(%d+)$")}
    if #chunks == 4 then
        for _, v in ipairs(chunks) do
            if not v or tonumber(v) > 255 then
                return false
            end
        end
        return true
    end
    
    -- IPv6格式验证
    local chunks = {ip:match("^"..(("([a-fA-F0-9]*):"):rep(8):gsub(":$","$")))}
    if #chunks == 8 then
        for _, v in ipairs(chunks) do
            if not v or #v > 4 then
                return false
            end
        end
        return true
    end
    
    return false
end

-- 初始化共享内存字典
local function init_shared_dict()
    if not ngx.shared.rate_limit then
        return nil, error_codes.new_error(
            error_codes.codes.RUNTIME_ERROR,
            {
                message = "速率限制字典未定义",
                details = "请在nginx配置中添加 lua_shared_dict rate_limit 10m;"
            }
        )
    end
    
    -- 检查共享内存容量
    local capacity = ngx.shared.rate_limit:capacity()
    local free_space = ngx.shared.rate_limit:free_space()
    if free_space / capacity < 0.1 then  -- 剩余空间小于10%
        ngx.log(ngx.WARN, "速率限制共享内存空间不足")
        -- 清理过期的键
        ngx.shared.rate_limit:flush_expired()
    end
    
    return true
end

-- 生成键
local function generate_key(ip, uri)
    return string.format("rate_limit:%s:%s", ip, uri)
end

-- 检查是否被阻断
local function is_blocked(key)
    local block_key = key .. ":blocked"
    local blocked_until = ngx.shared.rate_limit:get(block_key)
    if blocked_until then
        if blocked_until > ngx.time() then
            return true, blocked_until - ngx.time()
        else
            ngx.shared.rate_limit:delete(block_key)
        end
    end
    return false
end

-- 增加计数
local function increment_counter(key, window)
    local success, err, forcible = ngx.shared.rate_limit:incr(key, 1, 0, window)
    if not success then
        if forcible then
            -- 强制删除一些过期的键
            ngx.shared.rate_limit:flush_expired()
            -- 重试一次
            success, err, forcible = ngx.shared.rate_limit:incr(key, 1, 0, window)
        end
        if not success then
            return nil, error_codes.new_error(
                error_codes.codes.RUNTIME_ERROR,
                {
                    message = "增加计数失败",
                    details = err
                }
            )
        end
    end
    return success
end

-- 设置阻断
local function set_block(key, block_time)
    local block_key = key .. ":blocked"
    -- 使用毫秒级时间戳提高精度
    local block_until = (ngx.now() * 1000 + block_time * 1000) / 1000
    local success, err, forcible = ngx.shared.rate_limit:set(block_key, block_until)
    if not success then
        if forcible then
            -- 强制删除一些过期的键
            ngx.shared.rate_limit:flush_expired()
            -- 重试一次
            success, err, forcible = ngx.shared.rate_limit:set(block_key, block_until)
        end
        if not success then
            return nil, error_codes.new_error(
                error_codes.codes.RUNTIME_ERROR,
                {
                    message = "设置阻断失败",
                    details = err
                }
            )
        end
    end
    return block_until
end

-- 初始化
function _M.init(config)
    -- 验证配置
    if not config then
        config = default_config
    end
    
    -- 初始化共享内存
    local ok, err = init_shared_dict()
    if not ok then
        return err
    end
    
    -- 保存配置
    _M.config = {
        window = config.window or default_config.window,
        limit = config.limit or default_config.limit,
        block_time = config.block_time or default_config.block_time
    }
    
    return true
end

-- 检查速率限制
function _M.is_limited(ip, uri)
    if not _M.config then
        return nil, error_codes.new_error(
            error_codes.codes.CONFIG_ERROR,
            "速率限制未初始化"
        )
    end
    
    -- 验证IP地址格式
    if not is_valid_ip(ip) then
        return nil, error_codes.new_error(
            error_codes.codes.INVALID_PARAMS,
            "无效的IP地址格式"
        )
    end
    
    -- 生成键
    local key = generate_key(ip, uri)
    
    -- 检查是否被阻断
    local is_block, remaining = is_blocked(key)
    if is_block then
        return true, error_codes.new_error(
            error_codes.codes.RATE_LIMIT_ERROR,
            {
                message = "请求被限制",
                details = {
                    remaining_seconds = remaining
                }
            }
        )
    end
    
    -- 增加计数
    local count, err = increment_counter(key, _M.config.window)
    if not count then
        return nil, err
    end
    
    -- 检查是否超过限制
    if count > _M.config.limit then
        -- 设置阻断
        local block_until, err = set_block(key, _M.config.block_time)
        if not block_until then
            return nil, err
        end
        
        return true, error_codes.new_error(
            error_codes.codes.RATE_LIMIT_ERROR,
            {
                message = "请求频率超限",
                details = {
                    window = _M.config.window,
                    limit = _M.config.limit,
                    count = count,
                    block_time = _M.config.block_time,
                    block_until = block_until
                }
            }
        )
    end
    
    return false
end

return _M 