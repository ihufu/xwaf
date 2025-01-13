package model

import (
	"fmt"
	"sort"
	"time"

	"github.com/xwaf/rule_engine/internal/errors"
)

// RuleVariable 规则变量类型
type RuleVariable string

const (
	RuleVarRequestURI     RuleVariable = "request_uri"
	RuleVarRequestHeaders RuleVariable = "request_headers"
	RuleVarRequestArgs    RuleVariable = "request_args"
	RuleVarRequestBody    RuleVariable = "request_body"
	RuleVarRequestMethod  RuleVariable = "request_method"
	RuleVarResponse       RuleVariable = "response"
)

// RuleType 规则类型
type RuleType string

const (
	RuleTypeIP     RuleType = "ip"     // IP规则
	RuleTypeCC     RuleType = "cc"     // CC规则
	RuleTypeRegex  RuleType = "regex"  // 正则规则
	RuleTypeSQLi   RuleType = "sqli"   // SQL注入规则
	RuleTypeXSS    RuleType = "xss"    // XSS规则
	RuleTypeCustom RuleType = "custom" // 自定义规则
)

// ActionType 动作类型
type ActionType string

const (
	ActionBlock    ActionType = "block"    // 阻止
	ActionAllow    ActionType = "allow"    // 允许
	ActionLog      ActionType = "log"      // 记录日志
	ActionRedirect ActionType = "redirect" // 重定向
	ActionCaptcha  ActionType = "captcha"  // 验证码
)

// StatusType 状态类型
type StatusType string

const (
	StatusEnabled  StatusType = "enabled"  // 启用
	StatusDisabled StatusType = "disabled" // 禁用
)

// SeverityType 风险级别
type SeverityType string

const (
	SeverityHigh   SeverityType = "high"   // 高风险
	SeverityMedium SeverityType = "medium" // 中风险
	SeverityLow    SeverityType = "low"    // 低风险
)

// RuleStatus 规则状态
type RuleStatus string

const (
	RuleStatusEnabled  RuleStatus = "enabled"  // 启用
	RuleStatusDisabled RuleStatus = "disabled" // 禁用
)

// SQLInjectType SQL注入类型
type SQLInjectType string

const (
	SQLInjectTypeUnion     SQLInjectType = "union"   // UNION注入
	SQLInjectTypeError     SQLInjectType = "error"   // 错误注入
	SQLInjectTypeBoolean   SQLInjectType = "boolean" // 布尔注入
	SQLInjectTypeTime      SQLInjectType = "time"    // 时间注入
	SQLInjectTypeOutOfBand SQLInjectType = "oob"     // 带外注入
	SQLInjectTypeStacked   SQLInjectType = "stacked" // 堆叠注入
	SQLInjectTypeBlind     SQLInjectType = "blind"   // 盲注
	SQLInjectTypeSecond    SQLInjectType = "second"  // 二阶注入
)

// SQLInjectRule SQL注入规则
type SQLInjectRule struct {
	Type        SQLInjectType `json:"type"`        // 注入类型
	Pattern     string        `json:"pattern"`     // 匹配模式
	Description string        `json:"description"` // 规则描述
	Risk        int           `json:"risk"`        // 风险等级
	Examples    []string      `json:"examples"`    // 示例
	Solutions   []string      `json:"solutions"`   // 解决方案
	References  []string      `json:"references"`  // 参考资料
	Confidence  float64       `json:"confidence"`  // 置信度
	Impact      float64       `json:"impact"`      // 影响度
	CVSS        float64       `json:"cvss"`        // CVSS评分
	CWE         string        `json:"cwe"`         // CWE编号
}

// SQLInjectRules SQL注入规则列表
var SQLInjectRules = []SQLInjectRule{
	{
		Type:        SQLInjectTypeUnion,
		Pattern:     `(?i)(UNION.*SELECT|SELECT.*\bUNION\b)`,
		Description: "检测UNION SELECT注入",
		Risk:        9,
		Examples: []string{
			"' UNION SELECT 1,2,3--",
			"') UNION SELECT * FROM users--",
		},
		Solutions: []string{
			"使用参数化查询",
			"过滤UNION关键字",
			"限制查询结果集",
		},
		References: []string{
			"https://owasp.org/www-community/attacks/SQL_Injection",
			"https://portswigger.net/web-security/sql-injection/union-attacks",
		},
		Confidence: 0.9,
		Impact:     0.8,
		CVSS:       8.5,
		CWE:        "CWE-89",
	},
}

// SQLTokenType Token类型
type SQLTokenType int

const (
	SQLTokenError SQLTokenType = iota
	SQLTokenEOF
	SQLTokenKeyword
	SQLTokenIdentifier
	SQLTokenString
	SQLTokenNumber
	SQLTokenOperator
	SQLTokenComment
)

// Rule 规则定义
type Rule struct {
	ID             int64        `json:"id" db:"id"`
	GroupID        int64        `json:"group_id" db:"group_id"`
	Name           string       `json:"name" db:"name"`
	Description    string       `json:"description" db:"description"`
	Pattern        string       `json:"pattern" db:"pattern"`
	Params         string       `json:"params" db:"params"`
	Type           RuleType     `json:"type" db:"type"`
	RuleVariable   RuleVariable `json:"rule_variable" db:"rule_variable"`
	Action         ActionType   `json:"action" db:"action"`
	Priority       int          `json:"priority" db:"priority"`
	Status         StatusType   `json:"status" db:"status"`
	Severity       SeverityType `json:"severity" db:"severity"`
	RulesOperation string       `json:"rules_operation" db:"rules_operation"`
	Version        int64        `json:"version" db:"version"`
	Hash           string       `json:"hash" db:"hash"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
	CreatedBy      int64        `json:"created_by" db:"created_by"`
	UpdatedBy      int64        `json:"updated_by" db:"updated_by"`
}

// ValidateXSSRule 验证XSS规则
func ValidateXSSRule(rule *Rule) error {
	if rule.Type != RuleTypeXSS {
		return errors.NewError(errors.ErrRuleValidation, "规则类型必须是XSS")
	}
	if rule.Pattern == "" {
		return errors.NewError(errors.ErrRuleValidation, "规则匹配模式不能为空")
	}
	return nil
}

// Validate 验证规则的合法性
func (r *Rule) Validate() error {
	if r.Name == "" {
		return errors.NewError(errors.ErrRuleValidation, "规则名称不能为空")
	}
	if r.Pattern == "" {
		return errors.NewError(errors.ErrRuleValidation, "规则匹配模式不能为空")
	}
	if r.Type == "" {
		return errors.NewError(errors.ErrRuleValidation, "规则类型不能为空")
	}
	if r.Action == "" {
		return errors.NewError(errors.ErrRuleValidation, "规则动作不能为空")
	}
	if r.Status == "" {
		return errors.NewError(errors.ErrRuleValidation, "规则状态不能为空")
	}

	// 验证规则类型的合法性
	switch r.Type {
	case RuleTypeIP, RuleTypeCC, RuleTypeRegex, RuleTypeSQLi, RuleTypeXSS, RuleTypeCustom:
		// 合法的规则类型
	default:
		return errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("无效的规则类型: %s", r.Type))
	}

	// 验证动作类型的合法性
	switch r.Action {
	case ActionBlock, ActionAllow, ActionLog, ActionRedirect, ActionCaptcha:
		// 合法的动作类型
	default:
		return errors.NewError(errors.ErrRuleValidation, fmt.Sprintf("无效的动作类型: %s", r.Action))
	}

	return nil
}

// SortRulesByPriority 按优先级排序规则列表
func SortRulesByPriority(rules []*Rule) {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority // 优先级高的排在前面
	})
}
