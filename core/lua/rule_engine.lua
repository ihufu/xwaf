-- 规则引擎调用模块
local cjson = require "cjson"
local logger = require "logger"
local error_codes = require "error_codes"
local config = require "config"
local http_client = require "http_client"

local _M = {}

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
    local rule_engine_host = config.get("rule_engine.host")
    if not rule_engine_host then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "规则引擎地址未配置"
        )
    end

    local res, err = http_client.request("POST", 
        rule_engine_host .. "/api/v1/rules/check", 
        {
            body = body,
            parse_json = true
        }
    )
    
    if not res then
        return nil, err
    end

    -- 检查响应状态
    if res.status ~= 200 then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "规则引擎返回错误",
            res.parsed_body
        )
    end

    return res.parsed_body.data
end

-- 同步规则
function _M.sync_rules()
    local rule_engine_host = config.get("rule_engine.host")
    if not rule_engine_host then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "规则引擎地址未配置"
        )
    end

    local res, err = http_client.request("POST", 
        rule_engine_host .. "/api/v1/rules/sync",
        { parse_json = true }
    )
    
    if not res then
        return nil, err
    end

    if res.status ~= 200 then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "同步规则失败",
            res.parsed_body
        )
    end

    return true
end

-- 获取规则版本
function _M.get_rules_version()
    local rule_engine_host = config.get("rule_engine.host")
    if not rule_engine_host then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "规则引擎地址未配置"
        )
    end

    local res, err = http_client.request("GET", 
        rule_engine_host .. "/api/v1/rules/version",
        { parse_json = true }
    )
    
    if not res then
        return nil, err
    end

    if res.status ~= 200 then
        return nil, error_codes.new_error(
            error_codes.codes.RULE_ENGINE_ERROR,
            "获取规则版本失败",
            res.parsed_body
        )
    end

    return res.parsed_body.data.version
end

return _M 