<template>
  <Teleport to="body">
    <Transition name="search-fade">
      <div v-if="open" class="fixed inset-0 z-[100] flex items-start justify-center pt-[12vh]" @click.self="close">
        <!-- 背景遮罩 -->
        <div class="absolute inset-0 bg-black/50 dark:bg-black/70 backdrop-blur-sm"></div>

        <!-- 搜索面板 -->
        <div
          class="relative w-full max-w-xl bg-white dark:bg-gray-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden"
          @keydown.esc="close"
          @keydown.up.prevent="moveUp"
          @keydown.down.prevent="moveDown"
          @keydown.enter.prevent="selectCurrent"
        >
          <!-- 搜索输入区 -->
          <div class="flex items-center gap-3 px-5 py-4 border-b border-gray-200 dark:border-gray-700">
            <svg
              class="w-5 h-5 text-gray-400 dark:text-gray-500 flex-shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
            <input
              ref="searchInput"
              v-model="query"
              type="text"
              placeholder="搜索材料、卡片、练习题、标签..."
              class="flex-1 bg-transparent text-gray-900 dark:text-gray-100 text-base outline-none placeholder-gray-400 dark:placeholder-gray-500"
              @input="onInput"
            />
            <kbd
              class="hidden sm:inline-flex items-center px-2 py-0.5 rounded border border-gray-300 dark:border-gray-600 text-xs text-gray-400 dark:text-gray-500 font-mono"
            >
              Esc
            </kbd>
          </div>

          <!-- 结果区域 -->
          <div class="max-h-[45vh] overflow-y-auto">
            <!-- 加载状态 -->
            <div v-if="loading" class="px-5 py-8 text-center">
              <div
                class="inline-block w-5 h-5 border-2 border-primary-500 border-t-transparent rounded-full animate-spin"
              ></div>
              <p class="text-sm text-gray-500 dark:text-gray-400 mt-2">搜索中...</p>
            </div>

            <!-- 无结果 -->
            <div v-else-if="searched && results.length === 0" class="px-5 py-10 text-center">
              <svg
                class="w-12 h-12 mx-auto text-gray-300 dark:text-gray-600 mb-3"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="1.5"
                  d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                未找到与「
                <span class="text-gray-700 dark:text-gray-300">{{ query }}</span>
                」相关的结果
              </p>
            </div>

            <!-- 初始状态提示 -->
            <div v-else-if="!searched && query.length < 1" class="px-5 py-8 text-center">
              <p class="text-sm text-gray-400 dark:text-gray-500">输入关键词开始搜索</p>
              <div class="flex items-center justify-center gap-2 mt-3 text-xs text-gray-400 dark:text-gray-500">
                <span class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded">材料标题</span>
                <span class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded">概念名</span>
                <span class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded">题目</span>
                <span class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded">标签</span>
              </div>
            </div>

            <!-- 结果列表 -->
            <template v-else-if="results.length > 0">
              <!-- 按类型分组 -->
              <template v-for="group in groupedResults" :key="group.type">
                <div
                  class="px-5 py-2 text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider bg-gray-50 dark:bg-gray-800/50"
                >
                  {{ group.label }}
                  <span class="text-gray-300 dark:text-gray-600 ml-1">({{ group.items.length }})</span>
                </div>
                <div
                  v-for="item in group.items"
                  :key="item.id"
                  :class="resultClass(item)"
                  class="flex items-center gap-3 px-5 py-3 cursor-pointer transition-colors duration-100 border-l-2"
                  :style="{ borderLeftColor: item._flatIndex === activeIndex ? 'rgb(99, 102, 241)' : 'transparent' }"
                  @click="navigateTo(item)"
                  @mouseenter="activeIndex = item._flatIndex"
                >
                  <!-- 类型图标 -->
                  <div
                    :class="iconClass(item.type)"
                    class="w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0"
                  >
                    <!-- 材料图标 -->
                    <svg
                      v-if="item.type === 'material'"
                      class="w-4 h-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                      />
                    </svg>
                    <!-- 卡片图标 -->
                    <svg
                      v-else-if="item.type === 'card'"
                      class="w-4 h-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
                      />
                    </svg>
                    <!-- 题目图标 -->
                    <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                  </div>
                  <!-- 文本内容 -->
                  <div class="flex-1 min-w-0">
                    <p
                      class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate"
                      v-html="highlightText(item.title)"
                    ></p>
                    <p class="text-xs text-gray-500 dark:text-gray-400 truncate mt-0.5">
                      <span v-html="highlightText(item.subtitle)"></span>
                      <span v-if="item.material_title" class="text-primary-500 dark:text-primary-400">
                        ·
                        <span v-html="highlightText(item.material_title)"></span>
                      </span>
                    </p>
                  </div>
                  <!-- 相关度徽章 -->
                  <span
                    v-if="item.relevance >= 100"
                    class="flex-shrink-0 px-1.5 py-0.5 text-[10px] font-medium bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400 rounded"
                  >
                    精确
                  </span>
                  <!-- 跳转箭头 -->
                  <svg
                    class="w-4 h-4 text-gray-300 dark:text-gray-600 flex-shrink-0"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                  </svg>
                </div>
              </template>
            </template>
          </div>

          <!-- 高级搜索面板 -->
          <AdvancedSearch ref="advancedSearchRef" @update:filters="onFiltersUpdate" @search="doSearchWithFilters" />

          <!-- 底部快捷键提示 -->
          <div
            class="px-5 py-2.5 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 flex items-center gap-4 text-xs text-gray-400 dark:text-gray-500"
          >
            <span class="flex items-center gap-1">
              <kbd class="px-1.5 py-0.5 rounded border border-gray-300 dark:border-gray-600 font-mono">↑</kbd>
              <kbd class="px-1.5 py-0.5 rounded border border-gray-300 dark:border-gray-600 font-mono">↓</kbd>
              导航
            </span>
            <span class="flex items-center gap-1">
              <kbd class="px-1.5 py-0.5 rounded border border-gray-300 dark:border-gray-600 font-mono">↵</kbd>
              打开
            </span>
            <span class="flex items-center gap-1">
              <kbd class="px-1.5 py-0.5 rounded border border-gray-300 dark:border-gray-600 font-mono">Esc</kbd>
              关闭
            </span>
            <span v-if="results.length > 0" class="ml-auto">共 {{ results.length }} 条结果</span>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { globalSearch } from '../api/client'
import AdvancedSearch from './AdvancedSearch.vue'

const router = useRouter()
const open = ref(false)
const query = ref('')
const loading = ref(false)
const searched = ref(false)
const results = ref([])
const activeIndex = ref(0)
const searchInput = ref(null)
const advancedSearchRef = ref(null)
const currentFilters = ref({})

let debounceTimer = null

const typeLabels = {
  material: '材料',
  card: '知识卡片',
  quiz: '练习题'
}

const typeOrder = { material: 0, card: 1, quiz: 2 }

// 按类型分组结果
const groupedResults = computed(() => {
  const groups = {}
  results.value.forEach((item) => {
    if (!groups[item.type]) {
      groups[item.type] = []
    }
    groups[item.type].push(item)
  })

  return Object.keys(groups)
    .sort((a, b) => (typeOrder[a] ?? 99) - (typeOrder[b] ?? 99))
    .map((type) => ({
      type,
      label: typeLabels[type] || type,
      items: groups[type].map((item, idx) => ({ ...item, _groupIdx: idx }))
    }))
})

// 扁平化结果列表（用于键盘导航）
const flatResults = computed(() => {
  const flat = []
  groupedResults.value.forEach((group) => {
    group.items.forEach((item) => {
      flat.push(item)
    })
  })
  // 给每个 item 加上 _flatIndex
  flat.forEach((item, i) => {
    item._flatIndex = i
  })
  return flat
})

// 高亮匹配文本
function highlightText(text) {
  if (!text) return ''
  const q = query.value.trim()
  if (!q) return escapeHtml(text)
  // 转义搜索词中的正则特殊字符
  const escaped = q.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  // 先转义 HTML，然后替换匹配词为 mark 标签
  return escapeHtml(text).replace(
    new RegExp(`(${escapeHtml(escaped)})`, 'gi'),
    '<mark class="bg-yellow-200 dark:bg-yellow-700/40 text-yellow-900 dark:text-yellow-200 px-0.5 rounded">$1</mark>'
  )
}

function escapeHtml(str) {
  if (!str) return ''
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}

function openSearch() {
  open.value = true
  query.value = ''
  results.value = []
  searched.value = false
  activeIndex.value = 0
  nextTick(() => {
    searchInput.value?.focus()
  })
}

function close() {
  open.value = false
}

function onFiltersUpdate(filters) {
  currentFilters.value = filters
}

function doSearchWithFilters() {
  doSearch()
}

function onInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    doSearch()
  }, 300)
}

async function doSearch() {
  const q = query.value.trim()
  const filters = advancedSearchRef.value?.buildFilters() || currentFilters.value
  const hasFilters = Object.keys(filters).length > 0

  if (q.length < 1 && !hasFilters) {
    results.value = []
    searched.value = false
    return
  }

  loading.value = true
  searched.value = true
  activeIndex.value = 0

  try {
    const { data } = await globalSearch(q, filters)
    results.value = (data.results || []).map((item, i) => ({ ...item, _flatIndex: i }))
  } catch (err) {
    console.error('搜索失败:', err)
    results.value = []
  } finally {
    loading.value = false
  }
}

function moveUp() {
  if (flatResults.value.length === 0) return
  activeIndex.value = (activeIndex.value - 1 + flatResults.value.length) % flatResults.value.length
  scrollIntoView()
}

function moveDown() {
  if (flatResults.value.length === 0) return
  activeIndex.value = (activeIndex.value + 1) % flatResults.value.length
  scrollIntoView()
}

function scrollIntoView() {
  nextTick(() => {
    const el = document.querySelector(`[data-search-index="${activeIndex.value}"]`)
    el?.scrollIntoView({ block: 'nearest' })
  })
}

function selectCurrent() {
  const item = flatResults.value.find((r) => r._flatIndex === activeIndex.value)
  if (item) navigateTo(item)
}

function navigateTo(item) {
  close()
  if (item.type === 'material') {
    router.push(`/materials/${item.id}`)
  } else if (item.type === 'card') {
    router.push('/cards')
  } else if (item.type === 'quiz') {
    router.push('/quiz')
  }
}

function resultClass(item) {
  return item._flatIndex === activeIndex.value
    ? 'bg-primary-50 dark:bg-primary-900/20'
    : 'hover:bg-gray-50 dark:hover:bg-gray-700/30'
}

function iconClass(type) {
  const map = {
    material: 'bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400',
    card: 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400',
    quiz: 'bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400'
  }
  return map[type] || 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
}

// Ctrl+K / Cmd+K 快捷键
function handleKeydown(e) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    if (open.value) {
      close()
    } else {
      openSearch()
    }
  }
  if (e.key === 'Escape' && open.value) {
    close()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
  if (debounceTimer) clearTimeout(debounceTimer)
})

defineExpose({ openSearch })
</script>

<style scoped>
.search-fade-enter-active,
.search-fade-leave-active {
  transition: opacity 0.15s ease;
}
.search-fade-enter-from,
.search-fade-leave-to {
  opacity: 0;
}
</style>
