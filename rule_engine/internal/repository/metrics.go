package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/xwaf/rule_engine/internal/errors"
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
	// 如果记录失败会记录错误日志
	RecordRuleMatch(ctx context.Context, ruleID, ruleName string, action model.ActionType)

	// RecordAPIRequest 记录API请求
	// 如果记录失败会记录错误日志
	RecordAPIRequest(ctx context.Context, path, method, status string)

	// RecordAPIResponseTime 记录API响应时间
	// 如果记录失败会记录错误日志
	RecordAPIResponseTime(ctx context.Context, path, method string, duration time.Duration)

	// GetRuleMatchMetrics 获取规则匹配统计
	// 返回 ErrSystem - 查询Prometheus失败
	GetRuleMatchMetrics(ctx context.Context) ([]*model.RuleMatchMetrics, error)

	// GetAPIMetrics 获取API性能统计
	// 返回 ErrSystem - 查询Prometheus失败
	GetAPIMetrics(ctx context.Context) ([]*model.APIMetrics, error)
}

// PrometheusMetricsRepository Prometheus监控指标仓储实现
type PrometheusMetricsRepository struct {
	prometheusAddr string
}

// NewPrometheusMetricsRepository 创建Prometheus监控指标仓储
func NewPrometheusMetricsRepository(prometheusAddr string) MetricsRepository {
	return &PrometheusMetricsRepository{prometheusAddr}
}

// RecordRuleMatch 记录规则匹配
func (r *PrometheusMetricsRepository) RecordRuleMatch(ctx context.Context, ruleID, ruleName string, action model.ActionType) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("记录规则匹配指标失败: %v\n", err)
		}
	}()
	ruleMatchCounter.WithLabelValues(ruleID, ruleName, string(action)).Inc()
}

// RecordAPIRequest 记录API请求
func (r *PrometheusMetricsRepository) RecordAPIRequest(ctx context.Context, path, method, status string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("记录API请求指标失败: %v\n", err)
		}
	}()
	apiRequestCounter.WithLabelValues(path, method, status).Inc()
}

// RecordAPIResponseTime 记录API响应时间
func (r *PrometheusMetricsRepository) RecordAPIResponseTime(ctx context.Context, path, method string, duration time.Duration) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("记录API响应时间指标失败: %v\n", err)
		}
	}()
	apiResponseTimeHistogram.WithLabelValues(path, method).Observe(duration.Seconds())
}

// GetRuleMatchMetrics 获取规则匹配统计
func (r *PrometheusMetricsRepository) GetRuleMatchMetrics(ctx context.Context) ([]*model.RuleMatchMetrics, error) {
	// 从Prometheus查询规则匹配统计数据
	metrics := make([]*model.RuleMatchMetrics, 0)

	// 查询规则匹配总数
	query := `sum(waf_rule_matches_total) by (rule_id, rule_name, action)`
	result, err := r.queryPrometheus(ctx, query)
	if err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("查询规则匹配统计失败: %v", err))
	}

	// 解析查询结果
	for _, m := range result {
		metric := &model.RuleMatchMetrics{
			RuleID:    m.Labels["rule_id"],
			RuleName:  m.Labels["rule_name"],
			TotalHits: int64(m.Value),
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetAPIMetrics 获取API性能统计
func (r *PrometheusMetricsRepository) GetAPIMetrics(ctx context.Context) ([]*model.APIMetrics, error) {
	// 从Prometheus查询API性能统计数据
	metrics := make([]*model.APIMetrics, 0)

	// 查询API请求总数
	query := `sum(waf_api_requests_total) by (path, method)`
	result, err := r.queryPrometheus(ctx, query)
	if err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("查询API请求统计失败: %v", err))
	}

	// 查询API响应时间
	timeQuery := `histogram_quantile(0.95, sum(rate(waf_api_response_time_seconds_bucket[5m])) by (path, method, le))`
	timeResult, err := r.queryPrometheus(ctx, timeQuery)
	if err != nil {
		return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("查询API响应时间统计失败: %v", err))
	}

	// 解析查询结果
	for _, m := range result {
		metric := &model.APIMetrics{
			Path:          m.Labels["path"],
			Method:        m.Labels["method"],
			TotalRequests: int64(m.Value),
		}
		metrics = append(metrics, metric)
	}

	// 更新响应时间
	for _, m := range timeResult {
		for _, metric := range metrics {
			if metric.Path == m.Labels["path"] && metric.Method == m.Labels["method"] {
				metric.AvgResponseTime = m.Value
				break
			}
		}
	}

	return metrics, nil
}

// PrometheusMetric Prometheus指标数据
type PrometheusMetric struct {
	Labels map[string]string
	Value  float64
}

// queryPrometheus 查询Prometheus数据
func (r *PrometheusMetricsRepository) queryPrometheus(ctx context.Context, query string) ([]PrometheusMetric, error) {
	select {
	case <-ctx.Done():
		return nil, errors.NewError(errors.ErrSystem, "查询已取消")
	default:
		// 使用 Prometheus HTTP API 查询
		u := fmt.Sprintf("%s/api/v1/query", r.prometheusAddr)
		req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
		if err != nil {
			return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("创建请求失败: %v", err))
		}

		// 添加查询参数
		q := req.URL.Query()
		q.Add("query", query)
		req.URL.RawQuery = q.Encode()

		// 发送请求
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("发送请求失败: %v", err))
		}
		defer resp.Body.Close()

		// 解析响应
		var result struct {
			Status string `json:"status"`
			Data   struct {
				ResultType string `json:"resultType"`
				Result     []struct {
					Metric map[string]string `json:"metric"`
					Value  []interface{}     `json:"value"`
				} `json:"result"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, errors.NewError(errors.ErrSystem, fmt.Sprintf("解析响应失败: %v", err))
		}

		// 检查响应状态
		if result.Status != "success" {
			return nil, errors.NewError(errors.ErrSystem, "查询失败")
		}

		// 转换结果
		metrics := make([]PrometheusMetric, 0, len(result.Data.Result))
		for _, r := range result.Data.Result {
			// Prometheus 返回的 value 是 [timestamp, value] 格式
			if len(r.Value) != 2 {
				continue
			}

			// 转换值
			value, err := strconv.ParseFloat(r.Value[1].(string), 64)
			if err != nil {
				continue
			}

			metrics = append(metrics, PrometheusMetric{
				Labels: r.Metric,
				Value:  value,
			})
		}

		return metrics, nil
	}
}
