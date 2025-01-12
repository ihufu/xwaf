-- 创建规则审计日志表
CREATE TABLE IF NOT EXISTS rule_audit_logs (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
    rule_id     BIGINT UNSIGNED NOT NULL COMMENT '规则ID',
    action      VARCHAR(50) NOT NULL COMMENT '操作类型(create/update/delete/import)',
    operator    BIGINT UNSIGNED NOT NULL COMMENT '操作者ID',
    old_value   TEXT COMMENT '修改前的值',
    new_value   TEXT COMMENT '修改后的值',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id),
    INDEX idx_rule_id (rule_id),
    INDEX idx_operator (operator),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='规则审计日志表';

-- 创建规则匹配统计表
CREATE TABLE IF NOT EXISTS rule_match_stats (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '统计ID',
    rule_id     BIGINT UNSIGNED NOT NULL COMMENT '规则ID',
    match_count BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '匹配次数',
    match_time  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '统计时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_rule_time (rule_id, match_time),
    INDEX idx_match_time (match_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='规则匹配统计表';

-- 创建规则统计汇总表
CREATE TABLE IF NOT EXISTS rule_stats_summary (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '汇总ID',
    total_rules       BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '规则总数',
    enabled_rules     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '启用的规则数',
    disabled_rules    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '禁用的规则数',
    high_risk_rules   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '高风险规则数',
    medium_risk_rules BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '中风险规则数',
    low_risk_rules    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '低风险规则数',
    sqli_rules        BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'SQL注入规则数',
    xss_rules         BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'XSS规则数',
    cc_rules          BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'CC规则数',
    custom_rules      BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '自定义规则数',
    last_day_matches  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '最近一天匹配次数',
    last_week_matches BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '最近一周匹配次数',
    last_month_matches BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '最近一月匹配次数',
    total_matches     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '总匹配次数',
    last_updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='规则统计汇总表';

-- 创建定时任务，每小时更新规则统计汇总表
DELIMITER //
CREATE EVENT IF NOT EXISTS update_rule_stats_summary
ON SCHEDULE EVERY 1 HOUR
DO
BEGIN
    -- 更新规则统计汇总表
    UPDATE rule_stats_summary SET
        total_rules = (SELECT COUNT(*) FROM rules),
        enabled_rules = (SELECT COUNT(*) FROM rules WHERE status = 'enabled'),
        disabled_rules = (SELECT COUNT(*) FROM rules WHERE status = 'disabled'),
        high_risk_rules = (SELECT COUNT(*) FROM rules WHERE severity = 'high'),
        medium_risk_rules = (SELECT COUNT(*) FROM rules WHERE severity = 'medium'),
        low_risk_rules = (SELECT COUNT(*) FROM rules WHERE severity = 'low'),
        sqli_rules = (SELECT COUNT(*) FROM rules WHERE type = 'sqli'),
        xss_rules = (SELECT COUNT(*) FROM rules WHERE type = 'xss'),
        cc_rules = (SELECT COUNT(*) FROM rules WHERE type = 'cc'),
        custom_rules = (SELECT COUNT(*) FROM rules WHERE type = 'custom'),
        last_day_matches = (
            SELECT COALESCE(SUM(match_count), 0)
            FROM rule_match_stats
            WHERE match_time >= DATE_SUB(NOW(), INTERVAL 1 DAY)
        ),
        last_week_matches = (
            SELECT COALESCE(SUM(match_count), 0)
            FROM rule_match_stats
            WHERE match_time >= DATE_SUB(NOW(), INTERVAL 1 WEEK)
        ),
        last_month_matches = (
            SELECT COALESCE(SUM(match_count), 0)
            FROM rule_match_stats
            WHERE match_time >= DATE_SUB(NOW(), INTERVAL 1 MONTH)
        ),
        total_matches = (SELECT COALESCE(SUM(match_count), 0) FROM rule_match_stats)
    WHERE id = 1;

    -- 如果没有记录，则插入一条
    INSERT INTO rule_stats_summary (id)
    SELECT 1
    WHERE NOT EXISTS (SELECT 1 FROM rule_stats_summary WHERE id = 1);
END //
DELIMITER ;
