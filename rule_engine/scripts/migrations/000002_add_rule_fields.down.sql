-- 删除新增的字段
ALTER TABLE rules DROP COLUMN rule_variable;
ALTER TABLE rules DROP COLUMN severity;
ALTER TABLE rules DROP COLUMN rules_operation;

-- 恢复规则类型字段的注释
ALTER TABLE rules MODIFY COLUMN rule_type VARCHAR(20) NOT NULL COMMENT '规则类型'; 