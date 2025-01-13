package repository

import (
	"context"
	"fmt"

	"github.com/xwaf/rule_engine/internal/errors"
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
	if err := r.db.WithContext(ctx).Create(rule).Error; err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("创建IP规则失败: %v", err))
	}
	return nil
}

// UpdateIPRule 更新IP规则
func (r *IPRepository) UpdateIPRule(ctx context.Context, rule *model.IPRule) error {
	result := r.db.WithContext(ctx).Save(rule)
	if result.Error != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("更新IP规则失败: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("IP规则不存在: %d", rule.ID))
	}
	return nil
}

// DeleteIPRule 删除IP规则
func (r *IPRepository) DeleteIPRule(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&model.IPRule{}, id)
	if result.Error != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("删除IP规则失败: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("IP规则不存在: %d", id))
	}
	return nil
}

// GetIPRule 获取IP规则
func (r *IPRepository) GetIPRule(ctx context.Context, id int64) (*model.IPRule, error) {
	var rule model.IPRule
	err := r.db.WithContext(ctx).First(&rule, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("IP规则不存在: %d", id))
	}
	if err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取IP规则失败: %v", err))
	}
	return &rule, nil
}

// GetIPRuleByIP 根据IP获取规则
func (r *IPRepository) GetIPRuleByIP(ctx context.Context, ip string) (*model.IPRule, error) {
	var rule model.IPRule
	err := r.db.WithContext(ctx).Where("ip = ?", ip).First(&rule).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("IP规则不存在: %s", ip))
	}
	if err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取IP规则失败: %v", err))
	}
	return &rule, nil
}

// ListIPRules 获取IP规则列表
func (r *IPRepository) ListIPRules(ctx context.Context, query *model.IPRuleQuery, offset, limit int) ([]*model.IPRule, int64, error) {
	var rules []*model.IPRule
	var total int64

	db := r.db.WithContext(ctx).Model(&model.IPRule{})

	// 添加查询条件
	if query != nil {
		if query.Keyword != "" {
			db = db.Where("ip LIKE ? OR description LIKE ?",
				"%"+query.Keyword+"%", "%"+query.Keyword+"%")
		}
		if query.IPType != "" {
			db = db.Where("ip_type = ?", query.IPType)
		}
		if query.BlockType != "" {
			db = db.Where("block_type = ?", query.BlockType)
		}
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取IP规则总数失败: %v", err))
	}

	// 分页查询
	if err := db.Offset(offset).Limit(limit).Find(&rules).Error; err != nil {
		return nil, 0, errors.NewError(errors.ErrSystem, fmt.Sprintf("查询IP规则列表失败: %v", err))
	}

	return rules, total, nil
}

// ExistsByIP 检查IP是否存在规则
func (r *IPRepository) ExistsByIP(ctx context.Context, ip string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.IPRule{}).Where("ip = ?", ip).Count(&count).Error
	if err != nil {
		return false, errors.NewError(errors.ErrSystem, fmt.Sprintf("检查IP规则是否存在失败: %v", err))
	}
	return count > 0, nil
}
