package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	xerrors "github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/service"
)

// CCRuleHandler CC规则处理器
type CCRuleHandler struct {
	ccRuleService service.CCRuleService
}

// NewCCRuleHandler 创建CC规则处理器
func NewCCRuleHandler(ccRuleService service.CCRuleService) *CCRuleHandler {
	if ccRuleService == nil {
		panic("CC规则服务不能为空")
	}
	return &CCRuleHandler{
		ccRuleService: ccRuleService,
	}
}

// CreateCCRule 创建CC规则
func (h *CCRuleHandler) CreateCCRule(c *gin.Context) {
	var rule model.CCRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		ValidationError(c, "CC规则数据格式错误: "+err.Error())
		return
	}

	// 验证规则数据
	if err := rule.Validate(); err != nil {
		ValidationError(c, err.Error())
		return
	}

	if err := h.ccRuleService.CreateCCRule(c.Request.Context(), &rule); err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrRuleValidation:
				ValidationError(c, err.Error())
			case xerrors.ErrRuleConflict:
				ConflictError(c, err.Error())
			default:
				SystemError(c, "创建CC规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "创建CC规则失败: "+err.Error())
		}
		return
	}

	Success(c, rule)
}

// UpdateCCRule 更新CC规则
func (h *CCRuleHandler) UpdateCCRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ValidationError(c, "无效的规则ID")
		return
	}

	var rule model.CCRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		ValidationError(c, "CC规则数据格式错误: "+err.Error())
		return
	}

	// 验证规则数据
	if err := rule.Validate(); err != nil {
		ValidationError(c, err.Error())
		return
	}

	rule.ID = ruleID
	if err := h.ccRuleService.UpdateCCRule(c.Request.Context(), &rule); err != nil {
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
				SystemError(c, "更新CC规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "更新CC规则失败: "+err.Error())
		}
		return
	}

	Success(c, rule)
}

// DeleteCCRule 删除CC规则
func (h *CCRuleHandler) DeleteCCRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ValidationError(c, "无效的规则ID")
		return
	}

	if err := h.ccRuleService.DeleteCCRule(c.Request.Context(), ruleID); err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrRuleNotFound:
				NotFoundError(c, err.Error())
			default:
				SystemError(c, "删除CC规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "删除CC规则失败: "+err.Error())
		}
		return
	}

	Success(c, nil)
}

// GetCCRule 获取CC规则
func (h *CCRuleHandler) GetCCRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ValidationError(c, "无效的规则ID")
		return
	}

	rule, err := h.ccRuleService.GetCCRule(c.Request.Context(), ruleID)
	if err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrRuleNotFound:
				NotFoundError(c, err.Error())
			default:
				SystemError(c, "获取CC规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "获取CC规则失败: "+err.Error())
		}
		return
	}

	if rule == nil {
		NotFoundError(c, "CC规则不存在")
		return
	}

	Success(c, rule)
}

// ListCCRules 获取CC规则列表
func (h *CCRuleHandler) ListCCRules(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		ValidationError(c, "无效的页码")
		return
	}
	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		ValidationError(c, "无效的页大小")
		return
	}

	if page < 1 {
		ValidationError(c, "页码必须大于0")
		return
	}
	if size < 1 || size > 100 {
		ValidationError(c, "页大小必须在1-100之间")
		return
	}

	rules, total, err := h.ccRuleService.ListCCRules(c.Request.Context(), model.CCRuleQuery{}, page, size)
	if err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrSystem:
				SystemError(c, "获取CC规则列表失败: "+err.Error())
			default:
				SystemError(c, "获取CC规则列表失败: "+err.Error())
			}
		} else {
			SystemError(c, "获取CC规则列表失败: "+err.Error())
		}
		return
	}

	Success(c, gin.H{
		"total": total,
		"items": rules,
	})
}

// CheckCC 检查CC规则匹配
func (h *CCRuleHandler) CheckCC(c *gin.Context) {
	var req struct {
		IP     string `json:"ip" binding:"required"`
		Path   string `json:"path" binding:"required"`
		Method string `json:"method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "请求数据格式错误: "+err.Error())
		return
	}

	// 验证IP地址格式
	if req.IP == "" {
		ValidationError(c, "IP地址不能为空")
		return
	}

	// 验证请求路径
	if req.Path == "" {
		ValidationError(c, "请求路径不能为空")
		return
	}

	// 验证请求方法
	if req.Method == "" {
		ValidationError(c, "请求方法不能为空")
		return
	}

	isBlocked, err := h.ccRuleService.CheckCC(c.Request.Context(), req.IP, req.Path, req.Method)
	if err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrCache:
				ServiceUnavailableError(c, "CC检查服务暂时不可用: "+err.Error())
			case xerrors.ErrRuleEngine:
				SystemError(c, "CC规则检查失败: "+err.Error())
			default:
				SystemError(c, "CC规则检查失败: "+err.Error())
			}
		} else {
			SystemError(c, "CC规则检查失败: "+err.Error())
		}
		return
	}

	Success(c, gin.H{
		"is_blocked": isBlocked,
	})
}
