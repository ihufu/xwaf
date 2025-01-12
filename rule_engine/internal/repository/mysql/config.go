package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
)

// wafConfigRepository WAF配置MySQL仓储实现
type wafConfigRepository struct {
	db *sql.DB
}

// NewWAFConfigRepository 创建WAF配置仓储
func NewWAFConfigRepository(db *sql.DB) repository.WAFConfigRepository {
	return &wafConfigRepository{db: db}
}

// GetConfig 获取WAF配置
func (r *wafConfigRepository) GetConfig(ctx context.Context) (*model.WAFConfig, error) {
	query := `
		SELECT id, mode, description, created_by, updated_by, created_at, updated_at
		FROM waf_configs ORDER BY id DESC LIMIT 1
	`
	var config model.WAFConfig
	err := r.db.QueryRowContext(ctx, query).Scan(
		&config.ID, &config.Mode, &config.Description,
		&config.CreatedBy, &config.UpdatedBy,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取WAF配置失败: %v", err)
	}

	return &config, nil
}

// UpdateConfig 更新WAF配置
func (r *wafConfigRepository) UpdateConfig(ctx context.Context, config *model.WAFConfig) error {
	if config.ID == 0 {
		// 创建新配置
		query := `
			INSERT INTO waf_configs (mode, description, created_by, updated_by)
			VALUES (?, ?, ?, ?)
		`
		result, err := r.db.ExecContext(ctx, query,
			config.Mode, config.Description,
			config.CreatedBy, config.UpdatedBy,
		)
		if err != nil {
			return fmt.Errorf("创建WAF配置失败: %v", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("获取WAF配置ID失败: %v", err)
		}
		config.ID = id
	} else {
		// 更新现有配置
		query := `
			UPDATE waf_configs SET
				mode = ?, description = ?, updated_by = ?,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`
		result, err := r.db.ExecContext(ctx, query,
			config.Mode, config.Description,
			config.UpdatedBy, config.ID,
		)
		if err != nil {
			return fmt.Errorf("更新WAF配置失败: %v", err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("获取影响行数失败: %v", err)
		}
		if affected == 0 {
			return fmt.Errorf("WAF配置不存在: %d", config.ID)
		}
	}

	return nil
}
