<template>
  <div class="mt-5 pt-4 border-t border-gray-100 dark:border-gray-700">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-sm font-medium text-gray-600 dark:text-gray-400 flex items-center gap-1.5">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        执行时间线
      </h3>
      <span v-if="allDone && totalElapsed > 0" class="text-xs text-gray-400 dark:text-gray-500">
        总计 {{ formatDuration(totalElapsed) }}
      </span>
      <span v-else-if="anyStarted" class="text-xs text-blue-500 dark:text-blue-400 animate-pulse">
        执行中 {{ formatDuration(currentElapsed) }}
      </span>
    </div>

    <!-- Timeline rows -->
    <div class="space-y-2">
      <div v-for="bar in bars" :key="bar.key" class="flex items-center gap-3">
        <!-- Agent label -->
        <div class="w-20 flex items-center gap-1.5 shrink-0">
          <span class="text-xs">{{ bar.icon }}</span>
          <span
            class="text-xs font-medium truncate"
            :class="
              bar.done
                ? bar.error
                  ? 'text-red-500 dark:text-red-400'
                  : 'text-gray-600 dark:text-gray-400'
                : 'text-blue-600 dark:text-blue-400'
            "
          >
            {{ bar.label }}
          </span>
        </div>

        <!-- Bar track -->
        <div class="flex-1 h-7 bg-gray-50 dark:bg-gray-700/40 rounded-lg relative overflow-hidden">
          <!-- Grid lines -->
          <div
            v-for="tick in ticks"
            :key="tick.time"
            class="absolute top-0 bottom-0 w-px"
            :class="isDark ? 'bg-gray-600/20' : 'bg-gray-200/70'"
            :style="{ left: (tick.time / timeScale) * 100 + '%' }"
          ></div>

          <!-- Agent bar -->
          <div
            class="absolute top-1 bottom-1 rounded-md transition-all duration-500 ease-out flex items-center overflow-hidden"
            :class="
              bar.done
                ? bar.error
                  ? 'bg-red-100 dark:bg-red-900/30 border border-red-200 dark:border-red-700'
                  : 'bg-green-100 dark:bg-green-900/30 border border-green-200 dark:border-green-700'
                : 'bg-blue-100 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-700'
            "
            :style="barStyle(bar)"
          >
            <!-- Shimmer overlay for running agents -->
            <div v-if="!bar.done" class="absolute inset-0 shimmer-bg"></div>

            <!-- Duration label inside bar -->
            <span
              class="relative text-[10px] font-medium px-2 whitespace-nowrap"
              :class="
                bar.done
                  ? bar.error
                    ? 'text-red-600 dark:text-red-400'
                    : 'text-green-700 dark:text-green-400'
                  : 'text-blue-600 dark:text-blue-400'
              "
            >
              {{ formatDuration(bar.displayDuration) }}
            </span>
          </div>

          <!-- Pending indicator -->
          <div
            v-if="!bar.started"
            class="absolute top-1 bottom-1 left-1 w-1.5 rounded-full bg-gray-300 dark:bg-gray-600 animate-pulse"
          ></div>
        </div>

        <!-- Status icon -->
        <div class="w-5 shrink-0 text-center">
          <span v-if="bar.done && !bar.error" class="text-green-500 text-xs">✓</span>
          <span v-else-if="bar.done && bar.error" class="text-red-500 text-xs">✗</span>
          <span v-else-if="bar.started" class="inline-block w-2 h-2 bg-blue-500 rounded-full animate-pulse"></span>
          <span v-else class="text-gray-300 dark:text-gray-600 text-xs">—</span>
        </div>
      </div>
    </div>

    <!-- Time axis -->
    <div class="flex items-center gap-3 mt-2.5">
      <div class="w-20 shrink-0"></div>
      <div class="flex-1 relative h-4 border-t border-gray-200 dark:border-gray-600">
        <span
          v-for="tick in ticks"
          :key="'l-' + tick.time"
          class="absolute text-[9px] text-gray-400 dark:text-gray-500 -translate-x-1/2 leading-4"
          :style="{ left: (tick.time / timeScale) * 100 + '%' }"
        >
          {{ tick.label }}
        </span>
      </div>
      <div class="w-5 shrink-0"></div>
    </div>

    <!-- Legend -->
    <div class="flex items-center gap-4 mt-3 text-[10px] text-gray-400 dark:text-gray-500">
      <span class="flex items-center gap-1.5">
        <span
          class="inline-block w-5 h-2 rounded-sm bg-blue-100 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-700"
        ></span>
        执行中
      </span>
      <span class="flex items-center gap-1.5">
        <span
          class="inline-block w-5 h-2 rounded-sm bg-green-100 dark:bg-green-900/30 border border-green-200 dark:border-green-700"
        ></span>
        完成
      </span>
      <span class="flex items-center gap-1.5">
        <span
          class="inline-block w-5 h-2 rounded-sm bg-red-100 dark:bg-red-900/30 border border-red-200 dark:border-red-700"
        ></span>
        失败
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { useDarkMode } from '../composables/useDarkMode'

const { isDark } = useDarkMode()

const props = defineProps({
  agents: { type: Array, required: true },
  timelineStart: { type: Number, default: 0 },
  agentFinishTimes: { type: Object, default: () => ({}) },
  analyzing: { type: Boolean, default: false }
})

const agentLabels = {
  Analyst: '分析师',
  QuizMaster: '出题官',
  CardMaker: '卡片师',
  MapBuilder: '图谱师'
}

// Reactive tick for real-time bar growth
const tick = ref(0)
let tickInterval = null

onMounted(() => {
  tickInterval = setInterval(() => {
    if (props.analyzing && props.timelineStart > 0) {
      tick.value = Date.now()
    }
  }, 200)
})

onUnmounted(() => {
  if (tickInterval) clearInterval(tickInterval)
})

// Reset tick timer when analyzing state changes
watch(
  () => props.analyzing,
  (val) => {
    if (val) {
      tick.value = Date.now()
    }
  }
)

const allDone = computed(() => props.agents.every((a) => a.done))
const anyStarted = computed(() => props.analyzing && props.timelineStart > 0)

const currentElapsed = computed(() => {
  // tick.value ensures reactivity every 200ms
  const now = tick.value || Date.now()
  return props.timelineStart ? Math.max(0, now - props.timelineStart) : 0
})

const totalElapsed = computed(() => {
  if (!allDone.value) return 0
  const times = Object.values(props.agentFinishTimes)
  if (times.length === 0) return 0
  return Math.max(...times) - props.timelineStart
})

const bars = computed(() => {
  // Reference tick for real-time updates
  const _ = tick.value

  const result = props.agents.map((agent) => {
    const key = agent.name.split(' ')[0]
    const label = agentLabels[key] || key
    const icon = agent.icon || '🤖'
    const finishTime = props.agentFinishTimes[key] || 0

    if (!agent.done && !props.analyzing) {
      return { key, label, icon, done: false, error: false, started: false, startMs: 0, endMs: 0, displayDuration: 0 }
    }

    if (agent.done) {
      const durationMs = agent.duration || 0
      const startMs = finishTime > 0 ? finishTime - durationMs : 0
      return {
        key,
        label,
        icon,
        done: true,
        error: agent.error || false,
        started: true,
        startMs,
        endMs: finishTime,
        displayDuration: durationMs
      }
    }

    // Running agent
    const now = Date.now()
    const elapsed = props.timelineStart ? now - props.timelineStart : 0
    return {
      key,
      label,
      icon,
      done: false,
      error: false,
      started: true,
      startMs: 0,
      endMs: elapsed,
      displayDuration: elapsed
    }
  })

  // Sort: completed first (shortest duration first), then running, then not started
  return result.sort((a, b) => {
    if (a.done && !b.done) return -1
    if (!a.done && b.done) return 1
    if (a.done && b.done) return a.displayDuration - b.displayDuration
    return 0
  })
})

const timeScale = computed(() => {
  let maxTime = 10000 // minimum 10s scale
  for (const bar of bars.value) {
    if (bar.endMs > maxTime) maxTime = bar.endMs
  }
  // Add 15% padding
  return maxTime * 1.15
})

const ticks = computed(() => {
  const interval = getTickInterval(timeScale.value)
  const result = []
  for (let t = interval; t < timeScale.value; t += interval) {
    result.push({ time: t, label: formatTimeLabel(t) })
  }
  return result
})

function getTickInterval(total) {
  if (total <= 10000) return 2000
  if (total <= 20000) return 5000
  if (total <= 45000) return 10000
  if (total <= 120000) return 15000
  return 30000
}

function formatTimeLabel(ms) {
  const s = Math.round(ms / 1000)
  if (s < 60) return s + 's'
  const m = Math.floor(s / 60)
  const sec = s % 60
  return m + ':' + String(sec).padStart(2, '0')
}

function formatDuration(ms) {
  if (!ms || ms <= 0) return '0s'
  const s = ms / 1000
  if (s < 60) return s.toFixed(1) + 's'
  const m = Math.floor(s / 60)
  const sec = Math.round(s % 60)
  return m + 'm' + sec + 's'
}

function barStyle(bar) {
  if (!bar.started || timeScale.value <= 0) {
    return { width: '0%', marginLeft: '0%' }
  }
  const left = (bar.startMs / timeScale.value) * 100
  const width = Math.max(((bar.endMs - bar.startMs) / timeScale.value) * 100, 2.5)
  return {
    left: left + '%',
    width: width + '%'
  }
}
</script>

<style scoped>
.shimmer-bg {
  background: linear-gradient(90deg, transparent 0%, rgba(255, 255, 255, 0.45) 50%, transparent 100%);
  background-size: 200% 100%;
  animation: shimmer-move 1.6s ease-in-out infinite;
}
.dark .shimmer-bg {
  background: linear-gradient(90deg, transparent 0%, rgba(255, 255, 255, 0.12) 50%, transparent 100%);
  background-size: 200% 100%;
}
@keyframes shimmer-move {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}
</style>
