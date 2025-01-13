# WAF规则引擎API文档

## 1. 概述

### 1.1 版本控制
- 当前版本: v1
- 基础路径: `/api/v1`
- 版本策略: 使用URL路径版本控制，重大更新使用新的版本号

### 1.2 环境说明
- 开发环境: `https://dev-waf-api.example.com`
- 测试环境: `https://test-waf-api.example.com`
- 生产环境: `https://waf-api.example.com`

### 1.3 认证机制
```http
Authorization: Bearer <token>
```
- Token获取: 通过 `/auth/token` 接口获取
- Token有效期: 24小时
- Token刷新: 通过 `/auth/refresh` 接口刷新
- Token撤销: 通过 `/auth/revoke` 接口撤销

### 1.4 限流策略
- 普通接口: 1000次/分钟
- 批量操作: 100次/分钟
- 规则检查: 10000次/分钟
- 超限响应: HTTP 429 Too Many Requests

## 2. 通用说明

### 2.1 数据模型

#### Rule 规则模型
```json
{
    "id": "string",           // 规则ID
    "name": "string",         // 规则名称
    "description": "string",  // 规则描述
    "rule_type": {           // 规则类型
        "type": "string",     // 类型名称(ip/cc/regex/sqli/xss)
        "config": {}         // 类型特定配置
    },
    "rule_variable": {       // 规则变量
        "name": "string",     // 变量名称
        "path": "string",     // 变量路径
        "type": "string"      // 变量类型
    },
    "pattern": {             // 匹配模式
        "type": "string",     // 模式类型(regex/exact/wildcard)
        "value": "string",    // 模式值
        "flags": ["string"]   // 模式标志
    },
    "action": {              // 动作
        "type": "string",     // 动作类型(block/allow/log/captcha)
        "config": {}         // 动作配置
    },
    "priority": 0,          // 优先级(1-100)
    "status": "string",     // 状态(enabled/disabled)
    "severity": "string",   // 风险级别(high/medium/low)
    "rules_operation": {    // 规则组合操作
        "type": "string",    // 操作类型(and/or/not/any/all)
        "threshold": 0,      // ANY/ALL操作的阈值
        "children": []       // 子规则ID列表
    },
    "tags": ["string"],     // 标签列表
    "extra_data": {},       // 扩展数据
    "created_at": "string", // 创建时间
    "updated_at": "string"  // 更新时间
}
```

#### 规则组合操作说明
规则引擎支持以下组合操作类型：

1. AND (与)
   - 所有子规则都匹配时才算匹配
   - 示例: `{"type": "and", "children": [1, 2, 3]}`

2. OR (或)
   - 任意子规则匹配即算匹配
   - 示例: `{"type": "or", "children": [1, 2, 3]}`

3. NOT (非)
   - 子规则不匹配时算匹配
   - 示例: `{"type": "not", "children": [1]}`

4. ANY (任意N个)
   - 至少N个子规则匹配时算匹配
   - 示例: `{"type": "any", "threshold": 2, "children": [1, 2, 3, 4]}`

5. ALL (全部N个)
   - 至少N个子规则匹配时算匹配
   - 示例: `{"type": "all", "threshold": 3, "children": [1, 2, 3, 4]}`

#### 规则组合示例
```json
{
    "name": "复合SQL注入检测",
    "description": "组合多个SQL注入规则",
    "rules_operation": {
        "type": "any",
        "threshold": 2,
        "children": [1, 2, 3, 4]
    }
}
```

#### RuleType 规则类型配置
```json
{
    "ip": {
        "ip_type": "string",     // IP类型(single/range/cidr)
        "block_type": "string",  // 封禁类型(permanent/temporary)
        "expire_time": "string"  // 过期时间
    },
    "cc": {
        "threshold": 0,          // 阈值
        "interval": 0,           // 时间窗口(秒)
        "block_time": 0         // 封禁时间(秒)
    },
    "regex": {
        "case_sensitive": false, // 大小写敏感
        "multiline": false      // 多行匹配
    },
    "sqli": {
        "detect_types": ["string"], // 检测类型列表
        "risk_level": 0            // 风险等级
    },
    "xss": {
        "detect_types": ["string"], // 检测类型列表
        "sanitize": false          // 是否净化
    }
}
```

#### RuleVariable 规则变量
```json
{
    "request_args": {        // 请求参数
        "name": "string",     // 参数名
        "in": "string"       // 参数位置(query/path)
    },
    "request_headers": {     // 请求头
        "name": "string",     // 头部名称
        "required": false    // 是否必需
    },
    "request_body": {        // 请求体
        "path": "string",     // JSON路径
        "type": "string"     // 数据类型
    },
    "request_cookies": {     // Cookie
        "name": "string",     // Cookie名
        "domain": "string"   // 域名
    },
    "response_headers": {    // 响应头
        "name": "string",     // 头部名称
        "required": false    // 是否必需
    },
    "response_body": {       // 响应体
        "path": "string",     // JSON路径
        "type": "string"     // 数据类型
    }
}
```

#### Action 动作配置
```json
{
    "block": {
        "status_code": 403,   // HTTP状态码
        "response_body": {},  // 响应体
        "log_level": "string" // 日志级别
    },
    "allow": {
        "log": false,         // 是否记录日志
        "audit": false       // 是否审计
    },
    "log": {
        "level": "string",    // 日志级别
        "format": "string",   // 日志格式
        "target": "string"   // 日志目标
    },
    "captcha": {
        "type": "string",     // 验证码类型
        "timeout": 0,        // 超时时间(秒)
        "max_tries": 0       // 最大尝试次数
    }
}
```

#### MatchResult 匹配结果
```json
{
    "matched": false,        // 是否匹配
    "rule_id": "string",     // 规则ID
    "rule_type": "string",   // 规则类型
    "block_reason": "string", // 拦截原因
    "match_details": {       // 匹配详情
        "variable": "string",  // 匹配变量
        "pattern": "string",   // 匹配模式
        "value": "string",     // 匹配值
        "position": {         // 匹配位置
            "start": 0,
            "end": 0
        }
    },
    "action_taken": "string", // 执行的动作
    "process_time": 0,       // 处理时间(ms)
    "timestamp": "string"    // 匹配时间
}
```

### 2.2 请求格式
- Content-Type: application/json
- 字符编码: UTF-8
- 时间格式: ISO8601 (YYYY-MM-DDTHH:mm:ssZ)

### 2.3 响应格式
```json
{
    "code": 0,           // 状态码，0表示成功
    "message": "string", // 响应消息
    "data": {},         // 响应数据
    "request_id": "string", // 请求追踪ID
    "timestamp": "string"   // 响应时间戳
}
```

### 2.4 错误码说明

#### 2.4.1 通用错误码
| 错误码 | 说明 | HTTP状态码 | 处理建议 |
|--------|------|------------|----------|
| 0 | 成功 | 200 | 正常处理响应数据 |
| 1001 | 参数验证失败 | 400 | 检查请求参数是否符合要求 |
| 1002 | 认证失败 | 401 | 检查Token是否有效，必要时重新获取 |
| 1003 | 权限不足 | 403 | 确认是否有接口调用权限 |
| 1004 | 资源不存在 | 404 | 检查请求的资源ID是否正确 |
| 1005 | 请求过于频繁 | 429 | 实现退避重试，建议间隔1-5秒 |
| 3001 | 系统内部错误 | 500 | 记录错误日志，联系技术支持 |

#### 2.4.2 规则管理错误码
| 错误码 | 说明 | HTTP状态码 | 处理建议 |
|--------|------|------------|----------|
| 2001 | 规则创建失败 | 400 | 检查规则格式是否正确 |
| 2002 | 规则更新失败 | 400 | 确认规则ID存在且内容有效 |
| 2003 | 规则删除失败 | 400 | 检查规则是否被其他规则引用 |
| 2004 | 规则名称重复 | 400 | 更换规则名称 |
| 2005 | 规则模式无效 | 400 | 检查pattern格式是否正确 |
| 2006 | 规则变量无效 | 400 | 确认rule_variable是否支持 |
| 2007 | 规则动作无效 | 400 | 检查action是否在允许范围内 |
| 2008 | 规则优先级冲突 | 400 | 调整规则优先级 |

#### 2.4.3 规则检查错误码
| 错误码 | 说明 | HTTP状态码 | 处理建议 |
|--------|------|------------|----------|
| 2101 | 请求体过大 | 413 | 检查请求体是否超过限制 |
| 2102 | 规则引擎繁忙 | 503 | 实现退避重试，建议指数退避 |
| 2103 | 规则匹配超时 | 504 | 检查规则复杂度，考虑优化 |
| 2104 | 规则版本过期 | 409 | 重新同步规则后重试 |
| 2105 | 规则执行错误 | 500 | 检查规则配置是否正确 |

#### 2.4.4 IP规则错误码
| 错误码 | 说明 | HTTP状态码 | 处理建议 |
|--------|------|------------|----------|
| 2201 | IP格式无效 | 400 | 检查IP地址格式 |
| 2202 | IP规则冲突 | 409 | 检查是否与现有规则冲突 |
| 2203 | IP规则过期 | 410 | 更新或删除过期规则 |

#### 2.4.5 系统配置错误码
| 错误码 | 说明 | HTTP状态码 | 处理建议 |
|--------|------|------------|----------|
| 2301 | 配置格式无效 | 400 | 检查配置格式是否正确 |
| 2302 | 配置项不存在 | 404 | 确认配置项是否支持 |
| 2303 | 配置值无效 | 400 | 检查配置值是否在允许范围内 |

#### 2.4.6 错误处理最佳实践

1. 通用错误处理
```lua
-- 错误重试示例
local function retry_request(func, max_retries, base_delay)
    local tries = 0
    while tries < max_retries do
        local res, err = func()
        if res then
            return res
        end
        
        tries = tries + 1
        if tries < max_retries then
            -- 指数退避
            local delay = base_delay * (2 ^ (tries - 1))
            ngx.sleep(delay)
        end
    end
    return nil, err
end

-- 使用示例
local res, err = retry_request(function()
    return http_request("POST", "/api/v1/rules/check", body)
end, 3, 0.1)
```

2. 规则检查错误处理
```lua
-- 规则检查示例
local function check_rules(req)
    local res, err = http_request("POST", "/api/v1/rules/check", req)
    if not res then
        return nil, err
    end
    
    if res.status == 409 then  -- 规则版本过期
        -- 先同步规则
        local ok, sync_err = sync_rules()
        if not ok then
            return nil, sync_err
        end
        -- 重试检查
        return check_rules(req)
    end
    
    if res.status == 503 then  -- 服务繁忙
        -- 实现退避重试
        return retry_request(function()
            return http_request("POST", "/api/v1/rules/check", req)
        end, 3, 0.5)
    end
    
    return res
end
```

3. 错误日志记录
```lua
-- 错误日志示例
local function handle_error(err, context)
    local log_data = {
        error_code = err.code,
        error_msg = err.message,
        request_id = context.request_id,
        timestamp = ngx.now(),
        stack_trace = debug.traceback()
    }
    
    logger.error("规则引擎错误", log_data)
    metrics.increment("rule_engine_errors", 1, {code = err.code})
end
```

4. 错误恢复策略
```lua
-- 错误恢复示例
local function handle_check_error(err, config)
    if err.code == 2102 then  -- 服务繁忙
        -- 使用本地缓存规则
        return check_rules_local_cache()
    end
    
    if err.code == 2104 then  -- 规则版本过期
        -- 强制同步规则
        sync_rules(true)
        return check_rules_local_cache()
    end
    
    if config.fail_open then
        -- 失败开放模式
        return true
    end
    
    -- 失败关闭模式
    return false
end
```

#### 2.4.7 监控告警建议

1. 错误率监控
- 监控5分钟内错误率超过1%的情况
- 监控特定错误码的出现频率
- 监控重试次数超过阈值的情况

2. 性能监控
- 监控API响应时间超过100ms的情况
- 监控规则匹配时间超过50ms的情况
- 监控并发请求数超过阈值的情况

3. 健康度监控
- 监控服务可用性
- 监控规则同步状态
- 监控资源使用情况

### 2.5 公共请求参数
| 参数名 | 类型 | 说明 |
|--------|------|------|
| page | int | 页码，从1开始 |
| size | int | 每页大小，默认20 |
| sort | string | 排序字段 |
| order | string | 排序方式(asc/desc) |
| start_time | string | 开始时间 |
| end_time | string | 结束时间 |

## 3. 接口清单

### 3.1 认证接口

#### 获取访问令牌
```http
POST /auth/token
Content-Type: application/json

Request:
{
    "username": "string",
    "password": "string"
}

Response:
{
    "access_token": "string",
    "token_type": "Bearer",
    "expires_in": 86400,
    "refresh_token": "string"
}
```

#### 刷新访问令牌
```http
POST /auth/refresh
Authorization: Bearer <refresh_token>

Response:
{
    "access_token": "string",
    "token_type": "Bearer",
    "expires_in": 86400,
    "refresh_token": "string"
}
```

### 3.2 规则管理接口

#### 创建规则
```http
POST /rules
Content-Type: application/json

Request:
{
    "name": "string",           // 规则名称
    "description": "string",    // 规则描述
    "rule_type": "string",      // 规则类型(ip/cc/regex/sqli/xss)
    "rule_variable": "string",  // 规则变量(request_args/request_body/headers)
    "pattern": "string",        // 匹配模式
    "action": "string",         // 动作(block/allow/log/captcha)
    "priority": 0,             // 优先级(1-100)
    "status": "string",        // 状态(enabled/disabled)
    "severity": "string",      // 风险级别(high/medium/low)
    "rules_operation": "string", // 规则组合操作(and/or)
    "tags": ["string"],        // 规则标签
    "extra_data": {           // 扩展数据
        "key": "value"
    }
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "rule_id": "string",
        "created_at": "string",
        "updated_at": "string"
    }
}
```

#### 批量创建规则
```http
POST /rules/batch
Content-Type: application/json

Request:
{
    "rules": [
        {
            // 规则内容，同单个规则创建
        }
    ]
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "success_count": 0,
        "fail_count": 0,
        "fail_details": [
            {
                "index": 0,
                "error": "string"
            }
        ]
    }
}
```

#### 更新规则
```http
PUT /rules/{id}
Content-Type: application/json

Request:
{
    // 同创建规则
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "updated_at": "string"
    }
}
```

#### 规则列表查询
```http
GET /rules

Parameters:
- page: 页码
- size: 每页大小
- rule_type: 规则类型
- status: 规则状态
- severity: 风险级别
- tag: 规则标签
- keyword: 关键词搜索
- sort: 排序字段
- order: 排序方式
- start_time: 开始时间
- end_time: 结束时间

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 0,
        "items": [
            {
                // 规则详情
            }
        ]
    }
}
```

#### 规则检查
```http
POST /rules/check
Content-Type: application/json

Request:
{
    "request_id": "string",     // 请求ID
    "client_ip": "string",      // 客户端IP
    "method": "string",         // 请求方法
    "uri": "string",           // 请求URI
    "headers": {               // 请求头
        "string": "string"
    },
    "args": {                  // 请求参数
        "string": "string"
    },
    "body": "string"          // 请求体
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "matched": boolean,        // 是否匹配规则
        "rule_id": "string",      // 匹配的规则ID
        "rule_type": "string",    // 规则类型
        "block_reason": "string", // 拦截原因
        "match_details": {        // 匹配详情
            "matched_pattern": "string",
            "matched_position": "string",
            "matched_value": "string",
            "matched_time": "string"
        },
        "action_taken": "string", // 执行的动作
        "process_time": 0        // 处理时间(ms)
    }
}
```

### 3.3 规则模板接口

#### 获取规则模板列表
```http
GET /rules/templates

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "templates": [
            {
                "id": "string",
                "name": "string",
                "description": "string",
                "rule_type": "string",
                "template_content": {}
            }
        ]
    }
}
```

#### 基于模板创建规则
```http
POST /rules/from-template/{template_id}
Content-Type: application/json

Request:
{
    "name": "string",
    "description": "string",
    "parameters": {
        "key": "value"
    }
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "rule_id": "string"
    }
}
```

### 3.4 监控统计接口

#### 规则匹配统计
```http
GET /metrics/rules/matches

Parameters:
- start_time: 开始时间
- end_time: 结束时间
- rule_id: 规则ID
- rule_type: 规则类型
- aggregation: 聚合维度(minute/hour/day)

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "total_matches": 0,
        "total_blocks": 0,
        "series": [
            {
                "timestamp": "string",
                "matches": 0,
                "blocks": 0
            }
        ],
        "top_rules": [
            {
                "rule_id": "string",
                "matches": 0
            }
        ]
    }
}
```

#### 性能监控
```http
GET /metrics/performance

Parameters:
- start_time: 开始时间
- end_time: 结束时间
- metrics_type: 指标类型(cpu/memory/latency)

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "current": {
            "cpu_usage": 0,
            "memory_usage": 0,
            "avg_latency": 0
        },
        "series": [
            {
                "timestamp": "string",
                "value": 0
            }
        ]
    }
}
```

### 3.5 系统管理接口

#### 获取系统状态
```http
GET /system/status

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "version": "string",
        "uptime": 0,
        "rules_count": 0,
        "rules_version": "string",
        "last_reload": "string",
        "system_load": {
            "cpu": 0,
            "memory": 0,
            "connections": 0
        }
    }
}
```

#### 系统配置更新
```http
PUT /system/config
Content-Type: application/json

Request:
{
    "mode": "string",        // 运行模式(block/alert/bypass)
    "log_level": "string",   // 日志级别
    "performance": {
        "max_connections": 0,
        "timeout": 0
    }
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "updated_at": "string"
    }
}
```

### 3.6 运行模式管理接口

#### 3.6.1 获取当前运行模式
```http
GET /api/v1/config/mode

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "mode": "block"  // block: 阻断模式 / log: 日志模式 / bypass: 旁路模式
    }
}
```

#### 3.6.2 更新运行模式
```http
PUT /api/v1/config/mode
Content-Type: application/json

Request:
{
    "mode": "block",        // 运行模式: block/log/bypass
    "reason": "string",     // 变更原因（必填）
    "description": "string" // 变更描述（选填）
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "mode": "block",
        "updated_at": 1641916800
    }
}
```

#### 3.6.3 获取模式变更日志
```http
GET /api/v1/config/mode/logs?start_time=1641916800&end_time=1641999999&page=1&size=20

Parameters:
- start_time: 开始时间戳（秒）
- end_time: 结束时间戳（秒）
- page: 页码，默认1
- size: 每页大小，默认20

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 100,
        "logs": [
            {
                "id": 1,
                "old_mode": "log",
                "new_mode": "block",
                "operator": "admin",
                "reason": "安全策略要求",
                "description": "切换到阻断模式以应对安全威胁",
                "created_at": 1641916800
            }
        ]
    }
}
```

### 3.7 运行模式说明

#### 3.7.1 模式类型
1. **阻断模式 (block)**
   - 匹配规则时直接阻断请求
   - 返回 403 状态码
   - 记录阻断日志
   - 支持自定义阻断页面

2. **日志模式 (log)**
   - 只记录日志，不阻断请求
   - 记录规则匹配情况
   - 支持多级别日志
   - 用于规则测试和审计

3. **旁路模式 (bypass)**
   - 完全放行所有请求
   - 仍然记录基础访问日志
   - 用于紧急情况或维护时

#### 3.7.2 最佳实践
1. **模式切换建议**
   - 新规则上线时先使用日志模式观察
   - 确认规则稳定后再切换到阻断模式
   - 发现异常时可临时切换到旁路模式

2. **日志记录**
   - 所有模式变更必须记录变更原因
   - 重要变更需要添加详细描述
   - 定期审计模式变更日志

3. **权限控制**
   - 限制模式变更操作权限
   - 关键模式变更需要多人审批
   - 记录所有操作人信息

4. **监控告警**
   - 监控模式变更频率
   - 异常模式切换及时告警
   - 定期同步模式状态

## 4. 最佳实践

### 4.1 错误处理
- 始终检查响应的code字段
- 对HTTP 429状态码实现退避重试
- 保存request_id用于问题追踪

### 4.2 性能优化
- 使用批量接口处理大量数据
- 实现本地缓存减少请求
- 合理设置超时时间

### 4.3 安全建议
- 定期轮换访问令牌
- 使用HTTPS传输
- 实现请求签名机制
- 敏感数据加密传输

## 5. 更新日志

### v1.1.0 (2024-01-11)
- 添加规则模板功能
- 增强监控统计能力
- 优化性能指标采集

### v1.0.0 (2024-01-01)
- 首次发布 

## 3. API接口

### 3.1 规则管理接口

#### 3.1.1 创建规则
```http
POST /api/v1/rules
Content-Type: application/json

Request:
{
    "name": "SQL注入检测",
    "type": "sqli",
    "description": "检测SQL注入攻击",
    "pattern": "(?i)(union|select|update|delete|insert|drop)",
    "action": "block",
    "severity": "high",
    "status": "enabled"
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "id": "1",
        "created_at": "2025-01-12T02:58:11+08:00"
    }
}
```

#### 3.1.2 批量创建规则
```http
POST /api/v1/rules/batch
Content-Type: application/json

Request:
{
    "rules": [
        {
            "name": "规则1",
            "type": "sqli",
            "pattern": "pattern1"
        },
        {
            "name": "规则2",
            "type": "xss",
            "pattern": "pattern2"
        }
    ]
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "success_count": 2,
        "failed_count": 0,
        "failed_rules": []
    }
}
```

#### 3.1.3 更新规则
```http
PUT /api/v1/rules/{id}
Content-Type: application/json

Request:
{
    "name": "SQL注入检测-更新",
    "description": "更新后的描述",
    "status": "disabled"
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "updated_at": "2025-01-12T02:58:11+08:00"
    }
}
```

#### 3.1.4 批量更新规则
```http
PUT /api/v1/rules/batch
Content-Type: application/json

Request:
{
    "rules": [
        {
            "id": 1,
            "status": "disabled"
        },
        {
            "id": 2,
            "status": "enabled"
        }
    ]
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "success_count": 2,
        "failed_count": 0,
        "failed_rules": []
    }
}
```

#### 3.1.5 删除规则
```http
DELETE /api/v1/rules/{id}

Response:
{
    "code": 0,
    "message": "success"
}
```

#### 3.1.6 批量删除规则
```http
DELETE /api/v1/rules/batch
Content-Type: application/json

Request:
{
    "ids": [1, 2, 3]
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "success_count": 3,
        "failed_count": 0,
        "failed_rules": []
    }
}
```

### 3.2 规则统计接口

#### 3.2.1 获取规则统计
```http
GET /api/v1/rules/stats

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "total_rules": 100,
        "enabled_rules": 80,
        "disabled_rules": 20,
        "high_risk_rules": 30,
        "medium_risk_rules": 50,
        "low_risk_rules": 20,
        "sqli_rules": 40,
        "xss_rules": 30,
        "cc_rules": 20,
        "custom_rules": 10,
        "last_day_matches": 1000,
        "last_week_matches": 5000,
        "last_month_matches": 20000,
        "total_matches": 50000
    }
}
```

#### 3.2.2 获取规则匹配统计
```http
GET /api/v1/rules/{id}/match_stats?start_time=2025-01-01&end_time=2025-01-12

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "rule_id": 1,
        "total": 1000,
        "timeline": [
            {
                "timestamp": 1641916800,
                "count": 100
            },
            {
                "timestamp": 1641920400,
                "count": 200
            }
        ]
    }
}
```

### 3.3 规则审计接口

#### 3.3.1 获取规则审计日志
```http
GET /api/v1/rules/{id}/audit_logs?start_time=2025-01-01&end_time=2025-01-12&page=1&size=20

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 100,
        "logs": [
            {
                "id": 1,
                "rule_id": 1,
                "action": "create",
                "operator": 123,
                "old_value": null,
                "new_value": "{\"name\":\"SQL注入检测\"}",
                "created_at": "2025-01-12T02:58:11+08:00"
            }
        ]
    }
}
```

### 3.4 规则导入导出接口

#### 3.4.1 导出规则
```http
GET /api/v1/rules/export?type=sqli&status=enabled

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "rules": [
            {
                "id": 1,
                "name": "SQL注入检测",
                "type": "sqli",
                "pattern": "(?i)(union|select|update|delete|insert|drop)",
                "action": "block",
                "severity": "high",
                "status": "enabled"
            }
        ]
    }
}
```

#### 3.4.2 导入规则
```http
POST /api/v1/rules/import
Content-Type: application/json

Request:
{
    "rules": [
        {
            "name": "SQL注入检测",
            "type": "sqli",
            "pattern": "(?i)(union|select|update|delete|insert|drop)",
            "action": "block",
            "severity": "high",
            "status": "enabled"
        }
    ]
}

Response:
{
    "code": 0,
    "message": "success",
    "data": {
        "success_count": 1,
        "failed_count": 0,
        "failed_rules": []
    }
}
```

### 3.5 接口调用示例

#### 3.5.1 Go语言示例
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

func createRule() error {
    rule := map[string]interface{}{
        "name":        "SQL注入检测",
        "type":        "sqli",
        "description": "检测SQL注入攻击",
        "pattern":     "(?i)(union|select|update|delete|insert|drop)",
        "action":      "block",
        "severity":    "high",
        "status":      "enabled",
    }

    body, _ := json.Marshal(rule)
    req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/rules", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer <token>")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return err
    }

    fmt.Printf("Response: %+v\n", result)
    return nil
}
```

#### 3.5.2 Python示例
```python
import requests
import json

def create_rule():
    rule = {
        "name": "SQL注入检测",
        "type": "sqli",
        "description": "检测SQL注入攻击",
        "pattern": "(?i)(union|select|update|delete|insert|drop)",
        "action": "block",
        "severity": "high",
        "status": "enabled"
    }

    headers = {
        "Content-Type": "application/json",
        "Authorization": "Bearer <token>"
    }

    response = requests.post(
        "http://localhost:8080/api/v1/rules",
        headers=headers,
        data=json.dumps(rule)
    )

    print(f"Response: {response.json()}")