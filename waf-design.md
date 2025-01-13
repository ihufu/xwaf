# WAF 防火墙设计方案

## 1. 系统架构

### 1.1 整体架构
- **客户端:** 用户通过浏览器或 API 客户端访问 WAF 防火墙
- **反向代理:** Nginx 作为反向代理服务器，接收客户端请求，并将请求转发给 WAF 核心
- **WAF 核心:** OpenResty + ngx_lua 实现 WAF 核心功能，包括请求预处理、规则引擎调用、结果处理和日志记录
- **规则引擎:** Go 语言开发的规则引擎，负责加载、匹配和管理安全规则
- **管理平台:** Go 语言开发的管理平台，提供规则管理、WAF 配置、日志查看等功能
- **数据库:** Redis 用于规则缓存，MySQL 用于存储配置、用户数据和管理平台数据，Elasticsearch (可选) 用于日志存储和分析

### 1.2 模块划分
- **WAF 核心模块:**
    - 请求预处理模块
    - 规则引擎调用模块
    - 结果处理模块
    - 日志记录模块
    - 限流和频率控制模块
- **规则引擎模块:**
    - 规则加载模块
    - 规则匹配模块
    - API 接口模块
    - 缓存模块
    - 自定义规则模块
- **管理平台模块:**
    - 规则管理模块
    - WAF 配置模块
    - 日志查看模块
    - IP 黑白名单管理模块
    - 用户管理模块
    - 角色管理模块
    - 审计日志模块
- **日志和存储模块:**
    - 访问日志模块
    - 规则匹配日志模块
    - 操作日志模块
    - 本地文件系统存储模块
    - 数据库存储模块
    - 专门日志系统存储模块

## 2. 模块设计

### 2.1 WAF 核心模块
- **请求预处理模块:**
    - 解析请求头、URI、方法等
    - 提取必要信息，如 IP 地址、User-Agent 等
- **规则引擎调用模块:**
    - 通过 HTTP 请求调用规则引擎 API
    - 传递请求信息给规则引擎
- **结果处理模块:**
    - 根据规则引擎返回的结果，执行相应操作（放行、拦截、重定向）
    - 记录访问日志
- **日志记录模块:**
    - 记录访问日志到文件或数据库
    - 包含时间、IP、URI、方法、状态码等信息
- **限流和频率控制模块:**
    - 基于 IP 地址或用户进行限流
    - 使用 Nginx 共享内存实现

### 2.2 规则引擎模块
- **规则加载模块:**
    - 从数据库（Redis/MySQL）加载安全规则
    - 支持正则表达式、IP 黑白名单、CC 防御、SQL 注入、XSS 等规则
    - 支持规则版本控制和回滚
    - 记录规则同步状态
- **规则匹配模块:**
    - 对请求信息进行规则匹配
    - 返回匹配的规则 ID 和动作
    - 记录规则匹配统计信息
- **API 接口模块:**
    - 提供 RESTful API 接口供 WAF 核心调用
    - 支持规则热更新
    - 提供规则版本管理接口
- **缓存模块:**
    - 使用 Redis 缓存规则，减少数据库访问
    - 实现缓存自动更新机制
    - 支持缓存健康检查和自动重连
    - 监控缓存命中率
- **自定义规则模块:**
    - 支持用户自定义规则
    - 提供规则测试和验证功能
- **监控指标模块:**
    - 收集规则匹配次数和命中率
    - 监控缓存性能指标
    - 跟踪API响应时间
    - 支持指标数据导出和分析

### 2.3 管理平台模块
- **规则管理模块:**
    - 添加、修改、删除安全规则
    - 配置规则的启用/禁用状态
- **WAF 配置模块:**
    - 配置 WAF 核心的拦截模式、监控模式等
- **日志查看模块:**
    - 查看访问日志、规则匹配日志和操作日志
- **IP 黑白名单管理模块:**
    - 添加、删除 IP 黑白名单
- **用户管理模块:**
    - 用户注册、登录、权限管理
- **角色管理模块:**
    - 创建、修改、删除角色
    - 分配角色权限
- **审计日志模块:**
    - 记录用户操作日志

### 2.4 日志和存储模块
- **访问日志模块:**
    - 记录 WAF 核心的访问日志
    - 包含时间、IP、URI、方法、状态码等信息
- **规则匹配日志模块:**
    - 记录规则引擎的匹配日志
    - 包含匹配的规则 ID、动作等信息
- **操作日志模块:**
    - 记录管理平台的用户操作日志
    - 包含操作时间、用户、操作类型等信息
- **本地文件系统存储模块:**
    - 将日志记录到本地文件
    - 使用 logrotate 进行日志切割和备份
- **数据库存储模块:**
    - 将日志记录到数据库 (MySQL 或 Elasticsearch)
- **专门日志系统存储模块:**
    - 使用 ELK (Elasticsearch, Logstash, Kibana) 进行日志存储和分析

## 3. 数据流

### 3.1 请求处理流程
1.  客户端发送请求到 Nginx 反向代理服务器
2.  Nginx 将请求转发给 WAF 核心 (OpenResty + ngx_lua)
3.  WAF 核心进行请求预处理，提取必要信息
4.  WAF 核心通过 HTTP 请求调用规则引擎 API
5.  规则引擎进行规则匹配，返回匹配结果
6.  WAF 核心根据规则引擎返回的结果，执行相应操作（放行、拦截、重定向）
7.  WAF 核心记录访问日志
8.  如果请求被放行，则将请求转发给后端服务器
9.  后端服务器返回响应给 WAF 核心
10. WAF 核心将响应返回给客户端

### 3.2 规则管理流程
1.  管理员通过管理平台 Web 界面进行规则管理
2.  管理平台通过 API 调用规则引擎，添加、修改、删除规则
3.  规则引擎将规则存储到数据库 (MySQL)
4.  规则引擎将规则缓存到 Redis
5.  WAF 核心从 Redis 加载规则

### 3.3 日志处理流程
1.  WAF 核心记录访问日志到文件或数据库
2.  规则引擎记录规则匹配日志到文件或数据库
3.  管理平台记录操作日志到文件或数据库
4.  日志存储模块将日志存储到本地文件系统、数据库或专门日志系统

## 4. 安全设计

### 4.1 输入验证
- 对所有输入数据进行验证，防止恶意输入
- 使用正则表达式验证输入格式
- 对用户提供的字符串进行转义，防止 XSS 攻击

### 4.2 访问控制
- 使用 RBAC 进行权限控制
- 对 API 接口进行身份验证和授权
- 限制用户对管理平台的访问权限

### 4.3 数据加密
- 对敏感数据进行加密存储
- 使用 HTTPS 加密传输数据
- 使用 JWT 或 OAuth2 进行用户认证

### 4.4 安全审计
- 记录所有用户操作日志
- 定期进行安全审计
- 监控系统运行状态

## 5. 性能设计

### 5.1 缓存
- 使用 Redis 缓存规则，减少数据库访问
- 使用 Nginx 共享内存实现限流和频率控制
- 使用 CDN 缓存静态资源

### 5.2 异步处理
- 使用 Lua 协程处理异步操作
- 使用消息队列处理日志记录

### 5.3 负载均衡
- 使用 Nginx 进行负载均衡
- 使用 Docker Swarm 或 Kubernetes 进行容器编排

### 5.4 性能优化
- 使用 `local` 变量
- 避免全局变量
- 优化数据库查询
- 优化正则表达式匹配

## 6. 数据库设计

### 6.1 MySQL表设计
- **规则表 (rules):**
    ```sql
    CREATE TABLE rules (
        id INT PRIMARY KEY AUTO_INCREMENT,
        group_id INT NOT NULL,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        rule_type ENUM('regex', 'ip', 'cc', 'sql', 'xss') NOT NULL,
        pattern TEXT NOT NULL,
        action ENUM('allow', 'block', 'redirect', 'log') NOT NULL,
        priority INT DEFAULT 100,
        status ENUM('enabled', 'disabled') NOT NULL,
        created_by INT NULL,
        updated_by INT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    );
    ```

- **规则版本表 (rule_versions):**
    ```sql
    CREATE TABLE rule_versions (
        id INT PRIMARY KEY AUTO_INCREMENT,
        rule_id INT NOT NULL,
        version INT NOT NULL,
        content TEXT NOT NULL,
        status ENUM('active', 'archived') NOT NULL,
        created_by INT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (rule_id) REFERENCES rules(id)
    );
    ```

- **规则同步记录表 (rule_sync_logs):**
    ```sql
    CREATE TABLE rule_sync_logs (
        id INT PRIMARY KEY AUTO_INCREMENT,
        rule_id INT NOT NULL,
        version INT NOT NULL,
        status ENUM('success', 'failed') NOT NULL,
        message TEXT,
        sync_type ENUM('create', 'update', 'delete') NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (rule_id) REFERENCES rules(id)
    );
    ```

- **CC防御规则表 (cc_rules):**
    ```sql
    CREATE TABLE cc_rules (
        id INT PRIMARY KEY AUTO_INCREMENT,
        uri VARCHAR(255) NOT NULL,
        limit_rate INT NOT NULL,
        time_window INT NOT NULL,
        limit_unit ENUM('second', 'minute') NOT NULL,
        status ENUM('enabled', 'disabled') NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    );
    ```

### 6.2 Redis缓存设计
- **规则缓存:**
    - Key格式: `rule:<id>`
    - 字段: `name`, `pattern`, `action`, `priority`, `status`
    - 过期时间: 5分钟

- **CC防御规则缓存:**
    - Key格式: `cc_rule:<uri>`
    - 字段: `limit_rate`, `time_window`, `status`
    - 过期时间: 1分钟

- **规则版本缓存:**
    - Key格式: `rule_version:<rule_id>`
    - 字段: `version`, `content`, `status`
    - 过期时间: 10分钟

### 6.3 监控指标存储
- **规则匹配计数器:**
    - Key格式: `counter:rule:<rule_id>:matches`
    - 类型: String (计数器)
    - 统计周期: 1小时

- **缓存命中率:**
    - Key格式: `stats:cache:hit_rate`
    - 类型: Hash
    - 字段: `hits`, `misses`
    - 统计周期: 5分钟

- **API响应时间:**
    - Key格式: `stats:api:response_time`
    - 类型: List
    - 记录最近100次请求的响应时间
    - 统计周期: 实时