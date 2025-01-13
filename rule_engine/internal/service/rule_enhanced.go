package service

import (
	"context"
	"fmt"
	"time"

	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
)

// EnhancedRuleService 增强的规则服务接口
type EnhancedRuleService interface {
	RuleService // 继承基础规则服务接口

	// 批量操作
	BatchCreateRules(ctx context.Context, rules []*model.Rule) error
	BatchUpdateRules(ctx context.Context, rules []*model.Rule) error
	BatchDeleteRules(ctx context.Context, ids []int64) error

	// 导入导出
	ImportRules(ctx context.Context, rules []*model.Rule) error
	ExportRules(ctx context.Context, query *repository.RuleQuery) ([]*model.Rule, error)

	// 统计功能
	GetRuleStats(ctx context.Context) (*model.RuleStats, error)
	GetRuleMatchStats(ctx context.Context, ruleID int64, startTime, endTime time.Time) (*model.RuleMatchStat, error)

	// 审计日志
	CreateRuleAuditLog(ctx context.Context, log *model.RuleAuditLog) error
	GetRuleAuditLogs(ctx context.Context, ruleID int64) ([]*model.RuleAuditLog, error)
}

// enhancedRuleService 增强的规则服务实现
type enhancedRuleService struct {
	*ruleService // 继承基础规则服务
	auditRepo    repository.RuleAuditRepository
	statsRepo    repository.RuleStatsRepository
}

// NewEnhancedRuleService 创建增强的规则服务
func NewEnhancedRuleService(
	baseService *ruleService,
	auditRepo repository.RuleAuditRepository,
	statsRepo repository.RuleStatsRepository,
) EnhancedRuleService {
	return &enhancedRuleService{
		ruleService: baseService,
		auditRepo:   auditRepo,
		statsRepo:   statsRepo,
	}
}

// BatchCreateRules 批量创建规则
func (s *enhancedRuleService) BatchCreateRules(ctx context.Context, rules []*model.Rule) error {
	// 规则验证
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("规则[%s]验证失败: %v", rule.Name, err))
		}
	}

	// 开启事务
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("开启事务失败: %v", err))
	}
	defer tx.Rollback()

	// 批量创建规则
	for _, rule := range rules {
		if err := s.repo.CreateRule(ctx, rule); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("创建规则[%s]失败: %v", rule.Name, err))
		}
	}

	return tx.Commit()
}

// BatchUpdateRules 批量更新规则
func (s *enhancedRuleService) BatchUpdateRules(ctx context.Context, rules []*model.Rule) error {
	// 规则验证
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("规则[%s]验证失败: %v", rule.Name, err))
		}
	}

	// 开启事务
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("开启事务失败: %v", err))
	}
	defer tx.Rollback()

	// 批量更新规则
	for _, rule := range rules {
		if err := s.repo.UpdateRule(ctx, rule); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新规则[%s]失败: %v", rule.Name, err))
		}
	}

	return tx.Commit()
}

// BatchDeleteRules 批量删除规则
func (s *enhancedRuleService) BatchDeleteRules(ctx context.Context, ids []int64) error {
	// 开启事务
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("开启事务失败: %v", err))
	}
	defer tx.Rollback()

	// 批量删除规则
	for _, id := range ids {
		if err := s.repo.DeleteRule(ctx, id); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("删除规则[ID:%d]失败: %v", id, err))
		}
	}

	return tx.Commit()
}

// ImportRules 导入规则
func (s *enhancedRuleService) ImportRules(ctx context.Context, rules []*model.Rule) error {
	// 规则验证
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("规则[%s]验证失败: %v", rule.Name, err))
		}
	}

	// 开启事务
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("开启事务失败: %v", err))
	}
	defer tx.Rollback()

	// 导入规则
	for _, rule := range rules {
		// 检查规则是否已存在
		existingRule, err := s.repo.GetRuleByName(ctx, rule.Name)
		if err != nil && !repository.IsNotFound(err) {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("检查规则[%s]是否存在失败: %v", rule.Name, err))
		}

		if existingRule != nil {
			// 更新已存在的规则
			rule.ID = existingRule.ID
			if err := s.repo.UpdateRule(ctx, rule); err != nil {
				return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新规则[%s]失败: %v", rule.Name, err))
			}
		} else {
			// 创建新规则
			if err := s.repo.CreateRule(ctx, rule); err != nil {
				return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("创建规则[%s]失败: %v", rule.Name, err))
			}
		}
	}

	return tx.Commit()
}

// ExportRules 导出规则
func (s *enhancedRuleService) ExportRules(ctx context.Context, query *repository.RuleQuery) ([]*model.Rule, error) {
	// 不分页，导出所有符合条件的规则
	query.Page = 0
	query.PageSize = 0
	rules, _, err := s.repo.ListRules(ctx, query)
	if err != nil {
		return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("导出规则失败: %v", err))
	}
	return rules, nil
}

// GetRuleStats 获取规则统计信息
func (s *enhancedRuleService) GetRuleStats(ctx context.Context) (*model.RuleStats, error) {
	stats, err := s.statsRepo.GetRuleStats(ctx)
	if err != nil {
		return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则统计信息失败: %v", err))
	}
	return stats, nil
}

// GetRuleMatchStats 获取规则匹配统计
func (s *enhancedRuleService) GetRuleMatchStats(ctx context.Context, ruleID int64, startTime, endTime time.Time) (*model.RuleMatchStat, error) {
	stats, err := s.statsRepo.GetRuleMatchStats(ctx, ruleID, startTime, endTime)
	if err != nil {
		return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则[ID:%d]匹配统计失败: %v", ruleID, err))
	}
	return stats, nil
}

// CreateRuleAuditLog 创建规则审计日志
func (s *enhancedRuleService) CreateRuleAuditLog(ctx context.Context, log *model.RuleAuditLog) error {
	if err := s.auditRepo.CreateAuditLog(ctx, log); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("创建规则[ID:%d]审计日志失败: %v", log.RuleID, err))
	}
	return nil
}

// GetRuleAuditLogs 获取规则审计日志
func (s *enhancedRuleService) GetRuleAuditLogs(ctx context.Context, ruleID int64) ([]*model.RuleAuditLog, error) {
	logs, err := s.auditRepo.GetAuditLogs(ctx, ruleID)
	if err != nil {
		return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则[ID:%d]审计日志失败: %v", ruleID, err))
	}
	return logs, nil
}
