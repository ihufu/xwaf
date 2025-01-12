package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// CacheHits 缓存命中计数器
	CacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rule_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"}, // local 或 redis
	)

	// CacheMisses 缓存未命中计数器
	CacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rule_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	// CacheLatency 缓存操作延迟直方图
	CacheLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rule_cache_operation_seconds",
			Help:    "Latency of cache operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "cache_type"},
	)
)

func init() {
	// 注册指标
	prometheus.MustRegister(CacheHits)
	prometheus.MustRegister(CacheMisses)
	prometheus.MustRegister(CacheLatency)
}
