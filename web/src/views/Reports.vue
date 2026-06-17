<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-5xl mx-auto">
    <!-- 头部 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">学习报告</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">回顾你的学习数据，发现进步与不足</p>
      </div>
      <div class="flex items-center gap-3">
        <!-- 导出图片按钮 -->
        <button
          :disabled="loading"
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          :class="
            isDark
              ? 'bg-gray-700 text-gray-300 hover:bg-gray-600 disabled:opacity-50'
              : 'bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 disabled:opacity-50'
          "
          @click="exportReport"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
            />
          </svg>
          导出图片
        </button>
      </div>
    </div>

    <!-- 周/月切换 + 日期导航 -->
    <div class="flex flex-col sm:flex-row sm:items-center gap-3 mb-6">
      <!-- 模式切换 -->
      <div class="flex rounded-lg overflow-hidden border" :class="isDark ? 'border-gray-700' : 'border-gray-200'">
        <button
          class="px-4 py-2 text-sm font-medium transition-colors"
          :class="
            mode === 'weekly'
              ? 'bg-primary-600 text-white'
              : isDark
                ? 'bg-gray-800 text-gray-400 hover:text-white'
                : 'bg-white text-gray-600 hover:text-gray-900'
          "
          @click="switchMode('weekly')"
        >
          周报
        </button>
        <button
          class="px-4 py-2 text-sm font-medium transition-colors"
          :class="
            mode === 'monthly'
              ? 'bg-primary-600 text-white'
              : isDark
                ? 'bg-gray-800 text-gray-400 hover:text-white'
                : 'bg-white text-gray-600 hover:text-gray-900'
          "
          @click="switchMode('monthly')"
        >
          月报
        </button>
      </div>
      <!-- 日期导航 -->
      <div class="flex items-center gap-2">
        <button
          class="p-1.5 rounded-lg transition-colors"
          :class="isDark ? 'hover:bg-gray-700 text-gray-400' : 'hover:bg-gray-100 text-gray-500'"
          @click="navigate(-1)"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <span class="text-sm font-medium min-w-[160px] text-center" :class="isDark ? 'text-gray-200' : 'text-gray-700'">
          {{ periodLabel }}
        </span>
        <button
          :disabled="!canNavigateNext"
          class="p-1.5 rounded-lg transition-colors"
          :class="[
            isDark ? 'hover:bg-gray-700 text-gray-400' : 'hover:bg-gray-100 text-gray-500',
            { 'opacity-30 cursor-not-allowed': !canNavigateNext }
          ]"
          @click="navigate(1)"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>
        <button
          v-if="!isCurrentPeriod"
          class="ml-1 px-3 py-1 rounded-lg text-xs font-medium transition-colors"
          :class="
            isDark
              ? 'bg-primary-900/30 text-primary-400 hover:bg-primary-900/50'
              : 'bg-primary-50 text-primary-600 hover:bg-primary-100'
          "
          @click="goToCurrent"
        >
          本周
        </button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="space-y-4">
      <div v-for="i in 4" :key="i" class="rounded-xl p-6 animate-pulse" :class="isDark ? 'bg-gray-800' : 'bg-gray-100'">
        <div class="h-4 rounded w-1/3 mb-3" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'"></div>
        <div class="h-8 rounded w-1/4" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'"></div>
      </div>
    </div>

    <!-- 报告内容 -->
    <div v-else-if="report" ref="reportContent">
      <!-- 概览统计卡片 -->
      <div class="grid grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4 mb-6">
        <div
          v-for="stat in statCards"
          :key="stat.key"
          class="rounded-xl p-4 sm:p-5 border transition-colors"
          :class="isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'"
        >
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs font-medium" :class="isDark ? 'text-gray-400' : 'text-gray-500'">
              {{ stat.label }}
            </span>
            <span class="text-lg">{{ stat.icon }}</span>
          </div>
          <div class="text-2xl font-bold" :class="isDark ? 'text-gray-100' : 'text-gray-900'">{{ stat.value }}</div>
          <div v-if="stat.change" class="flex items-center gap-1 mt-1.5 text-xs">
            <span :class="changeColor(stat.change.direction)">
              {{ changeArrow(stat.change.direction) }} {{ Math.abs(stat.change.pct || 0).toFixed(1) }}%
            </span>
            <span :class="isDark ? 'text-gray-500' : 'text-gray-400'">vs {{ prevLabel }}</span>
          </div>
        </div>
      </div>

      <!-- 图表区域 -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-6 mb-6">
        <!-- 每日活动柱状图 -->
        <div
          class="rounded-xl p-4 sm:p-5 border"
          :class="isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'"
        >
          <h3 class="text-sm font-semibold mb-3" :class="isDark ? 'text-gray-200' : 'text-gray-700'">
            {{ mode === 'weekly' ? '每日活动量' : '每日活动量' }}
          </h3>
          <div ref="dailyChartRef" class="w-full" style="height: 220px"></div>
        </div>

        <!-- 本期 vs 上期 对比图 -->
        <div
          class="rounded-xl p-4 sm:p-5 border"
          :class="isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'"
        >
          <h3 class="text-sm font-semibold mb-3" :class="isDark ? 'text-gray-200' : 'text-gray-700'">
            {{ mode === 'weekly' ? '本周 vs 上周' : '本月 vs 上月' }}
          </h3>
          <div ref="compareChartRef" class="w-full" style="height: 220px"></div>
        </div>
      </div>

      <!-- 详细数据卡片 -->
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
        <!-- 练习题表现 -->
        <div
          class="rounded-xl p-4 sm:p-5 border"
          :class="isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'"
        >
          <h3
            class="text-sm font-semibold mb-4 flex items-center gap-2"
            :class="isDark ? 'text-gray-200' : 'text-gray-700'"
          >
            <svg class="w-4 h-4 text-amber-500" fill="currentColor" viewBox="0 0 20 20">
              <path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" />
              <path
                fill-rule="evenodd"
                d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z"
                clip-rule="evenodd"
              />
            </svg>
            练习题表现
          </h3>
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs" :class="isDark ? 'text-gray-400' : 'text-gray-500'">完成题数</span>
              <span class="text-sm font-semibold" :class="isDark ? 'text-gray-200' : 'text-gray-800'">
                {{ thisData.quizzes_done }}
              </span>
            </div>
            <div>
              <div class="flex items-center justify-between mb-1">
                <span class="text-xs" :class="isDark ? 'text-gray-400' : 'text-gray-500'">平均正确率</span>
                <span class="text-sm font-semibold" :class="accuracyColor(thisData.avg_accuracy)">
                  {{ (thisData.avg_accuracy || 0).toFixed(1) }}%
                </span>
              </div>
              <div class="w-full h-2 rounded-full overflow-hidden" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :class="accuracyBarColor(thisData.avg_accuracy)"
                  :style="{ width: thisData.avg_accuracy + '%' }"
                ></div>
              </div>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-xs" :class="isDark ? 'text-gray-400' : 'text-gray-500'">使用提示次数</span>
              <span class="text-sm font-semibold" :class="isDark ? 'text-gray-200' : 'text-gray-800'">
                {{ thisData.hints_used }}
              </span>
            </div>
          </div>
        </div>

        <!-- 卡片复习 -->
        <div
          class="rounded-xl p-4 sm:p-5 border"
          :class="isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'"
        >
          <h3
            class="text-sm font-semibold mb-4 flex items-center gap-2"
            :class="isDark ? 'text-gray-200' : 'text-gray-700'"
          >
            <svg class="w-4 h-4 text-green-500" fill="currentColor" viewBox="0 0 20 20">
              <path
                d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
              />
            </svg>
            卡片复习
          </h3>
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs" :class="isDark ? 'text-gray-400' : 'text-gray-500'">复习卡片数</span>
              <span class="text-sm font-semibold" :class="isDark ? 'text-gray-200' : 'text-gray-800'">
                {{ thisData.cards_reviewed }}
              </span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-xs" :class="isDark ? 'text-gray-400' : 'text-gray-500'">已掌握</span>
              <span class="text-sm font-semibold text-green-600 dark:text-green-400">
                {{ thisData.mastered_count }}
              </span>
            </div>
            <div>
              <div class="flex items-center justify-between mb-1">
                <span class="text-xs" :class="isDark ? 'text-gray-400' : 'text-gray-500'">掌握率</span>
                <span class="text-sm font-semibold" :class="accuracyColor(thisData.mastery_rate)">
                  {{ (thisData.mastery_rate || 0).toFixed(1) }}%
                </span>
              </div>
              <div class="w-full h-2 rounded-full overflow-hidden" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'">
                <div
                  class="h-full rounded-full transition-all duration-500 bg-green-500"
                  :style="{ width: thisData.mastery_rate + '%' }"
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 学习摘要 -->
      <div
        class="rounded-xl p-4 sm:p-5 border"
        :class="isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'"
      >
        <h3
          class="text-sm font-semibold mb-3 flex items-center gap-2"
          :class="isDark ? 'text-gray-200' : 'text-gray-700'"
        >
          <svg class="w-4 h-4 text-primary-500" fill="currentColor" viewBox="0 0 20 20">
            <path
              fill-rule="evenodd"
              d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
              clip-rule="evenodd"
            />
          </svg>
          学习摘要
        </h3>
        <div class="text-sm leading-relaxed space-y-2" :class="isDark ? 'text-gray-300' : 'text-gray-600'">
          <p>
            {{ mode === 'weekly' ? '本周' : '本月' }}你共学习了
            <strong :class="isDark ? 'text-gray-100' : 'text-gray-900'">{{ thisData.active_days }} 天</strong>
            ， 复习了
            <strong :class="isDark ? 'text-gray-100' : 'text-gray-900'">{{ thisData.cards_reviewed }} 张卡片</strong>
            ， 完成了
            <strong :class="isDark ? 'text-gray-100' : 'text-gray-900'">{{ thisData.quizzes_done }} 道练习题</strong>
            。
          </p>
          <p v-if="thisData.chat_messages > 0">
            你还发送了
            <strong :class="isDark ? 'text-gray-100' : 'text-gray-900'">{{ thisData.chat_messages }} 条对话消息</strong>
            ，
            {{
              thisData.new_materials > 0 ? `上传了 ${thisData.new_materials} 份新材料。` : '继续保持与 AI 的交流吧！'
            }}
          </p>
          <p v-if="thisData.longest_focus_min > 0">
            最长单次专注时长
            <strong :class="isDark ? 'text-gray-100' : 'text-gray-900'">{{ thisData.longest_focus_min }} 分钟</strong>
            ，
            {{ thisData.longest_focus_min >= 25 ? '非常棒！' : '试试番茄钟来提升专注力吧。' }}
          </p>
          <p v-if="overallChangeDirection === 'up'" class="text-green-600 dark:text-green-400 font-medium">
            整体趋势上升，继续保持！
          </p>
          <p v-else-if="overallChangeDirection === 'down'" class="text-amber-600 dark:text-amber-400 font-medium">
            相比{{ prevLabel }}有所下降，加油！
          </p>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="text-center py-16">
      <svg
        class="w-16 h-16 mx-auto mb-4"
        :class="isDark ? 'text-gray-600' : 'text-gray-300'"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
        />
      </svg>
      <p class="text-sm" :class="isDark ? 'text-gray-400' : 'text-gray-500'">加载报告中...</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useDarkMode } from '../composables/useDarkMode'
import { getWeeklyReport, getMonthlyReport } from '../api/client'
import * as echarts from 'echarts'

const { isDark } = useDarkMode()

const mode = ref('weekly') // 'weekly' | 'monthly'
const loading = ref(false)
const report = ref(null)
const currentDate = ref(new Date())

const dailyChartRef = ref(null)
const compareChartRef = ref(null)
const reportContent = ref(null)
let dailyChart = null
let compareChart = null

// ========== 计算属性 ==========

const thisData = computed(() => {
  if (!report.value) return {}
  return mode.value === 'weekly' ? report.value.this_week : report.value.this_month
})

const prevData = computed(() => {
  if (!report.value) return {}
  return mode.value === 'weekly' ? report.value.prev_week : report.value.prev_month
})

const changeData = computed(() => report.value?.change || {})

const prevLabel = computed(() => (mode.value === 'weekly' ? '上周' : '上月'))

const periodLabel = computed(() => {
  if (!report.value) return ''
  const p = report.value.period
  return `${p.start} ~ ${p.end}`
})

const isCurrentPeriod = computed(() => {
  const now = new Date()
  const d = currentDate.value
  if (mode.value === 'weekly') {
    // 同一年同一周
    const nowWeekStart = getMonday(now)
    const curWeekStart = getMonday(d)
    return nowWeekStart.getTime() === curWeekStart.getTime()
  } else {
    return now.getFullYear() === d.getFullYear() && now.getMonth() === d.getMonth()
  }
})

const canNavigateNext = computed(() => {
  return !isCurrentPeriod.value
})

const overallChangeDirection = computed(() => {
  const cd = changeData.value
  if (!cd) return 'same'
  let upCount = 0,
    downCount = 0
  for (const key of ['cards_reviewed', 'quizzes_done', 'chat_messages', 'active_days']) {
    if (cd[key]?.direction === 'up') upCount++
    else if (cd[key]?.direction === 'down') downCount++
  }
  if (upCount > downCount) return 'up'
  if (downCount > upCount) return 'down'
  return 'same'
})

const statCards = computed(() => {
  const d = thisData.value
  const c = changeData.value
  if (!d || !c) return []
  return [
    { key: 'cards', label: '复习卡片', value: d.cards_reviewed || 0, icon: '\u{1F4C7}', change: c.cards_reviewed },
    { key: 'quizzes', label: '完成练习', value: d.quizzes_done || 0, icon: '\u{270D}\u{FE0F}', change: c.quizzes_done },
    {
      key: 'accuracy',
      label: '正确率',
      value: (d.avg_accuracy || 0).toFixed(1) + '%',
      icon: '\u{1F3AF}',
      change: null
    },
    { key: 'messages', label: '对话消息', value: d.chat_messages || 0, icon: '\u{1F4AC}', change: c.chat_messages },
    { key: 'materials', label: '新材料', value: d.new_materials || 0, icon: '\u{1F4DA}', change: c.new_materials },
    { key: 'active', label: '活跃天数', value: d.active_days || 0, icon: '\u{1F525}', change: c.active_days }
  ]
})

// ========== 辅助函数 ==========

function getMonday(d) {
  const date = new Date(d)
  const day = date.getDay()
  const diff = date.getDate() - day + (day === 0 ? -6 : 1)
  date.setDate(diff)
  date.setHours(0, 0, 0, 0)
  return date
}

function changeArrow(dir) {
  if (dir === 'up') return '\u2191'
  if (dir === 'down') return '\u2193'
  return '\u2192'
}

function changeColor(dir) {
  if (dir === 'up') return 'text-green-600 dark:text-green-400'
  if (dir === 'down') return 'text-red-500 dark:text-red-400'
  return isDark.value ? 'text-gray-500' : 'text-gray-400'
}

function accuracyColor(val) {
  if (val >= 80) return 'text-green-600 dark:text-green-400'
  if (val >= 60) return 'text-amber-600 dark:text-amber-400'
  return 'text-red-500 dark:text-red-400'
}

function accuracyBarColor(val) {
  if (val >= 80) return 'bg-green-500'
  if (val >= 60) return 'bg-amber-500'
  return 'bg-red-500'
}

// ========== 数据加载 ==========

async function loadReport() {
  loading.value = true
  report.value = null
  try {
    const dateStr = formatDate(currentDate.value)
    const monthStr = formatMonth(currentDate.value)
    const res = mode.value === 'weekly' ? await getWeeklyReport(dateStr) : await getMonthlyReport(monthStr)
    report.value = res.data
    await nextTick()
    renderCharts()
  } catch (e) {
    console.error('加载报告失败:', e)
  } finally {
    loading.value = false
  }
}

function formatDate(d) {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function formatMonth(d) {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  return `${y}-${m}`
}

function switchMode(m) {
  if (mode.value === m) return
  mode.value = m
  currentDate.value = new Date()
  loadReport()
}

function navigate(dir) {
  const d = new Date(currentDate.value)
  if (mode.value === 'weekly') {
    d.setDate(d.getDate() + dir * 7)
  } else {
    d.setMonth(d.getMonth() + dir)
  }
  // 不允许导航到未来
  if (dir > 0 && d > new Date()) return
  currentDate.value = d
  loadReport()
}

function goToCurrent() {
  currentDate.value = new Date()
  loadReport()
}

// ========== 图表渲染 ==========

function renderCharts() {
  renderDailyChart()
  renderCompareChart()
}

function renderDailyChart() {
  if (!dailyChartRef.value || !report.value?.daily_activity) return

  if (dailyChart) {
    dailyChart.dispose()
  }
  dailyChart = echarts.init(dailyChartRef.value)

  const data = report.value.daily_activity
  const labels = data.map((d) => d.label)
  const values = data.map((d) => d.count)

  const textColor = isDark.value ? '#9ca3af' : '#6b7280'
  const barColor = isDark.value ? '#6366f1' : '#4f46e5'
  const gridColor = isDark.value ? '#374151' : '#e5e7eb'

  dailyChart.setOption({
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark.value ? '#1f2937' : '#fff',
      borderColor: isDark.value ? '#374151' : '#e5e7eb',
      textStyle: { color: isDark.value ? '#e5e7eb' : '#374151', fontSize: 12 },
      formatter: (params) => {
        const p = params[0]
        return `${p.name}<br/>活动次数: <strong>${p.value}</strong>`
      }
    },
    grid: { top: 10, right: 10, bottom: 24, left: 36 },
    xAxis: {
      type: 'category',
      data: labels,
      axisLabel: { color: textColor, fontSize: 11 },
      axisLine: { lineStyle: { color: gridColor } },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      minInterval: 1,
      axisLabel: { color: textColor, fontSize: 11 },
      splitLine: { lineStyle: { color: gridColor, type: 'dashed' } },
      axisLine: { show: false },
      axisTick: { show: false }
    },
    series: [
      {
        type: 'bar',
        data: values,
        barMaxWidth: 32,
        itemStyle: {
          color: barColor,
          borderRadius: [4, 4, 0, 0]
        },
        emphasis: {
          itemStyle: { color: isDark.value ? '#818cf8' : '#6366f1' }
        }
      }
    ]
  })
}

function renderCompareChart() {
  if (!compareChartRef.value || !report.value) return

  if (compareChart) {
    compareChart.dispose()
  }
  compareChart = echarts.init(compareChartRef.value)

  const thisD = thisData.value
  const prevD = prevData.value
  if (!thisD || !prevD) return

  const categories = ['卡片复习', '练习题', '对话', '新材料', '活跃天数']
  const thisValues = [
    thisD.cards_reviewed || 0,
    thisD.quizzes_done || 0,
    thisD.chat_messages || 0,
    thisD.new_materials || 0,
    thisD.active_days || 0
  ]
  const prevValues = [
    prevD.cards_reviewed || 0,
    prevD.quizzes_done || 0,
    prevD.chat_messages || 0,
    prevD.new_materials || 0,
    prevD.active_days || 0
  ]

  const textColor = isDark.value ? '#9ca3af' : '#6b7280'
  const gridColor = isDark.value ? '#374151' : '#e5e7eb'
  const thisColor = isDark.value ? '#6366f1' : '#4f46e5'
  const prevColor = isDark.value ? '#6b7280' : '#d1d5db'

  compareChart.setOption({
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark.value ? '#1f2937' : '#fff',
      borderColor: isDark.value ? '#374151' : '#e5e7eb',
      textStyle: { color: isDark.value ? '#e5e7eb' : '#374151', fontSize: 12 }
    },
    legend: {
      data: [mode.value === 'weekly' ? '本周' : '本月', prevLabel.value],
      top: 0,
      right: 0,
      textStyle: { color: textColor, fontSize: 11 },
      itemWidth: 12,
      itemHeight: 8
    },
    grid: { top: 30, right: 10, bottom: 24, left: 36 },
    xAxis: {
      type: 'category',
      data: categories,
      axisLabel: { color: textColor, fontSize: 10, interval: 0, rotate: categories.length > 5 ? 20 : 0 },
      axisLine: { lineStyle: { color: gridColor } },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      minInterval: 1,
      axisLabel: { color: textColor, fontSize: 11 },
      splitLine: { lineStyle: { color: gridColor, type: 'dashed' } },
      axisLine: { show: false },
      axisTick: { show: false }
    },
    series: [
      {
        name: mode.value === 'weekly' ? '本周' : '本月',
        type: 'bar',
        data: thisValues,
        barMaxWidth: 20,
        itemStyle: { color: thisColor, borderRadius: [3, 3, 0, 0] }
      },
      {
        name: prevLabel.value,
        type: 'bar',
        data: prevValues,
        barMaxWidth: 20,
        itemStyle: { color: prevColor, borderRadius: [3, 3, 0, 0] }
      }
    ]
  })
}

// ========== 导出图片 ==========

function exportReport() {
  if (!reportContent.value) return

  // 获取图表图片
  const dailyImg = dailyChart?.getDataURL({
    type: 'png',
    pixelRatio: 2,
    backgroundColor: isDark.value ? '#1f2937' : '#fff'
  })
  const compareImg = compareChart?.getDataURL({
    type: 'png',
    pixelRatio: 2,
    backgroundColor: isDark.value ? '#1f2937' : '#fff'
  })

  // 用 canvas 拼接
  const canvas = document.createElement('canvas')
  const ctx = canvas.getContext('2d')
  const width = 800
  const padding = 24
  const headerH = 60
  const statsH = 80
  const chartH = 250
  const detailH = 120
  const totalH = headerH + statsH + chartH + detailH + padding * 4

  canvas.width = width
  canvas.height = totalH

  // 背景
  ctx.fillStyle = isDark.value ? '#111827' : '#ffffff'
  ctx.fillRect(0, 0, width, totalH)

  // 标题
  ctx.fillStyle = isDark.value ? '#f3f4f6' : '#111827'
  ctx.font = 'bold 20px sans-serif'
  ctx.fillText(`StudyForge Pro - ${mode.value === 'weekly' ? '周' : '月'}学习报告`, padding, 36)
  ctx.fillStyle = isDark.value ? '#9ca3af' : '#6b7280'
  ctx.font = '13px sans-serif'
  ctx.fillText(periodLabel.value, padding, 54)

  // 统计数据行
  const d = thisData.value
  const statsY = headerH + padding
  const stats = [
    { label: '复习卡片', val: d.cards_reviewed },
    { label: '完成练习', val: d.quizzes_done },
    { label: '正确率', val: (d.avg_accuracy || 0).toFixed(1) + '%' },
    { label: '活跃天数', val: d.active_days },
    { label: '对话消息', val: d.chat_messages }
  ]
  const colW = (width - padding * 2) / stats.length
  stats.forEach((s, i) => {
    const x = padding + i * colW + colW / 2
    ctx.fillStyle = isDark.value ? '#f3f4f6' : '#111827'
    ctx.font = 'bold 22px sans-serif'
    ctx.textAlign = 'center'
    ctx.fillText(String(s.val), x, statsY + 30)
    ctx.fillStyle = isDark.value ? '#9ca3af' : '#6b7280'
    ctx.font = '11px sans-serif'
    ctx.fillText(s.label, x, statsY + 48)
  })
  ctx.textAlign = 'start'

  // 图表图片
  const chartY = headerH + statsH + padding * 2
  const imgW = (width - padding * 3) / 2
  const imgH = chartH - 20

  const drawImg = (src, x, y, w, h) => {
    return new Promise((resolve) => {
      if (!src) {
        resolve()
        return
      }
      const img = new Image()
      img.onload = () => {
        ctx.drawImage(img, x, y, w, h)
        resolve()
      }
      img.onerror = () => resolve()
      img.src = src
    })
  }

  Promise.all([
    drawImg(dailyImg, padding, chartY, imgW, imgH),
    drawImg(compareImg, padding * 2 + imgW, chartY, imgW, imgH)
  ]).then(() => {
    // 底部摘要
    const summaryY = chartY + chartH + padding
    ctx.fillStyle = isDark.value ? '#f3f4f6' : '#111827'
    ctx.font = '13px sans-serif'
    ctx.fillText(
      `${mode.value === 'weekly' ? '本周' : '本月'}共学习 ${d.active_days} 天，复习 ${d.cards_reviewed} 张卡片，完成 ${d.quizzes_done} 道题。`,
      padding,
      summaryY + 16
    )
    ctx.fillStyle = isDark.value ? '#6b7280' : '#9ca3af'
    ctx.font = '11px sans-serif'
    ctx.fillText(`Generated by StudyForge Pro - ${new Date().toLocaleDateString('zh-CN')}`, padding, summaryY + 36)

    // 下载
    canvas.toBlob((blob) => {
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `studyforge-report-${mode.value}-${periodLabel.value.replace(/\s/g, '')}.png`
      a.click()
      URL.revokeObjectURL(url)
    }, 'image/png')
  })
}

// ========== 生命周期 ==========

function handleResize() {
  dailyChart?.resize()
  compareChart?.resize()
}

onMounted(() => {
  loadReport()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  dailyChart?.dispose()
  compareChart?.dispose()
})

// 暗色模式切换时重绘图表
watch(isDark, () => {
  nextTick(() => {
    if (report.value) renderCharts()
  })
})
</script>
