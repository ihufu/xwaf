package repository

import (
	"context"
	"time"

	"github.com/xwaf/rule_engine/internal/model"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// 规则匹配计数器
	ruleMatchCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waf_rule_matches_total",
			Help: "Total number of rule matches",
		},
		[]string{"rule_id", "rule_name", "action"},
	)

	// API请求计数器
	apiRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waf_api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"path", "method", "status"},
	)

	// API响应时间直方图
	apiResponseTimeHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "waf_api_response_time_seconds",
			Help:    "API response time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)
)

func init() {
	// 注册Prometheus指标
	prometheus.MustRegister(ruleMatchCounter)
	prometheus.MustRegister(apiRequestCounter)
	prometheus.MustRegister(apiResponseTimeHistogram)
}

// MetricsRepository 监控指标仓储接口
type MetricsRepository interface {
	// RecordRuleMatch 记录规则匹配
	RecordRuleMatch(ctx context.Context, ruleID, ruleName string, action model.ActionType)

	// RecordAPIRequest 记录API请求
	RecordAPIRequest(ctx context.Context, path, method, status string)

	// RecordAPIResponseTime 记录API响应时间
	RecordAPIResponseTime(ctx context.Context, path, method string, duration time.Duration)

	// GetRuleMatchMetrics 获取规则匹配统计
	GetRuleMatchMetrics(ctx context.Context) ([]*model.RuleMatchMetrics, error)

	// GetAPIMetrics 获取API性能统计
	GetAPIMetrics(ctx context.Context) ([]*model.APIMetrics, error)
}

// PrometheusMetricsRepository Prometheus监控指标仓储实现
type PrometheusMetricsRepository struct{}

// NewPrometheusMetricsRepository 创建Prometheus监控指标仓储
func NewPrometheusMetricsRepository() MetricsRepository {
	return &PrometheusMetricsRepository{}
}

// RecordRuleMatch 记录规则匹配
func (r *PrometheusMetricsRepository) RecordRuleMatch(ctx context.Context, ruleID, ruleName string, action model.ActionType) {
	ruleMatchCounter.WithLabelValues(ruleID, ruleName, string(action)).Inc()
}

// RecordAPIRequest 记录API请求
func (r *PrometheusMetricsRepository) RecordAPIRequest(ctx context.Context, path, method, status string) {
	apiRequestCounter.WithLabelValues(path, method, status).Inc()
}

// RecordAPIResponseTime 记录API响应时间
func (r *PrometheusMetricsRepository) RecordAPIResponseTime(ctx context.Context, path, method string, duration time.Duration) {
	apiResponseTimeHistogram.WithLabelValues(path, method).Observe(duration.Seconds())
}

// GetRuleMatchMetrics 获取规则匹配统计
func (r *PrometheusMetricsRepository) GetRuleMatchMetrics(ctx context.Context) ([]*model.RuleMatchMetrics, error) {
	// 从Prometheus查询规则匹配统计数据
	metrics := make([]*model.RuleMatchMetrics, 0)

	// TODO: 实现从Prometheus查询数据的逻辑

	return metrics, nil
}

// GetAPIMetrics 获取API性能统计
func (r *PrometheusMetricsRepository) GetAPIMetrics(ctx context.Context) ([]*model.APIMetrics, error) {
	// 从Prometheus查询API性能统计数据
	metrics := make([]*model.APIMetrics, 0)

	// TODO: 实现从Prometheus查询数据的逻辑

	return metrics, nil
}
