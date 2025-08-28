package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func (app *App) uninstall() error {
	// 首先检查程序是否已安装
	installed, _ := app.checkInstallStatus()
	if !installed {
		app.logWithLevelOpt(ERROR, false, "程序尚未安装，无需卸载")
		return fmt.Errorf("程序未安装")
	}

	app.logWithLevelOpt(INFO, false, "准备卸载 GitHub Hosts 更新程序...")
	app.logWithLevelOpt(WARNING, false, "此操作将删除所有程序文件、配置和日志，且不可恢复")

	// 询问用户确认
	fmt.Print("确定要卸载吗？(y/N): ")
	var response string
	fmt.Scanln(&response)

	// 检查用户响应
	if response != "y" && response != "Y" {
		app.logWithLevelOpt(INFO, false, "已取消卸载")
		return nil
	}

	app.logWithLevelOpt(INFO, false, "开始卸载...")

	// 1. 清理 hosts 文件中的 GitHub 相关记录
	app.logWithLevelOpt(INFO, false, "正在清理 hosts 文件...")
	if err := app.cleanHostsFile(); err != nil {
		app.logWithLevelOpt(ERROR, false, "清理 hosts 文件失败: %v", err)
		return err
	}
	app.logWithLevelOpt(SUCCESS, false, "hosts 文件已清理")

	// 2. 移除定时任务
	app.logWithLevelOpt(INFO, false, "正在移除定时任务...")
	if runtime.GOOS == "darwin" {
		exec.Command("launchctl", "bootout", "system/com.github.hosts").Run()
		os.Remove(darwinPlistPath)
	} else if runtime.GOOS == "windows" {
		exec.Command("schtasks", "/delete", "/tn", windowsTaskName, "/f").Run()
	} else {
		os.Remove(linuxCronPath)
	}
	app.logWithLevelOpt(SUCCESS, false, "定时任务已移除")

	// 3. 删除程序文件和目录
	app.logWithLevelOpt(INFO, false, "正在删除程序文件...")

	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		app.logWithLevelOpt(ERROR, false, "获取用户主目录失败: %v", err)
		return err
	}

	// 需要删除的目录列表
	dirsToRemove := []string{
		app.baseDir,                        // 主程序目录
		app.backupDir,                      // 备份目录
		app.logDir,                         // 日志目录
		homeDir + "/.github-hosts",         // 配置目录
		homeDir + "/.github-hosts/backups", // 备份目录
		homeDir + "/.github-hosts/logs",    // 日志目录
	}

	// 删除所有相关目录
	for _, dir := range dirsToRemove {
		if err := os.RemoveAll(dir); err != nil {
			app.logWithLevelOpt(WARNING, false, "删除目录失败: %s: %v", dir, err)
		}
	}

	// 4. 刷新 DNS 缓存
	app.logWithLevelOpt(INFO, false, "正在刷新 DNS 缓存...")
	if err := app.flushDNSCache(); err != nil {
		app.logWithLevelOpt(WARNING, false, "DNS 缓存刷新失败: %v", err)
	}

	app.logWithLevelOpt(SUCCESS, false, "卸载完成")
	app.logWithLevelOpt(INFO, false, "所有程序文件和配置已清理干净")
	return nil
}

// cleanHostsFile 清理 hosts 文件中的 GitHub 相关记录
func (app *App) cleanHostsFile() error {
	// 读取 hosts 文件内容
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return fmt.Errorf("读取 hosts 文件失败: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	var lastLineEmpty bool = true // 用于跟踪上一行是否为空

	// 逐行处理，移除 GitHub 相关记录和多余的空行
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 跳过 GitHub 相关记录
		if strings.Contains(trimmedLine, "github") || strings.Contains(trimmedLine, "githubusercontent") {
			continue
		}

		// 处理空行：只有当上一行不是空行时才保留当前空行
		if trimmedLine == "" {
			if lastLineEmpty {
				continue // 跳过连续的空行
			}
			lastLineEmpty = true
		} else {
			lastLineEmpty = false
		}

		newLines = append(newLines, line)
	}

	// 确保文件末尾只有一个换行符
	for len(newLines) > 0 && strings.TrimSpace(newLines[len(newLines)-1]) == "" {
		newLines = newLines[:len(newLines)-1]
	}
	newLines = append(newLines, "") // 添加一个空行作为文件结尾

	// 写回文件
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(hostsFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("写入 hosts 文件失败: %w", err)
	}

	return nil
}
