# WAF 配置说明

## 1. 规则引擎配置 (rule_engine)
```json
{
    "rule_engine": {
        "host": "http://127.0.0.1:8080",  // 规则引擎主机地址
        "timeout": 1000,                   // 请求超时时间(毫秒)
        "fail_open": false,                // 引擎失败时是否放行
        "sync_interval": 60,               // 规则同步间隔(秒)
        "max_retries": 3,                  // 最大重试次数
        "retry_interval": 1                // 重试间隔(秒)
    }
}
```

### WAF运行模式说明
WAF运行模式由规则引擎统一管理，支持以下模式：
- **block**: 阻断模式，匹配规则时直接阻断请求
- **log**: 日志模式，只记录日志，不阻断请求
- **bypass**: 旁路模式，完全放行所有请求

可通过规则引擎的API进行查看和管理：
- 获取当前模式: GET /api/v1/config/mode
- 更新运行模式: PUT /api/v1/config/mode
- 查看变更日志: GET /api/v1/config/mode/logs

## 2. 速率限制配置 (rate_limit)
```json
{
    "rate_limit": {
        "enable": true,           // 是否启用速率限制
        "rate": 100,             // 请求速率限制(次/分钟)
        "window": 60,            // 统计窗口(秒)
        "burst": 50,             // 突发请求容量
        "block_time": 300        // 封禁时间(秒)
    }
}
```

## 3. Redis配置 (redis)
```json
{
    "redis": {
        "host": "127.0.0.1",              // Redis主机地址
        "port": 6379,                     // Redis端口
        "password": "",                   // Redis密码
        "db": 0,                         // Redis数据库编号
        "timeout": 1000,                 // 连接超时时间(毫秒)
        "pool_size": 100,                // 连接池大小
        "pool_timeout": 1000,            // 连接池超时时间(毫秒)
        "keepalive_timeout": 60000,      // 连接保活时间(毫秒)
        "keepalive_pool_size": 50,       // 保活连接池大小
        "connection_backlog": 1024        // 连接队列大小
    }
}
```

## 4. 日志配置 (log)
```json
{
    "log": {
        "level": "INFO",                // 日志级别(DEBUG/INFO/WARN/ERROR)
        "dir": "logs/waf",             // 日志目录路径
        "max_size": 104857600,         // 单个日志文件最大大小(字节)
        "max_files": 10,               // 最大保留文件数
        "buffer_size": 4096            // 日志缓冲区大小(字节)
    }
}
```

## 5. 缓存配置 (cache)
```json
{
    "cache": {
        "ttl": 3600,                   // 缓存过期时间(秒)
        "local_capacity": "10m",        // 本地缓存容量
        "negative_ttl": 60             // 负缓存过期时间(秒)
    }
}
```

## 6. 阻断配置 (block)
```json
{
    "block": {
        "status": 403,                 // 阻断状态码
        "message": "Request blocked by WAF",  // 阻断提示信息
        "template": "block.html",      // 阻断页面模板
        "template_dir": "templates"    // 模板目录
    }
}
```

## 7. 告警配置 (alert)
```json
{
    "alert": {
        "log_level": "WARN",          // 告警日志级别
        "notify": false,              // 是否发送告警通知
        "notify_url": "",             // 告警通知地址
        "notify_interval": 300        // 通知间隔(秒)
    }
}
```

## 8. 请求配置 (request)
```json
{
    "request": {
        "allowed_methods": [          // 允许的HTTP请求方法
            "GET", "POST", "PUT", 
            "DELETE", "HEAD", "OPTIONS"
        ],
        "max_body_size": 10485760,    // 请求体大小限制(字节)
        "max_uri_length": 4096,       // URI长度限制(字节)
        "max_header_size": 8192,      // 请求头大小限制(字节)
        "max_args": 100               // 最大参数个数
    }
}
```

## 9. 错误码配置 (error_codes)
详细的错误码定义请参考 [错误处理文档](../docs/error_handling.md)

## 10. 配置最佳实践
1. **性能调优**
   - 根据服务器配置调整连接池大小
   - 合理设置缓存容量和TTL
   - 适当配置日志级别和缓冲区

2. **安全建议**
   - 设置合适的请求限制
   - 启用必要的安全检查
   - 配置告警通知

3. **运维建议**
   - 定期同步规则配置
   - 监控错误日志
   - 及时更新配置文件

## 11. 环境变量
- `XWAF_CONFIG_PATH`: 配置文件路径
- `XWAF_LOG_LEVEL`: 日志级别
- `XWAF_MODE`: 运行模式