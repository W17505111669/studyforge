<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-6xl mx-auto">
    <!-- 页头 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-6">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white">学情诊断</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">多维度分析你的学习弱点，精准定位薄弱环节</p>
      </div>
      <button
        :disabled="loading"
        class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition text-sm flex items-center gap-2 self-start"
        @click="loadDiagnosis"
      >
        <svg v-if="loading" class="animate-spin h-4 w-4" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        <svg v-else class="w-4 h-4" viewBox="0 0 20 20" fill="currentColor">
          <path
            fill-rule="evenodd"
            d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0113.352 2.541l1.153-.577A1 1 0 0121 8.682l-1.268.634a7.002 7.002 0 01-5.5 6.153V17a1 1 0 11-2 0v-1.531a7.002 7.002 0 01-5.5-6.153L5.47 8.682A1 1 0 016.6 7.148l1.153.577A7.002 7.002 0 013 5.101V3a1 1 0 011-1z"
            clip-rule="evenodd"
          />
        </svg>
        {{ loading ? '分析中...' : '重新诊断' }}
      </button>
    </div>

    <!-- 加载骨架屏 -->
    <div v-if="loading" class="space-y-6">
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div class="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700">
          <div class="skeleton h-6 w-32 mb-4 rounded bg-gray-200 dark:bg-gray-700 animate-pulse"></div>
          <div class="skeleton h-64 w-full rounded bg-gray-200 dark:bg-gray-700 animate-pulse"></div>
        </div>
        <div class="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700">
          <div class="skeleton h-6 w-32 mb-4 rounded bg-gray-200 dark:bg-gray-700 animate-pulse"></div>
          <div class="space-y-3">
            <div
              v-for="i in 3"
              :key="i"
              class="skeleton h-20 w-full rounded bg-gray-200 dark:bg-gray-700 animate-pulse"
            ></div>
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div
      v-else-if="!diagnosis || (diagnosis.weak_points.length === 0 && diagnosis.quiz_stats.total_attempts === 0)"
      class="text-center py-16"
    >
      <svg class="w-20 h-20 mx-auto text-green-400 mb-4" viewBox="0 0 20 20" fill="currentColor">
        <path
          fill-rule="evenodd"
          d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
          clip-rule="evenodd"
        />
      </svg>
      <h3 class="text-lg font-semibold text-gray-700 dark:text-gray-200 mb-2">学习状态良好！</h3>
      <p class="text-gray-500 dark:text-gray-400 max-w-md mx-auto">
        当前没有发现明显的薄弱环节。继续保持学习习惯，积累更多答题记录可以获得更精准的诊断。
      </p>
      <div class="mt-6 flex items-center justify-center gap-3">
        <router-link
          to="/cards"
          class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition text-sm"
        >
          去学习卡片
        </router-link>
        <router-link
          to="/quiz"
          class="px-4 py-2 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition text-sm"
        >
          开始练习
        </router-link>
      </div>
    </div>

    <!-- 诊断结果 -->
    <div v-else class="space-y-6">
      <!-- 统计概览 -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4">
        <div class="bg-white dark:bg-gray-800 rounded-xl p-4 border border-gray-200 dark:border-gray-700">
          <div class="text-2xl font-bold text-indigo-600 dark:text-indigo-400">
            {{ Math.round(diagnosis.overall_score) }}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">综合得分</div>
        </div>
        <div class="bg-white dark:bg-gray-800 rounded-xl p-4 border border-gray-200 dark:border-gray-700">
          <div class="text-2xl font-bold text-emerald-600 dark:text-emerald-400">{{ diagnosis.card_stats.total }}</div>
          <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">知识卡片</div>
        </div>
        <div class="bg-white dark:bg-gray-800 rounded-xl p-4 border border-gray-200 dark:border-gray-700">
          <div
            class="text-2xl font-bold"
            :class="
              diagnosis.quiz_stats.accuracy >= 60
                ? 'text-green-600 dark:text-green-400'
                : 'text-amber-600 dark:text-amber-400'
            "
          >
            {{ diagnosis.quiz_stats.accuracy ? Math.round(diagnosis.quiz_stats.accuracy) + '%' : '--' }}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">答题正确率</div>
        </div>
        <div class="bg-white dark:bg-gray-800 rounded-xl p-4 border border-gray-200 dark:border-gray-700">
          <div class="text-2xl font-bold text-red-500 dark:text-red-400">{{ diagnosis.weak_points.length }}</div>
          <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">薄弱环节</div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- 雷达图 -->
        <div class="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">能力雷达图</h2>
          <div ref="radarChartRef" class="w-full" style="height: 300px"></div>
          <div class="mt-3 grid grid-cols-5 gap-1 text-center text-xs">
            <div v-for="dim in diagnosis.radar_dimensions" :key="dim.name">
              <span class="text-gray-500 dark:text-gray-400">{{ dim.name }}</span>
              <div class="font-semibold mt-0.5" :class="scoreColor(dim.score)">{{ Math.round(dim.score) }}</div>
            </div>
          </div>
        </div>

        <!-- 薄弱环节列表 -->
        <div class="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            薄弱环节
            <span class="text-sm font-normal text-gray-500 dark:text-gray-400 ml-2">按严重程度排序</span>
          </h2>
          <div
            v-if="diagnosis.weak_points.length === 0"
            class="text-center py-8 text-gray-500 dark:text-gray-400 text-sm"
          >
            暂无明显薄弱环节，继续保持！
          </div>
          <div v-else class="space-y-3 max-h-80 overflow-y-auto custom-scroll pr-1">
            <div
              v-for="(wp, idx) in diagnosis.weak_points"
              :key="idx"
              class="p-3 rounded-lg border transition hover:shadow-sm"
              :class="severityBorder(wp.severity)"
            >
              <div class="flex items-start justify-between gap-2">
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2 mb-1 flex-wrap">
                    <span class="text-xs px-1.5 py-0.5 rounded font-medium" :class="dimBadge(wp.dimension)">
                      {{ dimLabel(wp.dimension) }}
                    </span>
                    <span class="text-sm font-medium text-gray-900 dark:text-white truncate">{{ wp.label }}</span>
                    <span class="text-xs px-1.5 py-0.5 rounded-full font-medium" :class="severityBadge(wp.severity)">
                      {{ severityLabel(wp.severity) }}
                    </span>
                  </div>
                  <p class="text-xs text-gray-600 dark:text-gray-400 mb-1">{{ wp.description }}</p>
                  <p class="text-xs text-indigo-600 dark:text-indigo-400">{{ wp.suggestion }}</p>
                </div>
                <router-link
                  v-if="wp.action_url"
                  :to="wp.action_url"
                  class="shrink-0 text-xs text-indigo-600 dark:text-indigo-400 hover:underline mt-1"
                >
                  前往 →
                </router-link>
              </div>
              <div class="mt-2 text-xs text-gray-400 dark:text-gray-500">{{ wp.metric }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 题型 & 难度分析 -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div
          v-if="Object.keys(diagnosis.type_accuracy).length"
          class="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700"
        >
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-3">题型正确率</h3>
          <div class="space-y-2">
            <div v-for="(rate, type) in diagnosis.type_accuracy" :key="type" class="flex items-center gap-3">
              <span class="text-xs text-gray-600 dark:text-gray-400 w-16 shrink-0">
                {{ typeLabelMap[type] || type }}
              </span>
              <div class="flex-1 bg-gray-200 dark:bg-gray-700 rounded-full h-3 overflow-hidden">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :style="{ width: Math.max(rate, 3) + '%' }"
                  :class="rateColor(rate)"
                ></div>
              </div>
              <span
                class="text-xs font-medium w-10 text-right"
                :class="rate >= 60 ? 'text-green-600 dark:text-green-400' : 'text-red-500 dark:text-red-400'"
              >
                {{ Math.round(rate) }}%
              </span>
            </div>
          </div>
        </div>

        <div
          v-if="Object.keys(diagnosis.difficulty_accuracy).length"
          class="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700"
        >
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-3">难度正确率</h3>
          <div class="space-y-2">
            <div v-for="(rate, diff) in diagnosis.difficulty_accuracy" :key="diff" class="flex items-center gap-3">
              <span class="text-xs text-gray-600 dark:text-gray-400 w-16 shrink-0">
                {{ diffLabelMap[diff] || diff }}
              </span>
              <div class="flex-1 bg-gray-200 dark:bg-gray-700 rounded-full h-3 overflow-hidden">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :style="{ width: Math.max(rate, 3) + '%' }"
                  :class="rateColor(rate)"
                ></div>
              </div>
              <span
                class="text-xs font-medium w-10 text-right"
                :class="rate >= 60 ? 'text-green-600 dark:text-green-400' : 'text-red-500 dark:text-red-400'"
              >
                {{ Math.round(rate) }}%
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 生成针对性练习按钮 -->
      <div
        v-if="diagnosis.weak_points.length > 0"
        class="bg-gradient-to-r from-indigo-500 to-purple-600 rounded-xl p-6 text-white"
      >
        <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <h3 class="text-lg font-semibold mb-1">生成针对性练习</h3>
            <p class="text-sm text-indigo-100">基于你的薄弱环节，AI 将自动生成一份 10 题专项练习</p>
          </div>
          <button
            :disabled="generating"
            class="shrink-0 px-5 py-2.5 bg-white text-indigo-700 font-medium rounded-lg hover:bg-indigo-50 disabled:opacity-60 transition text-sm flex items-center gap-2"
            @click="generateTargetedQuiz"
          >
            <svg v-if="generating" class="animate-spin h-4 w-4" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            {{ generating ? '生成中...' : '开始生成' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { getDiagnosis } from '../api/client'
import { useToast } from '../composables/useToast'
import { useDarkMode } from '../composables/useDarkMode'

const router = useRouter()
const toast = useToast()
const { isDark } = useDarkMode()

const loading = ref(false)
const generating = ref(false)
const diagnosis = ref(null)
const radarChartRef = ref(null)
let radarChart = null

const typeLabelMap = { choice: '选择题', fill: '填空题', short_answer: '简答题', judge: '判断题' }
const diffLabelMap = { easy: '简单', medium: '中等', hard: '困难' }

async function loadDiagnosis() {
  loading.value = true
  try {
    const res = await getDiagnosis()
    diagnosis.value = res.data.diagnosis
    await nextTick()
    renderRadar()
  } catch (e) {
    console.error('诊断加载失败:', e)
    toast.error('加载诊断数据失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

function renderRadar() {
  if (!radarChartRef.value || !diagnosis.value?.radar_dimensions?.length) return
  if (radarChart) radarChart.dispose()
  radarChart = echarts.init(radarChartRef.value)

  const dims = diagnosis.value.radar_dimensions
  const indicator = dims.map((d) => ({ name: d.name, max: 100 }))
  const values = dims.map((d) => d.score)

  const areaColor = isDark.value ? 'rgba(99, 102, 241, 0.25)' : 'rgba(99, 102, 241, 0.15)'
  const lineColor = '#6366f1'
  const axisColor = isDark.value ? '#4b5563' : '#d1d5db'
  const textColor = isDark.value ? '#9ca3af' : '#6b7280'

  radarChart.setOption({
    radar: {
      indicator,
      shape: 'polygon',
      splitNumber: 4,
      axisName: { color: textColor, fontSize: 12, fontWeight: 'bold' },
      splitLine: { lineStyle: { color: axisColor } },
      splitArea: {
        areaStyle: {
          color: isDark.value
            ? ['rgba(55,65,81,0.3)', 'rgba(55,65,81,0.15)']
            : ['rgba(249,250,251,0.5)', 'rgba(243,244,246,0.5)']
        }
      },
      axisLine: { lineStyle: { color: axisColor } }
    },
    series: [
      {
        type: 'radar',
        data: [
          {
            value: values,
            name: '能力分布',
            symbol: 'circle',
            symbolSize: 6,
            lineStyle: { color: lineColor, width: 2 },
            areaStyle: { color: areaColor },
            itemStyle: { color: lineColor, borderColor: '#fff', borderWidth: 2 }
          }
        ]
      }
    ],
    tooltip: {
      trigger: 'item',
      backgroundColor: isDark.value ? '#1f2937' : '#fff',
      borderColor: isDark.value ? '#374151' : '#e5e7eb',
      textStyle: { color: isDark.value ? '#e5e7eb' : '#374151' }
    }
  })
}

function scoreColor(score) {
  if (score >= 70) return 'text-green-600 dark:text-green-400'
  if (score >= 50) return 'text-amber-600 dark:text-amber-400'
  return 'text-red-500 dark:text-red-400'
}

function rateColor(rate) {
  if (rate >= 80) return 'bg-green-500 dark:bg-green-400'
  if (rate >= 60) return 'bg-emerald-500 dark:bg-emerald-400'
  if (rate >= 40) return 'bg-amber-500 dark:bg-amber-400'
  return 'bg-red-500 dark:bg-red-400'
}

function severityBorder(severity) {
  if (severity === 'high') return 'border-red-300 dark:border-red-700 bg-red-50/50 dark:bg-red-900/10'
  if (severity === 'medium') return 'border-amber-300 dark:border-amber-700 bg-amber-50/50 dark:bg-amber-900/10'
  return 'border-green-300 dark:border-green-700 bg-green-50/50 dark:bg-green-900/10'
}

function severityBadge(severity) {
  if (severity === 'high') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
  if (severity === 'medium') return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
  return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
}

function severityLabel(severity) {
  if (severity === 'high') return '严重'
  if (severity === 'medium') return '中等'
  return '轻微'
}

function dimBadge(dim) {
  const map = {
    card: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    tag: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
    material: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
    type: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
    difficulty: 'bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-400'
  }
  return map[dim] || 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300'
}

function dimLabel(dim) {
  const map = { card: '卡片', tag: '标签', material: '材料', type: '题型', difficulty: '难度' }
  return map[dim] || dim
}

async function generateTargetedQuiz() {
  generating.value = true
  try {
    // 跳转到练习页并带上推荐难度参数
    router.push({ path: '/quiz', query: { recommended: 'true' } })
    toast.success('已跳转到智能推荐练习')
  } finally {
    generating.value = false
  }
}

watch(isDark, () => {
  nextTick(() => renderRadar())
})

function handleResize() {
  radarChart?.resize()
}
onMounted(() => {
  loadDiagnosis()
  window.addEventListener('resize', handleResize)
})
onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  radarChart?.dispose()
})
</script>

<style scoped>
.custom-scroll::-webkit-scrollbar {
  width: 4px;
}
.custom-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroll::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 2px;
}
.dark .custom-scroll::-webkit-scrollbar-thumb {
  background: #4b5563;
}
</style>
