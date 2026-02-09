import { Hono } from "hono"
import {
  formatHostsFile,
  getDomainData,
  getHostsData,
  resetHostsData,
} from "./services/hosts"
import { handleSchedule } from "./scheduled"
import { Bindings } from "./types"
import { GITHUB_URLS } from "./constants"
import { rateLimit } from "./middleware/rate-limit"

const app = new Hono<{ Bindings: Bindings }>()

const MAINTENANCE_MESSAGE = {
  error: "Service temporarily unavailable",
  message: "因接口被刷，KV 配额已耗尽，服务已临时下线，恢复时间待定。",
  timestamp: new Date().toISOString(),
}

const MAINTENANCE_HTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>服务已下线 - GitHub Host</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; 
           max-width: 600px; margin: 50px auto; padding: 20px; text-align: center; }
    .notice { background: #fff3cd; border: 1px solid #ffc107; border-radius: 8px; 
             padding: 30px; margin: 20px 0; }
    .notice h2 { color: #856404; margin-top: 0; }
    .notice p { color: #856404; font-size: 16px; line-height: 1.6; }
  </style>
</head>
<body>
  <div class="notice">
    <h2>⚠️ 服务已临时下线</h2>
    <p>因接口被刷，KV 配额已耗尽。<br><br>
    服务正在紧急修复中，恢复时间待定。<br><br>
    感谢您的理解与支持。</p>
  </div>
</body>
</html>`

// 所有接口返回维护公告
app.use("*", async (c) => {
  const accept = c.req.header("Accept") || ""
  if (accept.includes("text/html")) {
    return c.html(MAINTENANCE_HTML, 503)
  }
  return c.json(MAINTENANCE_MESSAGE, 503)
})

app.get("/", async (c) => {
  const html = await c.env.ASSETS.get("index.html")
  if (!html) {
    return c.text("Template not found", 404)
  }

  return c.html(html)
})

app.get("/hosts.json", async (c) => {
  const data = await getHostsData(c.env)
  return c.json(data)
})

app.get("/hosts", async (c) => {
  const data = await getHostsData(c.env)
  const hostsContent = formatHostsFile(data)
  return c.text(hostsContent)
})

// 管理接口限流: 5 请求/分钟
app.post("/reset", rateLimit({ limit: 5, windowMs: 60_000 }), async (c) => {
  const apiKey = c.req.query("key")

  // 验证 API key
  if (apiKey !== c.env.API_KEY) {
    return c.json({ error: "Unauthorized" }, 401)
  }

  const newEntries = await resetHostsData(c.env)

  return c.json({
    message: "Reset completed",
    entriesCount: newEntries.length,
    entries: newEntries,
  })
})

// 域名查询限流: 30 请求/分钟
app.get("/:domain", rateLimit({ limit: 30, windowMs: 60_000 }), async (c) => {
  const domain = c.req.param("domain")

  // 只允许查询预定义域名
  if (!ALLOWED_DOMAINS.has(domain)) {
    return c.json(
      {
        error: "Domain not in allowed list",
        hint: "Use /hosts endpoint to get all supported domains",
      },
      400
    )
  }

  const data = await getDomainData(c.env, domain)

  if (!data) {
    return c.json({ error: "Failed to resolve domain" }, 500)
  }

  return c.json(data)
})

export default {
  fetch: app.fetch,
  async scheduled(event: ScheduledEvent, env: Bindings, ctx: ExecutionContext) {
    ctx.waitUntil(handleSchedule(event, env))
  },
}
