package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"
)

// testConnection 测试网络连接
func (app *App) testConnection() error {
	app.logWithLevel(INFO, "开始网络连接测试...")
	fmt.Println("\n=== 连接测试结果 ===")

	// 读取 hosts 文件
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		fmt.Printf("\n❌ 严重错误：无法读取 hosts 文件\n")
		fmt.Printf("❌ 错误详情：%v\n", err)
		return err
	}

	// 解析 hosts 文件中的 GitHub 相关记录
	var tests []struct {
		name string
		ip   string
		host string
	}

	startMarker := "# ===== GitHub Hosts Start ====="
	endMarker := "# ===== GitHub Hosts End ====="
	inGithubSection := false
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == startMarker {
			inGithubSection = true
			continue
		}
		if line == endMarker {
			inGithubSection = false
			continue
		}

		if inGithubSection && line != "" && !strings.HasPrefix(line, "#") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				tests = append(tests, struct {
					name string
					ip   string
					host string
				}{
					name: fields[1],
					ip:   fields[0],
					host: fields[1],
				})
			}
		}
	}

	if len(tests) == 0 {
		fmt.Printf("\n❌ 错误：在 hosts 文件中未找到 GitHub 相关记录\n")
		return fmt.Errorf("no github hosts found")
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			DisableKeepAlives:     true,
		},
	}

	// 使用 tabwriter 创建表格
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// 输出表头
	fmt.Fprintf(w, "\n%s\t%s\t%s\t%s\t%s\n",
		"域名",
		"状态",
		"响应时间",
		"当前解析IP",
		"期望IP")
	fmt.Fprintln(w, strings.Repeat("-", 100))

	// 测试结果统计
	var (
		successCount = 0
		failCount    = 0
	)

	// 测试每个域名
	for _, test := range tests {
		start := time.Now()

		// 获取实际 DNS 解析结果
		actualIP := ""
		addrs, err := net.LookupHost(test.host)
		if err != nil {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				test.host,
				"✗ DNS失败",
				"-",
				"解析失败",
				test.ip)
			fmt.Printf("❌ DNS 解析失败: %v\n", err)
			failCount++
			continue
		}
		if len(addrs) > 0 {
			actualIP = addrs[0]
		}

		// 测试连接
		resp, err := client.Get("https://" + test.host)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				test.host,
				"✗ 连接失败",
				fmt.Sprintf("%.2fs", elapsed.Seconds()),
				actualIP,
				test.ip)
			fmt.Printf("❌ 连接失败: %v\n", err)
			failCount++
			continue
		}
		defer resp.Body.Close()

		// 检查 IP 匹配和连接状态
		status := "✓ 正常"
		if actualIP != test.ip {
			status = "! IP不匹配"
			fmt.Printf("⚠️  %s 的 IP 不匹配！当前: %s, 期望: %s\n", test.host, actualIP, test.ip)
			failCount++
		} else if resp.StatusCode != http.StatusOK {
			status = fmt.Sprintf("! 状态%d", resp.StatusCode)
			fmt.Printf("⚠️  %s 返回异常状态码: %d\n", test.host, resp.StatusCode)
			failCount++
		} else {
			successCount++
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			test.host,
			status,
			fmt.Sprintf("%.2fs", elapsed.Seconds()),
			actualIP,
			test.ip)
	}

	fmt.Fprintln(w, strings.Repeat("-", 100))
	w.Flush()

	// 输出总结
	fmt.Printf("\n测试总结:\n")
	fmt.Printf("总计测试: %d\n", len(tests))
	if successCount > 0 {
		fmt.Printf("✅ 成功: %d\n", successCount)
	}
	if failCount > 0 {
		fmt.Printf("❌ 失败: %d\n", failCount)
		fmt.Printf("\n⚠️  警告：检测到 %d 个问题，建议重新执行更新操作\n", failCount)
	} else {
		fmt.Printf("\n✅ 太好了！所有测试都通过了\n")
	}

	return nil
}

// flushDNSCache 刷新 DNS 缓存
func (app *App) flushDNSCache() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("killall", "-HUP", "mDNSResponder")
	case "linux":
		cmd = exec.Command("systemd-resolve", "--flush-caches")
		if err := cmd.Run(); err != nil {
			cmd = exec.Command("systemctl", "restart", "systemd-resolved")
		}
	case "windows":
		cmd = exec.Command("ipconfig", "/flushdns")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return cmd.Run()
}
