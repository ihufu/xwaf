package config

import (
	"fmt"
	"os"

	"github.com/xwaf/rule_engine/internal/server"
	"github.com/xwaf/rule_engine/pkg/logger"
	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Server *server.Config    `yaml:"server"`
	MySQL  *MySQLConfig      `yaml:"mysql"`
	Redis  *RedisConfig      `yaml:"redis"`
	Log    *logger.LogConfig `yaml:"log"`
	Rule   *RuleConfig       `yaml:"rule"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Charset         string `yaml:"charset"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime int    `yaml:"conn_max_idle_time"`
}

// RuleConfig 规则配置
type RuleConfig struct {
	SyncInterval         int `yaml:"sync_interval"`
	CacheTTL             int `yaml:"cache_ttl"`
	VersionCheckInterval int `yaml:"version_check_interval"`
}

// LoadConfig 加载配置
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	return &cfg, nil
}

// validateConfig 验证配置
func validateConfig(cfg *Config) error {
	// 验证服务器配置
	if cfg.Server == nil {
		return fmt.Errorf("服务器配置不能为空")
	}
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("无效的服务器端口: %d", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout <= 0 {
		return fmt.Errorf("无效的读取超时时间: %d", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout <= 0 {
		return fmt.Errorf("无效的写入超时时间: %d", cfg.Server.WriteTimeout)
	}

	// 验证MySQL配置
	if cfg.MySQL == nil {
		return fmt.Errorf("MySQL配置不能为空")
	}
	if cfg.MySQL.Port <= 0 || cfg.MySQL.Port > 65535 {
		return fmt.Errorf("无效的MySQL端口: %d", cfg.MySQL.Port)
	}
	if cfg.MySQL.MaxIdleConns <= 0 {
		return fmt.Errorf("无效的最大空闲连接数: %d", cfg.MySQL.MaxIdleConns)
	}
	if cfg.MySQL.MaxOpenConns <= 0 {
		return fmt.Errorf("无效的最大打开连接数: %d", cfg.MySQL.MaxOpenConns)
	}
	if cfg.MySQL.ConnMaxLifetime <= 0 {
		return fmt.Errorf("无效的连接最大生命周期: %d", cfg.MySQL.ConnMaxLifetime)
	}

	// 验证日志配置
	if cfg.Log == nil {
		return fmt.Errorf("日志配置不能为空")
	}
	if cfg.Log.MaxSize <= 0 {
		return fmt.Errorf("无效的日志文件最大大小: %d", cfg.Log.MaxSize)
	}
	if cfg.Log.MaxAge <= 0 {
		return fmt.Errorf("无效的日志文件最大保留天数: %d", cfg.Log.MaxAge)
	}
	if cfg.Log.MaxBackups <= 0 {
		return fmt.Errorf("无效的日志文件最大备份数: %d", cfg.Log.MaxBackups)
	}

	// 验证规则配置
	if cfg.Rule == nil {
		return fmt.Errorf("规则配置不能为空")
	}
	if cfg.Rule.SyncInterval <= 0 {
		return fmt.Errorf("无效的规则同步间隔: %d", cfg.Rule.SyncInterval)
	}
	if cfg.Rule.CacheTTL <= 0 {
		return fmt.Errorf("无效的规则缓存时间: %d", cfg.Rule.CacheTTL)
	}
	if cfg.Rule.VersionCheckInterval <= 0 {
		return fmt.Errorf("无效的规则版本检查间隔: %d", cfg.Rule.VersionCheckInterval)
	}

	return nil
}
