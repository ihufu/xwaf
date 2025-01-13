package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/service"
)

// IPRuleHandler IP规则处理器
type IPRuleHandler struct {
	ipRuleService service.IPRuleService
}

// NewIPRuleHandler 创建IP规则处理器
func NewIPRuleHandler(ipRuleService service.IPRuleService) *IPRuleHandler {
	return &IPRuleHandler{
		ipRuleService: ipRuleService,
	}
}

// CreateIPRule 创建IP规则
func (h *IPRuleHandler) CreateIPRule(c *gin.Context) {
	var rule model.IPRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.ipRuleService.CreateIPRule(c.Request.Context(), &rule); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, rule)
}

// UpdateIPRule 更新IP规则
func (h *IPRuleHandler) UpdateIPRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	var rule model.IPRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	rule.ID = ruleID
	if err := h.ipRuleService.UpdateIPRule(c.Request.Context(), &rule); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, rule)
}

// DeleteIPRule 删除IP规则
func (h *IPRuleHandler) DeleteIPRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	if err := h.ipRuleService.DeleteIPRule(c.Request.Context(), ruleID); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, nil)
}

// GetIPRule 获取IP规则
func (h *IPRuleHandler) GetIPRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	rule, err := h.ipRuleService.GetIPRule(c.Request.Context(), ruleID)
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

// ListIPRules 获取IP规则列表
func (h *IPRuleHandler) ListIPRules(c *gin.Context) {
	var query model.IPRuleQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

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

	rules, total, err := h.ipRuleService.ListIPRules(c.Request.Context(), query, page, size)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, gin.H{
		"total": total,
		"items": rules,
	})
}

// CheckIP 检查IP是否命中规则
func (h *IPRuleHandler) CheckIP(c *gin.Context) {
	var req struct {
		IP string `json:"ip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	isBlocked, err := h.ipRuleService.IsIPBlocked(c.Request.Context(), req.IP)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	isWhitelisted, err := h.ipRuleService.IsIPWhitelisted(c.Request.Context(), req.IP)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, gin.H{
		"is_blocked":     isBlocked,
		"is_whitelisted": isWhitelisted,
	})
}
