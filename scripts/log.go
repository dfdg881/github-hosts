package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// logWithLevel 输出带有级别的日志，并同时写入日志文件
// writeToFile 参数控制是否写入日志文件，默认为 true
func (app *App) logWithLevel(level LogLevel, format string, args ...interface{}) {
	app.logWithLevelOpt(level, true, format, args...)
}

// logWithLevelOpt 输出带有级别的日志，可选择是否写入日志文件
func (app *App) logWithLevelOpt(level LogLevel, writeToFile bool, format string, args ...interface{}) {
	var prefix string
	switch level {
	case INFO:
		prefix = "ℹ️  "
	case SUCCESS:
		prefix = "✅ "
	case WARNING:
		prefix = "⚠️  "
	case ERROR:
		prefix = "❌ "
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("%s[%s] %s\n", prefix, timestamp, message)

	// 输出到控制台
	fmt.Print(logLine)

	// 如果不需要写入文件，直接返回
	if !writeToFile {
		return
	}

	// 写入到日志文件
	logFile := filepath.Join(app.logDir, fmt.Sprintf("update_%s.log", time.Now().Format("20060102")))

	// 确保日志目录存在
	if err := os.MkdirAll(app.logDir, 0755); err != nil {
		fmt.Printf("❌ 创建日志目录失败: %v\n", err)
		return
	}

	// 以追加模式打开日志文件
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("❌ 打开日志文件失败: %v\n", err)
		return
	}
	defer f.Close()

	// 写入日志
	if _, err := f.WriteString(logLine); err != nil {
		fmt.Printf("❌ 写入日志失败: %v\n", err)
	}
}
