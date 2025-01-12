package repository

import (
	"context"
	"time"

	"github.com/xwaf/rule_engine/internal/model"
)

// RuleStatsRepository 规则统计仓库接口
type RuleStatsRepository interface {
	GetRuleStats(ctx context.Context) (*model.RuleStats, error)
	GetRuleMatchStats(ctx context.Context, ruleID int64, startTime, endTime time.Time) (*model.RuleMatchStat, error)
}

// RuleAuditRepository 规则审计仓库接口
type RuleAuditRepository interface {
	CreateAuditLog(ctx context.Context, log *model.RuleAuditLog) error
	GetAuditLogs(ctx context.Context, ruleID int64) ([]*model.RuleAuditLog, error)
}

// RuleCache 规则缓存接口
type RuleCache interface {
	GetRule(ctx context.Context, id int64) (*model.Rule, error)
	SetRule(ctx context.Context, rule *model.Rule) error
	DeleteRule(ctx context.Context, id int64) error
	ClearRules(ctx context.Context) error
}

// CacheRepository 缓存仓库接口
type CacheRepository interface {
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// RuleRepository 规则仓库接口
type RuleRepository interface {
	// 基础操作
	CreateRule(ctx context.Context, rule *model.Rule) error
	BatchCreateRules(ctx context.Context, rules []*model.Rule) error
	UpdateRule(ctx context.Context, rule *model.Rule) error
	BatchUpdateRules(ctx context.Context, rules []*model.Rule) error
	DeleteRule(ctx context.Context, id int64) error
	BatchDeleteRules(ctx context.Context, ids []int64) error
	GetRule(ctx context.Context, id int64) (*model.Rule, error)
	GetRuleByName(ctx context.Context, name string) (*model.Rule, error)
	ListRules(ctx context.Context, query *RuleQuery) ([]*model.Rule, int64, error)
	GetLatestVersion(ctx context.Context) (int64, error)
	ImportRules(ctx context.Context, rules []*model.Rule) error

	// 事务相关
	BeginTx(ctx context.Context) (Transaction, error)

	// 规则统计
	GetRuleStats(ctx context.Context) (*model.RuleStats, error)
	GetRuleMatchStats(ctx context.Context, ruleID int64, startTime, endTime time.Time) (*model.RuleMatchStat, error)
	IncrRuleMatchCount(ctx context.Context, ruleID int64) error
	GetRuleMatchCount(ctx context.Context, ruleID int64) (int64, error)

	// 规则审计
	GetRuleAuditLogs(ctx context.Context, ruleID int64) ([]*model.RuleAuditLog, error)
	CreateRuleAuditLog(ctx context.Context, log *model.RuleAuditLog) error
}

// Transaction 事务接口
type Transaction interface {
	Commit() error
	Rollback() error
}

// RuleVersionRepository 规则版本仓库接口
type RuleVersionRepository interface {
	CreateVersion(ctx context.Context, version *model.RuleVersion) error
	GetVersion(ctx context.Context, ruleID, version int64) (*model.RuleVersion, error)
	ListVersions(ctx context.Context, ruleID int64) ([]*model.RuleVersion, error)
	GetRulesByVersion(ctx context.Context, version int64) ([]*model.Rule, error)
	GetLatestVersion(ctx context.Context) (int64, error)
	RollbackRules(ctx context.Context, rules []*model.Rule, event *model.RuleUpdateEvent) error
	RefreshRules(ctx context.Context) error
	CreateSyncLog(ctx context.Context, log *model.RuleSyncLog) error
	ListSyncLogs(ctx context.Context, ruleID int64) ([]*model.RuleSyncLog, error)
}

// Pipeline 缓存管道接口
type Pipeline interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration)
	Exec(ctx context.Context) ([]interface{}, error)
}

// Lock 分布式锁接口
type Lock interface {
	Lock() bool
	Unlock() error
}
