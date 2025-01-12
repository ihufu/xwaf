package mysql

import (
	"context"
	"fmt"
	"time"

	"regexp"

	"github.com/go-redis/redis/v8"
	"github.com/xwaf/rule_engine/internal/model"
	"github.com/xwaf/rule_engine/internal/repository"
	"gorm.io/gorm"
)

type RuleRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewRuleRepository(db *gorm.DB, rdb *redis.Client) repository.RuleRepository {
	return &RuleRepository{
		db:  db,
		rdb: rdb,
	}
}

func (r *RuleRepository) CreateRule(ctx context.Context, rule *model.Rule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *RuleRepository) UpdateRule(ctx context.Context, rule *model.Rule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *RuleRepository) DeleteRule(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Rule{}, id).Error
}

func (r *RuleRepository) GetRule(ctx context.Context, id int64) (*model.Rule, error) {
	var rule model.Rule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *RuleRepository) ListRules(ctx context.Context, query *repository.RuleQuery) ([]*model.Rule, int64, error) {
	var rules []*model.Rule
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Rule{})

	if query.Keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.RuleType != "" {
		db = db.Where("rule_type = ?", query.RuleType)
	}
	if query.RuleVariable != "" {
		db = db.Where("rule_variable = ?", query.RuleVariable)
	}
	if query.Severity != "" {
		db = db.Where("severity = ?", query.Severity)
	}
	if query.RulesOperation != "" {
		db = db.Where("rules_operation = ?", query.RulesOperation)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Find(&rules).Error; err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

func (r *RuleRepository) GetLatestVersion(ctx context.Context) (int64, error) {
	var version int64
	err := r.db.WithContext(ctx).Model(&model.Rule{}).Select("COALESCE(MAX(version), 0)").Scan(&version).Error
	return version, err
}

func (r *RuleRepository) BatchCreateRules(ctx context.Context, rules []*model.Rule) error {
	return r.db.WithContext(ctx).Create(rules).Error
}

func (r *RuleRepository) BatchDeleteRules(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Delete(&model.Rule{}, ids).Error
}

// BatchUpdateRules 批量更新规则
func (r *RuleRepository) BatchUpdateRules(ctx context.Context, rules []*model.Rule) error {
	// 使用事务确保批量更新的原子性
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rule := range rules {
			if err := tx.Save(rule).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// CreateRuleAuditLog 创建规则审计日志
func (r *RuleRepository) CreateRuleAuditLog(ctx context.Context, log *model.RuleAuditLog) error {
	// 设置创建时间
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	// 将审计日志保存到数据库
	result := r.db.WithContext(ctx).Create(log)
	if result.Error != nil {
		return fmt.Errorf("创建规则审计日志失败: %w", result.Error)
	}

	return nil
}

// CreateRuleGroup 创建规则组
func (r *RuleRepository) CreateRuleGroup(ctx context.Context, group *model.RuleGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// UpdateRuleGroup 更新规则组
func (r *RuleRepository) UpdateRuleGroup(ctx context.Context, group *model.RuleGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// DeleteRuleGroup 删除规则组
func (r *RuleRepository) DeleteRuleGroup(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.RuleGroup{}, id).Error
}

// GetRuleGroup 获取规则组
func (r *RuleRepository) GetRuleGroup(ctx context.Context, id int64) (*model.RuleGroup, error) {
	var group model.RuleGroup
	err := r.db.WithContext(ctx).First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// ListRuleGroups 列出规则组
func (r *RuleRepository) ListRuleGroups(ctx context.Context, query *repository.RuleQuery) ([]*model.RuleGroup, int64, error) {
	db := r.db.WithContext(ctx).Model(&model.RuleGroup{})

	// 应用查询条件
	if query.Keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.CreatedBy != 0 {
		db = db.Where("created_by = ?", query.CreatedBy)
	}
	if query.UpdatedBy != 0 {
		db = db.Where("updated_by = ?", query.UpdatedBy)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	// 获取总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序
	if query.OrderBy != "" {
		if query.OrderDesc {
			db = db.Order(query.OrderBy + " DESC")
		} else {
			db = db.Order(query.OrderBy)
		}
	}
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}

	// 获取数据
	var groups []*model.RuleGroup
	if err := db.Find(&groups).Error; err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// GetRulesByVersion 获取指定版本的规则
func (r *RuleRepository) GetRulesByVersion(ctx context.Context, version int64) ([]*model.Rule, error) {
	var rules []*model.Rule
	err := r.db.WithContext(ctx).Where("version = ?", version).Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// RollbackRules 回滚规则
func (r *RuleRepository) RollbackRules(ctx context.Context, rules []*model.Rule, event *model.RuleUpdateEvent) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 记录当前规则状态
		var currentRules []*model.Rule
		if err := tx.Where("version = ?", event.Version).Find(&currentRules).Error; err != nil {
			return fmt.Errorf("获取当前版本规则失败: %v", err)
		}

		// 创建回滚事件
		event.Action = "rollback"
		event.RuleDiffs = make([]*model.RuleDiff, 0)

		// 记录规则变更
		ruleMap := make(map[int64]*model.Rule)
		for _, rule := range currentRules {
			ruleMap[rule.ID] = rule
		}

		// 删除当前版本的规则
		if err := tx.Where("version = ?", event.Version).Delete(&model.Rule{}).Error; err != nil {
			return fmt.Errorf("删除当前版本规则失败: %v", err)
		}

		// 创建回滚的规则
		for _, rule := range rules {
			// 记录规则变更
			if oldRule, exists := ruleMap[rule.ID]; exists {
				event.RuleDiffs = append(event.RuleDiffs, &model.RuleDiff{
					OldRule: oldRule,
					NewRule: rule,
				})
			} else {
				event.RuleDiffs = append(event.RuleDiffs, &model.RuleDiff{
					OldRule: oldRule,
					NewRule: rule,
				})
			}

			// 创建规则
			if err := tx.Create(rule).Error; err != nil {
				return fmt.Errorf("创建回滚规则失败: %v", err)
			}
		}

		// 记录被删除的规则
		for _, oldRule := range currentRules {
			found := false
			for _, rule := range rules {
				if rule.ID == oldRule.ID {
					found = true
					break
				}
			}
			if !found {
				event.RuleDiffs = append(event.RuleDiffs, &model.RuleDiff{
					OldRule: oldRule,
					NewRule: oldRule,
				})
			}
		}

		// 记录回滚事件
		if err := tx.Create(event).Error; err != nil {
			return fmt.Errorf("记录回滚事件失败: %v", err)
		}

		return nil
	})
}

// ValidateRule 验证规则
func (r *RuleRepository) ValidateRule(ctx context.Context, rule *model.Rule) error {
	// 检查规则名称是否重复
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Rule{}).
		Where("name = ? AND id != ?", rule.Name, rule.ID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("rule name %s already exists", rule.Name)
	}
	return nil
}

// ValidateRuleGroup 验证规则组
func (r *RuleRepository) ValidateRuleGroup(ctx context.Context, group *model.RuleGroup) error {
	// 检查规则组名称是否重复
	var count int64
	err := r.db.WithContext(ctx).Model(&model.RuleGroup{}).
		Where("name = ? AND id != ?", group.Name, group.ID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("rule group name %s already exists", group.Name)
	}
	return nil
}

// GetRuleStats 获取规则统计信息
func (r *RuleRepository) GetRuleStats(ctx context.Context) (*model.RuleStats, error) {
	var stats model.RuleStats

	// 获取总规则数
	if err := r.db.WithContext(ctx).Model(&model.Rule{}).Count(&stats.TotalRules).Error; err != nil {
		return nil, err
	}

	// 获取启用和禁用规则数
	if err := r.db.WithContext(ctx).Model(&model.Rule{}).Where("status = ?", model.StatusEnabled).Count(&stats.EnabledRules).Error; err != nil {
		return nil, err
	}
	stats.DisabledRules = stats.TotalRules - stats.EnabledRules

	return &stats, nil
}

// GetRuleMatchStats 获取规则匹配统计
func (r *RuleRepository) GetRuleMatchStats(ctx context.Context, ruleID int64, startTime, endTime time.Time) (*model.RuleMatchStat, error) {
	stats := &model.RuleMatchStat{
		RuleID:    ruleID,
		StartTime: startTime,
		EndTime:   endTime,
		Timeline:  make([]*model.RuleMatchPoint, 0),
	}

	var points []*model.RuleMatchPoint
	err := r.db.WithContext(ctx).Table("rule_match_stats").
		Select("UNIX_TIMESTAMP(match_time) as timestamp, match_count as count").
		Where("rule_id = ? AND match_time BETWEEN ? AND ?", ruleID, startTime, endTime).
		Order("match_time ASC").
		Find(&points).Error
	if err != nil {
		return nil, err
	}

	total := int64(0)
	for _, point := range points {
		total += point.Count
	}

	stats.Timeline = points
	stats.Total = total

	return stats, nil
}

// ImportRules 导入规则
func (r *RuleRepository) ImportRules(ctx context.Context, rules []*model.Rule) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rule := range rules {
			// 检查规则是否已存在
			var existingRule model.Rule
			err := tx.Where("name = ?", rule.Name).First(&existingRule).Error
			if err == nil {
				// 规则已存在，更新
				rule.ID = existingRule.ID
				if err := tx.Save(rule).Error; err != nil {
					return err
				}
			} else if err == gorm.ErrRecordNotFound {
				// 规则不存在，创建
				if err := tx.Create(rule).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return nil
	})
}

// ExportRules 导出规则
func (r *RuleRepository) ExportRules(ctx context.Context, query *repository.RuleQuery) ([]*model.Rule, error) {
	var rules []*model.Rule
	db := r.db.WithContext(ctx)

	// 应用查询条件
	if query.RuleType != "" {
		db = db.Where("rule_type = ?", query.RuleType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Severity != "" {
		db = db.Where("severity = ?", query.Severity)
	}
	if query.GroupID != 0 {
		db = db.Where("group_id = ?", query.GroupID)
	}

	err := db.Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// TestRule 测试规则
func (r *RuleRepository) TestRule(ctx context.Context, rule *model.Rule, testCase *model.RuleTestCase) (*model.RuleTestResult, error) {
	result := &model.RuleTestResult{
		TestCase: testCase,
	}

	// 记录开始时间
	startTime := time.Now()

	// 根据规则类型执行不同的匹配逻辑
	switch rule.Type {
	case model.RuleTypeRegex:
		// 正则匹配
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			result.Error = fmt.Sprintf("正则表达式编译错误: %v", err)
			return result, nil
		}
		result.IsMatch = re.MatchString(testCase.Input)
		if result.IsMatch {
			result.MatchResult = &model.RuleMatch{
				Rule:       rule,
				MatchedStr: re.FindString(testCase.Input),
				Position:   re.FindStringIndex(testCase.Input)[0],
				Score:      1.0,
			}
		}

	case model.RuleTypeSQLi:
		// SQL注入检测
		detector := model.NewSQLInjectionDetector()
		isInjection, reason := detector.DetectInjection(testCase.Input)
		result.IsMatch = isInjection
		if result.IsMatch {
			result.MatchResult = &model.RuleMatch{
				Rule:       rule,
				MatchedStr: testCase.Input,
				Position:   0,
				Score:      1.0,
			}
			result.Error = reason
		}

	case model.RuleTypeXSS:
		// XSS检测
		err := model.ValidateXSSRule(rule)
		if err != nil {
			result.Error = fmt.Sprintf("XSS规则验证错误: %v", err)
			return result, nil
		}
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			result.Error = fmt.Sprintf("XSS规则正则表达式编译错误: %v", err)
			return result, nil
		}
		result.IsMatch = re.MatchString(testCase.Input)
		if result.IsMatch {
			result.MatchResult = &model.RuleMatch{
				Rule:       rule,
				MatchedStr: re.FindString(testCase.Input),
				Position:   re.FindStringIndex(testCase.Input)[0],
				Score:      1.0,
			}
		}

	default:
		result.Error = fmt.Sprintf("不支持的规则类型: %s", rule.Type)
		return result, nil
	}

	// 记录执行时间
	result.Duration = time.Since(startTime)

	return result, nil
}

// ListRuleTestCases 列出规则测试用例
func (r *RuleRepository) ListRuleTestCases(ctx context.Context, ruleID int64) ([]*model.RuleTestCase, error) {
	var testCases []*model.RuleTestCase
	err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Find(&testCases).Error
	if err != nil {
		return nil, err
	}
	return testCases, nil
}

// GetRuleAuditLogs 获取规则审计日志
func (r *RuleRepository) GetRuleAuditLogs(ctx context.Context, ruleID int64) ([]*model.RuleAuditLog, error) {
	var logs []*model.RuleAuditLog
	err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// IncrRuleMatchCount 增加规则匹配计数
func (r *RuleRepository) IncrRuleMatchCount(ctx context.Context, ruleID int64) error {
	key := fmt.Sprintf("rule:match:count:%d", ruleID)
	return r.rdb.Incr(ctx, key).Err()
}

// GetRuleMatchCount 获取规则匹配计数
func (r *RuleRepository) GetRuleMatchCount(ctx context.Context, ruleID int64) (int64, error) {
	key := fmt.Sprintf("rule:match:count:%d", ruleID)
	return r.rdb.Get(ctx, key).Int64()
}

// GetVersion 获取规则版本
func (r *RuleRepository) GetVersion(ctx context.Context) (int64, error) {
	var version int64
	err := r.db.WithContext(ctx).Model(&model.Rule{}).Select("COALESCE(MAX(version), 0)").Scan(&version).Error
	return version, err
}

// GormTransaction 包装 GORM 事务
type GormTransaction struct {
	tx *gorm.DB
}

func (t *GormTransaction) Commit() error {
	return t.tx.Commit().Error
}

func (t *GormTransaction) Rollback() error {
	return t.tx.Rollback().Error
}

// BeginTx 开启事务
func (r *RuleRepository) BeginTx(ctx context.Context) (repository.Transaction, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &GormTransaction{tx: tx}, nil
}

// GetRuleByName 根据名称获取规则
func (r *RuleRepository) GetRuleByName(ctx context.Context, name string) (*model.Rule, error) {
	var rule model.Rule
	err := r.db.Where("name = ?", name).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}
