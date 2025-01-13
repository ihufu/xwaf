package matcher

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/xwaf/rule_engine/internal/errors"
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
	if len(matchers) == 0 {
		return &ParallelMatcher{
			matchers: make([]Matcher, 0),
			workers:  runtime.NumCPU(),
		}
	}
	return &ParallelMatcher{
		matchers: matchers,
		workers:  runtime.NumCPU(), // 使用CPU核心数作为工作协程数
	}
}

// Add 添加规则到所有匹配器
func (p *ParallelMatcher) Add(rule *model.Rule) error {
	if rule == nil {
		return errors.NewError(errors.ErrRuleMatch, "规则不能为空")
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, matcher := range p.matchers {
		if err := matcher.Add(rule); err != nil {
			return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("添加规则到匹配器失败: %v", err))
		}
	}
	return nil
}

// Remove 从所有匹配器中移除规则
func (p *ParallelMatcher) Remove(ruleID int64) error {
	if ruleID <= 0 {
		return errors.NewError(errors.ErrRuleMatch, "无效的规则ID")
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, matcher := range p.matchers {
		if err := matcher.Remove(ruleID); err != nil {
			return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("从匹配器移除规则失败: %v", err))
		}
	}
	return nil
}

// Match 并行执行规则匹配
func (p *ParallelMatcher) Match(ctx context.Context, req *model.CheckRequest) ([]*model.RuleMatch, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("上下文已取消: %v", err))
	}

	if req == nil {
		return nil, errors.NewError(errors.ErrRuleMatch, "请求参数不能为空")
	}

	if req.URI == "" {
		return nil, errors.NewError(errors.ErrRuleMatch, "请求URI不能为空")
	}

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if len(p.matchers) == 0 {
		return nil, errors.NewError(errors.ErrRuleMatch, "没有可用的匹配器")
	}

	// 创建工作任务通道
	tasks := make(chan Matcher, len(p.matchers))
	results := make(chan []*model.RuleMatch, len(p.matchers))
	errs := make(chan error, len(p.matchers))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for matcher := range tasks {
				matches, err := matcher.Match(ctx, req)
				if err != nil {
					errs <- err
					return
				}
				results <- matches
			}
		}()
	}

	// 分发任务
	for _, matcher := range p.matchers {
		select {
		case <-ctx.Done():
			return nil, errors.NewError(errors.ErrRuleMatch, "上下文已取消")
		case tasks <- matcher:
		}
	}
	close(tasks)

	// 等待所有工作完成
	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	// 收集结果
	var allMatches []*model.RuleMatch
	for matches := range results {
		allMatches = append(allMatches, matches...)
	}

	// 检查错误
	for err := range errs {
		if err != nil {
			return nil, errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("匹配器执行失败: %v", err))
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
			return errors.NewError(errors.ErrRuleMatch, fmt.Sprintf("清空匹配器失败: %v", err))
		}
	}
	return nil
}

// SetWorkers 设置工作协程数
func (p *ParallelMatcher) SetWorkers(workers int) error {
	if workers <= 0 {
		return errors.NewError(errors.ErrRuleMatch, "工作协程数必须大于0")
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.workers = workers
	return nil
}
