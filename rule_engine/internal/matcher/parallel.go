package matcher

import (
	"context"
	"runtime"
	"sync"

	"github.com/xwaf/rule_engine/internal/model"
)

// ParallelMatcher 并行匹配器
type ParallelMatcher struct {
	matchers []Matcher
	workers  int
	mutex    sync.RWMutex
}

// NewParallelMatcher 创建新的并行匹配器
func NewParallelMatcher(matchers []Matcher) *ParallelMatcher {
	return &ParallelMatcher{
		matchers: matchers,
		workers:  runtime.NumCPU(), // 使用CPU核心数作为工作协程数
	}
}

// Add 添加规则到所有匹配器
func (p *ParallelMatcher) Add(rule *model.Rule) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, matcher := range p.matchers {
		if err := matcher.Add(rule); err != nil {
			return err
		}
	}
	return nil
}

// Remove 从所有匹配器中移除规则
func (p *ParallelMatcher) Remove(ruleID int64) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, matcher := range p.matchers {
		if err := matcher.Remove(ruleID); err != nil {
			return err
		}
	}
	return nil
}

// Match 并行执行规则匹配
func (p *ParallelMatcher) Match(ctx context.Context, req *model.CheckRequest) ([]*model.RuleMatch, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// 创建工作任务通道
	tasks := make(chan Matcher, len(p.matchers))
	results := make(chan []*model.RuleMatch, len(p.matchers))
	errors := make(chan error, len(p.matchers))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for matcher := range tasks {
				matches, err := matcher.Match(ctx, req)
				if err != nil {
					errors <- err
					return
				}
				results <- matches
			}
		}()
	}

	// 分发任务
	for _, matcher := range p.matchers {
		tasks <- matcher
	}
	close(tasks)

	// 等待所有工作完成
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// 收集结果
	var allMatches []*model.RuleMatch
	for matches := range results {
		allMatches = append(allMatches, matches...)
	}

	// 检查错误
	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	// 对结果进行优先级排序
	model.SortRuleMatchesByPriority(allMatches)

	return allMatches, nil
}

// Clear 清空所有匹配器
func (p *ParallelMatcher) Clear() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, matcher := range p.matchers {
		if err := matcher.Clear(); err != nil {
			return err
		}
	}
	return nil
}

// SetWorkers 设置工作协程数
func (p *ParallelMatcher) SetWorkers(workers int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.workers = workers
}
