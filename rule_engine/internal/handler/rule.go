package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
	"github.com/xwaf/rule_engine/internal/service"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// RuleHandler 规则处理器
type RuleHandler struct {
	ruleService    service.RuleService
	versionService service.RuleVersionService
}

// NewRuleHandler 创建规则处理器
func NewRuleHandler(ruleService service.RuleService, versionService service.RuleVersionService) *RuleHandler {
	if ruleService == nil {
		panic(errors.NewError(errors.ErrConfig, "规则服务不能为空"))
	}
	if versionService == nil {
		panic(errors.NewError(errors.ErrConfig, "版本服务不能为空"))
	}
	return &RuleHandler{
		ruleService:    ruleService,
		versionService: versionService,
	}
}

// ListRules 获取规则列表
func (h *RuleHandler) ListRules(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("获取规则列表: RequestID=%s", requestID)

	// 获取分页参数
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "10")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum <= 0 {
		logger.Errorf("无效的页码: RequestID=%s, Page=%s", requestID, page)
		Error(c, errors.NewError(errors.ErrInvalidParams, "页码必须大于0"))
		return
	}

	pageSize, err := strconv.Atoi(size)
	if err != nil || pageSize <= 0 || pageSize > 100 {
		logger.Errorf("无效的页大小: RequestID=%s, Size=%s", requestID, size)
		Error(c, errors.NewError(errors.ErrInvalidParams, "页大小必须在1-100之间"))
		return
	}

	// 构建查询参数
	query := &repository.RuleQuery{
		Page:           pageNum,
		PageSize:       pageSize,
		Keyword:        c.Query("keyword"),
		Status:         model.StatusType(c.Query("status")),
		RuleType:       model.RuleType(c.Query("rule_type")),
		RuleVariable:   model.RuleVariable(c.Query("rule_variable")),
		Severity:       model.SeverityType(c.Query("severity")),
		RulesOperation: c.Query("rules_operation"),
		GroupID:        parseInt64(c.Query("group_id")),
		CreatedBy:      parseInt64(c.Query("created_by")),
		UpdatedBy:      parseInt64(c.Query("updated_by")),
		OrderBy:        c.Query("order_by"),
		OrderDesc:      c.Query("order_desc") == "true",
	}

	// 验证查询参数
	if query.Status != "" {
		switch query.Status {
		case model.StatusEnabled, model.StatusDisabled:
			// 合法的状态值
		default:
			logger.Errorf("无效的状态值: RequestID=%s, Status=%s", requestID, query.Status)
			Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的状态值: %s", query.Status)))
			return
		}
	}
	if query.RuleType != "" {
		switch query.RuleType {
		case model.RuleTypeIP, model.RuleTypeCC, model.RuleTypeRegex, model.RuleTypeSQLi, model.RuleTypeXSS, model.RuleTypeCustom:
			// 合法的规则类型
		default:
			logger.Errorf("无效的规则类型: RequestID=%s, RuleType=%s", requestID, query.RuleType)
			Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的规则类型: %s", query.RuleType)))
			return
		}
	}
	if query.Severity != "" {
		switch query.Severity {
		case model.SeverityHigh, model.SeverityMedium, model.SeverityLow:
			// 合法的严重程度
		default:
			logger.Errorf("无效的严重程度: RequestID=%s, Severity=%s", requestID, query.Severity)
			Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的严重程度: %s", query.Severity)))
			return
		}
	}

	// 解析时间范围
	if startTime := c.Query("start_time"); startTime != "" {
		t, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			logger.Errorf("无效的开始时间: RequestID=%s, StartTime=%s, Error=%v", requestID, startTime, err)
			Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的开始时间格式: %s", startTime)))
			return
		}
		query.StartTime = &t
	}
	if endTime := c.Query("end_time"); endTime != "" {
		t, err := time.Parse(time.RFC3339, endTime)
		if err != nil {
			logger.Errorf("无效的结束时间: RequestID=%s, EndTime=%s, Error=%v", requestID, endTime, err)
			Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的结束时间格式: %s", endTime)))
			return
		}
		query.EndTime = &t
	}

	// 验证时间范围
	if query.StartTime != nil && query.EndTime != nil && query.StartTime.After(*query.EndTime) {
		logger.Errorf("无效的时间范围: RequestID=%s, StartTime=%v, EndTime=%v",
			requestID, query.StartTime, query.EndTime)
		Error(c, errors.NewError(errors.ErrInvalidParams, "开始时间不能晚于结束时间"))
		return
	}

	rules, total, err := h.ruleService.ListRules(c.Request.Context(), query)
	if err != nil {
		logger.Errorf("获取规则列表失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取规则列表失败: %v", err)))
		return
	}

	logger.Infof("获取规则列表成功: RequestID=%s, Total=%d", requestID, total)
	Success(c, gin.H{
		"rules": rules,
		"total": total,
		"page":  pageNum,
		"size":  pageSize,
	})
}

// CreateRule 创建规则
func (h *RuleHandler) CreateRule(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("创建规则: RequestID=%s", requestID)

	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		logger.Errorf("请求参数错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求参数错误: %v", err)))
		return
	}

	// 规则验证
	if err := rule.Validate(); err != nil {
		logger.Errorf("规则验证失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, err)
		return
	}

	// 设置创建者
	if userID := getUserID(c); userID > 0 {
		rule.CreatedBy = userID
		rule.UpdatedBy = userID
	} else {
		logger.Errorf("获取用户ID失败: RequestID=%s", requestID)
		Error(c, errors.NewError(errors.ErrInvalidParams, "无法获取用户ID"))
		return
	}

	// 设置创建时间
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	if err := h.ruleService.CreateRule(c.Request.Context(), &rule); err != nil {
		logger.Errorf("创建规则失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("创建规则失败: %v", err)))
		return
	}

	// 创建审计日志
	auditLog := &model.RuleAuditLog{
		RuleID:    rule.ID,
		Action:    "create",
		Operator:  strconv.FormatInt(rule.CreatedBy, 10),
		NewValue:  toString(rule),
		CreatedAt: now,
	}
	if err := h.createAuditLog(c.Request.Context(), auditLog); err != nil {
		// 仅记录错误，不影响主流程
		logger.Errorf("创建审计日志失败: RequestID=%s, Error=%v", requestID, err)
	}

	logger.Infof("创建规则成功: RequestID=%s, RuleID=%d", requestID, rule.ID)
	Success(c, rule)
}

// UpdateRule 更新规则
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("更新规则: RequestID=%s", requestID)

	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		logger.Errorf("无效的规则ID: RequestID=%s, ID=%s", requestID, id)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的规则ID: %s", id)))
		return
	}

	// 获取原规则
	oldRule, err := h.ruleService.GetRule(c.Request.Context(), ruleID)
	if err != nil {
		logger.Errorf("获取原规则失败: RequestID=%s, RuleID=%d, Error=%v", requestID, ruleID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取原规则失败: %v", err)))
		return
	}
	if oldRule == nil {
		logger.Errorf("规则不存在: RequestID=%s, RuleID=%d", requestID, ruleID)
		Error(c, errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("规则不存在: %d", ruleID)))
		return
	}

	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		logger.Errorf("请求参数错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求参数错误: %v", err)))
		return
	}

	// 规则验证
	if err := rule.Validate(); err != nil {
		logger.Errorf("规则验证失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, err)
		return
	}

	rule.ID = ruleID
	// 设置更新者
	if userID := getUserID(c); userID > 0 {
		rule.UpdatedBy = userID
	} else {
		logger.Errorf("获取用户ID失败: RequestID=%s", requestID)
		Error(c, errors.NewError(errors.ErrInvalidParams, "无法获取用户ID"))
		return
	}

	// 设置更新时间
	rule.UpdatedAt = time.Now()

	if err := h.ruleService.UpdateRule(c.Request.Context(), &rule); err != nil {
		logger.Errorf("更新规则失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("更新规则失败: %v", err)))
		return
	}

	// 创建审计日志
	auditLog := &model.RuleAuditLog{
		RuleID:    rule.ID,
		Action:    "update",
		Operator:  strconv.FormatInt(rule.UpdatedBy, 10),
		OldValue:  toString(oldRule),
		NewValue:  toString(rule),
		CreatedAt: time.Now(),
	}
	if err := h.createAuditLog(c.Request.Context(), auditLog); err != nil {
		// 仅记录错误，不影响主流程
		logger.Errorf("创建审计日志失败: RequestID=%s, Error=%v", requestID, err)
	}

	logger.Infof("更新规则成功: RequestID=%s, RuleID=%d", requestID, rule.ID)
	Success(c, rule)
}

// DeleteRule 删除规则
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("删除规则: RequestID=%s", requestID)

	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		logger.Errorf("无效的规则ID: RequestID=%s, ID=%s", requestID, id)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("无效的规则ID: %s", id)))
		return
	}

	// 获取原规则
	oldRule, err := h.ruleService.GetRule(c.Request.Context(), ruleID)
	if err != nil {
		logger.Errorf("获取原规则失败: RequestID=%s, RuleID=%d, Error=%v", requestID, ruleID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("获取原规则失败: %v", err)))
		return
	}
	if oldRule == nil {
		logger.Errorf("规则不存在: RequestID=%s, RuleID=%d", requestID, ruleID)
		Error(c, errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("规则不存在: %d", ruleID)))
		return
	}

	// 获取当前用户ID
	userID := getUserID(c)
	if userID <= 0 {
		logger.Errorf("获取用户ID失败: RequestID=%s", requestID)
		Error(c, errors.NewError(errors.ErrInvalidParams, "无法获取用户ID"))
		return
	}

	if err := h.ruleService.DeleteRule(c.Request.Context(), ruleID); err != nil {
		logger.Errorf("删除规则失败: RequestID=%s, RuleID=%d, Error=%v", requestID, ruleID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("删除规则失败: %v", err)))
		return
	}

	// 创建审计日志
	auditLog := &model.RuleAuditLog{
		RuleID:    ruleID,
		Action:    "delete",
		Operator:  strconv.FormatInt(userID, 10),
		OldValue:  toString(oldRule),
		CreatedAt: time.Now(),
	}
	if err := h.createAuditLog(c.Request.Context(), auditLog); err != nil {
		// 仅记录错误，不影响主流程
		logger.Errorf("创建审计日志失败: RequestID=%s, Error=%v", requestID, err)
	}

	logger.Infof("删除规则成功: RequestID=%s, RuleID=%d", requestID, ruleID)
	Success(c, nil)
}

// CheckRule 检查规则匹配
func (h *RuleHandler) CheckRule(c *gin.Context) {
	requestID := c.GetString("request_id")
	logger.Infof("检查规则匹配: RequestID=%s", requestID)

	var req model.CheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("请求数据格式错误: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("请求数据格式错误: %v", err)))
		return
	}

	// 验证请求参数
	if err := req.Validate(); err != nil {
		logger.Errorf("请求参数验证失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, err)
		return
	}

	result, err := h.ruleService.CheckRequest(c.Request.Context(), &req)
	if err != nil {
		logger.Errorf("检查规则匹配失败: RequestID=%s, Error=%v", requestID, err)
		Error(c, errors.NewError(errors.ErrRuleEngine, fmt.Sprintf("检查规则匹配失败: %v", err)))
		return
	}

	if result.Matched {
		logger.Infof("规则匹配成功: RequestID=%s, Rule=%s, Action=%s", requestID, result.MatchedRule.Name, result.Action)
	} else {
		logger.Infof("未匹配任何规则: RequestID=%s", requestID)
	}

	Success(c, result)
}

// GetRule 获取单个规则
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

	Success(c, rule)
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

// GetRuleVersion 获取规则版本
func (h *RuleHandler) GetRuleVersion(c *gin.Context) {
	version, err := h.versionService.GetVersion(c.Request.Context(), 0, 0)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, gin.H{"version": version})
}

// GetRuleUpdateEvent 获取规则更新事件
func (h *RuleHandler) GetRuleUpdateEvent(c *gin.Context) {
	events, err := h.versionService.GetSyncLogs(c.Request.Context(), 0)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, gin.H{"events": events})
}

// BatchCreateRules 批量创建规则
func (h *RuleHandler) BatchCreateRules(c *gin.Context) {
	var rules []*model.Rule
	if err := c.ShouldBindJSON(&rules); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	// 规则验证
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("规则[%s]验证失败: %s", rule.Name, err.Error())))
			return
		}
	}

	// 设置创建者
	userID := getUserID(c)
	for _, rule := range rules {
		if userID > 0 {
			rule.CreatedBy = userID
			rule.UpdatedBy = userID
		}
	}

	if err := h.ruleService.BatchCreateRules(c.Request.Context(), rules); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	// 创建审计日志
	for _, rule := range rules {
		h.createAuditLog(c.Request.Context(), &model.RuleAuditLog{
			RuleID:    rule.ID,
			Action:    "batch_create",
			Operator:  strconv.FormatInt(rule.CreatedBy, 10),
			NewValue:  toString(rule),
			CreatedAt: time.Now(),
		})
	}

	Success(c, rules)
}

// BatchUpdateRules 批量更新规则
func (h *RuleHandler) BatchUpdateRules(c *gin.Context) {
	var rules []*model.Rule
	if err := c.ShouldBindJSON(&rules); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	// 规则验证
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("规则[%s]验证失败: %s", rule.Name, err.Error())))
			return
		}
	}

	// 设置更新者
	userID := getUserID(c)
	for _, rule := range rules {
		if userID > 0 {
			rule.UpdatedBy = userID
		}
	}

	// 获取原规则
	var oldRules = make(map[int64]*model.Rule)
	for _, rule := range rules {
		oldRule, err := h.ruleService.GetRule(c.Request.Context(), rule.ID)
		if err != nil {
			Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
			return
		}
		oldRules[rule.ID] = oldRule
	}

	if err := h.ruleService.BatchUpdateRules(c.Request.Context(), rules); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	// 创建审计日志
	for _, rule := range rules {
		h.createAuditLog(c.Request.Context(), &model.RuleAuditLog{
			RuleID:    rule.ID,
			Action:    "batch_update",
			Operator:  strconv.FormatInt(rule.UpdatedBy, 10),
			OldValue:  toString(oldRules[rule.ID]),
			NewValue:  toString(rule),
			CreatedAt: time.Now(),
		})
	}

	Success(c, rules)
}

// BatchDeleteRules 批量删除规则
func (h *RuleHandler) BatchDeleteRules(c *gin.Context) {
	var ruleIDs []int64
	if err := c.ShouldBindJSON(&ruleIDs); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	// 获取原规则
	var oldRules = make(map[int64]*model.Rule)
	for _, ruleID := range ruleIDs {
		oldRule, err := h.ruleService.GetRule(c.Request.Context(), ruleID)
		if err != nil {
			Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
			return
		}
		oldRules[ruleID] = oldRule
	}

	if err := h.ruleService.BatchDeleteRules(c.Request.Context(), ruleIDs); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	// 创建审计日志
	userID := getUserID(c)
	for _, ruleID := range ruleIDs {
		h.createAuditLog(c.Request.Context(), &model.RuleAuditLog{
			RuleID:    ruleID,
			Action:    "batch_delete",
			Operator:  strconv.FormatInt(userID, 10),
			OldValue:  toString(oldRules[ruleID]),
			CreatedAt: time.Now(),
		})
	}

	Success(c, nil)
}

// ImportRules 导入规则
func (h *RuleHandler) ImportRules(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "请选择要导入的文件"))
		return
	}

	// 读取文件内容
	f, err := file.Open()
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}
	defer f.Close()

	var rules []*model.Rule
	if err := json.NewDecoder(f).Decode(&rules); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "文件格式错误"))
		return
	}

	// 规则验证
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			Error(c, errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("规则[%s]验证失败: %s", rule.Name, err.Error())))
			return
		}
	}

	// 设置创建者
	userID := getUserID(c)
	for _, rule := range rules {
		if userID > 0 {
			rule.CreatedBy = userID
			rule.UpdatedBy = userID
		}
	}

	if err := h.ruleService.ImportRules(c.Request.Context(), rules); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	// 创建审计日志
	for _, rule := range rules {
		h.createAuditLog(c.Request.Context(), &model.RuleAuditLog{
			RuleID:    rule.ID,
			Action:    "import",
			Operator:  strconv.FormatInt(rule.CreatedBy, 10),
			NewValue:  toString(rule),
			CreatedAt: time.Now(),
		})
	}

	Success(c, gin.H{
		"total":   len(rules),
		"success": len(rules),
	})
}

// ExportRules 导出规则
func (h *RuleHandler) ExportRules(c *gin.Context) {
	query := &repository.RuleQuery{
		Keyword:        c.Query("keyword"),
		Status:         model.StatusType(c.Query("status")),
		RuleType:       model.RuleType(c.Query("rule_type")),
		RuleVariable:   model.RuleVariable(c.Query("rule_variable")),
		Severity:       model.SeverityType(c.Query("severity")),
		RulesOperation: c.Query("rules_operation"),
	}

	rules, err := h.ruleService.ExportRules(c.Request.Context(), query)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=rules.json")

	// 导出为JSON文件
	if err := json.NewEncoder(c.Writer).Encode(rules); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}
}

// GetRuleStats 获取规则统计信息
func (h *RuleHandler) GetRuleStats(c *gin.Context) {
	stats, err := h.ruleService.GetRuleStats(c.Request.Context())
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, stats)
}

// GetRuleMatchStats 获取规则匹配统计
func (h *RuleHandler) GetRuleMatchStats(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	// 解析时间范围
	startTime, _ := time.Parse(time.RFC3339, c.Query("start_time"))
	endTime, _ := time.Parse(time.RFC3339, c.Query("end_time"))
	if endTime.IsZero() {
		endTime = time.Now()
	}
	if startTime.IsZero() {
		startTime = endTime.AddDate(0, 0, -7) // 默认最近7天
	}

	stats, err := h.ruleService.GetRuleMatchStats(c.Request.Context(), ruleID, startTime, endTime)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, stats)
}

// GetRuleAuditLogs 获取规则审计日志
func (h *RuleHandler) GetRuleAuditLogs(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, "无效的规则ID"))
		return
	}

	logs, err := h.ruleService.GetRuleAuditLogs(c.Request.Context(), ruleID)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, logs)
}

// 辅助函数

// parseInt64 解析int64
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// getUserID 获取当前用户ID
func getUserID(c *gin.Context) int64 {
	if v, exists := c.Get("user_id"); exists {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

// toString 将对象转换为字符串
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// createAuditLog 创建审计日志
func (h *RuleHandler) createAuditLog(ctx context.Context, log *model.RuleAuditLog) error {
	if err := h.ruleService.CreateRuleAuditLog(ctx, log); err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("创建审计日志失败: %v", err))
	}
	return nil
}
