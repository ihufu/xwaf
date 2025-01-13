package model

import (
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/xwaf/rule_engine/internal/errors"
)

// BypassMode 旁路模式类型
type BypassMode string

const (
	BypassModeNone     BypassMode = "none"     // 不启用旁路
	BypassModeMonitor  BypassMode = "monitor"  // 监控模式
	BypassModePartial  BypassMode = "partial"  // 部分旁路
	BypassModeComplete BypassMode = "complete" // 完全旁路
)

// BypassConfig 旁路配置
type BypassConfig struct {
	Mode      BypassMode `json:"mode"`       // 旁路模式
	IPs       []string   `json:"ips"`        // 允许旁路的IP列表
	URLs      []string   `json:"urls"`       // 允许旁路的URL列表
	Headers   []string   `json:"headers"`    // 允许旁路的Header列表
	StartTime int64      `json:"start_time"` // 旁路开始时间
	EndTime   int64      `json:"end_time"`   // 旁路结束时间
	Reason    string     `json:"reason"`     // 旁路原因
}

// BypassAttempt 旁路尝试记录
type BypassAttempt struct {
	ID        uint64     `json:"id" gorm:"primaryKey"`
	IP        string     `json:"ip"`        // 来源IP
	URL       string     `json:"url"`       // 请求URL
	Headers   string     `json:"headers"`   // 请求头
	Mode      BypassMode `json:"mode"`      // 尝试的旁路模式
	Timestamp int64      `json:"timestamp"` // 尝试时间
	Success   bool       `json:"success"`   // 是否成功
	Reason    string     `json:"reason"`    // 原因说明
}

// ValidateBypassConfig 验证旁路配置
func ValidateBypassConfig(config *BypassConfig) error {
	if config == nil {
		return errors.NewError(errors.ErrValidation, "旁路配置不能为空")
	}

	// 验证旁路模式
	validModes := map[BypassMode]bool{
		BypassModeNone:     true,
		BypassModeMonitor:  true,
		BypassModePartial:  true,
		BypassModeComplete: true,
	}
	if !validModes[config.Mode] {
		return errors.NewError(errors.ErrValidation, fmt.Sprintf("无效的旁路模式: %s", config.Mode))
	}

	// 验证时间范围
	if config.StartTime > 0 && config.EndTime > 0 && config.StartTime >= config.EndTime {
		return errors.NewError(errors.ErrValidation, "旁路开始时间必须早于结束时间")
	}

	// 验证IP列表
	for _, ip := range config.IPs {
		if net.ParseIP(ip) == nil {
			return errors.NewError(errors.ErrValidation, fmt.Sprintf("无效的IP地址: %s", ip))
		}
	}

	// 验证URL列表
	for _, url := range config.URLs {
		if _, err := regexp.Compile(url); err != nil {
			return errors.NewError(errors.ErrValidation, fmt.Sprintf("无效的URL模式: %s, 错误: %v", url, err))
		}
	}

	// 验证Header列表
	for _, header := range config.Headers {
		if !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(header) {
			return errors.NewError(errors.ErrValidation, fmt.Sprintf("无效的Header名称: %s", header))
		}
	}

	return nil
}

// IsBypassAllowed 检查是否允许旁路
func IsBypassAllowed(config *BypassConfig, req *CheckRequest) (bool, error) {
	if config == nil {
		return false, errors.NewError(errors.ErrValidation, "旁路配置不能为空")
	}

	if req == nil {
		return false, errors.NewError(errors.ErrValidation, "请求参数不能为空")
	}

	// 检查旁路模式
	if config.Mode == BypassModeNone {
		return false, nil
	}

	// 检查时间范围
	now := time.Now().Unix()
	if config.StartTime > 0 && config.EndTime > 0 {
		if now < config.StartTime || now > config.EndTime {
			return false, nil
		}
	}

	// 完全旁路模式
	if config.Mode == BypassModeComplete {
		return true, nil
	}

	// 检查IP
	for _, ip := range config.IPs {
		if req.ClientIP == ip {
			return true, nil
		}
	}

	// 检查URL
	for _, urlPattern := range config.URLs {
		matched, err := regexp.MatchString(urlPattern, req.URI)
		if err != nil {
			return false, errors.NewError(errors.ErrValidation, fmt.Sprintf("URL模式匹配失败: %v", err))
		}
		if matched {
			return true, nil
		}
	}

	// 检查Headers
	for _, header := range config.Headers {
		if _, exists := req.Headers[header]; exists {
			return true, nil
		}
	}

	return false, nil
}
