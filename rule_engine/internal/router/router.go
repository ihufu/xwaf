package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/internal/handler"
	"github.com/xwaf/rule_engine/internal/middleware"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// RouterConfig 路由配置
type RouterConfig struct {
	RuleHandler    *handler.RuleHandler
	IPHandler      *handler.IPRuleHandler
	CCHandler      *handler.CCRuleHandler
	VersionHandler *handler.RuleVersionHandler
	ConfigHandler  *handler.ConfigHandler
}

// Validate 验证路由配置
func (c *RouterConfig) Validate() error {
	if c.RuleHandler == nil {
		return errors.NewError(errors.ErrConfig, "规则处理器不能为空")
	}
	if c.IPHandler == nil {
		return errors.NewError(errors.ErrConfig, "IP规则处理器不能为空")
	}
	if c.CCHandler == nil {
		return errors.NewError(errors.ErrConfig, "CC规则处理器不能为空")
	}
	if c.VersionHandler == nil {
		return errors.NewError(errors.ErrConfig, "版本处理器不能为空")
	}
	if c.ConfigHandler == nil {
		return errors.NewError(errors.ErrConfig, "配置处理器不能为空")
	}
	return nil
}

// SetupRouter 设置路由
// 配置所有API路由，添加必要的中间件，并进行参数验证
func SetupRouter(cfg *RouterConfig) (*gin.Engine, error) {
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, errors.NewError(errors.ErrConfig, fmt.Sprintf("路由配置无效: %v", err))
	}

	// 创建路由引擎
	r := gin.New()

	// 基础中间件
	r.Use(middleware.Cors())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.ErrorHandler())

	// API路由组
	api := r.Group("/api/v1")
	{
		// 规则相关路由
		rules := api.Group("/rules")
		{
			rules.POST("", cfg.RuleHandler.CreateRule)
			rules.PUT("/:id", validateIDParam(), cfg.RuleHandler.UpdateRule)
			rules.DELETE("/:id", validateIDParam(), cfg.RuleHandler.DeleteRule)
			rules.GET("/:id", validateIDParam(), cfg.RuleHandler.GetRule)
			rules.GET("", cfg.RuleHandler.ListRules)
			rules.POST("/reload", cfg.RuleHandler.ReloadRules)
			rules.GET("/version", cfg.RuleHandler.GetRuleVersion)
			rules.GET("/events", cfg.RuleHandler.GetRuleUpdateEvent)

			// 规则版本相关路由
			versions := rules.Group("/:rule_id/versions")
			versions.Use(validateIDParam("rule_id"))
			{
				versions.POST("", cfg.VersionHandler.CreateVersion)
				versions.GET("/:version", validateVersionParam(), cfg.VersionHandler.GetVersion)
				versions.GET("", cfg.VersionHandler.ListVersions)
			}

			// 规则同步日志相关路由
			syncLogs := rules.Group("/:rule_id/sync-logs")
			syncLogs.Use(validateIDParam("rule_id"))
			{
				syncLogs.GET("", cfg.VersionHandler.GetSyncLogs)
			}
		}

		// IP规则相关路由
		ips := api.Group("/ips")
		{
			ips.POST("", cfg.IPHandler.CreateIPRule)
			ips.PUT("/:id", validateIDParam(), cfg.IPHandler.UpdateIPRule)
			ips.DELETE("/:id", validateIDParam(), cfg.IPHandler.DeleteIPRule)
			ips.GET("/:id", validateIDParam(), cfg.IPHandler.GetIPRule)
			ips.GET("", cfg.IPHandler.ListIPRules)
		}

		// CC防护规则相关路由
		cc := api.Group("/cc-rules")
		{
			cc.POST("", cfg.CCHandler.CreateCCRule)
			cc.PUT("/:id", validateIDParam(), cfg.CCHandler.UpdateCCRule)
			cc.DELETE("/:id", validateIDParam(), cfg.CCHandler.DeleteCCRule)
			cc.GET("/:id", validateIDParam(), cfg.CCHandler.GetCCRule)
			cc.GET("", cfg.CCHandler.ListCCRules)
			cc.GET("/check/:uri", validateURIParam(), cfg.CCHandler.CheckCCLimit)
		}

		// 配置相关路由
		configGroup := api.Group("/config")
		{
			configGroup.GET("/mode", cfg.ConfigHandler.GetMode)
			configGroup.PUT("/mode", cfg.ConfigHandler.UpdateMode)
			configGroup.GET("/mode/logs", cfg.ConfigHandler.GetModeChangeLogs)
		}
	}

	logger.Infof("路由设置完成")
	return r, nil
}

// validateIDParam 验证ID参数中间件
func validateIDParam(paramName ...string) gin.HandlerFunc {
	name := "id"
	if len(paramName) > 0 {
		name = paramName[0]
	}
	return func(c *gin.Context) {
		id := c.Param(name)
		if id == "" {
			err := errors.NewError(errors.ErrInvalidParams, fmt.Sprintf("缺少%s参数", name))
			c.Error(err)
			c.Abort()
			return
		}
		c.Next()
	}
}

// validateVersionParam 验证版本参数中间件
func validateVersionParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := c.Param("version")
		if version == "" {
			err := errors.NewError(errors.ErrInvalidParams, "缺少version参数")
			c.Error(err)
			c.Abort()
			return
		}
		c.Next()
	}
}

// validateURIParam 验证URI参数中间件
func validateURIParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := c.Param("uri")
		if uri == "" {
			err := errors.NewError(errors.ErrInvalidParams, "缺少uri参数")
			c.Error(err)
			c.Abort()
			return
		}
		c.Next()
	}
}
