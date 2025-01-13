package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/service"
)

// CCRuleHandler CC规则处理器
type CCRuleHandler struct {
	ccService service.CCRuleService
}

// NewCCRuleHandler 创建CC规则处理器
func NewCCRuleHandler(ccService service.CCRuleService) *CCRuleHandler {
	return &CCRuleHandler{
		ccService: ccService,
	}
}

// CreateCCRule 创建CC规则
func (h *CCRuleHandler) CreateCCRule(c *gin.Context) {
	var rule model.CCRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.ccService.CreateCCRule(c.Request.Context(), &rule); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, rule)
}

// UpdateCCRule 更新CC规则
func (h *CCRuleHandler) UpdateCCRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	var rule model.CCRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}
	rule.ID = id

	if err := h.ccService.UpdateCCRule(c.Request.Context(), &rule); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, rule)
}

// DeleteCCRule 删除CC规则
func (h *CCRuleHandler) DeleteCCRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	if err := h.ccService.DeleteCCRule(c.Request.Context(), id); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, nil)
}

// GetCCRule 获取CC规则
func (h *CCRuleHandler) GetCCRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	rule, err := h.ccService.GetCCRule(c.Request.Context(), id)
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

// ListCCRules 获取CC规则列表
func (h *CCRuleHandler) ListCCRules(c *gin.Context) {
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

	rules, total, err := h.ccService.ListCCRules(c.Request.Context(), model.CCRuleQuery{}, page, size)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, gin.H{
		"items": rules,
		"pagination": gin.H{
			"page":  page,
			"size":  size,
			"total": total,
		},
	})
}

// CheckCCLimit 检查CC限制
func (h *CCRuleHandler) CheckCCLimit(c *gin.Context) {
	uri := c.Param("uri")
	if uri == "" {
		Error(c, errors.NewError(errors.ErrInvalidParams, "URI不能为空"))
		return
	}

	isLimited, err := h.ccService.CheckCCLimit(c.Request.Context(), uri)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, gin.H{
		"is_limited": isLimited,
	})
}

// CheckCC 检查CC规则
func (h *CCRuleHandler) CheckCC(c *gin.Context) {
	var req struct {
		IP     string `json:"ip" binding:"required"`
		Path   string `json:"path" binding:"required"`
		Method string `json:"method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	isBlocked, err := h.ccService.CheckCC(c.Request.Context(), req.IP, req.Path, req.Method)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, gin.H{
		"is_blocked": isBlocked,
	})
}
