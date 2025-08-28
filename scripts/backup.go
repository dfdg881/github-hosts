package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// listBackups 获取备份文件列表
func (app *App) listBackups() ([]string, error) {
	files, err := os.ReadDir(app.backupDir)
	if err != nil {
		return nil, err
	}

	var backups []string
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "hosts_") {
			backups = append(backups, file.Name())
		}
	}
	return backups, nil
}

// listBackupsWithDetails 显示备份列表详情
func (app *App) listBackupsWithDetails() error {
	backups, err := app.listBackups()
	if err != nil {
		return err
	}

	if len(backups) == 0 {
		app.logWithLevel(INFO, "没有找到备份文件")
		return nil
	}

	fmt.Println("\n可用的备份文件：")
	fmt.Println("序号\t备份时间\t\t文件大小")
	fmt.Println("----------------------------------------")

	// 按时间排序
	sort.Slice(backups, func(i, j int) bool {
		return backups[i] > backups[j] // 降序排列
	})

	for i, backup := range backups {
		path := filepath.Join(app.backupDir, backup)
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		timeStr := strings.TrimPrefix(backup, "hosts_")
		timeStr = strings.TrimSuffix(timeStr, filepath.Ext(timeStr))
		fmt.Printf("%d\t%s\t%.2f KB\n", i+1, timeStr, float64(info.Size())/1024)
	}
	return nil
}

// createNewBackup 创建新的备份
func (app *App) createNewBackup() error {
	app.logWithLevel(INFO, "创建新的备份...")
	if err := app.backupHosts(); err != nil {
		return err
	}
	app.logWithLevel(SUCCESS, "备份创建成功")
	return nil
}

// backupHosts 备份当前 hosts 文件
func (app *App) backupHosts() error {
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(app.backupDir, fmt.Sprintf("hosts_%s", timestamp))

	input, err := os.ReadFile(hostsFile)
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, input, 0644)
}

// restoreBackupMenu 显示恢复备份菜单
func (app *App) restoreBackupMenu() error {
	backups, err := app.listBackups()
	if err != nil {
		return err
	}

	if len(backups) == 0 {
		app.logWithLevel(INFO, "没有可用的备份文件")
		return nil
	}

	if err := app.listBackupsWithDetails(); err != nil {
		return err
	}

	fmt.Print("\n请选择要恢复的备份序号（0 取消）: ")
	var choice int
	fmt.Scanf("%d", &choice)

	if choice == 0 {
		return nil
	}

	if choice < 1 || choice > len(backups) {
		return fmt.Errorf("无效的选择")
	}

	// 确认恢复
	fmt.Print("确定要恢复这个备份吗？这将覆盖当前的 hosts 文件 [y/N]: ")
	var confirm string
	fmt.Scanf("%s", &confirm)

	if strings.ToLower(confirm) != "y" {
		app.logWithLevel(INFO, "已取消恢复操作")
		return nil
	}

	backupFile := filepath.Join(app.backupDir, backups[choice-1])
	return app.restoreBackup(backupFile)
}

// deleteBackupMenu 显示删除备份菜单
func (app *App) deleteBackupMenu() error {
	backups, err := app.listBackups()
	if err != nil {
		return err
	}

	if len(backups) == 0 {
		app.logWithLevel(INFO, "没有可用的备份文件")
		return nil
	}

	if err := app.listBackupsWithDetails(); err != nil {
		return err
	}

	fmt.Print("\n请选择要删除的备份序号（0 取消）: ")
	var choice int
	fmt.Scanf("%d", &choice)

	if choice == 0 {
		return nil
	}

	if choice < 1 || choice > len(backups) {
		return fmt.Errorf("无效的选择")
	}

	// 确认删除
	fmt.Print("确定要删除这个备份吗？此操作不可恢复 [y/N]: ")
	var confirm string
	fmt.Scanf("%s", &confirm)

	if strings.ToLower(confirm) != "y" {
		app.logWithLevel(INFO, "已取消删除操作")
		return nil
	}

	backupFile := filepath.Join(app.backupDir, backups[choice-1])
	if err := os.Remove(backupFile); err != nil {
		return fmt.Errorf("删除备份失败: %w", err)
	}

	app.logWithLevel(SUCCESS, "备份已删除")
	return nil
}

// restoreBackup 恢复指定的备份文件
func (app *App) restoreBackup(backupFile string) error {
	// 先创建当前 hosts 文件的备份
	if err := app.backupHosts(); err != nil {
		return fmt.Errorf("创建当前 hosts 备份失败: %w", err)
	}

	// 读取备份文件
	content, err := os.ReadFile(backupFile)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %w", err)
	}

	// 写入到 hosts 文件
	if err := os.WriteFile(hostsFile, content, 0644); err != nil {
		return fmt.Errorf("恢复 hosts 文件失败: %w", err)
	}

	// 刷新 DNS 缓存
	if err := app.flushDNSCache(); err != nil {
		app.logWithLevel(WARNING, "DNS 缓存刷新失败: %v", err)
	}

	app.logWithLevel(SUCCESS, "hosts 文件已恢复")
	return nil
}
