package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xwaf/rule_engine/internal/handler"
	"github.com/xwaf/rule_engine/pkg/metrics"
)

// SetupRouter 设置路由
func SetupRouter(ruleHandler *handler.RuleHandler, ipHandler *handler.IPRuleHandler, ccHandler *handler.CCRuleHandler, versionHandler *handler.RuleVersionHandler) *gin.Engine {
	r := gin.Default()

	// 添加中间件
	r.Use(metrics.APIMetricsMiddleware)

	// API路由组
	api := r.Group("/api/v1")
	{
		// 规则管理
		rules := api.Group("/rules")
		{
			rules.POST("", ruleHandler.CreateRule)
			rules.PUT("/:id", ruleHandler.UpdateRule)
			rules.DELETE("/:id", ruleHandler.DeleteRule)
			rules.GET("/:id", ruleHandler.GetRule)
			rules.GET("", ruleHandler.ListRules)
		}

		// IP规则管理
		ip := api.Group("/ip")
		{
			ip.POST("", ipHandler.CreateIPRule)
			ip.PUT("/:id", ipHandler.UpdateIPRule)
			ip.DELETE("/:id", ipHandler.DeleteIPRule)
			ip.GET("/:id", ipHandler.GetIPRule)
			ip.GET("", ipHandler.ListIPRules)
			ip.POST("/check", ipHandler.CheckIP)
		}

		// CC规则管理
		cc := api.Group("/cc")
		{
			cc.POST("", ccHandler.CreateCCRule)
			cc.PUT("/:id", ccHandler.UpdateCCRule)
			cc.DELETE("/:id", ccHandler.DeleteCCRule)
			cc.GET("/:id", ccHandler.GetCCRule)
			cc.GET("", ccHandler.ListCCRules)
			cc.POST("/check", ccHandler.CheckCC)
		}

		// 版本管理
		version := api.Group("/versions")
		{
			version.GET("/:rule_id", versionHandler.ListVersions)
			version.GET("/:rule_id/:version", versionHandler.GetVersion)
			version.GET("/:rule_id/logs", versionHandler.GetSyncLogs)
		}
	}

	// 监控指标
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r
}
