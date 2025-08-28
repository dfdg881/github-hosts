package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"
)

// setupCron 根据操作系统设置定时任务
func (app *App) setupCron(interval int) error {
	// 创建更新脚本
	scriptPath := filepath.Join(app.baseDir, "update.sh")
	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(app.baseDir, "update.bat")
	}

	if err := app.createUpdateScript(scriptPath); err != nil {
		return fmt.Errorf("创建更新脚本失败: %w", err)
	}

	// 根据操作系统选择不同的定时任务实现
	switch runtime.GOOS {
	case "darwin":
		return app.setupDarwinCron(interval, scriptPath)
	case "linux":
		return app.setupLinuxCron(interval, scriptPath)
	case "windows":
		return app.setupWindowsCron(interval, scriptPath)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// createUpdateScript 创建更新脚本
func (app *App) createUpdateScript(path string) error {
	var content string
	if runtime.GOOS == "windows" {
		// Windows 批处理脚本
		content = `@echo off
echo [%date% %time%] 开始更新 hosts 文件... >> "{{.LogDir}}\update.log"

:: 清理已存在的 GitHub Hosts 内容
powershell -Command "& {(Get-Content '{{.HostsFile}}') -notmatch '===== GitHub Hosts (Start|End) =====' | Set-Content '{{.HostsFile}}.tmp'}"
move /Y "{{.HostsFile}}.tmp" "{{.HostsFile}}"

:: 获取新的 hosts 内容
echo # ===== GitHub Hosts Start ===== >> "{{.HostsFile}}"
powershell -Command "& {(New-Object System.Net.WebClient).DownloadString('{{.HostsAPI}}')}" >> "{{.HostsFile}}"
echo # ===== GitHub Hosts End ===== >> "{{.HostsFile}}"

:: 刷新 DNS 缓存
ipconfig /flushdns

echo [%date% %time%] 更新完成 >> "{{.LogDir}}\update.log"
`
	} else {
		// Unix 系统脚本
		content = `#!/bin/bash
LOG_FILE="{{.LogDir}}/update_$(date +%Y%m%d).log"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

log() {
    echo "[$TIMESTAMP] $1" >> "$LOG_FILE"
}

log "开始更新 hosts 文件..."

# 清理已存在的 GitHub Hosts 内容
sed -i.bak '/# ===== GitHub Hosts Start =====/,/# ===== GitHub Hosts End =====/d' {{.HostsFile}}

# 获取新的 hosts 内容
echo "# ===== GitHub Hosts Start =====" >> {{.HostsFile}}
curl -fsSL {{.HostsAPI}} >> {{.HostsFile}}
echo "# ===== GitHub Hosts End =====" >> {{.HostsFile}}

# 刷新 DNS 缓存
if [ "$(uname)" == "Darwin" ]; then
    killall -HUP mDNSResponder
    log "已刷新 MacOS DNS 缓存"
else
    if systemd-resolve --flush-caches > /dev/null 2>&1; then
        log "已刷新 Linux DNS 缓存"
    elif systemctl restart systemd-resolved > /dev/null 2>&1; then
        log "已重启 systemd-resolved 服务"
    fi
fi

log "更新完成"
`
	}

	// 准备模板数据
	data := struct {
		LogDir    string
		HostsFile string
		HostsAPI  string
	}{
		LogDir:    app.logDir,
		HostsFile: hostsFile,
		HostsAPI:  hostsAPI,
	}

	// 解析并执行模板
	tmpl, err := template.New("script").Parse(content)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// setupWindowsCron 设置 Windows 计划任务
func (app *App) setupWindowsCron(interval int, scriptPath string) error {
	// 删除已存在的任务
	exec.Command("schtasks", "/delete", "/tn", windowsTaskName, "/f").Run()

	// 创建新任务
	cmd := exec.Command("schtasks", "/create", "/tn", windowsTaskName,
		"/tr", scriptPath,
		"/sc", "minute",
		"/mo", fmt.Sprintf("%d", interval),
		"/ru", "SYSTEM",
		"/f")

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("创建计划任务失败: %s, %v", string(output), err)
	}

	return nil
}

// setupDarwinCron 设置 macOS 定时任务
func (app *App) setupDarwinCron(interval int, scriptPath string) error {
	// 先尝试卸载已存在的服务
	exec.Command("launchctl", "bootout", "system/com.github.hosts").Run()
	// 删除旧的 plist 文件
	os.Remove(darwinPlistPath)

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.github.hosts</string>
    <key>ProgramArguments</key>
    <array>
        <string>/bin/bash</string>
        <string>%s</string>
    </array>
    <key>StartInterval</key>
    <integer>%d</integer>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>`, scriptPath, interval*60)

	// 写入新的 plist 文件
	if err := os.WriteFile(darwinPlistPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入 plist 文件失败: %w", err)
	}

	// 加载新的服务
	cmd := exec.Command("launchctl", "bootstrap", "system", darwinPlistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("加载服务失败: %v, 输出: %s", err, string(output))
	}

	return nil
}

// setupLinuxCron 设置 Linux 定时任务
func (app *App) setupLinuxCron(interval int, scriptPath string) error {
	var schedule string
	switch interval {
	case 30:
		schedule = "*/30 * * * *"
	case 60:
		schedule = "0 * * * *"
	case 120:
		schedule = "0 */2 * * *"
	default:
		return fmt.Errorf("无效的时间间隔: %d", interval)
	}

	content := fmt.Sprintf("%s root %s > %s/update.log 2>&1\n",
		schedule, scriptPath, app.logDir)

	if err := os.WriteFile(linuxCronPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入 cron 文件失败: %w", err)
	}

	// 重启 cron 服务
	if err := exec.Command("systemctl", "restart", "cron").Run(); err != nil {
		return fmt.Errorf("重启 cron 服务失败: %w", err)
	}

	return nil
}
