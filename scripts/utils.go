package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// checkDirPermissions 检查目录权限
func (app *App) checkDirPermissions(dir string) error {
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("目录不存在: %s", dir)
	}

	// 检查是否可写
	testFile := filepath.Join(dir, ".test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("目录不可写: %s", dir)
	}
	os.Remove(testFile)

	return nil
}

// runDiagnostics 运行系统诊断
func (app *App) runDiagnostics() error {
	app.logWithLevel(INFO, "开始系统诊断...")

	// 1. 检查系统信息
	app.logWithLevel(INFO, "系统信息:")
	app.logWithLevel(INFO, "  • 操作系统: %s", runtime.GOOS)
	app.logWithLevel(INFO, "  • 架构: %s", runtime.GOARCH)

	// 2. 检查目录权限
	app.logWithLevel(INFO, "检查目录权限...")
	dirs := []string{app.baseDir, app.backupDir, app.logDir}
	for _, dir := range dirs {
		if err := app.checkDirPermissions(dir); err != nil {
			app.logWithLevel(WARNING, "目录权限问题: %v", err)
		} else {
			app.logWithLevel(SUCCESS, "目录权限正常: %s", dir)
		}
	}

	// 3. 检查网络连接
	app.logWithLevel(INFO, "检查网络连接...")
	if err := app.testConnection(); err != nil {
		app.logWithLevel(WARNING, "网络连接问题: %v", err)
	}

	// 4. 检查配置文件
	app.logWithLevel(INFO, "检查配置文件...")
	if config, err := app.loadConfig(); err != nil {
		app.logWithLevel(WARNING, "配置文件问题: %v", err)
	} else {
		app.logWithLevel(SUCCESS, "配置文件正常")
		app.logWithLevel(INFO, "  • 版本: %s", config.Version)
		app.logWithLevel(INFO, "  • 上次更新: %s", config.LastUpdate.Local().Format("2006-01-02 15:04:05"))
	}

	app.logWithLevel(SUCCESS, "诊断完成")
	return nil
}
