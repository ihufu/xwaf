package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/service"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	configService service.WAFConfigService
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(configService service.WAFConfigService) *ConfigHandler {
	if configService == nil {
		panic(errors.NewError(errors.ErrConfig, "配置服务不能为空"))
	}
	return &ConfigHandler{
		configService: configService,
	}
}

// GetMode 获取运行模式
func (h *ConfigHandler) GetMode(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("获取运行模式: RequestID=%s", requestID)

	mode, err := h.configService.GetMode(c)
	if err != nil {
		logger.Errorf("获取运行模式失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取运行模式失败: %v", err)))
		return
	}

	Success(c, gin.H{
		"mode": mode,
	})
}

// UpdateMode 更新运行模式
func (h *ConfigHandler) UpdateMode(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("更新运行模式: RequestID=%s", requestID)

	var req struct {
		Mode        model.WAFMode `json:"mode" binding:"required"`
		Reason      string        `json:"reason" binding:"required"`
		Description string        `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("请求参数错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求参数错误: %v", err)))
		return
	}

	// 验证模式值
	if err := req.Mode.Validate(); err != nil {
		logger.Errorf("无效的运行模式: RequestID=%s, Mode=%s, Error=%v", requestID, req.Mode, err)
		Error(c, err)
		return
	}

	// 验证原因长度
	if len(req.Reason) > 200 {
		logger.Errorf("变更原因过长: RequestID=%s, Length=%d", requestID, len(req.Reason))
		Error(c, errors.NewError(errors.ErrInvalidParams, "变更原因不能超过200个字符"))
		return
	}

	// 获取当前配置
	config, err := h.configService.GetConfig(c)
	if err != nil {
		logger.Errorf("获取配置失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取配置失败: %v", err)))
		return
	}

	// 记录变更日志
	oldMode := config.Mode
	config.Mode = req.Mode
	config.UpdatedAt = time.Now().Unix()
	config.UpdatedBy = c.GetString("operator")
	config.Description = req.Description

	// 更新配置
	if err := h.configService.UpdateConfig(c, config); err != nil {
		logger.Errorf("更新运行模式失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrSystem, fmt.Sprintf("更新运行模式失败: %v", err)))
		return
	}

	// 记录变更日志
	log := &model.WAFModeChangeLog{
		OldMode:     oldMode,
		NewMode:     req.Mode,
		Operator:    c.GetString("operator"),
		Reason:      req.Reason,
		Description: req.Description,
		CreatedAt:   time.Now().Unix(),
	}

	if err := h.configService.LogModeChange(c, log); err != nil {
		logger.Errorf("记录模式变更日志失败: RequestID=%s, Error=%v", requestID, err)
		// 仅记录错误，不影响主流程
		c.Error(errors.NewError(errors.ErrSystem, fmt.Sprintf("记录模式变更日志失败: %v", err)))
	}

	logger.Infof("更新运行模式成功: RequestID=%s, OldMode=%s, NewMode=%s", requestID, oldMode, req.Mode)
	Success(c, gin.H{
		"mode":       req.Mode,
		"updated_at": config.UpdatedAt,
	})
}

// GetModeChangeLogs 获取模式变更日志
func (h *ConfigHandler) GetModeChangeLogs(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("获取模式变更日志: RequestID=%s", requestID)

	var req struct {
		StartTime int64 `form:"start_time"`
		EndTime   int64 `form:"end_time"`
		Page      int   `form:"page,default=1"`
		PageSize  int   `form:"page_size,default=20"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Errorf("请求参数错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求参数错误: %v", err)))
		return
	}

	// 验证分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// 验证时间范围
	if req.EndTime > 0 && req.StartTime > req.EndTime {
		logger.Errorf("无效的时间范围: RequestID=%s, StartTime=%d, EndTime=%d",
			requestID, req.StartTime, req.EndTime)
		Error(c, errors.NewError(errors.ErrInvalidParams, "开始时间不能大于结束时间"))
		return
	}

	logs, total, err := h.configService.GetModeChangeLogs(c, req.StartTime, req.EndTime, req.Page, req.PageSize)
	if err != nil {
		logger.Errorf("获取变更日志失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取变更日志失败: %v", err)))
		return
	}

	logger.Infof("获取模式变更日志成功: RequestID=%s, Total=%d", requestID, total)
	Success(c, gin.H{
		"total": total,
		"logs":  logs,
	})
}
