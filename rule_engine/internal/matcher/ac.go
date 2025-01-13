package matcher

import (
	"context"
	"sync"

	"github.com/xwaf/rule_engine/internal/model"
)

// ACNode AC自动机节点
type ACNode struct {
	children map[rune]*ACNode
	fail     *ACNode
	isEnd    bool
	rules    []*model.Rule
	depth    int
}

// ACMatcher AC自动机匹配器
type ACMatcher struct {
	root  *ACNode
	mutex sync.RWMutex
}

// NewACMatcher 创建新的AC自动机匹配器
func NewACMatcher() *ACMatcher {
	return &ACMatcher{
		root: &ACNode{
			children: make(map[rune]*ACNode),
			depth:    0,
		},
	}
}

// Add 添加规则到AC自动机
func (m *ACMatcher) Add(rule *model.Rule) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 构建字典树
	current := m.root
	for _, ch := range rule.Pattern {
		if current.children[ch] == nil {
			current.children[ch] = &ACNode{
				children: make(map[rune]*ACNode),
				depth:    current.depth + 1,
			}
		}
		current = current.children[ch]
	}
	current.isEnd = true
	current.rules = append(current.rules, rule)

	// 构建失败指针
	m.buildFailPointers()
	return nil
}

// buildFailPointers 构建失败指针
func (m *ACMatcher) buildFailPointers() {
	// 使用BFS构建失败指针
	queue := make([]*ACNode, 0)

	// 初始化第一层失败指针
	for _, child := range m.root.children {
		child.fail = m.root
		queue = append(queue, child)
	}

	// BFS构建其他层失败指针
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for ch, child := range current.children {
			queue = append(queue, child)

			// 查找失败指针
			failTo := current.fail
			for failTo != nil {
				if next := failTo.children[ch]; next != nil {
					child.fail = next
					break
				}
				failTo = failTo.fail
			}
			if failTo == nil {
				child.fail = m.root
			}
		}
	}
}

// Remove 从AC自动机中移除规则
func (m *ACMatcher) Remove(ruleID int64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 重建AC自动机
	newRoot := &ACNode{
		children: make(map[rune]*ACNode),
		depth:    0,
	}
	m.removeRule(m.root, ruleID, newRoot)
	m.root = newRoot
	m.buildFailPointers()
	return nil
}

// removeRule 递归移除规则
func (m *ACMatcher) removeRule(node *ACNode, ruleID int64, newNode *ACNode) {
	if node.isEnd {
		newRules := make([]*model.Rule, 0)
		for _, rule := range node.rules {
			if rule.ID != ruleID {
				newRules = append(newRules, rule)
			}
		}
		if len(newRules) > 0 {
			newNode.isEnd = true
			newNode.rules = newRules
		}
	}

	for ch, child := range node.children {
		if newNode.children[ch] == nil {
			newNode.children[ch] = &ACNode{
				children: make(map[rune]*ACNode),
				depth:    newNode.depth + 1,
			}
		}
		m.removeRule(child, ruleID, newNode.children[ch])
	}
}

// Match 使用AC自动机进行匹配
func (m *ACMatcher) Match(ctx context.Context, req *model.CheckRequest) ([]*model.RuleMatch, error) {
	// 检查 context 是否已取消
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	content := req.URI
	// 预分配一个合理的容量以减少内存分配
	matches := make([]*model.RuleMatch, 0, 16)
	current := m.root

	// 每处理 1000 个字符检查一次 context
	const checkInterval = 1000
	checkCounter := 0

	// AC自动机匹配
	for pos, ch := range content {
		// 定期检查 context 是否取消
		checkCounter++
		if checkCounter >= checkInterval {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			checkCounter = 0
		}

		// 查找匹配节点
		for current != nil && current.children[ch] == nil {
			current = current.fail
		}

		if current == nil {
			current = m.root
			continue
		}

		current = current.children[ch]

		// 收集所有匹配结果
		for p := current; p != nil; p = p.fail {
			if p.isEnd && len(p.rules) > 0 {
				matchedStr := content[pos-p.depth+1 : pos+1]
				// 使用临时切片存储当前节点的匹配结果
				nodeMatches := make([]*model.RuleMatch, 0, len(p.rules))

				for _, rule := range p.rules {
					nodeMatches = append(nodeMatches, &model.RuleMatch{
						Rule:       rule,
						MatchedStr: matchedStr,
						Position:   pos - p.depth + 1,
						Score:      1.0,
					})
				}

				// 一次性添加所有匹配结果
				matches = append(matches, nodeMatches...)
			}
		}
	}

	return matches, nil
}

// Clear 清空AC自动机
func (m *ACMatcher) Clear() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.root = &ACNode{
		children: make(map[rune]*ACNode),
		depth:    0,
	}
	return nil
}
