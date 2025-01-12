package service

import (
	"context"
	"fmt"
	"time"

	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// IPRuleService IP规则服务接口
type IPRuleService interface {
	CreateIPRule(ctx context.Context, rule *model.IPRule) error
	UpdateIPRule(ctx context.Context, rule *model.IPRule) error
	DeleteIPRule(ctx context.Context, id int64) error
	GetIPRule(ctx context.Context, id int64) (*model.IPRule, error)
	ListIPRules(ctx context.Context, query model.IPRuleQuery, page, size int) ([]*model.IPRule, int64, error)
	IsIPBlocked(ctx context.Context, ip string) (bool, error)
	IsIPWhitelisted(ctx context.Context, ip string) (bool, error)
	CheckIP(ctx context.Context, ip string) (bool, error)
}

// ipRuleService IP规则服务实现
type ipRuleService struct {
	ipRepo    repository.IPRuleRepository
	cacheRepo repository.CacheRepository
}

// NewIPRuleService 创建IP规则服务
func NewIPRuleService(ipRepo repository.IPRuleRepository, cacheRepo repository.CacheRepository) IPRuleService {
	return &ipRuleService{
		ipRepo:    ipRepo,
		cacheRepo: cacheRepo,
	}
}

// CreateIPRule 创建IP规则
func (s *ipRuleService) CreateIPRule(ctx context.Context, rule *model.IPRule) error {
	// 验证规则
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("规则验证失败: %v", err)
	}

	// 检查IP是否已存在
	exists, err := s.ipRepo.ExistsByIP(ctx, rule.IP)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("IP %s 已存在规则", rule.IP)
	}

	// 创建规则
	if err := s.ipRepo.CreateIPRule(ctx, rule); err != nil {
		return err
	}

	// 更新缓存
	return s.updateIPRuleCache(ctx, rule)
}

// UpdateIPRule 更新IP规则
func (s *ipRuleService) UpdateIPRule(ctx context.Context, rule *model.IPRule) error {
	// 验证规则
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("规则验证失败: %v", err)
	}

	// 检查规则是否存在
	oldRule, err := s.ipRepo.GetIPRule(ctx, rule.ID)
	if err != nil {
		return err
	}
	if oldRule == nil {
		return fmt.Errorf("规则不存在: %d", rule.ID)
	}

	// 更新规则
	if err := s.ipRepo.UpdateIPRule(ctx, rule); err != nil {
		return err
	}

	// 更新缓存
	return s.updateIPRuleCache(ctx, rule)
}

// DeleteIPRule 删除IP规则
func (s *ipRuleService) DeleteIPRule(ctx context.Context, id int64) error {
	// 删除规则
	if err := s.ipRepo.DeleteIPRule(ctx, id); err != nil {
		return err
	}

	// 删除缓存
	return s.deleteIPRuleCache(ctx, id)
}

// GetIPRule 获取IP规则
func (s *ipRuleService) GetIPRule(ctx context.Context, id int64) (*model.IPRule, error) {
	// 从缓存获取
	rule, err := s.getIPRuleFromCache(ctx, id)
	if err == nil && rule != nil {
		return rule, nil
	}

	// 从数据库获取
	rule, err = s.ipRepo.GetIPRule(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	if rule != nil {
		if err := s.updateIPRuleCache(ctx, rule); err != nil {
			logger.Warn("更新IP规则缓存失败:", err)
		}
	}

	return rule, nil
}

// ListIPRules 获取IP规则列表
func (s *ipRuleService) ListIPRules(ctx context.Context, query model.IPRuleQuery, page, size int) ([]*model.IPRule, int64, error) {
	offset := (page - 1) * size
	queryPtr := &query
	return s.ipRepo.ListIPRules(ctx, queryPtr, offset, size)
}

// IsIPBlocked 检查IP是否被封禁
func (s *ipRuleService) IsIPBlocked(ctx context.Context, ip string) (bool, error) {
	// 从缓存检查
	blocked, err := s.isIPBlockedFromCache(ctx, ip)
	if err == nil {
		return blocked, nil
	}

	// 从数据库查询
	rule, err := s.ipRepo.GetIPRuleByIP(ctx, ip)
	if err != nil {
		return false, err
	}

	// 检查是否在黑名单且未过期
	if rule != nil && rule.IPType == model.IPListTypeBlack {
		if rule.BlockType == model.BlockTypePermanent {
			return true, nil
		}
		return time.Now().Before(rule.ExpireTime), nil
	}

	return false, nil
}

// IsIPWhitelisted 检查IP是否在白名单
func (s *ipRuleService) IsIPWhitelisted(ctx context.Context, ip string) (bool, error) {
	// 从缓存检查
	whitelisted, err := s.isIPWhitelistedFromCache(ctx, ip)
	if err == nil {
		return whitelisted, nil
	}

	// 从数据库查询
	rule, err := s.ipRepo.GetIPRuleByIP(ctx, ip)
	if err != nil {
		return false, err
	}

	return rule != nil && rule.IPType == model.IPListTypeWhite, nil
}

// CheckIP 检查IP是否被规则阻止
func (s *ipRuleService) CheckIP(ctx context.Context, ip string) (bool, error) {
	rules, _, err := s.ListIPRules(ctx, model.IPRuleQuery{}, 1, 1000)
	if err != nil {
		return false, err
	}

	for _, rule := range rules {
		if rule.IP == ip {
			return true, nil
		}
	}
	return false, nil
}

// 缓存相关的辅助方法
func (s *ipRuleService) updateIPRuleCache(ctx context.Context, rule *model.IPRule) error {
	key := fmt.Sprintf("ip_rule:%d", rule.ID)
	return s.cacheRepo.Set(ctx, key, rule, 24*time.Hour)
}

func (s *ipRuleService) deleteIPRuleCache(ctx context.Context, id int64) error {
	key := fmt.Sprintf("ip_rule:%d", id)
	return s.cacheRepo.Delete(ctx, key)
}

func (s *ipRuleService) getIPRuleFromCache(ctx context.Context, id int64) (*model.IPRule, error) {
	key := fmt.Sprintf("ip_rule:%d", id)
	var rule model.IPRule
	err := s.cacheRepo.Get(ctx, key, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *ipRuleService) isIPBlockedFromCache(ctx context.Context, ip string) (bool, error) {
	key := fmt.Sprintf("ip_blocked:%s", ip)
	var blocked bool
	err := s.cacheRepo.Get(ctx, key, &blocked)
	if err != nil {
		return false, err
	}
	return blocked, nil
}

func (s *ipRuleService) isIPWhitelistedFromCache(ctx context.Context, ip string) (bool, error) {
	key := fmt.Sprintf("ip_whitelisted:%s", ip)
	var whitelisted bool
	err := s.cacheRepo.Get(ctx, key, &whitelisted)
	if err != nil {
		return false, err
	}
	return whitelisted, nil
}