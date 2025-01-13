package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/service"
)

// RuleVersionHandler 规则版本处理器
type RuleVersionHandler struct {
	versionService service.RuleVersionService
}

// NewRuleVersionHandler 创建规则版本处理器
func NewRuleVersionHandler(versionService service.RuleVersionService) *RuleVersionHandler {
	return &RuleVersionHandler{
		versionService: versionService,
	}
}

// CreateVersion 创建规则版本
func (h *RuleVersionHandler) CreateVersion(c *gin.Context) {
	var version model.RuleVersion
	if err := c.ShouldBindJSON(&version); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.versionService.CreateVersion(c.Request.Context(), &version); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, version)
}

// GetVersion 获取规则版本
func (h *RuleVersionHandler) GetVersion(c *gin.Context) {
	ruleID, err := strconv.ParseInt(c.Param("rule_id"), 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	version, err := strconv.ParseInt(c.Param("version"), 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的版本号"))
		return
	}

	v, err := h.versionService.GetVersion(c.Request.Context(), ruleID, version)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	if v == nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, "版本不存在"))
		return
	}

	Success(c, v)
}

// ListVersions 获取规则版本列表
func (h *RuleVersionHandler) ListVersions(c *gin.Context) {
	ruleID, err := strconv.ParseInt(c.Param("rule_id"), 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	versions, err := h.versionService.ListVersions(c.Request.Context(), ruleID)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, versions)
}

// GetSyncLogs 获取同步日志
func (h *RuleVersionHandler) GetSyncLogs(c *gin.Context) {
	ruleID, err := strconv.ParseInt(c.Param("rule_id"), 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	logs, err := h.versionService.GetSyncLogs(c.Request.Context(), ruleID)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, logs)
}
