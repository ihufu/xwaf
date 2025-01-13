package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/service"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	configService service.WAFConfigService
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(configService service.WAFConfigService) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
	}
}

// GetMode 获取运行模式
func (h *ConfigHandler) GetMode(c *gin.Context) {
	mode, err := h.configService.GetMode(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取运行模式失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"mode": mode,
		},
		"message": "success",
	})
}

// UpdateMode 更新运行模式
func (h *ConfigHandler) UpdateMode(c *gin.Context) {
	var req struct {
		Mode        model.WAFMode `json:"mode" binding:"required"`
		Reason      string        `json:"reason" binding:"required"`
		Description string        `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 获取当前配置
	config, err := h.configService.GetConfig(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取配置失败",
			"error":   err.Error(),
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新运行模式失败",
			"error":   err.Error(),
		})
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
		// 仅记录错误，不影响主流程
		c.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"mode":       req.Mode,
			"updated_at": config.UpdatedAt,
		},
	})
}

// GetModeChangeLogs 获取模式变更日志
func (h *ConfigHandler) GetModeChangeLogs(c *gin.Context) {
	var req struct {
		StartTime int64 `form:"start_time"`
		EndTime   int64 `form:"end_time"`
		Page      int   `form:"page,default=1"`
		PageSize  int   `form:"page_size,default=20"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	logs, total, err := h.configService.GetModeChangeLogs(c, req.StartTime, req.EndTime, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取变更日志失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"total": total,
			"logs":  logs,
		},
	})
}
