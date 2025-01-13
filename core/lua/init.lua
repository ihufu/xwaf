-- WAF 核心模块
local cjson = require "cjson"
local http = require "resty.http"
local logger = require "logger"
local metrics = require "metrics"
local ratelimit = require "ratelimit"
local config = require "config"
local error_codes = require "error_codes"

local _M = {}

-- WAF 配置
local waf_config = {
    mode = "block",  -- 默认模式为阻断
    last_update = 0  -- 上次更新时间
}

-- 创建带上下文的错误
local function create_error_with_context(code, message, context)
    return error_codes.new_error(code, message, context)
end

-- 初始化HTTP客户端
local function init_http_client()
    local httpc = http.new()
    httpc:set_timeout(config.get("rule_engine.timeout") or 1000)
    return httpc
end

-- 构建API请求头
local function build_api_headers()
    return {
        ["Content-Type"] = "application/json",
        ["Authorization"] = "Bearer " .. config.get("rule_engine.token"),
        ["X-Request-ID"] = ngx.var.request_id
    }
end

-- 获取WAF运行模式
local function get_waf_mode()
    if ngx.time() - waf_config.last_update < 60 then
        return waf_config.mode
    end

    local httpc = init_http_client()
    local rule_engine_host = config.get("rule_engine.host")
    if not rule_engine_host then
        logger.error("规则引擎地址未配置")
        return waf_config.mode
    end
    
    local start_time = ngx.now()
    local res, err = httpc:request_uri(rule_engine_host .. "/api/v1/config/mode", {
        method = "GET",
        headers = build_api_headers()
    })
    metrics.record_api_latency("get_mode", ngx.now() - start_time)
    
    if not res then
        logger.error("获取WAF模式失败: " .. tostring(err))
        return waf_config.mode
    end
    
    local ok, response = pcall(cjson.decode, res.body)
    if not ok or not response.data or not response.data.mode then
        logger.error("解析WAF模式响应失败")
        return waf_config.mode
    end
    
    waf_config.mode = response.data.mode
    waf_config.last_update = ngx.time()
    logger.info("WAF模式更新为: " .. waf_config.mode)
    
    return waf_config.mode
end

-- 处理请求
function _M.handle_request()
    local start_time = ngx.now()
    
    -- 获取当前WAF模式
    local mode = get_waf_mode()
    if mode == "bypass" then
        return true
    end
    
    -- 检查请求体大小
    local max_body_size = config.get("request.max_body_size") or 1048576  -- 默认1MB
    local content_length = tonumber(ngx.var.http_content_length) or 0
    if content_length > max_body_size then
        return false, create_error_with_context(
            error_codes.codes.WAF_REQUEST_TOO_LARGE,
            "请求体超过大小限制"
        )
    end
    
    -- 检查规则
    local httpc = init_http_client()
    local rule_engine_host = config.get("rule_engine.host")
    
    -- 构建请求数据
    local check_data = {
        request_id = ngx.var.request_id,
        client_ip = ngx.var.remote_addr,
        method = ngx.req.get_method(),
        uri = ngx.var.uri,
        headers = ngx.req.get_headers(),
        args = ngx.req.get_uri_args(),
        body = ngx.req.get_body_data(),
        mode = mode
    }
    
    -- 发送检查请求
    local res, err = httpc:request_uri(rule_engine_host .. "/api/v1/rules/check", {
        method = "POST",
        body = cjson.encode(check_data),
        headers = build_api_headers()
    })
    
    -- 关闭HTTP连接
    httpc:close()
    
    -- 记录API延迟
    metrics.record_api_latency("check_rules", ngx.now() - start_time)
    
    -- 处理响应
    if not res then
        logger.error("规则检查请求失败: " .. tostring(err))
        if config.get("rule_engine.fail_open") then
            return true
        end
        return false, create_error_with_context(
            error_codes.codes.WAF_RULE_ENGINE_REQUEST_ERROR,
            "规则引擎请求失败",
            err
        )
    end
    
    local ok, result = pcall(cjson.decode, res.body)
    if not ok or not result.data then
        logger.error("解析规则检查响应失败")
        if config.get("rule_engine.fail_open") then
            return true
        end
        return false, create_error_with_context(
            error_codes.codes.WAF_RULE_ENGINE_RESPONSE_ERROR,
            "规则引擎响应解析失败"
        )
    end
    
    -- 记录指标
    if result.data.matched then
        metrics.record_rule_match(result.data.rule_id, result.data.action)
    end
    
    -- 根据规则引擎的决定处理
    if result.data.should_block then
        return false, create_error_with_context(
            error_codes.codes.WAF_RULE_MATCH_ERROR,
            result.data.block_reason,
            result.data.context
        )
    end
    
    return true
end

-- 初始化worker
function _M.init_worker()
    -- 加载配置
    local ok, err = config.load()
    if err then  -- 修复错误判断逻辑
        logger.error("加载配置失败: " .. error_codes.to_json(err))
        return
    end
    
    -- 初始化速率限制
    ratelimit.init()
    
    -- 初始化监控指标
    metrics.init()
end

return _M 