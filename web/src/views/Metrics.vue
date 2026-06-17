<template>
  <div class="p-4 sm:p-6 lg:p-8">
    <!-- 顶部标题栏 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
          <svg class="w-6 h-6 text-primary-500" viewBox="0 0 20 20" fill="currentColor">
            <path
              fill-rule="evenodd"
              d="M3 3a1 1 0 000 2v8a2 2 0 002 2h2.586l-1.293 1.293a1 1 0 101.414 1.414L10 15.414l2.293 2.293a1 1 0 001.414-1.414L12.414 15H15a2 2 0 002-2V5a1 1 0 100-2H3zm11.707 4.707a1 1 0 00-1.414-1.414L10 9.586 8.707 8.293a1 1 0 00-1.414 0l-2 2a1 1 0 101.414 1.414L8 10.414l1.293 1.293a1 1 0 001.414 0l4-4z"
              clip-rule="evenodd"
            />
          </svg>
          API 性能监控
        </h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
          实时 API 延迟、请求量、错误率分析
        </p>
      </div>
      <div class="flex items-center gap-3">
        <!-- 时间范围选择 -->
        <div class="flex bg-gray-100 dark:bg-gray-800 rounded-lg p-1">
          <button
            v-for="opt in rangeOptions"
            :key="opt.value"
            class="px-3 py-1.5 text-xs font-medium rounded-md transition-all duration-200"
            :class="
              rangeHours === opt.value
                ? 'bg-primary-500 text-white shadow-sm'
                : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'
            "
            @click="switchRange(opt.value)"
          >
            {{ opt.label }}
          </button>
        </div>
        <!-- 刷新按钮 -->
        <button
          class="p-2 rounded-lg text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          :class="{ 'animate-spin': loading }"
          @click="loadData"
          title="刷新"
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
      </div>
    </div>

    <!-- 加载骨架屏 -->
    <div v-if="loading" class="space-y-6">
      <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
        <div
          v-for="i in 6"
          :key="i"
          class="bg-white dark:bg-gray-800 rounded-xl p-4 animate-pulse"
        >
          <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mb-3"></div>
          <div class="h-6 bg-gray-200 dark:bg-gray-700 rounded w-3/4"></div>
        </div>
      </div>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div
          v-for="i in 4"
          :key="i"
          class="bg-white dark:bg-gray-800 rounded-xl p-4 h-80 animate-pulse"
        >
          <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/3 mb-4"></div>
          <div class="h-48 bg-gray-200 dark:bg-gray-700 rounded"></div>
        </div>
      </div>
    </div>

    <!-- 加载失败状态 -->
    <div
      v-else-if="error"
      class="text-center py-20"
    >
      <svg
        class="w-16 h-16 mx-auto text-red-300 dark:text-red-800 mb-4"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
        />
      </svg>
      <h3 class="text-lg font-medium text-gray-500 dark:text-gray-400 mb-2">加载失败</h3>
      <p class="text-sm text-red-400 dark:text-red-500 max-w-md mx-auto mb-4">
        {{ error }}
      </p>
      <button
        class="inline-flex items-center gap-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white text-sm font-medium rounded-lg transition"
        @click="loadData"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
          />
        </svg>
        重新加载
      </button>
    </div>

    <!-- 空状态 -->
    <div
      v-else-if="data && data.total_requests === 0"
      class="text-center py-20"
    >
      <svg
        class="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-4"
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
      <h3 class="text-lg font-medium text-gray-500 dark:text-gray-400 mb-2">暂无监控数据</h3>
      <p class="text-sm text-gray-400 dark:text-gray-500 max-w-md mx-auto">
        监控中间件会在后台自动记录所有 API 请求的延迟数据。
        请先使用一段时间后再来查看性能报告。
      </p>
      <p class="text-xs text-gray-400 dark:text-gray-500 mt-4">
        时间范围：{{ data?.range_label || rangeLabel }}
      </p>
    </div>

    <!-- 数据面板 -->
    <div v-else class="space-y-6">
      <!-- 概览统计卡片 -->
      <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 sm:gap-4">
        <div
          v-for="card in summaryCards"
          :key="card.label"
          class="bg-white dark:bg-gray-800 rounded-xl p-4 border border-gray-100 dark:border-gray-700"
        >
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">{{ card.label }}</p>
          <p class="text-lg sm:text-xl font-bold" :class="card.color">{{ card.value }}</p>
          <p v-if="card.sub" class="text-xs text-gray-400 dark:text-gray-500 mt-1">{{ card.sub }}</p>
        </div>
      </div>

      <!-- 图表区域 -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- API 延迟折线图 -->
        <div class="bg-white dark:bg-gray-800 rounded-xl p-4 sm:p-5 border border-gray-100 dark:border-gray-700">
          <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-4">
            API 延迟趋势
          </h3>
          <div ref="latencyChartRef" class="w-full h-72"></div>
        </div>

        <!-- 端点延迟柱状图 (Top-10) -->
        <div class="bg-white dark:bg-gray-800 rounded-xl p-4 sm:p-5 border border-gray-100 dark:border-gray-700">
          <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-4">
            端点延迟 Top-10
          </h3>
          <div ref="endpointChartRef" class="w-full h-72"></div>
        </div>

        <!-- 请求量热力图 -->
        <div class="bg-white dark:bg-gray-800 rounded-xl p-4 sm:p-5 border border-gray-100 dark:border-gray-700">
          <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-4">
            请求量热力图（小时×天）
          </h3>
          <div ref="heatmapChartRef" class="w-full h-72"></div>
        </div>

        <!-- 状态码分布饼图 -->
        <div class="bg-white dark:bg-gray-800 rounded-xl p-4 sm:p-5 border border-gray-100 dark:border-gray-700">
          <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-4">
            HTTP 状态码分布
          </h3>
          <div ref="statusChartRef" class="w-full h-72"></div>
        </div>
      </div>

      <!-- 端点详情表格 -->
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 overflow-hidden">
        <div class="p-4 sm:p-5 border-b border-gray-100 dark:border-gray-700">
          <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">
            全部端点统计
            <span class="text-xs font-normal text-gray-400 ml-2">({{ data.endpoints?.length || 0 }} 个端点)</span>
          </h3>
        </div>
        <div class="overflow-x-auto custom-scroll">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-50 dark:bg-gray-900/50">
                <th class="text-left px-4 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">端点</th>
                <th class="text-right px-4 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">请求数</th>
                <th class="text-right px-4 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">平均</th>
                <th class="text-right px-4 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">P50</th>
                <th class="text-right px-4 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">P95</th>
                <th class="text-right px-4 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">P99</th>
                <th class="text-right px-4 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">最大</th>
                <th class="text-right px-4 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">错误率</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
              <tr
                v-for="ep in data.endpoints"
                :key="ep.method + ep.path"
                class="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
              >
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2">
                    <span
                      class="px-1.5 py-0.5 text-[10px] font-bold rounded uppercase"
                      :class="methodClass(ep.method)"
                    >{{ ep.method }}</span>
                    <span class="text-gray-700 dark:text-gray-300 font-mono text-xs truncate max-w-[200px] sm:max-w-xs" :title="ep.path">
                      {{ ep.path }}
                    </span>
                  </div>
                </td>
                <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-400 tabular-nums">{{ ep.count }}</td>
                <td class="px-4 py-3 text-right tabular-nums" :class="latencyColor(ep.avg_ms)">{{ ep.avg_ms }}ms</td>
                <td class="px-4 py-3 text-right tabular-nums text-gray-600 dark:text-gray-400">{{ ep.p50_ms }}ms</td>
                <td class="px-4 py-3 text-right tabular-nums" :class="latencyColor(ep.p95_ms)">{{ ep.p95_ms }}ms</td>
                <td class="px-4 py-3 text-right tabular-nums" :class="latencyColor(ep.p99_ms)">{{ ep.p99_ms }}ms</td>
                <td class="px-4 py-3 text-right tabular-nums text-gray-500 dark:text-gray-400">{{ ep.max_ms }}ms</td>
                <td class="px-4 py-3 text-right">
                  <span
                    class="px-1.5 py-0.5 text-xs rounded"
                    :class="ep.error_rate > 10 ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' : ep.error_rate > 0 ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400' : 'text-gray-400'"
                  >{{ ep.error_rate }}%</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import { getAPIMetrics } from '../api/client'
import { useToast } from '../composables/useToast'
import { useDarkMode } from '../composables/useDarkMode'

const { toast } = useToast()
const { isDark } = useDarkMode()

// ===== 状态 =====
const loading = ref(true)
const data = ref(null)
const error = ref(null)
const rangeHours = ref(24)

const rangeOptions = [
  { label: '1 小时', value: 1 },
  { label: '24 小时', value: 24 },
  { label: '7 天', value: 168 }
]

const rangeLabel = computed(() => {
  const opt = rangeOptions.find((o) => o.value === rangeHours.value)
  return opt ? opt.label : `${rangeHours.value} 小时`
})

// ===== 图表引用 =====
const latencyChartRef = ref(null)
const endpointChartRef = ref(null)
const heatmapChartRef = ref(null)
const statusChartRef = ref(null)

let latencyChart = null
let endpointChart = null
let heatmapChart = null
let statusChart = null

// ===== 概览卡片 =====
const summaryCards = computed(() => {
  if (!data.value) return []
  const d = data.value
  return [
    { label: '总请求数', value: d.total_requests?.toLocaleString() || '0', color: 'text-gray-900 dark:text-white' },
    {
      label: '平均延迟',
      value: `${d.avg_ms}ms`,
      color: d.avg_ms > 500 ? 'text-red-600 dark:text-red-400' : d.avg_ms > 200 ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'
    },
    {
      label: 'P50 延迟',
      value: `${d.p50_ms}ms`,
      color: 'text-blue-600 dark:text-blue-400'
    },
    {
      label: 'P95 延迟',
      value: `${d.p95_ms}ms`,
      color: d.p95_ms > 1000 ? 'text-red-600 dark:text-red-400' : 'text-violet-600 dark:text-violet-400'
    },
    {
      label: '错误率',
      value: `${d.error_rate}%`,
      color: d.error_rate > 5 ? 'text-red-600 dark:text-red-400' : d.error_rate > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400',
      sub: d.error_rate === 0 ? '无错误' : '4xx + 5xx'
    },
    { label: 'QPS', value: d.qps?.toFixed(1) || '0', color: 'text-primary-600 dark:text-primary-400', sub: '请求/秒' }
  ]
})

// ===== 数据加载 =====
async function loadData() {
  loading.value = true
  error.value = null
  try {
    const res = await getAPIMetrics(rangeHours.value)
    data.value = res.data || { total_requests: 0, endpoints: [], time_series: [], heatmap: [], status_dist: [] }
    await nextTick()
    renderCharts()
  } catch (err) {
    const msg = err.response?.data?.error || err.message || '未知错误'
    error.value = msg
    toast.error('加载监控数据失败: ' + msg)
  } finally {
    loading.value = false
  }
}

function switchRange(val) {
  if (rangeHours.value === val) return
  rangeHours.value = val
  loadData()
}

// ===== ECharts 渲染 =====
function chartTextColor() {
  return isDark.value ? '#9CA3AF' : '#6B7280'
}
function chartBgColor() {
  return isDark.value ? '#1F2937' : '#FFFFFF'
}
function chartBorderColor() {
  return isDark.value ? '#374151' : '#E5E7EB'
}

function renderCharts() {
  renderLatencyChart()
  renderEndpointChart()
  renderHeatmapChart()
  renderStatusChart()
}

function renderLatencyChart() {
  if (!latencyChartRef.value || !data.value?.time_series?.length) return

  if (latencyChart) latencyChart.dispose()
  latencyChart = echarts.init(latencyChartRef.value)

  const ts = data.value.time_series
  latencyChart.setOption({
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark.value ? '#1F2937' : '#FFF',
      borderColor: isDark.value ? '#374151' : '#E5E7EB',
      textStyle: { color: chartTextColor(), fontSize: 12 }
    },
    legend: {
      data: ['平均延迟', 'P95 延迟'],
      textStyle: { color: chartTextColor(), fontSize: 11 },
      top: 0,
      right: 0
    },
    grid: { left: 50, right: 20, top: 36, bottom: 30 },
    xAxis: {
      type: 'category',
      data: ts.map((t) => t.time),
      axisLabel: { color: chartTextColor(), fontSize: 10, rotate: ts.length > 12 ? 45 : 0 },
      axisLine: { lineStyle: { color: chartBorderColor() } }
    },
    yAxis: {
      type: 'value',
      name: 'ms',
      nameTextStyle: { color: chartTextColor(), fontSize: 10 },
      axisLabel: { color: chartTextColor(), fontSize: 10 },
      splitLine: { lineStyle: { color: chartBorderColor(), type: 'dashed' } }
    },
    series: [
      {
        name: '平均延迟',
        type: 'line',
        data: ts.map((t) => t.avg_ms),
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        lineStyle: { width: 2, color: '#6366F1' },
        itemStyle: { color: '#6366F1' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: isDark.value ? 'rgba(99,102,241,0.3)' : 'rgba(99,102,241,0.15)' },
            { offset: 1, color: 'rgba(99,102,241,0)' }
          ])
        }
      },
      {
        name: 'P95 延迟',
        type: 'line',
        data: ts.map((t) => t.p95_ms),
        smooth: true,
        symbol: 'diamond',
        symbolSize: 4,
        lineStyle: { width: 2, color: '#F59E0B', type: 'dashed' },
        itemStyle: { color: '#F59E0B' }
      }
    ]
  })
}

function renderEndpointChart() {
  if (!endpointChartRef.value || !data.value?.endpoints?.length) return

  if (endpointChart) endpointChart.dispose()
  endpointChart = echarts.init(endpointChartRef.value)

  const top10 = data.value.endpoints.slice(0, 10).reverse()
  const labels = top10.map((e) => {
    const path = e.path.length > 25 ? '...' + e.path.slice(-22) : e.path
    return `${e.method} ${path}`
  })

  endpointChart.setOption({
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: isDark.value ? '#1F2937' : '#FFF',
      borderColor: isDark.value ? '#374151' : '#E5E7EB',
      textStyle: { color: chartTextColor(), fontSize: 12 },
      formatter: (params) => {
        const ep = top10[params[0].dataIndex]
        return `<strong>${ep.method} ${ep.path}</strong><br/>请求数: ${ep.count}<br/>平均: ${ep.avg_ms}ms<br/>P95: ${ep.p95_ms}ms<br/>错误率: ${ep.error_rate}%`
      }
    },
    grid: { left: 160, right: 30, top: 10, bottom: 10 },
    xAxis: {
      type: 'value',
      name: 'ms',
      nameTextStyle: { color: chartTextColor(), fontSize: 10 },
      axisLabel: { color: chartTextColor(), fontSize: 10 },
      splitLine: { lineStyle: { color: chartBorderColor(), type: 'dashed' } }
    },
    yAxis: {
      type: 'category',
      data: labels,
      axisLabel: { color: chartTextColor(), fontSize: 10 },
      axisLine: { lineStyle: { color: chartBorderColor() } }
    },
    series: [
      {
        type: 'bar',
        data: top10.map((e) => e.avg_ms),
        barWidth: 16,
        itemStyle: {
          borderRadius: [0, 4, 4, 0],
          color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
            { offset: 0, color: '#6366F1' },
            { offset: 1, color: '#818CF8' }
          ])
        },
        label: {
          show: true,
          position: 'right',
          formatter: '{c}ms',
          color: chartTextColor(),
          fontSize: 10
        }
      }
    ]
  })
}

function renderHeatmapChart() {
  if (!heatmapChartRef.value || !data.value?.heatmap?.length) return

  if (heatmapChart) heatmapChart.dispose()
  heatmapChart = echarts.init(heatmapChartRef.value)

  const days = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  const hours = Array.from({ length: 24 }, (_, i) => `${i}:00`)
  const hmData = data.value.heatmap.map((c) => [c.hour, c.day, c.value])
  const maxVal = Math.max(...data.value.heatmap.map((c) => c.value), 1)

  heatmapChart.setOption({
    backgroundColor: 'transparent',
    tooltip: {
      backgroundColor: isDark.value ? '#1F2937' : '#FFF',
      borderColor: isDark.value ? '#374151' : '#E5E7EB',
      textStyle: { color: chartTextColor(), fontSize: 12 },
      formatter: (p) => `${days[p.value[1]]} ${hours[p.value[0]]}<br/>请求数: <strong>${p.value[2]}</strong>`
    },
    grid: { left: 50, right: 40, top: 10, bottom: 30 },
    xAxis: {
      type: 'category',
      data: hours,
      axisLabel: {
        color: chartTextColor(),
        fontSize: 9,
        interval: 2
      },
      axisLine: { lineStyle: { color: chartBorderColor() } },
      splitArea: { show: false }
    },
    yAxis: {
      type: 'category',
      data: days,
      axisLabel: { color: chartTextColor(), fontSize: 10 },
      axisLine: { lineStyle: { color: chartBorderColor() } },
      splitArea: { show: false }
    },
    visualMap: {
      min: 0,
      max: maxVal,
      calculable: false,
      orient: 'vertical',
      right: 0,
      top: 'center',
      inRange: {
        color: isDark.value
          ? ['#1E293B', '#312E81', '#4338CA', '#6366F1', '#818CF8']
          : ['#EEF2FF', '#C7D2FE', '#818CF8', '#6366F1', '#4338CA']
      },
      textStyle: { color: chartTextColor(), fontSize: 10 }
    },
    series: [
      {
        type: 'heatmap',
        data: hmData,
        emphasis: {
          itemStyle: { borderColor: isDark.value ? '#FFF' : '#333', borderWidth: 1 }
        }
      }
    ]
  })
}

function renderStatusChart() {
  if (!statusChartRef.value || !data.value?.status_dist?.length) return

  if (statusChart) statusChart.dispose()
  statusChart = echarts.init(statusChartRef.value)

  const statusColors = {
    200: '#10B981',
    201: '#34D399',
    204: '#6EE7B7',
    301: '#60A5FA',
    302: '#3B82F6',
    304: '#93C5FD',
    400: '#F59E0B',
    401: '#F97316',
    403: '#EF4444',
    404: '#DC2626',
    429: '#A855F7',
    500: '#B91C1C',
    502: '#991B1B',
    503: '#7F1D1D'
  }

  const pieData = data.value.status_dist.map((s) => ({
    name: `${s.status}`,
    value: s.count,
    itemStyle: { color: statusColors[s.status] || (s.status < 400 ? '#6366F1' : s.status < 500 ? '#F59E0B' : '#EF4444') }
  }))

  statusChart.setOption({
    backgroundColor: 'transparent',
    tooltip: {
      backgroundColor: isDark.value ? '#1F2937' : '#FFF',
      borderColor: isDark.value ? '#374151' : '#E5E7EB',
      textStyle: { color: chartTextColor(), fontSize: 12 },
      formatter: (p) => `HTTP ${p.name}<br/>数量: ${p.value} (${p.percent}%)`
    },
    legend: {
      orient: 'vertical',
      right: 10,
      top: 'center',
      textStyle: { color: chartTextColor(), fontSize: 11 }
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        center: ['35%', '50%'],
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: 6,
          borderColor: chartBgColor(),
          borderWidth: 2
        },
        label: {
          show: true,
          position: 'outside',
          formatter: '{b}\n{d}%',
          color: chartTextColor(),
          fontSize: 10
        },
        emphasis: {
          label: { show: true, fontSize: 13, fontWeight: 'bold' },
          itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0,0,0,0.3)' }
        },
        data: pieData
      }
    ]
  })
}

// ===== 辅助函数 =====
function methodClass(method) {
  const classes = {
    GET: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
    POST: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    PUT: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
    DELETE: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    PATCH: 'bg-violet-100 text-violet-700 dark:bg-violet-900/30 dark:text-violet-400'
  }
  return classes[method] || 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400'
}

function latencyColor(ms) {
  if (ms > 1000) return 'text-red-600 dark:text-red-400'
  if (ms > 500) return 'text-amber-600 dark:text-amber-400'
  if (ms > 200) return 'text-yellow-600 dark:text-yellow-400'
  return 'text-emerald-600 dark:text-emerald-400'
}

// ===== 生命周期 =====
function handleResize() {
  latencyChart?.resize()
  endpointChart?.resize()
  heatmapChart?.resize()
  statusChart?.resize()
}

watch(isDark, () => {
  // 暗色模式切换时重绘图表
  nextTick(() => renderCharts())
})

onMounted(() => {
  loadData()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  latencyChart?.dispose()
  endpointChart?.dispose()
  heatmapChart?.dispose()
  statusChart?.dispose()
})
</script>

<style scoped>
.custom-scroll::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
.custom-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroll::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}
:root.dark .custom-scroll::-webkit-scrollbar-thumb {
  background: #4b5563;
}
</style>
