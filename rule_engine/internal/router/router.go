package router

import (
	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/handler"
	"github.com/xwaf/rule_engine/internal/middleware"
)

// SetupRouter 设置路由
func SetupRouter(ruleHandler *handler.RuleHandler, ipHandler *handler.IPRuleHandler,
	ccHandler *handler.CCRuleHandler, versionHandler *handler.RuleVersionHandler) *gin.Engine {
	r := gin.New()

	// 中间件
	r.Use(middleware.Cors())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// API路由组
	api := r.Group("/api/v1")
	{
		// 规则相关路由
		rules := api.Group("/rules")
		{
			rules.POST("", ruleHandler.CreateRule)
			rules.PUT("/:id", ruleHandler.UpdateRule)
			rules.DELETE("/:id", ruleHandler.DeleteRule)
			rules.GET("/:id", ruleHandler.GetRule)
			rules.GET("", ruleHandler.ListRules)
			rules.POST("/reload", ruleHandler.ReloadRules)
			rules.GET("/version", ruleHandler.GetRuleVersion)
			rules.GET("/events", ruleHandler.GetRuleUpdateEvent)

			// 规则版本相关路由
			versions := rules.Group("/:rule_id/versions")
			{
				versions.POST("", versionHandler.CreateVersion)
				versions.GET("/:version", versionHandler.GetVersion)
				versions.GET("", versionHandler.ListVersions)
			}

			// 规则同步日志相关路由
			syncLogs := rules.Group("/:rule_id/sync-logs")
			{
				syncLogs.GET("", versionHandler.GetSyncLogs)
			}
		}

		// IP规则相关路由
		ips := api.Group("/ips")
		{
			ips.POST("", ipHandler.CreateIPRule)
			ips.PUT("/:id", ipHandler.UpdateIPRule)
			ips.DELETE("/:id", ipHandler.DeleteIPRule)
			ips.GET("/:id", ipHandler.GetIPRule)
			ips.GET("", ipHandler.ListIPRules)
		}

		// CC防护规则相关路由
		cc := api.Group("/cc-rules")
		{
			cc.POST("", ccHandler.CreateCCRule)
			cc.PUT("/:id", ccHandler.UpdateCCRule)
			cc.DELETE("/:id", ccHandler.DeleteCCRule)
			cc.GET("/:id", ccHandler.GetCCRule)
			cc.GET("", ccHandler.ListCCRules)
			cc.GET("/check/:uri", ccHandler.CheckCCLimit)
		}
	}

	return r
}
