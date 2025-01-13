package handler

import (
	"fmt"
	"net"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/service"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// IPRuleHandler IP规则处理器
type IPRuleHandler struct {
	ipService service.IPRuleService
}

// NewIPRuleHandler 创建IP规则处理器
func NewIPRuleHandler(ipService service.IPRuleService) *IPRuleHandler {
	if ipService == nil {
		panic(errors.NewError(errors.ErrConfig, "IP规则服务不能为空"))
	}
	return &IPRuleHandler{
		ipService: ipService,
	}
}

// CreateIPRule 创建IP规则
func (h *IPRuleHandler) CreateIPRule(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("创建IP规则: RequestID=%s", requestID)

	var rule model.IPRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		logger.Errorf("请求参数错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求参数错误: %v", err)))
		return
	}

	// 验证规则
	if err := rule.Validate(); err != nil {
		logger.Errorf("规则验证失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, err)
		return
	}

	if err := h.ipService.CreateIPRule(c.Request.Context(), &rule); err != nil {
		logger.Errorf("创建IP规则失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("创建IP规则失败: %v", err)))
		return
	}

	logger.Infof("创建IP规则成功: RequestID=%s, RuleID=%d", requestID, rule.ID)
	Success(c, rule)
}

// UpdateIPRule 更新IP规则
func (h *IPRuleHandler) UpdateIPRule(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("更新IP规则: RequestID=%s", requestID)

	// 获取规则ID
	id := c.Param("id")
	if id == "" {
		logger.Errorf("缺少规则ID: RequestID=%s", requestID)
		Error(c, errors.NewError(errors.ErrInvalidParams, "规则ID不能为空"))
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		logger.Errorf("无效的规则ID: RequestID=%s, ID=%s", requestID, id)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的规则ID: %s", id)))
		return
	}

	var rule model.IPRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		logger.Errorf("请求参数错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求参数错误: %v", err)))
		return
	}

	// 设置规则ID
	rule.ID = idInt

	// 验证规则
	if err := rule.Validate(); err != nil {
		logger.Errorf("规则验证失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, err)
		return
	}

	if err := h.ipService.UpdateIPRule(c.Request.Context(), &rule); err != nil {
		logger.Errorf("更新IP规则失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新IP规则失败: %v", err)))
		return
	}

	logger.Infof("更新IP规则成功: RequestID=%s, RuleID=%d", requestID, rule.ID)
	Success(c, rule)
}

// DeleteIPRule 删除IP规则
func (h *IPRuleHandler) DeleteIPRule(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("删除IP规则: RequestID=%s", requestID)

	id := c.Param("id")
	if id == "" {
		logger.Errorf("缺少规则ID: RequestID=%s", requestID)
		Error(c, errors.NewError(errors.ErrInvalidParams, "规则ID不能为空"))
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		logger.Errorf("无效的规则ID: RequestID=%s, ID=%s", requestID, id)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的规则ID: %s", id)))
		return
	}

	if err := h.ipService.DeleteIPRule(c.Request.Context(), idInt); err != nil {
		logger.Errorf("删除IP规则失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("删除IP规则失败: %v", err)))
		return
	}

	logger.Infof("删除IP规则成功: RequestID=%s, RuleID=%d", requestID, idInt)
	Success(c, nil)
}

// GetIPRule 获取IP规则
func (h *IPRuleHandler) GetIPRule(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("获取IP规则: RequestID=%s", requestID)

	id := c.Param("id")
	if id == "" {
		logger.Errorf("缺少规则ID: RequestID=%s", requestID)
		Error(c, errors.NewError(errors.ErrInvalidParams, "规则ID不能为空"))
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		logger.Errorf("无效的规则ID: RequestID=%s, ID=%s", requestID, id)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的规则ID: %s", id)))
		return
	}

	rule, err := h.ipService.GetIPRule(c.Request.Context(), idInt)
	if err != nil {
		logger.Errorf("获取IP规则失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取IP规则失败: %v", err)))
		return
	}

	if rule == nil {
		logger.Errorf("IP规则不存在: RequestID=%s, RuleID=%d", requestID, idInt)
		Error(c, errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("IP规则不存在: %d", idInt)))
		return
	}

	logger.Infof("获取IP规则成功: RequestID=%s, RuleID=%d", requestID, idInt)
	Success(c, rule)
}

// ListIPRules 获取IP规则列表
func (h *IPRuleHandler) ListIPRules(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("获取IP规则列表: RequestID=%s", requestID)

	var query model.IPRuleQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.Errorf("请求参数错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求参数错误: %v", err)))
		return
	}

	// 验证分页参数
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		logger.Errorf("无效的页码: RequestID=%s, Page=%s", requestID, c.Query("page"))
		Error(c, errors.NewError(errors.ErrInvalidParams, "页码必须大于0"))
		return
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || size < 1 || size > 100 {
		logger.Errorf("无效的页大小: RequestID=%s, Size=%s", requestID, c.Query("size"))
		Error(c, errors.NewError(errors.ErrInvalidParams, "页大小必须在1-100之间"))
		return
	}

	// 验证IP类型
	if query.IPType != "" {
		switch query.IPType {
		case model.IPListTypeWhite, model.IPListTypeBlack:
			// 合法的IP类型
		default:
			logger.Errorf("无效的IP类型: RequestID=%s, IPType=%s", requestID, query.IPType)
			Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的IP类型: %s", query.IPType)))
			return
		}
	}

	// 验证封禁类型
	if query.BlockType != "" {
		switch query.BlockType {
		case model.BlockTypePermanent, model.BlockTypeTemporary:
			// 合法的封禁类型
		default:
			logger.Errorf("无效的封禁类型: RequestID=%s, BlockType=%s", requestID, query.BlockType)
			Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的封禁类型: %s", query.BlockType)))
			return
		}
	}

	rules, total, err := h.ipService.ListIPRules(c.Request.Context(), query, page, size)
	if err != nil {
		logger.Errorf("获取IP规则列表失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取IP规则列表失败: %v", err)))
		return
	}

	logger.Infof("获取IP规则列表成功: RequestID=%s, Total=%d", requestID, total)
	Success(c, gin.H{
		"total": total,
		"items": rules,
	})
}

// CheckIP 检查IP规则
func (h *IPRuleHandler) CheckIP(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("检查IP规则: RequestID=%s", requestID)

	var req struct {
		IP string `json:"ip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("请求参数错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求参数错误: %v", err)))
		return
	}

	// 验证IP地址格式
	if net.ParseIP(req.IP) == nil {
		logger.Errorf("无效的IP地址: RequestID=%s, IP=%s", requestID, req.IP)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的IP地址: %s", req.IP)))
		return
	}

	isBlocked, err := h.ipService.CheckIP(c.Request.Context(), req.IP)
	if err != nil {
		logger.Errorf("检查IP规则失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("检查IP规则失败: %v", err)))
		return
	}

	logger.Infof("检查IP规则成功: RequestID=%s, IP=%s, IsBlocked=%v", requestID, req.IP, isBlocked)
	Success(c, gin.H{
		"is_blocked": isBlocked,
	})
}
