package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xwaf/rule_engine/internal/errors"
	"github.com/xwaf/rule_engine/pkg/logger"
)

// Config HTTP服务器配置
type Config struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	ReadTimeout     int    `yaml:"read_timeout"`
	WriteTimeout    int    `yaml:"write_timeout"`
	ShutdownTimeout int    `yaml:"shutdown_timeout"`
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return errors.NewError(errors.ErrConfig, fmt.Sprintf("无效的端口号: %d", c.Port))
	}
	if c.ReadTimeout <= 0 {
		return errors.NewError(errors.ErrConfig, "读取超时时间必须大于0")
	}
	if c.WriteTimeout <= 0 {
		return errors.NewError(errors.ErrConfig, "写入超时时间必须大于0")
	}
	if c.ShutdownTimeout <= 0 {
		return errors.NewError(errors.ErrConfig, "关闭超时时间必须大于0")
	}
	return nil
}

// Server HTTP服务器
type Server struct {
	server *http.Server
	engine *gin.Engine
	config *Config
}

// NewServer 创建HTTP服务器
func NewServer(cfg *Config, engine *gin.Engine) (*Server, error) {
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, errors.NewError(errors.ErrConfig, fmt.Sprintf("服务器配置无效: %v", err))
	}

	// 记录服务器配置
	logger.Infof("创建HTTP服务器: Host=%s, Port=%d, ReadTimeout=%ds, WriteTimeout=%ds, ShutdownTimeout=%ds",
		cfg.Host, cfg.Port, cfg.ReadTimeout, cfg.WriteTimeout, cfg.ShutdownTimeout)

	return &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler:      engine,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		},
		engine: engine,
		config: cfg,
	}, nil
}

// Start 启动服务器
func (s *Server) Start() error {
	logger.Infof("正在启动HTTP服务器: %s", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("启动服务器失败: %v", err))
	}

	return nil
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	logger.Infof("正在关闭HTTP服务器...")

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(s.config.ShutdownTimeout)*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := s.server.Shutdown(timeoutCtx); err != nil {
		return errors.NewError(errors.ErrSystem, fmt.Sprintf("关闭服务器失败: %v", err))
	}

	logger.Infof("HTTP服务器已关闭")
	return nil
}
