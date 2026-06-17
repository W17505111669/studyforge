import { ref } from 'vue'

const isDark = ref(false)
let initialized = false

// 初始化暗色模式（从 localStorage 或系统偏好读取）
function initDarkMode() {
  if (initialized) return
  initialized = true

  const stored = localStorage.getItem('theme')
  if (stored === 'dark') {
    isDark.value = true
  } else if (stored === 'light') {
    isDark.value = false
  } else {
    // 跟随系统偏好
    isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  applyDarkMode()
}

// 应用暗色模式到 <html> 元素
function applyDarkMode() {
  const root = document.documentElement
  if (isDark.value) {
    root.classList.add('dark')
  } else {
    root.classList.remove('dark')
  }
}

// 切换暗色模式
function toggleDarkMode() {
  isDark.value = !isDark.value
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
  applyDarkMode()
}

// 监听系统偏好变化
function setupMediaListener() {
  const mq = window.matchMedia('(prefers-color-scheme: dark)')
  mq.addEventListener('change', (e) => {
    // 仅在用户没有手动设置过主题时跟随系统
    if (!localStorage.getItem('theme')) {
      isDark.value = e.matches
      applyDarkMode()
    }
  })
}

// 自动初始化
initDarkMode()
setupMediaListener()

export function useDarkMode() {
  return {
    isDark,
    toggleDarkMode
  }
}
