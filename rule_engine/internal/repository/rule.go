package repository

import (
	"time"

	"github.com/xwaf/rule_engine/internal/model"
)

// RuleQuery 规则查询参数
type RuleQuery struct {
	Page           int                `form:"page"`            // 页码
	PageSize       int                `form:"size"`            // 每页大小
	Keyword        string             `form:"keyword"`         // 关键词
	Status         model.StatusType   `form:"status"`          // 状态
	RuleType       model.RuleType     `form:"rule_type"`       // 规则类型
	RuleVariable   model.RuleVariable `form:"rule_variable"`   // 规则变量
	Severity       model.SeverityType `form:"severity"`        // 风险级别
	RulesOperation string             `form:"rules_operation"` // 规则组合操作
	GroupID        int64              `form:"group_id"`        // 规则组ID
	CreatedBy      int64              `form:"created_by"`      // 创建者ID
	UpdatedBy      int64              `form:"updated_by"`      // 更新者ID
	StartTime      *time.Time         `form:"start_time"`      // 开始时间
	EndTime        *time.Time         `form:"end_time"`        // 结束时间
	OrderBy        string             `form:"order_by"`        // 排序字段
	OrderDesc      bool               `form:"order_desc"`      // 是否降序
}

// RuleStats 规则统计信息
type RuleStats struct {
	TotalRules    int64     `json:"total_rules"`     // 总规则数
	EnabledRules  int64     `json:"enabled_rules"`   // 启用规则数
	LastUpdatedAt time.Time `json:"last_updated_at"` // 最后更新时间
}
