package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (app *App) installMenu() error {
	app.logWithLevel(INFO, "检查系统状态...")

	// 首先检查是否已存在 hosts 数据
	content, err := os.ReadFile(hostsFile)
	if err == nil && strings.Contains(string(content), "GitHub Hosts") {
		// 已存在 GitHub Hosts 数据，询问是否更新
		fmt.Print("\n检测到已存在 GitHub Hosts 数据，是否要更新？[Y/n]: ")
		var updateResponse string
		fmt.Scanf("%s", &updateResponse)

		if updateResponse == "n" || updateResponse == "N" {
			app.logWithLevel(INFO, "已取消更新操作")
			return nil
		}

		// 用户选择更新，直接执行更新操作
		if err := app.updateHosts(); err != nil {
			app.logWithLevel(ERROR, "更新 hosts 失败: %v", err)
			return fmt.Errorf("更新 hosts 失败: %w", err)
		}
		app.logWithLevel(SUCCESS, "hosts 文件更新完成")
		return nil
	}

	// 不存在 GitHub Hosts 数据，执行完整的安装流程
	app.logWithLevel(INFO, "开始安装配置向导...")

	// 1. 选择是否开启自动更新
	var autoUpdate bool = true // 默认开启
	fmt.Print("\n是否开启自动更新？[Y/n]: ")
	var response string
	fmt.Scanf("%s", &response)

	var interval int = 60 // 默认 60 分钟
	if response == "n" || response == "N" {
		autoUpdate = false
		app.logWithLevel(INFO, "已禁用自动更新")
	} else {
		app.logWithLevel(INFO, "已启用自动更新")

		fmt.Println("\n请选择更新间隔：")
		fmt.Println("1. 每 30 分钟")
		fmt.Println("2. 每 60 分钟")
		fmt.Println("3. 每 120 分钟")
		fmt.Print("请输入选项 (1-3): ")

		var choice int
		fmt.Scanf("%d", &choice)

		switch choice {
		case 1:
			interval = 30
		case 2:
			interval = 60
		case 3:
			interval = 120
		default:
			app.logWithLevel(ERROR, "无效的选项，将使用默认间隔（60分钟）")
			interval = 60
		}
		app.logWithLevel(INFO, "选择的更新间隔: %d 分钟", interval)
	}

	app.logWithLevel(INFO, "开始执行安装流程...")

	// 1. Setup directories
	app.logWithLevel(INFO, "第 1/4 步: 创建必要的目录结构")
	if err := app.setupDirectories(); err != nil {
		app.logWithLevel(ERROR, "创建目录失败: %v", err)
		return fmt.Errorf("创建目录失败: %w", err)
	}
	app.logWithLevel(SUCCESS, "目录创建完成")
	app.logWithLevel(INFO, "  - 基础目录: %s", app.baseDir)
	app.logWithLevel(INFO, "  - 配置文件: %s", app.configFile)
	app.logWithLevel(INFO, "  - 备份目录: %s", app.backupDir)
	app.logWithLevel(INFO, "  - 日志目录: %s", app.logDir)

	// 2. Update config
	app.logWithLevel(INFO, "第 2/4 步: 更新配置文件")
	if err := app.updateConfig(interval, autoUpdate); err != nil {
		app.logWithLevel(ERROR, "更新配置失败: %v", err)
		return fmt.Errorf("更新配置失败: %w", err)
	}
	app.logWithLevel(SUCCESS, "配置文件更新完成")

	// 3. Update hosts
	app.logWithLevel(INFO, "第 3/4 步: 更新 hosts 文件")
	if err := app.updateHosts(); err != nil {
		app.logWithLevel(ERROR, "更新 hosts 失败: %v", err)
		return fmt.Errorf("更新 hosts 失败: %w", err)
	}
	app.logWithLevel(SUCCESS, "hosts 文件更新完成")

	// 4. Setup cron
	if autoUpdate {
		app.logWithLevel(INFO, "第 4/4 步: 设置定时更新任务")
		if err := app.setupCron(interval); err != nil {
			app.logWithLevel(ERROR, "设置定时任务失败: %v", err)
			return fmt.Errorf("设置定时任务失败: %w", err)
		}
		app.logWithLevel(SUCCESS, "定时任务设置完成")
	} else {
		app.logWithLevel(INFO, "已跳过定时任务设置（自动更新已禁用）")
	}

	// 显示安装完成信息
	app.logWithLevel(SUCCESS, "安装完成！")
	app.logWithLevel(INFO, "系统配置信息：")
	if autoUpdate {
		app.logWithLevel(INFO, "  • 更新间隔: 每 %d 分钟", interval)
	}
	app.logWithLevel(INFO, "  • 自动更新: %s", map[bool]string{true: "已启用", false: "已禁用"}[autoUpdate])
	app.logWithLevel(INFO, "  • 配置文件: %s", app.configFile)
	app.logWithLevel(INFO, "  • 日志文件: %s", filepath.Join(app.logDir, "update.log"))
	app.logWithLevel(INFO, "  • 备份目录: %s", app.backupDir)

	// 显示当前 hosts 文件内容
	app.logWithLevel(INFO, "\n当前 hosts 文件内容：")
	fmt.Println("----------------------------------------")
	content, err = os.ReadFile(hostsFile)
	if err != nil {
		app.logWithLevel(ERROR, "读取 hosts 文件失败: %v", err)
	} else {
		fmt.Println(string(content))
	}
	fmt.Println("----------------------------------------")

	// 自动执行网络连接测试
	app.logWithLevel(INFO, "\n开始测试网络连接...")
	if err := app.testConnection(); err != nil {
		app.logWithLevel(WARNING, "网络连接测试出现问题: %v", err)
	}

	return nil
}

func (app *App) setupDirectories() error {
	dirs := []string{app.baseDir, app.backupDir, app.logDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (app *App) updateConfig(interval int, autoUpdate bool) error {
	config := Config{
		UpdateInterval: interval,
		LastUpdate:     time.Now().UTC(),
		Version:        "1.0.0",
		AutoUpdate:     autoUpdate,
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(app.configFile, data, 0644)
}

func (app *App) updateHosts() error {
	app.logWithLevel(INFO, "开始备份当前 hosts 文件")
	if err := app.backupHosts(); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	app.logWithLevel(SUCCESS, "hosts 文件备份完成")

	// 先清理已存在的 GitHub Hosts 内容
	app.logWithLevel(INFO, "清理已存在的 GitHub Hosts 内容")
	if err := app.cleanHostsFile(); err != nil {
		return fmt.Errorf("清理已存在内容失败: %w", err)
	}
	app.logWithLevel(SUCCESS, "已清理旧的 hosts 内容")

	app.logWithLevel(INFO, "正在从服务器获取最新 hosts 数据")
	resp, err := http.Get(hostsAPI)
	if err != nil {
		return fmt.Errorf("failed to download hosts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	app.logWithLevel(SUCCESS, "成功获取最新 hosts 数据")

	app.logWithLevel(INFO, "正在更新本地 hosts 文件")
	f, err := os.OpenFile(hostsFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open hosts file: %w", err)
	}
	defer f.Close()

	// 添加开始标记和更新时间
	startMarker := fmt.Sprintf("\n# ===== GitHub Hosts Start ===== \n# (Updated: %s)\n",
		time.Now().Format("2006-01-02 15:04:05"))
	if _, err := f.WriteString(startMarker); err != nil {
		return fmt.Errorf("failed to write start marker: %w", err)
	}

	// 写入 hosts 内容
	if _, err := f.Write(content); err != nil {
		return fmt.Errorf("failed to write hosts content: %w", err)
	}

	// 添加结束标记
	endMarker := "# ===== GitHub Hosts End =====\n"
	if _, err := f.WriteString(endMarker); err != nil {
		return fmt.Errorf("failed to write end marker: %w", err)
	}

	app.logWithLevel(SUCCESS, "hosts 文件更新成功")

	app.logWithLevel(INFO, "正在刷新 DNS 缓存")
	if err := app.flushDNSCache(); err != nil {
		app.logWithLevel(WARNING, "DNS 缓存刷新失败: %v", err)
	} else {
		app.logWithLevel(SUCCESS, "DNS 缓存刷新完成")
	}

	return nil
}
