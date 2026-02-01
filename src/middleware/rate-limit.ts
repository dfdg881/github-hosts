import { Context, MiddlewareHandler } from "hono"
import { Bindings } from "../types"

export interface RateLimitConfig {
  limit: number
  windowMs: number
}

const DEFAULT_CONFIG: RateLimitConfig = {
  limit: 60,
  windowMs: 60_000,
}

/**
 * Normalize IPv6 to /64 prefix to prevent bypass via suffix rotation.
 * IPv4 addresses are returned as-is.
 */
function normalizeIP(ip: string): string {
  if (!ip.includes(":")) {
    return ip
  }

  // Expand :: shorthand to full form
  const parts = ip.split(":")
  if (parts.length < 8) {
    const emptyIndex = parts.indexOf("")
    if (emptyIndex !== -1) {
      const missing = 8 - parts.filter((p) => p !== "").length
      const fill = Array(missing).fill("0000")
      parts.splice(emptyIndex, 1, ...fill)
    }
  }

  // Take first 4 groups (/64 prefix)
  return parts.slice(0, 4).join(":")
}

function getWindowStart(windowMs: number): number {
  return Math.floor(Date.now() / windowMs) * windowMs
}

export function rateLimit(
  config: Partial<RateLimitConfig> = {}
): MiddlewareHandler<{ Bindings: Bindings }> {
  const { limit, windowMs } = { ...DEFAULT_CONFIG, ...config }

  return async (c, next) => {
    const ip = c.req.header("cf-connecting-ip") || "unknown"
    const normalizedIP = normalizeIP(ip)
    const windowStart = getWindowStart(windowMs)
    const key = `ratelimit:${normalizedIP}:${windowStart}`

    const kv = c.env.github_hosts
    const current = parseInt((await kv.get(key)) || "0", 10)

    const windowEnd = windowStart + windowMs
    const resetTimestamp = Math.ceil(windowEnd / 1000)
    const remaining = Math.max(0, limit - current - 1)
    const retryAfter = Math.ceil((windowEnd - Date.now()) / 1000)

    c.header("RateLimit-Limit", String(limit))
    c.header("RateLimit-Remaining", String(Math.max(0, limit - current - 1)))
    c.header("RateLimit-Reset", String(resetTimestamp))

    if (current >= limit) {
      c.header("Retry-After", String(retryAfter))
      return c.json(
        { error: "Too Many Requests", retryAfter },
        429
      )
    }

    // Increment counter with TTL (window duration + 60s buffer)
    const ttl = Math.ceil(windowMs / 1000) + 60
    await kv.put(key, String(current + 1), { expirationTtl: ttl })

    await next()
  }
}
