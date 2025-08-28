// 获取当前页面的基础 URL
const baseUrl = window.location.origin

function escapeHtml(str) {
  const div = document.createElement("div")
  div.textContent = str
  return div.innerHTML
}

async function copyToClipboard(btn) {
  try {
    const hostsElement = document.getElementById("hosts")
    await navigator.clipboard.writeText(hostsElement.textContent)

    const originalText = btn.textContent
    btn.textContent = "已复制"

    setTimeout(() => {
      btn.textContent = originalText
    }, 1000)
  } catch (err) {
    console.error("复制失败:", err)
  }
}

async function loadHosts() {
  const hostsElement = document.getElementById("hosts")
  // 如果元素不存在，直接返回
  if (!hostsElement) return

  try {
    const response = await fetch(`${baseUrl}/hosts`)
    if (!response.ok) throw new Error("Failed to load hosts")
    const hostsContent = await response.text()
    hostsElement.textContent = hostsContent
  } catch (error) {
    hostsElement.textContent = "加载 hosts 内容失败，请稍后重试"
    console.error("Error loading hosts:", error)
  }
}

function setupEventListeners() {
  document.querySelectorAll(".copy-btn").forEach((btn) => {
    btn.addEventListener("click", () => copyToClipboard(btn))
  })

  document.addEventListener("click", (e) => {
    if (e.target.classList.contains("response-collapse-btn")) {
      toggleCollapse(e.target)
    }

    if (e.target.closest(".response-area.collapsed")) {
      const collapseBtn = e.target
        .closest(".response-area")
        .querySelector(".response-collapse-btn")
      if (collapseBtn) {
        toggleCollapse(collapseBtn)
      }
    }
  })
}

// 确保 DOM 完全加载后再执行
document.addEventListener("DOMContentLoaded", () => {
  loadHosts()
  setupEventListeners()
})
