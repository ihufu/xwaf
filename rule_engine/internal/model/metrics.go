package model

// RuleMatchMetrics 规则匹配统计
type RuleMatchMetrics struct {
	RuleID      string `json:"rule_id"`       // 规则ID
	RuleName    string `json:"rule_name"`     // 规则名称
	TotalHits   int64  `json:"total_hits"`    // 总命中次数
	BlockCount  int64  `json:"block_count"`   // 拦截次数
	AllowCount  int64  `json:"allow_count"`   // 放行次数
	LastHitTime string `json:"last_hit_time"` // 最后命中时间
}

// CacheMetrics 缓存统计
type CacheMetrics struct {
	TotalRequests int64   `json:"total_requests"` // 总请求数
	HitCount      int64   `json:"hit_count"`      // 命中次数
	MissCount     int64   `json:"miss_count"`     // 未命中次数
	HitRate       float64 `json:"hit_rate"`       // 命中率
}

// APIMetrics API性能统计
type APIMetrics struct {
	Path            string  `json:"path"`              // API路径
	Method          string  `json:"method"`            // HTTP方法
	TotalRequests   int64   `json:"total_requests"`    // 总请求数
	AvgResponseTime float64 `json:"avg_response_time"` // 平均响应时间
	MaxResponseTime float64 `json:"max_response_time"` // 最大响应时间
	MinResponseTime float64 `json:"min_response_time"` // 最小响应时间
	ErrorCount      int64   `json:"error_count"`       // 错误次数
	ErrorRate       float64 `json:"error_rate"`        // 错误率
}
