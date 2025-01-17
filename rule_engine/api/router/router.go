package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/handler"
	"github.com/xwaf/rule_engine/internal/middleware"
	"github.com/xwaf/rule_engine/pkg/logger"
	"github.com/xwaf/rule_engine/pkg/metrics"
)

// SetupRouter 设置路由
func SetupRouter(ruleHandler *handler.RuleHandler, ipHandler *handler.IPRuleHandler, ccHandler *handler.CCRuleHandler, versionHandler *handler.RuleVersionHandler) (*gin.Engine, error) {
	// 验证处理器
	if ruleHandler == nil {
		return nil, errors.NewError(errors.ErrInit, "规则处理器不能为空")
	}
	if ipHandler == nil {
		return nil, errors.NewError(errors.ErrInit, "IP规则处理器不能为空")
	}
	if ccHandler == nil {
		return nil, errors.NewError(errors.ErrInit, "CC规则处理器不能为空")
	}
	if versionHandler == nil {
		return nil, errors.NewError(errors.ErrInit, "版本处理器不能为空")
	}

	r := gin.Default()

	// 添加中间件
	r.Use(middleware.RequestID()) // 生成请求ID
	r.Use(middleware.Logger())    // 请求日志
	r.Use(middleware.Recovery())  // 错误恢复
	r.Use(metrics.APIMetricsMiddleware)

	logger.Info("初始化路由...")

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

	logger.Info("路由初始化完成")
	return r, nil
}
