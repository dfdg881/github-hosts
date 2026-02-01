export interface Bindings {
  HOSTS_STORE: KVNamespace
  API_KEY: string
  ASSETS: { get(key: string): Promise<string | null> }
  github_hosts: KVNamespace
}

export interface RateLimitConfig {
  limit: number
  windowMs: number
}
