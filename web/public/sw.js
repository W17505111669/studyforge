// StudyForge Pro Service Worker
// 缓存策略: 静态资源 Cache-First, API/导航 Network-First + 离线回退

const CACHE_NAME = 'studyforge-v1'
const STATIC_CACHE = 'studyforge-static-v1'

// 预缓存的核心资源
const PRECACHE_URLS = [
  '/',
  '/offline.html',
  '/manifest.json',
  '/icon.svg',
  '/icon-192.png',
  '/icon-512.png',
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
  const keepCaches = [CACHE_NAME, STATIC_CACHE]
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

  // API 请求: Network-First, 失败则返回离线页
  if (url.pathname.startsWith('/api/') || url.pathname.startsWith('/ws')) {
    event.respondWith(networkFirstWithFallback(request))
    return
  }

  // 静态资源 (JS/CSS/字体/图片): Cache-First
  if (isStaticAsset(url.pathname)) {
    event.respondWith(cacheFirstWithNetwork(request))
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

// Network-First → 缓存回退
async function networkFirstWithFallback(request) {
  try {
    const response = await fetch(request)
    // 成功响应存入缓存
    if (response.ok) {
      const cache = await caches.open(CACHE_NAME)
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

// Cache-First → 网络回退 + 缓存新资源
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
    // 图片资源返回占位符
    if (request.destination === 'image') {
      return new Response('', { status: 404 })
    }
    return new Response('', { status: 503 })
  }
}

// Network-First → 离线回退页面
async function networkFirstWithOfflinePage(request) {
  try {
    const response = await fetch(request)
    // 更新缓存中的页面
    if (response.ok) {
      const cache = await caches.open(CACHE_NAME)
      cache.put(request, response.clone())
    }
    return response
  } catch {
    // 尝试返回缓存页面
    const cached = await caches.match(request)
    if (cached) return cached

    // 回退到离线页面
    const offlinePage = await caches.match('/offline.html')
    return offlinePage || new Response('StudyForge Pro 离线中', {
      status: 503,
      headers: { 'Content-Type': 'text/plain; charset=utf-8' },
    })
  }
}

// Network-First → 缓存回退
async function networkFirstWithCache(request) {
  try {
    const response = await fetch(request)
    if (response.ok) {
      const cache = await caches.open(CACHE_NAME)
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
  return /\.(js|css|woff2?|ttf|eot|svg|png|jpg|jpeg|gif|ico|webp)(\?.*)?$/i.test(pathname)
    || pathname.startsWith('/assets/')
}

// ===== 消息通信: 手动更新缓存 =====
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
})
