package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
)

// wafConfigRepository WAF配置MySQL仓储实现
type wafConfigRepository struct {
	db *sql.DB
}

// NewWAFConfigRepository 创建WAF配置仓储
func NewWAFConfigRepository(db *sql.DB) repository.WAFConfigRepository {
	if db == nil {
		panic("数据库连接不能为空")
	}
	return &wafConfigRepository{db: db}
}

// GetConfig 获取WAF配置
func (r *wafConfigRepository) GetConfig(ctx context.Context) (*model.WAFConfig, error) {
	if ctx == nil {
		return nil, errors.NewError(errors.ErrValidation, "上下文不能为空")
	}

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
		return nil, errors.NewError(errors.ErrConfig, "WAF配置不存在，请先创建配置")
	}
	if err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取WAF配置失败: %v", err))
	}

	return &config, nil
}

// UpdateConfig 更新WAF配置
func (r *wafConfigRepository) UpdateConfig(ctx context.Context, config *model.WAFConfig) error {
	if ctx == nil {
		return errors.NewError(errors.ErrValidation, "上下文不能为空")
	}
	if config == nil {
		return errors.NewError(errors.ErrValidation, "配置不能为空")
	}
	if config.Mode == "" {
		return errors.NewError(errors.ErrValidation, "WAF模式不能为空")
	}
	if config.CreatedBy == "" {
		return errors.NewError(errors.ErrValidation, "创建者不能为空")
	}
	if config.UpdatedBy == "" {
		return errors.NewError(errors.ErrValidation, "更新者不能为空")
	}

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
			return errors.NewError(errors.ErrSystem, fmt.Sprintf("创建WAF配置失败: %v", err))
		}

		id, err := result.LastInsertId()
		if err != nil {
			return errors.NewError(errors.ErrSystem, fmt.Sprintf("获取WAF配置ID失败: %v", err))
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
			return errors.NewError(errors.ErrSystem, fmt.Sprintf("更新WAF配置失败: %v", err))
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return errors.NewError(errors.ErrSystem, fmt.Sprintf("获取影响行数失败: %v", err))
		}
		if affected == 0 {
			return errors.NewError(errors.ErrConfig, fmt.Sprintf("WAF配置不存在: ID=%d", config.ID))
		}
	}

	return nil
}

// LogModeChange 记录模式变更日志
func (r *wafConfigRepository) LogModeChange(ctx context.Context, log *model.WAFModeChangeLog) error {
	if ctx == nil {
		return errors.NewError(errors.ErrValidation, "上下文不能为空")
	}
	if log == nil {
		return errors.NewError(errors.ErrValidation, "日志不能为空")
	}
	if log.OldMode == "" {
		return errors.NewError(errors.ErrValidation, "原WAF模式不能为空")
	}
	if log.NewMode == "" {
		return errors.NewError(errors.ErrValidation, "新WAF模式不能为空")
	}
	if log.Operator == "" {
		return errors.NewError(errors.ErrValidation, "操作者不能为空")
	}
	if log.Reason == "" {
		return errors.NewError(errors.ErrValidation, "变更原因不能为空")
	}
	if log.CreatedAt == 0 {
		return errors.NewError(errors.ErrValidation, "创建时间不能为空")
	}

	query := `
		INSERT INTO waf_mode_change_logs (old_mode, new_mode, operator, reason, description, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.OldMode, log.NewMode, log.Operator,
		log.Reason, log.Description, log.CreatedAt,
	)
	if err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("记录模式变更日志失败: %v", err))
	}
	return nil
}

// GetModeChangeLogs 获取模式变更日志
func (r *wafConfigRepository) GetModeChangeLogs(ctx context.Context, startTime, endTime int64, page, pageSize int) ([]*model.WAFModeChangeLog, int64, error) {
	if ctx == nil {
		return nil, 0, errors.NewError(errors.ErrValidation, "上下文不能为空")
	}
	if startTime < 0 {
		return nil, 0, errors.NewError(errors.ErrValidation, "开始时间不能为负数")
	}
	if endTime < startTime {
		return nil, 0, errors.NewError(errors.ErrValidation, "结束时间不能小于开始时间")
	}
	if page <= 0 {
		return nil, 0, errors.NewError(errors.ErrValidation, "页码必须大于0")
	}
	if pageSize <= 0 {
		return nil, 0, errors.NewError(errors.ErrValidation, "每页大小必须大于0")
	}

	// 获取总数
	countQuery := `
		SELECT COUNT(*) FROM waf_mode_change_logs 
		WHERE created_at BETWEEN ? AND ?
	`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, startTime, endTime).Scan(&total)
	if err != nil {
		return nil, 0, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取模式变更日志总数失败: %v", err))
	}

	// 获取日志列表
	offset := (page - 1) * pageSize
	query := `
		SELECT id, old_mode, new_mode, operator, reason, description, created_at
		FROM waf_mode_change_logs
		WHERE created_at BETWEEN ? AND ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, startTime, endTime, pageSize, offset)
	if err != nil {
		return nil, 0, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取模式变更日志失败: %v", err))
	}
	defer rows.Close()

	var logs []*model.WAFModeChangeLog
	for rows.Next() {
		var log model.WAFModeChangeLog
		err := rows.Scan(
			&log.ID, &log.OldMode, &log.NewMode,
			&log.Operator, &log.Reason, &log.Description,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, errors.NewError(errors.ErrSystem, fmt.Sprintf("扫描模式变更日志失败: %v", err))
		}
		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.NewError(errors.ErrSystem, fmt.Sprintf("遍历模式变更日志失败: %v", err))
	}

	return logs, total, nil
}
