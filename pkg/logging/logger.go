package logging

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 全局日志对象
var Sugar *zap.SugaredLogger
var Logger *zap.Logger

// InitLogger 初始化日志组件
// mode: "development" 或 "production"
// level: "debug", "info", "warn", "error"
func InitLogger(mode string, level string) {
	var config zap.Config
	var err error

	// 1. 根据模式选择配置
	if mode == "production" {
		// 生产环境：JSON 格式，便于机器解析，只记录 Warning 及以上
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	} else {
		// 开发环境：Console 格式，便于人阅读，记录 Debug 及以上
		config = zap.NewDevelopmentConfig()
	}

	// 2. 解析日志级别
	// 允许通过传入字符串设置级别，覆盖默认配置
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err == nil {
		config.Level = zap.NewAtomicLevelAt(zapLevel)
	}

	// 3. 更改日志输出路径 (可选，这里仍输出到标准错误)
	// config.OutputPaths = []string{"/var/log/usb-monitor.log"}

	// 4. 构建 Logger
	Logger, err = config.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error setting up logger: %v\n", err)
		os.Exit(1)
	}
	// 5. 提供 SugaredLogger 以便简化调用 (不需要结构化 Key-Value)
	Sugar = Logger.Sugar()
}

// CloseLogger 确保程序退出时，所有缓冲区的日志都被写入
func CloseLogger() {
	if Logger != nil {
		Logger.Sync()
	}
}
