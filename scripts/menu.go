package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// showHostsContent 显示 hosts 文件内容
func (app *App) showHostsContent() error {
	// 读取 hosts 文件内容
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return fmt.Errorf("读取 hosts 文件失败: %w", err)
	}

	// 显示完整内容
	fmt.Printf("\n当前 hosts 文件内容 (%s)：\n", hostsFile)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(string(content))
	fmt.Println(strings.Repeat("-", 80))

	// 显示文件信息
	if info, err := os.Stat(hostsFile); err == nil {
		fmt.Printf("文件大小: %.2f KB\n", float64(info.Size())/1024)
		fmt.Printf("修改时间: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
	}

	return nil
}

// openConfigDir 打开配置目录
func (app *App) openConfigDir() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", app.baseDir)
	case "linux":
		cmd = exec.Command("xdg-open", app.baseDir)
	case "windows":
		cmd = exec.Command("explorer", app.baseDir)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("打开目录失败: %w", err)
	}

	return nil
}

// showUpdateLogs 显示更新日志
func (app *App) showUpdateLogs() error {
	logFile := filepath.Join(app.logDir, fmt.Sprintf("update_%s.log", time.Now().Format("20060102")))

	content, err := os.ReadFile(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			app.logWithLevel(INFO, "今日暂无更新日志")
			return nil
		}
		return fmt.Errorf("读取日志文件失败: %w", err)
	}

	fmt.Println("\n最近的更新日志:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(string(content))
	fmt.Println(strings.Repeat("-", 80))
	return nil
}

// checkStatus 检查系统状态
func (app *App) checkStatus() error {
	app.logWithLevel(INFO, "开始检查系统状态...")

	// 1. 检查配置文件
	config, err := app.loadConfig()
	if err != nil {
		app.logWithLevel(ERROR, "配置文件检查失败: %v", err)
	} else {
		app.logWithLevel(INFO, "配置文件状态:")
		app.logWithLevel(INFO, "  • 更新间隔: %d 分钟", config.UpdateInterval)
		app.logWithLevel(INFO, "  • 自动更新: %s", map[bool]string{true: "已启用", false: "已禁用"}[config.AutoUpdate])
		app.logWithLevel(INFO, "  • 最后更新: %s", config.LastUpdate.Local().Format("2006-01-02 15:04:05"))
		app.logWithLevel(INFO, "  • 版本: %s", config.Version)
	}

	// 2. 检查 hosts 文件
	if _, err := os.Stat(hostsFile); err != nil {
		app.logWithLevel(ERROR, "hosts 文件检查失败: %v", err)
	} else {
		content, err := os.ReadFile(hostsFile)
		if err != nil {
			app.logWithLevel(ERROR, "读取 hosts 文件失败: %v", err)
		} else {
			lines := strings.Split(string(content), "\n")
			githubCount := 0
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), "github") {
					githubCount++
				}
			}
			app.logWithLevel(INFO, "hosts 文件状态:")
			app.logWithLevel(INFO, "  • 文件大小: %.2f KB", float64(len(content))/1024)
			app.logWithLevel(INFO, "  • GitHub 相关记录数: %d", githubCount)
		}
	}

	// 3. 检查定时任务状态
	app.logWithLevel(INFO, "定时任务状态:")
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("launchctl", "list", "com.github.hosts")
		if err := cmd.Run(); err == nil {
			app.logWithLevel(SUCCESS, "  • 定时任务运行正常")
		} else {
			app.logWithLevel(WARNING, "  • 定时任务未运行")
		}
	case "windows":
		cmd := exec.Command("schtasks", "/query", "/tn", windowsTaskName)
		if err := cmd.Run(); err == nil {
			app.logWithLevel(SUCCESS, "  • 定时任务运行正常")
		} else {
			app.logWithLevel(WARNING, "  • 定时任务未运行")
		}
	case "linux":
		if _, err := os.Stat(linuxCronPath); err == nil {
			app.logWithLevel(SUCCESS, "  • 定时任务配置正常")
		} else {
			app.logWithLevel(WARNING, "  • 定时任务配置不存在")
		}
	}

	// 4. 检查目录权限
	app.logWithLevel(INFO, "目录权限检查:")
	dirs := []string{app.baseDir, app.backupDir, app.logDir}
	for _, dir := range dirs {
		if err := app.checkDirPermissions(dir); err != nil {
			app.logWithLevel(WARNING, "  • %s: %v", dir, err)
		} else {
			app.logWithLevel(SUCCESS, "  • %s: 权限正常", dir)
		}
	}

	// 5. 检查备份状态
	if backups, err := app.listBackups(); err != nil {
		app.logWithLevel(ERROR, "备份检查失败: %v", err)
	} else {
		app.logWithLevel(INFO, "备份状态:")
		app.logWithLevel(INFO, "  • 备份文件数量: %d", len(backups))
		if len(backups) > 0 {
			app.logWithLevel(INFO, "  • 最新备份: %s", backups[len(backups)-1])
		}
	}

	return nil
}
