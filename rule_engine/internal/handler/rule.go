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
)

// RuleHandler 规则处理器
type RuleHandler struct {
	ruleService    service.RuleService
	versionService service.RuleVersionService
}

// NewRuleHandler 创建规则处理器
func NewRuleHandler(ruleService service.RuleService, versionService service.RuleVersionService) *RuleHandler {
	return &RuleHandler{
		ruleService:    ruleService,
		versionService: versionService,
	}
}

// ListRules 获取规则列表
func (h *RuleHandler) ListRules(c *gin.Context) {
	// 获取分页参数
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "10")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum <= 0 {
		Error(c, errors.NewError(errors.ErrInvalidParams, "页码必须大于0"))
		return
	}

	pageSize, err := strconv.Atoi(size)
	if err != nil || pageSize <= 0 || pageSize > 100 {
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

	// 解析时间范围
	if startTime := c.Query("start_time"); startTime != "" {
		t, err := time.Parse(time.RFC3339, startTime)
		if err == nil {
			query.StartTime = &t
		}
	}
	if endTime := c.Query("end_time"); endTime != "" {
		t, err := time.Parse(time.RFC3339, endTime)
		if err == nil {
			query.EndTime = &t
		}
	}

	rules, total, err := h.ruleService.ListRules(c.Request.Context(), query)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	Success(c, gin.H{
		"rules": rules,
		"total": total,
		"page":  pageNum,
		"size":  pageSize,
	})
}

// CreateRule 创建规则
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	// 规则验证
	if err := rule.Validate(); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	// 设置创建者
	if userID := getUserID(c); userID > 0 {
		rule.CreatedBy = userID
		rule.UpdatedBy = userID
	}

	if err := h.ruleService.CreateRule(c.Request.Context(), &rule); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	// 创建审计日志
	h.createAuditLog(c.Request.Context(), &model.RuleAuditLog{
		RuleID:    rule.ID,
		Action:    "create",
		Operator:  strconv.FormatInt(rule.CreatedBy, 10),
		NewValue:  toString(rule),
		CreatedAt: time.Now(),
	})

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

	// 获取原规则
	oldRule, err := h.ruleService.GetRule(c.Request.Context(), ruleID)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	var rule model.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	// 规则验证
	if err := rule.Validate(); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	rule.ID = ruleID
	// 设置更新者
	if userID := getUserID(c); userID > 0 {
		rule.UpdatedBy = userID
	}

	if err := h.ruleService.UpdateRule(c.Request.Context(), &rule); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	// 创建审计日志
	h.createAuditLog(c.Request.Context(), &model.RuleAuditLog{
		RuleID:    rule.ID,
		Action:    "update",
		Operator:  strconv.FormatInt(rule.UpdatedBy, 10),
		OldValue:  toString(oldRule),
		NewValue:  toString(rule),
		CreatedAt: time.Now(),
	})

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

	// 获取原规则
	oldRule, err := h.ruleService.GetRule(c.Request.Context(), ruleID)
	if err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	if err := h.ruleService.DeleteRule(c.Request.Context(), ruleID); err != nil {
		Error(c, errors.NewError(errors.ErrRuleEngine, err.Error()))
		return
	}

	// 创建审计日志
	h.createAuditLog(c.Request.Context(), &model.RuleAuditLog{
		RuleID:    ruleID,
		Action:    "delete",
		Operator:  strconv.FormatInt(getUserID(c), 10),
		OldValue:  toString(oldRule),
		CreatedAt: time.Now(),
	})

	Success(c, nil)
}

// CheckRule 检查规则匹配
func (h *RuleHandler) CheckRule(c *gin.Context) {
	var req model.CheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, errors.NewError(errors.ErrInvalidParams, err.Error()))
		return
	}

	// 验证请求参数
	if err := req.Validate(); err != nil {
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
func (h *RuleHandler) createAuditLog(ctx context.Context, log *model.RuleAuditLog) {
	if err := h.ruleService.CreateRuleAuditLog(ctx, log); err != nil {
		// 只记录错误日志，不影响主流程
		fmt.Printf("创建审计日志失败: %v\n", err)
	}
}
