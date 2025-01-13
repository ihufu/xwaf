-- WAF 核心模块
local cjson = require "cjson"
local logger = require "logger"
local metrics = require "metrics"
local ratelimit = require "ratelimit"
local config = require "config"
local error_codes = require "error_codes"
local http_client = require "http_client"

local _M = {}

-- 获取WAF运行模式
local function get_waf_mode()
    -- 检查本地缓存
    local mode = ngx.shared.waf_config:get("mode")
    if mode then
        return mode
    end

    -- 从规则引擎获取模式
    local rule_engine_host = config.get("rule_engine.host")
    if not rule_engine_host then
        logger.error("规则引擎地址未配置")
        return "block" -- 默认阻断模式
    end
    
    local start_time = ngx.now()
    local res, err = http_client.request("GET",
        rule_engine_host .. "/api/v1/config/mode",
        {
            headers = {
                ["Authorization"] = "Bearer " .. config.get("rule_engine.token"),
                ["X-Request-ID"] = ngx.var.request_id
            },
            parse_json = true
        }
    )
    metrics.record_api_latency("get_mode", ngx.now() - start_time)
    
    if not res then
        logger.error("获取WAF模式失败: " .. tostring(err))
        return "block" -- 默认阻断模式
    end
    
    if not res.parsed_body or not res.parsed_body.data or not res.parsed_body.data.mode then
        logger.error("解析WAF模式响应失败")
        return "block" -- 默认阻断模式
    end
    
    -- 更新本地缓存
    local mode = res.parsed_body.data.mode
    local ok, err = ngx.shared.waf_config:set("mode", mode, 60) -- 缓存60秒
    if not ok then
        logger.warn("更新WAF模式缓存失败: " .. tostring(err))
    end
    
    -- 记录模式变更
    local old_mode = ngx.shared.waf_config:get("last_mode")
    if old_mode and old_mode ~= mode then
        logger.info("WAF模式已变更: " .. old_mode .. " -> " .. mode)
        metrics.record_mode_change(old_mode, mode)
    end
    ngx.shared.waf_config:set("last_mode", mode)
    
    return mode
end

-- 处理请求
function _M.handle_request()
    local start_time = ngx.now()
    
    -- 获取当前WAF模式
    local mode = get_waf_mode()
    
    -- 根据不同模式处理请求
    if mode == "bypass" then
        return true
    elseif mode == "log" then
        local matched, result = match_rules()
        if matched then
            logger.warn("规则匹配记录: " .. cjson.encode(result))
        end
        return true
    elseif mode == "block" then
        local matched, result = match_rules()
        if matched then
            return handle_block(result)
        end
    else
        logger.error("未知的WAF模式: " .. mode)
        return true -- 未知模式时放行
    end
    
    -- 记录处理时间
    metrics.record_request_time(ngx.now() - start_time)
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