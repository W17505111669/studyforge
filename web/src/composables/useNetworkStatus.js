import { ref, readonly, onMounted } from 'vue'

// 模块级单例状态
const isOnline = ref(navigator.onLine)
const wasOffline = ref(false) // 标记是否曾经离线过 (用于回连提示)
const offlineSince = ref(null) // 离线开始时间

let listenersAttached = false
const listeners = new Set()

function handleOnline() {
  isOnline.value = true
  if (offlineSince.value) {
    wasOffline.value = true
    offlineSince.value = null
  }
  notifyListeners('online')
}

function handleOffline() {
  isOnline.value = false
  offlineSince.value = Date.now()
  notifyListeners('offline')
}

function notifyListeners(event) {
  for (const cb of listeners) {
    try {
      cb(event)
    } catch {}
  }
}

/**
 * 网络状态 composable (模块级单例)
 * - isOnline: 当前是否在线
 * - wasOffline: 本次会话是否曾经离线
 * - offlineSince: 离线开始时间戳
 * - offlineDuration: 离线持续时长 (ms)
 * - onStatusChange: 注册状态变化回调
 * - isResponseFromCache: 检测响应是否来自离线缓存
 */
export function useNetworkStatus() {
  function attachListeners() {
    if (listenersAttached) return
    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)
    listenersAttached = true
  }

  onMounted(() => {
    attachListeners()
    // 同步最新状态
    isOnline.value = navigator.onLine
  })

  function onStatusChange(callback) {
    listeners.add(callback)
    return () => listeners.delete(callback)
  }

  function offlineDuration() {
    if (offlineSince.value) return Date.now() - offlineSince.value
    return 0
  }

  /**
   * 检测 axios/fetch 响应是否来自离线缓存
   * SW 会在离线缓存响应上附加 x-offline-cache: true 头
   */
  function isResponseFromCache(response) {
    if (!response) return false
    // axios response
    if (response.headers && response.headers['x-offline-cache'] === 'true') return true
    // native Response
    if (response.headers instanceof Headers && response.headers.get('x-offline-cache') === 'true') return true
    return false
  }

  function resetWasOffline() {
    wasOffline.value = false
  }

  return {
    isOnline: readonly(isOnline),
    wasOffline: readonly(wasOffline),
    offlineSince: readonly(offlineSince),
    offlineDuration,
    onStatusChange,
    isResponseFromCache,
    resetWasOffline
  }
}
