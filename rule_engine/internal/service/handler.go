package service

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/xwaf/rule_engine/internal/model"
)

// defaultRuleFactory 默认规则工厂实现
type defaultRuleFactory struct {
	handlers map[model.RuleType]RuleHandler
}

// NewDefaultRuleFactory 创建默认规则工厂
func NewDefaultRuleFactory() RuleFactory {
	factory := &defaultRuleFactory{
		handlers: make(map[model.RuleType]RuleHandler),
	}

	// 注册规则处理器
	factory.handlers[model.RuleTypeIP] = &ipRuleHandler{}
	factory.handlers[model.RuleTypeCC] = &ccRuleHandler{}
	factory.handlers[model.RuleTypeRegex] = &regexRuleHandler{}
	factory.handlers[model.RuleTypeSQLi] = &sqlInjectionRuleHandler{}
	factory.handlers[model.RuleTypeXSS] = &xssRuleHandler{}

	return factory
}

// CreateRuleHandler 创建规则处理器
func (f *defaultRuleFactory) CreateRuleHandler(ruleType model.RuleType) (RuleHandler, error) {
	handler, ok := f.handlers[ruleType]
	if !ok {
		return nil, fmt.Errorf("unsupported rule type: %s", ruleType)
	}
	return handler, nil
}

// ipRuleHandler IP规则处理器
type ipRuleHandler struct {
	regexCache sync.Map // 用于缓存编译后的正则表达式
}

func (h *ipRuleHandler) Match(ctx context.Context, rule *model.Rule, req *model.CheckRequest) (bool, error) {
	// 从缓存中获取正则表达式
	cached, ok := h.regexCache.Load(rule.ID)
	var re *regexp.Regexp
	var err error

	if !ok {
		// 如果缓存中没有，则编译并存储
		re, err = regexp.Compile(rule.Pattern)
		if err != nil {
			return false, fmt.Errorf("编译IP规则正则表达式失败: %v", err)
		}
		h.regexCache.Store(rule.ID, re)
	} else {
		re = cached.(*regexp.Regexp)
	}

	return re.MatchString(req.ClientIP), nil
}

// ccRuleHandler CC规则处理器
type ccRuleHandler struct {
	rdb redis.UniversalClient // Redis客户端
}

// NewCCRuleHandler 创建CC规则处理器
func NewCCRuleHandler(rdb redis.UniversalClient) *ccRuleHandler {
	return &ccRuleHandler{
		rdb: rdb,
	}
}

func (h *ccRuleHandler) Match(ctx context.Context, rule *model.Rule, req *model.CheckRequest) (bool, error) {
	// 解析规则参数
	var params struct {
		Window  int64 `json:"window"`  // 时间窗口（秒）
		MaxReqs int64 `json:"maxReqs"` // 最大请求数
	}

	if err := json.Unmarshal([]byte(rule.Params), &params); err != nil {
		return false, fmt.Errorf("解析CC规则参数失败: %v", err)
	}

	// 验证参数
	if params.Window <= 0 || params.MaxReqs <= 0 {
		return false, fmt.Errorf("无效的CC规则参数: window=%d, maxReqs=%d", params.Window, params.MaxReqs)
	}

	// 构造Redis键
	key := fmt.Sprintf("cc:%d:%s", rule.ID, req.ClientIP)

	// 使用Redis的MULTI/EXEC保证原子性
	pipe := h.rdb.Pipeline()

	// 增加计数器
	incr := pipe.Incr(ctx, key)
	// 设置过期时间
	pipe.Expire(ctx, key, time.Duration(params.Window)*time.Second)

	// 执行命令
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("redis操作失败: %v", err)
	}

	// 获取当前计数
	count := incr.Val()

	// 判断是否超过限制
	return count > params.MaxReqs, nil
}

// regexRuleHandler 正则规则处理器
type regexRuleHandler struct {
	regexCache sync.Map // 用于缓存编译后的正则表达式
}

func (h *regexRuleHandler) Match(ctx context.Context, rule *model.Rule, req *model.CheckRequest) (bool, error) {
	// 从缓存中获取正则表达式
	cached, ok := h.regexCache.Load(rule.ID)
	var re *regexp.Regexp
	var err error

	if !ok {
		// 如果缓存中没有，则编译并存储
		re, err = regexp.Compile(rule.Pattern)
		if err != nil {
			return false, fmt.Errorf("编译正则表达式失败: %v", err)
		}
		h.regexCache.Store(rule.ID, re)
	} else {
		re = cached.(*regexp.Regexp)
	}

	// 根据规则变量类型检查不同的请求部分
	switch rule.RuleVariable {
	case model.RuleVarRequestURI:
		return re.MatchString(req.URI), nil
	case model.RuleVarRequestHeaders:
		for _, v := range req.Headers {
			if re.MatchString(v) {
				return true, nil
			}
		}
	case model.RuleVarRequestArgs:
		for _, v := range req.Args {
			if re.MatchString(v) {
				return true, nil
			}
		}
	case model.RuleVarRequestBody:
		return re.MatchString(req.Body), nil
	}

	return false, nil
}

// sqlInjectionRuleHandler SQL注入规则处理器
type sqlInjectionRuleHandler struct{}

func (h *sqlInjectionRuleHandler) Match(ctx context.Context, rule *model.Rule, req *model.CheckRequest) (bool, error) {
	detector := model.NewSQLInjectionDetector()

	switch rule.RuleVariable {
	case model.RuleVarRequestURI:
		if isInjection, _ := detector.DetectInjection(req.URI); isInjection {
			return true, nil
		}
	case model.RuleVarRequestArgs:
		for _, v := range req.Args {
			if isInjection, _ := detector.DetectInjection(v); isInjection {
				return true, nil
			}
		}
	case model.RuleVarRequestBody:
		if isInjection, _ := detector.DetectInjection(req.Body); isInjection {
			return true, nil
		}
	}

	return false, nil
}

// xssRuleHandler XSS规则处理器
type xssRuleHandler struct{}

func (h *xssRuleHandler) Match(ctx context.Context, rule *model.Rule, req *model.CheckRequest) (bool, error) {
	// 根据规则变量类型检查不同的请求部分
	switch rule.RuleVariable {
	case model.RuleVarRequestURI:
		if containsXSS(req.URI) {
			return true, nil
		}
	case model.RuleVarRequestArgs:
		for _, v := range req.Args {
			if containsXSS(v) {
				return true, nil
			}
		}
	case model.RuleVarRequestBody:
		if containsXSS(req.Body) {
			return true, nil
		}
	}

	return false, nil
}

// containsXSS 检查是否包含XSS攻击
func containsXSS(input string) bool {
	patterns := []string{
		`(?i)<script[^>]*>.*?</script>`,
		`(?i)<[^>]*\b(on\w+|style|javascript:)`,
		`(?i)(javascript|vbscript|expression|data):\s*`,
		`(?i)<(iframe|object|embed|applet)`,
		`(?i)<\w+[^>]*\s+src\s*=`,
		`(?i)<\w+[^>]*\s+href\s*=`,
		`(?i)<\w+[^>]*\s+data\s*=`,
	}

	for _, pattern := range patterns {
		if matched, err := regexp.MatchString(pattern, input); err == nil && matched {
			return true
		}
	}
	return false
}
