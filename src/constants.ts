export const GITHUB_URLS = [
  "alive.github.com",
  "api.github.com",
  "assets-cdn.github.com",
  "avatars.githubusercontent.com",
  "avatars0.githubusercontent.com",
  "avatars1.githubusercontent.com",
  "avatars2.githubusercontent.com",
  "avatars3.githubusercontent.com",
  "avatars4.githubusercontent.com",
  "avatars5.githubusercontent.com",
  "camo.githubusercontent.com",
  "central.github.com",
  "cloud.githubusercontent.com",
  "codeload.github.com",
  "collector.github.com",
  "desktop.githubusercontent.com",
  "favicons.githubusercontent.com",
  "gist.github.com",
  "github-cloud.s3.amazonaws.com",
  "github-com.s3.amazonaws.com",
  "github-production-release-asset-2e65be.s3.amazonaws.com",
  "github-production-repository-file-5c1aeb.s3.amazonaws.com",
  "github-production-user-asset-6210df.s3.amazonaws.com",
  "github.blog",
  "github.com",
  "github.community",
  "github.githubassets.com",
  "github.global.ssl.fastly.net",
  "github.io",
  "github.map.fastly.net",
  "githubstatus.com",
  "live.github.com",
  "media.githubusercontent.com",
  "objects.githubusercontent.com",
  "pipelines.actions.githubusercontent.com",
  "raw.githubusercontent.com",
  "user-images.githubusercontent.com",
  "vscode.dev",
  "education.github.com",
  "private-user-images.githubusercontent.com",
]

export const HOSTS_TEMPLATE = `# github hosts
# 加速 GitHub 访问

{content}

# 数据更新时间：{updateTime}
`

export const GITHUB_API_BASE = "https://api.github.com"

export const HOSTS_PATH = "hosts"

export const DNS_PROVIDERS = [
  {
    url: (domain: string) => `https://1.1.1.1/dns-query?name=${domain}&type=A`,
    headers: { Accept: "application/dns-json" },
    name: "Cloudflare DNS",
  },
  {
    url: (domain: string) => `https://dns.google/resolve?name=${domain}&type=A`,
    headers: { Accept: "application/dns-json" },
    name: "Google DNS",
  },
]
