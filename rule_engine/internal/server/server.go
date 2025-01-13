package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Config HTTP服务器配置
type Config struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	ReadTimeout     int    `yaml:"read_timeout"`
	WriteTimeout    int    `yaml:"write_timeout"`
	ShutdownTimeout int    `yaml:"shutdown_timeout"`
}

// Server HTTP服务器
type Server struct {
	server *http.Server
	engine *gin.Engine
}

// NewServer 创建HTTP服务器
func NewServer(cfg *Config, engine *gin.Engine) *Server {
	return &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler:      engine,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		},
		engine: engine,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
