-- OpenResty ngx 对象的 mock 实现
local _M = {}

-- 日志级别
_M.ERR = 4
_M.WARN = 5
_M.INFO = 6
_M.DEBUG = 7

-- 共享字典
_M.shared = {
    waf_cache = {
        get = function() return nil end,
        set = function() return true end
    }
}

-- 变量
_M.var = {
    request_id = "test-request-id",
    remote_addr = "127.0.0.1",
    uri = "/test",
    request_method = "GET",
    host = "localhost"
}

-- 请求方法
_M.HTTP_GET = "GET"
_M.HTTP_POST = "POST"
_M.HTTP_PUT = "PUT"
_M.HTTP_DELETE = "DELETE"
_M.HTTP_HEAD = "HEAD"

-- 状态码
_M.HTTP_OK = 200
_M.HTTP_FORBIDDEN = 403

-- 日志函数
_M.log = function() end

-- 时间函数
_M.time = function() return os.time() end

-- 请求相关函数
_M.req = {
    get_headers = function()
        return {
            ["Accept"] = "text/html",
            ["User-Agent"] = "curl/7.64.1"
        }
    end,
    get_uri_args = function()
        return {}
    end,
    get_body_data = function()
        return nil
    end
}

-- 响应相关函数
_M.header = {}
_M.status = 200
_M.say = function() end
_M.exit = function() end

-- 配置相关函数
_M.config = {
    prefix = function()
        return "/usr/local/openresty/nginx/"
    end
}

return _M 