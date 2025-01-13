package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	xerrors "github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/service"
)

// IPRuleHandler IP规则处理器
type IPRuleHandler struct {
	ipRuleService service.IPRuleService
}

// NewIPRuleHandler 创建IP规则处理器
func NewIPRuleHandler(ipRuleService service.IPRuleService) *IPRuleHandler {
	if ipRuleService == nil {
		panic("IP规则服务不能为空")
	}
	return &IPRuleHandler{
		ipRuleService: ipRuleService,
	}
}

// CreateIPRule 创建IP规则
func (h *IPRuleHandler) CreateIPRule(c *gin.Context) {
	var rule model.IPRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		ValidationError(c, "IP规则数据格式错误: "+err.Error())
		return
	}

	// 验证规则数据
	if err := rule.Validate(); err != nil {
		ValidationError(c, err.Error())
		return
	}

	if err := h.ipRuleService.CreateIPRule(c.Request.Context(), &rule); err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrRuleValidation:
				ValidationError(c, err.Error())
			case xerrors.ErrRuleConflict:
				ConflictError(c, err.Error())
			default:
				SystemError(c, "创建IP规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "创建IP规则失败: "+err.Error())
		}
		return
	}

	Success(c, rule)
}

// UpdateIPRule 更新IP规则
func (h *IPRuleHandler) UpdateIPRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ValidationError(c, "无效的规则ID")
		return
	}

	var rule model.IPRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		ValidationError(c, "IP规则数据格式错误: "+err.Error())
		return
	}

	// 验证规则数据
	if err := rule.Validate(); err != nil {
		ValidationError(c, err.Error())
		return
	}

	rule.ID = ruleID
	if err := h.ipRuleService.UpdateIPRule(c.Request.Context(), &rule); err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrRuleValidation:
				ValidationError(c, err.Error())
			case xerrors.ErrRuleNotFound:
				NotFoundError(c, err.Error())
			case xerrors.ErrRuleConflict:
				ConflictError(c, err.Error())
			default:
				SystemError(c, "更新IP规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "更新IP规则失败: "+err.Error())
		}
		return
	}

	Success(c, rule)
}

// DeleteIPRule 删除IP规则
func (h *IPRuleHandler) DeleteIPRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ValidationError(c, "无效的规则ID")
		return
	}

	if err := h.ipRuleService.DeleteIPRule(c.Request.Context(), ruleID); err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrRuleNotFound:
				NotFoundError(c, err.Error())
			default:
				SystemError(c, "删除IP规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "删除IP规则失败: "+err.Error())
		}
		return
	}

	Success(c, nil)
}

// GetIPRule 获取IP规则
func (h *IPRuleHandler) GetIPRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ValidationError(c, "无效的规则ID")
		return
	}

	rule, err := h.ipRuleService.GetIPRule(c.Request.Context(), ruleID)
	if err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrRuleNotFound:
				NotFoundError(c, err.Error())
			default:
				SystemError(c, "获取IP规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "获取IP规则失败: "+err.Error())
		}
		return
	}

	if rule == nil {
		NotFoundError(c, "IP规则不存在")
		return
	}

	Success(c, rule)
}

// ListIPRules 获取IP规则列表
func (h *IPRuleHandler) ListIPRules(c *gin.Context) {
	var query model.IPRuleQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		ValidationError(c, "查询参数格式错误: "+err.Error())
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		ValidationError(c, "无效的页码")
		return
	}
	if page <= 0 {
		ValidationError(c, "页码必须大于0")
		return
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		ValidationError(c, "无效的页大小")
		return
	}
	if size <= 0 {
		ValidationError(c, "每页大小必须大于0")
		return
	}
	if size > 100 {
		ValidationError(c, "每页大小不能超过100")
		return
	}

	rules, total, err := h.ipRuleService.ListIPRules(c.Request.Context(), query, page, size)
	if err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrValidation:
				ValidationError(c, err.Error())
			default:
				SystemError(c, "查询IP规则列表失败: "+err.Error())
			}
		} else {
			SystemError(c, "查询IP规则列表失败: "+err.Error())
		}
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
		ValidationError(c, "请求数据格式错误: "+err.Error())
		return
	}

	// 验证IP地址格式
	if err := (&model.IPRule{IP: req.IP}).Validate(); err != nil {
		ValidationError(c, err.Error())
		return
	}

	isBlocked, err := h.ipRuleService.IsIPBlocked(c.Request.Context(), req.IP)
	if err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrCache:
				ServiceUnavailableError(c, err.Error())
			default:
				SystemError(c, "检查IP是否被封禁失败: "+err.Error())
			}
		} else {
			SystemError(c, "检查IP是否被封禁失败: "+err.Error())
		}
		return
	}

	isWhitelisted, err := h.ipRuleService.IsIPWhitelisted(c.Request.Context(), req.IP)
	if err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrCache:
				ServiceUnavailableError(c, err.Error())
			default:
				SystemError(c, "检查IP是否在白名单失败: "+err.Error())
			}
		} else {
			SystemError(c, "检查IP是否在白名单失败: "+err.Error())
		}
		return
	}

	Success(c, gin.H{
		"is_blocked":     isBlocked,
		"is_whitelisted": isWhitelisted,
	})
}
