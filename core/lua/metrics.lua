-- 监控指标模块
local prometheus = require "prometheus"

local _M = {}

-- Prometheus 指标对象
local metrics = nil

-- 初始化监控指标
function _M.init()
    if metrics then
        return true
    end

    local ok, prom = pcall(prometheus.init, "waf_metrics", "waf_")
    if not ok then
        ngx.log(ngx.ERR, "初始化Prometheus指标失败: " .. tostring(prom))
        return nil, "初始化Prometheus指标失败"
    end
    
    metrics = prom
    if not metrics then
        return nil, "初始化Prometheus指标失败"
    end

    -- 请求计数器
    metrics.requests = metrics:counter(
        "requests_total",
        "WAF请求总数",
        {"method", "uri"}
    )

    -- 阻断计数器
    metrics.blocks = metrics:counter(
        "blocks_total",
        "WAF阻断总数",
        {"reason", "client_ip"}
    )

    -- 放行计数器
    metrics.allows = metrics:counter(
        "allows_total",
        "WAF放行总数",
        {"reason", "client_ip"}
    )

    -- 错误计数器
    metrics.errors = metrics:counter(
        "errors_total",
        "WAF错误总数",
        {"code"}
    )

    -- 规则匹配计数器
    metrics.rule_matches = metrics:counter(
        "rule_matches_total",
        "WAF规则匹配总数",
        {"rule_id", "action"}
    )

    -- 规则同步状态
    metrics.rule_sync = metrics:gauge(
        "rule_sync_status",
        "WAF规则同步状态",
        {"status"}
    )

    -- 健康检查状态
    metrics.health_check = metrics:gauge(
        "health_check_status",
        "WAF健康检查状态"
    )

    -- 规则引擎延迟直方图
    metrics.rule_engine_latency = metrics:histogram(
        "rule_engine_latency_seconds",
        "WAF规则引擎延迟",
        {"api"},
        {0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1}
    )

    -- 缓存延迟直方图
    metrics.cache_latency = metrics:histogram(
        "cache_latency_seconds",
        "WAF缓存延迟",
        {"operation"},
        {0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1}
    )

    -- 缓存命中计数器
    metrics.cache_hits = metrics:counter(
        "cache_hits_total",
        "WAF缓存命中总数",
        {"type"}
    )

    -- 缓存未命中计数器
    metrics.cache_misses = metrics:counter(
        "cache_misses_total",
        "WAF缓存未命中总数",
        {"type"}
    )

    return true
end

-- 记录请求
function _M.record_request(method, uri)
    if not metrics then
        return
    end
    metrics.requests:inc(1, {method, uri})
end

-- 记录阻断
function _M.record_block(reason, client_ip)
    if not metrics then
        return
    end
    metrics.blocks:inc(1, {reason, client_ip})
end

-- 记录放行
function _M.record_allow(reason, client_ip)
    if not metrics then
        return
    end
    metrics.allows:inc(1, {reason, client_ip})
end

-- 记录错误
function _M.record_error(code)
    if not metrics then
        return
    end
    metrics.errors:inc(1, {tostring(code)})
end

-- 记录规则匹配
function _M.record_rule_match(rule_id, action)
    if not metrics then
        return
    end
    metrics.rule_matches:inc(1, {rule_id, action})
end

-- 记录规则同步状态
function _M.record_sync_status(success)
    if not metrics then
        return
    end
    if success then
        metrics.rule_sync:set(1, {"success"})
        metrics.rule_sync:set(0, {"failure"})
    else
        metrics.rule_sync:set(0, {"success"})
        metrics.rule_sync:set(1, {"failure"})
    end
end

-- 记录健康检查状态
function _M.record_health_check(healthy, latency)
    if not metrics then
        return
    end
    metrics.health_check:set(healthy and 1 or 0)
    if latency then
        metrics.rule_engine_latency:observe(latency / 1000, {"health_check"})
    end
end

-- 记录缓存延迟
function _M.observe_cache_latency(latency)
    if not metrics then
        return
    end
    metrics.cache_latency:observe(latency, {"get"})
end

-- 记录缓存未命中
function _M.record_cache_miss()
    if not metrics then
        return
    end
    metrics.cache_misses:inc(1, {"rules"})
end

-- 记录规则引擎延迟（毫秒转换为秒）
function _M.record_rule_engine_latency(api, latency)
    if not metrics then
        return
    end
    metrics.rule_engine_latency:observe(latency / 1000, {api})
end

-- 导出指标
function _M.export()
    if not metrics then
        return nil, "监控指标未初始化"
    end
    return metrics:collect()
end

return _M 