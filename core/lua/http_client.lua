-- HTTP客户端模块
local http = require "resty.http"
local cjson = require "cjson"
local logger = require "logger"
local error_codes = require "error_codes"
local config = require "config"

local _M = {}

-- HTTP请求封装
function _M.request(method, url, opts)
    opts = opts or {}
    local timeout = opts.timeout or config.get("rule_engine.timeout") or 1000
    local retry_times = opts.retry_times or config.get("rule_engine.max_retries") or 3
    local retry_interval = opts.retry_interval or config.get("rule_engine.retry_interval") or 0.1
    
    local httpc = http.new()
    httpc:set_timeout(timeout)
    
    -- 设置默认headers
    local headers = opts.headers or {}
    headers["Content-Type"] = headers["Content-Type"] or "application/json"
    
    -- 处理请求体
    local body = opts.body
    if body and type(body) == "table" then
        body = cjson.encode(body)
    end
    
    local tries = 0
    while tries < retry_times do
        local res, err = httpc:request_uri(url, {
            method = method,
            body = body,
            headers = headers,
            keepalive_timeout = opts.keepalive_timeout or 60000,
            keepalive_pool = opts.keepalive_pool or 10
        })
        
        -- 确保关闭连接
        httpc:close()
        
        if not err then
            -- 如果需要解析JSON响应
            if opts.parse_json and res.body then
                local ok, result = pcall(cjson.decode, res.body)
                if not ok then
                    return nil, error_codes.new_error(
                        error_codes.codes.HTTP_ERROR,
                        "解析响应JSON失败",
                        result
                    )
                end
                res.parsed_body = result
            end
            return res
        end
        
        tries = tries + 1
        if tries < retry_times then
            logger.warn(string.format("HTTP请求失败,准备第%d次重试: %s", tries + 1, err))
            ngx.sleep(retry_interval)
        else
            return nil, error_codes.new_error(
                error_codes.codes.HTTP_ERROR,
                "HTTP请求失败",
                err
            )
        end
    end
end

return _M 