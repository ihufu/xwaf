package matcher

import (
	"context"

	"github.com/xwaf/rule_engine/internal/model"
)

// Matcher 定义规则匹配器接口
type Matcher interface {
	// Add 添加规则到匹配器
	Add(rule *model.Rule) error

	// Remove 从匹配器中移除规则
	Remove(ruleID int64) error

	// Match 执行规则匹配
	Match(ctx context.Context, req *model.CheckRequest) ([]*model.RuleMatch, error)

	// Clear 清空所有规则
	Clear() error
}

// Factory 匹配器工厂接口
type Factory interface {
	// CreateMatcher 创建对应类型的匹配器
	CreateMatcher(matcherType string) (Matcher, error)
}
