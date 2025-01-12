# WAF 规则引擎

WAF 规则引擎是一个基于 Go 语言开发的 Web 应用防火墙规则管理系统，提供规则的存储、管理、匹配和同步功能。

## 功能特性

- 支持多种规则类型：
  - 正则表达式规则
  - IP 黑白名单规则
  - CC 防护规则
  - SQL 注入规则
  - XSS 攻击规则

- 高效的规则匹配：
  - Trie树URL路径匹配
  - AC自动机多模式匹配
  - 正则表达式优化
  - 并行匹配处理

- 灵活的规则组合：
  - AND (与) 组合
  - OR (或) 组合
  - NOT (非) 操作
  - ANY (任意N个) 匹配
  - ALL (全部N个) 匹配

- 规则管理功能：
  - 规则的增删改查
  - 规则版本控制
  - 规则热更新
  - 规则同步状态监控

- 高性能设计：
  - Redis 缓存加速
  - 规则优先级排序
  - 高效的匹配算法
  - 并发匹配处理

- 完善的监控：
  - 规则匹配统计
  - 缓存命中率
  - API 响应时间
  - Prometheus 指标导出

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.2+
- Docker & Docker Compose

### 安装部署

1. 克隆代码：

```bash
git clone https://github.com/xwaf/rule_engine.git
cd rule_engine
```

2. 修改配置：

编辑 `configs/config.yaml` 文件，根据实际环境修改 MySQL、Redis 等配置。

3. 启动服务：

```bash
docker-compose up -d
```

4. 检查服务状态：

```bash
docker-compose ps
```

### API 接口

#### 规则检查接口

```http
POST /api/v1/check
Content-Type: application/json

{
    "request_id": "string",    // 请求ID
    "client_ip": "string",     // 客户端IP
    "method": "string",        // HTTP方法
    "uri": "string",          // 请求URI
    "headers": {              // 请求头
        "string": "string"
    },
    "args": {                 // 请求参数
        "string": "string"
    },
    "body": "string"          // 请求体
}
```

#### 规则管理接口

- 创建规则：`POST /api/v1/rules`
- 更新规则：`PUT /api/v1/rules/{id}`
- 删除规则：`DELETE /api/v1/rules/{id}`
- 获取规则：`GET /api/v1/rules/{id}`
- 规则列表：`GET /api/v1/rules?page={page}&size={size}`
- 重新加载：`POST /api/v1/rules/reload`

#### 监控接口

- 规则匹配统计：`GET /api/v1/metrics/rules/matches`
- 缓存命中率：`GET /api/v1/metrics/cache/hit_rate`
- API响应时间：`GET /api/v1/metrics/api/response_time`

### 目录结构

```
waf/rule_engine/
├── api/            # API接口层
│   ├── handler/    # 请求处理器
│   └── router/     # 路由配置
├── cmd/            # 程序入口
│   └── server/     # 服务启动
├── configs/        # 配置文件
├── internal/       # 内部实现
│   ├── config/     # 配置管理
│   ├── model/      # 数据模型
│   ├── repository/ # 数据访问层
│   └── service/    # 业务逻辑层
├── pkg/            # 公共包
│   ├── logger/     # 日志工具
│   └── metrics/    # 监控指标
└── scripts/        # 脚本文件
```

## 开发说明

### 添加新规则类型

1. 在 `internal/model/rule.go` 中添加新的规则类型常量
2. 在 `internal/service/rule.go` 中实现规则匹配逻辑
3. 更新数据库表结构（如果需要）
4. 添加相应的测试用例

### 监控指标扩展

1. 在 `pkg/metrics/metrics.go` 中定义新的指标
2. 在相应的代码位置记录指标
3. 在 Prometheus 中配置指标采集
4. 更新 Grafana 面板（如果使用）

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交代码
4. 创建 Pull Request

## 许可证

MIT License

## 新增功能

### 1. 批量规则管理

- **批量创建规则**：支持一次创建多个规则，提高规则管理效率
- **批量更新规则**：支持一次更新多个规则的状态或配置
- **批量删除规则**：支持一次删除多个规则
- **规则导入导出**：支持规则的批量导入和导出，方便规则迁移和备份

### 2. 规则统计分析

- **规则总体统计**
  - 规则总数统计
  - 各类型规则数量统计
  - 各风险级别规则统计
  - 启用/禁用规则统计

- **规则匹配统计**
  - 规则匹配次数统计
  - 规则匹配时间线分析
  - 最近一天/一周/一月匹配统计
  - 总匹配次数统计

### 3. 规则审计功能

- **操作审计**
  - 记录规则的创建、更新、删除操作
  - 记录操作人员信息
  - 记录操作时间
  - 记录变更内容

- **审计日志查询**
  - 支持按时间范围查询
  - 支持按操作类型查询
  - 支持按规则ID查询
  - 支持分页查询

### 4. 性能优化

- **规则缓存优化**
  - 使用Redis缓存热点规则
  - 支持规则缓存自动更新
  - 支持缓存过期策略
  - 支持缓存预热

- **规则匹配优化**
  - 优化正则表达式性能
  - 支持规则优先级排序
  - 支持规则并行匹配
  - 支持规则提前终止

### 5. 监控告警

- **性能监控**
  - API响应时间监控
  - 规则匹配时间监控
  - 缓存命中率监控
  - 系统资源使用监控

- **异常告警**
  - 规则匹配异常告警
  - 系统错误告警
  - 性能瓶颈告警
  - 资源使用告警

### 6. 使用示例

#### 6.1 批量创建规则
```bash
curl -X POST http://localhost:8080/api/v1/rules/batch \
  -H "Content-Type: application/json" \
  -d '{
    "rules": [
      {
        "name": "SQL注入检测",
        "type": "sqli",
        "pattern": "(?i)(union|select|update|delete|insert|drop)",
        "action": "block",
        "severity": "high"
      },
      {
        "name": "XSS攻击检测",
        "type": "xss",
        "pattern": "(?i)(<script|javascript:)",
        "action": "block",
        "severity": "high"
      }
    ]
  }'
```

#### 6.2 查看规则统计
```bash
curl http://localhost:8080/api/v1/rules/stats
```

#### 6.3 查看规则审计日志
```bash
curl "http://localhost:8080/api/v1/rules/1/audit_logs?start_time=2025-01-01&end_time=2025-01-12"
```

### 7. 最佳实践

1. **规则管理**
   - 按业务场景对规则分组
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

### 规则配置示例

1. 简单规则
```json
{
    "name": "SQL注入检测",
    "type": "sqli",
    "pattern": "(?i)(union[\\s\\(\\+]+select)",
    "action": "block"
}
```

2. 组合规则
```json
{
    "name": "复合攻击检测",
    "rules_operation": {
        "type": "any",
        "threshold": 2,
        "children": [
            {"id": 1, "type": "sqli"},
            {"id": 2, "type": "xss"},
            {"id": 3, "type": "path"}
        ]
    },
    "action": "block"
}
```

3. 条件组合
```json
{
    "name": "高级防护规则",
    "rules_operation": {
        "type": "and",
        "children": [
            {"id": 1, "type": "ip"},
            {
                "type": "or",
                "children": [
                    {"id": 2, "type": "sqli"},
                    {"id": 3, "type": "xss"}
                ]
            }
        ]
    },
    "action": "block"
}
```
