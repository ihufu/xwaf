package handler

import (
	"fmt"
	"net"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/service"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// CCRuleHandler CC规则处理器
type CCRuleHandler struct {
	ccService service.CCRuleService
}

// NewCCRuleHandler 创建CC规则处理器
func NewCCRuleHandler(ccService service.CCRuleService) *CCRuleHandler {
	if ccService == nil {
		panic(errors.NewError(errors.ErrConfig, "CC规则服务不能为空"))
	}
	return &CCRuleHandler{
		ccService: ccService,
	}
}

// validateCCRule 验证CC规则
func (h *CCRuleHandler) validateCCRule(rule *model.CCRule) error {
	if rule.URI == "" {
		return errors.NewError(errors.ErrInvalidParams, "URI不能为空")
	}

	// 验证URI格式
	if !regexp.MustCompile(`^/[\w\-./]*$`).MatchString(rule.URI) {
		return errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的URI格式: %s", rule.URI))
	}

	// 验证限制速率
	if rule.LimitRate <= 0 {
		return errors.NewError(errors.ErrInvalidParams, "限制速率必须大于0")
	}

	// 验证时间窗口
	if rule.TimeWindow <= 0 {
		return errors.NewError(errors.ErrInvalidParams, "时间窗口必须大于0")
	}

	// 验证限制单位
	switch rule.LimitUnit {
	case model.LimitUnitSecond, model.LimitUnitMinute, model.LimitUnitHour, model.LimitUnitDay:
		// 合法的限制单位
	default:
		return errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的限制单位: %s", rule.LimitUnit))
	}

	// 验证状态
	switch rule.Status {
	case model.CCStatusEnabled, model.CCStatusDisabled:
		// 合法的状态
	default:
		return errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的状态: %s", rule.Status))
	}

	return nil
}

// CreateCCRule 创建CC规则
func (h *CCRuleHandler) CreateCCRule(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("创建CC规则: RequestID=%s", requestID)

	var rule model.CCRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		logger.Errorf("请求参数错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求参数错误: %v", err)))
		return
	}

	// 验证规则
	if err := h.validateCCRule(&rule); err != nil {
		logger.Errorf("规则验证失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, err)
		return
	}

	if err := h.ccService.CreateCCRule(c.Request.Context(), &rule); err != nil {
		logger.Errorf("创建CC规则失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("创建CC规则失败: %v", err)))
		return
	}

	logger.Infof("创建CC规则成功: RequestID=%s, RuleID=%d", requestID, rule.ID)
	Success(c, rule)
}

// UpdateCCRule 更新CC规则
func (h *CCRuleHandler) UpdateCCRule(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("更新CC规则: RequestID=%s", requestID)

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

	var rule model.CCRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		logger.Errorf("请求参数错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求参数错误: %v", err)))
		return
	}

	rule.ID = idInt

	// 验证规则
	if err := h.validateCCRule(&rule); err != nil {
		logger.Errorf("规则验证失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, err)
		return
	}

	if err := h.ccService.UpdateCCRule(c.Request.Context(), &rule); err != nil {
		logger.Errorf("更新CC规则失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新CC规则失败: %v", err)))
		return
	}

	logger.Infof("更新CC规则成功: RequestID=%s, RuleID=%d", requestID, rule.ID)
	Success(c, rule)
}

// DeleteCCRule 删除CC规则
func (h *CCRuleHandler) DeleteCCRule(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("删除CC规则: RequestID=%s", requestID)

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

	if err := h.ccService.DeleteCCRule(c.Request.Context(), idInt); err != nil {
		logger.Errorf("删除CC规则失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("删除CC规则失败: %v", err)))
		return
	}

	logger.Infof("删除CC规则成功: RequestID=%s, RuleID=%d", requestID, idInt)
	Success(c, nil)
}

// GetCCRule 获取CC规则
func (h *CCRuleHandler) GetCCRule(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("获取CC规则: RequestID=%s", requestID)

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

	rule, err := h.ccService.GetCCRule(c.Request.Context(), idInt)
	if err != nil {
		logger.Errorf("获取CC规则失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取CC规则失败: %v", err)))
		return
	}

	if rule == nil {
		logger.Errorf("CC规则不存在: RequestID=%s, RuleID=%d", requestID, idInt)
		Error(c, errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("CC规则不存在: %d", idInt)))
		return
	}

	logger.Infof("获取CC规则成功: RequestID=%s, RuleID=%d", requestID, idInt)
	Success(c, rule)
}

// ListCCRules 获取CC规则列表
func (h *CCRuleHandler) ListCCRules(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("获取CC规则列表: RequestID=%s", requestID)

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

	// 解析查询参数
	var query model.CCRuleQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.Errorf("请求参数错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求参数错误: %v", err)))
		return
	}

	rules, total, err := h.ccService.ListCCRules(c.Request.Context(), query, page, size)
	if err != nil {
		logger.Errorf("获取CC规则列表失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取CC规则列表失败: %v", err)))
		return
	}

	logger.Infof("获取CC规则列表成功: RequestID=%s, Total=%d", requestID, total)
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
	requestID := c.GetString("request_id")
	logger.Infof("检查CC限制: RequestID=%s", requestID)

	uri := c.Param("uri")
	if uri == "" {
		logger.Errorf("缺少URI参数: RequestID=%s", requestID)
		Error(c, errors.NewError(errors.ErrInvalidParams, "URI不能为空"))
		return
	}

	// 验证URI格式
	if !regexp.MustCompile(`^/[\w\-./]*$`).MatchString(uri) {
		logger.Errorf("无效的URI格式: RequestID=%s, URI=%s", requestID, uri)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的URI格式: %s", uri)))
		return
	}

	isLimited, err := h.ccService.CheckCCLimit(c.Request.Context(), uri)
	if err != nil {
		logger.Errorf("检查CC限制失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("检查CC限制失败: %v", err)))
		return
	}

	logger.Infof("检查CC限制成功: RequestID=%s, URI=%s, IsLimited=%v", requestID, uri, isLimited)
	Success(c, gin.H{
		"is_limited": isLimited,
	})
}

// CheckCC 检查CC规则
func (h *CCRuleHandler) CheckCC(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("检查CC规则: RequestID=%s", requestID)

	var req struct {
		IP     string `json:"ip" binding:"required"`
		Path   string `json:"path" binding:"required"`
		Method string `json:"method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("请求参数错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求参数错误: %v", err)))
		return
	}

	// 验证IP地址
	if net.ParseIP(req.IP) == nil {
		logger.Errorf("无效的IP地址: RequestID=%s, IP=%s", requestID, req.IP)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的IP地址: %s", req.IP)))
		return
	}

	// 验证Path格式
	if !regexp.MustCompile(`^/[\w\-./]*$`).MatchString(req.Path) {
		logger.Errorf("无效的Path格式: RequestID=%s, Path=%s", requestID, req.Path)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的Path格式: %s", req.Path)))
		return
	}

	// 验证Method
	switch req.Method {
	case "GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH":
		// 合法的HTTP方法
	default:
		logger.Errorf("无效的HTTP方法: RequestID=%s, Method=%s", requestID, req.Method)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的HTTP方法: %s", req.Method)))
		return
	}

	isBlocked, err := h.ccService.CheckCC(c.Request.Context(), req.IP, req.Path, req.Method)
	if err != nil {
		logger.Errorf("检查CC规则失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("检查CC规则失败: %v", err)))
		return
	}

	logger.Infof("检查CC规则成功: RequestID=%s, IP=%s, Path=%s, Method=%s, IsBlocked=%v",
		requestID, req.IP, req.Path, req.Method, isBlocked)
	Success(c, gin.H{
		"is_blocked": isBlocked,
	})
}
