package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
)

// ccRuleRepository CC规则MySQL仓储实现
type ccRuleRepository struct {
	db *sql.DB
}

// NewCCRuleRepository 创建CC规则仓储
func NewCCRuleRepository(db *sql.DB) repository.CCRuleRepository {
	return &ccRuleRepository{db: db}
}

// CreateCCRule 创建CC规则
func (r *ccRuleRepository) CreateCCRule(ctx context.Context, rule *model.CCRule) error {
	query := `
		INSERT INTO cc_rules (uri, limit_rate, time_window, limit_unit, status)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		rule.URI, rule.LimitRate, rule.TimeWindow, rule.LimitUnit, rule.Status,
	)
	if err != nil {
		return fmt.Errorf("创建CC规则失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取CC规则ID失败: %v", err)
	}
	rule.ID = id

	return nil
}

// UpdateCCRule 更新CC规则
func (r *ccRuleRepository) UpdateCCRule(ctx context.Context, rule *model.CCRule) error {
	query := `
		UPDATE cc_rules SET
			limit_rate = ?, time_window = ?, limit_unit = ?, status = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		rule.LimitRate, rule.TimeWindow, rule.LimitUnit, rule.Status, rule.ID,
	)
	if err != nil {
		return fmt.Errorf("更新CC规则失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("CC规则不存在: %d", rule.ID)
	}

	return nil
}

// DeleteCCRule 删除CC规则
func (r *ccRuleRepository) DeleteCCRule(ctx context.Context, id int64) error {
	query := "DELETE FROM cc_rules WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除CC规则失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("CC规则不存在: %d", id)
	}

	return nil
}

// GetCCRule 获取CC规则
func (r *ccRuleRepository) GetCCRule(ctx context.Context, id int64) (*model.CCRule, error) {
	query := `
		SELECT id, uri, limit_rate, time_window, limit_unit, status,
			created_at, updated_at
		FROM cc_rules WHERE id = ?
	`
	var rule model.CCRule
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID, &rule.URI, &rule.LimitRate, &rule.TimeWindow,
		&rule.LimitUnit, &rule.Status, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取CC规则失败: %v", err)
	}

	return &rule, nil
}

// ListCCRules 获取CC规则列表
func (r *ccRuleRepository) ListCCRules(ctx context.Context, offset, limit int) ([]*model.CCRule, error) {
	query := `
		SELECT id, uri, limit_rate, time_window, limit_unit, status,
			created_at, updated_at
		FROM cc_rules
		ORDER BY id DESC LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询CC规则列表失败: %v", err)
	}
	defer rows.Close()

	var rules []*model.CCRule
	for rows.Next() {
		var rule model.CCRule
		err := rows.Scan(
			&rule.ID, &rule.URI, &rule.LimitRate, &rule.TimeWindow,
			&rule.LimitUnit, &rule.Status, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描CC规则数据失败: %v", err)
		}
		rules = append(rules, &rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历CC规则数据失败: %v", err)
	}

	return rules, nil
}
