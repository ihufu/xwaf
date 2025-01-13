package matcher

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
)

const (
	maxRegexLength = 1000 // 正则表达式最大长度
	minPrefixLen   = 3    // 最小前缀长度
	maxPrefixLen   = 10   // 最大前缀长度
)

// RegexRule 正则规则结构
type RegexRule struct {
	rule    *model.Rule
	regex   *regexp.Regexp
	prefix  string
	suffix  string
	literal bool
}

// RegexMatcher 正则表达式匹配器
type RegexMatcher struct {
	rules    map[int64]*RegexRule
	prefixes map[string][]*RegexRule // 前缀索引
	mutex    sync.RWMutex
}

// NewRegexMatcher 创建新的正则匹配器
func NewRegexMatcher() *RegexMatcher {
	return &RegexMatcher{
		rules:    make(map[int64]*RegexRule),
		prefixes: make(map[string][]*RegexRule),
	}
}

// optimizeRegex 优化正则表达式
func (m *RegexMatcher) optimizeRegex(pattern string) (string, string, bool) {
	// 提取字面前缀
	prefix := ""
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '^' {
			continue
		}
		if isMetachar(pattern[i]) {
			break
		}
		prefix += string(pattern[i])
	}

	// 提取字面后缀
	suffix := ""
	for i := len(pattern) - 1; i >= 0; i-- {
		if pattern[i] == '$' {
			continue
		}
		if isMetachar(pattern[i]) {
			break
		}
		suffix = string(pattern[i]) + suffix
	}

	// 检查是否为纯字面量
	literal := !containsMetachars(pattern)

	return prefix, suffix, literal
}

// isMetachar 检查字符是否为正则元字符
func isMetachar(c byte) bool {
	switch c {
	case '*', '+', '?', '.', '^', '$', '(', ')', '[', ']', '{', '}', '|', '\\':
		return true
	}
	return false
}

// containsMetachars 检查字符串是否包含正则元字符
func containsMetachars(s string) bool {
	for i := 0; i < len(s); i++ {
		if isMetachar(s[i]) {
			return true
		}
	}
	return false
}

// validateRegex 验证正则表达式的复杂度
func (m *RegexMatcher) validateRegex(pattern string) error {
	if len(pattern) > maxRegexLength {
		return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("正则表达式过长，最大允许长度为 %d", maxRegexLength))
	}

	// 检查嵌套深度
	depth := 0
	maxDepth := 0
	for _, ch := range pattern {
		if ch == '(' {
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
		} else if ch == ')' {
			depth--
		}
	}

	if maxDepth > 5 {
		return errors.NewError(errors.ErrRuleMatch, "正则表达式嵌套深度过高，最大允许深度为 5")
	}

	// 检查重复次数
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '{' {
			end := strings.IndexByte(pattern[i:], '}')
			if end != -1 {
				repeat := pattern[i+1 : i+end]
				if n, err := strconv.Atoi(repeat); err == nil && n > 100 {
					return errors.NewError(errors.ErrRuleMatch, "正则表达式重复次数过多，最大允许重复次数为 100")
				}
			}
		}
	}

	return nil
}

// Add 添加规则到正则匹配器
func (m *RegexMatcher) Add(rule *model.Rule) error {
	// 验证正则表达式复杂度
	if err := m.validateRegex(rule.Pattern); err != nil {
		return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("正则表达式验证失败: %v", err))
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 编译正则表达式
	regex, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("编译正则表达式失败: %v", err))
	}

	// 优化正则表达式
	prefix, suffix, literal := m.optimizeRegex(rule.Pattern)
	regexRule := &RegexRule{
		rule:    rule,
		regex:   regex,
		prefix:  prefix,
		suffix:  suffix,
		literal: literal,
	}

	// 存储规则
	m.rules[rule.ID] = regexRule

	// 更新前缀索引
	if len(prefix) >= minPrefixLen {
		if len(prefix) > maxPrefixLen {
			prefix = prefix[:maxPrefixLen]
		}
		m.prefixes[prefix] = append(m.prefixes[prefix], regexRule)
	}

	return nil
}

// Remove 从正则匹配器中移除规则
func (m *RegexMatcher) Remove(ruleID int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	rule, exists := m.rules[ruleID]
	if !exists {
		return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("规则不存在: %d", ruleID))
	}

	// 从前缀索引中移除
	if rule.prefix != "" {
		rules := m.prefixes[rule.prefix]
		for i, r := range rules {
			if r.rule.ID == ruleID {
				m.prefixes[rule.prefix] = append(rules[:i], rules[i+1:]...)
				break
			}
		}
	}

	// 从规则映射中移除
	delete(m.rules, ruleID)
	return nil
}

// Match 执行正则匹配
func (m *RegexMatcher) Match(ctx context.Context, req *model.CheckRequest) ([]*model.RuleMatch, error) {
	// 检查 context 是否已取消
	if err := ctx.Err(); err != nil {
		return nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("上下文已取消: %v", err))
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	content := req.URI
	matches := make([]*model.RuleMatch, 0, 16)

	// 使用前缀索引进行快速过滤
	candidateRules := make([]*RegexRule, 0)
	for i := minPrefixLen; i <= len(content) && i <= maxPrefixLen; i++ {
		prefix := content[:i]
		if rules, ok := m.prefixes[prefix]; ok {
			candidateRules = append(candidateRules, rules...)
		}
	}

	// 如果没有找到前缀匹配，则使用所有规则
	if len(candidateRules) == 0 {
		for _, rule := range m.rules {
			candidateRules = append(candidateRules, rule)
		}
	}

	// 执行正则匹配
	for _, rule := range candidateRules {
		// 定期检查 context 是否取消
		if err := ctx.Err(); err != nil {
			return nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("上下文已取消: %v", err))
		}

		// 如果是字面量匹配，使用字符串查找
		if rule.literal {
			if idx := strings.Index(content, rule.prefix); idx != -1 {
				matches = append(matches, &model.RuleMatch{
					Rule:       rule.rule,
					MatchedStr: rule.prefix,
					Position:   idx,
					Score:      1.0,
				})
			}
			continue
		}

		// 执行正则匹配
		if match := rule.regex.FindStringIndex(content); match != nil {
			matches = append(matches, &model.RuleMatch{
				Rule:       rule.rule,
				MatchedStr: content[match[0]:match[1]],
				Position:   match[0],
				Score:      1.0,
			})
		}
	}

	return matches, nil
}

// Clear 清空正则匹配器
func (m *RegexMatcher) Clear() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.rules = make(map[int64]*RegexRule)
	m.prefixes = make(map[string][]*RegexRule)
	return nil
}
