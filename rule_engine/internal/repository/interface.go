package repository

import (
	"context"
	"time"

	"github.com/xwaf/rule_engine/internal/model"
)

// RuleStatsRepository 规则统计仓库接口
type RuleStatsRepository interface {
	// GetRuleStats 获取规则统计信息
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	GetRuleStats(ctx context.Context) (*model.RuleStats, error)

	// GetRuleMatchStats 获取规则匹配统计信息
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	// - ErrCache: 缓存操作失败
	GetRuleMatchStats(ctx context.Context, ruleID int64, startTime, endTime time.Time) (*model.RuleMatchStat, error)
}

// RuleAuditRepository 规则审计仓库接口
type RuleAuditRepository interface {
	// CreateAuditLog 创建审计日志
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库写入失败
	// - ErrValidation: 日志数据验证失败
	CreateAuditLog(ctx context.Context, log *model.RuleAuditLog) error

	// GetAuditLogs 获取审计日志列表
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	// - ErrRuleNotFound: 规则不存在
	GetAuditLogs(ctx context.Context, ruleID int64) ([]*model.RuleAuditLog, error)
}

// RuleCache 规则缓存接口
type RuleCache interface {
	// GetRule 获取规则缓存
	// 返回错误:
	// - ErrCache: 缓存操作失败
	// - ErrCacheMiss: 缓存未命中
	GetRule(ctx context.Context, id int64) (*model.Rule, error)

	// SetRule 设置规则缓存
	// 返回错误:
	// - ErrCache: 缓存操作失败
	// - ErrValidation: 规则数据验证失败
	SetRule(ctx context.Context, rule *model.Rule) error

	// DeleteRule 删除规则缓存
	// 返回错误:
	// - ErrCache: 缓存操作失败
	DeleteRule(ctx context.Context, id int64) error

	// ClearRules 清空规则缓存
	// 返回错误:
	// - ErrCache: 缓存操作失败
	ClearRules(ctx context.Context) error
}

// CacheRepository 缓存仓库接口
type CacheRepository interface {
	// Get 获取缓存
	// 返回错误:
	// - ErrCache: 缓存操作失败
	// - ErrCacheMiss: 缓存未命中
	Get(ctx context.Context, key string, value interface{}) error

	// Set 设置缓存
	// 返回错误:
	// - ErrCache: 缓存操作失败
	// - ErrValidation: 数据验证失败
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Delete 删除缓存
	// 返回错误:
	// - ErrCache: 缓存操作失败
	Delete(ctx context.Context, key string) error
}

// RuleRepository 规则仓库接口
type RuleRepository interface {
	// 基础操作
	// CreateRule 创建规则
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库写入失败
	// - ErrValidation: 规则验证失败
	// - ErrRuleConflict: 规则名称冲突
	CreateRule(ctx context.Context, rule *model.Rule) error

	// BatchCreateRules 批量创建规则
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库写入失败
	// - ErrValidation: 规则验证失败
	// - ErrRuleConflict: 规则名称冲突
	BatchCreateRules(ctx context.Context, rules []*model.Rule) error

	// UpdateRule 更新规则
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库更新失败
	// - ErrValidation: 规则验证失败
	// - ErrRuleNotFound: 规则不存在
	// - ErrRuleConflict: 规则名称冲突
	UpdateRule(ctx context.Context, rule *model.Rule) error

	// BatchUpdateRules 批量更新规则
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库更新失败
	// - ErrValidation: 规则验证失败
	// - ErrRuleNotFound: 规则不存在
	// - ErrRuleConflict: 规则名称冲突
	BatchUpdateRules(ctx context.Context, rules []*model.Rule) error

	// DeleteRule 删除规则
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库删除失败
	// - ErrRuleNotFound: 规则不存在
	DeleteRule(ctx context.Context, id int64) error

	// BatchDeleteRules 批量删除规则
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库删除失败
	// - ErrRuleNotFound: 规则不存在
	BatchDeleteRules(ctx context.Context, ids []int64) error

	// GetRule 获取规则
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	// - ErrRuleNotFound: 规则不存在
	GetRule(ctx context.Context, id int64) (*model.Rule, error)

	// GetRuleByName 根据名称获取规则
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	// - ErrRuleNotFound: 规则不存在
	// - ErrValidation: 规则名称验证失败
	GetRuleByName(ctx context.Context, name string) (*model.Rule, error)

	// ListRules 获取规则列表
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	// - ErrValidation: 查询参数验证失败
	ListRules(ctx context.Context, query *RuleQuery) ([]*model.Rule, int64, error)

	// GetLatestVersion 获取最新版本号
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	GetLatestVersion(ctx context.Context) (int64, error)

	// ImportRules 导入规则
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库写入失败
	// - ErrValidation: 规则验证失败
	// - ErrRuleConflict: 规则名称冲突
	ImportRules(ctx context.Context, rules []*model.Rule) error

	// 事务相关
	// BeginTx 开启事务
	// 返回错误:
	// - ErrSystem: 系统错误，如事务开启失败
	BeginTx(ctx context.Context) (Transaction, error)

	// 规则统计
	// GetRuleStats 获取规则统计信息
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	GetRuleStats(ctx context.Context) (*model.RuleStats, error)

	// GetRuleMatchStats 获取规则匹配统计信息
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	// - ErrCache: 缓存操作失败
	GetRuleMatchStats(ctx context.Context, ruleID int64, startTime, endTime time.Time) (*model.RuleMatchStat, error)

	// IncrRuleMatchCount 增加规则匹配计数
	// 返回错误:
	// - ErrCache: 缓存操作失败
	IncrRuleMatchCount(ctx context.Context, ruleID int64) error

	// GetRuleMatchCount 获取规则匹配计数
	// 返回错误:
	// - ErrCache: 缓存操作失败
	GetRuleMatchCount(ctx context.Context, ruleID int64) (int64, error)

	// 规则审计
	// GetRuleAuditLogs 获取规则审计日志
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	// - ErrRuleNotFound: 规则不存在
	GetRuleAuditLogs(ctx context.Context, ruleID int64) ([]*model.RuleAuditLog, error)

	// CreateRuleAuditLog 创建规则审计日志
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库写入失败
	// - ErrValidation: 日志数据验证失败
	CreateRuleAuditLog(ctx context.Context, log *model.RuleAuditLog) error
}

// Transaction 事务接口
type Transaction interface {
	// Commit 提交事务
	// 返回错误:
	// - ErrSystem: 系统错误，如事务提交失败
	Commit() error

	// Rollback 回滚事务
	// 返回错误:
	// - ErrSystem: 系统错误，如事务回滚失败
	Rollback() error
}

// RuleVersionRepository 规则版本仓库接口
type RuleVersionRepository interface {
	// CreateVersion 创建规则版本
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库写入失败
	// - ErrValidation: 版本数据验证失败
	CreateVersion(ctx context.Context, version *model.RuleVersion) error

	// GetVersion 获取规则版本
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	// - ErrRuleNotFound: 规则不存在
	GetVersion(ctx context.Context, ruleID, version int64) (*model.RuleVersion, error)

	// ListVersions 获取规则版本列表
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	// - ErrRuleNotFound: 规则不存在
	ListVersions(ctx context.Context, ruleID int64) ([]*model.RuleVersion, error)

	// GetRulesByVersion 获取指定版本的规则列表
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	GetRulesByVersion(ctx context.Context, version int64) ([]*model.Rule, error)

	// GetLatestVersion 获取最新版本号
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	GetLatestVersion(ctx context.Context) (int64, error)

	// RollbackRules 回滚规则
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库操作失败
	// - ErrValidation: 回滚数据验证失败
	RollbackRules(ctx context.Context, rules []*model.Rule, event *model.RuleUpdateEvent) error

	// RefreshRules 刷新规则
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库操作失败
	RefreshRules(ctx context.Context) error

	// CreateSyncLog 创建同步日志
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库写入失败
	// - ErrValidation: 日志数据验证失败
	CreateSyncLog(ctx context.Context, log *model.RuleSyncLog) error

	// ListSyncLogs 获取同步日志列表
	// 返回错误:
	// - ErrSystem: 系统错误，如数据库查询失败
	// - ErrRuleNotFound: 规则不存在
	ListSyncLogs(ctx context.Context, ruleID int64) ([]*model.RuleSyncLog, error)
}

// Pipeline 缓存管道接口
type Pipeline interface {
	// Set 设置缓存
	// 返回错误:
	// - ErrCache: 缓存操作失败
	// - ErrValidation: 数据验证失败
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration)

	// Exec 执行管道操作
	// 返回错误:
	// - ErrCache: 缓存操作失败
	Exec(ctx context.Context) ([]interface{}, error)
}

// Lock 分布式锁接口
type Lock interface {
	// Lock 获取锁
	// 返回:
	// - true: 获取锁成功
	// - false: 获取锁失败
	Lock() bool

	// Unlock 释放锁
	// 返回错误:
	// - ErrSystem: 系统错误，如锁释放失败
	Unlock() error
}
