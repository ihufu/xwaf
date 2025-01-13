package service

import (
	"context"
	"time"

	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
)

// RuleService 统一的规则服务接口
type RuleService interface {
	// 规则管理
	CreateRule(ctx context.Context, rule *model.Rule) error
	BatchCreateRules(ctx context.Context, rules []*model.Rule) error
	UpdateRule(ctx context.Context, rule *model.Rule) error
	BatchUpdateRules(ctx context.Context, rules []*model.Rule) error
	DeleteRule(ctx context.Context, id int64) error
	BatchDeleteRules(ctx context.Context, ids []int64) error
	GetRule(ctx context.Context, id int64) (*model.Rule, error)
	ListRules(ctx context.Context, query *repository.RuleQuery) ([]*model.Rule, int64, error)

	// 规则检查
	CheckRequest(ctx context.Context, req *model.CheckRequest) (*model.CheckResult, error)

	// 规则同步
	ReloadRules(ctx context.Context) error
	GetVersion(ctx context.Context) (int64, error)

	// 规则导入导出
	ImportRules(ctx context.Context, rules []*model.Rule) error
	ExportRules(ctx context.Context, query *repository.RuleQuery) ([]*model.Rule, error)

	// 规则审计
	GetRuleAuditLogs(ctx context.Context, ruleID int64) ([]*model.RuleAuditLog, error)
	CreateRuleAuditLog(ctx context.Context, log *model.RuleAuditLog) error

	// 规则统计（简化后的接口）
	IncrRuleMatchCount(ctx context.Context, ruleID int64) error
	GetRuleMatchCount(ctx context.Context, ruleID int64) (int64, error)
	GetRuleStats(ctx context.Context) (*model.RuleStats, error)
	GetRuleMatchStats(ctx context.Context, ruleID int64, startTime, endTime time.Time) (*model.RuleMatchStat, error)
}

// RuleFactory 规则工厂接口
type RuleFactory interface {
	CreateRuleHandler(ruleType model.RuleType) (RuleHandler, error)
}

// RuleHandler 规则处理器接口
type RuleHandler interface {
	// 规则匹配
	Match(ctx context.Context, rule *model.Rule, req *model.CheckRequest) (bool, error)
}
