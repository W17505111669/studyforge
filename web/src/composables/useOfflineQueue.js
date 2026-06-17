import { ref, readonly, computed } from 'vue'

// IndexedDB 配置
const DB_NAME = 'studyforge-offline'
const DB_VERSION = 1
const QUEUE_STORE = 'action_queue'

// 模块级单例状态
const pendingActions = ref([])
const syncing = ref(false)
const lastSyncResult = ref(null) // { success: number, failed: number, time: Date }
let dbInstance = null
let initialized = false
const listeners = new Set()

// ========== IndexedDB 操作 ==========

function openDB() {
  if (dbInstance) return Promise.resolve(dbInstance)

  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION)

    request.onupgradeneeded = (event) => {
      const db = event.target.result
      if (!db.objectStoreNames.contains(QUEUE_STORE)) {
        const store = db.createObjectStore(QUEUE_STORE, { keyPath: 'id', autoIncrement: true })
        store.createIndex('type', 'type', { unique: false })
        store.createIndex('createdAt', 'createdAt', { unique: false })
      }
    }

    request.onsuccess = (event) => {
      dbInstance = event.target.result
      resolve(dbInstance)
    }

    request.onerror = (event) => {
      console.error('[OfflineQueue] IndexedDB 打开失败:', event.target.error)
      reject(event.target.error)
    }
  })
}

async function loadQueue() {
  try {
    const db = await openDB()
    return new Promise((resolve, reject) => {
      const tx = db.transaction(QUEUE_STORE, 'readonly')
      const store = tx.objectStore(QUEUE_STORE)
      const request = store.getAll()
      request.onsuccess = () => {
        // 按创建时间排序
        const items = request.result.sort((a, b) => a.createdAt - b.createdAt)
        pendingActions.value = items
        resolve(items)
      }
      request.onerror = () => reject(request.error)
    })
  } catch (err) {
    console.error('[OfflineQueue] 加载队列失败:', err)
    return []
  }
}

async function addToDB(action) {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(QUEUE_STORE, 'readwrite')
    const store = tx.objectStore(QUEUE_STORE)
    const request = store.add(action)
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

async function removeFromDB(id) {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(QUEUE_STORE, 'readwrite')
    const store = tx.objectStore(QUEUE_STORE)
    const request = store.delete(id)
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

async function clearDB() {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(QUEUE_STORE, 'readwrite')
    const store = tx.objectStore(QUEUE_STORE)
    const request = store.clear()
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

// ========== API 请求重放 ==========

/**
 * 重放单个离线操作
 * 返回 { success: boolean, error?: string }
 */
async function replayAction(action) {
  const token = localStorage.getItem('token')
  if (!token) return { success: false, error: '未登录' }

  try {
    const response = await fetch(`/api${action.path}`, {
      method: action.method,
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`
      },
      body: action.body ? JSON.stringify(action.body) : undefined
    })

    if (response.ok) {
      return { success: true }
    }

    // 401 不再重试 (token 过期)
    if (response.status === 401) {
      return { success: false, error: '认证已过期' }
    }

    // 409/404 说明资源已变更, 标记为成功 (无需重放)
    if (response.status === 409 || response.status === 404) {
      return { success: true, skipped: true }
    }

    return { success: false, error: `HTTP ${response.status}` }
  } catch (err) {
    return { success: false, error: err.message }
  }
}

// ========== 同步 ==========

async function syncQueue() {
  if (syncing.value || !navigator.onLine) return
  if (pendingActions.value.length === 0) return

  syncing.value = true
  let success = 0
  let failed = 0

  const actions = [...pendingActions.value]

  for (const action of actions) {
    const result = await replayAction(action)

    if (result.success) {
      await removeFromDB(action.id)
      success++
    } else if (result.error === '认证已过期' || result.error === '未登录') {
      // 认证问题, 停止同步
      failed += actions.length - success - failed
      break
    } else {
      // 网络或其他错误, 保留在队列中, 停止后续同步
      failed++
      break
    }
  }

  // 刷新队列
  await loadQueue()

  lastSyncResult.value = { success, failed, time: new Date() }
  syncing.value = false

  notifyListeners('synced', { success, failed })

  return lastSyncResult.value
}

function notifyListeners(event, data) {
  for (const cb of listeners) {
    try {
      cb(event, data)
    } catch {}
  }
}

// ========== 支持的离线操作类型定义 ==========

const ACTION_TYPES = {
  CARD_REVIEW: 'card_review', // 卡片复习
  CARD_BOOKMARK: 'card_bookmark', // 书签切换
  CARD_NOTE: 'card_note', // 卡片笔记
  NOTE_SAVE: 'note_save', // 笔记保存
  NOTE_DELETE: 'note_delete', // 笔记删除
  TASK_TOGGLE: 'task_toggle', // 任务切换
  NOTIFICATION_READ: 'notification_read', // 通知已读
  MISTAKE_REVIEW: 'mistake_review', // 错题标记复习
  CUSTOM: 'custom' // 自定义操作
}

const ACTION_LABELS = {
  [ACTION_TYPES.CARD_REVIEW]: '卡片复习',
  [ACTION_TYPES.CARD_BOOKMARK]: '书签切换',
  [ACTION_TYPES.CARD_NOTE]: '卡片笔记',
  [ACTION_TYPES.NOTE_SAVE]: '笔记保存',
  [ACTION_TYPES.NOTE_DELETE]: '笔记删除',
  [ACTION_TYPES.TASK_TOGGLE]: '任务切换',
  [ACTION_TYPES.NOTIFICATION_READ]: '通知已读',
  [ACTION_TYPES.MISTAKE_REVIEW]: '错题标记',
  [ACTION_TYPES.CUSTOM]: '操作'
}

// ========== Composable 导出 ==========

/**
 * 离线操作队列 composable (模块级单例)
 * IndexedDB 持久化 + 联网自动同步
 */
export function useOfflineQueue() {
  // 初始化 (仅首次)
  async function init() {
    if (initialized) return
    initialized = true

    await loadQueue()

    // 监听 SW 同步消息
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.addEventListener('message', (event) => {
        if (event.data?.type === 'SYNC_QUEUE') {
          syncQueue()
        }
      })
    }

    // 联网时自动同步
    window.addEventListener('online', () => {
      setTimeout(() => syncQueue(), 1000) // 延迟 1s 等网络稳定
    })
  }

  /**
   * 添加操作到离线队列
   * @param {string} type - 操作类型 (ACTION_TYPES)
   * @param {string} method - HTTP 方法
   * @param {string} path - API 路径 (不含 /api 前缀)
   * @param {object} body - 请求体
   * @param {string} label - 用户可读描述
   */
  async function enqueue(type, method, path, body = null, label = '') {
    const action = {
      type,
      method,
      path,
      body,
      label: label || ACTION_LABELS[type] || type,
      createdAt: Date.now()
    }

    try {
      const id = await addToDB(action)
      action.id = id
      pendingActions.value = [...pendingActions.value, action]
      notifyListeners('enqueued', action)
      return id
    } catch (err) {
      console.error('[OfflineQueue] 入队失败:', err)
      return null
    }
  }

  /**
   * 移除指定操作
   */
  async function remove(id) {
    await removeFromDB(id)
    pendingActions.value = pendingActions.value.filter((a) => a.id !== id)
  }

  /**
   * 清空所有待处理操作
   */
  async function clearAll() {
    await clearDB()
    pendingActions.value = []
  }

  function onQueueEvent(callback) {
    listeners.add(callback)
    return () => listeners.delete(callback)
  }

  const pendingCount = computed(() => pendingActions.value.length)

  return {
    pendingActions: readonly(pendingActions),
    pendingCount,
    syncing: readonly(syncing),
    lastSyncResult: readonly(lastSyncResult),
    init,
    enqueue,
    remove,
    clearAll,
    syncQueue,
    onQueueEvent,
    ACTION_TYPES
  }
}
