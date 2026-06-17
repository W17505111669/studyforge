<template>
  <div class="flex flex-col h-full bg-gray-50 dark:bg-gray-900 transition-colors duration-300">
    <!-- Header -->
    <div class="flex-shrink-0 px-4 sm:px-6 lg:px-8 pt-4 sm:pt-6 pb-3">
      <div class="flex items-center justify-between mb-3">
        <div>
          <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-gray-100">今日复习</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">智能排序 · 优先复习到期和困难卡片</p>
        </div>
        <div class="flex items-center gap-2">
          <!-- 排序方式选择 -->
          <select
            v-model="sortMode"
            class="text-xs px-2.5 py-1.5 rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
            @change="handleSortChange"
          >
            <option value="ai">AI 推荐</option>
            <option value="overdue">到期时间</option>
            <option value="difficulty">难度优先</option>
            <option value="random">随机排序</option>
          </select>
          <router-link
            to="/study"
            class="text-sm text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 transition flex items-center gap-1.5 px-3 py-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
              />
            </svg>
            全部卡片
          </router-link>
        </div>
      </div>

      <!-- 统计条 -->
      <div class="flex items-center gap-4 mb-3">
        <div class="flex items-center gap-3 text-sm">
          <span class="flex items-center gap-1.5">
            <span class="w-2 h-2 rounded-full bg-red-500 animate-pulse"></span>
            <span class="text-gray-600 dark:text-gray-400">待复习</span>
            <span class="font-bold text-red-600 dark:text-red-400 tabular-nums">{{ dueCount }}</span>
          </span>
          <span class="text-gray-300 dark:text-gray-600">|</span>
          <span class="flex items-center gap-1.5">
            <svg class="w-4 h-4 text-emerald-500" fill="currentColor" viewBox="0 0 20 20">
              <path
                fill-rule="evenodd"
                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                clip-rule="evenodd"
              />
            </svg>
            <span class="text-gray-600 dark:text-gray-400">今日已复习</span>
            <span class="font-bold text-emerald-600 dark:text-emerald-400 tabular-nums">{{ todayReviewed }}</span>
          </span>
          <span class="text-gray-300 dark:text-gray-600">|</span>
          <span class="flex items-center gap-1.5 text-xs">
            <span class="text-gray-500 dark:text-gray-400">AI:</span>
            <span class="text-red-600 dark:text-red-400 font-semibold">{{ distribution.high }}</span>
            <span class="text-gray-400">高</span>
            <span class="text-amber-600 dark:text-amber-400 font-semibold">{{ distribution.medium }}</span>
            <span class="text-gray-400">中</span>
            <span class="text-green-600 dark:text-green-400 font-semibold">{{ distribution.low }}</span>
            <span class="text-gray-400">低</span>
          </span>
        </div>
        <!-- 进度条 -->
        <div class="flex-1 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
          <div
            class="h-full rounded-full transition-all duration-500 ease-out"
            :class="progressColor"
            :style="{ width: progressPct + '%' }"
          ></div>
        </div>
        <span class="text-xs text-gray-400 dark:text-gray-500 tabular-nums whitespace-nowrap">
          {{ reviewedInSession }}/{{ dueCount }}
        </span>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <div class="flex flex-col items-center gap-3">
        <svg class="animate-spin w-8 h-8 text-primary-500" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
        <span class="text-sm text-gray-400 dark:text-gray-500">加载待复习卡片...</span>
      </div>
    </div>

    <!-- 空状态 — 所有卡片复习完毕 -->
    <div
      v-else-if="cards.length === 0 || currentIndex >= cards.length"
      class="flex-1 flex items-center justify-center p-6"
    >
      <div class="text-center max-w-md">
        <div
          class="w-28 h-28 mx-auto mb-6 bg-gradient-to-br from-emerald-100 to-green-50 dark:from-emerald-900/40 dark:to-green-800/30 rounded-full flex items-center justify-center"
        >
          <svg
            class="w-14 h-14 text-emerald-500 dark:text-emerald-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.5"
              d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>
        <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">
          {{ reviewedInSession > 0 ? '本轮复习完成！' : '所有卡片都复习完毕，今天可以休息了！' }}
        </h2>
        <p class="text-gray-500 dark:text-gray-400 mb-4">
          {{
            reviewedInSession > 0
              ? `本轮复习了 ${reviewedInSession} 张卡片，继续保持好习惯！`
              : '目前没有需要复习的卡片，去上传新材料或休息一下。'
          }}
        </p>
        <!-- 本轮统计 -->
        <div v-if="reviewedInSession > 0" class="flex justify-center gap-6 mb-6 text-sm">
          <div class="text-center">
            <p class="text-2xl font-bold text-emerald-500">{{ sessionStats.mastered }}</p>
            <p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">已掌握</p>
          </div>
          <div class="text-center">
            <p class="text-2xl font-bold text-red-500">{{ sessionStats.unfamiliar }}</p>
            <p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">不熟练</p>
          </div>
          <div class="text-center">
            <p class="text-2xl font-bold text-amber-500">{{ sessionStats.fuzzy }}</p>
            <p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">模糊跳过</p>
          </div>
        </div>
        <div class="flex gap-3 justify-center">
          <button
            class="px-5 py-2.5 bg-primary-600 text-white rounded-xl text-sm font-medium hover:bg-primary-700 transition-colors shadow-sm"
            @click="restart"
          >
            重新加载
          </button>
          <router-link
            to="/cards"
            class="px-5 py-2.5 border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 rounded-xl text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
          >
            浏览卡片库
          </router-link>
        </div>
      </div>
    </div>

    <!-- 学习区域 -->
    <div
      v-else
      class="flex-1 flex flex-col items-center justify-center px-4 sm:px-6 lg:px-8 pb-4 overflow-hidden select-none"
    >
      <!-- Card container with touch/swipe -->
      <div
        ref="cardContainerRef"
        class="relative w-full max-w-md mx-auto"
        style="perspective: 1200px"
        @touchstart="onTouchStart"
        @touchmove="onTouchMove"
        @touchend="onTouchEnd"
      >
        <!-- Stacked background cards -->
        <div
          v-if="currentIndex + 2 < cards.length"
          class="absolute inset-0 flex items-center justify-center pointer-events-none"
          style="transform: scale(0.92) translateY(16px); opacity: 0.3"
        >
          <div
            class="w-full h-72 sm:h-80 rounded-2xl bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 shadow-sm"
          ></div>
        </div>
        <div
          v-if="currentIndex + 1 < cards.length"
          class="absolute inset-0 flex items-center justify-center pointer-events-none"
          style="transform: scale(0.96) translateY(8px); opacity: 0.6"
        >
          <div
            class="w-full h-72 sm:h-80 rounded-2xl bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 shadow-sm"
          ></div>
        </div>

        <!-- Main flip card -->
        <div class="relative w-full cursor-pointer" :style="cardStyle" @click="flipCard">
          <div
            class="flip-card-inner w-full"
            :class="{ flipped: isFlipped }"
            style="transform-style: preserve-3d; position: relative"
          >
            <!-- Front face -->
            <div
              class="flip-card-face w-full rounded-2xl border shadow-lg p-6 sm:p-8 flex flex-col items-center justify-center text-center"
              :class="
                isDark
                  ? 'bg-gradient-to-br from-gray-800 to-gray-700 border-gray-600'
                  : 'bg-gradient-to-br from-white to-gray-50 border-gray-200'
              "
              style="backface-visibility: hidden"
            >
              <!-- Priority badge -->
              <div class="absolute top-3 right-3 flex items-center gap-1.5">
                <span class="px-2 py-0.5 rounded-full text-xs font-semibold" :class="priorityBadgeClass(currentCard)">
                  {{ priorityLabel(currentCard) }}
                </span>
                <span
                  v-if="currentCard._priorityScore !== undefined"
                  class="text-xs text-gray-400 dark:text-gray-500 tabular-nums"
                >
                  {{ Math.round(currentCard._priorityScore) }}
                </span>
              </div>

              <!-- Difficulty badge -->
              <div class="absolute top-3 left-3">
                <span class="px-2 py-0.5 rounded-full text-xs font-medium" :class="diffClass(currentCard.difficulty)">
                  {{ diffLabel(currentCard.difficulty) }}
                </span>
              </div>

              <p class="text-xs text-gray-400 dark:text-gray-500 mb-3 uppercase tracking-wider">
                {{ currentCard.tags ? currentCard.tags.split(',')[0].trim() : '知识卡片' }}
              </p>
              <h2
                class="text-xl sm:text-2xl font-bold mb-3 leading-snug"
                :class="isDark ? 'text-gray-100' : 'text-gray-800'"
                v-html="renderMath(currentCard.concept)"
              ></h2>
              <p class="text-xs mt-2" :class="isDark ? 'text-gray-500' : 'text-gray-400'">
                {{ currentCard.review_count || 0 }} 次复习 · {{ easeLabel(currentCard.ease_factor) }}
              </p>
              <!-- Overdue info -->
              <p v-if="getOverdueDays(currentCard) > 0" class="text-xs mt-1 text-red-500 dark:text-red-400 font-medium">
                已到期 {{ getOverdueDays(currentCard) }} 天
              </p>
              <p
                v-else-if="!currentCard.next_review_at"
                class="text-xs mt-1 text-blue-500 dark:text-blue-400 font-medium"
              >
                新卡片 · 尚未复习
              </p>

              <p class="text-xs mt-4" :class="isDark ? 'text-gray-500' : 'text-gray-400'">点击翻转查看详情</p>
            </div>

            <!-- Back face -->
            <div
              class="flip-card-face w-full rounded-2xl border shadow-lg overflow-hidden"
              :class="
                isDark
                  ? 'bg-gradient-to-br from-gray-800 to-gray-700 border-gray-600'
                  : 'bg-gradient-to-br from-primary-50 to-indigo-50 border-primary-100'
              "
              style="backface-visibility: hidden; transform: rotateY(180deg); position: absolute; top: 0; left: 0"
            >
              <div class="p-5 sm:p-6 overflow-y-auto custom-scroll" style="max-height: clamp(320px, 55vh, 460px)">
                <h3
                  class="text-sm font-bold mb-3"
                  :class="isDark ? 'text-gray-200' : 'text-gray-700'"
                  v-html="renderMath(currentCard.concept)"
                ></h3>

                <!-- Detail -->
                <div
                  class="text-sm leading-relaxed mb-4"
                  :class="isDark ? 'text-gray-300' : 'text-gray-600'"
                  v-html="renderMath(currentCard.detail)"
                ></div>

                <!-- Formula -->
                <div
                  v-if="currentCard.formula"
                  class="mb-3 px-3 py-2 rounded-lg text-sm"
                  :class="
                    isDark
                      ? 'bg-blue-900/30 text-blue-300 border border-blue-800'
                      : 'bg-blue-50 text-blue-700 border border-blue-100'
                  "
                >
                  <span class="text-xs font-semibold opacity-70">公式:</span>
                  <span v-html="renderMath(currentCard.formula)"></span>
                </div>

                <!-- Memory tip -->
                <div
                  v-if="currentCard.memory_tip"
                  class="mb-3 px-3 py-2 rounded-lg text-sm"
                  :class="
                    isDark
                      ? 'bg-amber-900/20 text-amber-300 border border-amber-800'
                      : 'bg-amber-50 text-amber-700 border border-amber-100'
                  "
                >
                  <span class="text-xs font-semibold opacity-70">记忆技巧:</span>
                  <span v-html="renderMath(currentCard.memory_tip)"></span>
                </div>

                <!-- Interval / Ease stats -->
                <div
                  class="flex items-center gap-3 text-xs mt-3 pt-3 border-t"
                  :class="isDark ? 'border-gray-600 text-gray-400' : 'border-primary-100 text-gray-500'"
                >
                  <span>间隔 {{ currentCard.interval_days || 0 }} 天</span>
                  <span>·</span>
                  <span>系数 {{ (currentCard.ease_factor || 2.5).toFixed(1) }}</span>
                  <span>·</span>
                  <span>下次 {{ fmtDate(currentCard.next_review_at) }}</span>
                </div>

                <!-- Sort reason -->
                <p
                  v-if="sortMode === 'ai' && currentCard._priorityReason"
                  class="text-xs mt-2 italic"
                  :class="isDark ? 'text-gray-500' : 'text-gray-400'"
                >
                  排序理由: {{ currentCard._priorityReason }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Review action buttons -->
      <div class="flex items-center gap-3 sm:gap-4 mt-6">
        <button
          :disabled="reviewing"
          class="flex flex-col items-center gap-1 px-5 sm:px-6 py-3 rounded-xl text-sm font-medium transition-all bg-red-50 text-red-600 hover:bg-red-100 dark:bg-red-900/30 dark:text-red-400 dark:hover:bg-red-900/50 disabled:opacity-50 active:scale-95"
          @click.stop="submitReview('review')"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
            />
          </svg>
          不熟
        </button>
        <button
          :disabled="reviewing"
          class="flex flex-col items-center gap-1 px-5 sm:px-6 py-3 rounded-xl text-sm font-medium transition-all bg-amber-50 text-amber-600 hover:bg-amber-100 dark:bg-amber-900/30 dark:text-amber-400 dark:hover:bg-amber-900/50 disabled:opacity-50 active:scale-95"
          @click.stop="markFuzzy"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          模糊
        </button>
        <button
          :disabled="reviewing"
          class="flex flex-col items-center gap-1 px-5 sm:px-6 py-3 rounded-xl text-sm font-medium transition-all bg-emerald-50 text-emerald-600 hover:bg-emerald-100 dark:bg-emerald-900/30 dark:text-emerald-400 dark:hover:bg-emerald-900/50 disabled:opacity-50 active:scale-95"
          @click.stop="submitReview('mastered')"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          掌握
        </button>
      </div>

      <!-- Keyboard shortcuts hint (desktop only) -->
      <div class="hidden sm:flex items-center gap-4 mt-3 text-xs text-gray-400 dark:text-gray-600">
        <span>
          <kbd class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded text-gray-500 dark:text-gray-400 font-mono">
            ←
          </kbd>
          <kbd
            class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded text-gray-500 dark:text-gray-400 font-mono ml-0.5"
          >
            →
          </kbd>
          切换
        </span>
        <span>
          <kbd class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded text-gray-500 dark:text-gray-400 font-mono">
            Space
          </kbd>
          翻转
        </span>
        <span>
          <kbd class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded text-gray-500 dark:text-gray-400 font-mono">
            1
          </kbd>
          <kbd
            class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded text-gray-500 dark:text-gray-400 font-mono ml-0.5"
          >
            2
          </kbd>
          <kbd
            class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded text-gray-500 dark:text-gray-400 font-mono ml-0.5"
          >
            3
          </kbd>
          评价
        </span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { reviewCard, getUserStats, getDueCards } from '../api/client'
import { useToast } from '../composables/useToast'
import { renderMath } from '../composables/useMathRender'
import { useDarkMode } from '../composables/useDarkMode'

const { success: toastSuccess } = useToast()
const { isDark } = useDarkMode()

// ── Data ──
const cards = ref([])
const loading = ref(true)
const currentIndex = ref(0)
const sortMode = ref('ai')
const dueCount = ref(0)
const todayReviewed = ref(0)
const distribution = ref({ high: 0, medium: 0, low: 0 })

// ── Card state ──
const isFlipped = ref(false)
const cardTransform = ref('translateX(0) rotate(0deg)')
const cardTransition = ref(true)
const reviewing = ref(false)
const cardContainerRef = ref(null)

// ── Touch state ──
const touchStartX = ref(0)
const touchStartY = ref(0)
const swipeOffset = ref(0)

// ── Review tracking ──
const sessionStats = ref({ mastered: 0, unfamiliar: 0, fuzzy: 0 })

// ── Computed ──
const currentCard = computed(() => cards.value[currentIndex.value] || {})

const reviewedInSession = computed(
  () => sessionStats.value.mastered + sessionStats.value.unfamiliar + sessionStats.value.fuzzy
)

const progressPct = computed(() => {
  if (dueCount.value === 0) return 100
  return Math.min((reviewedInSession.value / dueCount.value) * 100, 100)
})

const progressColor = computed(() => {
  const pct = progressPct.value
  if (pct >= 80) return 'bg-gradient-to-r from-emerald-500 to-green-400'
  if (pct >= 50) return 'bg-gradient-to-r from-amber-500 to-yellow-400'
  return 'bg-gradient-to-r from-red-500 to-orange-400'
})

// ── Swipe style ──
const cardStyle = computed(() => ({
  transform: cardTransform.value,
  transition: cardTransition.value ? 'transform 0.1s cubic-bezier(0.4, 0, 0.2, 1)' : 'none'
}))

// ── Priority scoring (server-side via GetDueCards) ──
function getOverdueDays(card) {
  if (!card.next_review_at) return 0
  const now = new Date()
  const nextReview = new Date(card.next_review_at)
  return Math.max(0, Math.floor((now - nextReview) / (1000 * 60 * 60 * 24)))
}

function priorityBadgeClass(card) {
  const score = card._priorityScore || 0
  if (score >= 60) return 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400'
  if (score >= 30) return 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-400'
  return 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-400'
}

function priorityLabel(card) {
  const score = card._priorityScore || 0
  if (score >= 60) return '优先'
  if (score >= 30) return '一般'
  return '轻松'
}

function handleSortChange() {
  if (sortMode.value === 'random') {
    // Client-side Fisher-Yates shuffle
    for (let i = cards.value.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1))
      ;[cards.value[i], cards.value[j]] = [cards.value[j], cards.value[i]]
    }
    currentIndex.value = 0
    isFlipped.value = false
  } else {
    // Server-side sort: reload data
    loadCards()
  }
}

function _applySorting() {
  // Server already handles sorting; this is now a no-op for non-random modes
  if (sortMode.value === 'random') {
    for (let i = cards.value.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1))
      ;[cards.value[i], cards.value[j]] = [cards.value[j], cards.value[i]]
    }
  }
  currentIndex.value = 0
  isFlipped.value = false
}

// ── Helpers ──
function diffClass(d) {
  return (
    {
      easy: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
      medium: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
      hard: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    }[d] || 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
  )
}

function diffLabel(d) {
  return { easy: '简单', medium: '中等', hard: '困难' }[d] || '未分级'
}

function easeLabel(ease) {
  if (!ease || ease >= 2.5) return '较轻松'
  if (ease >= 2.0) return '适中'
  if (ease >= 1.5) return '较困难'
  return '困难'
}

function fmtDate(dateStr) {
  if (!dateStr) return '今天'
  const d = new Date(dateStr)
  const now = new Date()
  const diffDays = Math.ceil((d - now) / (1000 * 60 * 60 * 24))
  if (diffDays <= 0) return '今天'
  if (diffDays === 1) return '明天'
  if (diffDays <= 7) return `${diffDays}天后`
  return `${d.getMonth() + 1}/${d.getDate()}`
}

// ── Load data ──
async function loadCards() {
  loading.value = true
  try {
    // Map frontend sort modes to backend sort params
    const backendSort =
      {
        ai: 'priority',
        overdue: 'due_date',
        difficulty: 'difficulty'
      }[sortMode.value] || 'priority'

    // Load due cards (with server-side priority scoring) and stats in parallel
    const [dueRes, statsRes] = await Promise.allSettled([
      getDueCards({ sort: backendSort, limit: 200 }),
      getUserStats()
    ])

    if (dueRes.status === 'fulfilled') {
      const d = dueRes.value.data
      const rawCards = d.data || []
      // Map server fields to local aliases
      rawCards.forEach((c) => {
        c._priorityScore = c.priority_score || 0
        c._priorityReason = (c.priority_reasons || []).join(' + ')
        c._priorityLevel = c.priority_level || 'low'
      })
      cards.value = rawCards
      dueCount.value = rawCards.length

      // Distribution from server
      if (d.distribution) {
        distribution.value = d.distribution
      }
    }

    if (statsRes.status === 'fulfilled') {
      const d = statsRes.value.data
      dueCount.value = d.due_card_count || dueCount.value
      todayReviewed.value = d.today_reviewed_count || 0
    }
  } catch (e) {
    console.error('Failed to load review cards:', e)
  } finally {
    loading.value = false
  }
}

// ── Touch handlers ──
function onTouchStart(e) {
  touchStartX.value = e.touches[0].clientX
  touchStartY.value = e.touches[0].clientY
  swipeOffset.value = 0
  cardTransition.value = false
}

function onTouchMove(e) {
  const dx = e.touches[0].clientX - touchStartX.value
  const dy = e.touches[0].clientY - touchStartY.value
  if (Math.abs(dx) > Math.abs(dy) && Math.abs(dx) > 10) {
    e.preventDefault()
    swipeOffset.value = dx
    const rotation = (dx / window.innerWidth) * 12
    cardTransform.value = `translateX(${dx}px) rotate(${rotation}deg)`
  }
}

function onTouchEnd() {
  cardTransition.value = true
  const threshold = Math.min(window.innerWidth * 0.25, 120)

  if (Math.abs(swipeOffset.value) > threshold) {
    if (swipeOffset.value < 0 && currentIndex.value + 1 < cards.value.length) {
      doNav(1)
    } else if (swipeOffset.value > 0 && currentIndex.value > 0) {
      doNav(-1)
    } else {
      resetSwipe()
    }
  } else {
    resetSwipe()
  }
}

function resetSwipe() {
  cardTransition.value = true
  cardTransform.value = 'translateX(0) rotate(0deg)'
  swipeOffset.value = 0
}

// ── Navigation ──
function doNav(direction) {
  if (reviewing.value) return
  isFlipped.value = false
  const offset = direction > 0 ? -window.innerWidth : window.innerWidth
  const rotation = direction > 0 ? -15 : 15
  cardTransition.value = true
  cardTransform.value = `translateX(${offset}px) rotate(${rotation}deg)`

  setTimeout(() => {
    cardTransition.value = false
    const entryOffset = direction > 0 ? window.innerWidth : -window.innerWidth
    cardTransform.value = `translateX(${entryOffset}px) rotate(${-rotation}deg)`
    void cardContainerRef.value?.offsetWidth
    cardTransition.value = true
    currentIndex.value += direction
    cardTransform.value = 'translateX(0) rotate(0deg)'
    setTimeout(() => {
      swipeOffset.value = 0
    }, 100)
  }, 100)
}

function flipCard() {
  if (Math.abs(swipeOffset.value) > 8) return
  isFlipped.value = !isFlipped.value
}

// ── Review ──
async function submitReview(result) {
  if (reviewing.value) return
  reviewing.value = true
  try {
    const card = currentCard.value
    const res = await reviewCard(card.id, result)
    const data = res.data
    card.review_count = data.review_count
    card.interval_days = data.interval_days
    card.ease_factor = data.ease_factor
    card.next_review_at = data.next_review_at
    card.last_reviewed_at = new Date().toISOString()

    if (result === 'mastered') {
      sessionStats.value.mastered++
      todayReviewed.value++
      toastSuccess(`已掌握「${card.concept.slice(0, 12)}」，${data.interval_days} 天后复习`)
    } else {
      sessionStats.value.unfamiliar++
      todayReviewed.value++
      toastSuccess(`「${card.concept.slice(0, 12)}」明天再复习`)
    }
    setTimeout(() => doNav(1), 60)
  } catch (e) {
    console.error('Review failed:', e)
    toastSuccess('提交失败，请重试')
  } finally {
    reviewing.value = false
  }
}

function markFuzzy() {
  if (reviewing.value) return
  sessionStats.value.fuzzy++
  const card = currentCard.value
  toastSuccess(`「${card.concept.slice(0, 12)}」标记为模糊，下一张`)
  doNav(1)
}

// ── Keyboard ──
function handleKeydown(e) {
  const tag = e.target?.tagName
  if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || e.target?.contentEditable === 'true') return

  if (e.key === 'ArrowLeft') {
    e.preventDefault()
    currentIndex.value > 0 && doNav(-1)
  } else if (e.key === 'ArrowRight') {
    e.preventDefault()
    currentIndex.value + 1 < cards.value.length && doNav(1)
  } else if (e.key === ' ') {
    e.preventDefault()
    flipCard()
  } else if (e.key === '1') submitReview('review')
  else if (e.key === '2') markFuzzy()
  else if (e.key === '3') submitReview('mastered')
}

// ── Lifecycle ──
function restart() {
  currentIndex.value = 0
  isFlipped.value = false
  sessionStats.value = { mastered: 0, unfamiliar: 0, fuzzy: 0 }
  loadCards()
}

onMounted(() => {
  loadCards()
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.flip-card-inner {
  transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.flip-card-inner.flipped {
  transform: rotateY(180deg);
}
.flip-card-face {
  min-height: clamp(320px, 55vh, 460px);
}

/* Custom scrollbar for back face */
:deep(.custom-scroll)::-webkit-scrollbar {
  width: 3px;
}
:deep(.custom-scroll)::-webkit-scrollbar-track {
  background: transparent;
}
:deep(.custom-scroll)::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.4);
  border-radius: 3px;
}
:deep(.custom-scroll)::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.6);
}

/* Dark mode KaTeX */
:deep(.katex) {
  color: inherit;
}
</style>
