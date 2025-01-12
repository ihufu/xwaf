package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/xwaf/rule_engine/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// 规则相关指标
	ruleMatchTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waf_rule_match_total",
			Help: "规则匹配总次数",
		},
		[]string{"rule_id", "rule_name", "rule_type", "action", "status"},
	)

	ruleMatchLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "waf_rule_match_latency_seconds",
			Help:    "规则匹配延迟",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"rule_type"},
	)

	// 规则同步指标
	ruleSyncStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "waf_rule_sync_status",
			Help: "规则同步状态(0:失败,1:成功)",
		},
		[]string{"node_id"},
	)

	ruleSyncLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "waf_rule_sync_latency_seconds",
			Help:    "规则同步延迟",
			Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"node_id"},
	)

	ruleSyncTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waf_rule_sync_total",
			Help: "规则同步总次数",
		},
		[]string{"node_id", "status"},
	)

	// 缓存监控指标
	cacheOperationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waf_cache_operation_total",
			Help: "缓存操作总次数",
		},
		[]string{"operation", "status"},
	)

	cacheHitRatio = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "waf_cache_hit_ratio",
			Help: "缓存命中率",
		},
		[]string{"cache_type"},
	)

	cacheLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "waf_cache_latency_seconds",
			Help:    "缓存操作延迟",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25},
		},
		[]string{"operation"},
	)

	// 健康检查指标
	componentHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "waf_component_health",
			Help: "组件健康状态(0:不健康,1:健康)",
		},
		[]string{"component", "node_id"},
	)

	healthCheckLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "waf_health_check_latency_seconds",
			Help:    "健康检查延迟",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"component"},
	)

	// 请求相关指标
	requestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "waf_request_total",
			Help: "请求总数",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "waf_request_duration_seconds",
			Help:    "请求处理时间",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5},
		},
		[]string{"method", "path", "status"},
	)
)

// RecordRuleMatch 记录规则匹配
func RecordRuleMatch(rule *model.Rule, matched bool, duration time.Duration) {
	status := "miss"
	if matched {
		status = "hit"
	}

	ruleMatchTotal.WithLabelValues(
		strconv.FormatInt(rule.ID, 10),
		rule.Name,
		string(rule.Type),
		string(rule.Action),
		status,
	).Inc()

	ruleMatchLatency.WithLabelValues(string(rule.Type)).Observe(duration.Seconds())
}

// RecordRuleSync 记录规则同步
func RecordRuleSync(nodeID string, success bool, duration time.Duration) {
	status := "failed"
	statusCode := float64(0)
	if success {
		status = "success"
		statusCode = 1
	}

	ruleSyncStatus.WithLabelValues(nodeID).Set(statusCode)
	ruleSyncLatency.WithLabelValues(nodeID).Observe(duration.Seconds())
	ruleSyncTotal.WithLabelValues(nodeID, status).Inc()
}

// RecordCacheOperation 记录缓存操作
func RecordCacheOperation(operation string, hit bool, duration time.Duration) {
	status := "miss"
	if hit {
		status = "hit"
	}

	cacheOperationTotal.WithLabelValues(operation, status).Inc()
	cacheLatency.WithLabelValues(operation).Observe(duration.Seconds())
}

// UpdateCacheHitRatio 更新缓存命中率
func UpdateCacheHitRatio(cacheType string, ratio float64) {
	cacheHitRatio.WithLabelValues(cacheType).Set(ratio)
}

// RecordComponentHealth 记录组件健康状态
func RecordComponentHealth(component, nodeID string, healthy bool, checkDuration time.Duration) {
	status := float64(0)
	if healthy {
		status = 1
	}

	componentHealth.WithLabelValues(component, nodeID).Set(status)
	healthCheckLatency.WithLabelValues(component).Observe(checkDuration.Seconds())
}

// MetricsHandler 返回Prometheus指标处理器
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// MetricsMiddleware 监控中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		duration := time.Since(startTime)
		status := strconv.Itoa(c.Writer.Status())

		// 记录请求指标
		requestTotal.With(prometheus.Labels{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": status,
		}).Inc()

		requestDuration.With(prometheus.Labels{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": status,
		}).Observe(duration.Seconds())
	}
}
