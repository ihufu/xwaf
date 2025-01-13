package matcher

import (
	"context"
	"fmt"
	"sync"

	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
)

// ExprType 表达式类型
type ExprType int

const (
	ExprTypeRule ExprType = iota // 单个规则
	ExprTypeAnd                  // AND组合
	ExprTypeOr                   // OR组合
	ExprTypeNot                  // NOT操作
	ExprTypeAny                  // 任意匹配N个
	ExprTypeAll                  // 全部匹配N个
)

// Expression 规则表达式
type Expression struct {
	Type      ExprType
	Rule      *model.Rule  // 单个规则
	Children  []Expression // 子表达式
	Threshold int          // ANY/ALL阈值
}

// ExpressionMatcher 表达式匹配器
type ExpressionMatcher struct {
	expressions []Expression
	matchers    map[string]Matcher
	mutex       sync.RWMutex
}

// NewExpressionMatcher 创建表达式匹配器
func NewExpressionMatcher(matchers map[string]Matcher) *ExpressionMatcher {
	return &ExpressionMatcher{
		expressions: make([]Expression, 0),
		matchers:    matchers,
	}
}

// Add 添加规则表达式
func (m *ExpressionMatcher) Add(rule *model.Rule) error {
	if rule == nil {
		return errors.NewError(errors.ErrRuleMatch, "规则不能为空")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 解析规则组合操作
	expr, err := m.parseRuleOperation(rule)
	if err != nil {
		return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("解析规则操作失败: %v", err))
	}

	// 添加到表达式列表
	m.expressions = append(m.expressions, expr)

	// 将规则添加到对应的基础匹配器
	for _, matcher := range m.matchers {
		if err := matcher.Add(rule); err != nil {
			return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("添加规则到基础匹配器失败: %v", err))
		}
	}

	return nil
}

// parseRuleOperation 解析规则组合操作
func (m *ExpressionMatcher) parseRuleOperation(rule *model.Rule) (Expression, error) {
	switch rule.RulesOperation {
	case "and":
		return Expression{
			Type:     ExprTypeAnd,
			Children: []Expression{{Type: ExprTypeRule, Rule: rule}},
		}, nil
	case "or":
		return Expression{
			Type:     ExprTypeOr,
			Children: []Expression{{Type: ExprTypeRule, Rule: rule}},
		}, nil
	case "not":
		return Expression{
			Type:     ExprTypeNot,
			Children: []Expression{{Type: ExprTypeRule, Rule: rule}},
		}, nil
	case "any":
		return Expression{
			Type:      ExprTypeAny,
			Children:  []Expression{{Type: ExprTypeRule, Rule: rule}},
			Threshold: 1, // 默认阈值
		}, nil
	case "all":
		return Expression{
			Type:      ExprTypeAll,
			Children:  []Expression{{Type: ExprTypeRule, Rule: rule}},
			Threshold: 1, // 默认阈值
		}, nil
	default:
		return Expression{Type: ExprTypeRule, Rule: rule}, nil
	}
}

// Match 执行表达式匹配
func (m *ExpressionMatcher) Match(ctx context.Context, req *model.CheckRequest) ([]*model.RuleMatch, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("上下文已取消: %v", err))
	}

	if req == nil {
		return nil, errors.NewError(errors.ErrRuleMatch, "请求参数不能为空")
	}

	if req.URI == "" {
		return nil, errors.NewError(errors.ErrRuleMatch, "请求URI不能为空")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var matches []*model.RuleMatch
	for _, expr := range m.expressions {
		if matched, match, err := m.evaluateExpression(ctx, expr, req); err != nil {
			return nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("评估表达式失败: %v", err))
		} else if matched && match != nil {
			matches = append(matches, match)
		}
	}

	return matches, nil
}

// evaluateExpression 评估表达式
func (m *ExpressionMatcher) evaluateExpression(ctx context.Context, expr Expression, req *model.CheckRequest) (bool, *model.RuleMatch, error) {
	if err := ctx.Err(); err != nil {
		return false, nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("上下文已取消: %v", err))
	}

	switch expr.Type {
	case ExprTypeRule:
		// 使用基础匹配器进行匹配
		for _, matcher := range m.matchers {
			if matches, err := matcher.Match(ctx, req); err != nil {
				return false, nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("基础匹配器匹配失败: %v", err))
			} else if len(matches) > 0 {
				return true, matches[0], nil
			}
		}
		return false, nil, nil

	case ExprTypeAnd:
		for _, child := range expr.Children {
			if matched, _, err := m.evaluateExpression(ctx, child, req); err != nil {
				return false, nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("AND表达式评估失败: %v", err))
			} else if !matched {
				return false, nil, nil
			}
		}
		return true, &model.RuleMatch{Rule: expr.Rule}, nil

	case ExprTypeOr:
		for _, child := range expr.Children {
			if matched, match, err := m.evaluateExpression(ctx, child, req); err != nil {
				return false, nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("OR表达式评估失败: %v", err))
			} else if matched {
				return true, match, nil
			}
		}
		return false, nil, nil

	case ExprTypeNot:
		matched, _, err := m.evaluateExpression(ctx, expr.Children[0], req)
		if err != nil {
			return false, nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("NOT表达式评估失败: %v", err))
		}
		return !matched, &model.RuleMatch{Rule: expr.Rule}, nil

	case ExprTypeAny:
		matchCount := 0
		for _, child := range expr.Children {
			if matched, _, err := m.evaluateExpression(ctx, child, req); err != nil {
				return false, nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("ANY表达式评估失败: %v", err))
			} else if matched {
				matchCount++
				if matchCount >= expr.Threshold {
					return true, &model.RuleMatch{Rule: expr.Rule}, nil
				}
			}
		}
		return false, nil, nil

	case ExprTypeAll:
		matchCount := 0
		for _, child := range expr.Children {
			if matched, _, err := m.evaluateExpression(ctx, child, req); err != nil {
				return false, nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("ALL表达式评估失败: %v", err))
			} else if matched {
				matchCount++
			}
		}
		return matchCount >= expr.Threshold, &model.RuleMatch{Rule: expr.Rule}, nil

	default:
		return false, nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("未知的表达式类型: %v", expr.Type))
	}
}

// Remove 移除规则
func (m *ExpressionMatcher) Remove(ruleID int64) error {
	if ruleID <= 0 {
		return errors.NewError(errors.ErrRuleMatch, "无效的规则ID")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查规则是否存在
	exists := false
	for _, expr := range m.expressions {
		if expr.Rule.ID == ruleID {
			exists = true
			break
		}
	}

	if !exists {
		return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("规则不存在: %d", ruleID))
	}

	// 从表达式列表中移除
	newExpressions := make([]Expression, 0)
	for _, expr := range m.expressions {
		if expr.Rule.ID != ruleID {
			newExpressions = append(newExpressions, expr)
		}
	}
	m.expressions = newExpressions

	// 从基础匹配器中移除
	for _, matcher := range m.matchers {
		if err := matcher.Remove(ruleID); err != nil {
			return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("从基础匹配器移除规则失败: %v", err))
		}
	}

	return nil
}

// Clear 清空所有规则
func (m *ExpressionMatcher) Clear() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.expressions = make([]Expression, 0)
	for _, matcher := range m.matchers {
		if err := matcher.Clear(); err != nil {
			return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("清空基础匹配器失败: %v", err))
		}
	}
	return nil
}
