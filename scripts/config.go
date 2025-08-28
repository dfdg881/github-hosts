package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// toggleAutoUpdate 切换自动更新状态
func (app *App) toggleAutoUpdate() error {
	config, err := app.loadConfig()
	if err != nil {
		return fmt.Errorf("读取配置失败: %w", err)
	}

	currentStatus := map[bool]string{true: "开启", false: "关闭"}[config.AutoUpdate]
	targetStatus := map[bool]string{true: "关闭", false: "开启"}[config.AutoUpdate]

	fmt.Printf("\n当前自动更新已%s，是否%s？[y/N]: ", currentStatus, targetStatus)

	var response string
	fmt.Scanf("%s", &response)

	if response != "y" && response != "Y" {
		app.logWithLevel(INFO, "保持当前状态不变")
		return nil
	}

	// 更新配置
	config.AutoUpdate = !config.AutoUpdate
	if err := app.updateConfig(config.UpdateInterval, config.AutoUpdate); err != nil {
		return fmt.Errorf("更新配置失败: %w", err)
	}

	if config.AutoUpdate {
		// 开启自动更新时，设置定时任务
		if err := app.setupCron(config.UpdateInterval); err != nil {
			app.logWithLevel(ERROR, "设置定时任务失败: %v", err)
			// 回滚配置
			config.AutoUpdate = false
			app.updateConfig(config.UpdateInterval, false)
			return fmt.Errorf("设置定时任务失败: %w", err)
		}
		app.logWithLevel(SUCCESS, "自动更新已开启，更新间隔为 %d 分钟", config.UpdateInterval)
	} else {
		// 关闭自动更新时，移除定时任务
		if runtime.GOOS == "darwin" {
			exec.Command("launchctl", "bootout", "system/com.github.hosts").Run()
			os.Remove(darwinPlistPath)
		} else if runtime.GOOS == "windows" {
			exec.Command("schtasks", "/delete", "/tn", windowsTaskName, "/f").Run()
		} else {
			os.Remove(linuxCronPath)
		}
		app.logWithLevel(SUCCESS, "自动更新已关闭")
	}

	return nil
}

// changeUpdateInterval 修改更新间隔
func (app *App) changeUpdateInterval() error {
	config, err := app.loadConfig()
	if err != nil {
		return fmt.Errorf("读取配置失败: %w", err)
	}

	fmt.Println("\n请选择新的更新间隔：")
	fmt.Println("1. 每 30 分钟")
	fmt.Println("2. 每 60 分钟")
	fmt.Println("3. 每 120 分钟")

	var choice int
	fmt.Scanf("%d", &choice)

	var interval int
	switch choice {
	case 1:
		interval = 30
	case 2:
		interval = 60
	case 3:
		interval = 120
	default:
		return fmt.Errorf("无效的选项")
	}

	// 更新配置
	if err := app.updateConfig(interval, config.AutoUpdate); err != nil {
		return fmt.Errorf("更新配置失败: %w", err)
	}

	// 如果启用了自动更新，则更新定时任务
	if config.AutoUpdate {
		if err := app.setupCron(interval); err != nil {
			return fmt.Errorf("更新定时任务失败: %w", err)
		}
	}

	app.logWithLevel(SUCCESS, "更新间隔已修改为 %d 分钟", interval)
	return nil
}

// exportConfigToFile 导出配置到文件
func (app *App) exportConfigToFile() error {
	exportPath := filepath.Join(app.baseDir, fmt.Sprintf("config_export_%s.json", time.Now().Format("20060102_150405")))

	data, err := os.ReadFile(app.configFile)
	if err != nil {
		return fmt.Errorf("读取配置失败: %w", err)
	}

	if err := os.WriteFile(exportPath, data, 0644); err != nil {
		return fmt.Errorf("导出配置失败: %w", err)
	}

	app.logWithLevel(SUCCESS, "配置已导出到: %s", exportPath)
	return nil
}

// importConfigFromFile 从文件导入配置
func (app *App) importConfigFromFile() error {
	fmt.Print("请输入配置文件路径: ")
	var path string
	fmt.Scanf("%s", &path)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("配置文件格式无效: %w", err)
	}

	if err := os.WriteFile(app.configFile, data, 0644); err != nil {
		return fmt.Errorf("更新配置失败: %w", err)
	}

	app.logWithLevel(SUCCESS, "配置导入成功")
	return nil
}
