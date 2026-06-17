import { ref } from 'vue'
import router from '../router'

// ========== Module-level singleton state (shared across all components) ==========
const pendingPrefix = ref('') // 'g' when waiting for second key
const showKeyHint = ref(false) // Show bottom-right key hint bubble
const showHelpModal = ref(false) // Show shortcut help modal

// Signal for Pomodoro toggle (Pomodoro.vue watches this)
const pomodoroToggleSignal = ref(0)

// ========== Default shortcuts ==========
const NAV_SHORTCUTS = {
  d: { path: '/', label: '仪表盘' },
  u: { path: '/upload', label: '上传分析' },
  c: { path: '/cards', label: '知识卡片' },
  q: { path: '/quiz', label: '练习场' },
  h: { path: '/chat', label: 'AI 对话' },
  g: { path: '/graph', label: '知识图谱' },
  s: { path: '/study', label: '学习模式' }
}

const NAV_HINT_LABELS = 'd / u / c / q / h / g / s'

// ========== Custom shortcut mappings (localStorage) ==========
const STORAGE_KEY = 'studyforge-shortcuts'

function loadCustomShortcuts() {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    return stored ? JSON.parse(stored) : {}
  } catch {
    return {}
  }
}

function saveCustomShortcuts(custom) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(custom))
  } catch {}
}

/**
 * Get effective navigation shortcuts (custom overrides merged with defaults)
 */
function getEffectiveNavShortcuts() {
  const custom = loadCustomShortcuts()
  const effective = { ...NAV_SHORTCUTS }
  for (const [key, value] of Object.entries(custom)) {
    if (key.startsWith('g_') && value.path) {
      const navKey = key.slice(2)
      effective[navKey] = { path: value.path, label: value.label || value.path }
    }
  }
  return effective
}

/**
 * Set a custom shortcut for a navigation key (g + key)
 */
function setCustomShortcut(navKey, path, label) {
  const custom = loadCustomShortcuts()
  custom[`g_${navKey}`] = { path, label }
  saveCustomShortcuts(custom)
}

/**
 * Reset a custom shortcut back to default
 */
function resetCustomShortcut(navKey) {
  const custom = loadCustomShortcuts()
  delete custom[`g_${navKey}`]
  saveCustomShortcuts(custom)
}

/**
 * Reset all custom shortcuts to defaults
 */
function resetAllShortcuts() {
  try {
    localStorage.removeItem(STORAGE_KEY)
  } catch {}
}

// ========== Prefix timeout ==========
let prefixTimer = null

function resetPrefix() {
  pendingPrefix.value = ''
  showKeyHint.value = false
  if (prefixTimer) {
    clearTimeout(prefixTimer)
    prefixTimer = null
  }
}

// ========== Editable element detection ==========
function isEditableElement(el) {
  if (!el) return false
  const tag = el.tagName?.toLowerCase()
  if (tag === 'input' || tag === 'textarea' || tag === 'select') return true
  if (el.isContentEditable) return true
  return false
}

// ========== All shortcuts definition (for help panel) ==========
const ALL_SHORTCUTS = [
  {
    category: '导航',
    icon: '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7"/></svg>',
    shortcuts: [
      { keys: ['g', 'd'], desc: '仪表盘' },
      { keys: ['g', 'u'], desc: '上传分析' },
      { keys: ['g', 'c'], desc: '知识卡片' },
      { keys: ['g', 'q'], desc: '练习场' },
      { keys: ['g', 'h'], desc: 'AI 对话' },
      { keys: ['g', 'g'], desc: '知识图谱' },
      { keys: ['g', 's'], desc: '学习模式' }
    ]
  },
  {
    category: '功能',
    icon: '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>',
    shortcuts: [
      { keys: ['n'], desc: '快速创建笔记' },
      { keys: ['t'], desc: '番茄钟开始/暂停' }
    ]
  },
  {
    category: '通用',
    icon: '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"/></svg>',
    shortcuts: [
      { keys: ['Ctrl', 'K'], desc: '全局搜索' },
      { keys: ['?'], desc: '快捷键帮助' },
      { keys: ['Esc'], desc: '关闭弹窗/面板' }
    ]
  },
  {
    category: '聊天',
    icon: '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/></svg>',
    shortcuts: [
      { keys: ['/'], desc: '聚焦聊天输入' },
      { keys: ['Ctrl', 'Enter'], desc: '发送消息' }
    ]
  }
]

// ========== Main keydown handler ==========
let initialized = false

function handleGlobalKeydown(e) {
  // Ignore when modifier keys are held (except for specific shortcuts)
  if (e.altKey) return

  const isEditable = isEditableElement(e.target)

  // Ctrl+K / Cmd+K — Global search is handled by GlobalSearch.vue
  // Let it propagate naturally

  // If in editable element, only handle Esc
  if (isEditable) {
    if (e.key === 'Escape') {
      // Let other handlers deal with Esc in inputs
      return
    }
    return // Don't intercept typing
  }

  // ---- Navigation prefix 'g' ----
  if (pendingPrefix.value === 'g') {
    const key = e.key.toLowerCase()
    const navShortcuts = getEffectiveNavShortcuts()
    if (navShortcuts[key]) {
      e.preventDefault()
      router.push(navShortcuts[key].path)
      resetPrefix()
      return
    }
    // 'g' pressed again — ignore (reset and start fresh)
    if (key === 'g') {
      // Restart prefix timer
      resetPrefix()
      pendingPrefix.value = 'g'
      showKeyHint.value = true
      prefixTimer = setTimeout(resetPrefix, 1500)
      return
    }
    // Invalid key — cancel prefix
    resetPrefix()
    return
  }

  // Start 'g' prefix
  if (e.key === 'g' && !e.ctrlKey && !e.metaKey) {
    pendingPrefix.value = 'g'
    showKeyHint.value = true
    prefixTimer = setTimeout(resetPrefix, 1500)
    return
  }

  // '?' — Toggle help modal
  if (e.key === '?' && !e.ctrlKey && !e.metaKey) {
    e.preventDefault()
    showHelpModal.value = !showHelpModal.value
    return
  }

  // 'n' — Navigate to notes
  if (e.key === 'n' && !e.ctrlKey && !e.metaKey) {
    e.preventDefault()
    router.push('/notes')
    return
  }

  // 't' — Pomodoro toggle (dispatch custom event for Pomodoro.vue to pick up)
  if (e.key === 't' && !e.ctrlKey && !e.metaKey) {
    e.preventDefault()
    pomodoroToggleSignal.value++
    window.dispatchEvent(new CustomEvent('shortcut-pomodoro-toggle'))
    return
  }

  // Esc — close help modal
  if (e.key === 'Escape' && showHelpModal.value) {
    showHelpModal.value = false
    return
  }
}

// ========== Auto-initialize (once) ==========
function initShortcuts() {
  if (initialized) return
  initialized = true
  document.addEventListener('keydown', handleGlobalKeydown)
}

function destroyShortcuts() {
  document.removeEventListener('keydown', handleGlobalKeydown)
  resetPrefix()
  initialized = false
}

// ========== Public API ==========
export function useShortcuts() {
  return {
    // Reactive state
    pendingPrefix,
    showKeyHint,
    showHelpModal,

    // Actions
    openHelp: () => {
      showHelpModal.value = true
    },
    closeHelp: () => {
      showHelpModal.value = false
    },

    // Data
    allShortcuts: ALL_SHORTCUTS,
    navHintLabels: NAV_HINT_LABELS,
    defaultNavShortcuts: NAV_SHORTCUTS,

    // Customization
    getEffectiveNavShortcuts,
    setCustomShortcut,
    resetCustomShortcut,
    resetAllShortcuts,

    // Lifecycle
    initShortcuts,
    destroyShortcuts
  }
}
