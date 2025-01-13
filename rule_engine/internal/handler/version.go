package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/service"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// RuleVersionHandler 规则版本处理器
type RuleVersionHandler struct {
	versionService service.RuleVersionService
}

// NewRuleVersionHandler 创建规则版本处理器
func NewRuleVersionHandler(versionService service.RuleVersionService) *RuleVersionHandler {
	if versionService == nil {
		panic("version service cannot be nil")
	}
	return &RuleVersionHandler{
		versionService: versionService,
	}
}

// CreateVersion 创建规则版本
func (h *RuleVersionHandler) CreateVersion(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("创建规则版本: RequestID=%s", requestID)

	var version model.RuleVersion
	if err := c.ShouldBindJSON(&version); err != nil {
		logger.Errorf("请求数据格式错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求数据格式错误: %v", err)))
		return
	}

	if err := h.versionService.CreateVersion(c.Request.Context(), &version); err != nil {
		logger.Errorf("创建规则版本失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("创建规则版本失败: %v", err)))
		return
	}

	logger.Infof("创建规则版本成功: RequestID=%s, RuleID=%d, Version=%d", requestID, version.RuleID, version.Version)
	Success(c, version)
}

// GetVersion 获取规则版本
func (h *RuleVersionHandler) GetVersion(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("获取规则版本: RequestID=%s", requestID)

	ruleID, err := strconv.ParseInt(c.Param("rule_id"), 10, 64)
	if err != nil {
		logger.Errorf("无效的规则ID: RequestID=%s, RuleID=%s", requestID, c.Param("rule_id"))
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的规则ID: %s", c.Param("rule_id"))))
		return
	}

	version, err := strconv.ParseInt(c.Param("version"), 10, 64)
	if err != nil {
		logger.Errorf("无效的版本号: RequestID=%s, Version=%s", requestID, c.Param("version"))
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的版本号: %s", c.Param("version"))))
		return
	}

	v, err := h.versionService.GetVersion(c.Request.Context(), ruleID, version)
	if err != nil {
		logger.Errorf("获取规则版本失败: RequestID=%s, RuleID=%d, Version=%d, Error=%v", requestID, ruleID, version, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则版本失败: %v", err)))
		return
	}

	if v == nil {
		logger.Errorf("规则版本不存在: RequestID=%s, RuleID=%d, Version=%d", requestID, ruleID, version)
		Error(c, errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("规则版本不存在: RuleID=%d, Version=%d", ruleID, version)))
		return
	}

	logger.Infof("获取规则版本成功: RequestID=%s, RuleID=%d, Version=%d", requestID, ruleID, version)
	Success(c, v)
}

// ListVersions 获取规则版本列表
func (h *RuleVersionHandler) ListVersions(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("获取规则版本列表: RequestID=%s", requestID)

	ruleID, err := strconv.ParseInt(c.Param("rule_id"), 10, 64)
	if err != nil {
		logger.Errorf("无效的规则ID: RequestID=%s, RuleID=%s", requestID, c.Param("rule_id"))
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的规则ID: %s", c.Param("rule_id"))))
		return
	}

	versions, err := h.versionService.ListVersions(c.Request.Context(), ruleID)
	if err != nil {
		logger.Errorf("获取规则版本列表失败: RequestID=%s, RuleID=%d, Error=%v", requestID, ruleID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则版本列表失败: %v", err)))
		return
	}

	logger.Infof("获取规则版本列表成功: RequestID=%s, RuleID=%d, Count=%d", requestID, ruleID, len(versions))
	Success(c, versions)
}

// GetSyncLogs 获取同步日志
func (h *RuleVersionHandler) GetSyncLogs(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("获取同步日志: RequestID=%s", requestID)

	ruleID, err := strconv.ParseInt(c.Param("rule_id"), 10, 64)
	if err != nil {
		logger.Errorf("无效的规则ID: RequestID=%s, RuleID=%s", requestID, c.Param("rule_id"))
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的规则ID: %s", c.Param("rule_id"))))
		return
	}

	logs, err := h.versionService.GetSyncLogs(c.Request.Context(), ruleID)
	if err != nil {
		logger.Errorf("获取同步日志失败: RequestID=%s, RuleID=%d, Error=%v", requestID, ruleID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取同步日志失败: %v", err)))
		return
	}

	logger.Infof("获取同步日志成功: RequestID=%s, RuleID=%d, Count=%d", requestID, ruleID, len(logs))
	Success(c, logs)
}
