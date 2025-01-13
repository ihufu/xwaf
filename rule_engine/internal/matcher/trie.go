package matcher

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/model"
)

// TrieNode Trie树节点
type TrieNode struct {
	children map[string]*TrieNode
	rules    []*model.Rule
	isEnd    bool
}

// TrieMatcher 基于Trie树的URL匹配器
type TrieMatcher struct {
	root  *TrieNode
	mutex sync.RWMutex
}

// NewTrieMatcher 创建新的Trie匹配器
func NewTrieMatcher() *TrieMatcher {
	return &TrieMatcher{
		root: &TrieNode{
			children: make(map[string]*TrieNode),
		},
	}
}

// Add 添加规则到Trie树
func (t *TrieMatcher) Add(rule *model.Rule) error {
	if rule == nil {
		return errors.NewError(errors.ErrRuleMatch, "规则不能为空")
	}

	// 仅处理URL类型的规则
	if rule.RuleVariable != "request_uri" {
		return nil
	}

	if rule.Pattern == "" {
		return errors.NewError(errors.ErrRuleMatch, "规则模式不能为空")
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	// 分割URL路径
	parts := strings.Split(strings.Trim(rule.Pattern, "/"), "/")
	current := t.root

	for _, part := range parts {
		if current.children[part] == nil {
			current.children[part] = &TrieNode{
				children: make(map[string]*TrieNode),
			}
		}
		current = current.children[part]
	}

	current.isEnd = true
	current.rules = append(current.rules, rule)
	return nil
}

// Remove 从Trie树中移除规则
func (t *TrieMatcher) Remove(ruleID int64) error {
	if ruleID <= 0 {
		return errors.NewError(errors.ErrRuleMatch, "无效的规则ID")
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	var found bool
	var removeFromNode func(*TrieNode) bool
	removeFromNode = func(node *TrieNode) bool {
		if node == nil {
			return false
		}

		// 从当前节点的规则列表中移除
		if node.rules != nil {
			for i, rule := range node.rules {
				if rule.ID == ruleID {
					node.rules = append(node.rules[:i], node.rules[i+1:]...)
					if len(node.rules) == 0 {
						node.isEnd = false
					}
					found = true
					return true
				}
			}
		}

		// 递归检查子节点
		for _, child := range node.children {
			if removeFromNode(child) {
				return true
			}
		}

		return false
	}

	removeFromNode(t.root)
	if !found {
		return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("规则不存在: %d", ruleID))
	}
	return nil
}

// Match 在Trie树中匹配URL
func (t *TrieMatcher) Match(ctx context.Context, req *model.CheckRequest) ([]*model.RuleMatch, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("上下文已取消: %v", err))
	}

	if req == nil {
		return nil, errors.NewError(errors.ErrRuleMatch, "请求参数不能为空")
	}

	if req.URI == "" {
		return nil, errors.NewError(errors.ErrRuleMatch, "请求URI不能为空")
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	matches := make([]*model.RuleMatch, 0)
	parts := strings.Split(strings.Trim(req.URI, "/"), "/")

	var matchPath func(node *TrieNode, depth int)
	matchPath = func(node *TrieNode, depth int) {
		if node == nil {
			return
		}

		// 如果是终结节点，添加所有规则到匹配结果
		if node.isEnd {
			for _, rule := range node.rules {
				matches = append(matches, &model.RuleMatch{
					Rule:       rule,
					MatchedStr: req.URI,
					Position:   0,
					Score:      1.0,
				})
			}
		}

		// 如果已经到达路径末尾，返回
		if depth >= len(parts) {
			return
		}

		// 精确匹配当前路径部分
		if child, ok := node.children[parts[depth]]; ok {
			matchPath(child, depth+1)
		}

		// 通配符匹配
		if child, ok := node.children["*"]; ok {
			matchPath(child, depth+1)
		}
	}

	matchPath(t.root, 0)
	return matches, nil
}

// Clear 清空Trie树
func (t *TrieMatcher) Clear() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.root = &TrieNode{
		children: make(map[string]*TrieNode),
	}
	return nil
}
