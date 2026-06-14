<template>
  <div class="flex flex-col h-full bg-gray-50 dark:bg-gray-900 transition-colors duration-300">
    <!-- Header -->
    <div class="flex-shrink-0 px-4 sm:px-6 lg:px-8 pt-4 sm:pt-6 pb-3">
      <div class="flex items-center justify-between mb-3">
        <div>
          <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-gray-100">学习模式</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
            滑动切换 · 点击翻转 · 键盘 1/2/3 快捷评价
          </p>
        </div>
        <router-link to="/cards"
          class="text-sm text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 transition flex items-center gap-1.5 px-3 py-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          返回卡片库
        </router-link>
      </div>

      <!-- Progress -->
      <div class="flex items-center gap-3">
        <div class="flex-1 h-2.5 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
          <div
            class="h-full bg-gradient-to-r from-primary-500 to-primary-400 rounded-full transition-all duration-500 ease-out"
            :style="{ width: progressPct + '%' }"
          ></div>
        </div>
        <div class="flex items-center gap-2 sm:gap-3 text-xs sm:text-sm whitespace-nowrap">
          <span class="text-primary-600 dark:text-primary-400 font-semibold tabular-nums">{{ currentIndex + 1 }}</span>
          <span class="text-gray-400 dark:text-gray-500">/</span>
          <span class="text-gray-500 dark:text-gray-400 tabular-nums">{{ totalDisplay }}</span>
          <span v-if="dueDisplayCount > 0"
            class="px-2 py-0.5 bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400 rounded-full text-xs font-medium">
            {{ dueDisplayCount }} 待复习
          </span>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <div class="flex flex-col items-center gap-3">
        <svg class="animate-spin w-8 h-8 text-primary-500" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
        <span class="text-sm text-gray-400 dark:text-gray-500">加载卡片中...</span>
      </div>
    </div>

    <!-- All done -->
    <div v-else-if="cards.length === 0 || currentIndex >= cards.length" class="flex-1 flex items-center justify-center p-6">
      <div class="text-center max-w-md">
        <div class="w-24 h-24 mx-auto mb-6 bg-gradient-to-br from-emerald-100 to-green-50 dark:from-emerald-900/40 dark:to-green-800/30 rounded-full flex items-center justify-center">
          <svg class="w-12 h-12 text-emerald-500 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">
          {{ reviewedCount > 0 ? '本轮复习完成！' : '今天都复习完了' }}
        </h2>
        <p class="text-gray-500 dark:text-gray-400 mb-2">
          {{ reviewedCount > 0 ? `本轮复习了 ${reviewedCount} 张卡片，继续保持！` : '所有卡片都已掌握或暂无卡片，休息一下吧' }}
        </p>
        <!-- Review stats -->
        <div v-if="reviewedCount > 0" class="flex justify-center gap-6 mb-6 text-sm">
          <div class="text-center">
            <p class="text-2xl font-bold text-emerald-500">{{ reviewStats.mastered }}</p>
            <p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">已掌握</p>
          </div>
          <div class="text-center">
            <p class="text-2xl font-bold text-red-500">{{ reviewStats.unfamiliar }}</p>
            <p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">不熟练</p>
          </div>
          <div class="text-center">
            <p class="text-2xl font-bold text-amber-500">{{ reviewStats.fuzzy }}</p>
            <p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">模糊跳过</p>
          </div>
        </div>
        <div class="flex gap-3 justify-center">
          <button @click="restart"
            class="px-5 py-2.5 bg-primary-600 text-white rounded-xl text-sm font-medium hover:bg-primary-700 transition-colors shadow-sm">
            再来一轮
          </button>
          <router-link to="/cards"
            class="px-5 py-2.5 border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 rounded-xl text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
            返回卡片库
          </router-link>
        </div>
      </div>
    </div>

    <!-- Study Area -->
    <div v-else class="flex-1 flex flex-col items-center justify-center px-4 sm:px-6 lg:px-8 pb-4 overflow-hidden select-none">
      <!-- Card container with touch/swipe -->
      <div
        ref="cardContainerRef"
        class="relative w-full max-w-md mx-auto"
        style="perspective: 1200px"
        @touchstart="onTouchStart"
        @touchmove="onTouchMove"
        @touchend="onTouchEnd"
      >
        <!-- Decorative stack behind -->
        <div v-if="currentIndex + 1 < cards.length"
          class="absolute inset-0 bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 scale-[0.96] translate-y-2 opacity-60 pointer-events-none"
          style="z-index: 0"
        ></div>
        <div v-if="currentIndex + 2 < cards.length"
          class="absolute inset-0 bg-white dark:bg-gray-800 rounded-2xl border border-gray-100 dark:border-gray-700 scale-[0.92] translate-y-4 opacity-30 pointer-events-none"
          style="z-index: 0"
        ></div>

        <!-- Main card (click to flip) -->
        <div
          class="flip-card relative cursor-pointer"
          :class="{ flipped: isFlipped }"
          :style="cardStyle"
          @click="flipCard"
        >
          <div class="flip-card-inner">
            <!-- ====== FRONT ====== -->
            <div class="flip-card-front bg-white dark:bg-gray-800 rounded-2xl shadow-xl border border-gray-100 dark:border-gray-700 p-6 sm:p-8 flex flex-col">
              <div class="flex items-start justify-between mb-4">
                <span class="px-2.5 py-1 rounded-lg text-xs font-semibold" :class="diffClass(currentCard.difficulty)">
                  {{ diffLabel(currentCard.difficulty) }}
                </span>
                <div class="flex items-center gap-1.5">
                  <span v-if="currentCard.review_count > 0"
                    class="text-xs px-2 py-1 bg-emerald-50 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400 rounded-lg font-medium">
                    复习 {{ currentCard.review_count }}次
                  </span>
                  <span v-else
                    class="text-xs px-2 py-1 bg-blue-50 text-blue-500 dark:bg-blue-900/30 dark:text-blue-400 rounded-lg font-medium">
                    新卡片
                  </span>
                </div>
              </div>

              <div class="flex-1 flex flex-col justify-center">
                <h2 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-gray-100 text-center leading-relaxed"
                  v-html="renderMath(currentCard.concept)">
                </h2>
              </div>

              <div v-if="currentCard.tags && currentCard.tags.length" class="flex flex-wrap gap-1.5 justify-center mt-4">
                <span v-for="tag in currentCard.tags" :key="tag"
                  class="px-2 py-0.5 bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 rounded text-xs">
                  {{ tag }}
                </span>
              </div>

              <p class="text-xs text-gray-300 dark:text-gray-600 mt-4 text-center">点击翻转查看详情</p>
            </div>

            <!-- ====== BACK ====== -->
            <div class="flip-card-back bg-gradient-to-br from-primary-50 to-indigo-50 dark:from-gray-800 dark:to-gray-700 rounded-2xl shadow-xl border border-primary-100 dark:border-gray-600 p-6 sm:p-8 flex flex-col">
              <h3 class="text-base sm:text-lg font-bold text-primary-700 dark:text-primary-400 mb-3"
                v-html="renderMath(currentCard.concept)">
              </h3>

              <div class="flex-1 overflow-y-auto pr-1 -mr-1 custom-scroll">
                <p class="text-sm text-gray-700 dark:text-gray-300 leading-relaxed whitespace-pre-wrap"
                  v-html="renderMath(currentCard.detail)">
                </p>

                <div v-if="currentCard.formula"
                  class="mt-3 p-3 bg-white/80 dark:bg-gray-700/80 rounded-xl border border-primary-100 dark:border-gray-600">
                  <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">关键公式/代码</p>
                  <code class="text-sm text-primary-600 dark:text-primary-400 whitespace-pre-wrap break-all" v-html="renderMath(currentCard.formula)"></code>
                </div>

                <div v-if="currentCard.memory_tip"
                  class="mt-3 p-3 bg-amber-50/80 dark:bg-amber-900/20 rounded-xl border border-amber-100 dark:border-amber-800">
                  <p class="text-xs text-amber-600 dark:text-amber-400 mb-1">记忆技巧</p>
                  <p class="text-sm text-amber-800 dark:text-amber-300 leading-relaxed" v-html="renderMath(currentCard.memory_tip)"></p>
                </div>
              </div>

              <div class="mt-3 pt-2 border-t border-primary-100/50 dark:border-gray-600 flex justify-between text-xs text-gray-400 dark:text-gray-500">
                <span>间隔 {{ currentCard.interval_days || 0 }}天 · 系数 {{ (currentCard.ease_factor || 2.5).toFixed(1) }}</span>
                <span v-if="currentCard.next_review_at">下次: {{ fmtDate(currentCard.next_review_at) }}</span>
                <span v-else>尚未复习</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Desktop side arrows -->
      <div class="hidden sm:flex items-center justify-between w-full max-w-xl mx-auto mt-4 pointer-events-none">
        <button
          v-if="currentIndex > 0"
          @click="nav(-1)"
          class="pointer-events-auto p-2.5 rounded-full text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition"
          title="上一张 (←)"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <div v-else class="pointer-events-none w-10"></div>

        <span class="text-xs text-gray-400 dark:text-gray-500">
          <kbd class="px-1.5 py-0.5 rounded border border-gray-300 dark:border-gray-600 text-[10px] font-mono">←</kbd>
          <kbd class="px-1.5 py-0.5 rounded border border-gray-300 dark:border-gray-600 text-[10px] font-mono ml-1">→</kbd>
          切换
        </span>

        <button
          v-if="currentIndex + 1 < cards.length"
          @click="nav(1)"
          class="pointer-events-auto p-2.5 rounded-full text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition"
          title="下一张 (→)"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>
        <div v-else class="pointer-events-none w-10"></div>
      </div>

      <!-- Action buttons -->
      <div class="flex gap-3 sm:gap-4 mt-5 sm:mt-6 w-full max-w-md mx-auto">
        <button
          @click="submitReview('review')"
          :disabled="reviewing"
          class="flex-1 py-3 sm:py-3.5 rounded-xl text-sm font-semibold transition-all
            bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800
            text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/30
            active:scale-[0.97] disabled:opacity-40 disabled:cursor-not-allowed"
        >
          <span class="block text-lg sm:hidden">✗</span>
          <span class="hidden sm:inline">不熟</span>
          <span class="block text-[10px] text-red-400 dark:text-red-500 mt-0.5 hidden sm:block">重置间隔</span>
        </button>

        <button
          @click="markFuzzy"
          :disabled="reviewing"
          class="flex-1 py-3 sm:py-3.5 rounded-xl text-sm font-semibold transition-all
            bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800
            text-amber-600 dark:text-amber-400 hover:bg-amber-100 dark:hover:bg-amber-900/30
            active:scale-[0.97] disabled:opacity-40 disabled:cursor-not-allowed"
        >
          <span class="block text-lg sm:hidden">~</span>
          <span class="hidden sm:inline">模糊</span>
          <span class="block text-[10px] text-amber-400 dark:text-amber-500 mt-0.5 hidden sm:block">跳过</span>
        </button>

        <button
          @click="submitReview('mastered')"
          :disabled="reviewing"
          class="flex-1 py-3 sm:py-3.5 rounded-xl text-sm font-semibold transition-all
            bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800
            text-emerald-600 dark:text-emerald-400 hover:bg-emerald-100 dark:hover:bg-emerald-900/30
            active:scale-[0.97] disabled:opacity-40 disabled:cursor-not-allowed"
        >
          <span class="block text-lg sm:hidden">✓</span>
          <span class="hidden sm:inline">掌握</span>
          <span class="block text-[10px] text-emerald-400 dark:text-emerald-500 mt-0.5 hidden sm:block">延长间隔</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { listCards, reviewCard } from '../api/client'
import { useToast } from '../composables/useToast'
import { renderMath } from '../composables/useMathRender'

const { success: toastSuccess } = useToast()

// ── Data ──
const cards = ref([])
const loading = ref(true)
const currentIndex = ref(0)

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
const reviewStats = ref({ mastered: 0, unfamiliar: 0, fuzzy: 0 })

// ── Computed ──
const currentCard = computed(() => cards.value[currentIndex.value] || {})

const dueCards = computed(() => {
  const now = new Date()
  return cards.value.filter(c => {
    if (!c.next_review_at) return true
    return new Date(c.next_review_at) <= now
  })
})

const dueDisplayCount = computed(() => Math.min(dueCards.value.length, 99))

const totalDisplay = computed(() => Math.min(cards.value.length, 999))

const progressPct = computed(() => {
  if (cards.value.length === 0) return 0
  return Math.min(((currentIndex.value) / cards.value.length) * 100, 100)
})

const reviewedCount = computed(() =>
  reviewStats.value.mastered + reviewStats.value.unfamiliar + reviewStats.value.fuzzy
)

// ── Swipe style ──
const cardStyle = computed(() => ({
  transform: cardTransform.value,
  transition: cardTransition.value ? 'transform 0.1s cubic-bezier(0.4, 0, 0.2, 1)' : 'none',
}))

// ── Helpers ──
function diffClass(d) {
  return {
    easy: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
    medium: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
    hard: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  }[d] || 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
}

function diffLabel(d) {
  return { easy: '简单', medium: '中等', hard: '困难' }[d] || '未分级'
}

function fmtDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diffDays = Math.ceil((d - now) / (1000 * 60 * 60 * 24))
  if (diffDays <= 0) return '今天'
  if (diffDays === 1) return '明天'
  if (diffDays <= 7) return `${diffDays}天后`
  return `${d.getMonth() + 1}/${d.getDate()}`
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

    // Force reflow so the position is applied before transition resumes
    void cardContainerRef.value?.offsetWidth

    cardTransition.value = true
    currentIndex.value += direction
    cardTransform.value = 'translateX(0) rotate(0deg)'

    setTimeout(() => {
      swipeOffset.value = 0
    }, 100)
  }, 100)
}

function nav(dir) {
  if (dir > 0 && currentIndex.value + 1 >= cards.value.length) return
  if (dir < 0 && currentIndex.value <= 0) return
  doNav(dir)
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
      reviewStats.value.mastered++
      toastSuccess(`已掌握「${card.concept.slice(0, 12)}」，${data.interval_days} 天后复习`)
    } else {
      reviewStats.value.unfamiliar++
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
  reviewStats.value.fuzzy++
  const card = currentCard.value
  toastSuccess(`「${card.concept.slice(0, 12)}」标记为模糊，下一张`)
  doNav(1)
}

// ── Keyboard ──
function handleKeydown(e) {
  const tag = e.target.tagName
  if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return

  switch (e.key) {
    case 'ArrowLeft':
      nav(-1)
      e.preventDefault()
      break
    case 'ArrowRight':
      nav(1)
      e.preventDefault()
      break
    case ' ':
      flipCard()
      e.preventDefault()
      break
    case '1':
      submitReview('review')
      break
    case '2':
      markFuzzy()
      break
    case '3':
      submitReview('mastered')
      break
  }
}

// ── Lifecycle ──
async function loadCards() {
  loading.value = true
  try {
    const res = await listCards({ limit: 200 })
    const all = (res.data.data || []).map(c => ({
      ...c,
      tags: typeof c.tags === 'string' && c.tags ? c.tags.split(',').map(t => t.trim()) : [],
    }))
    // Sort: due cards first, then by next_review_at
    const now = new Date()
    all.sort((a, b) => {
      const aDue = !a.next_review_at || new Date(a.next_review_at) <= now
      const bDue = !b.next_review_at || new Date(b.next_review_at) <= now
      if (aDue && !bDue) return -1
      if (!aDue && bDue) return 1
      return 0
    })
    cards.value = all
  } catch (e) {
    console.error('Load cards failed:', e)
  } finally {
    loading.value = false
  }
}

function restart() {
  currentIndex.value = 0
  isFlipped.value = false
  reviewStats.value = { mastered: 0, unfamiliar: 0, fuzzy: 0 }
  cardTransform.value = 'translateX(0) rotate(0deg)'
  loadCards()
}

onMounted(() => {
  loadCards()
  document.addEventListener('keydown', handleKeydown)
  if (cardContainerRef.value) {
    cardContainerRef.value.addEventListener('touchstart', onTouchStart, { passive: false })
  }
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.flip-card {
  width: 100%;
  height: clamp(320px, 55vh, 460px);
  perspective: 1200px;
  z-index: 1;
  position: relative;
}

.flip-card-inner {
  position: relative;
  width: 100%;
  height: 100%;
  transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  transform-style: preserve-3d;
}

.flip-card.flipped .flip-card-inner {
  transform: rotateY(180deg);
}

.flip-card-front,
.flip-card-back {
  position: absolute;
  inset: 0;
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
  overflow: hidden;
}

.flip-card-back {
  transform: rotateY(180deg);
}

.custom-scroll::-webkit-scrollbar {
  width: 3px;
}
.custom-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroll::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.12);
  border-radius: 2px;
}
:global(.dark) .custom-scroll::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.12);
}

/* KaTeX 暗色模式适配 */
:deep(.katex) {
  color: inherit;
}
</style>
