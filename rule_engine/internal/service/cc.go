package service

import (
	"context"
	"fmt"
	"time"

	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// CCRuleService CC 防护服务接口
type CCRuleService interface {
	CreateCCRule(ctx context.Context, rule *model.CCRule) error
	UpdateCCRule(ctx context.Context, rule *model.CCRule) error
	DeleteCCRule(ctx context.Context, id int64) error
	GetCCRule(ctx context.Context, id int64) (*model.CCRule, error)
	ListCCRules(ctx context.Context, query model.CCRuleQuery, page, size int) ([]*model.CCRule, int64, error)
	CheckCCLimit(ctx context.Context, uri string) (bool, error)
	ReloadRules(ctx context.Context) error
	CheckCC(ctx context.Context, ip string, path string, method string) (bool, error)
}

// ccRuleService CC 防护服务
type ccRuleService struct {
	ccRepo    repository.CCRuleRepository
	cacheRepo repository.CacheRepository
}

// NewCCRuleService 创建 CC 防护服务
func NewCCRuleService(ccRepo repository.CCRuleRepository, cacheRepo repository.CacheRepository) CCRuleService {
	return &ccRuleService{
		ccRepo:    ccRepo,
		cacheRepo: cacheRepo,
	}
}

// CreateCCRule 创建 CC 规则
func (s *ccRuleService) CreateCCRule(ctx context.Context, rule *model.CCRule) error {
	// 验证规则
	if err := rule.Validate(); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("CC规则验证失败: %v", err))
	}

	if err := s.ccRepo.CreateCCRule(ctx, rule); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("创建CC规则失败: %v", err))
	}
	return nil
}

// UpdateCCRule 更新 CC 规则
func (s *ccRuleService) UpdateCCRule(ctx context.Context, rule *model.CCRule) error {
	// 验证规则
	if err := rule.Validate(); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("CC规则验证失败: %v", err))
	}

	if err := s.ccRepo.UpdateCCRule(ctx, rule); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新CC规则失败: %v", err))
	}
	return nil
}

// DeleteCCRule 删除 CC 规则
func (s *ccRuleService) DeleteCCRule(ctx context.Context, id int64) error {
	if err := s.ccRepo.DeleteCCRule(ctx, id); err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("删除CC规则失败: %v", err))
	}
	return nil
}

// GetCCRule 获取 CC 规则
func (s *ccRuleService) GetCCRule(ctx context.Context, id int64) (*model.CCRule, error) {
	rule, err := s.ccRepo.GetCCRule(ctx, id)
	if err != nil {
		return nil, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取CC规则失败: %v", err))
	}
	return rule, nil
}

// ListCCRules 获取 CC 规则列表
func (s *ccRuleService) ListCCRules(ctx context.Context, query model.CCRuleQuery, page, size int) ([]*model.CCRule, int64, error) {
	rules, err := s.ccRepo.ListCCRules(ctx, page*size, size)
	if err != nil {
		return nil, 0, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取CC规则列表失败: %v", err))
	}
	return rules, int64(len(rules)), nil
}

// CheckCCLimit 检查是否超过 CC 限制
func (s *ccRuleService) CheckCCLimit(ctx context.Context, uri string) (bool, error) {
	rules, err := s.ccRepo.ListCCRules(ctx, 0, 1000)
	if err != nil {
		return true, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取CC规则列表失败: %v", err))
	}

	if len(rules) == 0 {
		return false, nil
	}

	for _, rule := range rules {
		if rule.URI == uri {
			return s.checkLimit(ctx, rule)
		}
	}
	return false, nil
}

// checkLimit 使用滑动窗口算法进行限流
func (s *ccRuleService) checkLimit(ctx context.Context, rule *model.CCRule) (bool, error) {
	key := fmt.Sprintf("cc_limit:%s", rule.URI)
	now := time.Now()

	limitUnit := rule.LimitUnit
	timeWindow := rule.TimeWindow

	var window time.Duration
	if limitUnit == "second" {
		window = time.Duration(timeWindow) * time.Second
	} else if limitUnit == "minute" {
		window = time.Duration(timeWindow) * time.Minute
	} else if limitUnit == "hour" {
		window = time.Duration(timeWindow) * time.Hour
	} else {
		return true, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("无效的限制单位: %s", limitUnit))
	}

	var windowEnd = now.Add(window)

	var list []time.Time

	var cached interface{}
	err := s.cacheRepo.Get(ctx, key, &cached)
	if err == nil && cached != nil {
		list = cached.([]time.Time)
	}

	var validRequests []time.Time
	for _, requestTime := range list {
		if requestTime.After(now.Add(-window)) && requestTime.Before(windowEnd) {
			validRequests = append(validRequests, requestTime)
		}
	}

	if len(validRequests) >= rule.LimitRate {
		logger.Warnf("CC 防护触发，URI: %s, 当前请求数: %d, 限制: %d", rule.URI, len(validRequests), rule.LimitRate)
		return true, nil
	}

	validRequests = append(validRequests, now)
	if err := s.cacheRepo.Set(ctx, key, validRequests, window); err != nil {
		return false, errors.NewError(errors.ErrCache, fmt.Sprintf("设置CC限制缓存失败: %v", err))
	}

	return false, nil
}

// ReloadRules 重新加载规则
func (s *ccRuleService) ReloadRules(ctx context.Context) error {
	rules, err := s.ccRepo.ListCCRules(ctx, 0, 1000)
	if err != nil {
		return errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取CC规则列表失败: %v", err))
	}

	for _, rule := range rules {
		key := fmt.Sprintf("cc_limit:%s", rule.URI)
		if err := s.cacheRepo.Delete(ctx, key); err != nil {
			logger.Errorf("删除CC限制缓存失败, key: %s, error: %v", key, err)
		}
	}
	return nil
}

// CheckCC 检查CC规则匹配
func (s *ccRuleService) CheckCC(ctx context.Context, ip string, path string, method string) (bool, error) {
	return s.CheckCCLimit(ctx, path)
}
