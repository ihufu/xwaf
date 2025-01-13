package model

import (
	"fmt"
	"time"

	"github.com/xwaf/rule_engine/internal/errors"
)

// LimitUnit 限制单位
type LimitUnit string

const (
	LimitUnitSecond LimitUnit = "second" // 每秒
	LimitUnitMinute LimitUnit = "minute" // 每分钟
	LimitUnitHour   LimitUnit = "hour"   // 每小时
	LimitUnitDay    LimitUnit = "day"    // 每天
)

// CCStatus CC规则状态
type CCStatus string

const (
	CCStatusEnabled  CCStatus = "enabled"  // 启用
	CCStatusDisabled CCStatus = "disabled" // 禁用
)

// CCRule CC防护规则
type CCRule struct {
	ID         int64     `json:"id" db:"id"`
	URI        string    `json:"uri" db:"uri"`
	LimitRate  int       `json:"limit_rate" db:"limit_rate"`
	TimeWindow int       `json:"time_window" db:"time_window"`
	LimitUnit  LimitUnit `json:"limit_unit" db:"limit_unit"`
	Status     CCStatus  `json:"status" db:"status"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// CCRuleQuery CC规则查询条件
type CCRuleQuery struct {
	URI       string    `json:"uri"`
	Status    CCStatus  `json:"status"`
	LimitUnit LimitUnit `json:"limit_unit"`
}

// Validate 验证CC规则
func (r *CCRule) Validate() error {
	// 验证URI
	if r.URI == "" {
		return errors.NewError(errors.ErrRuleValidation, "URI不能为空")
	}
	if len(r.URI) > 255 {
		return errors.NewError(errors.ErrRuleValidation, "URI长度不能超过255个字符")
	}

	// 验证限制速率
	if r.LimitRate <= 0 {
		return errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("无效的限制速率: %d", r.LimitRate))
	}

	// 验证时间窗口
	if r.TimeWindow <= 0 {
		return errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("无效的时间窗口: %d", r.TimeWindow))
	}

	// 验证限制单位
	switch r.LimitUnit {
	case LimitUnitSecond, LimitUnitMinute, LimitUnitHour, LimitUnitDay:
		// 合法的限制单位
	default:
		return errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("无效的限制单位: %s", r.LimitUnit))
	}

	// 验证状态
	switch r.Status {
	case CCStatusEnabled, CCStatusDisabled:
		// 合法的状态
	default:
		return errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("无效的状态: %s", r.Status))
	}

	return nil
}
