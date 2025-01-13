# XWAF 使用说明

## 简介
XWAF 是一个基于 OpenResty 的 Web 应用防火墙，它可以帮助你保护你的 Web 应用免受各种网络攻击。

## 架构
XWAF 由以下几个部分组成：

-   **Nginx**: 作为 Web 服务器和反向代理。
-   **Lua 模块**: 负责实现 WAF 的核心逻辑。
-   **规则引擎**: 负责存储和管理 WAF 的规则（Go语言）。

## 配置文件
WAF 的配置文件是 `waf/core/lua/config.json`。其中包含了以下配置项：

-   `rule_engine_api`: 规则引擎的 API 地址。
-   `enable_limit`: 是否启用限流。
-   `limit_rate`: 限流速率。
-   `limit_window`: 限流时间窗口。
-   `log_level`: 日志级别。
-   `log_dir`: 日志目录。
-   `block_status`: 拦截请求时的 HTTP 状态码。
-   `block_message`: 拦截请求时的错误消息。
-  `allowed_methods`: 允许的 HTTP 方法。
-   `rule_engine_timeout`: 请求规则引擎的超时时间。
-   `max_body_size`: 允许的最大请求体大小。
-   `rule_cache_ttl`: 规则缓存的 TTL。
-   `error_messages`: 错误消息配置。

## 使用方法
1.  **构建镜像:**
    ```bash
    docker-compose build waf
    ```
2.  **启动容器:**
    ```bash
    docker-compose up -d
    ```
3.  **配置规则引擎**：
    你需要启动规则引擎服务，并且配置 `config.json` 中的 `rule_engine_api` 为规则引擎的地址。目前规则引擎服务在 `docker-compose.yml` 中被注释掉了，你需要启用它。

4.  **测试 WAF**
    你可以向你的 Web 应用发送请求，WAF 将会根据配置的规则进行过滤。

## 注意事项
-  确保 `config.json` 文件存在于 `/usr/local/openresty/waf/core/lua/` 目录下。
-  你需要在 `docker-compose.yml` 文件中启用 `rule_engine` 服务，否则 WAF 无法工作。
-  你可以通过修改 `config.json` 文件来自定义 WAF 的行为。
-  WAF 的日志文件存储在 `logs/waf` 目录下。

## 功能说明

### 核心功能

1. **请求防护**
   - **SQL注入防护**: 检测和阻止SQL注入攻击
   - **XSS防护**: 防止跨站脚本攻击
   - **CC攻击防护**: 防止CC（Challenge Collapsar）攻击
   - **恶意爬虫防护**: 识别和阻止恶意爬虫
   - **敏感信息泄露防护**: 防止敏感信息外泄

2. **访问控制**
   - **IP黑白名单**: 支持IP地址的黑白名单管理
   - **地域访问控制**: 可以根据IP地理位置进行访问控制
   - **请求频率限制**: 基于IP或用户的访问频率控制
   - **自定义访问规则**: 支持自定义访问控制规则

3. **监控告警**
   - **实时监控**: 监控系统运行状态和攻击情况
   - **告警通知**: 支持邮件、短信等多种告警方式
   - **日志分析**: 提供详细的访问日志和攻击日志分析

### 规则引擎功能说明

#### 1. 规则管理

##### 1.1 基础规则管理
- **创建规则**
  ```json
  POST /api/v1/rules
  {
    "name": "SQL注入检测",
    "type": "sqli",
    "description": "检测SQL注入攻击",
    "pattern": "(?i)(union|select|update|delete|insert|drop)",
    "action": "block",
    "severity": "high",
    "status": "enabled"
  }
  ```

- **更新规则**
  ```json
  PUT /api/v1/rules/{id}
  {
    "name": "SQL注入检测-更新",
    "description": "更新后的描述",
    "status": "disabled"
  }
  ```

- **删除规则**
  ```json
  DELETE /api/v1/rules/{id}
  ```

- **查询规则**
  ```json
  GET /api/v1/rules?type=sqli&status=enabled&keyword=注入&page=1&page_size=10
  ```

##### 1.2 批量操作
- **批量创建规则**
  ```json
  POST /api/v1/rules/batch
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
  ```

- **批量更新规则**
  ```json
  PUT /api/v1/rules/batch
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
  ```

- **批量删除规则**
  ```json
  DELETE /api/v1/rules/batch
  {
    "ids": [1, 2, 3]
  }
  ```

##### 1.3 导入导出
- **导出规则**
  ```json
  GET /api/v1/rules/export?type=sqli&status=enabled
  ```

- **导入规则**
  ```json
  POST /api/v1/rules/import
  {
    "rules": [/* 规则JSON数组 */]
  }
  ```

#### 2. 规则统计

##### 2.1 规则总体统计
```json
GET /api/v1/rules/stats

Response:
{
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
```

##### 2.2 规则匹配统计
```json
GET /api/v1/rules/{id}/match_stats?start_time=2025-01-01&end_time=2025-01-12

Response:
{
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
```

#### 3. 规则审计

##### 3.1 审计日志查询
```json
GET /api/v1/rules/{id}/audit_logs

Response:
{
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
```

#### 4. 配置示例

##### 4.1 规则引擎配置
```yaml
rule_engine:
  # MySQL配置
  mysql:
    host: "localhost"
    port: 3306
    user: "root"
    password: "password"
    database: "waf"
    charset: "utf8mb4"

  # Redis配置（用于规则缓存）
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 0

  # 服务配置
  server:
    host: "0.0.0.0"
    port: 8080
    read_timeout: 5s
    write_timeout: 5s

  # 日志配置
  log:
    level: "info"
    path: "/var/log/waf"
    max_size: 100  # MB
    max_age: 7     # 天
    max_backups: 5

  # WAF配置
  waf:
    mode: "block"        # block/monitor
    cache_ttl: 300       # 规则缓存时间（秒）
    sync_interval: 60    # 规则同步间隔（秒）
    max_request_size: 10 # 最大请求体大小（MB）
```

#### 5. API文档

##### 5.1 规则管理API
| 方法   | 路径                    | 描述         |
|--------|------------------------|--------------|
| POST   | /api/v1/rules         | 创建规则      |
| PUT    | /api/v1/rules/{id}    | 更新规则      |
| DELETE | /api/v1/rules/{id}    | 删除规则      |
| GET    | /api/v1/rules/{id}    | 获取规则      |
| GET    | /api/v1/rules         | 查询规则列表   |
| POST   | /api/v1/rules/batch   | 批量创建规则   |
| PUT    | /api/v1/rules/batch   | 批量更新规则   |
| DELETE | /api/v1/rules/batch   | 批量删除规则   |
| GET    | /api/v1/rules/export  | 导出规则      |
| POST   | /api/v1/rules/import  | 导入规则      |

##### 5.2 规则统计API
| 方法   | 路径                              | 描述           |
|--------|----------------------------------|----------------|
| GET    | /api/v1/rules/stats              | 获取规则统计    |
| GET    | /api/v1/rules/{id}/match_stats   | 获取匹配统计    |

##### 5.3 规则审计API
| 方法   | 路径                              | 描述           |
|--------|----------------------------------|----------------|
| GET    | /api/v1/rules/{id}/audit_logs    | 获取审计日志    |

#### 6. 错误码说明

| 错误码  | 描述                    | 处理建议                    |
|--------|------------------------|----------------------------|
| 400    | 请求参数错误            | 检查请求参数是否符合要求      |
| 401    | 未授权                 | 检查认证信息是否正确          |
| 403    | 无权限                 | 确认是否有操作权限            |
| 404    | 资源不存在             | 检查请求的资源是否存在        |
| 409    | 资源冲突               | 检查是否有重复的规则名称      |
| 500    | 服务器内部错误          | 查看服务器日志并联系管理员    |

#### 7. 最佳实践

1. **规则管理**
   - 按照业务场景对规则进行分组
   - 为规则添加清晰的描述
   - 定期审查和更新规则
   - 使用批量操作提高效率

2. **性能优化**
   - 合理使用规则缓存
   - 避免过于复杂的正则表达式
   - 定期清理无用的规则
   - 使用合适的规则优先级

3. **监控告警**
   - 配置关键指标的告警阈值
   - 定期分析规则匹配统计
   - 关注异常的规则触发
   - 及时处理审计日志

4. **安全建议**
   - 定期备份规则配置
   - 严格控制规则管理权限
   - 记录所有规则变更
   - 进行定期安全审计

### 使用指南

1. **安装部署**
   ```bash
   # 克隆项目
   git clone <项目地址>
   cd xwaf

   # 构建镜像
   docker-compose build waf

   # 启动服务
   docker-compose up -d
   ```

2. **配置说明**
   - **基础配置**
     ```json
     {
       "rule_engine_api": "http://rule-engine:8080",  // 规则引擎API地址
       "enable_limit": true,                          // 是否启用限流
       "limit_rate": 100,                            // 限流速率（每秒请求数）
       "limit_window": 60,                           // 限流时间窗口（秒）
       "log_level": "info",                          // 日志级别（debug/info/warn/error）
       "block_status": 403                           // 拦截状态码
     }
     ```

   - **规则配置**
     ```json
     {
       "rules": [
         {
           "id": "sql_injection",
           "type": "regex",
           "pattern": "(?i)(select|union|update|delete|insert|table|from|ascii|hex|unhex|drop)",
           "action": "block"
         }
       ]
     }
     ```

3. **管理平台使用**
   - 访问地址：`http://<your-domain>:8080`
   - 默认管理员账号：admin
   - 默认密码：admin123

4. **日志查看**
   - 访问日志：`logs/waf/access.log`
   - 错误日志：`logs/waf/error.log`
   - 攻击日志：`logs/waf/attack.log`

5. **常见问题**
   - **问题1：WAF无法启动**
     - 检查配置文件是否正确
     - 确保规则引擎服务正常运行
     - 查看错误日志获取详细信息

   - **问题2：规则不生效**
     - 检查规则配置是否正确
     - 确认规则是否已启用
     - 查看规则匹配日志

   - **问题3：误拦截**
     - 调整规则匹配条件
     - 添加白名单
     - 降低规则匹配严格度

### 性能优化建议

1. **系统配置优化**
   - 调整worker进程数
   - 优化内存使用
   - 配置合适的缓存策略

2. **规则优化**
   - 减少规则数量
   - 优化规则匹配顺序
   - 使用高效的正则表达式

3. **监控优化**
   - 设置合理的日志级别
   - 配置日志轮转
   - 使用专门的日志分析系统

## 改进方向
-   增加更多的 WAF 规则。
-   支持更多的日志格式。
-   支持动态更新 WAF 配置。
-   增加对 Web 攻击的监控和报警功能。

## 未来功能

我将在这里添加关于我提供的所有功能的详细说明，包括：

-   **功能名称**: 功能的用途、使用方法、参数说明、返回值说明等。
-   **参数说明**: 每个参数的名称、类型、是否必选、默认值、取值范围等。
-   **返回值说明**: 返回值的类型、含义、示例等。

请持续关注此文档的更新。