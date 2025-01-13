package model

import (
	"fmt"
)

// WAFMode WAF运行模式
type WAFMode string

const (
	WAFModeBlock  WAFMode = "block"  // 阻断模式
	WAFModeLog    WAFMode = "log"    // 日志模式
	WAFModeBypass WAFMode = "bypass" // 旁路模式
)

// Validate 验证 WAF 模式
func (m WAFMode) Validate() error {
	switch m {
	case WAFModeBlock, WAFModeLog, WAFModeBypass:
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

// WAFConfig WAF配置
type WAFConfig struct {
	ID          int64   `json:"id" gorm:"primary_key"`                 // 配置ID
	Mode        WAFMode `json:"mode" gorm:"column:mode"`               // 运行模式
	UpdatedAt   int64   `json:"updated_at" gorm:"column:updated_at"`   // 更新时间
	UpdatedBy   string  `json:"updated_by" gorm:"column:updated_by"`   // 更新人
	CreatedAt   int64   `json:"created_at" gorm:"column:created_at"`   // 创建时间
	CreatedBy   string  `json:"created_by" gorm:"column:created_by"`   // 创建人
	Description string  `json:"description" gorm:"column:description"` // 描述
}

// WAFModeChangeLog WAF运行模式变更日志
type WAFModeChangeLog struct {
	ID          int64   `json:"id" gorm:"primary_key"`
	OldMode     WAFMode `json:"old_mode" gorm:"column:old_mode"`       // 原模式
	NewMode     WAFMode `json:"new_mode" gorm:"column:new_mode"`       // 新模式
	Operator    string  `json:"operator" gorm:"column:operator"`       // 操作人
	Reason      string  `json:"reason" gorm:"column:reason"`           // 变更原因
	CreatedAt   int64   `json:"created_at" gorm:"column:created_at"`   // 创建时间
	Description string  `json:"description" gorm:"column:description"` // 描述
}

// Validate 验证配置
func (c *WAFConfig) Validate() error {
	switch c.Mode {
	case WAFModeBlock, WAFModeLog, WAFModeBypass:
		return nil
	default:
		return fmt.Errorf("无效的运行模式: %s", c.Mode)
	}
}

// RuleEngineConfig 规则引擎配置
type RuleEngineConfig struct {
	SyncInterval int `yaml:"sync_interval" json:"sync_interval" toml:"sync_interval"`
}
