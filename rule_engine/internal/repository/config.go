package repository

import (
	"context"

	"github.com/xwaf/rule_engine/internal/model"
)

// WAFConfigRepository WAF配置仓储接口
type WAFConfigRepository interface {
	// GetConfig 获取WAF配置
	GetConfig(ctx context.Context) (*model.WAFConfig, error)

	// UpdateConfig 更新WAF配置
	UpdateConfig(ctx context.Context, config *model.WAFConfig) error
}
