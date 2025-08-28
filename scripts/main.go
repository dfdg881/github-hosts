package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// checkAndElevateSudo 检查权限并在需要时提权
func checkAndElevateSudo() error {
	// Windows 系统使用不同的权限检查方式
	if runtime.GOOS == "windows" {
		// 检查是否以管理员权限运行
		isAdmin, err := isWindowsAdmin()
		if err != nil {
			return fmt.Errorf("检查 Windows 权限失败: %w", err)
		}

		if !isAdmin {
			fmt.Println("需要管理员权限来修改 hosts 文件")
			fmt.Println("请右键点击程序，选择'以管理员身份运行'")

			// 获取当前可执行文件的路径
			exe, err := os.Executable()
			if err != nil {
				return fmt.Errorf("获取程序路径失败: %w", err)
			}

			// 使用 runas 命令提权运行
			cmd := exec.Command("powershell", "Start-Process", exe, "-Verb", "RunAs")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("提权失败: %w", err)
			}

			// 退出当前的非管理员进程
			os.Exit(0)
		}
		return nil
	}

	// Unix 系统的权限检查
	if os.Geteuid() == 0 {
		return nil
	}

	// 检查命令是否以 sudo 运行
	sudoUID := os.Getenv("SUDO_UID")
	if sudoUID != "" {
		return nil
	}

	fmt.Println("需要管理员权限来修改 hosts 文件")
	fmt.Println("请输入 sudo 密码：")

	// 获取当前可执行文件的路径
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}

	// 构建使用 sudo 运行的命令
	cmd := exec.Command("sudo", "-S", exe)

	// 将当前程序的标准输入输出连接到新进程
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 运行提权后的程序
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("提权失败: %w", err)
	}

	// 退出当前的非 root 进程
	os.Exit(0)
	return nil
}

// isWindowsAdmin 检查当前进程是否具有管理员权限
func isWindowsAdmin() (bool, error) {
	if runtime.GOOS != "windows" {
		return false, fmt.Errorf("不是 Windows 系统")
	}

	// 创建一个测试文件在系统目录
	testPath := filepath.Join(os.Getenv("windir"), ".test")
	err := os.WriteFile(testPath, []byte("test"), 0644)
	if err == nil {
		// 如果成功创建，则删除测试文件
		os.Remove(testPath)
		return true, nil
	}

	// 如果创建失败，检查是否是权限问题
	if os.IsPermission(err) {
		return false, nil
	}

	return false, err
}

// clearScreen 清空控制台
func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default: // linux, darwin, etc
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func main() {
	// 检查权限并在需要时提权
	if err := checkAndElevateSudo(); err != nil {
		fmt.Printf("错误: %v\n", err)
		fmt.Println("请使用管理员权限运行此程序")
		os.Exit(1)
	}

	clearScreen() // 启动时先清屏
	fmt.Println(banner)

	app, err := NewApp()
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}

	for {
		// 显示安装状态
		installed, _ := app.checkInstallStatus()
		app.displayInstallStatus()

		fmt.Println("\n[基础功能]")
		fmt.Println("1.  安装/更新")
		if installed {
			fmt.Println("2.  卸载程序")
			fmt.Println("3.  查看 hosts 内容")

			fmt.Println("\n[自动更新]")
			// 动态显示自动更新选项
			config, err := app.loadConfig()
			if err == nil {
				if config.AutoUpdate {
					fmt.Println("4.  关闭自动更新")
				} else {
					fmt.Println("4.  开启自动更新")
				}
			} else {
				fmt.Println("4.  切换自动更新")
			}
			fmt.Println("5.  修改更新间隔")

			fmt.Println("\n[系统工具]")
			fmt.Println("6.  测试网络连接")
			fmt.Println("7.  检查系统状态")
			fmt.Println("8.  查看更新日志")
			fmt.Println("9.  打开配置目录")
			fmt.Println("10. 系统诊断")
		}

		fmt.Println("\n[系统]")
		fmt.Println("11. 打开 hosts 文件")

		fmt.Println("\n[关于]")
		fmt.Println("12. 🐙 访问项目主页")

		fmt.Println("\n0.  退出程序")
		fmt.Printf("\n请输入选项 (0-12 或 q 退出): ")

		// 读取用户输入
		var input string
		fmt.Scanln(&input)

		// 检查是否是退出命令
		if input == "q" || input == "Q" {
			fmt.Println("感谢使用，再见！")
			return
		}

		// 转换输入为数字
		var choice int
		_, err := fmt.Sscanf(input, "%d", &choice)
		if err != nil {
			fmt.Println("无效的选项，请重试")
			waitForEnter()
			continue
		}

		// 在未安装状态下限制某些选项的访问
		if !installed && (choice >= 2 && choice <= 10) {
			fmt.Println("\n❌ 请先安装程序才能使用该功能")
			waitForEnter()
			continue
		}

		switch choice {
		case 1: // 安装/更新
			if err := app.installMenu(); err != nil {
				log.Printf("安装失败: %v", err)
			}
			waitForEnter()
		case 2: // 卸载
			if !installed {
				continue
			}
			if err := app.uninstall(); err != nil {
				log.Printf("卸载失败: %v", err)
			}
			waitForEnter()
		case 3: // 查看 hosts
			if err := app.showHostsContent(); err != nil {
				log.Printf("查看 hosts 内容失败: %v", err)
			}
			waitForEnter()
		case 4: // 切换自动更新
			if err := app.toggleAutoUpdate(); err != nil {
				log.Printf("切换自动更新失败: %v", err)
			}
			waitForEnter()
		case 5: // 修改更新间隔
			if err := app.changeUpdateInterval(); err != nil {
				log.Printf("修改更新间隔失败: %v", err)
			}
			waitForEnter()
		case 6: // 测试网络连接
			if err := app.testConnection(); err != nil {
				log.Printf("网络测试失败: %v", err)
			}
			waitForEnter()
		case 7: // 检查系统状态
			if err := app.checkStatus(); err != nil {
				log.Printf("状态检查失败: %v", err)
			}
			waitForEnter()
		case 8: // 查看更新日志
			if err := app.showUpdateLogs(); err != nil {
				log.Printf("查看日志失败: %v", err)
			}
			waitForEnter()
		case 9: // 打开配置目录
			if err := app.openConfigDir(); err != nil {
				log.Printf("打开配置目录失败: %v", err)
			}
			waitForEnter()
		case 10: // 系统诊断
			if err := app.runDiagnostics(); err != nil {
				log.Printf("系统诊断失败: %v", err)
			}
			waitForEnter()
		case 11: // 打开 hosts 文件
			if err := app.openHostsFile(); err != nil {
				log.Printf("打开 hosts 文件失败: %v", err)
			}
			waitForEnter()
		case 12: // 访问项目主页
			if err := app.openGitHubRepo(); err != nil {
				log.Printf("打开项目主页失败: %v", err)
			}
			waitForEnter()
		case 0: // 退出
			fmt.Println("感谢使用，再见！")
			return
		default:
			fmt.Println("无效的选项，请重试")
			waitForEnter()
		}
	}
}

// NewApp 创建新的应用实例
func NewApp() (*App, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".github-hosts")
	app := &App{
		baseDir:    baseDir,
		configFile: filepath.Join(baseDir, "config.json"),
		backupDir:  filepath.Join(baseDir, "backups"),
		logDir:     filepath.Join(baseDir, "logs"),
		logger:     log.New(os.Stdout, "", log.LstdFlags),
	}

	return app, nil
}

// openGitHubRepo 打开项目主页
func (app *App) openGitHubRepo() error {
	repoURL := "https://github.com/TinsFox/github-hosts"
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", repoURL)
	case "linux":
		cmd = exec.Command("xdg-open", repoURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", repoURL)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("打开浏览器失败: %w", err)
	}

	app.logWithLevel(SUCCESS, "已在浏览器中打开项目主页")
	return nil
}

// loadConfig 加载配置文件
func (app *App) loadConfig() (*Config, error) {
	data, err := os.ReadFile(app.configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// waitForEnter 等待用户按回车并重新显示界面
func waitForEnter() {
	fmt.Print("\n按回车键继续...")
	fmt.Scanln()        // 等待用户按下回车键
	clearScreen()       // 清空控制台
	fmt.Println(banner) // 重新显示 banner
}

// checkInstallStatus 检查程序安装状态
func (app *App) checkInstallStatus() (bool, *InstallStatus) {
	status := &InstallStatus{
		IsInstalled:    false,
		AutoUpdate:     false,
		UpdateInterval: 0,
		LastUpdate:     "",
		Version:        "v1.0.0", // 当前程序版本
	}

	// 检查配置文件是否存在
	config, err := app.loadConfig()
	if err == nil && config != nil {
		status.IsInstalled = true
		status.AutoUpdate = config.AutoUpdate
		status.UpdateInterval = config.UpdateInterval

		// 获取最后更新时间
		if stat, err := os.Stat(app.configFile); err == nil {
			status.LastUpdate = stat.ModTime().Format("2006-01-02 15:04:05")
		}
	}

	return status.IsInstalled, status
}

// displayInstallStatus 显示安装状态
func (app *App) displayInstallStatus() {
	installed, status := app.checkInstallStatus()

	fmt.Println("\n=== 系统状态 ===")
	if installed {
		fmt.Println("📦 安装状态: ✅ 已安装")
		fmt.Printf("🔄 自动更新: %s\n", formatBool(status.AutoUpdate))
		if status.AutoUpdate {
			fmt.Printf("⏱️  更新间隔: %d 小时\n", status.UpdateInterval)
		}
		fmt.Printf("🕒 上次更新: %s\n", status.LastUpdate)
		fmt.Printf("📌 程序版本: %s\n", status.Version)

		// 检查 hosts 文件中的 GitHub 记录数量
		count, _ := app.countGitHubHosts()
		fmt.Printf("📝 GitHub Hosts 记录数: %d\n", count)
	} else {
		fmt.Println("📦 安装状态: ❌ 未安装")
		fmt.Println("💡 提示: 请选择选项 1 进行安装")
	}
	fmt.Println(strings.Repeat("-", 30))
}

// formatBool 格式化布尔值显示
func formatBool(b bool) string {
	if b {
		return "✅ 已开启"
	}
	return "❌ 已关闭"
}

// countGitHubHosts 统计 hosts 文件中的 GitHub 相关记录数量
func (app *App) countGitHubHosts() (int, error) {
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return 0, err
	}

	count := 0
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "github") || strings.Contains(line, "githubusercontent") {
			count++
		}
	}
	return count, nil
}

// InstallStatus 安装状态结构体
type InstallStatus struct {
	IsInstalled    bool
	AutoUpdate     bool
	UpdateInterval int
	LastUpdate     string
	Version        string
}

// openHostsFile 打开 hosts 文件
func (app *App) openHostsFile() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		// macOS 使用 open 命令
		cmd = exec.Command("open", hostsFile)
	case "linux":
		// Linux 使用 xdg-open 命令
		cmd = exec.Command("xdg-open", hostsFile)
	case "windows":
		// Windows 使用 notepad 打开
		cmd = exec.Command("notepad", hostsFile)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("打开 hosts 文件失败: %w", err)
	}

	app.logWithLevel(SUCCESS, "已打开 hosts 文件")
	return nil
}
