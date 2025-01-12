package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "HTTP request duration in seconds",
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

// APIMetricsMiddleware Gin中间件，用于收集API指标
func APIMetricsMiddleware(c *gin.Context) {
	start := time.Now()
	c.Next()

	status := strconv.Itoa(c.Writer.Status())
	duration := time.Since(start).Seconds()

	httpRequestsTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, status).Inc()
	httpRequestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path).Observe(duration)
}
