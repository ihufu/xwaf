package matcher

import (
	"fmt"
	"sync"
)

// MatcherFactory 匹配器工厂实现
type MatcherFactory struct {
	matchers map[string]Matcher
	mutex    sync.RWMutex
}

// NewMatcherFactory 创建新的匹配器工厂
func NewMatcherFactory() *MatcherFactory {
	return &MatcherFactory{
		matchers: make(map[string]Matcher),
	}
}

// RegisterMatcher 注册匹配器
func (f *MatcherFactory) RegisterMatcher(matcherType string, matcher Matcher) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.matchers[matcherType] = matcher
}

// CreateMatcher 创建匹配器
func (f *MatcherFactory) CreateMatcher(matcherType string) (Matcher, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	matcher, exists := f.matchers[matcherType]
	if !exists {
		return nil, fmt.Errorf("未知的匹配器类型: %s", matcherType)
	}

	return matcher, nil
}

// CreateDefaultMatchers 创建默认的匹配器集合
func CreateDefaultMatchers() ([]Matcher, error) {
	matchers := make([]Matcher, 0)

	// 创建 Trie 匹配器
	trieMatcher := NewTrieMatcher()
	matchers = append(matchers, trieMatcher)

	// 创建正则匹配器
	regexMatcher := NewRegexMatcher()
	matchers = append(matchers, regexMatcher)

	// 创建 AC 自动机匹配器
	acMatcher := NewACMatcher()
	matchers = append(matchers, acMatcher)

	return matchers, nil
}

// CreateParallelMatcher 创建并行匹配器
func CreateParallelMatcher() (Matcher, error) {
	matchers, err := CreateDefaultMatchers()
	if err != nil {
		return nil, err
	}
	return NewParallelMatcher(matchers), nil
}
