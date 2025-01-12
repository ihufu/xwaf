package repository

import (
	"context"

	"github.com/xwaf/rule_engine/internal/model"
	"gorm.io/gorm"
)

// IPRuleRepository IP规则仓储接口
type IPRuleRepository interface {
	// CreateIPRule 创建IP规则
	CreateIPRule(ctx context.Context, rule *model.IPRule) error

	// UpdateIPRule 更新IP规则
	UpdateIPRule(ctx context.Context, rule *model.IPRule) error

	// DeleteIPRule 删除IP规则
	DeleteIPRule(ctx context.Context, id int64) error

	// GetIPRule 获取IP规则
	GetIPRule(ctx context.Context, id int64) (*model.IPRule, error)

	// GetIPRuleByIP 根据IP获取规则
	GetIPRuleByIP(ctx context.Context, ip string) (*model.IPRule, error)

	// ListIPRules 获取IP规则列表
	ListIPRules(ctx context.Context, query *model.IPRuleQuery, offset, limit int) ([]*model.IPRule, int64, error)

	// ExistsByIP 检查IP是否存在规则
	ExistsByIP(ctx context.Context, ip string) (bool, error)
}

type IPRepository struct {
	db *gorm.DB
}

func NewIPRepository(db *gorm.DB) *IPRepository {
	return &IPRepository{
		db: db,
	}
}

// CreateIPRule 创建IP规则
func (r *IPRepository) CreateIPRule(ctx context.Context, rule *model.IPRule) error {
	return r.db.Create(rule).Error
}

// UpdateIPRule 更新IP规则
func (r *IPRepository) UpdateIPRule(ctx context.Context, rule *model.IPRule) error {
	return r.db.Save(rule).Error
}

// DeleteIPRule 删除IP规则
func (r *IPRepository) DeleteIPRule(ctx context.Context, id int64) error {
	return r.db.Delete(&model.IPRule{}, id).Error
}

// GetIPRule 获取IP规则
func (r *IPRepository) GetIPRule(ctx context.Context, id int64) (*model.IPRule, error) {
	var rule model.IPRule
	err := r.db.First(&rule, id).Error
	return &rule, err
}

// GetIPRuleByIP 根据IP获取规则
func (r *IPRepository) GetIPRuleByIP(ctx context.Context, ip string) (*model.IPRule, error) {
	var rule model.IPRule
	err := r.db.Where("ip = ?", ip).First(&rule).Error
	return &rule, err
}

// ListIPRules 获取IP规则列表
func (r *IPRepository) ListIPRules(ctx context.Context, query *model.IPRuleQuery, offset, limit int) ([]*model.IPRule, int64, error) {
	var rules []*model.IPRule
	var total int64
	db := r.db.Model(&model.IPRule{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Offset(offset).Limit(limit).Find(&rules).Error
	return rules, total, err
}

// ExistsByIP 检查IP是否存在规则
func (r *IPRepository) ExistsByIP(ctx context.Context, ip string) (bool, error) {
	var count int64
	err := r.db.Model(&model.IPRule{}).Where("ip = ?", ip).Count(&count).Error
	return count > 0, err
}
