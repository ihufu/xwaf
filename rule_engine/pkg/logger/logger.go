package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
)

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

// Init 初始化日志
func Init(cfg *LogConfig) error {
	// 配置日志级别
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return fmt.Errorf("解析日志级别失败: %v", err)
	}

	// 配置日志输出
	hook := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,    // 每个日志文件最大尺寸(MB)
		MaxBackups: cfg.MaxBackups, // 保留的旧日志文件最大数量
		MaxAge:     cfg.MaxAge,     // 保留的旧日志文件最大天数
		Compress:   cfg.Compress,   // 是否压缩旧日志文件
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 配置核心
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(hook)),
		level,
	)

	// 创建日志记录器
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = logger.Sugar()

	return nil
}

// Debug 记录调试级别日志
func Debug(msg string, args ...interface{}) {
	sugar.Debugf(msg, args...)
}

// Info 记录信息级别日志
func Info(msg string, args ...interface{}) {
	sugar.Infof(msg, args...)
}

// Warn 记录警告级别日志
func Warn(msg string, args ...interface{}) {
	sugar.Warnf(msg, args...)
}

// Error 记录错误级别日志
func Error(msg string, args ...interface{}) {
	sugar.Errorf(msg, args...)
}

// Fatal 记录致命级别日志
func Fatal(msg string, args ...interface{}) {
	sugar.Fatalf(msg, args...)
}

// Sync 同步日志
func Sync() error {
	return logger.Sync()
}

// With 添加字段
func With(fields ...interface{}) *zap.SugaredLogger {
	return sugar.With(fields...)
}

// Named 添加名称
func Named(name string) *zap.Logger {
	return logger.Named(name)
}

// Warnf 记录格式化的警告级别日志
func Warnf(format string, args ...interface{}) {
	sugar.Warnf(format, args...)
}

// Errorf 记录格式化的错误级别日志
func Errorf(format string, args ...interface{}) {
	sugar.Errorf(format, args...)
}

// Infof 记录格式化的信息级别日志
func Infof(format string, args ...interface{}) {
	sugar.Infof(format, args...)
}
