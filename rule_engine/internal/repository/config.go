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

	// LogModeChange 记录模式变更日志
	LogModeChange(ctx context.Context, log *model.WAFModeChangeLog) error

	// GetModeChangeLogs 获取模式变更日志
	GetModeChangeLogs(ctx context.Context, startTime, endTime int64, page, pageSize int) ([]*model.WAFModeChangeLog, int64, error)
}
