<template>
  <div class="min-h-full p-4 sm:p-6 lg:p-8">
    <!-- 页面标题 -->
    <div class="max-w-2xl mx-auto">
      <div class="mb-8">
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-100">专注计时</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">番茄工作法 · 专注学习，高效成长</p>
      </div>

      <!-- 模式切换标签 -->
      <div class="flex justify-center mb-8">
        <div class="inline-flex rounded-xl p-1" :class="isDark ? 'bg-gray-800' : 'bg-gray-100'">
          <button
            v-for="mode in modes"
            :key="mode.key"
            class="px-5 py-2.5 rounded-lg text-sm font-medium transition-all duration-200"
            :class="
              currentMode === mode.key
                ? 'bg-primary-500 text-white shadow-lg shadow-primary-500/25'
                : isDark
                  ? 'text-gray-400 hover:text-gray-200'
                  : 'text-gray-600 hover:text-gray-900'
            "
            @click="switchMode(mode.key)"
          >
            {{ mode.label }}
          </button>
        </div>
      </div>

      <!-- 倒计时圆环 -->
      <div class="flex justify-center mb-8">
        <div class="relative">
          <svg :width="ringSize" :height="ringSize" viewBox="0 0 300 300" class="transform -rotate-90">
            <!-- 背景圆环 -->
            <circle
              cx="150"
              cy="150"
              :r="ringRadius"
              fill="none"
              :stroke="isDark ? '#1f2937' : '#e5e7eb'"
              :stroke-width="ringStrokeWidth"
            />
            <!-- 进度圆环 -->
            <circle
              cx="150"
              cy="150"
              :r="ringRadius"
              fill="none"
              :stroke="progressColor"
              :stroke-width="ringStrokeWidth"
              :stroke-dasharray="circumference"
              :stroke-dashoffset="dashOffset"
              stroke-linecap="round"
              class="transition-all duration-1000 ease-linear"
            />
          </svg>
          <!-- 中心文字 -->
          <div class="absolute inset-0 flex flex-col items-center justify-center">
            <span
              class="text-5xl sm:text-6xl font-mono font-bold tracking-wider"
              :class="isDark ? 'text-gray-100' : 'text-gray-900'"
            >
              {{ displayTime }}
            </span>
            <span class="text-sm mt-2" :class="isDark ? 'text-gray-400' : 'text-gray-500'">
              {{ statusLabel }}
            </span>
          </div>
        </div>
      </div>

      <!-- 控制按钮 -->
      <div class="flex items-center justify-center gap-4 mb-8">
        <!-- 重置按钮 -->
        <button
          v-if="timerState !== 'idle'"
          class="w-12 h-12 rounded-full flex items-center justify-center transition-all duration-200"
          :class="
            isDark
              ? 'bg-gray-800 text-gray-400 hover:bg-gray-700 hover:text-gray-200'
              : 'bg-gray-100 text-gray-500 hover:bg-gray-200 hover:text-gray-700'
          "
          title="重置"
          @click="resetTimer"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
            />
          </svg>
        </button>

        <!-- 开始/暂停按钮 -->
        <button
          class="w-16 h-16 rounded-full flex items-center justify-center text-white transition-all duration-200 shadow-lg"
          :class="
            timerState === 'running'
              ? 'bg-amber-500 hover:bg-amber-600 shadow-amber-500/30'
              : 'bg-primary-500 hover:bg-primary-600 shadow-primary-500/30'
          "
          @click="toggleTimer"
        >
          <!-- 播放图标 -->
          <svg v-if="timerState !== 'running'" class="w-7 h-7 ml-1" fill="currentColor" viewBox="0 0 24 24">
            <path d="M8 5v14l11-7z" />
          </svg>
          <!-- 暂停图标 -->
          <svg v-else class="w-7 h-7" fill="currentColor" viewBox="0 0 24 24">
            <path d="M6 4h4v16H6zM14 4h4v16h-4z" />
          </svg>
        </button>

        <!-- 占位（保持居中） -->
        <div v-if="timerState !== 'idle'" class="w-12 h-12"></div>
      </div>

      <!-- 自定义时长 -->
      <div class="flex justify-center mb-8">
        <div class="flex items-center gap-3 text-sm">
          <span :class="isDark ? 'text-gray-400' : 'text-gray-500'">时长：</span>
          <button
            v-for="preset in currentPresets"
            :key="preset.value"
            class="px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-200"
            :class="
              customMinutes === preset.value
                ? 'bg-primary-500 text-white'
                : isDark
                  ? 'bg-gray-800 text-gray-400 hover:bg-gray-700'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            "
            :disabled="timerState === 'running'"
            @click="setDuration(preset.value)"
          >
            {{ preset.label }}
          </button>
        </div>
      </div>

      <!-- 今日统计 -->
      <div
        class="rounded-xl p-6 mb-8 border"
        :class="isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-100'"
      >
        <h3 class="text-sm font-semibold mb-4" :class="isDark ? 'text-gray-300' : 'text-gray-700'">今日专注</h3>
        <div class="grid grid-cols-3 gap-4 text-center">
          <div>
            <div class="text-2xl font-bold text-primary-500">{{ todayStats.completed_count }}</div>
            <div class="text-xs mt-1" :class="isDark ? 'text-gray-400' : 'text-gray-500'">完成番茄</div>
          </div>
          <div>
            <div class="text-2xl font-bold" :class="isDark ? 'text-emerald-400' : 'text-emerald-600'">
              {{ formatMinutes(todayStats.total_minutes) }}
            </div>
            <div class="text-xs mt-1" :class="isDark ? 'text-gray-400' : 'text-gray-500'">专注时长</div>
          </div>
          <div>
            <div class="text-2xl font-bold" :class="isDark ? 'text-amber-400' : 'text-amber-600'">
              {{ todayStats.session_count }}
            </div>
            <div class="text-xs mt-1" :class="isDark ? 'text-gray-400' : 'text-gray-500'">总会话数</div>
          </div>
        </div>
      </div>

      <!-- 本周/本月统计 -->
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-5 mb-8">
        <!-- 本周 -->
        <div class="rounded-xl p-5 border" :class="isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-100'">
          <h3 class="text-sm font-semibold mb-3" :class="isDark ? 'text-gray-300' : 'text-gray-700'">本周</h3>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <div class="text-xl font-bold text-primary-500">{{ weekStats.completed_count }}</div>
              <div class="text-xs mt-0.5" :class="isDark ? 'text-gray-500' : 'text-gray-400'">番茄数</div>
            </div>
            <div>
              <div class="text-xl font-bold" :class="isDark ? 'text-emerald-400' : 'text-emerald-600'">
                {{ formatMinutes(weekStats.total_minutes) }}
              </div>
              <div class="text-xs mt-0.5" :class="isDark ? 'text-gray-500' : 'text-gray-400'">专注时长</div>
            </div>
          </div>
        </div>
        <!-- 本月 -->
        <div class="rounded-xl p-5 border" :class="isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-100'">
          <h3 class="text-sm font-semibold mb-3" :class="isDark ? 'text-gray-300' : 'text-gray-700'">本月</h3>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <div class="text-xl font-bold text-primary-500">{{ monthStats.completed_count }}</div>
              <div class="text-xs mt-0.5" :class="isDark ? 'text-gray-500' : 'text-gray-400'">番茄数</div>
            </div>
            <div>
              <div class="text-xl font-bold" :class="isDark ? 'text-emerald-400' : 'text-emerald-600'">
                {{ formatMinutes(monthStats.total_minutes) }}
              </div>
              <div class="text-xs mt-0.5" :class="isDark ? 'text-gray-500' : 'text-gray-400'">专注时长</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 快捷键提示 -->
      <div class="text-center text-xs" :class="isDark ? 'text-gray-600' : 'text-gray-400'">
        <kbd
          class="px-1.5 py-0.5 rounded border text-[10px] font-mono"
          :class="isDark ? 'border-gray-700 text-gray-500' : 'border-gray-300 text-gray-400'"
        >
          Space
        </kbd>
        暂停/继续
      </div>
    </div>

    <!-- 迷你浮窗计时器（离开番茄钟页面时显示） -->
    <Teleport to="body">
      <Transition name="mini-fade">
        <div v-if="showMiniTimer" class="fixed top-4 right-4 z-[9999] cursor-pointer select-none" @click="goToPomodoro">
          <div
            class="w-14 h-14 rounded-full flex items-center justify-center shadow-lg border-2 relative"
            :class="
              isDark
                ? 'bg-gray-800 border-primary-500 shadow-black/40'
                : 'bg-white border-primary-500 shadow-gray-300/50'
            "
          >
            <!-- 迷你进度环 -->
            <svg class="absolute inset-0 w-full h-full transform -rotate-90" viewBox="0 0 56 56">
              <circle cx="28" cy="28" r="24" fill="none" :stroke="isDark ? '#1f2937' : '#e5e7eb'" stroke-width="3" />
              <circle
                cx="28"
                cy="28"
                r="24"
                fill="none"
                :stroke="progressColor"
                stroke-width="3"
                :stroke-dasharray="miniCircumference"
                :stroke-dashoffset="miniDashOffset"
                stroke-linecap="round"
              />
            </svg>
            <span class="text-xs font-mono font-bold relative z-10" :class="isDark ? 'text-gray-100' : 'text-gray-900'">
              {{ miniTime }}
            </span>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { startPomodoro, endPomodoro, getPomodoroStats } from '../api/client'
import { useDarkMode } from '../composables/useDarkMode'
import { useToast } from '../composables/useToast'

const route = useRoute()
const router = useRouter()
const { isDark } = useDarkMode()
const { success, info } = useToast()

// ========== 模式配置 ==========
const modes = [
  { key: 'work', label: '专注', defaultMinutes: 25 },
  { key: 'short_break', label: '短休', defaultMinutes: 5 },
  { key: 'long_break', label: '长休', defaultMinutes: 15 }
]

const modePresets = {
  work: [
    { label: '15 分钟', value: 15 },
    { label: '25 分钟', value: 25 },
    { label: '30 分钟', value: 30 },
    { label: '45 分钟', value: 45 }
  ],
  short_break: [
    { label: '3 分钟', value: 3 },
    { label: '5 分钟', value: 5 },
    { label: '10 分钟', value: 10 }
  ],
  long_break: [
    { label: '10 分钟', value: 10 },
    { label: '15 分钟', value: 15 },
    { label: '20 分钟', value: 20 },
    { label: '30 分钟', value: 30 }
  ]
}

// ========== 状态 ==========
const currentMode = ref('work')
const customMinutes = ref(25)
const timerState = ref('idle') // idle | running | paused
const remainingSeconds = ref(25 * 60)
const sessionId = ref(null)
let timerInterval = null

// ========== 统计数据 ==========
const todayStats = ref({ completed_count: 0, total_minutes: 0, session_count: 0 })
const weekStats = ref({ completed_count: 0, total_minutes: 0, session_count: 0 })
const monthStats = ref({ completed_count: 0, total_minutes: 0, session_count: 0 })

// ========== 圆环配置 ==========
const ringSize = 280
const ringStrokeWidth = 8
const ringRadius = computed(() => (300 - ringStrokeWidth) / 2)
const circumference = computed(() => 2 * Math.PI * ringRadius.value)
const dashOffset = computed(() => {
  const total = customMinutes.value * 60
  if (total === 0) return circumference.value
  const progress = remainingSeconds.value / total
  return circumference.value * (1 - progress)
})

const progressColor = computed(() => {
  if (currentMode.value === 'work') return isDark.value ? '#6366f1' : '#4f46e5'
  if (currentMode.value === 'short_break') return isDark.value ? '#34d399' : '#10b981'
  return isDark.value ? '#f59e0b' : '#f59e0b'
})

// ========== 迷你浮窗 ==========
const showMiniTimer = computed(() => {
  return timerState.value !== 'idle' && route.path !== '/pomodoro'
})

const miniCircumference = computed(() => 2 * Math.PI * 24)
const miniDashOffset = computed(() => {
  const total = customMinutes.value * 60
  if (total === 0) return miniCircumference.value
  const progress = remainingSeconds.value / total
  return miniCircumference.value * (1 - progress)
})

const miniTime = computed(() => {
  const m = Math.floor(remainingSeconds.value / 60)
  const s = remainingSeconds.value % 60
  return `${m}:${String(s).padStart(2, '0')}`
})

// ========== 计算属性 ==========
const displayTime = computed(() => {
  const m = Math.floor(remainingSeconds.value / 60)
  const s = remainingSeconds.value % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
})

const statusLabel = computed(() => {
  if (timerState.value === 'idle') return '准备开始'
  if (timerState.value === 'paused') return '已暂停'
  const modeLabels = { work: '专注中', short_break: '短休息中', long_break: '长休息中' }
  return modeLabels[currentMode.value] || '进行中'
})

const currentPresets = computed(() => modePresets[currentMode.value] || [])

// ========== 方法 ==========
function switchMode(mode) {
  if (timerState.value === 'running') {
    // 运行中不允许切换模式
    return
  }
  currentMode.value = mode
  const defaultMin = modes.find((m) => m.key === mode)?.defaultMinutes || 25
  customMinutes.value = defaultMin
  remainingSeconds.value = defaultMin * 60
  timerState.value = 'idle'
  sessionId.value = null
  updateTitle()
}

function setDuration(minutes) {
  if (timerState.value === 'running') return
  customMinutes.value = minutes
  remainingSeconds.value = minutes * 60
  timerState.value = 'idle'
  sessionId.value = null
  updateTitle()
}

async function toggleTimer() {
  if (timerState.value === 'idle' || timerState.value === 'paused') {
    await startTimer()
  } else {
    pauseTimer()
  }
}

async function startTimer() {
  if (timerState.value === 'idle') {
    // 向后端创建会话
    try {
      const res = await startPomodoro({
        type: currentMode.value,
        planned_minutes: customMinutes.value
      })
      sessionId.value = res.data.id
    } catch (e) {
      return
    }
    timerState.value = 'running'
    startInterval()
  } else if (timerState.value === 'paused') {
    timerState.value = 'running'
    startInterval()
  }
}

function startInterval() {
  if (timerInterval) clearInterval(timerInterval)
  timerInterval = setInterval(tick, 1000)
}

async function tick() {
  if (remainingSeconds.value <= 0) {
    clearInterval(timerInterval)
    timerInterval = null
    await completeTimer(true)
    return
  }
  remainingSeconds.value--
  updateTitle()
}

function pauseTimer() {
  timerState.value = 'paused'
  if (timerInterval) {
    clearInterval(timerInterval)
    timerInterval = null
  }
  updateTitle()
}

async function resetTimer() {
  if (timerInterval) {
    clearInterval(timerInterval)
    timerInterval = null
  }

  // 如果会话已开始但未完成，标记为未完成
  if (sessionId.value && timerState.value !== 'idle') {
    try {
      await endPomodoro({ session_id: sessionId.value, completed: false })
    } catch {}
  }

  remainingSeconds.value = customMinutes.value * 60
  timerState.value = 'idle'
  sessionId.value = null
  updateTitle()
}

async function completeTimer(completed) {
  timerState.value = 'idle'
  if (timerInterval) {
    clearInterval(timerInterval)
    timerInterval = null
  }

  // 结束后端会话
  if (sessionId.value) {
    try {
      await endPomodoro({ session_id: sessionId.value, completed })
    } catch {}
    sessionId.value = null
  }

  // 播放提示音
  playNotificationSound()

  // 提示
  if (completed) {
    if (currentMode.value === 'work') {
      success('番茄钟完成！休息一下吧 🍅')
    } else {
      info('休息结束，准备开始下一个番茄钟！')
    }
  }

  // 重置计时器
  remainingSeconds.value = customMinutes.value * 60
  updateTitle()

  // 刷新统计
  loadStats()
}

// ========== Web Audio API 提示音 ==========
function playNotificationSound() {
  try {
    const audioCtx = new (window.AudioContext || window.webkitAudioContext)()
    const notes = [523.25, 659.25, 783.99] // C5, E5, G5

    notes.forEach((freq, i) => {
      const oscillator = audioCtx.createOscillator()
      const gainNode = audioCtx.createGain()

      oscillator.connect(gainNode)
      gainNode.connect(audioCtx.destination)

      oscillator.frequency.value = freq
      oscillator.type = 'sine'

      const startTime = audioCtx.currentTime + i * 0.2
      gainNode.gain.setValueAtTime(0.3, startTime)
      gainNode.gain.exponentialRampToValueAtTime(0.01, startTime + 0.3)

      oscillator.start(startTime)
      oscillator.stop(startTime + 0.3)
    })
  } catch {}
}

// ========== 动态标题 ==========
function updateTitle() {
  if (timerState.value === 'idle') {
    document.title = '专注计时 - StudyForge Pro'
  } else {
    document.title = `${displayTime.value} - StudyForge Pro`
  }
}

// ========== 键盘快捷键 ==========
function handleKeydown(e) {
  if (e.code === 'Space' && route.path === '/pomodoro') {
    const tag = document.activeElement?.tagName?.toLowerCase()
    if (tag === 'input' || tag === 'textarea' || tag === 'select' || tag === 'button') return
    e.preventDefault()
    toggleTimer()
  }
}

// 全局快捷键 't' 自定义事件处理器（由 useShortcuts.js 触发）
function handleShortcutToggle() {
  toggleTimer()
}

// ========== 加载统计 ==========
async function loadStats() {
  try {
    const res = await getPomodoroStats()
    todayStats.value = res.data.today || todayStats.value
    weekStats.value = res.data.week || weekStats.value
    monthStats.value = res.data.month || monthStats.value
  } catch {}
}

// ========== 格式化分钟 ==========
function formatMinutes(minutes) {
  if (minutes < 60) return `${minutes}分钟`
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  return m > 0 ? `${h}时${m}分` : `${h}小时`
}

// ========== 导航 ==========
function goToPomodoro() {
  router.push('/pomodoro')
}

// ========== 生命周期 ==========
onMounted(() => {
  loadStats()
  document.addEventListener('keydown', handleKeydown)
  // 全局快捷键 't' 通过自定义事件触发番茄钟切换
  window.addEventListener('shortcut-pomodoro-toggle', handleShortcutToggle)
})

onUnmounted(() => {
  if (timerInterval) {
    clearInterval(timerInterval)
    timerInterval = null
  }
  document.removeEventListener('keydown', handleKeydown)
  window.removeEventListener('shortcut-pomodoro-toggle', handleShortcutToggle)
  // 重置标题
  if (timerState.value === 'idle') {
    document.title = 'StudyForge Pro'
  }
})

// 监听路由变化更新标题
watch(
  () => route.path,
  () => {
    if (route.path !== '/pomodoro' && timerState.value === 'idle') {
      document.title = 'StudyForge Pro'
    }
  }
)
</script>

<style scoped>
/* 迷你浮窗过渡动画 */
.mini-fade-enter-active,
.mini-fade-leave-active {
  transition:
    opacity 0.3s ease,
    transform 0.3s ease;
}
.mini-fade-enter-from,
.mini-fade-leave-to {
  opacity: 0;
  transform: scale(0.8) translateY(-10px);
}
.mini-fade-enter-to,
.mini-fade-leave-from {
  opacity: 1;
  transform: scale(1) translateY(0);
}
</style>
