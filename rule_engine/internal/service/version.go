package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
)

// RuleVersionService 规则版本服务接口
type RuleVersionService interface {
	// CreateVersion 创建规则版本
	CreateVersion(ctx context.Context, version *model.RuleVersion) error

	// GetVersion 获取规则版本
	GetVersion(ctx context.Context, ruleID, version int64) (*model.RuleVersion, error)

	// ListVersions 获取规则版本列表
	ListVersions(ctx context.Context, ruleID int64) ([]*model.RuleVersion, error)

	// SyncRules 同步规则
	SyncRules(ctx context.Context, event *model.RuleUpdateEvent) error

	// GetSyncLogs 获取同步日志
	GetSyncLogs(ctx context.Context, ruleID int64) ([]*model.RuleSyncLog, error)

	// RollbackToVersion 回滚到指定版本
	RollbackToVersion(ctx context.Context, version int64) error
}

// ruleVersionService 规则版本服务实现
type ruleVersionService struct {
	versionRepo repository.RuleVersionRepository
}

// NewRuleVersionService 创建规则版本服务
func NewRuleVersionService(versionRepo repository.RuleVersionRepository) RuleVersionService {
	return &ruleVersionService{
		versionRepo: versionRepo,
	}
}

// CreateVersion 创建规则版本
func (s *ruleVersionService) CreateVersion(ctx context.Context, version *model.RuleVersion) error {
	return s.versionRepo.CreateVersion(ctx, version)
}

// GetVersion 获取规则版本
func (s *ruleVersionService) GetVersion(ctx context.Context, ruleID, version int64) (*model.RuleVersion, error) {
	return s.versionRepo.GetVersion(ctx, ruleID, version)
}

// ListVersions 获取规则版本列表
func (s *ruleVersionService) ListVersions(ctx context.Context, ruleID int64) ([]*model.RuleVersion, error) {
	return s.versionRepo.ListVersions(ctx, ruleID)
}

// SyncRules 同步规则
func (s *ruleVersionService) SyncRules(ctx context.Context, event *model.RuleUpdateEvent) error {
	// 遍历规则差异
	for _, diff := range event.RuleDiffs {
		// 创建规则版本
		content, err := json.Marshal(diff)
		if err != nil {
			return fmt.Errorf("序列化规则差异失败: %v", err)
		}

		version := &model.RuleVersion{
			RuleID:     diff.RuleID,
			Version:    event.Version,
			Hash:       diff.NewRule.Hash,
			Content:    string(content),
			ChangeType: diff.Operation,
			Status:     "synced",
			CreatedBy:  diff.NewRule.UpdatedBy,
		}

		if err := s.versionRepo.CreateVersion(ctx, version); err != nil {
			return fmt.Errorf("创建规则版本失败: %v", err)
		}

		// 创建同步日志
		log := &model.RuleSyncLog{
			RuleID:   diff.RuleID,
			Version:  event.Version,
			Status:   "success",
			Message:  fmt.Sprintf("规则%s成功", diff.Operation),
			SyncType: diff.Operation,
		}

		if err := s.versionRepo.CreateSyncLog(ctx, log); err != nil {
			return fmt.Errorf("创建同步日志失败: %v", err)
		}
	}

	return nil
}

// GetSyncLogs 获取同步日志
func (s *ruleVersionService) GetSyncLogs(ctx context.Context, ruleID int64) ([]*model.RuleSyncLog, error) {
	return s.versionRepo.ListSyncLogs(ctx, ruleID)
}

// RollbackToVersion 回滚到指定版本
func (s *ruleVersionService) RollbackToVersion(ctx context.Context, version int64) error {
	// 获取目标版本的规则
	rules, err := s.versionRepo.GetRulesByVersion(ctx, version)
	if err != nil {
		return fmt.Errorf("获取历史版本规则失败: %v", err)
	}

	// 验证版本有效性
	currentVersion, err := s.versionRepo.GetLatestVersion(ctx)
	if err != nil {
		return fmt.Errorf("获取当前版本失败: %v", err)
	}

	if version >= currentVersion {
		return fmt.Errorf("无法回滚到更新的版本")
	}

	// 创建回滚事件
	event := &model.RuleUpdateEvent{
		Version:   currentVersion + 1,
		Action:    "rollback",
		RuleDiffs: make([]*model.RuleDiff, 0),
	}

	// 执行回滚
	if err := s.versionRepo.RollbackRules(ctx, rules, event); err != nil {
		return fmt.Errorf("规则回滚失败: %v", err)
	}

	// 更新缓存
	if err := s.versionRepo.RefreshRules(ctx); err != nil {
		return fmt.Errorf("刷新规则缓存失败: %v", err)
	}

	return nil
}
