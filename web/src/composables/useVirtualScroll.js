/**
 * useVirtualScroll — 轻量虚拟滚动 composable
 *
 * 仅渲染可视区域内的列表项，适用于大列表性能优化。
 * 支持固定行高模式（Upload/Mistakes/Chat）和网格行模式（Cards）。
 *
 * @param {import('vue').Ref<HTMLElement|null>} containerRef - 滚动容器 DOM 引用
 * @param {import('vue').Ref<number>|number} totalCount - 列表总条数
 * @param {Object} options
 * @param {number} options.itemHeight - 每项预估高度（px）
 * @param {number} [options.buffer=5] - 上下各缓冲条数
 * @param {number} [options.threshold=100] - 超过此数量才启用虚拟滚动
 * @returns {Object}
 */
import { ref, computed, watch, onMounted, onBeforeUnmount, toValue } from 'vue'

export function useVirtualScroll(containerRef, totalCount, options = {}) {
  const { itemHeight = 120, buffer = 5, threshold = 100 } = options

  const scrollTop = ref(0)
  const viewportHeight = ref(600)
  const _rafId = ref(null)

  // 是否启用虚拟滚动（总数超过阈值时）
  const shouldVirtualize = computed(() => {
    const count = toValue(totalCount)
    return count > threshold
  })

  // 总高度（占位用）
  const totalHeight = computed(() => {
    const count = toValue(totalCount)
    return count * itemHeight
  })

  // 可见起始索引
  const startIndex = computed(() => {
    if (!shouldVirtualize.value) return 0
    const count = toValue(totalCount)
    const idx = Math.floor(scrollTop.value / itemHeight) - buffer
    return Math.max(0, Math.min(idx, count - 1))
  })

  // 可见结束索引
  const endIndex = computed(() => {
    if (!shouldVirtualize.value) return toValue(totalCount)
    const count = toValue(totalCount)
    const visibleCount = Math.ceil(viewportHeight.value / itemHeight)
    const idx = Math.floor(scrollTop.value / itemHeight) + visibleCount + buffer
    return Math.max(0, Math.min(idx, count))
  })

  // 顶部占位高度
  const topSpacerHeight = computed(() => startIndex.value * itemHeight)

  // 底部占位高度
  const bottomSpacerHeight = computed(() => {
    const count = toValue(totalCount)
    return Math.max(0, (count - endIndex.value) * itemHeight)
  })

  // 滚动事件处理（使用 rAF 节流）
  const onScroll = () => {
    if (_rafId.value) return
    _rafId.value = requestAnimationFrame(() => {
      _rafId.value = null
      if (containerRef.value) {
        scrollTop.value = containerRef.value.scrollTop
      }
    })
  }

  // 测量视口高度
  const measure = () => {
    if (containerRef.value) {
      viewportHeight.value = containerRef.value.clientHeight || 600
    }
  }

  // 滚动到指定索引
  const scrollToIndex = (index) => {
    if (containerRef.value) {
      containerRef.value.scrollTop = index * itemHeight
    }
  }

  // resize observer
  let resizeObserver = null

  onMounted(() => {
    const el = containerRef.value
    if (!el) return
    el.addEventListener('scroll', onScroll, { passive: true })
    measure()

    // 监听容器大小变化
    if (typeof ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(() => measure())
      resizeObserver.observe(el)
    }

    // 开发模式性能监控
    if (import.meta.env.DEV) {
      const count = toValue(totalCount)
      if (count > threshold) {
        console.time(`[VirtualScroll] render ${count} items`)
        requestAnimationFrame(() => {
          console.timeEnd(`[VirtualScroll] render ${count} items`)
        })
      }
    }
  })

  onBeforeUnmount(() => {
    const el = containerRef.value
    if (el) {
      el.removeEventListener('scroll', onScroll)
    }
    if (_rafId.value) {
      cancelAnimationFrame(_rafId.value)
    }
    if (resizeObserver) {
      resizeObserver.disconnect()
      resizeObserver = null
    }
  })

  // 总数变化时重新计算
  watch(
    () => toValue(totalCount),
    () => {
      if (import.meta.env.DEV) {
        const count = toValue(totalCount)
        if (count > threshold) {
          console.time(`[VirtualScroll] re-render ${count} items`)
          requestAnimationFrame(() => {
            console.timeEnd(`[VirtualScroll] re-render ${count} items`)
          })
        }
      }
    }
  )

  return {
    startIndex,
    endIndex,
    topSpacerHeight,
    bottomSpacerHeight,
    totalHeight,
    shouldVirtualize,
    onScroll,
    scrollToIndex,
    measure
  }
}

/**
 * useGridVirtualScroll — 网格布局虚拟滚动（Cards 等网格列表专用）
 *
 * 将卡片按行分组，每行包含 cols 张卡片，对行进行虚拟化。
 *
 * @param {import('vue').Ref<HTMLElement|null>} containerRef
 * @param {import('vue').Ref<number>|number} totalCount
 * @param {Object} options
 * @param {number} options.rowHeight - 每行高度（px），含 gap
 * @param {import('vue').Ref<number>|number} options.cols - 当前列数
 * @param {number} [options.buffer=3] - 上下各缓冲行数
 * @param {number} [options.threshold=200] - 超过此卡片数才启用
 */
export function useGridVirtualScroll(containerRef, totalCount, options = {}) {
  const { rowHeight = 340, cols = ref(3), buffer = 3, threshold = 200 } = options

  const scrollTop = ref(0)
  const viewportHeight = ref(600)
  const _rafId = ref(null)

  const colCount = computed(() => toValue(cols))

  const totalRows = computed(() => {
    const count = toValue(totalCount)
    return Math.ceil(count / colCount.value)
  })

  const shouldVirtualize = computed(() => {
    const count = toValue(totalCount)
    return count > threshold
  })

  const totalHeight = computed(() => totalRows.value * rowHeight)

  const startRow = computed(() => {
    if (!shouldVirtualize.value) return 0
    return Math.max(0, Math.floor(scrollTop.value / rowHeight) - buffer)
  })

  const endRow = computed(() => {
    if (!shouldVirtualize.value) return totalRows.value
    const visibleRows = Math.ceil(viewportHeight.value / rowHeight)
    return Math.min(totalRows.value, Math.floor(scrollTop.value / rowHeight) + visibleRows + buffer)
  })

  // 可见卡片的起始和结束索引
  const startIndex = computed(() => startRow.value * colCount.value)
  const endIndex = computed(() => {
    const count = toValue(totalCount)
    return Math.min(endRow.value * colCount.value, count)
  })

  const topSpacerHeight = computed(() => startRow.value * rowHeight)
  const bottomSpacerHeight = computed(() => {
    return Math.max(0, (totalRows.value - endRow.value) * rowHeight)
  })

  const onScroll = () => {
    if (_rafId.value) return
    _rafId.value = requestAnimationFrame(() => {
      _rafId.value = null
      if (containerRef.value) {
        scrollTop.value = containerRef.value.scrollTop
      }
    })
  }

  const measure = () => {
    if (containerRef.value) {
      viewportHeight.value = containerRef.value.clientHeight || 600
    }
  }

  let resizeObserver = null

  onMounted(() => {
    const el = containerRef.value
    if (!el) return
    el.addEventListener('scroll', onScroll, { passive: true })
    measure()

    if (typeof ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(() => measure())
      resizeObserver.observe(el)
    }

    if (import.meta.env.DEV) {
      const count = toValue(totalCount)
      if (count > threshold) {
        console.time(`[GridVirtualScroll] render ${count} cards`)
        requestAnimationFrame(() => {
          console.timeEnd(`[GridVirtualScroll] render ${count} cards`)
        })
      }
    }
  })

  onBeforeUnmount(() => {
    const el = containerRef.value
    if (el) el.removeEventListener('scroll', onScroll)
    if (_rafId.value) cancelAnimationFrame(_rafId.value)
    if (resizeObserver) {
      resizeObserver.disconnect()
      resizeObserver = null
    }
  })

  return {
    startIndex,
    endIndex,
    topSpacerHeight,
    bottomSpacerHeight,
    totalHeight,
    shouldVirtualize,
    onScroll,
    measure,
    colCount
  }
}
