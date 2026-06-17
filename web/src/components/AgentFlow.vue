<template>
  <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-6 mb-8">
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-2">
        <div class="w-2 h-2 rounded-full animate-pulse" :class="allDone ? 'bg-green-500' : 'bg-blue-500'"></div>
        <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">
          {{ allDone ? '分析完成' : 'Agent 并发执行中...' }}
        </h2>
      </div>
      <div v-if="allDone && elapsed > 0" class="text-sm text-gray-500 dark:text-gray-400">
        总耗时 {{ (elapsed / 1000).toFixed(1) }}s
      </div>
    </div>

    <!-- ECharts 流程图 -->
    <div ref="chartRef" class="w-full" style="height: 340px"></div>

    <!-- Agent 详情卡片 -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mt-4">
      <div
        v-for="agent in agents"
        :key="agent.name"
        class="rounded-lg border p-3 transition-all duration-500"
        :class="cardClass(agent)"
      >
        <div class="flex items-center justify-between mb-1">
          <span class="text-sm font-medium text-gray-800 dark:text-gray-200">
            {{ agent.icon }} {{ shortName(agent.name) }}
          </span>
          <StatusBadge :agent="agent" />
        </div>
        <div
          v-if="agent.done && agent.output"
          class="text-xs text-gray-600 dark:text-gray-400 mt-1 truncate"
          :title="agent.output"
        >
          {{ agent.output }}
        </div>
        <!-- 质量评分：星级 + 分数 + Judge 评语 tooltip -->
        <div v-if="agent.qualityScore" class="mt-1.5 group/score relative">
          <div class="flex items-center gap-1.5">
            <div class="flex gap-0.5">
              <svg
                v-for="star in 5"
                :key="star"
                class="w-3.5 h-3.5"
                :class="starClass(agent.qualityScore, star)"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"
                />
              </svg>
            </div>
            <span class="text-xs font-medium" :class="scoreClass(agent.qualityScore)">
              {{ agent.qualityScore.toFixed(1) }}
            </span>
          </div>
          <!-- Judge 评语 tooltip -->
          <div
            v-if="agent.judgeComment"
            class="absolute bottom-full left-0 mb-2 w-64 p-2.5 rounded-lg text-xs leading-relaxed bg-gray-900 dark:bg-gray-700 text-gray-100 dark:text-gray-200 shadow-lg opacity-0 invisible group-hover/score:opacity-100 group-hover/score:visible transition-all duration-200 z-20 pointer-events-none"
          >
            <div class="font-medium text-amber-400 dark:text-amber-300 mb-1 flex items-center gap-1">
              <svg class="w-3 h-3" viewBox="0 0 20 20" fill="currentColor">
                <path
                  fill-rule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                  clip-rule="evenodd"
                />
              </svg>
              Judge 评语
            </div>
            {{ agent.judgeComment }}
            <div
              class="absolute top-full left-4 w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-gray-900 dark:border-t-gray-700"
            ></div>
          </div>
        </div>
        <div v-if="agent.duration" class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">
          耗时 {{ (agent.duration / 1000).toFixed(1) }}s
        </div>
      </div>
    </div>

    <!-- Agent 执行时间线 (甘特图) -->
    <AgentTimeline
      v-if="timelineStart > 0"
      :agents="agents"
      :timeline-start="timelineStart"
      :agent-finish-times="agentFinishTimes"
      :analyzing="analyzing"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick, h } from 'vue'
import * as echarts from 'echarts'
import { useDarkMode } from '../composables/useDarkMode'
import AgentTimeline from './AgentTimeline.vue'

const { isDark } = useDarkMode()

const props = defineProps({
  agents: { type: Array, required: true }, // [{name, icon, output, done, qualityScore, judgeComment, duration}]
  analyzing: { type: Boolean, default: false },
  timelineStart: { type: Number, default: 0 },
  agentFinishTimes: { type: Object, default: () => ({}) }
})

const chartRef = ref(null)
let chart = null
let animTimer = null
const elapsed = ref(0)
let startTime = null

const allDone = computed(() => props.agents.every((a) => a.done))

// Agent 名称映射
const agentKeys = ['Analyst', 'QuizMaster', 'CardMaker', 'MapBuilder']

function shortName(name) {
  return name.split(' ')[0]
}

function cardClass(agent) {
  if (agent.error) return 'border-red-200 bg-red-50/50 dark:border-red-800 dark:bg-red-900/20'
  if (agent.done) return 'border-green-200 bg-green-50/30 dark:border-green-800 dark:bg-green-900/20'
  if (props.analyzing) return 'border-blue-200 bg-blue-50/30 animate-pulse dark:border-blue-800 dark:bg-blue-900/20'
  return 'border-gray-200 bg-gray-50/30 dark:border-gray-700 dark:bg-gray-700/30'
}

function scoreClass(score) {
  if (score >= 8) return 'text-green-600 font-medium'
  if (score >= 6) return 'text-amber-600'
  return 'text-red-500'
}

// 星级评分：0-10 分映射 5 颗星，支持半星
function starClass(score, starIndex) {
  const starValue = score / 2 // 10分制转5星制
  if (starIndex <= starValue) return 'text-amber-400 dark:text-amber-300'
  if (starIndex - 0.5 <= starValue) return 'text-amber-300/50 dark:text-amber-400/50'
  return 'text-gray-300 dark:text-gray-600'
}

// 获取每个 Agent 的当前状态
function getAgentState(key) {
  const agent = props.agents.find((a) => a.name.includes(key))
  if (!agent) return 'waiting'
  if (agent.error || (agent.output && agent.output.includes('失败'))) return 'error'
  if (agent.done) return 'done'
  if (props.analyzing) return 'running'
  return 'waiting'
}

// 颜色配置
const stateColors = {
  waiting: '#d1d5db',
  running: '#3b82f6',
  done: '#22c55e',
  error: '#ef4444'
}

const stateGlows = {
  waiting: 0,
  running: 15,
  done: 8,
  error: 12
}

// 构建 ECharts 图谱数据
function buildChartData() {
  const nodes = []
  const links = []

  // Orchestrator 节点（顶部中心）
  const orchState = props.analyzing ? (allDone.value ? 'done' : 'running') : 'waiting'
  nodes.push({
    name: 'Orchestrator',
    x: 300,
    y: 50,
    symbolSize: 50,
    symbol: 'roundRect',
    itemStyle: {
      color: stateColors[orchState],
      shadowBlur: stateGlows[orchState],
      shadowColor: stateColors[orchState] + '80'
    },
    label: {
      show: true,
      formatter: '编排器',
      fontSize: 13,
      fontWeight: 'bold',
      color: '#fff'
    }
  })

  // 4 个 Agent 节点
  const agentXPositions = [90, 230, 370, 510]
  agentKeys.forEach((key, i) => {
    const state = getAgentState(key)
    const labels = { Analyst: '分析师', QuizMaster: '出题官', CardMaker: '卡片师', MapBuilder: '图谱师' }
    const icons = { Analyst: '🔍', QuizMaster: '✏️', CardMaker: '🃏', MapBuilder: '🗺️' }

    nodes.push({
      name: key,
      x: agentXPositions[i],
      y: 170,
      symbolSize: state === 'running' ? 45 : 40,
      symbol: 'circle',
      itemStyle: {
        color: stateColors[state],
        shadowBlur: stateGlows[state],
        shadowColor: stateColors[state] + '80'
      },
      label: {
        show: true,
        formatter: `${icons[key]}\n${labels[key]}`,
        fontSize: 11,
        color: state === 'waiting' ? (isDark.value ? '#9ca3af' : '#6b7280') : '#fff',
        lineHeight: 16
      }
    })

    // Orchestrator → Agent 边
    links.push({
      source: 'Orchestrator',
      target: key,
      lineStyle: {
        color: state === 'waiting' ? (isDark.value ? '#4b5563' : '#e5e7eb') : stateColors[state],
        width: state === 'running' ? 3 : 2,
        curveness: 0.1,
        type: state === 'waiting' ? 'dashed' : 'solid'
      },
      symbol: ['none', 'arrow'],
      symbolSize: 8
    })

    // 输出节点（底部）
    const outputLabels = { Analyst: '分析报告', QuizMaster: '练习题', CardMaker: '知识卡片', MapBuilder: '知识图谱' }
    const outputIcons = { Analyst: '📊', QuizMaster: '📝', CardMaker: '🃏', MapBuilder: '🕸️' }

    nodes.push({
      name: `${key}_output`,
      x: agentXPositions[i],
      y: 280,
      symbolSize: 30,
      symbol: 'roundRect',
      itemStyle: {
        color: state === 'done' ? (isDark.value ? '#1a3a2a' : '#dcfce7') : isDark.value ? '#334155' : '#f9fafb',
        borderColor: state === 'done' ? (isDark.value ? '#22c55e' : '#86efac') : isDark.value ? '#4b5563' : '#e5e7eb',
        borderWidth: 1
      },
      label: {
        show: true,
        formatter: state === 'done' ? `${outputIcons[key]} ${outputLabels[key]}` : outputLabels[key],
        fontSize: 10,
        color: state === 'done' ? (isDark.value ? '#4ade80' : '#16a34a') : isDark.value ? '#9ca3af' : '#9ca3af'
      }
    })

    // Agent → Output 边
    links.push({
      source: key,
      target: `${key}_output`,
      lineStyle: {
        color: state === 'done' ? (isDark.value ? '#22c55e' : '#86efac') : isDark.value ? '#4b5563' : '#e5e7eb',
        width: state === 'done' ? 2 : 1,
        type: state === 'done' ? 'solid' : 'dotted'
      },
      symbol: ['none', 'arrow'],
      symbolSize: 6
    })
  })

  return { nodes, links }
}

// 初始化 / 更新图表
function updateChart() {
  if (!chart) return

  const data = buildChartData()

  chart.setOption({
    series: [
      {
        type: 'graph',
        layout: 'none',
        roam: false,
        animation: true,
        animationDuration: 500,
        animationEasingUpdate: 'cubicInOut',
        data: data.nodes,
        links: data.links,
        edgeSymbol: ['none', 'arrow'],
        edgeSymbolSize: [0, 8],
        emphasis: {
          disabled: true
        },
        lineStyle: {
          opacity: 1
        }
      }
    ]
  })
}

// 运行时的脉冲动画
function startPulse() {
  if (animTimer) return
  animTimer = setInterval(() => {
    if (!chart) return
    if (allDone.value || !props.analyzing) {
      stopPulse()
      updateChart()
      return
    }
    // 刷新图表以更新 running 状态的脉冲效果
    updateChart()
  }, 1500)
}

function stopPulse() {
  if (animTimer) {
    clearInterval(animTimer)
    animTimer = null
  }
}

// 监听 agents 变化（Agent 完成时刷新图表）
watch(
  () =>
    props.agents.map((a) => ({
      done: a.done,
      output: a.output,
      qualityScore: a.qualityScore,
      judgeComment: a.judgeComment,
      duration: a.duration
    })),
  () => {
    updateChart()
    if (allDone.value) {
      stopPulse()
      if (startTime) {
        elapsed.value = Date.now() - startTime
      }
    }
  },
  { deep: true }
)

watch(
  () => props.analyzing,
  (val) => {
    if (val) {
      startTime = Date.now()
      elapsed.value = 0
      startPulse()
    } else {
      stopPulse()
      if (startTime) elapsed.value = Date.now() - startTime
    }
    updateChart()
  }
)

watch(isDark, () => {
  nextTick(() => updateChart())
})

onMounted(() => {
  nextTick(() => {
    if (chartRef.value) {
      chart = echarts.init(chartRef.value)
      updateChart()
    }
    if (props.analyzing) startPulse()
  })

  window.addEventListener('resize', handleResize)
})

function handleResize() {
  chart?.resize()
}

onUnmounted(() => {
  stopPulse()
  chart?.dispose()
  window.removeEventListener('resize', handleResize)
})

// StatusBadge 子组件
const StatusBadge = {
  props: ['agent'],
  setup(p) {
    return () => {
      if (p.agent.error || (p.agent.output && p.agent.output.includes('失败'))) {
        return h('span', { class: 'text-xs text-red-500 font-medium' }, '✗ 失败')
      }
      if (p.agent.done) {
        return h('span', { class: 'text-xs text-green-600 font-medium' }, '✓')
      }
      if (props.analyzing) {
        return h('span', { class: 'text-xs text-blue-500 animate-pulse' }, '...')
      }
      return h('span', { class: 'text-xs text-gray-400' }, '—')
    }
  }
}
</script>
