<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-4xl mx-auto">
    <!-- 页面标题 -->
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">数据导出</h1>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">导出你的学习数据为 CSV 或 JSON 格式，方便备份和分析</p>
    </div>

    <!-- 数据类型选择 -->
    <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 sm:p-6 mb-4">
      <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">选择数据类型</h2>
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <label
          v-for="item in dataTypes"
          :key="item.value"
          class="relative flex items-center p-3 rounded-lg border cursor-pointer transition-all"
          :class="
            selectedTypes.has(item.value)
              ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20 dark:border-primary-400'
              : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
          "
        >
          <input
            type="checkbox"
            :value="item.value"
            :checked="selectedTypes.has(item.value)"
            class="sr-only"
            @change="toggleType(item.value)"
          />
          <div class="flex-1 ml-1">
            <div class="flex items-center gap-2">
              <span class="w-5 h-5" :class="item.color" v-html="item.icon"></span>
              <span class="text-sm font-medium text-gray-900 dark:text-gray-100">{{ item.label }}</span>
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ item.desc }}</p>
          </div>
          <div
            class="w-5 h-5 rounded border-2 flex items-center justify-center ml-2 shrink-0"
            :class="
              selectedTypes.has(item.value)
                ? 'bg-primary-500 border-primary-500'
                : 'border-gray-300 dark:border-gray-600'
            "
          >
            <svg
              v-if="selectedTypes.has(item.value)"
              class="w-3 h-3 text-white"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fill-rule="evenodd"
                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                clip-rule="evenodd"
              />
            </svg>
          </div>
        </label>
      </div>
      <div class="mt-3 flex items-center gap-3">
        <button class="text-xs text-primary-600 dark:text-primary-400 hover:underline" @click="selectAll">全选</button>
        <span class="text-gray-300 dark:text-gray-600">|</span>
        <button class="text-xs text-gray-500 dark:text-gray-400 hover:underline" @click="clearAll">清除</button>
      </div>
    </div>

    <!-- 格式选择 -->
    <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 sm:p-6 mb-4">
      <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">导出格式</h2>
      <div class="flex gap-3">
        <button
          v-for="fmt in formats"
          :key="fmt.value"
          class="flex-1 sm:flex-none px-4 py-3 rounded-lg border-2 transition-all"
          :class="
            format === fmt.value
              ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
              : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
          "
          @click="format = fmt.value"
        >
          <div class="flex items-center gap-2 justify-center">
            <span
              class="w-5 h-5"
              :class="
                format === fmt.value ? 'text-primary-600 dark:text-primary-400' : 'text-gray-400 dark:text-gray-500'
              "
              v-html="fmt.icon"
            ></span>
            <span
              class="text-sm font-medium"
              :class="
                format === fmt.value ? 'text-primary-700 dark:text-primary-300' : 'text-gray-700 dark:text-gray-300'
              "
            >
              {{ fmt.label }}
            </span>
          </div>
          <p
            class="text-xs mt-1"
            :class="
              format === fmt.value ? 'text-primary-500 dark:text-primary-400' : 'text-gray-400 dark:text-gray-500'
            "
          >
            {{ fmt.desc }}
          </p>
        </button>
      </div>
    </div>

    <!-- 时间范围 -->
    <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 sm:p-6 mb-4">
      <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">时间范围</h2>
      <div class="flex flex-wrap gap-2 mb-3">
        <button
          v-for="preset in datePresets"
          :key="preset.label"
          class="px-3 py-1.5 text-xs rounded-full transition-all"
          :class="
            activePreset === preset.label
              ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
              : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-400 dark:hover:bg-gray-600'
          "
          @click="applyDatePreset(preset)"
        >
          {{ preset.label }}
        </button>
      </div>
      <div class="flex flex-col sm:flex-row gap-3">
        <div class="flex-1">
          <label class="block text-xs text-gray-500 dark:text-gray-400 mb-1">开始日期</label>
          <input
            v-model="dateFrom"
            type="date"
            class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
          />
        </div>
        <div class="flex items-end pb-2 text-gray-400 dark:text-gray-500">
          <svg class="w-5 h-5 hidden sm:block" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
          </svg>
        </div>
        <div class="flex-1">
          <label class="block text-xs text-gray-500 dark:text-gray-400 mb-1">结束日期</label>
          <input
            v-model="dateTo"
            type="date"
            class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
          />
        </div>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="flex flex-col sm:flex-row gap-3 mb-6">
      <button
        :disabled="selectedTypes.size === 0 || previewLoading"
        class="flex-1 sm:flex-none px-6 py-3 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
        @click="loadPreview"
      >
        <svg v-if="previewLoading" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
          />
        </svg>
        预览数据
      </button>
      <button
        :disabled="selectedTypes.size === 0 || downloading"
        class="flex-1 px-6 py-3 rounded-lg bg-primary-600 hover:bg-primary-700 text-white text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
        @click="downloadExport"
      >
        <svg v-if="downloading" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
          />
        </svg>
        下载{{ format === 'csv' ? 'CSV' : 'JSON' }}
      </button>
    </div>

    <!-- 预览区域 -->
    <div
      v-if="previewData || previewError"
      class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden"
    >
      <div v-if="previewError" class="p-6 text-center">
        <svg
          class="w-10 h-10 mx-auto text-red-400 dark:text-red-500 mb-2"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <p class="text-sm text-red-600 dark:text-red-400">{{ previewError }}</p>
      </div>

      <div v-else-if="previewData">
        <!-- 预览头部 -->
        <div
          class="px-4 sm:px-6 py-3 bg-gray-50 dark:bg-gray-700/50 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between"
        >
          <div class="flex items-center gap-2">
            <svg class="w-4 h-4 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ previewTypeLabel }}</span>
          </div>
          <span class="text-xs text-gray-500 dark:text-gray-400">
            预览 {{ previewData.data?.length || 0 }} 条（共 {{ previewData.total || 0 }} 条）
          </span>
        </div>

        <!-- 预览表格 -->
        <div class="overflow-x-auto max-h-96 custom-scroll">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 dark:bg-gray-700/30 sticky top-0">
              <tr>
                <th
                  v-for="col in previewColumns"
                  :key="col.key"
                  class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 whitespace-nowrap"
                >
                  {{ col.label }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
              <tr v-for="(row, idx) in previewData.data" :key="idx" class="hover:bg-gray-50 dark:hover:bg-gray-700/20">
                <td
                  v-for="col in previewColumns"
                  :key="col.key"
                  class="px-3 py-2 text-gray-700 dark:text-gray-300 max-w-xs truncate"
                  :title="String(row[col.key] ?? '')"
                >
                  <span v-if="col.key === 'is_bookmarked'">
                    {{ row[col.key] ? '是' : '否' }}
                  </span>
                  <span
                    v-else-if="col.key === 'difficulty'"
                    class="inline-block px-1.5 py-0.5 rounded text-xs"
                    :class="diffClass(row[col.key])"
                  >
                    {{ diffLabel(row[col.key]) }}
                  </span>
                  <span v-else-if="col.key === 'type'" class="text-xs">
                    {{ typeLabel(row[col.key]) }}
                  </span>
                  <span v-else>{{ formatCell(row[col.key]) }}</span>
                </td>
              </tr>
            </tbody>
          </table>

          <div
            v-if="!previewData.data || previewData.data.length === 0"
            class="p-8 text-center text-gray-400 dark:text-gray-500"
          >
            <p>没有符合条件的数据</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="selectedTypes.size === 0" class="text-center py-12">
      <svg
        class="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-3"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10"
        />
      </svg>
      <p class="text-gray-500 dark:text-gray-400">请至少选择一种数据类型</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { exportDataPreview, exportDataDownload } from '../api/client'
import { useToast } from '../composables/useToast'

const toast = useToast()

// ===== 数据类型 =====
const dataTypes = [
  {
    value: 'cards',
    label: '知识卡片',
    desc: '概念、详情、标签',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"/></svg>',
    color: 'text-emerald-500'
  },
  {
    value: 'quizzes',
    label: '练习题',
    desc: '题目、答案、解析',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z"/><path fill-rule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3z" clip-rule="evenodd"/></svg>',
    color: 'text-blue-500'
  },
  {
    value: 'chats',
    label: '对话记录',
    desc: '会话、消息内容',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M18 10c0 3.866-3.582 7-8 7a8.841 8.841 0 01-4.083-.98L2 17l1.338-3.123C2.493 12.767 2 11.434 2 10c0-3.866 3.582-7 8-7s8 3.134 8 7zM7 9H5v2h2V9zm8 0h-2v2h2V9zm-4 0H9v2h2V9z" clip-rule="evenodd"/></svg>',
    color: 'text-violet-500'
  },
  {
    value: 'mistakes',
    label: '错题记录',
    desc: '错题、正确答案',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M4 2a2 2 0 00-2 2v11a3 3 0 003 3h10a3 3 0 003-3V4a2 2 0 00-2-2H4zm1 3a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1zm0 4a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1z" clip-rule="evenodd"/></svg>',
    color: 'text-red-500'
  }
]

const formats = [
  {
    value: 'csv',
    label: 'CSV',
    desc: 'Excel 兼容，适合表格分析',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M6 2a2 2 0 00-2 2v12a2 2 0 002 2h8a2 2 0 002-2V7.414A2 2 0 0015.414 6L12 2.586A2 2 0 0010.586 2H6zm2 10a1 1 0 10-2 0v3a1 1 0 102 0v-3zm2-3a1 1 0 011 1v5a1 1 0 11-2 0v-5a1 1 0 011-1zm4-1a1 1 0 10-2 0v6a1 1 0 102 0V8z" clip-rule="evenodd"/></svg>'
  },
  {
    value: 'json',
    label: 'JSON',
    desc: '程序友好，适合二次处理',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M12.316 3.051a1 1 0 01.633 1.265l-4 12a1 1 0 11-1.898-.632l4-12a1 1 0 011.265-.633zM5.707 6.293a1 1 0 010 1.414L3.414 10l2.293 2.293a1 1 0 11-1.414 1.414l-3-3a1 1 0 010-1.414l3-3a1 1 0 011.414 0zm8.586 0a1 1 0 011.414 0l3 3a1 1 0 010 1.414l-3 3a1 1 0 11-1.414-1.414L16.586 10l-2.293-2.293a1 1 0 010-1.414z" clip-rule="evenodd"/></svg>'
  }
]

const datePresets = [
  { label: '全部时间', days: 0 },
  { label: '最近 7 天', days: 7 },
  { label: '最近 30 天', days: 30 },
  { label: '最近 90 天', days: 90 },
  { label: '今年', days: -1 }
]

// ===== 状态 =====
const selectedTypes = ref(new Set(['cards', 'quizzes', 'chats', 'mistakes']))
const format = ref('csv')
const dateFrom = ref('')
const dateTo = ref('')
const activePreset = ref('全部时间')
const previewLoading = ref(false)
const previewData = ref(null)
const previewError = ref(null)
const previewType = ref('')
const downloading = ref(false)

// ===== 表格列定义 =====
const columnDefs = {
  cards: [
    { key: 'concept', label: '概念' },
    { key: 'detail', label: '详情' },
    { key: 'tags', label: '标签' },
    { key: 'difficulty', label: '难度' },
    { key: 'review_count', label: '复习次数' },
    { key: 'ease_factor', label: '难度系数' },
    { key: 'is_bookmarked', label: '书签' },
    { key: 'material_title', label: '所属材料' },
    { key: 'created_at', label: '创建时间' }
  ],
  quizzes: [
    { key: 'question', label: '题目' },
    { key: 'type', label: '题型' },
    { key: 'difficulty', label: '难度' },
    { key: 'answer', label: '答案' },
    { key: 'explanation', label: '解析' },
    { key: 'material_title', label: '所属材料' },
    { key: 'created_at', label: '创建时间' }
  ],
  chats: [
    { key: 'conversation_title', label: '会话标题' },
    { key: 'role', label: '角色' },
    { key: 'content', label: '消息内容' },
    { key: 'created_at', label: '时间' }
  ],
  mistakes: [
    { key: 'question', label: '题目' },
    { key: 'type', label: '题型' },
    { key: 'difficulty', label: '难度' },
    { key: 'user_answer', label: '我的答案' },
    { key: 'correct_answer', label: '正确答案' },
    { key: 'mistake_at', label: '错误时间' }
  ]
}

const previewTypeLabel = computed(() => {
  const item = dataTypes.find((d) => d.value === previewType.value)
  return item ? item.label : ''
})

const previewColumns = computed(() => {
  return columnDefs[previewType.value] || []
})

// ===== 方法 =====
function toggleType(value) {
  const s = new Set(selectedTypes.value)
  if (s.has(value)) {
    s.delete(value)
  } else {
    s.add(value)
  }
  selectedTypes.value = s
}

function selectAll() {
  selectedTypes.value = new Set(dataTypes.map((d) => d.value))
}

function clearAll() {
  selectedTypes.value = new Set()
}

function applyDatePreset(preset) {
  activePreset.value = preset.label
  if (preset.days === 0) {
    dateFrom.value = ''
    dateTo.value = ''
  } else if (preset.days === -1) {
    // 今年
    const now = new Date()
    dateFrom.value = `${now.getFullYear()}-01-01`
    dateTo.value = ''
  } else {
    const now = new Date()
    const from = new Date(now.getTime() - preset.days * 24 * 60 * 60 * 1000)
    dateFrom.value = from.toISOString().split('T')[0]
    dateTo.value = now.toISOString().split('T')[0]
  }
}

async function loadPreview() {
  if (selectedTypes.value.size === 0) return

  // 预览取第一个选中的类型
  const firstType = Array.from(selectedTypes.value)[0]
  previewType.value = firstType
  previewLoading.value = true
  previewError.value = null
  previewData.value = null

  try {
    const params = { type: firstType }
    if (dateFrom.value) params.date_from = dateFrom.value
    if (dateTo.value) params.date_to = dateTo.value
    const res = await exportDataPreview(params)
    previewData.value = res.data
  } catch (err) {
    previewError.value = '加载预览失败，请重试'
  } finally {
    previewLoading.value = false
  }
}

async function downloadExport() {
  if (selectedTypes.value.size === 0) return

  downloading.value = true
  try {
    const params = {
      type: Array.from(selectedTypes.value).join(','),
      format: format.value
    }
    if (dateFrom.value) params.date_from = dateFrom.value
    if (dateTo.value) params.date_to = dateTo.value

    const blob = await exportDataDownload(params)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url

    // 从 Content-Disposition 获取文件名，fallback 用默认名
    const typeStr = params.type === 'cards,quizzes,chats,mistakes' ? 'all' : params.type.replace(/,/g, '-')
    const ext = format.value
    a.download = `studyforge-${typeStr}-${new Date().toISOString().slice(0, 10)}.${ext}`

    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)

    toast.success('导出成功，文件已开始下载')
  } catch (err) {
    toast.error('导出失败，请重试')
  } finally {
    downloading.value = false
  }
}

function diffClass(difficulty) {
  switch (difficulty) {
    case 'easy':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
    case 'medium':
      return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    case 'hard':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
  }
}

function diffLabel(d) {
  switch (d) {
    case 'easy':
      return '简单'
    case 'medium':
      return '中等'
    case 'hard':
      return '困难'
    default:
      return d || '-'
  }
}

function typeLabel(t) {
  switch (t) {
    case 'choice':
      return '选择题'
    case 'fill':
      return '填空题'
    case 'judge':
      return '判断题'
    case 'short_answer':
      return '简答题'
    case 'user':
      return '用户'
    case 'assistant':
      return 'AI'
    default:
      return t || '-'
  }
}

function formatCell(val) {
  if (val === null || val === undefined || val === '') return '-'
  const str = String(val)
  return str.length > 60 ? str.slice(0, 57) + '...' : str
}
</script>

<style scoped>
.custom-scroll::-webkit-scrollbar {
  width: 4px;
  height: 4px;
}
.custom-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroll::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 4px;
}
:deep(.dark) .custom-scroll::-webkit-scrollbar-thumb {
  background: #4b5563;
}
</style>
