import { describe, it, expect, vi, beforeEach } from "vitest"
import { fetchIPFromIPAddress } from "../hosts"

describe("fetchIpFromIpaddress", () => {
  beforeEach(() => {
    // 清除所有模拟
    vi.clearAllMocks()
  })

  it("should successfully extract IP from DNS section", async () => {
    // 模拟 fetch 响应
    global.fetch = vi.fn().mockResolvedValue({
      text: () =>
        Promise.resolve(`
        <html>
          <body>
            <div id="dns">
              <table>
                <tr>
                  <td>140.82.114.25</td>
                </tr>
              </table>
            </div>
          </body>
        </html>
      `),
    })

    const result = await fetchIPFromIPAddress("github.com")
    expect(result).toBe("140.82.114.25")
    expect(fetch).toHaveBeenCalledWith(
      "https://sites.ipaddress.com/github.com",
      expect.any(Object)
    )
  })

  it("should return null when DNS section is not found", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      text: () =>
        Promise.resolve(`
        <html>
          <body>
            <div>No DNS section here</div>
          </body>
        </html>
      `),
    })

    const result = await fetchIPFromIPAddress("invalid-domain.com")
    expect(result).toBeNull()
  })

  it("should handle fetch errors gracefully", async () => {
    global.fetch = vi.fn().mockRejectedValue(new Error("Network error"))

    const result = await fetchIPFromIPAddress("github.com")
    expect(result).toBeNull()
  })

  it("should use fallback IP when DNS section IP is not available", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      text: () =>
        Promise.resolve(`
        <html>
          <body>
            <div id="dns"></div>
            <div>IP Address: 192.168.1.1</div>
          </body>
        </html>
      `),
    })

    const result = await fetchIPFromIPAddress("github.com")
    expect(result).toBe("192.168.1.1")
  })

  it("should handle multiple IPs in DNS section", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      text: () =>
        Promise.resolve(`
        <html>
          <body>
            <div id="dns">
              <table>
                <tr>
                  <td>140.82.114.4</td>
                  <td>140.82.114.5</td>
                </tr>
              </table>
            </div>
          </body>
        </html>
      `),
    })

    const result = await fetchIPFromIPAddress("github.com")
    expect(result).toBe("140.82.114.4") // 应该返回第一个找到的 IP
  })
})
