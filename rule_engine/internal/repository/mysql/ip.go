package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
)

// ipRuleRepository IP规则MySQL仓储实现
type ipRuleRepository struct {
	db *sql.DB
}

// NewIPRuleRepository 创建IP规则仓储
func NewIPRuleRepository(db *sql.DB) repository.IPRuleRepository {
	return &ipRuleRepository{db: db}
}

// CreateIPRule 创建IP规则
func (r *ipRuleRepository) CreateIPRule(ctx context.Context, rule *model.IPRule) error {
	query := `
		INSERT INTO ip_rules (
			ip, ip_type, block_type, expire_time, description,
			created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		rule.IP, rule.IPType, rule.BlockType, rule.ExpireTime, rule.Description,
		rule.CreatedBy, rule.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("创建IP规则失败: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取IP规则ID失败: %v", err)
	}
	rule.ID = id

	return nil
}

// UpdateIPRule 更新IP规则
func (r *ipRuleRepository) UpdateIPRule(ctx context.Context, rule *model.IPRule) error {
	query := `
		UPDATE ip_rules SET
			ip_type = ?, block_type = ?, expire_time = ?, description = ?,
			updated_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		rule.IPType, rule.BlockType, rule.ExpireTime, rule.Description,
		rule.UpdatedBy, rule.ID,
	)
	if err != nil {
		return fmt.Errorf("更新IP规则失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("IP规则不存在: %d", rule.ID)
	}

	return nil
}

// DeleteIPRule 删除IP规则
func (r *ipRuleRepository) DeleteIPRule(ctx context.Context, id int64) error {
	query := "DELETE FROM ip_rules WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除IP规则失败: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("IP规则不存在: %d", id)
	}

	return nil
}

// GetIPRule 获取IP规则
func (r *ipRuleRepository) GetIPRule(ctx context.Context, id int64) (*model.IPRule, error) {
	query := `
		SELECT id, ip, ip_type, block_type, expire_time, description,
			created_by, updated_by, created_at, updated_at
		FROM ip_rules WHERE id = ?
	`
	var rule model.IPRule
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID, &rule.IP, &rule.IPType, &rule.BlockType, &rule.ExpireTime,
		&rule.Description, &rule.CreatedBy, &rule.UpdatedBy,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取IP规则失败: %v", err)
	}

	return &rule, nil
}

// GetIPRuleByIP 根据IP获取规则
func (r *ipRuleRepository) GetIPRuleByIP(ctx context.Context, ip string) (*model.IPRule, error) {
	query := `
		SELECT id, ip, ip_type, block_type, expire_time, description,
			created_by, updated_by, created_at, updated_at
		FROM ip_rules WHERE ip = ?
	`
	var rule model.IPRule
	err := r.db.QueryRowContext(ctx, query, ip).Scan(
		&rule.ID, &rule.IP, &rule.IPType, &rule.BlockType, &rule.ExpireTime,
		&rule.Description, &rule.CreatedBy, &rule.UpdatedBy,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取IP规则失败: %v", err)
	}

	return &rule, nil
}

// ListIPRules 获取IP规则列表
func (r *ipRuleRepository) ListIPRules(ctx context.Context, query *model.IPRuleQuery, offset, limit int) ([]*model.IPRule, int64, error) {
	// 构建查询条件
	conditions := []string{"1 = 1"}
	args := []interface{}{}

	if query.Keyword != "" {
		conditions = append(conditions, "(ip LIKE ? OR description LIKE ?)")
		keyword := "%" + query.Keyword + "%"
		args = append(args, keyword, keyword)
	}
	if query.IPType != "" {
		conditions = append(conditions, "ip_type = ?")
		args = append(args, query.IPType)
	}
	if query.BlockType != "" {
		conditions = append(conditions, "block_type = ?")
		args = append(args, query.BlockType)
	}

	// 查询总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM ip_rules WHERE %s
	`, joinConditions(conditions))
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("获取IP规则总数失败: %v", err)
	}

	// 查询列表
	listQuery := fmt.Sprintf(`
		SELECT id, ip, ip_type, block_type, expire_time, description,
			created_by, updated_by, created_at, updated_at
		FROM ip_rules WHERE %s
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, joinConditions(conditions))
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询IP规则列表失败: %v", err)
	}
	defer rows.Close()

	var rules []*model.IPRule
	for rows.Next() {
		var rule model.IPRule
		err := rows.Scan(
			&rule.ID, &rule.IP, &rule.IPType, &rule.BlockType, &rule.ExpireTime,
			&rule.Description, &rule.CreatedBy, &rule.UpdatedBy,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描IP规则数据失败: %v", err)
		}
		rules = append(rules, &rule)
	}

	return rules, total, nil
}

// ExistsByIP 检查IP是否存在规则
func (r *ipRuleRepository) ExistsByIP(ctx context.Context, ip string) (bool, error) {
	query := "SELECT COUNT(*) FROM ip_rules WHERE ip = ?"
	var count int
	err := r.db.QueryRowContext(ctx, query, ip).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查IP规则是否存在失败: %v", err)
	}
	return count > 0, nil
}

// 辅助函数：拼接查询条件
func joinConditions(conditions []string) string {
	result := conditions[0]
	for i := 1; i < len(conditions); i++ {
		result += " AND " + conditions[i]
	}
	return result
}
