import { DNS_PROVIDERS, GITHUB_URLS, HOSTS_TEMPLATE } from "../constants"
import { Bindings } from "../types"

export type HostEntry = [string, string]

interface DomainData {
  ip: string
  lastUpdated: string
  lastChecked: string
}

interface DomainDataList {
  [key: string]: DomainData
}
interface KVData {
  domain_data: DomainDataList
  lastUpdated: string
}

interface DnsQuestion {
  name: string
  type: number
}

interface DnsAnswer {
  name: string
  type: number
  TTL: number
  data: string
}

interface DnsResponse {
  Status: number
  TC: boolean
  RD: boolean
  RA: boolean
  AD: boolean
  CD: boolean
  Question: DnsQuestion[]
  Answer: DnsAnswer[]
}

async function retry<T>(
  fn: () => Promise<T>,
  retries: number = 3,
  delay: number = 1000
): Promise<T> {
  try {
    return await fn()
  } catch (error) {
    if (retries === 0) throw error
    await new Promise((resolve) => setTimeout(resolve, delay))
    return retry(fn, retries - 1, delay * 2)
  }
}

export async function fetchIPFromIPAddress(
  domain: string,
  providerName?: string
): Promise<string | null> {
  const provider =
    DNS_PROVIDERS.find((p) => p.name === providerName) || DNS_PROVIDERS[0]

  try {
    const response = await retry(() =>
      fetch(provider.url(domain), { headers: provider.headers })
    )

    if (!response.ok) return null

    const data = (await response.json()) as DnsResponse

    // 查找类型为 1 (A记录) 的答案
    const aRecord = data.Answer?.find((answer) => answer.type === 1)
    const ip = aRecord?.data

    if (ip && /^\d+\.\d+\.\d+\.\d+$/.test(ip)) {
      return ip
    }
  } catch (error) {
    console.error(`Error with DNS provider:`, error)
  }

  return null
}

export async function fetchLatestHostsData(): Promise<HostEntry[]> {
  const entries: HostEntry[] = []
  const batchSize = 5

  for (let i = 0; i < GITHUB_URLS.length; i += batchSize) {
    console.log(
      `Processing batch ${i / batchSize + 1}/${Math.ceil(
        GITHUB_URLS.length / batchSize
      )}`
    )

    const batch = GITHUB_URLS.slice(i, i + batchSize)
    const batchResults = await Promise.all(
      batch.map(async (domain) => {
        const ip = await fetchIPFromIPAddress(domain)
        console.log(`Domain: ${domain}, IP: ${ip}`)
        return ip ? ([ip, domain] as HostEntry) : null
      })
    )

    entries.push(
      ...batchResults.filter((result): result is HostEntry => result !== null)
    )

    if (i + batchSize < GITHUB_URLS.length) {
      await new Promise((resolve) => setTimeout(resolve, 2000))
    }
  }

  console.log(`Total entries found: ${entries.length}`)
  return entries
}
export async function storeData(
  env: Bindings,
  data: HostEntry[]
): Promise<void> {

  await updateHostsData(env, data)
}
export async function getHostsData(env: Bindings): Promise<HostEntry[]> {
  const kvData = (await env.github_hosts.get("domain_data", {
    type: "json",
  })) as KVData | null

  // 如果数据不存在，或者最后更新时间超过1小时，获取新数据
  if (
    !kvData?.lastUpdated ||
    new Date(kvData.lastUpdated).getTime() + 1000 * 60 * 60 < Date.now() ||
    Object.keys(kvData.domain_data || {}).length === 0
  ) {
    const newEntries = await fetchLatestHostsData()
    await storeData(env, newEntries)
    return newEntries
  }

  try {
    // 从 KV 获取所有域名的数据
    const entries: HostEntry[] = []
    for (const domain of GITHUB_URLS) {
      const domainData = kvData.domain_data[domain]
      if (domainData) {
        entries.push([domainData.ip, domain])
      }
    }
    return entries
  } catch (error) {
    console.error("Error getting hosts data:", error)
    return []
  }
}

export async function updateHostsData(
  env: Bindings,
  newEntries?: HostEntry[]
): Promise<void> {
  try {
    const currentTime = new Date().toISOString()
    const kvData = (await env.github_hosts.get("domain_data", {
      type: "json",
    })) as KVData | null || { domain_data: {}, lastUpdated: currentTime }

    if (!newEntries) {
      // 只更新检查时间
      for (const domain in kvData.domain_data) {
        kvData.domain_data[domain] = {
          ...kvData.domain_data[domain],
          lastChecked: currentTime,
        }
      }
    } else {
      // 更新域名数据
      for (const [ip, domain] of newEntries) {
        const oldData = kvData.domain_data[domain]
        const hasChanged = !oldData || oldData.ip !== ip

        kvData.domain_data[domain] = {
          ip,
          lastUpdated: hasChanged ? currentTime : oldData?.lastUpdated || currentTime,
          lastChecked: currentTime,
        }
      }
    }

    kvData.lastUpdated = currentTime
    await env.github_hosts.put("domain_data", JSON.stringify(kvData))
  } catch (error) {
    console.error("Error updating hosts data:", error)
  }
}

export function formatHostsFile(entries: HostEntry[]): string {
  const content = entries
    .map(([ip, domain]) => `${ip.padEnd(30)}${domain}`)
    .join("\n")

  const updateTime = new Date().toLocaleString("en-US", {
    timeZone: "Asia/Shanghai",
    hour12: false,
  })

  return HOSTS_TEMPLATE.replace("{content}", content).replace(
    "{updateTime}",
    updateTime
  )
}

// 修改：获取单个域名数据的方法，直接从爬虫获取实时数据
export async function getDomainData(
  env: Bindings,
  domain: string
): Promise<DomainData | null> {
  try {
    const ip = await fetchIPFromIPAddress(domain)
    if (!ip) {
      return null
    }

    const currentTime = new Date().toISOString()
    const kvData = (await env.github_hosts.get("domain_data", {
      type: "json",
    })) as KVData | null || { domain_data: {}, lastUpdated: currentTime }

    const newData: DomainData = {
      ip,
      lastUpdated: currentTime,
      lastChecked: currentTime,
    }

    kvData.domain_data[domain] = newData
    kvData.lastUpdated = currentTime
    await env.github_hosts.put("domain_data", JSON.stringify(kvData))

    return newData
  } catch (error) {
    console.error(`Error getting data for domain ${domain}:`, error)
    return null
  }
}

// 修改：清空 KV 并重新获取所有数据
export async function resetHostsData(env: Bindings): Promise<HostEntry[]> {
  try {
    console.log("Clearing KV data...")
    await env.github_hosts.delete("domain_data")
    console.log("KV data cleared")

    console.log("Fetching new data...")
    const newEntries = await fetchLatestHostsData()
    console.log("New entries fetched:", newEntries)

    await updateHostsData(env, newEntries)
    console.log("New data stored in KV")

    return newEntries
  } catch (error) {
    console.error("Error resetting hosts data:", error)
    return []
  }
}
