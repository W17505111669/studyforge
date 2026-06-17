// StudyForge Pro Service Worker v2
// 缓存策略: 静态 Cache-First / API Network-First+缓存回退 / 字体+图片 StaleWhileRevalidate / 导航 Network-First+离线回退
// 离线可用端点白名单: Dashboard/Cards/Notes 等 GET 数据可缓存供离线读取

const CACHE_VERSION = 'studyforge-v2'
const STATIC_CACHE = 'studyforge-static-v2'
const DATA_CACHE = 'studyforge-data-v2'

// 预缓存的核心资源
const PRECACHE_URLS = [
  '/',
  '/offline.html',
  '/manifest.json',
  '/icon.svg',
  '/icon-192.png',
  '/icon-512.png',
]

// 离线可用的 API 端点白名单 (GET 请求, 缓存最新成功响应)
const CACHEABLE_API_PATTERNS = [
  /^\/api\/materials(\?.*)?$/,
  /^\/api\/cards(\?.*)?$/,
  /^\/api\/notes(\?.*)?$/,
  /^\/api\/notes\/folders(\?.*)?$/,
  /^\/api\/dashboard\/metrics$/,
  /^\/api\/dashboard\/activity$/,
  /^\/api\/stats$/,
  /^\/api\/stats\/calendar(\?.*)?$/,
  /^\/api\/streaks$/,
  /^\/api\/achievements$/,
  /^\/api\/pomodoro\/stats$/,
  /^\/api\/goals(\?.*)?$/,
  /^\/api\/goals\/progress$/,
  /^\/api\/daily-tasks(\?.*)?$/,
  /^\/api\/notifications\/unread-count$/,
  /^\/api\/friends\/count$/,
  /^\/api\/conversations(\?.*)?$/,
  /^\/api\/quizzes(\?.*)?$/,
  /^\/api\/mistakes(\?.*)?$/,
  /^\/api\/mistakes\/stats$/,
  /^\/api\/cards\/due(\?.*)?$/,
  /^\/api\/tags$/,
  /^\/api\/leaderboard(\?.*)?$/,
  /^\/api\/exams$/,
]

// 离线不可用的 API 端点 (写操作/AI 调用, 返回明确离线提示)
const OFFLINE_UNAVAILABLE_PATTERNS = [
  /^\/api\/chat/,
  /^\/api\/materials\/.*\/analyze$/,
  /^\/api\/materials\/upload/,
  /^\/api\/materials\/batch-analyze$/,
  /^\/api\/debate/,
  /^\/api\/explain/,
  /^\/api\/exam\/generate$/,
  /^\/api\/learning-path/,
  /^\/api\/import\//,
  /^\/api\/seed$/,
  /^\/ws/,
]

// ===== Install: 预缓存核心资源 =====
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(STATIC_CACHE)
      .then((cache) => cache.addAll(PRECACHE_URLS))
      .then(() => self.skipWaiting())
  )
})

// ===== Activate: 清理旧缓存 =====
self.addEventListener('activate', (event) => {
  const keepCaches = [CACHE_VERSION, STATIC_CACHE, DATA_CACHE]
  event.waitUntil(
    caches.keys()
      .then((keys) =>
        Promise.all(
          keys
            .filter((key) => !keepCaches.includes(key))
            .map((key) => caches.delete(key))
        )
      )
      .then(() => self.clients.claim())
  )
})

// ===== Fetch: 智能路由策略 =====
self.addEventListener('fetch', (event) => {
  const { request } = event
  const url = new URL(request.url)

  // 跳过非同源请求
  if (url.origin !== location.origin) return

  // WebSocket 不拦截
  if (request.url.startsWith('ws://') || request.url.startsWith('wss://')) return

  // API 请求
  if (url.pathname.startsWith('/api/')) {
    // POST/PUT/DELETE 写操作: Network-Only, 离线返回排队提示
    if (request.method !== 'GET') {
      event.respondWith(networkOnlyWithOfflineHint(request))
      return
    }

    // 离线不可用端点
    if (isOfflineUnavailable(url.pathname)) {
      event.respondWith(networkFirstWithOfflineHint(request))
      return
    }

    // 可缓存 GET 端点: Network-First + 更新数据缓存
    if (isCacheableApi(url.pathname)) {
      event.respondWith(networkFirstWithDataCache(request))
      return
    }

    // 其他 API GET: Network-First + 缓存回退
    event.respondWith(networkFirstWithFallback(request))
    return
  }

  // 静态资源 (JS/CSS/字体): Cache-First
  if (isStaticAsset(url.pathname)) {
    event.respondWith(cacheFirstWithNetwork(request))
    return
  }

  // 图片和字体: StaleWhileRevalidate (先返回缓存, 后台更新)
  if (isImageOrFont(url.pathname)) {
    event.respondWith(staleWhileRevalidate(request))
    return
  }

  // 导航请求: Network-First + 离线回退页
  if (request.mode === 'navigate') {
    event.respondWith(networkFirstWithOfflinePage(request))
    return
  }

  // 其他资源: Network-First + 缓存回退
  event.respondWith(networkFirstWithCache(request))
})

// ===== 策略实现 =====

// Network-First → 缓存回退 (通用)
async function networkFirstWithFallback(request) {
  try {
    const response = await fetch(request)
    if (response.ok) {
      const cache = await caches.open(CACHE_VERSION)
      cache.put(request, response.clone())
    }
    return response
  } catch {
    const cached = await caches.match(request)
    if (cached) return cached
    return new Response(JSON.stringify({ error: 'offline' }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' },
    })
  }
}

// Network-First → 数据缓存 (可缓存 API GET 端点)
// 成功时更新缓存, 失败时返回缓存数据附加 x-offline: true 头
async function networkFirstWithDataCache(request) {
  try {
    const response = await fetch(request)
    if (response.ok) {
      const cache = await caches.open(DATA_CACHE)
      cache.put(request, response.clone())
    }
    return response
  } catch {
    const cached = await caches.match(request)
    if (cached) {
      // 附加离线标记头, 前端可检测
      const headers = new Headers(cached.headers)
      headers.set('x-offline-cache', 'true')
      return new Response(cached.body, {
        status: cached.status,
        statusText: cached.statusText,
        headers,
      })
    }
    return new Response(JSON.stringify({ error: 'offline', data: [], total: 0 }), {
      status: 503,
      headers: { 'Content-Type': 'application/json', 'x-offline-cache': 'true' },
    })
  }
}

// Network-Only → 离线返回排队提示 (写操作)
async function networkOnlyWithOfflineHint(request) {
  try {
    return await fetch(request)
  } catch {
    return new Response(JSON.stringify({
      error: 'offline',
      message: '当前处于离线状态，操作已排队，联网后将自动同步',
      queued: true,
    }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' },
    })
  }
}

// Network-First → 离线不可用提示 (AI/实时功能)
async function networkFirstWithOfflineHint(request) {
  try {
    const response = await fetch(request)
    if (response.ok) {
      const cache = await caches.open(CACHE_VERSION)
      cache.put(request, response.clone())
    }
    return response
  } catch {
    return new Response(JSON.stringify({
      error: 'offline',
      message: '此功能需要网络连接',
      requires_network: true,
    }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' },
    })
  }
}

// Cache-First → 网络回退 (静态资源)
async function cacheFirstWithNetwork(request) {
  const cached = await caches.match(request)
  if (cached) return cached

  try {
    const response = await fetch(request)
    if (response.ok) {
      const cache = await caches.open(STATIC_CACHE)
      cache.put(request, response.clone())
    }
    return response
  } catch {
    if (request.destination === 'image') {
      return new Response('', { status: 404 })
    }
    return new Response('', { status: 503 })
  }
}

// StaleWhileRevalidate: 先返回缓存, 后台更新
async function staleWhileRevalidate(request) {
  const cache = await caches.open(STATIC_CACHE)
  const cached = await cache.match(request)

  // 后台更新 Promise
  const fetchPromise = fetch(request).then((response) => {
    if (response.ok) {
      cache.put(request, response.clone())
    }
    return response
  }).catch(() => null)

  // 有缓存先返回缓存, 后台静默更新; 无缓存则等待网络
  if (cached) {
    // 不 await, 后台更新
    fetchPromise.catch(() => {})
    return cached
  }

  const response = await fetchPromise
  return response || new Response('', { status: 503 })
}

// Network-First → 离线回退页面
async function networkFirstWithOfflinePage(request) {
  try {
    const response = await fetch(request)
    if (response.ok) {
      const cache = await caches.open(CACHE_VERSION)
      cache.put(request, response.clone())
    }
    return response
  } catch {
    const cached = await caches.match(request)
    if (cached) return cached

    const offlinePage = await caches.match('/offline.html')
    return offlinePage || new Response('StudyForge Pro 离线中', {
      status: 503,
      headers: { 'Content-Type': 'text/plain; charset=utf-8' },
    })
  }
}

// Network-First → 缓存回退 (其他资源)
async function networkFirstWithCache(request) {
  try {
    const response = await fetch(request)
    if (response.ok) {
      const cache = await caches.open(CACHE_VERSION)
      cache.put(request, response.clone())
    }
    return response
  } catch {
    const cached = await caches.match(request)
    return cached || new Response('', { status: 503 })
  }
}

// ===== 辅助函数 =====

function isStaticAsset(pathname) {
  return /\.(js|css|woff2?|ttf|eot)(\?.*)?$/i.test(pathname)
    || pathname.startsWith('/assets/')
}

function isImageOrFont(pathname) {
  return /\.(png|jpg|jpeg|gif|ico|webp|svg)(\?.*)?$/i.test(pathname)
}

function isCacheableApi(pathname) {
  return CACHEABLE_API_PATTERNS.some((pattern) => pattern.test(pathname))
}

function isOfflineUnavailable(pathname) {
  return OFFLINE_UNAVAILABLE_PATTERNS.some((pattern) => pattern.test(pathname))
}

// ===== 消息通信 =====
self.addEventListener('message', (event) => {
  if (event.data === 'SKIP_WAITING') {
    self.skipWaiting()
  }
  if (event.data === 'CLEAR_CACHE') {
    event.waitUntil(
      caches.keys().then((keys) =>
        Promise.all(keys.map((key) => caches.delete(key)))
      ).then(() => {
        event.source.postMessage({ type: 'CACHE_CLEARED' })
      })
    )
  }
  // 查询离线缓存状态
  if (event.data === 'CACHE_STATUS') {
    event.waitUntil(
      Promise.all([
        caches.open(DATA_CACHE).then((c) => c.keys()),
        caches.open(STATIC_CACHE).then((c) => c.keys()),
      ]).then(([dataKeys, staticKeys]) => {
        if (event.source) {
          event.source.postMessage({
            type: 'CACHE_STATUS',
            dataEntries: dataKeys.length,
            staticEntries: staticKeys.length,
          })
        }
      })
    )
  }
  // 预缓存指定 URL (前端触发)
  if (event.data && event.data.type === 'PRECACHE_URLS' && Array.isArray(event.data.urls)) {
    event.waitUntil(
      caches.open(DATA_CACHE)
        .then((cache) =>
          Promise.all(
            event.data.urls.map((url) =>
              fetch(url).then((res) => {
                if (res.ok) cache.put(url, res)
              }).catch(() => {}) // 静默失败
            )
          )
        )
        .then(() => {
          if (event.source) {
            event.source.postMessage({ type: 'PRECACHE_DONE' })
          }
        })
    )
  }
})

// ===== 后台同步 (如果支持) =====
self.addEventListener('sync', (event) => {
  if (event.tag === 'sync-offline-queue') {
    event.waitUntil(syncOfflineQueue())
  }
})

// 离线队列同步 (由 SW 端触发)
async function syncOfflineQueue() {
  // 通知所有客户端执行同步
  const clients = await self.clients.matchAll({ type: 'window' })
  for (const client of clients) {
    client.postMessage({ type: 'SYNC_QUEUE' })
  }
}
