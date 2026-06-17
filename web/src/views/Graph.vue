<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-6xl mx-auto">
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">知识图谱</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">可视化展示知识点之间的关系网络</p>
      </div>
      <div class="flex items-center gap-3">
        <!-- 布局切换 -->
        <div
          v-if="selectedMaterial"
          class="inline-flex rounded-lg border border-gray-200 dark:border-gray-600 overflow-hidden"
        >
          <button
            :class="
              layoutMode === 'hierarchical'
                ? 'bg-primary-50 text-primary-600 dark:bg-primary-900/20 dark:text-primary-400'
                : 'text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700'
            "
            class="px-3 py-1.5 text-sm transition-colors"
            title="层次布局 — 从上到下按层级排列"
            @click="layoutMode = 'hierarchical'"
          >
            <span class="inline-flex items-center gap-1">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
              层次
            </span>
          </button>
          <button
            :class="
              layoutMode === 'force'
                ? 'bg-primary-50 text-primary-600 dark:bg-primary-900/20 dark:text-primary-400'
                : 'text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700'
            "
            class="px-3 py-1.5 text-sm transition-colors border-l border-gray-200 dark:border-gray-600"
            title="力导向布局 — 物理模拟自动排列"
            @click="layoutMode = 'force'"
          >
            <span class="inline-flex items-center gap-1">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              力导
            </span>
          </button>
          <button
            :class="
              layoutMode === 'mindmap'
                ? 'bg-primary-50 text-primary-600 dark:bg-primary-900/20 dark:text-primary-400'
                : 'text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700'
            "
            class="px-3 py-1.5 text-sm transition-colors border-l border-gray-200 dark:border-gray-600"
            title="思维导图 — 从核心概念径向展开"
            @click="layoutMode = 'mindmap'"
          >
            <span class="inline-flex items-center gap-1">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 3v3m0 12v3M5.636 5.636l2.121 2.121m8.486 8.486l2.121 2.121M3 12h3m12 0h3M5.636 18.364l2.121-2.121m8.486-8.486l2.121-2.121"
                />
                <circle cx="12" cy="12" r="3" stroke-width="2" />
              </svg>
              脑图
            </span>
          </button>
        </div>
        <select
          v-model="selectedMaterial"
          class="px-4 py-2 rounded-lg border border-gray-200 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 text-sm"
          @change="loadGraph"
        >
          <option value="">选择材料...</option>
          <option value="all">全部材料 (跨材料图谱)</option>
          <option v-for="m in materials" :key="m.id" :value="m.id">{{ m.title }}</option>
        </select>
        <!-- 缩放适应 -->
        <button
          v-if="selectedMaterial"
          class="px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-600 text-sm text-gray-600 dark:text-gray-400 hover:border-primary-300 hover:text-primary-600 dark:hover:border-primary-500 dark:hover:text-primary-400 transition flex items-center gap-1.5"
          title="缩放适应，显示全部节点"
          @click="fitToView"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5v-4m0 4h-4m4 0l-5-5"
            />
          </svg>
          <span>适应</span>
        </button>
        <!-- 导出 PNG -->
        <button
          v-if="selectedMaterial"
          class="px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-600 text-sm text-gray-600 dark:text-gray-400 hover:border-primary-300 hover:text-primary-600 dark:hover:border-primary-500 dark:hover:text-primary-400 transition flex items-center gap-1.5"
          title="导出图谱为 PNG 图片"
          @click="exportPNG"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
            />
          </svg>
          <span>导出 PNG</span>
        </button>
        <!-- 跨材料关系高亮 -->
        <button
          v-if="isAllMaterials"
          class="px-3 py-2 rounded-lg border text-sm transition flex items-center gap-1.5"
          :class="
            crossMaterialHighlight
              ? 'bg-purple-50 border-purple-300 text-purple-700 dark:bg-purple-900/20 dark:border-purple-700 dark:text-purple-400'
              : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:border-purple-300 hover:text-purple-600 dark:hover:border-purple-500 dark:hover:text-purple-400'
          "
          title="高亮显示连接不同材料的边"
          @click="toggleCrossHighlight()"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
            />
          </svg>
          <span>跨材料关系</span>
        </button>
      </div>
    </div>

    <!-- 空状态 -->
    <div
      v-if="!selectedMaterial"
      class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-16 text-center"
    >
      <p class="text-5xl mb-4">🗺️</p>
      <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">选择一份已分析的材料</h3>
      <p class="text-gray-400 dark:text-gray-500">查看 AI 生成的知识关系图谱</p>
    </div>

    <!-- 图谱容器 -->
    <div
      v-else
      class="relative bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 overflow-hidden"
    >
      <!-- 搜索框浮层 -->
      <div class="absolute top-3 left-3 z-10 flex items-center gap-2">
        <div class="relative">
          <svg
            class="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 dark:text-gray-500"
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
            v-model="searchQuery"
            type="text"
            placeholder="搜索节点..."
            class="pl-8 pr-8 py-1.5 w-52 rounded-lg border border-gray-200 dark:border-gray-600 bg-white/90 dark:bg-gray-700/90 backdrop-blur-sm text-sm dark:text-gray-200 dark:placeholder-gray-400 focus:outline-none focus:border-primary-400 focus:ring-1 focus:ring-primary-200 transition"
            @input="onSearchInput"
          />
          <button
            v-if="searchQuery"
            class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300"
            @click="clearSearch"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <span
          v-if="searchQuery && matchCount >= 0"
          class="text-xs text-gray-400 dark:text-gray-500 bg-white/80 dark:bg-gray-800/80 px-2 py-1 rounded"
        >
          {{ matchCount }} 个匹配
        </span>
      </div>

      <!-- 图例 -->
      <div class="absolute top-3 right-3 z-10 flex flex-wrap gap-2 max-w-xs">
        <!-- 材料图例（跨材料模式） -->
        <span
          v-for="mat in graphMaterials"
          :key="mat.id"
          class="inline-flex items-center gap-1 text-xs bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm border border-gray-100 dark:border-gray-700 rounded-full px-2.5 py-0.5"
        >
          <span class="w-2.5 h-2.5 rounded-full" :style="{ backgroundColor: mat.color }"></span>
          {{ mat.title.length > 8 ? mat.title.slice(0, 8) + '...' : mat.title }}
        </span>
        <!-- 分类图例（单材料模式） -->
        <template v-if="!isAllMaterials">
          <span
            v-for="(color, cat) in activeCategories"
            :key="cat"
            class="inline-flex items-center gap-1 text-xs bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm border border-gray-100 dark:border-gray-700 rounded-full px-2.5 py-0.5"
          >
            <span class="w-2.5 h-2.5 rounded-full" :style="{ backgroundColor: color }"></span>
            {{ cat }}
          </span>
        </template>
      </div>

      <div ref="chartContainer" class="w-full" style="height: 600px"></div>

      <!-- 右键菜单 -->
      <div
        v-if="contextMenu.visible"
        class="absolute z-20 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 py-1 min-w-[160px] animate-fade-in-up"
        :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
        @click.stop
      >
        <button
          class="w-full px-3 py-1.5 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-primary-50 dark:hover:bg-primary-900/20 hover:text-primary-600 dark:hover:text-primary-400 transition flex items-center gap-2"
          @click="ctxShowDetail"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          查看详情
        </button>
        <button
          class="w-full px-3 py-1.5 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-primary-50 dark:hover:bg-primary-900/20 hover:text-primary-600 dark:hover:text-primary-400 transition flex items-center gap-2"
          @click="ctxHighlightConnections"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
            />
          </svg>
          高亮关联
        </button>
        <button
          class="w-full px-3 py-1.5 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-primary-50 dark:hover:bg-primary-900/20 hover:text-primary-600 dark:hover:text-primary-400 transition flex items-center gap-2"
          @click="ctxZoomToNode"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v3m0 0v3m0-3h3m-3 0H7"
            />
          </svg>
          缩放到此节点
        </button>
        <div class="border-t border-gray-100 dark:border-gray-700 my-1"></div>
        <button
          class="w-full px-3 py-1.5 text-left text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition flex items-center gap-2"
          @click="ctxHideNode"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"
            />
          </svg>
          隐藏节点
        </button>
      </div>
    </div>

    <!-- 节点详情 -->
    <div
      v-if="selectedNode"
      class="mt-6 bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-6 animate-fade-in-up"
    >
      <div class="flex items-start justify-between">
        <div class="flex-1">
          <h3 class="text-lg font-bold text-gray-900 dark:text-gray-100 mb-2">{{ selectedNode.name }}</h3>
          <p class="text-sm text-gray-600 dark:text-gray-400">{{ selectedNode.description || '暂无详细描述' }}</p>
        </div>
        <button
          class="ml-4 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition"
          @click="selectedNode = null"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div class="flex items-center gap-3 mt-3">
        <span v-if="selectedNode.category" class="px-2 py-1 bg-primary-50 text-primary-600 rounded text-xs font-medium">
          {{ selectedNode.category }}
        </span>
        <span v-if="selectedNodeConnections.length" class="text-xs text-gray-400 dark:text-gray-500">
          关联 {{ selectedNodeConnections.length }} 个节点
        </span>
      </div>
      <!-- 关联节点列表 -->
      <div v-if="selectedNodeConnections.length" class="mt-3 flex flex-wrap gap-2">
        <button
          v-for="conn in selectedNodeConnections"
          :key="conn"
          class="px-2.5 py-1 bg-gray-50 dark:bg-gray-700 hover:bg-primary-50 dark:hover:bg-primary-900/20 text-gray-600 dark:text-gray-400 hover:text-primary-600 dark:hover:text-primary-400 rounded-full text-xs transition border border-gray-200 dark:border-gray-600 hover:border-primary-200 dark:hover:border-primary-700"
          @click="focusNode(conn)"
        >
          {{ conn }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { getKnowledgeGraph, getAllGraphs, listMaterials } from '../api/client'
import { useDarkMode } from '../composables/useDarkMode'

const { isDark } = useDarkMode()
const route = useRoute()
const router = useRouter()
const chartContainer = ref(null)
const materials = ref([])
const selectedMaterial = ref(route.params.materialId || '')
const selectedNode = ref(null)
const searchQuery = ref('')
const matchCount = ref(-1)
const hiddenNodes = ref(new Set())
const isAllMaterials = ref(false)
const crossMaterialHighlight = ref(false)
const graphMaterials = ref([]) // [{id, title, color}] from /graph/all
const layoutMode = ref('hierarchical') // 'hierarchical' | 'force' | 'mindmap'
let chart = null
let currentGraphData = null
let searchTimer = null

// 右键菜单
const contextMenu = ref({ visible: false, x: 0, y: 0, node: null })

// ==================== 常量 ====================

const CARD_WIDTH = 160
const CARD_HEIGHT = 56

const categoryColors = {
  核心概念: '#6366f1',
  数据单元: '#8b5cf6',
  寻址: '#ec4899',
  协议: '#f59e0b',
  技术: '#10b981',
  设备: '#3b82f6',
  功能: '#06b6d4',
  概念: '#6366f1',
  原理: '#8b5cf6',
  算法: '#ec4899',
  应用: '#10b981',
  模型: '#3b82f6',
  方法: '#06b6d4',
  公式: '#f59e0b'
}

const categoryIcons = {
  核心概念: '◆',
  概念: '◆',
  原理: '⊕',
  数据单元: '□',
  算法: '∑',
  寻址: '⊞',
  应用: '▶',
  技术: '★',
  模型: '△',
  设备: '▣',
  方法: '⊞',
  功能: '◎',
  公式: '∫',
  协议: '↔'
}

// 示例图谱
const demoGraph = {
  nodes: [
    { name: '数据链路层', symbolSize: 50, category: '核心概念', description: 'OSI 模型第二层，负责节点间可靠数据传输' },
    { name: '帧', symbolSize: 35, category: '数据单元', description: '数据链路层的协议数据单元' },
    { name: 'MAC 地址', symbolSize: 35, category: '方法', description: '48位物理地址，标识网络设备' },
    { name: 'CSMA/CD', symbolSize: 30, category: '算法', description: '载波监听多路访问/冲突检测' },
    { name: '以太网', symbolSize: 40, category: '应用', description: '最广泛的局域网技术' },
    { name: '交换机', symbolSize: 35, category: '应用', description: '工作在数据链路层的网络设备' },
    { name: 'VLAN', symbolSize: 30, category: '方法', description: '虚拟局域网，逻辑隔离广播域' },
    { name: 'ARP', symbolSize: 30, category: '算法', description: '地址解析协议，IP 到 MAC 的映射' },
    { name: '差错检测', symbolSize: 28, category: '功能', description: 'CRC 循环冗余校验' },
    { name: '流量控制', symbolSize: 28, category: '功能', description: '滑动窗口机制控制发送速率' }
  ],
  edges: [
    { source: '数据链路层', target: '帧', label: '包含' },
    { source: '数据链路层', target: 'MAC 地址', label: '依赖' },
    { source: '数据链路层', target: '差错检测', label: '包含' },
    { source: '数据链路层', target: '流量控制', label: '包含' },
    { source: '以太网', target: 'CSMA/CD', label: '基于' },
    { source: '以太网', target: '数据链路层', label: '实例化' },
    { source: '交换机', target: '数据链路层', label: '基于' },
    { source: '交换机', target: 'VLAN', label: '包含' },
    { source: '交换机', target: 'MAC 地址', label: '依赖' },
    { source: 'ARP', target: 'MAC 地址', label: '依赖' },
    { source: '帧', target: '差错检测', label: '关联' }
  ]
}

// ==================== 工具函数 ====================

function getCategoryColor(node) {
  return isAllMaterials.value ? node.material_color || '#6366f1' : categoryColors[node.category] || '#6366f1'
}

function getCategoryIcon(category) {
  return categoryIcons[category] || '•'
}

function truncate(str, maxLen) {
  if (!str) return ''
  return str.length > maxLen ? str.slice(0, maxLen) + '…' : str
}

// ==================== 卡片节点构建器 ====================

function buildCardNode(n, overrides = {}) {
  const catColor = getCategoryColor(n)
  const icon = getCategoryIcon(n.category)
  const hasDesc = !!(n.description && n.description.length > 0)
  const cardH = hasDesc ? CARD_HEIGHT : CARD_HEIGHT - 14

  const node = {
    ...n,
    symbol: 'roundRect',
    symbolSize: [CARD_WIDTH, cardH],
    itemStyle: {
      color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
        { offset: 0, color: catColor + '25' },
        { offset: 0.12, color: isDark.value ? '#1e293b' : '#ffffff' },
        { offset: 1, color: isDark.value ? '#1e293b' : '#ffffff' }
      ]),
      borderColor: overrides.borderColor || (isDark.value ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.06)'),
      borderWidth: overrides.borderWidth || 1,
      borderRadius: 8,
      shadowBlur: overrides.shadowBlur ?? 4,
      shadowColor: overrides.shadowColor || (isDark.value ? 'rgba(0,0,0,0.3)' : 'rgba(0,0,0,0.08)'),
      shadowOffsetY: 2,
      opacity: overrides.opacity ?? 1
    },
    label: {
      show: true,
      formatter: hasDesc
        ? `{icon|${icon}} {name|${n.name}}\n{desc|${truncate(n.description, 16)}}`
        : `{icon|${icon}} {name|${n.name}}`,
      rich: {
        icon: {
          fontSize: 14,
          color: overrides.iconColor || catColor,
          padding: [0, 4, 0, 0]
        },
        name: {
          fontSize: 13,
          fontWeight: 'bold',
          color: overrides.nameColor || (isDark.value ? '#e2e8f0' : '#1e293b'),
          width: CARD_WIDTH - 36,
          overflow: 'truncate'
        },
        desc: {
          fontSize: 10,
          color: overrides.descColor || (isDark.value ? '#94a3b8' : '#64748b'),
          width: CARD_WIDTH - 24,
          overflow: 'truncate',
          padding: [3, 0, 0, 18],
          lineHeight: 14
        }
      },
      position: 'inside',
      align: 'left',
      padding: [6, 10]
    }
  }

  // Preserve x,y for hierarchical layout
  if (n.x !== undefined) node.x = n.x
  if (n.y !== undefined) node.y = n.y

  return node
}

// ==================== 边样式构建器 ====================

function buildEdgeConfig(e, overrides = {}) {
  const labelText = e._labelText || ''
  const showLabel = overrides.showLabel !== undefined ? overrides.showLabel : !!labelText
  const isDashed = labelText === '对比' || labelText === '关联'

  return {
    ...e,
    _labelText: labelText,
    label: {
      show: showLabel,
      formatter: showLabel ? ` ${labelText} ` : '',
      fontSize: overrides.labelFontSize || 10,
      fontWeight: '500',
      color: overrides.labelColor || (isDark.value ? '#c4b5fd' : '#6366f1'),
      backgroundColor: overrides.labelBg || (isDark.value ? 'rgba(30,41,59,0.92)' : 'rgba(255,255,255,0.95)'),
      borderRadius: 10,
      padding: [3, 8]
    },
    lineStyle: {
      color: overrides.edgeColor || (isDark.value ? '#6b7280' : '#94a3b8'),
      width: overrides.edgeWidth || 1.5,
      curveness: 0.2,
      opacity: overrides.edgeOpacity ?? 0.6,
      type: isDashed ? [5, 3] : 'solid'
    }
  }
}

// ==================== 层次布局算法 ====================

function computeHierarchicalLayout(nodes, edges, width, height) {
  if (nodes.length === 0) return new Map()

  const nodeNames = new Set(nodes.map((n) => n.name))
  const outgoing = new Map() // parent -> children
  const incoming = new Map() // child -> parents

  nodes.forEach((n) => {
    outgoing.set(n.name, [])
    incoming.set(n.name, [])
  })

  edges.forEach((e) => {
    if (nodeNames.has(e.source) && nodeNames.has(e.target)) {
      outgoing.get(e.source).push(e.target)
      incoming.get(e.target).push(e.source)
    }
  })

  // Find root nodes (no incoming edges)
  let roots = nodes.filter((n) => incoming.get(n.name).length === 0).map((n) => n.name)

  // Prefer "核心概念" category nodes as roots
  if (roots.length > 0) {
    const coreRoots = roots.filter((name) => {
      const node = nodes.find((n) => n.name === name)
      return node && (node.category === '核心概念' || node.category === '概念')
    })
    if (coreRoots.length > 0) roots = coreRoots
  }

  // If no roots (fully cyclic), pick highest-degree node
  if (roots.length === 0) {
    const degree = new Map()
    nodes.forEach((n) => {
      degree.set(n.name, (outgoing.get(n.name)?.length || 0) + (incoming.get(n.name)?.length || 0))
    })
    roots = [
      nodes.reduce((best, n) => ((degree.get(n.name) || 0) > (degree.get(best) || 0) ? n.name : best), nodes[0].name)
    ]
  }

  // BFS layer assignment
  const layer = new Map()
  const visited = new Set()
  const queue = []

  roots.forEach((r) => {
    layer.set(r, 0)
    visited.add(r)
    queue.push(r)
  })

  while (queue.length > 0) {
    const current = queue.shift()
    const currentLayer = layer.get(current)
    for (const child of outgoing.get(current) || []) {
      if (!visited.has(child)) {
        visited.add(child)
        layer.set(child, currentLayer + 1)
        queue.push(child)
      }
    }
  }

  // Handle disconnected nodes
  let maxLayer = 0
  layer.forEach((l) => {
    if (l > maxLayer) maxLayer = l
  })
  nodes.forEach((n) => {
    if (!layer.has(n.name)) {
      maxLayer++
      layer.set(n.name, maxLayer)
    }
  })

  // Group nodes by layer
  const layers = new Map()
  layer.forEach((l, name) => {
    if (!layers.has(l)) layers.set(l, [])
    layers.get(l).push(name)
  })

  const numLayers = maxLayer + 1
  const verticalSpacing = Math.min(130, Math.max(80, (height - 120) / Math.max(numLayers, 1)))
  const topMargin = 50

  // Initial horizontal placement (centered per layer)
  const positions = new Map()

  layers.forEach((nodesInLayer, layerIndex) => {
    const y = topMargin + layerIndex * verticalSpacing
    const hSpacing = Math.min(200, Math.max(170, (width - 100) / Math.max(nodesInLayer.length, 1)))
    const startX = (width - (nodesInLayer.length - 1) * hSpacing) / 2

    nodesInLayer.forEach((name, i) => {
      positions.set(name, { x: startX + i * hSpacing, y })
    })
  })

  // Barycenter refinement: sort each layer by average parent x-position
  for (let l = 1; l < numLayers; l++) {
    const nodesInLayer = layers.get(l) || []
    if (nodesInLayer.length <= 1) continue

    const barycenters = nodesInLayer.map((name) => {
      const parents = (incoming.get(name) || []).filter((p) => layer.get(p) === l - 1)
      if (parents.length === 0) return { name, bc: width / 2 }
      const avgX = parents.reduce((sum, p) => sum + (positions.get(p)?.x || width / 2), 0) / parents.length
      return { name, bc: avgX }
    })
    barycenters.sort((a, b) => a.bc - b.bc)

    const hSpacing = Math.min(200, Math.max(170, (width - 100) / Math.max(nodesInLayer.length, 1)))
    const startX = (width - (nodesInLayer.length - 1) * hSpacing) / 2
    const y = topMargin + l * verticalSpacing

    barycenters.forEach((item, i) => {
      positions.set(item.name, { x: startX + i * hSpacing, y })
    })
  }

  return positions
}

// ==================== 思维导图径向布局 ====================

function computeMindmapLayout(nodes, edges, width, height) {
  if (nodes.length === 0) return new Map()

  const nodeNames = new Set(nodes.map((n) => n.name))
  const adj = new Map() // undirected adjacency for tree traversal

  nodes.forEach((n) => adj.set(n.name, []))
  edges.forEach((e) => {
    if (nodeNames.has(e.source) && nodeNames.has(e.target)) {
      adj.get(e.source).push(e.target)
      adj.get(e.target).push(e.source)
    }
  })

  // Pick root: prefer "核心概念" category, else highest-degree node
  let root = nodes[0].name
  const coreNodes = nodes.filter((n) => n.category === '核心概念' || n.category === '概念')
  if (coreNodes.length > 0) {
    root = coreNodes.reduce(
      (best, n) => ((adj.get(n.name)?.length || 0) > (adj.get(best)?.length || 0) ? n.name : best),
      coreNodes[0].name
    )
  } else {
    root = nodes.reduce(
      (best, n) => ((adj.get(n.name)?.length || 0) > (adj.get(best)?.length || 0) ? n.name : best),
      nodes[0].name
    )
  }

  // BFS to build tree (parent → children) and assign depths
  const parent = new Map()
  const depth = new Map()
  const children = new Map()
  const visited = new Set()

  nodes.forEach((n) => children.set(n.name, []))

  const queue = [root]
  visited.add(root)
  depth.set(root, 0)
  parent.set(root, null)

  while (queue.length > 0) {
    const cur = queue.shift()
    for (const neighbor of adj.get(cur) || []) {
      if (!visited.has(neighbor)) {
        visited.add(neighbor)
        parent.set(neighbor, cur)
        depth.set(neighbor, depth.get(cur) + 1)
        children.get(cur).push(neighbor)
        queue.push(neighbor)
      }
    }
  }

  // Handle disconnected nodes: attach to root at depth 1
  nodes.forEach((n) => {
    if (!visited.has(n.name)) {
      visited.add(n.name)
      parent.set(n.name, root)
      depth.set(n.name, 1)
      children.get(root).push(n.name)
    }
  })

  // Count leaves in each subtree (for angular weight)
  const leafCount = new Map()
  function countLeaves(name) {
    const kids = children.get(name) || []
    if (kids.length === 0) {
      leafCount.set(name, 1)
      return 1
    }
    let count = 0
    for (const child of kids) {
      count += countLeaves(child)
    }
    leafCount.set(name, count)
    return count
  }
  countLeaves(root)

  // Radial layout: assign (x, y) using polar coordinates
  const positions = new Map()
  const cx = width / 2
  const cy = height / 2

  // Radius per depth level
  const maxDepth = Math.max(...nodes.map((n) => depth.get(n.name) || 0), 1)
  const maxRadius = Math.min(width, height) / 2 - 80 // leave margin for card size
  const radiusStep = maxRadius / Math.max(maxDepth, 1)

  // Recursive radial placement
  function placeRadial(name, startAngle, endAngle) {
    const d = depth.get(name)
    const r = d === 0 ? 0 : d * radiusStep
    const midAngle = (startAngle + endAngle) / 2

    const x = cx + r * Math.cos(midAngle)
    const y = cy + r * Math.sin(midAngle)
    positions.set(name, { x, y })

    const kids = children.get(name) || []
    if (kids.length === 0) return

    const totalLeaves = kids.reduce((sum, k) => sum + (leafCount.get(k) || 1), 0)
    let currentAngle = startAngle

    for (const child of kids) {
      const childLeaves = leafCount.get(child) || 1
      const sweep = (endAngle - startAngle) * (childLeaves / totalLeaves)
      placeRadial(child, currentAngle, currentAngle + sweep)
      currentAngle += sweep
    }
  }

  // Start with full circle, offset by -90° so first branch goes up
  placeRadial(root, -Math.PI / 2, -Math.PI / 2 + 2 * Math.PI)

  return positions
}

const activeCategories = computed(() => {
  if (!currentGraphData) return {}
  const cats = {}
  currentGraphData.nodes.forEach((n) => {
    if (n.category && !cats[n.category]) {
      cats[n.category] = categoryColors[n.category] || '#6366f1'
    }
  })
  return cats
})

// 获取某节点的关联节点名称
function getConnectedNodes(nodeName) {
  if (!currentGraphData) return []
  const connected = new Set()
  currentGraphData.edges.forEach((e) => {
    if (e.source === nodeName) connected.add(e.target)
    if (e.target === nodeName) connected.add(e.source)
  })
  return [...connected]
}

const selectedNodeConnections = computed(() => {
  if (!selectedNode.value) return []
  return getConnectedNodes(selectedNode.value.name)
})

// ==================== 图表初始化 ====================

function initChart(data) {
  if (!chartContainer.value) return

  if (chart) chart.dispose()
  chart = echarts.init(chartContainer.value)
  currentGraphData = data

  // 过滤隐藏节点
  const visibleNodes = data.nodes.filter((n) => !hiddenNodes.value.has(n.name))
  const visibleNodeNames = new Set(visibleNodes.map((n) => n.name))
  const visibleEdges = data.edges.filter((e) => visibleNodeNames.has(e.source) && visibleNodeNames.has(e.target))

  // 计算布局坐标
  let layoutPositions = null
  if (layoutMode.value === 'hierarchical') {
    const rect = chartContainer.value.getBoundingClientRect()
    layoutPositions = computeHierarchicalLayout(visibleNodes, visibleEdges, rect.width || 900, 600)
  } else if (layoutMode.value === 'mindmap') {
    const rect = chartContainer.value.getBoundingClientRect()
    layoutPositions = computeMindmapLayout(visibleNodes, visibleEdges, rect.width || 900, rect.height || 600)
  }

  // Build node color map (for edge coloring)
  const nodeColorMap = {}
  visibleNodes.forEach((n) => {
    nodeColorMap[n.name] = getCategoryColor(n)
  })

  // Build card-style nodes
  const nodes = visibleNodes.map((n) => {
    const pos = layoutPositions?.get(n.name)
    const nodeData = pos ? { ...n, x: pos.x, y: pos.y } : { ...n }
    // In force mode, strip stale x/y so force simulation can position freely
    if (layoutMode.value === 'force') {
      delete nodeData.x
      delete nodeData.y
    }
    return buildCardNode(nodeData)
  })

  // Build edges with source-colored lines
  const edges = visibleEdges.map((e) =>
    buildEdgeConfig(
      { ...e, _labelText: e.label || '' },
      { edgeColor: nodeColorMap[e.source] || (isDark.value ? '#6b7280' : '#94a3b8') }
    )
  )

  const option = {
    tooltip: {
      trigger: 'item',
      formatter: (p) => {
        if (p.dataType === 'node') {
          let text = `<b>${p.data.name}</b>`
          if (isAllMaterials.value && p.data.material_title) {
            text += `<br/>材料: ${p.data.material_title}`
          }
          text += `<br/>${p.data.category || ''}`
          text += `<br/>${p.data.description || ''}`
          return text
        }
        if (p.dataType === 'edge') {
          const label = p.data._labelText
            ? ` <span style="color:#6366f1;font-weight:600">${p.data._labelText}</span>`
            : ''
          return `${p.data.source} →${label} ${p.data.target}`
        }
        return `${p.data.source} → ${p.data.target}`
      },
      backgroundColor: isDark.value ? 'rgba(30,41,59,0.95)' : 'rgba(255,255,255,0.95)',
      borderColor: isDark.value ? '#374151' : '#e5e7eb',
      textStyle: { color: isDark.value ? '#d1d5db' : '#374151', fontSize: 12 }
    },
    animationDuration: 1500,
    animationDurationUpdate: 800,
    animationEasingUpdate: 'cubicInOut',
    series: [
      {
        type: 'graph',
        layout: layoutMode.value === 'force' ? 'force' : 'none',
        data: nodes,
        links: edges,
        roam: true,
        draggable: true,
        edgeSymbol: ['none', 'arrow'],
        edgeSymbolSize: [0, 6],
        ...(layoutMode.value === 'force'
          ? {
              force: {
                repulsion: 2000,
                edgeLength: [200, 400],
                gravity: 0.1,
                friction: 0.2,
                layoutAnimation: true
              }
            }
          : {}),
        emphasis: {
          focus: 'adjacency',
          lineStyle: { width: 3, color: '#6366f1' },
          itemStyle: {
            shadowBlur: 12,
            shadowColor: 'rgba(99,102,241,0.3)'
          },
          label: { show: true },
          edgeLabel: { show: true }
        },
        lineStyle: { opacity: 0.8 }
      }
    ]
  }

  chart.setOption(option)

  // 点击节点
  chart.on('click', (params) => {
    if (params.dataType === 'node') {
      selectedNode.value = params.data
      highlightNodeConnections(params.data.name)
      // 跨材料模式：点击节点跳转到对应材料详情页
      if (isAllMaterials.value && params.data.material_id) {
        router.push(`/materials/${params.data.material_id}`)
      }
    }
  })

  // 右键菜单
  chart.on('contextmenu', (params) => {
    if (params.dataType === 'node' && params.event) {
      const e = params.event.event || params.event
      const rect = chartContainer.value.getBoundingClientRect()
      contextMenu.value = {
        visible: true,
        x: e.clientX - rect.left,
        y: e.clientY - rect.top,
        node: params.data
      }
    } else {
      contextMenu.value.visible = false
    }
  })

  // 搜索恢复
  if (searchQuery.value) {
    applySearchHighlight(searchQuery.value)
  }
}

// ==================== 搜索 ====================

function onSearchInput() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    applySearchHighlight(searchQuery.value)
  }, 200)
}

function clearSearch() {
  searchQuery.value = ''
  matchCount.value = -1
  resetNodeStyles()
}

function applySearchHighlight(query) {
  if (!chart || !currentGraphData) return

  const q = query.trim().toLowerCase()
  if (!q) {
    matchCount.value = -1
    resetNodeStyles()
    return
  }

  const matchNames = new Set()
  currentGraphData.nodes.forEach((n) => {
    if (hiddenNodes.value.has(n.name)) return
    if (
      n.name.toLowerCase().includes(q) ||
      (n.category || '').toLowerCase().includes(q) ||
      (n.description || '').toLowerCase().includes(q)
    ) {
      matchNames.add(n.name)
    }
  })

  matchCount.value = matchNames.size

  const series = chart.getOption().series[0]

  const updatedNodes = series.data.map((n) => {
    if (matchNames.has(n.name)) {
      return buildCardNode(n, {
        borderColor: '#fbbf24',
        borderWidth: 3,
        shadowBlur: 12,
        shadowColor: 'rgba(251,191,36,0.4)'
      })
    }
    return buildCardNode(n, { opacity: 0.25 })
  })

  const updatedEdges = series.links.map((e) =>
    buildEdgeConfig(e, {
      edgeOpacity: matchNames.has(e.source) || matchNames.has(e.target) ? 0.7 : 0.08
    })
  )

  chart.setOption({
    series: [{ data: updatedNodes, links: updatedEdges }]
  })
}

function resetNodeStyles() {
  if (!chart || !currentGraphData) return

  const series = chart.getOption().series[0]
  const nodeColorMap = {}
  currentGraphData.nodes.forEach((n) => {
    nodeColorMap[n.name] = getCategoryColor(n)
  })

  const updatedNodes = series.data.map((n) => buildCardNode(n))

  const updatedEdges = series.links.map((e) =>
    buildEdgeConfig(e, {
      edgeColor: nodeColorMap[e.source] || (isDark.value ? '#6b7280' : '#94a3b8')
    })
  )

  chart.setOption({
    series: [{ data: updatedNodes, links: updatedEdges }]
  })
}

// ==================== 节点高亮关联边 ====================

function highlightNodeConnections(nodeName) {
  if (!chart || !currentGraphData) return

  const connectedNames = new Set(getConnectedNodes(nodeName))
  connectedNames.add(nodeName)

  const series = chart.getOption().series[0]

  const updatedNodes = series.data.map((n) => {
    const isCenter = n.name === nodeName
    const isConnected = connectedNames.has(n.name)
    if (isCenter) {
      return buildCardNode(n, {
        borderColor: '#f59e0b',
        borderWidth: 4,
        shadowBlur: 15,
        shadowColor: 'rgba(245,158,11,0.4)'
      })
    }
    if (isConnected) {
      return buildCardNode(n, {
        borderColor: '#6366f1',
        borderWidth: 2,
        shadowBlur: 8,
        shadowColor: 'rgba(99,102,241,0.2)'
      })
    }
    return buildCardNode(n, { opacity: 0.2 })
  })

  const updatedEdges = series.links.map((e) => {
    const isRelated = e.source === nodeName || e.target === nodeName
    return buildEdgeConfig(e, {
      showLabel: isRelated,
      edgeColor: isRelated ? '#6366f1' : isDark.value ? '#4b5563' : '#e2e8f0',
      edgeWidth: isRelated ? 3 : 1,
      edgeOpacity: isRelated ? 1 : 0.1,
      labelColor: isRelated ? '#6366f1' : isDark.value ? '#9ca3af' : '#6b7280'
    })
  })

  chart.setOption({
    series: [{ data: updatedNodes, links: updatedEdges }]
  })
}

// ==================== 跨材料关系高亮 ====================

function toggleCrossHighlight() {
  crossMaterialHighlight.value = !crossMaterialHighlight.value
  updateChart()
}

function updateChart() {
  if (!chart || !currentGraphData) return
  if (crossMaterialHighlight.value) {
    highlightCrossMaterial()
  } else {
    resetNodeStyles()
  }
}

function highlightCrossMaterial() {
  if (!chart || !currentGraphData || !isAllMaterials.value) return

  // 找出跨材料的边（source 和 target 来自不同材料）
  const nodeMaterialMap = {}
  currentGraphData.nodes.forEach((n) => {
    nodeMaterialMap[n.name] = n.material_id
  })

  const crossEdgeSet = new Set()
  const connectedNodes = new Set()
  currentGraphData.edges.forEach((e) => {
    const srcMat = nodeMaterialMap[e.source]
    const tgtMat = nodeMaterialMap[e.target]
    if (srcMat && tgtMat && srcMat !== tgtMat) {
      crossEdgeSet.add(`${e.source}|${e.target}`)
      connectedNodes.add(e.source)
      connectedNodes.add(e.target)
    }
  })

  const series = chart.getOption().series[0]

  const updatedNodes = series.data.map((n) => {
    if (connectedNodes.has(n.name)) {
      return buildCardNode(n, {
        borderColor: '#a855f7',
        borderWidth: 3,
        shadowBlur: 10,
        shadowColor: 'rgba(168,85,247,0.4)'
      })
    }
    return buildCardNode(n, { opacity: 0.15 })
  })

  const updatedEdges = series.links.map((e) => {
    const key = `${e.source}|${e.target}`
    const isCross = crossEdgeSet.has(key)
    return buildEdgeConfig(e, {
      edgeColor: isCross ? '#a855f7' : isDark.value ? '#4b5563' : '#e2e8f0',
      edgeWidth: isCross ? 3 : 1,
      edgeOpacity: isCross ? 1 : 0.08
    })
  })

  chart.setOption({
    series: [{ data: updatedNodes, links: updatedEdges }]
  })
}

// ==================== 缩放适应 ====================

function fitToView() {
  if (!chart) return
  // 重置所有缩放和平移
  chart.dispatchAction({
    type: 'restore'
  })
  // 重新应用当前搜索高亮
  if (searchQuery.value) {
    setTimeout(() => applySearchHighlight(searchQuery.value), 300)
  }
}

// ==================== 右键菜单操作 ====================

function closeContextMenu() {
  contextMenu.value.visible = false
}

function ctxShowDetail() {
  if (contextMenu.value.node) {
    selectedNode.value = contextMenu.value.node
    highlightNodeConnections(contextMenu.value.node.name)
  }
  closeContextMenu()
}

function ctxHighlightConnections() {
  if (contextMenu.value.node) {
    selectedNode.value = contextMenu.value.node
    highlightNodeConnections(contextMenu.value.node.name)
  }
  closeContextMenu()
}

function ctxZoomToNode() {
  if (!chart || !contextMenu.value.node) return
  chart.dispatchAction({
    type: 'highlight',
    seriesIndex: 0,
    name: contextMenu.value.node.name
  })
  selectedNode.value = contextMenu.value.node
  closeContextMenu()
}

function ctxHideNode() {
  if (!contextMenu.value.node) return
  hiddenNodes.value.add(contextMenu.value.node.name)
  if (selectedNode.value?.name === contextMenu.value.node.name) {
    selectedNode.value = null
  }
  initChart(currentGraphData)
  closeContextMenu()
}

// ==================== 聚焦节点 ====================

function focusNode(nodeName) {
  if (!chart) return
  chart.dispatchAction({
    type: 'highlight',
    seriesIndex: 0,
    name: nodeName
  })
  const node = currentGraphData?.nodes.find((n) => n.name === nodeName)
  if (node) {
    selectedNode.value = node
    highlightNodeConnections(nodeName)
  }
}

// ==================== 加载数据 ====================

async function loadGraph() {
  if (!selectedMaterial.value) return
  // 重置隐藏节点
  hiddenNodes.value.clear()
  selectedNode.value = null
  crossMaterialHighlight.value = false

  if (selectedMaterial.value === 'all') {
    isAllMaterials.value = true
    try {
      const res = await getAllGraphs()
      const data = res.data
      graphMaterials.value = data.materials || []
      if (data.nodes && data.nodes.length > 0) {
        await nextTick()
        initChart(data)
      } else {
        await nextTick()
        initChart({ nodes: [], edges: [], materials: [] })
      }
    } catch (e) {
      console.error('跨材料图谱加载失败:', e)
      await nextTick()
      initChart({ nodes: [], edges: [], materials: [] })
    }
    return
  }

  isAllMaterials.value = false
  graphMaterials.value = []

  try {
    const res = await getKnowledgeGraph(selectedMaterial.value)
    const data = res.data
    if (data && data.nodes && data.nodes.length > 0) {
      await nextTick()
      initChart(data)
    } else {
      await nextTick()
      initChart(demoGraph)
    }
  } catch (e) {
    console.error('图谱加载失败，使用示例数据:', e)
    await nextTick()
    initChart(demoGraph)
  }
}

async function loadMaterials() {
  try {
    const res = await listMaterials()
    materials.value = (res.data.data || []).filter((m) => m.status === 'completed')
  } catch (e) {
    console.error('材料列表加载失败:', e)
  }
}

function exportPNG() {
  if (!chart) return
  const url = chart.getDataURL({
    type: 'png',
    pixelRatio: 2,
    backgroundColor: isDark.value ? '#1f2937' : '#fff'
  })
  const a = document.createElement('a')
  a.href = url
  const materialName = materials.value.find((m) => m.id === selectedMaterial.value)?.title || 'graph'
  a.download = `studyforge_graph_${materialName}_${new Date().toISOString().slice(0, 10)}.png`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

// ==================== 全局事件 ====================

function handleGlobalClick() {
  if (contextMenu.value.visible) {
    closeContextMenu()
  }
}

function handleGlobalKeydown(e) {
  if (e.key === 'Escape' && contextMenu.value.visible) {
    closeContextMenu()
  }
}

function handleResize() {
  chart?.resize()
}

// Layout mode change → re-render
watch(layoutMode, () => {
  if (currentGraphData) {
    nextTick(() => initChart(currentGraphData))
  }
})

watch(isDark, () => {
  if (currentGraphData) {
    nextTick(() => initChart(currentGraphData))
  }
})

onMounted(async () => {
  await loadMaterials()
  if (selectedMaterial.value) {
    await loadGraph()
  }
  window.addEventListener('resize', handleResize)
  window.addEventListener('click', handleGlobalClick)
  document.addEventListener('keydown', handleGlobalKeydown)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('click', handleGlobalClick)
  document.removeEventListener('keydown', handleGlobalKeydown)
  clearTimeout(searchTimer)
  chart?.dispose()
})
</script>

<style scoped>
.animate-fade-in-up {
  animation: fadeInUp 0.2s ease-out;
}
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
