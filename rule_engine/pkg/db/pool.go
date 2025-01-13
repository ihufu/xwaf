package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/xwaf/rule_engine/pkg/metrics"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config 数据库配置
type Config struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	User         string        `yaml:"user"`
	Password     string        `yaml:"password"`
	Database     string        `yaml:"database"`
	MaxOpenConns int           `yaml:"max_open_conns"`
	MaxIdleConns int           `yaml:"max_idle_conns"`
	MaxLifetime  time.Duration `yaml:"max_lifetime"`
	MaxIdleTime  time.Duration `yaml:"max_idle_time"`
}

// Pool 数据库连接池
type Pool struct {
	db     *gorm.DB
	sqlDB  *sql.DB
	config *Config
}

// NewPool 创建数据库连接池
func NewPool(config *Config) (*Pool, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.Database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.MaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.MaxIdleTime)

	pool := &Pool{
		db:     db,
		sqlDB:  sqlDB,
		config: config,
	}

	// 启动连接池监控
	go pool.monitorPool()

	return pool, nil
}

// DB 获取数据库连接
func (p *Pool) DB() *gorm.DB {
	return p.db
}

// Close 关闭连接池
func (p *Pool) Close() error {
	return p.sqlDB.Close()
}

// Stats 获取连接池统计信息
func (p *Pool) Stats() sql.DBStats {
	return p.sqlDB.Stats()
}

// 监控连接池状态
func (p *Pool) monitorPool() {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for range ticker.C {
		stats := p.sqlDB.Stats()

		// 更新连接池指标
		metrics.DBOpenConnections.Set(float64(stats.OpenConnections))
		metrics.DBInUseConnections.Set(float64(stats.InUse))
		metrics.DBIdleConnections.Set(float64(stats.Idle))
		metrics.DBWaitCount.Set(float64(stats.WaitCount))
		metrics.DBMaxIdleClosedTotal.Add(float64(stats.MaxIdleClosed))
		metrics.DBMaxLifetimeClosedTotal.Add(float64(stats.MaxLifetimeClosed))
	}
}

// WithContext 创建带有上下文的数据库会话
func (p *Pool) WithContext(ctx context.Context) *gorm.DB {
	return p.db.WithContext(ctx)
}

// Transaction 执行事务
func (p *Pool) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return p.db.WithContext(ctx).Transaction(fn)
}
