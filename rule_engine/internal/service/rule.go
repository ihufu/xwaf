package service

import (
	"context"
	"fmt"
	"time"

	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
)

// ruleService 规则服务实现
type ruleService struct {
	repo    repository.RuleRepository
	factory RuleFactory
	cache   repository.RuleCache
}

// NewRuleService 创建规则服务
func NewRuleService(repo repository.RuleRepository, factory RuleFactory, cache repository.RuleCache) RuleService {
	return &ruleService{
		repo:    repo,
		factory: factory,
		cache:   cache,
	}
}

// CreateRule 创建规则
func (s *ruleService) CreateRule(ctx context.Context, rule *model.Rule) error {
	// 验证规则
	if err := rule.Validate(); err != nil {
		return err
	}

	// 创建规则
	if err := s.repo.CreateRule(ctx, rule); err != nil {
		return err
	}

	// 更新缓存
	return s.cache.SetRule(ctx, rule)
}

// UpdateRule 更新规则
func (s *ruleService) UpdateRule(ctx context.Context, rule *model.Rule) error {
	// 验证规则
	if err := rule.Validate(); err != nil {
		return err
	}

	// 更新规则
	if err := s.repo.UpdateRule(ctx, rule); err != nil {
		return err
	}

	// 更新缓存
	return s.cache.SetRule(ctx, rule)
}

// DeleteRule 删除规则
func (s *ruleService) DeleteRule(ctx context.Context, id int64) error {
	// 删除规则
	if err := s.repo.DeleteRule(ctx, id); err != nil {
		return err
	}

	// 删除缓存
	return s.cache.DeleteRule(ctx, id)
}

// GetRule 获取规则
func (s *ruleService) GetRule(ctx context.Context, id int64) (*model.Rule, error) {
	// 从缓存获取
	rule, err := s.cache.GetRule(ctx, id)
	if err == nil && rule != nil {
		return rule, nil
	}

	// 从数据库获取
	rule, err = s.repo.GetRule(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	if rule != nil {
		if err := s.cache.SetRule(ctx, rule); err != nil {
			return nil, err
		}
	}

	return rule, nil
}

// ListRules 获取规则列表
func (s *ruleService) ListRules(ctx context.Context, query *repository.RuleQuery) ([]*model.Rule, int64, error) {
	return s.repo.ListRules(ctx, query)
}

// CheckRequest 检查规则匹配
func (s *ruleService) CheckRequest(ctx context.Context, req *model.CheckRequest) (*model.CheckResult, error) {
	// 获取所有规则
	rules, _, err := s.repo.ListRules(ctx, &repository.RuleQuery{
		Status: model.StatusEnabled,
	})
	if err != nil {
		return nil, err
	}

	// 按优先级排序规则
	model.SortRulesByPriority(rules)

	// 检查每个规则
	for _, rule := range rules {
		// 检查规则类型是否需要处理
		if len(req.RuleTypes) > 0 {
			matched := false
			for _, t := range req.RuleTypes {
				if rule.Type == t {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// 获取规则处理器
		handler, err := s.factory.CreateRuleHandler(rule.Type)
		if err != nil {
			return nil, fmt.Errorf("创建规则处理器失败: %v", err)
		}

		// 执行规则匹配
		matched, err := handler.Match(ctx, rule, req)
		if err != nil {
			return nil, fmt.Errorf("规则匹配失败: %v", err)
		}

		// 如果匹配成功，返回结果
		if matched {
			return &model.CheckResult{
				Matched:     true,
				Action:      rule.Action,
				MatchedRule: rule,
				Message:     fmt.Sprintf("命中规则: %s", rule.Name),
			}, nil
		}
	}

	// 未匹配任何规则
	return &model.CheckResult{
		Matched: false,
		Action:  model.ActionAllow,
		Message: "未命中任何规则",
	}, nil
}

// ReloadRules 重新加载规则
func (s *ruleService) ReloadRules(ctx context.Context) error {
	// 清空缓存
	if err := s.cache.ClearRules(ctx); err != nil {
		return err
	}

	// 重新加载规则到缓存
	rules, _, err := s.repo.ListRules(ctx, &repository.RuleQuery{})
	if err != nil {
		return err
	}

	for _, rule := range rules {
		if err := s.cache.SetRule(ctx, rule); err != nil {
			return err
		}
	}

	return nil
}

// GetVersion 获取规则版本
func (s *ruleService) GetVersion(ctx context.Context) (int64, error) {
	return s.repo.GetLatestVersion(ctx)
}

// BatchCreateRules 批量创建规则
func (s *ruleService) BatchCreateRules(ctx context.Context, rules []*model.Rule) error {
	// 验证所有规则
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("规则验证失败: %w", err)
		}
	}

	// 批量创建规则
	if err := s.repo.BatchCreateRules(ctx, rules); err != nil {
		return fmt.Errorf("批量创建规则失败: %w", err)
	}

	// 更新缓存
	for _, rule := range rules {
		if err := s.cache.SetRule(ctx, rule); err != nil {
			return fmt.Errorf("更新规则缓存失败: %w", err)
		}
	}

	return nil
}

// BatchUpdateRules 批量更新规则
func (s *ruleService) BatchUpdateRules(ctx context.Context, rules []*model.Rule) error {
	// 验证所有规则
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("规则验证失败 (ID: %d): %w", rule.ID, err)
		}
	}

	// 批量更新规则
	for _, rule := range rules {
		if err := s.repo.UpdateRule(ctx, rule); err != nil {
			return fmt.Errorf("更新规则失败 (ID: %d): %w", rule.ID, err)
		}
		// 更新缓存
		if err := s.cache.SetRule(ctx, rule); err != nil {
			return fmt.Errorf("更新规则缓存失败 (ID: %d): %w", rule.ID, err)
		}
	}

	return nil
}

// BatchDeleteRules 批量删除规则
func (s *ruleService) BatchDeleteRules(ctx context.Context, ids []int64) error {
	// 删除规则
	if err := s.repo.BatchDeleteRules(ctx, ids); err != nil {
		return fmt.Errorf("删除规则失败: %w", err)
	}

	// 从缓存中删除规则
	for _, id := range ids {
		if err := s.cache.DeleteRule(ctx, id); err != nil {
			return fmt.Errorf("从缓存中删除规则失败: %w", err)
		}
	}

	return nil
}

// ExportRules 导出规则
func (s *ruleService) ExportRules(ctx context.Context, query *repository.RuleQuery) ([]*model.Rule, error) {
	// 不分页，导出所有符合条件的规则
	query.Page = 0
	query.PageSize = 0
	rules, _, err := s.repo.ListRules(ctx, query)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// GetRuleAuditLogs 获取规则审计日志
func (s *ruleService) GetRuleAuditLogs(ctx context.Context, ruleID int64) ([]*model.RuleAuditLog, error) {
	return s.repo.GetRuleAuditLogs(ctx, ruleID)
}

// GetRuleMatchStats 获取规则匹配统计
func (s *ruleService) GetRuleMatchStats(ctx context.Context, ruleID int64, startTime, endTime time.Time) (*model.RuleMatchStat, error) {
	stats, err := s.repo.GetRuleMatchStats(ctx, ruleID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// GetRuleStats 获取规则统计信息
func (s *ruleService) GetRuleStats(ctx context.Context) (*model.RuleStats, error) {
	return s.repo.GetRuleStats(ctx)
}

// ImportRules 导入规则
func (s *ruleService) ImportRules(ctx context.Context, rules []*model.Rule) error {
	return s.repo.ImportRules(ctx, rules)
}

// CreateRuleAuditLog 创建规则审计日志
func (s *ruleService) CreateRuleAuditLog(ctx context.Context, log *model.RuleAuditLog) error {
	return s.repo.CreateRuleAuditLog(ctx, log)
}

// IncrRuleMatchCount 增加规则匹配计数
func (s *ruleService) IncrRuleMatchCount(ctx context.Context, ruleID int64) error {
	return s.repo.IncrRuleMatchCount(ctx, ruleID)
}

// GetRuleMatchCount 获取规则匹配计数
func (s *ruleService) GetRuleMatchCount(ctx context.Context, ruleID int64) (int64, error) {
	return s.repo.GetRuleMatchCount(ctx, ruleID)
}
