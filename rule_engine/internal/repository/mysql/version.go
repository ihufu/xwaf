package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
)

// ruleVersionRepository 规则版本MySQL仓储实现
type ruleVersionRepository struct {
	db *sql.DB
}

// NewRuleVersionRepository 创建规则版本仓储
func NewRuleVersionRepository(db *sql.DB) repository.RuleVersionRepository {
	return &ruleVersionRepository{db: db}
}

// CreateVersion 创建规则版本
func (r *ruleVersionRepository) CreateVersion(ctx context.Context, version *model.RuleVersion) error {
	query := `
		INSERT INTO rule_versions (rule_id, version, hash, content, change_type, status, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		version.RuleID, version.Version, version.Hash, version.Content,
		version.ChangeType, version.Status, version.CreatedBy,
	)
	if err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("创建规则版本失败: %v", err))
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("获取规则版本ID失败: %v", err))
	}
	version.ID = id

	return nil
}

// GetVersion 获取规则版本
func (r *ruleVersionRepository) GetVersion(ctx context.Context, ruleID, version int64) (*model.RuleVersion, error) {
	query := `
		SELECT id, rule_id, version, hash, content, change_type, status, created_by, created_at
		FROM rule_versions WHERE rule_id = ? AND version = ?
	`
	var v model.RuleVersion
	err := r.db.QueryRowContext(ctx, query, ruleID, version).Scan(
		&v.ID, &v.RuleID, &v.Version, &v.Hash, &v.Content,
		&v.ChangeType, &v.Status, &v.CreatedBy, &v.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("规则版本不存在: rule_id=%d, version=%d", ruleID, version))
	}
	if err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取规则版本失败: %v", err))
	}
	return &v, nil
}

// ListVersions 获取规则版本列表
func (r *ruleVersionRepository) ListVersions(ctx context.Context, ruleID int64) ([]*model.RuleVersion, error) {
	query := `
		SELECT id, rule_id, version, hash, content, change_type, status, created_by, created_at
		FROM rule_versions WHERE rule_id = ? ORDER BY version DESC
	`
	rows, err := r.db.QueryContext(ctx, query, ruleID)
	if err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("查询规则版本列表失败: %v", err))
	}
	defer rows.Close()

	var versions []*model.RuleVersion
	for rows.Next() {
		var v model.RuleVersion
		err := rows.Scan(
			&v.ID, &v.RuleID, &v.Version, &v.Hash, &v.Content,
			&v.ChangeType, &v.Status, &v.CreatedBy, &v.CreatedAt,
		)
		if err != nil {
			return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("扫描规则版本数据失败: %v", err))
		}
		versions = append(versions, &v)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("遍历规则版本数据失败: %v", err))
	}

	return versions, nil
}

// GetLatestVersion 获取最新版本号
func (r *ruleVersionRepository) GetLatestVersion(ctx context.Context) (int64, error) {
	query := "SELECT COALESCE(MAX(version), 0) FROM rule_versions"
	var version int64
	err := r.db.QueryRowContext(ctx, query).Scan(&version)
	if err != nil {
		return 0, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取最新版本号失败: %v", err))
	}
	return version, nil
}

// CreateSyncLog 创建同步日志
func (r *ruleVersionRepository) CreateSyncLog(ctx context.Context, log *model.RuleSyncLog) error {
	query := `
		INSERT INTO rule_sync_logs (rule_id, version, status, message, created_by)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		log.RuleID, log.Version, log.Status, log.Message, log.CreatedBy,
	)
	if err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("创建同步日志失败: %v", err))
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("获取同步日志ID失败: %v", err))
	}
	log.ID = id

	return nil
}

// ListSyncLogs 获取同步日志列表
func (r *ruleVersionRepository) ListSyncLogs(ctx context.Context, ruleID int64) ([]*model.RuleSyncLog, error) {
	query := `
		SELECT id, rule_id, version, status, message, created_by, created_at
		FROM rule_sync_logs WHERE rule_id = ? ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, ruleID)
	if err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("查询同步日志列表失败: %v", err))
	}
	defer rows.Close()

	var logs []*model.RuleSyncLog
	for rows.Next() {
		var log model.RuleSyncLog
		err := rows.Scan(
			&log.ID, &log.RuleID, &log.Version, &log.Status,
			&log.Message, &log.CreatedBy, &log.CreatedAt,
		)
		if err != nil {
			return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("扫描同步日志数据失败: %v", err))
		}
		logs = append(logs, &log)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("遍历同步日志数据失败: %v", err))
	}

	return logs, nil
}

// GetRulesByVersion 获取指定版本的规则列表
func (r *ruleVersionRepository) GetRulesByVersion(ctx context.Context, version int64) ([]*model.Rule, error) {
	query := `
		SELECT r.* FROM rules r
		INNER JOIN rule_versions rv ON r.id = rv.rule_id
		WHERE rv.version = ?
	`
	rows, err := r.db.QueryContext(ctx, query, version)
	if err != nil {
		return nil, fmt.Errorf("查询规则列表失败: %v", err)
	}
	defer rows.Close()

	var rules []*model.Rule
	for rows.Next() {
		var rule model.Rule
		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Type,
			&rule.Action, &rule.Priority, &rule.Status, &rule.CreatedBy,
			&rule.UpdatedBy, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描规则数据失败: %v", err)
		}
		rules = append(rules, &rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历规则数据失败: %v", err)
	}

	return rules, nil
}

// RefreshRules 刷新规则
func (r *ruleVersionRepository) RefreshRules(ctx context.Context) error {
	query := `UPDATE rules SET updated_at = CURRENT_TIMESTAMP`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("刷新规则失败: %v", err)
	}
	return nil
}

// RollbackRules 回滚规则
func (r *ruleVersionRepository) RollbackRules(ctx context.Context, rules []*model.Rule, event *model.RuleUpdateEvent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 更新规则
	for _, rule := range rules {
		query := `
			UPDATE rules 
			SET name = ?, description = ?, type = ?, action = ?,
				priority = ?, status = ?, updated_by = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`
		_, err := tx.ExecContext(ctx, query,
			rule.Name, rule.Description, rule.Type, rule.Action,
			rule.Priority, rule.Status, rule.UpdatedBy, rule.ID,
		)
		if err != nil {
			return fmt.Errorf("更新规则失败: %v", err)
		}
	}

	// 记录更新事件
	query := `
		INSERT INTO rule_update_events (version, action, rule_diffs)
		VALUES (?, ?, ?)
	`
	_, err = tx.ExecContext(ctx, query,
		event.Version, event.Action, event.RuleDiffs,
	)
	if err != nil {
		return fmt.Errorf("记录更新事件失败: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}
