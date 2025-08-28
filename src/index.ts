import { Hono } from "hono"
import {
  formatHostsFile,
  getDomainData,
  getHostsData,
  resetHostsData,
} from "./services/hosts"
import { handleSchedule } from "./scheduled"
import { Bindings } from "./types"

const app = new Hono<{ Bindings: Bindings }>()

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

app.post("/reset", async (c) => {
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

app.get("/:domain", async (c) => {
  const domain = c.req.param("domain")
  const data = await getDomainData(c.env, domain)

  if (!data) {
    return c.json({ error: "Domain not found" }, 404)
  }

  return c.json(data)
})

export default {
  fetch: app.fetch,
  async scheduled(event: ScheduledEvent, env: Bindings, ctx: ExecutionContext) {
    ctx.waitUntil(handleSchedule(event, env))
  },
}
