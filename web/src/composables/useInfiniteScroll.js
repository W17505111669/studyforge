/**
 * useInfiniteScroll — 无限滚动加载 composable
 *
 * 使用 IntersectionObserver 检测滚动到底部，自动触发加载更多。
 * 兼容现有 offset+limit 分页 API。
 *
 * @param {Object} options
 * @param {Function} options.onLoad - 加载回调 (offset, limit) => Promise<{ items, total }>
 * @param {Function} [options.onItems] - 新项加载回调 (items) => void，IntersectionObserver 触发时调用
 * @param {number} [options.limit=20] - 每页条数
 * @param {number} [options.threshold=0.1] - IntersectionObserver threshold (距底部比例)
 * @param {string} [options.rootMargin='200px'] - 提前触发距离
 * @returns {Object}
 */
import { ref, onBeforeUnmount } from 'vue'

export function useInfiniteScroll(options = {}) {
  const { onLoad, onItems, limit = 20, threshold = 0.1, rootMargin = '200px' } = options

  const loading = ref(false)
  const hasMore = ref(true)
  const error = ref(null)
  const total = ref(0)
  const offset = ref(0)

  let observer = null
  let sentinelEl = null
  let _scrollContainer = null
  let initialized = false

  // 加载下一页
  async function loadNext() {
    if (loading.value || !hasMore.value || error.value) return

    loading.value = true
    error.value = null

    try {
      const result = await onLoad(offset.value, limit)
      if (result && result.items) {
        const newItems = result.items
        if (result.total !== undefined) {
          total.value = result.total
        }
        offset.value += newItems.length
        // 如果没有更多数据或者返回数量少于 limit，标记结束
        if (newItems.length < limit || offset.value >= total.value) {
          hasMore.value = false
        }
        return newItems
      }
      return []
    } catch (e) {
      error.value = e.message || '加载失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  // 重置状态（用于筛选条件变化时重新加载）
  function reset() {
    offset.value = 0
    hasMore.value = true
    error.value = null
    total.value = 0
    loading.value = false
  }

  // 重试失败的请求
  function retry() {
    error.value = null
    return loadNext()
  }

  // 初始化 IntersectionObserver
  function init(sentinel, root) {
    if (initialized) return
    sentinelEl = sentinel
    _scrollContainer = root

    if (!sentinelEl) return

    observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0]
        if (entry && entry.isIntersecting) {
          loadNext()
            .then((items) => {
              if (items && items.length > 0 && onItems) {
                onItems(items)
              }
            })
            .catch(() => {})
        }
      },
      {
        root: root || null,
        rootMargin,
        threshold
      }
    )

    observer.observe(sentinelEl)
    initialized = true
  }

  // 清理
  function destroy() {
    if (observer) {
      observer.disconnect()
      observer = null
    }
    initialized = false
    sentinelEl = null
    _scrollContainer = null
  }

  onBeforeUnmount(() => {
    destroy()
  })

  return {
    loading,
    hasMore,
    error,
    total,
    offset,
    loadNext,
    reset,
    retry,
    init,
    destroy
  }
}

/**
 * useScrollToTop — 返回顶部按钮 composable
 *
 * @param {import('vue').Ref<HTMLElement|null>} scrollContainer - 滚动容器
 * @param {number} [showThreshold=600] - 显示按钮的滚动距离（px），约 2 屏
 */
export function useScrollToTop(scrollContainer, showThreshold = 600) {
  const showButton = ref(false)

  function onScroll() {
    if (scrollContainer.value) {
      showButton.value = scrollContainer.value.scrollTop > showThreshold
    } else {
      showButton.value = window.scrollY > showThreshold
    }
  }

  function scrollToTop() {
    if (scrollContainer.value) {
      scrollContainer.value.scrollTo({ top: 0, behavior: 'smooth' })
    } else {
      window.scrollTo({ top: 0, behavior: 'smooth' })
    }
  }

  function initScrollListener() {
    const target = scrollContainer.value || window
    target.addEventListener('scroll', onScroll, { passive: true })
  }

  function destroyScrollListener() {
    const target = scrollContainer.value || window
    target.removeEventListener('scroll', onScroll)
  }

  return {
    showButton,
    scrollToTop,
    initScrollListener,
    destroyScrollListener
  }
}
