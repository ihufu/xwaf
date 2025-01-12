-- 规则引擎调用模块
local http = require "resty.http"
local cjson = require "cjson"
local logger = require "logger"
local error_codes = require "error_codes"

local _M = {}

-- 默认配置
local default_config = {
    timeout = 1000,                     -- 超时时间(ms)
    retry_times = 3,                    -- 重试次数
    retry_interval = 0.1,               -- 重试间隔(s)
    cache_ttl = 60,                     -- 缓存时间(s)
}

-- 当前配置
local config = {}

-- 初始化配置
function _M.init(opts)
    if not opts or not opts.host then
        return nil, "规则引擎地址未配置"
    end
    
    config = setmetatable(opts, { __index = default_config })
    return true
end

-- HTTP请求封装
local function http_request(method, path, body, headers)
    local httpc = http.new()
    httpc:set_timeout(config.timeout)

    headers = headers or {}
    headers["Content-Type"] = "application/json"

    local tries = 0
    while tries < config.retry_times do
        local res, err = httpc:request_uri(config.host .. path, {
            method = method,
            body = body and cjson.encode(body) or nil,
            headers = headers
        })
        
        -- 确保关闭连接
        httpc:close()

        if not err then
            return res
        end

        tries = tries + 1
        if tries < config.retry_times then
            logger.warn("规则引擎请求失败,准备第" .. tries + 1 .. "次重试: " .. err)
            ngx.sleep(config.retry_interval)
        else
            return nil, error_codes.new_error(
                error_codes.codes.RULE_ENGINE_ERROR,
                "规则引擎请求失败",
                err
            )
        end
    end
end

-- 检查规则
function _M.check_rules(req)
    if not req then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "请求参数不能为空"
        )
    end

    -- 构建请求体
    local body = {
        request_id = req.request_id,
        client_ip = req.remote_addr,
        method = req.request_method,
        uri = req.uri,
        headers = req.headers,
        args = req.args,
        body = req.body
    }

    -- 发送请求
    local res, err = http_request("POST", "/api/v1/rules/check", body)
    if not res then
        return nil, err
    end

    -- 解析响应
    local ok, result = pcall(cjson.decode, res.body)
    if not ok then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "解析规则引擎响应失败",
            result
        )
    end

    -- 检查响应状态
    if res.status ~= 200 then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "规则引擎返回错误",
            result
        )
    end

    return result.data
end

-- 同步规则
function _M.sync_rules()
    local res, err = http_request("POST", "/api/v1/rules/sync")
    if not res then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "同步规则失败",
            err
        )
    end

    if res.status ~= 200 then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "同步规则失败",
            res.body
        )
    end

    return true
end

-- 获取规则版本
function _M.get_rules_version()
    local res, err = http_request("GET", "/api/v1/rules/version")
    if not res then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "获取规则版本失败",
            err
        )
    end

    if res.status ~= 200 then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "获取规则版本失败",
            res.body
        )
    end

    local ok, result = pcall(cjson.decode, res.body)
    if not ok then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "解析规则版本响应失败",
            result
        )
    end

    return result.data.version
end

return _M 