package service

import (
	"context"
	"fmt"
	"time"

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
			return fmt.Errorf("规则[%s]验证失败: %s", rule.Name, err.Error())
		}
	}

	// 开启事务
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 批量创建规则
	for _, rule := range rules {
		if err := s.repo.CreateRule(ctx, rule); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// BatchUpdateRules 批量更新规则
func (s *enhancedRuleService) BatchUpdateRules(ctx context.Context, rules []*model.Rule) error {
	// 规则验证
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("规则[%s]验证失败: %s", rule.Name, err.Error())
		}
	}

	// 开启事务
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 批量更新规则
	for _, rule := range rules {
		if err := s.repo.UpdateRule(ctx, rule); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// BatchDeleteRules 批量删除规则
func (s *enhancedRuleService) BatchDeleteRules(ctx context.Context, ids []int64) error {
	// 开启事务
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 批量删除规则
	for _, id := range ids {
		if err := s.repo.DeleteRule(ctx, id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ImportRules 导入规则
func (s *enhancedRuleService) ImportRules(ctx context.Context, rules []*model.Rule) error {
	// 规则验证
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("规则[%s]验证失败: %s", rule.Name, err.Error())
		}
	}

	// 开启事务
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 导入规则
	for _, rule := range rules {
		// 检查规则是否已存在
		existingRule, err := s.repo.GetRuleByName(ctx, rule.Name)
		if err != nil && !repository.IsNotFound(err) {
			return err
		}

		if existingRule != nil {
			// 更新已存在的规则
			rule.ID = existingRule.ID
			if err := s.repo.UpdateRule(ctx, rule); err != nil {
				return err
			}
		} else {
			// 创建新规则
			if err := s.repo.CreateRule(ctx, rule); err != nil {
				return err
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
		return nil, err
	}
	return rules, nil
}

// GetRuleStats 获取规则统计信息
func (s *enhancedRuleService) GetRuleStats(ctx context.Context) (*model.RuleStats, error) {
	return s.statsRepo.GetRuleStats(ctx)
}

// GetRuleMatchStats 获取规则匹配统计
func (s *enhancedRuleService) GetRuleMatchStats(ctx context.Context, ruleID int64, startTime, endTime time.Time) (*model.RuleMatchStat, error) {
	return s.statsRepo.GetRuleMatchStats(ctx, ruleID, startTime, endTime)
}

// CreateRuleAuditLog 创建规则审计日志
func (s *enhancedRuleService) CreateRuleAuditLog(ctx context.Context, log *model.RuleAuditLog) error {
	return s.auditRepo.CreateAuditLog(ctx, log)
}

// GetRuleAuditLogs 获取规则审计日志
func (s *enhancedRuleService) GetRuleAuditLogs(ctx context.Context, ruleID int64) ([]*model.RuleAuditLog, error) {
	return s.auditRepo.GetAuditLogs(ctx, ruleID)
}
