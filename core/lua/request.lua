-- WAF请求处理模块
local cache = require "cache"
local cjson = require "cjson.safe"
local http = require "resty.http"
local metrics = require "metrics"
local rule_engine = require "rule_engine"
local config = require "config"
local error_codes = require "error_codes"

local _M = {}

-- 规则动作定义
local ACTIONS = {
    ALLOW = "allow",     -- 放行
    BLOCK = "block",     -- 阻断
    LOG = "log",         -- 仅记录
    CAPTCHA = "captcha"  -- 验证码
}

-- 从规则引擎或缓存获取规则
local function get_rules()
    local start_time = ngx.now()
    
    -- 尝试从缓存获取规则
    local rules = cache.get_rules()
    if rules then
        metrics.observe_cache_latency(ngx.now() - start_time)
        return rules
    end
    
    -- 缓存未命中，记录指标
    metrics.record_cache_miss()
    
    -- 从规则引擎获取规则并缓存
    local ok = cache.sync_rules(true)
    if not ok then
        ngx.log(ngx.ERR, "failed to sync rules from engine")
        return nil
    end
    
    rules = cache.get_rules()
    if rules then
        metrics.observe_cache_latency(ngx.now() - start_time)
        return rules
    end
    
    return nil
end

-- 获取请求信息
local function get_request_info()
    -- 检查请求体大小
    local content_length = tonumber(ngx.req.get_headers()["content-length"]) or 0
    local max_body_size = config.get("max_body_size") or 1024 * 1024  -- 默认1MB
    
    if content_length > max_body_size then
        return nil, error_codes.new_error(
            error_codes.codes.REQUEST_TOO_LARGE,
            string.format("请求体超过最大限制: %d > %d", content_length, max_body_size)
        )
    end
    
    -- 读取请求体
    ngx.req.read_body()
    local body = ngx.req.get_body_data()
    
    -- 构建请求信息
    return {
        client_ip = ngx.var.remote_addr,
        method = ngx.req.get_method(),
        uri = ngx.var.uri,
        headers = ngx.req.get_headers(),
        args = ngx.req.get_uri_args(),
        body = body,
        content_length = content_length
    }
end

-- 规则匹配
local function match_rule(rule, req_info)
    if not rule or not req_info then
        return false
    end

    -- 优先检查基本条件
    if rule.method and rule.method ~= req_info.method then
        return false
    end
    
    if rule.uri_pattern and not ngx.re.match(req_info.uri, rule.uri_pattern, "jo") then
        return false
    end

    -- 构建检查请求
    local check_req = {
        request_id = ngx.var.request_id,
        client_ip = req_info.client_ip,
        method = req_info.method,
        uri = req_info.uri,
        headers = req_info.headers,
        args = req_info.args,
        body = req_info.body,
        rule_types = {rule.type}  -- 只检查当前规则类型
    }

    -- 记录开始时间
    local start_time = ngx.now()

    -- 调用规则引擎检查
    local ok, res = pcall(rule_engine.check_rules, check_req)
    
    -- 记录规则检查延迟
    metrics.record_rule_engine_latency("check_rules", (ngx.now() - start_time) * 1000)
    
    if not ok then
        ngx.log(ngx.ERR, "规则匹配失败: ", res)
        return false, nil
    end

    -- 检查结果
    if res and res.matched and res.matched_rule and res.matched_rule.id == rule.id then
        return true, res
    end
    
    return false, nil
end

-- 处理匹配的规则
local function handle_matched_rule(rule, match_result)
    if not rule or not rule.action then
        return ngx.exit(ngx.HTTP_FORBIDDEN)
    end

    local action = rule.action
    if action == ACTIONS.ALLOW then
        return ngx.OK
    elseif action == ACTIONS.BLOCK then
        ngx.log(ngx.WARN, "请求被规则拦截: ", rule.id)
        return ngx.exit(ngx.HTTP_FORBIDDEN)
    elseif action == ACTIONS.LOG then
        ngx.log(ngx.WARN, "规则匹配记录: ", cjson.encode(match_result))
        return ngx.OK
    elseif action == ACTIONS.CAPTCHA then
        -- TODO: 实现验证码逻辑
        return ngx.exit(ngx.HTTP_FORBIDDEN)
    end

    return ngx.exit(ngx.HTTP_FORBIDDEN)
end

-- 检查请求
function _M.check_request()
    -- 获取规则
    local rules = get_rules()
    if not rules then
        return ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
    end

    -- 获取请求信息
    local req_info, err = get_request_info()
    if not req_info then
        return ngx.exit(ngx.HTTP_BAD_REQUEST)
    end
    
    -- 执行规则匹配
    for _, rule in ipairs(rules) do
        local matched, result = match_rule(rule, req_info)
        if matched then
            return handle_matched_rule(rule, result)
        end
    end
    
    -- 未匹配任何规则，放行
    return ngx.OK
end

return _M 