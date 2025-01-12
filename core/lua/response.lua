-- 响应处理模块
local cjson = require "cjson"
local template = require "resty.template"
local config = require "config"
local error_codes = require "error_codes"
local logger = require "logger"

local _M = {}

-- 渲染阻断页面
local function render_block_page(context)
    -- 获取模板路径
    local template_path = ngx.config.prefix() .. "html/block.html"
    
    -- 检查模板文件是否存在
    local f = io.open(template_path, "r")
    if not f then
        local err = error_codes.new_error(
            error_codes.codes.TEMPLATE_ERROR,
            "阻断页面模板不存在",
            template_path
        )
        logger.error("渲染阻断页面失败: " .. error_codes.to_json(err))
        return nil, err
    end
    
    -- 读取模板内容并限制大小
    local max_template_size = 1024 * 1024  -- 1MB
    local content = f:read("*a")
    f:close()
    
    if #content > max_template_size then
        local err = error_codes.new_error(
            error_codes.codes.TEMPLATE_ERROR,
            "模板文件过大"
        )
        logger.error("渲染阻断页面失败: " .. error_codes.to_json(err))
        return nil, err
    end
    
    -- 限制渲染内存使用
    local max_memory = 1024 * 1024 * 10  -- 10MB
    local memory_before = collectgarbage("count") * 1024
    
    -- 渲染模板
    local html, err = template.render_file(template_path, context)
    
    -- 检查内存使用
    local memory_after = collectgarbage("count") * 1024
    if memory_after - memory_before > max_memory then
        collectgarbage("collect")
        local err = error_codes.new_error(
            error_codes.codes.TEMPLATE_ERROR,
            "模板渲染内存超限"
        )
        logger.error("渲染阻断页面失败: " .. error_codes.to_json(err))
        return nil, err
    end
    
    if err then
        local render_err = error_codes.new_error(
            error_codes.codes.TEMPLATE_ERROR,
            "渲染阻断页面失败",
            err
        )
        logger.error("渲染阻断页面失败: " .. error_codes.to_json(render_err))
        return nil, render_err
    end
    
    return html
end

-- 发送阻断页面
function _M.send_block_page(code, error)
    -- 构建上下文(移除敏感信息)
    local context = {
        code = code,
        message = error and error.message or "请求被阻断",
        request_id = ngx.var.request_id,
        timestamp = ngx.time()
    }

    -- 记录阻断日志(包含完整信息供内部使用)
    logger.info("请求被阻断", {
        code = code,
        error = error,
        request_id = context.request_id,
        client_ip = ngx.var.remote_addr,
        uri = ngx.var.uri
    })

    -- 检查请求是否接受HTML
    local accept_header = ngx.req.get_headers()["Accept"]
    local want_json = accept_header and string.find(accept_header:lower(), "application/json")

    -- 设置响应头
    ngx.status = 403
    
    -- 限制响应头大小
    local max_header_size = 4096  -- 4KB
    local headers = {}
    if want_json then
        headers["content-type"] = "application/json; charset=utf-8"
    else
        headers["content-type"] = "text/html; charset=utf-8"
    end
    
    -- 检查响应头大小
    local header_size = 0
    for k, v in pairs(headers) do
        header_size = header_size + #k + #v + 4  -- 4 for ": " and "\r\n"
    end
    
    if header_size > max_header_size then
        ngx.log(ngx.ERR, "响应头超过最大限制")
        return ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
    end
    
    -- 设置响应头
    for k, v in pairs(headers) do
        ngx.header[k] = v
    end
    
    if want_json then
        ngx.say(cjson.encode({
            code = code,
            message = context.message,
            request_id = context.request_id,
            timestamp = context.timestamp
        }))
    else
        local html, err = render_block_page(context)
        if err then
            -- 如果渲染失败，返回简单的错误页面
            ngx.say(string.format(
                "<h1>请求被阻断</h1><p>错误代码: %d</p>",
                code
            ))
        else
            ngx.say(html)
        end
    end
    
    return ngx.exit(ngx.HTTP_FORBIDDEN)
end

return _M 