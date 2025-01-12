package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// DBOpenConnections 当前打开的连接数
	DBOpenConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_open_connections",
			Help: "The number of established connections both in use and idle",
		},
	)

	// DBInUseConnections 正在使用的连接数
	DBInUseConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_in_use_connections",
			Help: "The number of connections currently in use",
		},
	)

	// DBIdleConnections 空闲连接数
	DBIdleConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_idle_connections",
			Help: "The number of idle connections",
		},
	)

	// DBWaitCount 等待连接的总数
	DBWaitCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_wait_count",
			Help: "The total number of connections waited for",
		},
	)

	// DBMaxIdleClosedTotal 因最大空闲时间而关闭的连接总数
	DBMaxIdleClosedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "db_max_idle_closed_total",
			Help: "The total number of connections closed due to SetMaxIdleConns",
		},
	)

	// DBMaxLifetimeClosedTotal 因最大生命周期而关闭的连接总数
	DBMaxLifetimeClosedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "db_max_lifetime_closed_total",
			Help: "The total number of connections closed due to SetConnMaxLifetime",
		},
	)

	// DBQueryDuration 查询耗时
	DBQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "The duration of database queries",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

func init() {
	// 注册指标
	prometheus.MustRegister(DBOpenConnections)
	prometheus.MustRegister(DBInUseConnections)
	prometheus.MustRegister(DBIdleConnections)
	prometheus.MustRegister(DBWaitCount)
	prometheus.MustRegister(DBMaxIdleClosedTotal)
	prometheus.MustRegister(DBMaxLifetimeClosedTotal)
	prometheus.MustRegister(DBQueryDuration)
}
