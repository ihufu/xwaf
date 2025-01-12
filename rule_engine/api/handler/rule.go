package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
	"github.com/xwaf/rule_engine/internal/service"
)

// RuleHandler API规则处理器
type RuleHandler struct {
	ruleService service.RuleService
}

// NewRuleHandler 创建API规则处理器
func NewRuleHandler(ruleService service.RuleService) *RuleHandler {
	return &RuleHandler{
		ruleService: ruleService,
	}
}

// CreateRule 创建规则
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.ruleService.CreateRule(c.Request.Context(), &rule); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, rule)
}

// UpdateRule 更新规则
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	rule.ID = ruleID
	if err := h.ruleService.UpdateRule(c.Request.Context(), &rule); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, rule)
}

// DeleteRule 删除规则
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	if err := h.ruleService.DeleteRule(c.Request.Context(), ruleID); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, nil)
}

// GetRule 获取规则
func (h *RuleHandler) GetRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	rule, err := h.ruleService.GetRule(c.Request.Context(), ruleID)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	if rule == nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, "规则不存在"))
		return
	}

	Success(c, rule)
}

// ListRules 获取规则列表
func (h *RuleHandler) ListRules(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的页码"))
		return
	}
	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的页大小"))
		return
	}

	query := &repository.RuleQuery{
		Page:     page,
		PageSize: size,
	}

	rules, total, err := h.ruleService.ListRules(c.Request.Context(), query)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, gin.H{
		"total": total,
		"items": rules,
	})
}

// CheckRule 检查规则匹配
func (h *RuleHandler) CheckRule(c *gin.Context) {
	var req model.CheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	result, err := h.ruleService.CheckRequest(c.Request.Context(), &req)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, result)
}

// ReloadRules 重新加载规则
func (h *RuleHandler) ReloadRules(c *gin.Context) {
	err := h.ruleService.ReloadRules(c.Request.Context())
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, nil)
}
