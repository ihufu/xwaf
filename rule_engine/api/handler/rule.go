package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	xerrors "github.com/xwaf/rule_engine/internal/errors"
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
	if ruleService == nil {
		panic("规则服务不能为空")
	}
	return &RuleHandler{
		ruleService: ruleService,
	}
}

// CreateRule 创建规则
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		ValidationError(c, "规则数据格式错误: "+err.Error())
		return
	}

	// 验证规则数据
	if rule.Name == "" {
		ValidationError(c, "规则名称不能为空")
		return
	}
	if rule.Pattern == "" {
		ValidationError(c, "规则匹配模式不能为空")
		return
	}
	if rule.Action == "" {
		ValidationError(c, "规则动作不能为空")
		return
	}

	if err := h.ruleService.CreateRule(c.Request.Context(), &rule); err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrRuleValidation:
				ValidationError(c, err.Error())
			case xerrors.ErrRuleConflict:
				ConflictError(c, err.Error())
			default:
				SystemError(c, "创建规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "创建规则失败: "+err.Error())
		}
		return
	}

	Success(c, rule)
}

// UpdateRule 更新规则
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ValidationError(c, "无效的规则ID")
		return
	}
	if ruleID <= 0 {
		ValidationError(c, "规则ID必须大于0")
		return
	}

	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		ValidationError(c, "规则数据格式错误: "+err.Error())
		return
	}

	// 验证规则数据
	if rule.Name == "" {
		ValidationError(c, "规则名称不能为空")
		return
	}
	if rule.Pattern == "" {
		ValidationError(c, "规则匹配模式不能为空")
		return
	}
	if rule.Action == "" {
		ValidationError(c, "规则动作不能为空")
		return
	}

	rule.ID = ruleID
	if err := h.ruleService.UpdateRule(c.Request.Context(), &rule); err != nil {
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
				SystemError(c, "更新规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "更新规则失败: "+err.Error())
		}
		return
	}

	Success(c, rule)
}

// DeleteRule 删除规则
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ValidationError(c, "无效的规则ID")
		return
	}
	if ruleID <= 0 {
		ValidationError(c, "规则ID必须大于0")
		return
	}

	if err := h.ruleService.DeleteRule(c.Request.Context(), ruleID); err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrRuleNotFound:
				NotFoundError(c, err.Error())
			default:
				SystemError(c, "删除规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "删除规则失败: "+err.Error())
		}
		return
	}

	Success(c, nil)
}

// GetRule 获取规则
func (h *RuleHandler) GetRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ValidationError(c, "无效的规则ID")
		return
	}
	if ruleID <= 0 {
		ValidationError(c, "规则ID必须大于0")
		return
	}

	rule, err := h.ruleService.GetRule(c.Request.Context(), ruleID)
	if err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrRuleNotFound:
				NotFoundError(c, err.Error())
			default:
				SystemError(c, "获取规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "获取规则失败: "+err.Error())
		}
		return
	}

	if rule == nil {
		NotFoundError(c, "规则不存在")
		return
	}

	Success(c, rule)
}

// ListRules 获取规则列表
func (h *RuleHandler) ListRules(c *gin.Context) {
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

	query := &repository.RuleQuery{
		Page:     page,
		PageSize: size,
	}

	rules, total, err := h.ruleService.ListRules(c.Request.Context(), query)
	if err != nil {
		SystemError(c, "查询规则列表失败: "+err.Error())
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
		ValidationError(c, "请求数据格式错误: "+err.Error())
		return
	}

	// 验证请求数据
	if req.URI == "" {
		ValidationError(c, "请求URL不能为空")
		return
	}
	if req.Method == "" {
		ValidationError(c, "请求方法不能为空")
		return
	}

	result, err := h.ruleService.CheckRequest(c.Request.Context(), &req)
	if err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrRuleMatch:
				ValidationError(c, err.Error())
			case xerrors.ErrRuleEngine:
				SystemError(c, "规则引擎错误: "+err.Error())
			default:
				SystemError(c, "检查规则匹配失败: "+err.Error())
			}
		} else {
			SystemError(c, "检查规则匹配失败: "+err.Error())
		}
		return
	}

	Success(c, result)
}

// ReloadRules 重新加载规则
func (h *RuleHandler) ReloadRules(c *gin.Context) {
	err := h.ruleService.ReloadRules(c.Request.Context())
	if err != nil {
		var e *xerrors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case xerrors.ErrCache:
				ServiceUnavailableError(c, "缓存服务不可用: "+err.Error())
			case xerrors.ErrRuleEngine:
				SystemError(c, "规则引擎错误: "+err.Error())
			default:
				SystemError(c, "重新加载规则失败: "+err.Error())
			}
		} else {
			SystemError(c, "重新加载规则失败: "+err.Error())
		}
		return
	}

	Success(c, nil)
}
