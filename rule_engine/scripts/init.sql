-- 创建数据库
CREATE DATABASE IF NOT EXISTS waf DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE waf;

-- 修改rules表的rule_type字段为type
ALTER TABLE rules CHANGE COLUMN rule_type type VARCHAR(50) NOT NULL COMMENT '规则类型';

-- 创建规则表
CREATE TABLE IF NOT EXISTS rules (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '规则ID',
    group_id        BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '规则组ID',
    name            VARCHAR(255)     NOT NULL COMMENT '规则名称',
    description     TEXT            COMMENT '规则描述',
    type            VARCHAR(50)      NOT NULL COMMENT '规则类型',
    rule_variable   VARCHAR(50)      NOT NULL COMMENT '规则变量类型',
    pattern         VARCHAR(255)     NOT NULL COMMENT '匹配模式',
    action          VARCHAR(50)      NOT NULL COMMENT '动作',
    priority        INT             NOT NULL DEFAULT 0 COMMENT '优先级',
    status          VARCHAR(50)      NOT NULL DEFAULT 'enabled' COMMENT '状态',
    severity        VARCHAR(50)      NOT NULL DEFAULT 'medium' COMMENT '风险级别',
    rules_operation VARCHAR(10)      NOT NULL DEFAULT 'and' COMMENT '规则组合操作',
    version         BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '版本号',
    hash            VARCHAR(32)      NOT NULL DEFAULT '' COMMENT '规则哈希',
    created_by      BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建者',
    updated_by      BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新者',
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    INDEX idx_group_id (group_id),
    INDEX idx_status (status),
    INDEX idx_priority (priority),
    INDEX idx_version (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='规则表';

-- 创建规则版本表
CREATE TABLE IF NOT EXISTS rule_versions (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '版本ID',
    rule_id     BIGINT UNSIGNED NOT NULL COMMENT '规则ID',
    version     BIGINT UNSIGNED NOT NULL COMMENT '版本号',
    hash        VARCHAR(32)      NOT NULL COMMENT '内容哈希值',
    content     TEXT            NOT NULL COMMENT '规则内容',
    change_type VARCHAR(50)      NOT NULL COMMENT '变更类型',
    status      VARCHAR(50)      NOT NULL DEFAULT 'enabled' COMMENT '状态',
    created_by  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建者',
    created_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id),
    INDEX idx_rule_id (rule_id),
    INDEX idx_version (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='规则版本表';

-- 创建规则同步日志表
CREATE TABLE IF NOT EXISTS rule_sync_logs (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
    rule_id     BIGINT UNSIGNED NOT NULL COMMENT '规则ID',
    version     BIGINT UNSIGNED NOT NULL COMMENT '版本号',
    status      VARCHAR(50)      NOT NULL COMMENT '同步状态',
    message     TEXT            COMMENT '详细信息',
    sync_type   VARCHAR(50)      NOT NULL COMMENT '同步类型',
    created_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id),
    INDEX idx_rule_id (rule_id),
    INDEX idx_version (version),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='规则同步日志表';

-- 创建规则更新事件表
CREATE TABLE IF NOT EXISTS rule_update_events (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '事件ID',
    version     BIGINT UNSIGNED NOT NULL COMMENT '更新版本号',
    changes     TEXT NOT NULL COMMENT '变更列表(JSON)',
    status      VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '事件状态',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_version (version),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='规则更新事件表';

-- 创建CC防护规则表
CREATE TABLE IF NOT EXISTS cc_rules (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '规则ID',
    uri         VARCHAR(200) NOT NULL COMMENT '请求URI',
    limit_rate  INT NOT NULL COMMENT '限制速率',
    time_window INT NOT NULL COMMENT '时间窗口',
    limit_unit  VARCHAR(20) NOT NULL COMMENT '限制单位(second/minute/hour)',
    status      VARCHAR(20) NOT NULL DEFAULT 'enabled' COMMENT '状态(enabled/disabled)',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_uri (uri),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CC防护规则表';

-- 创建IP规则表
CREATE TABLE IF NOT EXISTS ip_rules (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '规则ID',
    ip          VARCHAR(50) NOT NULL COMMENT 'IP地址',
    ip_type     VARCHAR(20) NOT NULL COMMENT 'IP类型(white/black)',
    block_type  VARCHAR(20) NOT NULL COMMENT '封禁类型(permanent/temporary)',
    expire_time TIMESTAMP NULL COMMENT '过期时间',
    description TEXT COMMENT '规则描述',
    created_by  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建者',
    updated_by  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新者',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_ip (ip),
    INDEX idx_ip_type (ip_type),
    INDEX idx_block_type (block_type),
    INDEX idx_expire_time (expire_time),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='IP规则表';

-- 创建WAF配置表
CREATE TABLE IF NOT EXISTS waf_configs (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '配置ID',
    mode        VARCHAR(20) NOT NULL DEFAULT 'block' COMMENT 'WAF运行模式(block/monitor)',
    description TEXT COMMENT '配置描述',
    created_by  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建者',
    updated_by  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新者',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    INDEX idx_mode (mode),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='WAF配置表'; 