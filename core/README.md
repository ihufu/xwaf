# WAF 核心模块

WAF 核心模块是基于 OpenResty/Lua 开发的 Web 应用防火墙核心防护模块，提供实时请求检测和攻击防护功能。

## 环境要求

- OpenResty 1.19.0+
- Redis 6.0+
- WAF 规则引擎服务

## 快速开始

### 1. 安装依赖
```bash
# 安装 OpenResty
wget https://openresty.org/download/openresty-1.19.3.1.tar.gz
tar -xzvf openresty-1.19.3.1.tar.gz
cd openresty-1.19.3.1
./configure
make && sudo make install

# 安装 Redis
sudo apt-get install redis-server  # Ubuntu/Debian
sudo systemctl start redis-server
```

### 2. 部署 WAF 核心
```bash
# 创建必要目录
sudo mkdir -p /usr/local/openresty/nginx/conf/lua
sudo mkdir -p /usr/local/openresty/nginx/logs/waf

# 复制 WAF 核心文件
sudo cp -r lua/* /usr/local/openresty/nginx/conf/lua/
sudo cp conf/nginx.conf /usr/local/openresty/nginx/conf/

# 设置权限
sudo chown -R nobody:nobody /usr/local/openresty/nginx/logs/waf
```

### 3. 修改配置
```bash
# 修改配置文件
sudo vim /usr/local/openresty/nginx/conf/lua/config.json

# 修改 Nginx 配置
sudo vim /usr/local/openresty/nginx/conf/nginx.conf
```

### 4. 启动服务
```bash
# 检查配置
sudo /usr/local/openresty/nginx/sbin/nginx -t

# 启动 OpenResty
sudo /usr/local/openresty/nginx/sbin/nginx

# 重新加载配置
sudo /usr/local/openresty/nginx/sbin/nginx -s reload
```

### 5. 验证安装
```bash
# 检查 WAF 状态
curl http://localhost/api/v1/config/mode

# 检查规则同步
curl http://localhost/api/v1/metrics/rules/matches
```

## 目录结构
```
waf/core/
├── lua/                # Lua 代码目录
│   ├── init.lua       # 初始化模块
│   ├── config.lua     # 配置管理
│   ├── api.lua        # API 接口
│   ├── metrics.lua    # 监控指标
│   └── logger.lua     # 日志模块
├── conf/              # 配置文件目录
│   ├── nginx.conf     # Nginx 配置
│   └── waf.conf       # WAF 配置
└── html/              # 静态文件目录
    └── block.html     # 阻断页面
```

## 功能特性

### 1. 核心模块与规则引擎的关系

WAF（Web 应用防火墙）就像一个保安，保护你的网站不受到坏人的攻击。它主要分为两个部分：核心模块和规则引擎。

-   **核心模块**：就像保安的大脑，负责接收请求，判断是否有攻击行为，并根据规则进行处理。它使用 Lua 语言编写，运行在 OpenResty 上。
-   **规则引擎**：就像保安的规则手册，负责存储和管理各种安全规则。它使用 Go 语言编写，并提供 API 接口给核心模块调用。

核心模块依赖规则引擎提供的规则进行工作。你可以把规则引擎想象成一个规则库，核心模块每次需要判断一个请求是否安全时，都会从规则引擎获取最新的规则。

### 2. 运行模式

### 1. 运行模式
- **阻断模式（Block）**：检测到攻击时直接阻断请求
- **监控模式（Alert）**：检测到攻击时仅记录告警信息
- **旁路模式（Bypass）**：不进行攻击检测，直接放行请求

### 2. 规则类型
- **IP 黑白名单**
  - 支持 IP 精确匹配和 IP 段匹配
  - 支持永久封禁和临时封禁
  - 支持过期时间设置

- **CC 防护规则**
  - 通过规则引擎提供CC防护功能
  - 支持URI级别的限制
  - 支持多种时间单位的限制
  - 支持动态调整限制策略

- **SQL 注入防护**
  - 通过规则引擎进行检测
  - 支持 URI/参数/请求体检测
  - 支持自定义匹配模式

- **XSS 攻击防护**
  - 通过规则引擎进行检测
  - 支持 HTML/JavaScript 代码检测
  - 支持事件处理器检测

- **正则表达式规则**
  - 支持自定义正则表达式
  - 支持多个检测目标（URI/参数/请求头/请求体）
  - 支持正则选项配置

- **规则组合**
  - 支持 AND/OR/NOT 逻辑组合
  - 支持多规则优先级排序
  - 支持规则状态控制

### 3. 监控指标
- **请求统计**
  - 总请求数
  - 阻断请求数
  - 告警请求数
  - 旁路请求数

- **规则匹配**
  - 规则匹配次数
  - 各类型规则匹配统计
  - 规则组合匹配统计

- **性能指标**
  - 响应时间统计（平均值/P50/P90/P99）
  - 缓存命中率
  - 规则同步状态

### 4. 健康检查
- **规则引擎检查**
  - 定期检查规则版本
  - 自动同步规则变更
  - 支持失败重试

- **状态监控**
  - 记录同步状态
  - 记录错误次数
  - 记录成功率

## 配置说明

### 1. 基础配置
```json
{
    "waf_mode": "block",  // block/alert/bypass
    "rule_engine": {
        "host": "127.0.0.1",
        "port": 8080,
        "api": "http://127.0.0.1:8080",
        "timeout": 1000
    }
}
```

### 2. Redis 配置
```json
{
    "redis": {
        "host": "127.0.0.1",
        "port": 6379,
        "password": "",
        "db": 0
    }
}
```

### 3. 日志配置
```json
{
    "log": {
        "level": "INFO",
        "dir": "logs/waf"
    }
}
```

## API 接口

### 1. 获取运行模式
```http
GET /api/v1/config/mode
```

### 2. 获取规则匹配统计
```http
GET /api/v1/metrics/rules/matches
```

### 3. 获取缓存命中率
```http
GET /api/v1/metrics/cache/hit_rate
```

### 4. 获取响应时间统计
```http
GET /api/v1/metrics/api/response_time
```

## 开发说明

### 1. 添加新规则类型
1. 在 `init.lua` 中的 `rule_types` 表中添加新规则类型
2. 实现对应的规则匹配函数
3. 在 `match_rules` 函数中添加新规则类型的处理逻辑
4. 更新监控指标统计

### 2. 修改规则匹配逻辑
1. 修改对应规则类型的匹配函数
2. 确保正确处理规则状态和优先级
3. 添加必要的日志记录
4. 更新相关监控指标

### 3. 添加新的监控指标
1. 在 `metrics.lua` 中添加新的计数器
2. 实现相应的记录函数
3. 在规则匹配时调用记录函数
4. 在指标导出时添加新指标

## 注意事项

1.  CC防护功能现已完全迁移到规则引擎,核心模块只负责调用规则引擎的API进行检查
2.  所有CC防护规则的管理都通过规则引擎的管理接口进行
3.  确保规则引擎服务正常运行,否则CC防护功能将无法使用
4.  可以通过规则引擎的API接口查看和管理CC防护规则
