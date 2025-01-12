-- 添加规则变量类型字段
ALTER TABLE rules ADD COLUMN rule_variable VARCHAR(20) NOT NULL DEFAULT 'request_args' COMMENT '规则变量类型';

-- 添加风险级别字段
ALTER TABLE rules ADD COLUMN severity VARCHAR(10) NOT NULL DEFAULT 'medium' COMMENT '风险级别';

-- 添加规则组合操作字段
ALTER TABLE rules ADD COLUMN rules_operation VARCHAR(3) NOT NULL DEFAULT 'and' COMMENT '规则组合操作(and/or)';

-- 更新规则类型字段的注释
ALTER TABLE rules MODIFY COLUMN rule_type VARCHAR(20) NOT NULL COMMENT '规则类型(xss/webshell/sql_inject等)'; 