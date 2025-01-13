package model

import (
	"fmt"
	"time"
)

// WAFMode WAF 运行模式
type WAFMode string

const (
	WAFModeBlock   WAFMode = "block"  // 阻断模式
	WAFModeAlert   WAFMode = "alert"  // 监控模式
	WAFModeBypass  WAFMode = "bypass" // 旁路模式
	WAFModeMonitor WAFMode = "monitor"
)

// Validate 验证 WAF 模式
func (m WAFMode) Validate() error {
	switch m {
	case WAFModeBlock, WAFModeAlert, WAFModeBypass:
		return nil
	default:
		return fmt.Errorf("invalid WAF mode: %s", m)
	}
}

// Config 配置
type Config struct {
	MySQL      MySQLConfig      `yaml:"mysql" json:"mysql" toml:"mysql"`
	Redis      RedisConfig      `yaml:"redis" json:"redis" toml:"redis"`
	Server     ServerConfig     `yaml:"server" json:"server" toml:"server"`
	Log        LogConfig        `yaml:"log" json:"log" toml:"log"`
	WAF        WAFConfig        `yaml:"waf" json:"waf" toml:"waf"`
	RuleEngine RuleEngineConfig `yaml:"rule_engine" json:"rule_engine" toml:"rule_engine"`
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Host     string `yaml:"host" json:"host" toml:"host"`
	Port     int    `yaml:"port" json:"port" toml:"port"`
	User     string `yaml:"user" json:"user" toml:"user"`
	Password string `yaml:"password" json:"password" toml:"password"`
	Database string `yaml:"database" json:"database" toml:"database"`
	Charset  string `yaml:"charset" json:"charset" toml:"charset"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `yaml:"host" json:"host" toml:"host"`
	Port     int    `yaml:"port" json:"port" toml:"port"`
	Password string `yaml:"password" json:"password" toml:"password"`
	DB       int    `yaml:"db" json:"db" toml:"db"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port int    `yaml:"port" json:"port" toml:"port"`
	Mode string `yaml:"mode" json:"mode" toml:"mode"`
}

// LogConfig 日志配置
type LogConfig struct {
	Filename   string `yaml:"filename" json:"filename" toml:"filename"`
	Level      string `yaml:"level" json:"level" toml:"level"`
	MaxSize    int    `yaml:"max_size" json:"max_size" toml:"max_size"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups" toml:"max_backups"`
	MaxAge     int    `yaml:"max_age" json:"max_age" toml:"max_age"`
	Compress   bool   `yaml:"compress" json:"compress" toml:"compress"`
}

// WAFConfig WAF 配置
type WAFConfig struct {
	ID          int64     `json:"id" db:"id"`
	Mode        WAFMode   `json:"mode" db:"mode"`
	Description string    `json:"description" db:"description"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	UpdatedBy   string    `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Validate 验证配置
func (c *WAFConfig) Validate() error {
	if c.Mode != WAFModeBlock && c.Mode != WAFModeMonitor {
		return fmt.Errorf("无效的WAF模式: %s", c.Mode)
	}
	return nil
}

// RuleEngineConfig 规则引擎配置
type RuleEngineConfig struct {
	SyncInterval int `yaml:"sync_interval" json:"sync_interval" toml:"sync_interval"`
}
