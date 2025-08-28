package main

import (
	"log"
	"runtime"
	"time"
)

// App 应用程序结构体
type App struct {
	baseDir    string
	configFile string
	backupDir  string
	logDir     string
	logger     *log.Logger
}

// Config 配置文件结构体
type Config struct {
	UpdateInterval int       `json:"updateInterval"`
	LastUpdate     time.Time `json:"lastUpdate"`
	Version        string    `json:"version"`
	AutoUpdate     bool      `json:"autoUpdate"`
}

// LogLevel 定义日志级别
type LogLevel int

const (
	INFO LogLevel = iota
	SUCCESS
	WARNING
	ERROR
)

const (
	MaxMenuOption = 12 // Maximum menu option number
)

// displayOption 定义菜单选项
type displayOption struct {
	id          int
	name        string
	description string
	handler     func(*App) error
}

// 系统相关常量
var (
	// hostsAPI 定义 API 地址
	hostsAPI = "https://github-hosts.tinsfox.com/hosts"

	// hostsFile 根据操作系统定义 hosts 文件路径
	hostsFile = getHostsFilePath()

	// 定时任务相关路径
	windowsTaskName = "GitHubHostsUpdate"
	darwinPlistPath = "/Library/LaunchDaemons/com.github.hosts.plist"
	linuxCronPath   = "/etc/cron.d/github-hosts"
)

// getHostsFilePath 根据操作系统返回 hosts 文件路径
func getHostsFilePath() string {
	if runtime.GOOS == "windows" {
		return "C:\\Windows\\System32\\drivers\\etc\\hosts"
	}
	return "/etc/hosts"
}
