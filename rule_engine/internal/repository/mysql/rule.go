package mysql

import (
	"context"
	"fmt"
	"time"

	"regexp"

	"github.com/go-redis/redis/v8"
	"github.com/xwaf/rule_engine/internal/errors"
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
	if err := r.db.WithContext(ctx).Create(rule).Error; err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("创建规则失败: %v", err))
	}
	return nil
}

func (r *RuleRepository) UpdateRule(ctx context.Context, rule *model.Rule) error {
	result := r.db.WithContext(ctx).Save(rule)
	if result.Error != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("更新规则失败: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("规则不存在: %d", rule.ID))
	}
	return nil
}

func (r *RuleRepository) DeleteRule(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&model.Rule{}, id)
	if result.Error != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("删除规则失败: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("规则不存在: %d", id))
	}
	return nil
}

func (r *RuleRepository) GetRule(ctx context.Context, id int64) (*model.Rule, error) {
	var rule model.Rule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("规则不存在: %d", id))
		}
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取规则失败: %v", err))
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
		return nil, 0, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取规则总数失败: %v", err))
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Find(&rules).Error; err != nil {
		return nil, 0, errors.NewError(errors.ErrSystem, fmt.Sprintf("查询规则列表失败: %v", err))
	}

	return rules, total, nil
}

func (r *RuleRepository) GetLatestVersion(ctx context.Context) (int64, error) {
	var version int64
	if err := r.db.WithContext(ctx).Model(&model.Rule{}).Select("COALESCE(MAX(version), 0)").Scan(&version).Error; err != nil {
		return 0, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取最新版本号失败: %v", err))
	}
	return version, nil
}

func (r *RuleRepository) BatchCreateRules(ctx context.Context, rules []*model.Rule) error {
	if err := r.db.WithContext(ctx).Create(rules).Error; err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("批量创建规则失败: %v", err))
	}
	return nil
}

func (r *RuleRepository) BatchDeleteRules(ctx context.Context, ids []int64) error {
	result := r.db.WithContext(ctx).Delete(&model.Rule{}, ids)
	if result.Error != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("批量删除规则失败: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return errors.NewError(errors.ErrRuleNotFound, "未找到要删除的规则")
	}
	return nil
}

func (r *RuleRepository) BatchUpdateRules(ctx context.Context, rules []*model.Rule) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rule := range rules {
			result := tx.Save(rule)
			if result.Error != nil {
				return errors.NewError(errors.ErrSystem, fmt.Sprintf("更新规则失败: %v", result.Error))
			}
			if result.RowsAffected == 0 {
				return errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("规则不存在: %d", rule.ID))
			}
		}
		return nil
	})
}

func (r *RuleRepository) CreateRuleAuditLog(ctx context.Context, log *model.RuleAuditLog) error {
	// 设置创建时间
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	// 将审计日志保存到数据库
	result := r.db.WithContext(ctx).Create(log)
	if result.Error != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("创建规则审计日志失败: %v", result.Error))
	}

	return nil
}

func (r *RuleRepository) CreateRuleGroup(ctx context.Context, group *model.RuleGroup) error {
	if err := r.ValidateRuleGroup(ctx, group); err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(group).Error; err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("创建规则组失败: %v", err))
	}
	return nil
}

func (r *RuleRepository) UpdateRuleGroup(ctx context.Context, group *model.RuleGroup) error {
	if err := r.ValidateRuleGroup(ctx, group); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Save(group)
	if result.Error != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("更新规则组失败: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("规则组不存在: %d", group.ID))
	}
	return nil
}

func (r *RuleRepository) DeleteRuleGroup(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&model.RuleGroup{}, id)
	if result.Error != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("删除规则组失败: %v", result.Error))
	}
	if result.RowsAffected == 0 {
		return errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("规则组不存在: %d", id))
	}
	return nil
}

func (r *RuleRepository) GetRuleGroup(ctx context.Context, id int64) (*model.RuleGroup, error) {
	var group model.RuleGroup
	err := r.db.WithContext(ctx).First(&group, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("规则组不存在: %d", id))
		}
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取规则组失败: %v", err))
	}
	return &group, nil
}

func (r *RuleRepository) ListRuleGroups(ctx context.Context, query *repository.RuleQuery) ([]*model.RuleGroup, int64, error) {
	db := r.db.WithContext(ctx).Model(&model.RuleGroup{})

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

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

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

	var groups []*model.RuleGroup
	if err := db.Find(&groups).Error; err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

func (r *RuleRepository) GetRulesByVersion(ctx context.Context, version int64) ([]*model.Rule, error) {
	var rules []*model.Rule
	err := r.db.WithContext(ctx).Where("version = ?", version).Find(&rules).Error
	if err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取指定版本规则失败: %v", err))
	}
	return rules, nil
}

func (r *RuleRepository) RollbackRules(ctx context.Context, version int64) error {
	// 获取当前规则
	var currentRules []*model.Rule
	if err := r.db.WithContext(ctx).Find(&currentRules).Error; err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("获取当前规则失败: %v", err))
	}

	// 获取目标版本规则
	var targetRules []*model.Rule
	if err := r.db.WithContext(ctx).Where("version = ?", version).Find(&targetRules).Error; err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("获取目标版本规则失败: %v", err))
	}

	// 创建回滚事件
	event := &model.RuleUpdateEvent{
		Version:   version,
		Action:    model.RuleUpdateTypeRollback,
		RuleDiffs: make([]*model.RuleDiff, 0),
		CreatedAt: time.Now(),
	}

	// 记录规则变更
	for _, rule := range currentRules {
		diff := &model.RuleDiff{
			RuleID:     rule.ID,
			Name:       rule.Name,
			Pattern:    rule.Pattern,
			Action:     rule.Action,
			Status:     rule.Status,
			Version:    version,
			UpdateType: model.RuleUpdateTypeDelete,
			UpdateTime: time.Now(),
		}
		event.RuleDiffs = append(event.RuleDiffs, diff)
	}

	for _, rule := range targetRules {
		diff := &model.RuleDiff{
			RuleID:     rule.ID,
			Name:       rule.Name,
			Pattern:    rule.Pattern,
			Action:     rule.Action,
			Status:     rule.Status,
			Version:    version,
			UpdateType: model.RuleUpdateTypeCreate,
			UpdateTime: time.Now(),
		}
		event.RuleDiffs = append(event.RuleDiffs, diff)
	}

	// 开启事务
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("开启事务失败: %v", tx.Error))
	}

	// 删除当前规则
	if err := tx.Delete(&model.Rule{}).Error; err != nil {
		tx.Rollback()
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("删除当前规则失败: %v", err))
	}

	// 创建目标版本规则
	if err := tx.Create(&targetRules).Error; err != nil {
		tx.Rollback()
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("创建目标版本规则失败: %v", err))
	}

	// 记录回滚事件
	if err := tx.Create(event).Error; err != nil {
		tx.Rollback()
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("记录回滚事件失败: %v", err))
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("提交事务失败: %v", err))
	}

	return nil
}

func (r *RuleRepository) ValidateRule(ctx context.Context, rule *model.Rule) error {
	if rule == nil {
		return errors.NewError(errors.ErrRuleValidation, "规则不能为空")
	}

	if rule.Name == "" {
		return errors.NewError(errors.ErrRuleValidation, "规则名称不能为空")
	}

	if rule.Pattern == "" {
		return errors.NewError(errors.ErrRuleValidation, "规则模式不能为空")
	}

	// 检查规则名称是否重复
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Rule{}).
		Where("name = ? AND id != ?", rule.Name, rule.ID).
		Count(&count).Error
	if err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("检查规则名称重复失败: %v", err))
	}
	if count > 0 {
		return errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("规则名称已存在: %s", rule.Name))
	}

	// 验证正则表达式
	if _, err := regexp.Compile(rule.Pattern); err != nil {
		return errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("规则模式不是有效的正则表达式: %v", err))
	}

	return nil
}

func (r *RuleRepository) ValidateRuleGroup(ctx context.Context, group *model.RuleGroup) error {
	if group == nil {
		return errors.NewError(errors.ErrRuleValidation, "规则组不能为空")
	}

	if group.Name == "" {
		return errors.NewError(errors.ErrRuleValidation, "规则组名称不能为空")
	}

	// 检查规则组名称是否重复
	var count int64
	err := r.db.WithContext(ctx).Model(&model.RuleGroup{}).
		Where("name = ? AND id != ?", group.Name, group.ID).
		Count(&count).Error
	if err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("检查规则组名称重复失败: %v", err))
	}
	if count > 0 {
		return errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("规则组名称已存在: %s", group.Name))
	}
	return nil
}

func (r *RuleRepository) GetRuleStats(ctx context.Context) (*model.RuleStats, error) {
	var stats model.RuleStats

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取总规则数
		if err := tx.Model(&model.Rule{}).Count(&stats.TotalRules).Error; err != nil {
			return errors.NewError(errors.ErrSystem, fmt.Sprintf("获取规则总数失败: %v", err))
		}

		// 获取启用规则数
		if err := tx.Model(&model.Rule{}).Where("status = ?", model.RuleStatusEnabled).Count(&stats.EnabledRules).Error; err != nil {
			return errors.NewError(errors.ErrSystem, fmt.Sprintf("获取启用规则数失败: %v", err))
		}

		// 获取禁用规则数
		if err := tx.Model(&model.Rule{}).Where("status = ?", model.RuleStatusDisabled).Count(&stats.DisabledRules).Error; err != nil {
			return errors.NewError(errors.ErrSystem, fmt.Sprintf("获取禁用规则数失败: %v", err))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *RuleRepository) GetRuleMatchStats(ctx context.Context, ruleID int64, startTime, endTime time.Time) (*model.RuleMatchStat, error) {
	// 从缓存获取
	key := fmt.Sprintf("rule_match_stats:%d:%d:%d", ruleID, startTime.Unix(), endTime.Unix())
	var stat model.RuleMatchStat
	if err := r.rdb.Get(ctx, key).Scan(&stat); err == nil {
		return &stat, nil
	}

	// 从数据库获取
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Rule{}).Where("id = ?", ruleID).Count(&total).Error; err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取规则匹配统计失败: %v", err))
	}

	stat = model.RuleMatchStat{
		RuleID:    ruleID,
		StartTime: startTime,
		EndTime:   endTime,
		Total:     total,
		Timeline:  make([]*model.RuleMatchPoint, 0),
	}

	// 缓存统计结果
	if err := r.rdb.Set(ctx, key, &stat, time.Hour).Err(); err != nil {
		return nil, errors.NewError(errors.ErrCache, fmt.Sprintf("缓存规则匹配统计失败: %v", err))
	}

	return &stat, nil
}

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
					return errors.NewError(errors.ErrSystem, fmt.Sprintf("更新规则失败: %v", err))
				}
			} else if err == gorm.ErrRecordNotFound {
				// 规则不存在，创建
				if err := tx.Create(rule).Error; err != nil {
					return errors.NewError(errors.ErrSystem, fmt.Sprintf("创建规则失败: %v", err))
				}
			} else {
				return errors.NewError(errors.ErrSystem, fmt.Sprintf("检查规则是否存在失败: %v", err))
			}
		}
		return nil
	})
}

func (r *RuleRepository) ExportRules(ctx context.Context, query *repository.RuleQuery) ([]*model.Rule, error) {
	var rules []*model.Rule
	db := r.db.WithContext(ctx)

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
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("导出规则失败: %v", err))
	}
	return rules, nil
}

func (r *RuleRepository) TestRule(ctx context.Context, rule *model.Rule, testCase *model.RuleTestCase) (*model.RuleTestResult, error) {
	result := &model.RuleTestResult{
		TestCase: testCase,
	}

	// 记录开始时间
	startTime := time.Now()

	// 根据规则类型执行不同的匹配逻辑
	switch rule.Type {
	case model.RuleTypeSQLi:
		detector := model.NewSQLInjectionDetector()
		isInjection, err := detector.DetectInjection(testCase.Input)
		if err != nil {
			result.Error = errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("SQL注入检测失败: %v", err)).Error()
			return result, nil
		}
		result.IsMatch = isInjection
		if isInjection {
			result.MatchResult = &model.RuleMatch{
				Rule:       rule,
				MatchedStr: testCase.Input,
				Position:   0,
				Score:      1.0,
			}
			result.Error = errors.NewError(errors.ErrRuleValidation, "检测到SQL注入攻击").Error()
		}

	case model.RuleTypeXSS:
		// XSS检测
		err := model.ValidateXSSRule(rule)
		if err != nil {
			result.Error = errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("XSS规则验证失败: %v", err)).Error()
			return result, nil
		}
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			result.Error = errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("XSS规则正则表达式编译错误: %v", err)).Error()
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
		return nil, errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("不支持的规则类型: %s", rule.Type))
	}

	// 记录执行时间
	result.Duration = time.Since(startTime)

	return result, nil
}

func (r *RuleRepository) ListRuleTestCases(ctx context.Context, ruleID int64) ([]*model.RuleTestCase, error) {
	var testCases []*model.RuleTestCase
	err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Find(&testCases).Error
	if err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取规则测试用例失败: %v", err))
	}
	return testCases, nil
}

func (r *RuleRepository) GetRuleAuditLogs(ctx context.Context, ruleID int64) ([]*model.RuleAuditLog, error) {
	var logs []*model.RuleAuditLog
	err := r.db.WithContext(ctx).Where("rule_id = ?", ruleID).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取规则审计日志失败: %v", err))
	}
	return logs, nil
}

func (r *RuleRepository) IncrRuleMatchCount(ctx context.Context, ruleID int64) error {
	key := fmt.Sprintf("rule:match:count:%d", ruleID)
	if err := r.rdb.Incr(ctx, key).Err(); err != nil {
		return errors.NewError(errors.ErrCache, fmt.Sprintf("增加规则匹配计数失败: %v", err))
	}
	return nil
}

func (r *RuleRepository) GetRuleMatchCount(ctx context.Context, ruleID int64) (int64, error) {
	key := fmt.Sprintf("rule:match:count:%d", ruleID)
	count, err := r.rdb.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, errors.NewError(errors.ErrCache, fmt.Sprintf("获取规则匹配计数失败: %v", err))
	}
	return count, nil
}

func (r *RuleRepository) GetVersion(ctx context.Context) (int64, error) {
	var version int64
	err := r.db.WithContext(ctx).Model(&model.Rule{}).Select("COALESCE(MAX(version), 0)").Scan(&version).Error
	if err != nil {
		return 0, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取规则版本失败: %v", err))
	}
	return version, nil
}

type GormTransaction struct {
	tx *gorm.DB
}

func (t *GormTransaction) Commit() error {
	return t.tx.Commit().Error
}

func (t *GormTransaction) Rollback() error {
	return t.tx.Rollback().Error
}

func (r *RuleRepository) BeginTx(ctx context.Context) (repository.Transaction, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("开启事务失败: %v", tx.Error))
	}
	return &GormTransaction{tx: tx}, nil
}

func (r *RuleRepository) GetRuleByName(ctx context.Context, name string) (*model.Rule, error) {
	if name == "" {
		return nil, errors.NewError(errors.ErrRuleValidation, "规则名称不能为空")
	}

	var rule model.Rule
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&rule).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewError(errors.ErrRuleNotFound, fmt.Sprintf("规则不存在: %s", name))
		}
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("获取规则失败: %v", err))
	}
	return &rule, nil
}
