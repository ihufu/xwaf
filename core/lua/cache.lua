local redis = require "resty.redis"
local cjson = require "cjson.safe"
local config = require "config"
local ngx = ngx
local log = ngx.log
local ERR = ngx.ERR
local INFO = ngx.INFO

local _M = {
    _VERSION = "1.0.0"
}

-- 本地缓存配置
local local_rules = ngx.shared.rules
local local_ttl = config.get("cache.ttl") or 300  -- 从配置文件读取 TTL，默认 5 分钟

-- Redis 配置从配置文件读取
local redis_config = config.get("redis")
if not redis_config then
    log(ERR, "Redis configuration not found")
end
redis_config.prefix = "waf:rules:"  -- 添加规则前缀

-- 初始化 Redis 连接
local function get_redis()
    local red = redis:new()
    red:set_timeout(redis_config.timeout)
    
    local ok, err = red:connect(redis_config.host, redis_config.port)
    if not ok then
        log(ERR, "failed to connect Redis: ", err)
        return nil, err
    end
    
    return red
end

-- 设置本地缓存
function _M.set_local(key, value, ttl)
    ttl = ttl or local_ttl
    local ok, err = local_rules:set(key, value, ttl)
    if not ok then
        log(ERR, "failed to set local cache: ", err)
        return false
    end
    return true
end

-- 获取本地缓存
function _M.get_local(key)
    local value, err = local_rules:get(key)
    if err then
        log(ERR, "failed to get local cache: ", err)
        return nil
    end
    return value
end

-- 从规则引擎获取规则
local function fetch_rules_from_engine()
    local http = require "resty.http"
    local httpc = http.new()
    
    -- 设置超时
    httpc:set_timeout(1000)
    
    -- 发送请求
    local res, err = httpc:request_uri(config.get("rule_engine.host") .. "/api/v1/rules", {
        method = "GET",
        headers = {
            ["Content-Type"] = "application/json"
        }
    })
    
    -- 关闭连接
    httpc:close()
    
    if not res then
        log(ERR, "获取规则失败: " .. tostring(err))
        return nil
    end
    
    -- 解析响应
    local ok, rules = pcall(cjson.decode, res.body)
    if not ok then
        log(ERR, "解析规则失败: " .. tostring(rules))
        return nil
    end
    
    return rules.data
end

-- 检查规则引擎版本
local function check_engine_version()
    local http = require "resty.http"
    local httpc = http.new()
    
    -- 设置超时
    httpc:set_timeout(1000)
    
    -- 发送请求
    local res, err = httpc:request_uri(config.get("rule_engine.host") .. "/api/v1/version", {
        method = "GET",
        headers = {
            ["Content-Type"] = "application/json"
        }
    })
    
    -- 关闭连接
    httpc:close()
    
    if not res then
        log(ERR, "获取版本失败: " .. tostring(err))
        return nil
    end
    
    -- 解析响应
    local ok, version = pcall(cjson.decode, res.body)
    if not ok then
        log(ERR, "解析版本失败: " .. tostring(version))
        return nil
    end
    
    return version.data.version
end

-- 设置分布式缓存
function _M.set_distributed(key, value, ttl)
    local red, err = get_redis()
    if not red then
        return false
    end
    
    local ok, err = red:set(redis_config.prefix .. key, value)
    if not ok then
        log(ERR, "failed to set Redis cache: " .. tostring(err))
        red:close()
        return false
    end
    
    if ttl then
        ok, err = red:expire(redis_config.prefix .. key, ttl)
        if not ok then
            log(ERR, "failed to set Redis TTL: " .. tostring(err))
        end
    end
    
    -- 关闭连接
    red:close()
    
    return true
end

-- 获取分布式缓存
function _M.get_distributed(key)
    local red, err = get_redis()
    if not red then
        return nil
    end
    
    local value, err = red:get(redis_config.prefix .. key)
    
    -- 关闭连接
    red:close()
    
    if not value or err then
        log(ERR, "failed to get Redis cache: " .. tostring(err))
        return nil
    end
    
    return value
end

-- 规则版本控制
function _M.get_rule_version()
    local red, err = get_redis()
    if not red then
        return nil
    end
    
    local version, err = red:get(redis_config.prefix .. "version")
    if not version or err then
        log(ERR, "failed to get rule version: ", err)
        return nil
    end
    
    return version
end

-- 设置规则版本
function _M.set_rule_version(version)
    local red, err = get_redis()
    if not red then
        return false
    end
    
    local ok, err = red:set(redis_config.prefix .. "version", version)
    if not ok then
        log(ERR, "failed to set rule version: ", err)
        return false
    end
    
    return true
end

-- 同步规则
function _M.sync_rules(force)
    local current_version = _M.get_rule_version()
    if not force and current_version then
        -- 检查规则引擎版本
        local engine_version = _M.check_engine_version()
        if engine_version == current_version then
            return true  -- 规则已是最新
        end
    end
    
    -- 从规则引擎获取规则
    local rules = _M.fetch_rules_from_engine()
    if not rules then
        return false
    end
    
    -- 更新本地和分布式缓存
    local rules_str = cjson.encode(rules)
    _M.set_local("rules", rules_str)
    _M.set_distributed("rules", rules_str)
    
    return true
end

-- 获取规则（优先从本地缓存获取）
function _M.get_rules()
    -- 先查本地缓存
    local rules = _M.get_local("rules")
    if rules then
        return cjson.decode(rules)
    end
    
    -- 本地缓存未命中，查分布式缓存
    rules = _M.get_distributed("rules")
    if rules then
        -- 更新本地缓存
        _M.set_local("rules", rules)
        return cjson.decode(rules)
    end
    
    -- 缓存都未命中，从规则引擎获取
    _M.sync_rules(true)
    rules = _M.get_local("rules")
    if rules then
        return cjson.decode(rules)
    end
    
    return nil
end

return _M 