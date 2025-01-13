package service

import (
	"context"
	"fmt"
	"time"

	"github.com/xwaf/rule_engine/internal/errors"
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
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("规则验证失败: %v", err))
	}

	// 创建规则
	if err := s.repo.CreateRule(ctx, rule); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("创建规则失败: %v", err))
	}

	// 更新缓存
	if err := s.cache.SetRule(ctx, rule); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新规则缓存失败: %v", err))
	}

	return nil
}

// UpdateRule 更新规则
func (s *ruleService) UpdateRule(ctx context.Context, rule *model.Rule) error {
	// 验证规则
	if err := rule.Validate(); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("规则验证失败: %v", err))
	}

	// 更新规则
	if err := s.repo.UpdateRule(ctx, rule); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新规则失败: %v", err))
	}

	// 更新缓存
	if err := s.cache.SetRule(ctx, rule); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新规则缓存失败: %v", err))
	}

	return nil
}

// DeleteRule 删除规则
func (s *ruleService) DeleteRule(ctx context.Context, id int64) error {
	// 删除规则
	if err := s.repo.DeleteRule(ctx, id); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("删除规则失败: %v", err))
	}

	// 删除缓存
	if err := s.cache.DeleteRule(ctx, id); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("删除规则缓存失败: %v", err))
	}

	return nil
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
		return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则失败: %v", err))
	}

	// 更新缓存
	if rule != nil {
		if err := s.cache.SetRule(ctx, rule); err != nil {
			return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新规则缓存失败: %v", err))
		}
	}

	return rule, nil
}

// ListRules 获取规则列表
func (s *ruleService) ListRules(ctx context.Context, query *repository.RuleQuery) ([]*model.Rule, int64, error) {
	rules, total, err := s.repo.ListRules(ctx, query)
	if err != nil {
		return nil, 0, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则列表失败: %v", err))
	}
	return rules, total, nil
}

// CheckRequest 检查规则匹配
func (s *ruleService) CheckRequest(ctx context.Context, req *model.CheckRequest) (*model.CheckResult, error) {
	// 获取所有规则
	rules, _, err := s.repo.ListRules(ctx, &repository.RuleQuery{
		Status: model.StatusEnabled,
	})
	if err != nil {
		return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则列表失败: %v", err))
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
			return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("创建规则处理器失败: %v", err))
		}

		// 执行规则匹配
		matched, err := handler.Match(ctx, rule, req)
		if err != nil {
			return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("规则匹配失败: %v", err))
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
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("清空规则缓存失败: %v", err))
	}

	// 重新加载规则到缓存
	rules, _, err := s.repo.ListRules(ctx, &repository.RuleQuery{})
	if err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则列表失败: %v", err))
	}

	for _, rule := range rules {
		if err := s.cache.SetRule(ctx, rule); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新规则缓存失败: %v", err))
		}
	}

	return nil
}

// GetVersion 获取规则版本
func (s *ruleService) GetVersion(ctx context.Context) (int64, error) {
	version, err := s.repo.GetLatestVersion(ctx)
	if err != nil {
		return 0, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则版本失败: %v", err))
	}
	return version, nil
}

// BatchCreateRules 批量创建规则
func (s *ruleService) BatchCreateRules(ctx context.Context, rules []*model.Rule) error {
	// 验证所有规则
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("规则验证失败: %v", err))
		}
	}

	// 批量创建规则
	if err := s.repo.BatchCreateRules(ctx, rules); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("批量创建规则失败: %v", err))
	}

	// 更新缓存
	for _, rule := range rules {
		if err := s.cache.SetRule(ctx, rule); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新规则缓存失败: %v", err))
		}
	}

	return nil
}

// BatchUpdateRules 批量更新规则
func (s *ruleService) BatchUpdateRules(ctx context.Context, rules []*model.Rule) error {
	// 验证所有规则
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("规则验证失败 (ID: %d): %v", rule.ID, err))
		}
	}

	// 批量更新规则
	for _, rule := range rules {
		if err := s.repo.UpdateRule(ctx, rule); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新规则失败 (ID: %d): %v", rule.ID, err))
		}
		// 更新缓存
		if err := s.cache.SetRule(ctx, rule); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新规则缓存失败 (ID: %d): %v", rule.ID, err))
		}
	}

	return nil
}

// BatchDeleteRules 批量删除规则
func (s *ruleService) BatchDeleteRules(ctx context.Context, ids []int64) error {
	// 删除规则
	if err := s.repo.BatchDeleteRules(ctx, ids); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("删除规则失败: %v", err))
	}

	// 从缓存中删除规则
	for _, id := range ids {
		if err := s.cache.DeleteRule(ctx, id); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("从缓存中删除规则失败: %v", err))
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
		return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("导出规则失败: %v", err))
	}
	return rules, nil
}

// GetRuleAuditLogs 获取规则审计日志
func (s *ruleService) GetRuleAuditLogs(ctx context.Context, ruleID int64) ([]*model.RuleAuditLog, error) {
	logs, err := s.repo.GetRuleAuditLogs(ctx, ruleID)
	if err != nil {
		return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则审计日志失败: %v", err))
	}
	return logs, nil
}

// GetRuleMatchStats 获取规则匹配统计
func (s *ruleService) GetRuleMatchStats(ctx context.Context, ruleID int64, startTime, endTime time.Time) (*model.RuleMatchStat, error) {
	stats, err := s.repo.GetRuleMatchStats(ctx, ruleID, startTime, endTime)
	if err != nil {
		return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则匹配统计失败: %v", err))
	}
	return stats, nil
}

// GetRuleStats 获取规则统计信息
func (s *ruleService) GetRuleStats(ctx context.Context) (*model.RuleStats, error) {
	stats, err := s.repo.GetRuleStats(ctx)
	if err != nil {
		return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则统计信息失败: %v", err))
	}
	return stats, nil
}

// ImportRules 导入规则
func (s *ruleService) ImportRules(ctx context.Context, rules []*model.Rule) error {
	// 验证所有规则
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("规则验证失败: %v", err))
		}
	}

	if err := s.repo.ImportRules(ctx, rules); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("导入规则失败: %v", err))
	}
	return nil
}

// CreateRuleAuditLog 创建规则审计日志
func (s *ruleService) CreateRuleAuditLog(ctx context.Context, log *model.RuleAuditLog) error {
	if err := s.repo.CreateRuleAuditLog(ctx, log); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("创建规则审计日志失败: %v", err))
	}
	return nil
}

// IncrRuleMatchCount 增加规则匹配计数
func (s *ruleService) IncrRuleMatchCount(ctx context.Context, ruleID int64) error {
	if err := s.repo.IncrRuleMatchCount(ctx, ruleID); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("增加规则匹配计数失败: %v", err))
	}
	return nil
}

// GetRuleMatchCount 获取规则匹配计数
func (s *ruleService) GetRuleMatchCount(ctx context.Context, ruleID int64) (int64, error) {
	count, err := s.repo.GetRuleMatchCount(ctx, ruleID)
	if err != nil {
		return 0, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则匹配计数失败: %v", err))
	}
	return count, nil
}
