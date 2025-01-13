package repository

import (
	"context"

	"github.com/xwaf/rule_engine/internal/model"
)

// CCRuleRepository CC规则仓储接口
type CCRuleRepository interface {
	// CreateCCRule 创建CC规则
	CreateCCRule(ctx context.Context, rule *model.CCRule) error

	// UpdateCCRule 更新CC规则
	UpdateCCRule(ctx context.Context, rule *model.CCRule) error

	// DeleteCCRule 删除CC规则
	DeleteCCRule(ctx context.Context, id int64) error

	// GetCCRule 获取CC规则
	GetCCRule(ctx context.Context, id int64) (*model.CCRule, error)

	// ListCCRules 获取CC规则列表
	ListCCRules(ctx context.Context, offset, limit int) ([]*model.CCRule, error)
}
