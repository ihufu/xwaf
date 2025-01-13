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

// WAFConfigService WAF配置服务接口
type WAFConfigService interface {
	// GetConfig 获取WAF配置
	GetConfig(ctx context.Context) (*model.WAFConfig, error)

	// UpdateConfig 更新WAF配置
	UpdateConfig(ctx context.Context, config *model.WAFConfig) error

	// GetMode 获取WAF运行模式
	GetMode(ctx context.Context) (model.WAFMode, error)

	// LogModeChange 记录模式变更日志
	LogModeChange(ctx context.Context, log *model.WAFModeChangeLog) error

	// GetModeChangeLogs 获取模式变更日志
	GetModeChangeLogs(ctx context.Context, startTime, endTime int64, page, pageSize int) ([]*model.WAFModeChangeLog, int64, error)
}

// wafConfigService WAF配置服务实现
type wafConfigService struct {
	configRepo repository.WAFConfigRepository
	cacheRepo  repository.CacheRepository
}

// NewWAFConfigService 创建WAF配置服务
func NewWAFConfigService(configRepo repository.WAFConfigRepository, cacheRepo repository.CacheRepository) WAFConfigService {
	return &wafConfigService{
		configRepo: configRepo,
		cacheRepo:  cacheRepo,
	}
}

// GetConfig 获取WAF配置
func (s *wafConfigService) GetConfig(ctx context.Context) (*model.WAFConfig, error) {
	if ctx == nil {
		return nil, errors.NewError(errors.ErrConfig, "上下文不能为空")
	}

	// 从缓存获取
	config, err := s.getConfigFromCache(ctx)
	if err == nil && config != nil {
		return config, nil
	}

	// 从数据库获取
	config, err = s.configRepo.GetConfig(ctx)
	if err != nil {
		return nil, errors.NewError(errors.ErrConfig, fmt.Sprintf("获取WAF配置失败: %v", err))
	}

	// 更新缓存
	if config != nil {
		if err := s.updateConfigCache(ctx, config); err != nil {
			logger.Warnf("更新WAF配置缓存失败: %v", err)
		}
	}

	return config, nil
}

// UpdateConfig 更新WAF配置
func (s *wafConfigService) UpdateConfig(ctx context.Context, config *model.WAFConfig) error {
	if ctx == nil {
		return errors.NewError(errors.ErrConfig, "上下文不能为空")
	}
	if config == nil {
		return errors.NewError(errors.ErrConfig, "配置不能为空")
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return errors.NewError(errors.ErrConfig, fmt.Sprintf("配置验证失败: %v", err))
	}

	// 更新配置
	if err := s.configRepo.UpdateConfig(ctx, config); err != nil {
		return errors.NewError(errors.ErrConfig, fmt.Sprintf("更新WAF配置失败: %v", err))
	}

	// 更新缓存
	return s.updateConfigCache(ctx, config)
}

// GetMode 获取WAF运行模式
func (s *wafConfigService) GetMode(ctx context.Context) (model.WAFMode, error) {
	if ctx == nil {
		return "", errors.NewError(errors.ErrConfig, "上下文不能为空")
	}

	// 从缓存获取
	mode, err := s.getModeFromCache(ctx)
	if err == nil {
		return mode, nil
	}

	// 从数据库获取
	config, err := s.GetConfig(ctx)
	if err != nil {
		return "", errors.NewError(errors.ErrConfig, fmt.Sprintf("获取WAF配置失败: %v", err))
	}
	if config == nil {
		return model.WAFModeBlock, nil // 默认为阻断模式
	}

	return config.Mode, nil
}

// 缓存相关的辅助方法
func (s *wafConfigService) updateConfigCache(ctx context.Context, config *model.WAFConfig) error {
	if ctx == nil {
		return errors.NewError(errors.ErrCache, "上下文不能为空")
	}
	if config == nil {
		return errors.NewError(errors.ErrCache, "配置不能为空")
	}

	key := "waf:config"
	if err := s.cacheRepo.Set(ctx, key, config, 24*time.Hour); err != nil {
		return errors.NewError(errors.ErrCache, fmt.Sprintf("设置WAF配置缓存失败: %v", err))
	}
	return nil
}

func (s *wafConfigService) getConfigFromCache(ctx context.Context) (*model.WAFConfig, error) {
	if ctx == nil {
		return nil, errors.NewError(errors.ErrCache, "上下文不能为空")
	}

	key := "waf:config"
	var config model.WAFConfig
	err := s.cacheRepo.Get(ctx, key, &config)
	if err != nil {
		return nil, errors.NewError(errors.ErrCache, fmt.Sprintf("获取WAF配置缓存失败: %v", err))
	}
	return &config, nil
}

func (s *wafConfigService) getModeFromCache(ctx context.Context) (model.WAFMode, error) {
	if ctx == nil {
		return "", errors.NewError(errors.ErrCache, "上下文不能为空")
	}

	key := "waf:mode"
	var mode model.WAFMode
	err := s.cacheRepo.Get(ctx, key, &mode)
	if err != nil {
		return "", errors.NewError(errors.ErrCache, fmt.Sprintf("获取WAF模式缓存失败: %v", err))
	}
	return mode, nil
}

// LogModeChange 记录模式变更日志
func (s *wafConfigService) LogModeChange(ctx context.Context, log *model.WAFModeChangeLog) error {
	if ctx == nil {
		return errors.NewError(errors.ErrConfig, "上下文不能为空")
	}
	if log == nil {
		return errors.NewError(errors.ErrConfig, "日志不能为空")
	}

	if err := s.configRepo.LogModeChange(ctx, log); err != nil {
		return errors.NewError(errors.ErrConfig, fmt.Sprintf("记录WAF模式变更日志失败: %v", err))
	}
	return nil
}

// GetModeChangeLogs 获取模式变更日志
func (s *wafConfigService) GetModeChangeLogs(ctx context.Context, startTime, endTime int64, page, pageSize int) ([]*model.WAFModeChangeLog, int64, error) {
	if ctx == nil {
		return nil, 0, errors.NewError(errors.ErrConfig, "上下文不能为空")
	}
	if startTime > endTime {
		return nil, 0, errors.NewError(errors.ErrConfig, "开始时间不能大于结束时间")
	}
	if page <= 0 || pageSize <= 0 {
		return nil, 0, errors.NewError(errors.ErrConfig, "分页参数必须大于0")
	}

	logs, total, err := s.configRepo.GetModeChangeLogs(ctx, startTime, endTime, page, pageSize)
	if err != nil {
		return nil, 0, errors.NewError(errors.ErrConfig, fmt.Sprintf("获取WAF模式变更日志失败: %v", err))
	}
	return logs, total, nil
}
