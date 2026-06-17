<template>
  <div class="p-4 sm:p-6 lg:p-8">
    <!-- 页面标题 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-6">
      <div>
        <h1 class="text-2xl font-bold" :class="isDark ? 'text-white' : 'text-gray-900'">知识洞察</h1>
        <p class="text-sm mt-1" :class="isDark ? 'text-gray-400' : 'text-gray-500'">
          发现跨材料的知识关联，探索知识交叉点
        </p>
      </div>
      <button
        :disabled="loading"
        class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200"
        :class="
          isDark ? 'bg-indigo-600 hover:bg-indigo-500 text-white' : 'bg-indigo-600 hover:bg-indigo-700 text-white'
        "
        @click="loadConnections"
      >
        <svg v-if="loading" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
          />
        </svg>
        {{ loading ? '分析中...' : '重新分析' }}
      </button>
    </div>

    <!-- 骨架屏 -->
    <div v-if="loading" class="space-y-6">
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div
          v-for="i in 3"
          :key="i"
          class="rounded-xl p-5 animate-pulse"
          :class="isDark ? 'bg-gray-800' : 'bg-gray-100'"
        >
          <div class="h-8 rounded w-1/2 mb-2" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'"></div>
          <div class="h-4 rounded w-1/3" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'"></div>
        </div>
      </div>
      <div class="rounded-xl p-6 animate-pulse" :class="isDark ? 'bg-gray-800' : 'bg-gray-100'" style="height: 400px">
        <div class="flex items-center justify-center h-full">
          <p class="text-sm animate-pulse" :class="isDark ? 'text-gray-500' : 'text-gray-400'">
            正在分析跨材料知识关联...
          </p>
        </div>
      </div>
    </div>

    <!-- 内容区 -->
    <div v-else-if="insights" class="space-y-6">
      <!-- 统计卡片 -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div
          class="rounded-xl p-5 border"
          :class="isDark ? 'bg-gray-800/50 border-gray-700' : 'bg-white border-gray-200'"
        >
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg flex items-center justify-center bg-indigo-100 dark:bg-indigo-900/30">
              <svg
                class="w-5 h-5 text-indigo-600 dark:text-indigo-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
                />
              </svg>
            </div>
            <div>
              <p class="text-2xl font-bold" :class="isDark ? 'text-white' : 'text-gray-900'">
                {{ insights.total_pairs }}
              </p>
              <p class="text-xs" :class="isDark ? 'text-gray-400' : 'text-gray-500'">知识关联对</p>
            </div>
          </div>
        </div>
        <div
          class="rounded-xl p-5 border"
          :class="isDark ? 'bg-gray-800/50 border-gray-700' : 'bg-white border-gray-200'"
        >
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg flex items-center justify-center bg-rose-100 dark:bg-rose-900/30">
              <svg
                class="w-5 h-5 text-rose-600 dark:text-rose-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z"
                />
              </svg>
            </div>
            <div>
              <p class="text-2xl font-bold" :class="isDark ? 'text-white' : 'text-gray-900'">
                {{ insights.strong_count }}
              </p>
              <p class="text-xs" :class="isDark ? 'text-gray-400' : 'text-gray-500'">强关联</p>
            </div>
          </div>
        </div>
        <div
          class="rounded-xl p-5 border"
          :class="isDark ? 'bg-gray-800/50 border-gray-700' : 'bg-white border-gray-200'"
        >
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg flex items-center justify-center bg-amber-100 dark:bg-amber-900/30">
              <svg
                class="w-5 h-5 text-amber-600 dark:text-amber-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                />
              </svg>
            </div>
            <div>
              <p class="text-2xl font-bold" :class="isDark ? 'text-white' : 'text-gray-900'">
                {{ insights.top_concepts?.length || 0 }}
              </p>
              <p class="text-xs" :class="isDark ? 'text-gray-400' : 'text-gray-500'">交叉概念</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 关联可视化 -->
      <div
        class="rounded-xl border overflow-hidden"
        :class="isDark ? 'bg-gray-800/50 border-gray-700' : 'bg-white border-gray-200'"
      >
        <div
          class="px-5 py-4 border-b flex items-center justify-between"
          :class="isDark ? 'border-gray-700' : 'border-gray-200'"
        >
          <h2 class="text-base font-semibold" :class="isDark ? 'text-white' : 'text-gray-900'">跨材料知识关联图</h2>
          <div class="flex items-center gap-2 text-xs" :class="isDark ? 'text-gray-400' : 'text-gray-500'">
            <span class="flex items-center gap-1">
              <span class="inline-block w-3 h-0.5 bg-indigo-500 rounded"></span>
              强关联
            </span>
            <span class="flex items-center gap-1">
              <span class="inline-block w-3 h-0.5 bg-amber-500 rounded"></span>
              中关联
            </span>
            <span class="flex items-center gap-1">
              <span class="inline-block w-3 h-0.5 bg-gray-400 rounded"></span>
              弱关联
            </span>
          </div>
        </div>
        <div ref="graphRef" class="w-full" style="height: 420px"></div>
      </div>

      <!-- 交叉概念排行 -->
      <div
        v-if="insights.top_concepts && insights.top_concepts.length > 0"
        class="rounded-xl border overflow-hidden"
        :class="isDark ? 'bg-gray-800/50 border-gray-700' : 'bg-white border-gray-200'"
      >
        <div class="px-5 py-4 border-b" :class="isDark ? 'border-gray-700' : 'border-gray-200'">
          <h2 class="text-base font-semibold" :class="isDark ? 'text-white' : 'text-gray-900'">
            交叉概念 Top {{ insights.top_concepts.length }}
          </h2>
          <p class="text-xs mt-0.5" :class="isDark ? 'text-gray-400' : 'text-gray-500'">在多个材料中共同出现的知识点</p>
        </div>
        <div class="p-5 space-y-3">
          <div
            v-for="(concept, idx) in insights.top_concepts"
            :key="idx"
            class="flex items-start gap-3 p-3 rounded-lg transition-colors"
            :class="isDark ? 'hover:bg-gray-700/50' : 'hover:bg-gray-50'"
          >
            <div
              class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0"
              :class="
                idx < 3
                  ? isDark
                    ? 'bg-indigo-900/50 text-indigo-300'
                    : 'bg-indigo-100 text-indigo-700'
                  : isDark
                    ? 'bg-gray-700 text-gray-400'
                    : 'bg-gray-100 text-gray-500'
              "
            >
              {{ idx + 1 }}
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="text-sm font-medium" :class="isDark ? 'text-white' : 'text-gray-900'">
                  {{ concept.concept }}
                </span>
                <span
                  class="text-xs px-2 py-0.5 rounded-full"
                  :class="isDark ? 'bg-indigo-900/30 text-indigo-300' : 'bg-indigo-50 text-indigo-600'"
                >
                  {{ concept.material_count }} 个材料
                </span>
              </div>
              <div class="flex flex-wrap gap-1 mt-1.5">
                <span
                  v-for="(title, tIdx) in concept.material_titles"
                  :key="tIdx"
                  class="text-xs px-2 py-0.5 rounded"
                  :class="isDark ? 'bg-gray-700 text-gray-300' : 'bg-gray-100 text-gray-600'"
                >
                  {{ title }}
                </span>
              </div>
            </div>
            <button
              class="shrink-0 px-2.5 py-1 rounded-md text-xs font-medium transition-colors"
              :class="
                isDark
                  ? 'bg-indigo-600/20 text-indigo-300 hover:bg-indigo-600/40'
                  : 'bg-indigo-50 text-indigo-600 hover:bg-indigo-100'
              "
              @click="debateConcept(concept.concept)"
            >
              深入学习
            </button>
          </div>
        </div>
      </div>

      <!-- 关联详情列表 -->
      <div
        v-if="insights.connections.length > 0"
        class="rounded-xl border overflow-hidden"
        :class="isDark ? 'bg-gray-800/50 border-gray-700' : 'bg-white border-gray-200'"
      >
        <div
          class="px-5 py-4 border-b flex items-center justify-between"
          :class="isDark ? 'border-gray-700' : 'border-gray-200'"
        >
          <h2 class="text-base font-semibold" :class="isDark ? 'text-white' : 'text-gray-900'">关联详情</h2>
          <span class="text-xs" :class="isDark ? 'text-gray-400' : 'text-gray-500'">
            共 {{ insights.connections.length }} 对关联
          </span>
        </div>
        <div
          class="divide-y max-h-96 overflow-y-auto custom-scroll"
          :class="isDark ? 'divide-gray-700' : 'divide-gray-100'"
        >
          <div
            v-for="(conn, idx) in insights.connections"
            :key="idx"
            class="p-4 transition-colors"
            :class="isDark ? 'hover:bg-gray-700/30' : 'hover:bg-gray-50'"
          >
            <div class="flex items-start justify-between gap-3 mb-2">
              <div class="flex items-center gap-2 flex-wrap">
                <router-link
                  :to="`/materials/${conn.material_a.id}`"
                  class="text-sm font-medium hover:underline"
                  :class="isDark ? 'text-indigo-400' : 'text-indigo-600'"
                >
                  {{ conn.material_a.title }}
                </router-link>
                <svg
                  class="w-4 h-4 shrink-0"
                  :class="isDark ? 'text-gray-500' : 'text-gray-400'"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                </svg>
                <router-link
                  :to="`/materials/${conn.material_b.id}`"
                  class="text-sm font-medium hover:underline"
                  :class="isDark ? 'text-indigo-400' : 'text-indigo-600'"
                >
                  {{ conn.material_b.title }}
                </router-link>
              </div>
              <span
                class="shrink-0 text-xs font-bold px-2 py-0.5 rounded-full"
                :class="scoreClass(conn.similarity_score)"
              >
                {{ (conn.similarity_score * 100).toFixed(0) }}%
              </span>
            </div>
            <div class="flex flex-wrap gap-1.5">
              <span
                v-for="(sc, scIdx) in conn.shared_concepts.slice(0, 8)"
                :key="scIdx"
                class="text-xs px-2 py-0.5 rounded-full"
                :class="
                  sc.match_type === 'exact'
                    ? isDark
                      ? 'bg-emerald-900/30 text-emerald-300'
                      : 'bg-emerald-50 text-emerald-700'
                    : isDark
                      ? 'bg-amber-900/30 text-amber-300'
                      : 'bg-amber-50 text-amber-700'
                "
              >
                {{ sc.concept_a }}
                <template v-if="sc.concept_a !== sc.concept_b">↔ {{ sc.concept_b }}</template>
              </span>
              <span
                v-if="conn.shared_concepts.length > 8"
                class="text-xs px-2 py-0.5 rounded-full"
                :class="isDark ? 'bg-gray-700 text-gray-400' : 'bg-gray-100 text-gray-500'"
              >
                +{{ conn.shared_concepts.length - 8 }} 更多
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="flex flex-col items-center justify-center py-20">
      <div
        class="w-20 h-20 rounded-full flex items-center justify-center mb-6"
        :class="isDark ? 'bg-gray-800' : 'bg-gray-100'"
      >
        <svg
          class="w-10 h-10"
          :class="isDark ? 'text-gray-600' : 'text-gray-300'"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
          />
        </svg>
      </div>
      <h3 class="text-lg font-semibold mb-2" :class="isDark ? 'text-gray-300' : 'text-gray-700'">暂无知识关联</h3>
      <p class="text-sm text-center max-w-md mb-6" :class="isDark ? 'text-gray-500' : 'text-gray-400'">
        {{
          insights && insights.materials.length < 2
            ? '至少需要 2 个已完成分析的材料才能发现跨材料关联。'
            : '尚未发现跨材料的知识交叉点，尝试上传更多相关材料。'
        }}
      </p>
      <div class="flex gap-3">
        <router-link
          to="/upload"
          class="px-4 py-2 rounded-lg text-sm font-medium transition-colors bg-indigo-600 hover:bg-indigo-700 text-white"
        >
          上传材料
        </router-link>
        <button
          class="px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          :class="
            isDark ? 'bg-gray-700 hover:bg-gray-600 text-gray-300' : 'bg-gray-100 hover:bg-gray-200 text-gray-700'
          "
          @click="loadConnections"
        >
          重新分析
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { getConnections } from '../api/client'
import { useToast } from '../composables/useToast'
import { useDarkMode } from '../composables/useDarkMode'

const router = useRouter()
const toast = useToast()
const { isDark } = useDarkMode()

const loading = ref(false)
const insights = ref(null)
const graphRef = ref(null)
let chart = null

function scoreClass(score) {
  if (score >= 0.5) return isDark.value ? 'bg-rose-900/30 text-rose-300' : 'bg-rose-50 text-rose-700'
  if (score >= 0.2) return isDark.value ? 'bg-amber-900/30 text-amber-300' : 'bg-amber-50 text-amber-700'
  return isDark.value ? 'bg-gray-700 text-gray-400' : 'bg-gray-100 text-gray-500'
}

function debateConcept(concept) {
  router.push({ path: '/debate', query: { concept } })
}

async function loadConnections() {
  loading.value = true
  try {
    const res = await getConnections()
    insights.value = res.data
    await nextTick()
    if (insights.value.connections.length > 0) {
      renderGraph()
    }
  } catch (e) {
    console.error('加载知识关联失败:', e)
    toast.error('加载知识关联失败')
  } finally {
    loading.value = false
  }
}

function renderGraph() {
  if (!graphRef.value || !insights.value) return
  if (chart) {
    chart.dispose()
  }
  chart = echarts.init(graphRef.value)

  const materials = insights.value.materials || []
  const connections = insights.value.connections || []

  // 构建节点：材料节点 + 概念节点
  const nodes = []
  const nodeSet = new Set()

  // 材料节点
  const materialColors = [
    '#6366f1',
    '#ec4899',
    '#f59e0b',
    '#10b981',
    '#3b82f6',
    '#8b5cf6',
    '#ef4444',
    '#14b8a6',
    '#f97316',
    '#06b6d4'
  ]
  materials.forEach((m, idx) => {
    const id = `mat_${m.id}`
    if (!nodeSet.has(id)) {
      nodeSet.add(id)
      nodes.push({
        id,
        name: m.title.length > 12 ? m.title.slice(0, 12) + '...' : m.title,
        symbol: 'roundRect',
        symbolSize: [120, 40],
        category: 0,
        itemStyle: { color: materialColors[idx % materialColors.length], borderRadius: 8 },
        label: { show: true, color: '#fff', fontSize: 11, fontWeight: 'bold' },
        _isMaterial: true,
        _materialId: m.id
      })
    }
  })

  // 概念节点（交叉概念）
  const conceptNodes = new Map()
  connections.forEach((conn) => {
    conn.shared_concepts.forEach((sc) => {
      const name = sc.concept_a
      if (!conceptNodes.has(name)) {
        conceptNodes.set(name, { count: 0, materials: new Set() })
      }
      const info = conceptNodes.get(name)
      info.count++
      info.materials.add(conn.material_a.id)
      info.materials.add(conn.material_b.id)
    })
  })

  conceptNodes.forEach((info, name) => {
    const id = `concept_${name}`
    if (!nodeSet.has(id)) {
      nodeSet.add(id)
      const size = Math.min(30 + info.count * 5, 50)
      nodes.push({
        id,
        name: name.length > 10 ? name.slice(0, 10) + '...' : name,
        symbol: 'circle',
        symbolSize: size,
        category: 1,
        itemStyle: { color: isDark.value ? '#a78bfa' : '#7c3aed', opacity: 0.85 },
        label: { show: true, color: isDark.value ? '#e5e7eb' : '#374151', fontSize: 10 },
        _isConcept: true,
        _conceptName: name
      })
    }
  })

  // 构建边
  const edges = []
  const edgeSet = new Set()

  connections.forEach((conn) => {
    const matAId = `mat_${conn.material_a.id}`
    const matBId = `mat_${conn.material_b.id}`
    // 材料间直接关联边
    const edgeKey = `${matAId}-${matBId}`
    if (!edgeSet.has(edgeKey)) {
      edgeSet.add(edgeKey)
      const lineWidth = Math.max(1, Math.min(conn.similarity_score * 6, 5))
      const color =
        conn.similarity_score >= 0.5
          ? '#6366f1'
          : conn.similarity_score >= 0.2
            ? '#f59e0b'
            : isDark.value
              ? '#4b5563'
              : '#9ca3af'
      edges.push({
        source: matAId,
        target: matBId,
        lineStyle: { width: lineWidth, color, opacity: 0.6, curveness: 0.15 },
        label: {
          show: true,
          formatter: `${conn.shared_concepts.length} 个共同概念`,
          fontSize: 9,
          color: isDark.value ? '#9ca3af' : '#6b7280'
        },
        _sharedConcepts: conn.shared_concepts,
        _score: conn.similarity_score
      })
    }

    // 材料到概念的边
    conn.shared_concepts.forEach((sc) => {
      const conceptId = `concept_${sc.concept_a}`
      // Material A -> Concept
      const eKeyA = `${matAId}-${conceptId}`
      if (!edgeSet.has(eKeyA) && nodeSet.has(conceptId)) {
        edgeSet.add(eKeyA)
        edges.push({
          source: matAId,
          target: conceptId,
          lineStyle: { width: 1, color: isDark.value ? '#374151' : '#d1d5db', opacity: 0.4, type: 'dashed' }
        })
      }
      // Material B -> Concept
      const eKeyB = `${matBId}-${conceptId}`
      if (!edgeSet.has(eKeyB) && nodeSet.has(conceptId)) {
        edgeSet.add(eKeyB)
        edges.push({
          source: matBId,
          target: conceptId,
          lineStyle: { width: 1, color: isDark.value ? '#374151' : '#d1d5db', opacity: 0.4, type: 'dashed' }
        })
      }
    })
  })

  const option = {
    tooltip: {
      trigger: 'item',
      backgroundColor: isDark.value ? '#1f2937' : '#fff',
      borderColor: isDark.value ? '#374151' : '#e5e7eb',
      textStyle: { color: isDark.value ? '#e5e7eb' : '#374151', fontSize: 12 },
      formatter: (params) => {
        if (params.dataType === 'node') {
          if (params.data._isMaterial) {
            return `<b>📚 ${params.data.name}</b><br/>点击查看详情`
          }
          return `<b>💡 ${params.data.name}</b><br/>交叉概念`
        }
        if (params.dataType === 'edge' && params.data._sharedConcepts) {
          const concepts = params.data._sharedConcepts
            .slice(0, 5)
            .map((sc) => sc.concept_a)
            .join('、')
          const score = (params.data._score * 100).toFixed(0)
          return `<b>关联度: ${score}%</b><br/>共同概念: ${concepts}${params.data._sharedConcepts.length > 5 ? '...' : ''}`
        }
        return ''
      }
    },
    series: [
      {
        type: 'graph',
        layout: 'force',
        data: nodes,
        links: edges,
        roam: true,
        draggable: true,
        force: {
          repulsion: 600,
          edgeLength: [100, 250],
          gravity: 0.1,
          friction: 0.2
        },
        emphasis: {
          focus: 'adjacency',
          lineStyle: { width: 4 }
        },
        categories: [{ name: '材料' }, { name: '概念' }]
      }
    ]
  }

  chart.setOption(option)

  // 点击材料节点跳转详情
  chart.on('click', 'series.graph', (params) => {
    if (params.dataType === 'node' && params.data._isMaterial) {
      router.push(`/materials/${params.data._materialId}`)
    }
    if (params.dataType === 'node' && params.data._isConcept) {
      debateConcept(params.data._conceptName)
    }
  })
}

function handleResize() {
  chart?.resize()
}

watch(isDark, () => {
  if (insights.value?.connections?.length > 0) {
    renderGraph()
  }
})

onMounted(() => {
  loadConnections()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chart?.dispose()
})
</script>

<style scoped>
.custom-scroll::-webkit-scrollbar {
  width: 5px;
}
.custom-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroll::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.3);
  border-radius: 10px;
}
.custom-scroll::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.5);
}
</style>
