package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	goredis "github.com/go-redis/redis/v8"
	"github.com/xwaf/rule_engine/internal/config"
	"github.com/xwaf/rule_engine/internal/handler"
	"github.com/xwaf/rule_engine/internal/repository"
	"github.com/xwaf/rule_engine/internal/repository/mysql"
	redisrepo "github.com/xwaf/rule_engine/internal/repository/redis"
	"github.com/xwaf/rule_engine/internal/router"
	"github.com/xwaf/rule_engine/internal/server"
	"github.com/xwaf/rule_engine/internal/service"
	"github.com/xwaf/rule_engine/pkg/logger"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "config", "configs/config.yaml", "配置文件路径")
}

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(cfg.Log); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 初始化MySQL
	if err := mysql.InitMySQL(&mysql.Config{
		Username: cfg.MySQL.Username,
		Password: cfg.MySQL.Password,
		Addr:     cfg.MySQL.Host,
		Port:     fmt.Sprintf("%d", cfg.MySQL.Port),
		Database: cfg.MySQL.Database,
	}); err != nil {
		logger.Fatal("初始化MySQL失败: %v", err)
	}

	// 获取数据库连接
	db := mysql.GetDB()
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("获取原始数据库连接失败: %v", err)
	}

	// 初始化Redis客户端
	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 初始化仓库
	ruleRepo := mysql.NewRuleRepository(db, redisClient)
	ipRepo := mysql.NewIPRuleRepository(sqlDB)
	cacheRepo := redisrepo.NewCacheRepository(redisClient)
	ccRepo := mysql.NewCCRuleRepository(sqlDB)
	versionRepo := mysql.NewRuleVersionRepository(sqlDB)

	// 初始化服务
	ruleFactory := service.NewDefaultRuleFactory()
	ruleService := service.NewRuleService(ruleRepo, ruleFactory, cacheRepo.(repository.RuleCache))
	ipService := service.NewIPRuleService(ipRepo, cacheRepo)
	ccService := service.NewCCRuleService(ccRepo, cacheRepo)
	versionService := service.NewRuleVersionService(versionRepo)
	configService := service.NewWAFConfigService(mysql.NewWAFConfigRepository(sqlDB), cacheRepo)

	// 初始化处理器
	ruleHandler := handler.NewRuleHandler(ruleService, versionService)
	ipHandler := handler.NewIPRuleHandler(ipService)
	ccHandler := handler.NewCCRuleHandler(ccService)
	versionHandler := handler.NewRuleVersionHandler(versionService)
	configHandler := handler.NewConfigHandler(configService)

	// 设置路由
	routerConfig := &router.RouterConfig{
		RuleHandler:    ruleHandler,
		IPHandler:      ipHandler,
		CCHandler:      ccHandler,
		VersionHandler: versionHandler,
		ConfigHandler:  configHandler,
	}
	r, err := router.SetupRouter(routerConfig)
	if err != nil {
		logger.Fatal("设置路由失败: %v", err)
	}

	// 创建HTTP服务器
	srv, err := server.NewServer(cfg.Server, r)
	if err != nil {
		logger.Fatal("创建HTTP服务器失败: %v", err)
	}

	// 启动HTTP服务器
	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal("启动HTTP服务器失败: %v", err)
		}
	}()

	logger.Info("服务启动成功，监听端口: %d", cfg.Server.Port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	logger.Info("正在关闭服务...")
	if err := srv.Stop(context.Background()); err != nil {
		logger.Error("关闭服务失败: %v", err)
	}
}
