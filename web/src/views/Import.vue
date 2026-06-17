<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-5xl mx-auto">
    <!-- 页面标题 -->
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Anki 牌组导入</h1>
      <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">上传 .apkg 文件，将 Anki 卡片导入到 StudyForge</p>
    </div>

    <!-- 步骤指示器 -->
    <div class="flex items-center gap-1 mb-6 text-sm flex-wrap">
      <template v-for="(step, i) in steps" :key="i">
        <span
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-full transition-colors"
          :class="
            currentStep >= i
              ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
              : 'bg-gray-100 text-gray-400 dark:bg-gray-800 dark:text-gray-500'
          "
        >
          <span
            class="w-5 h-5 rounded-full flex items-center justify-center text-xs font-bold"
            :class="
              currentStep >= i
                ? 'bg-primary-500 text-white'
                : 'bg-gray-300 dark:bg-gray-600 text-gray-500 dark:text-gray-400'
            "
          >
            {{ i + 1 }}
          </span>
          {{ step }}
        </span>
        <svg
          v-if="i < steps.length - 1"
          class="w-4 h-4 text-gray-300 dark:text-gray-600 flex-shrink-0"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </template>
    </div>

    <!-- 步骤1：上传文件 -->
    <div v-if="currentStep === 0">
      <div
        class="border-2 border-dashed rounded-xl p-8 sm:p-12 text-center transition-all duration-200 cursor-pointer"
        :class="[
          isDragging
            ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
            : 'border-gray-300 dark:border-gray-600 hover:border-primary-400 dark:hover:border-primary-500'
        ]"
        @dragover.prevent="isDragging = true"
        @dragleave.prevent="isDragging = false"
        @drop.prevent="handleDrop"
        @click="fileInput?.click()"
      >
        <input ref="fileInput" type="file" accept=".apkg" class="hidden" @change="handleFileSelect" />
        <svg
          class="w-16 h-16 mx-auto mb-4 text-gray-400 dark:text-gray-500"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
          />
        </svg>
        <p class="text-lg font-medium text-gray-700 dark:text-gray-300">
          拖拽 .apkg 文件到此处，或
          <span class="text-primary-500 hover:text-primary-600">点击选择</span>
        </p>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-2">支持 Anki 导出的 .apkg 文件，最大 50MB</p>
      </div>

      <!-- 错误提示 -->
      <div
        v-if="errorMsg"
        class="mt-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-700 dark:text-red-400 text-sm"
      >
        {{ errorMsg }}
      </div>

      <!-- 解析进度 -->
      <div v-if="parsing" class="mt-6">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-5 h-5 border-2 border-primary-500 border-t-transparent rounded-full animate-spin"></div>
          <span class="text-sm text-gray-700 dark:text-gray-300">正在解析 {{ fileName }}...</span>
        </div>
        <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
          <div class="bg-primary-500 h-2 rounded-full animate-pulse" style="width: 60%"></div>
        </div>
      </div>
    </div>

    <!-- 步骤2：预览与选择 -->
    <div v-if="currentStep === 1 && previewData">
      <!-- 文件信息卡片 -->
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 sm:p-6 mb-6">
        <div class="flex flex-wrap items-center gap-4">
          <div class="w-12 h-12 bg-blue-100 dark:bg-blue-900/30 rounded-xl flex items-center justify-center">
            <svg class="w-6 h-6 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="font-semibold text-gray-900 dark:text-gray-100 truncate">{{ previewData.filename }}</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              模型: {{ previewData.model_name }} · 字段: {{ (previewData.field_names || []).join(', ') }}
            </p>
          </div>
          <div class="text-right">
            <p class="text-2xl font-bold text-primary-600 dark:text-primary-400">{{ previewData.total }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">张卡片</p>
          </div>
        </div>
      </div>

      <!-- 操作栏 -->
      <div class="flex flex-wrap items-center gap-3 mb-4">
        <button
          class="px-4 py-2 text-sm font-medium rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
          @click="selectAll"
        >
          全选 ({{ previewData.cards.length }})
        </button>
        <button
          class="px-4 py-2 text-sm font-medium rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
          @click="selectNone"
        >
          取消全选
        </button>
        <button
          class="px-4 py-2 text-sm font-medium rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
          @click="invertSelection"
        >
          反选
        </button>
        <span class="text-sm text-gray-500 dark:text-gray-400 ml-auto">
          已选
          <span class="font-semibold text-primary-600 dark:text-primary-400">{{ selectedCount }}</span>
          / {{ previewData.cards.length }}
        </span>
      </div>

      <!-- 搜索过滤 -->
      <div class="mb-4">
        <input
          v-model="filterQuery"
          type="text"
          placeholder="搜索卡片内容..."
          class="w-full sm:w-80 px-4 py-2 text-sm rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-colors"
        />
      </div>

      <!-- 卡片列表 -->
      <div class="border border-gray-200 dark:border-gray-700 rounded-xl overflow-hidden">
        <div class="max-h-[500px] overflow-y-auto custom-scroll">
          <div
            v-for="card in filteredCards"
            :key="card.index"
            class="flex items-start gap-3 p-3 sm:p-4 border-b border-gray-100 dark:border-gray-700/50 last:border-b-0 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
          >
            <input
              type="checkbox"
              :checked="selectedSet.has(card.index)"
              class="mt-1 w-4 h-4 rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500"
              @change="toggleCard(card.index)"
            />
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-gray-900 dark:text-gray-100 line-clamp-2">
                {{ card.concept || '(空)' }}
              </p>
              <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 line-clamp-2">
                {{ card.detail || '(无背面内容)' }}
              </p>
              <div v-if="card.tags" class="flex flex-wrap gap-1 mt-1.5">
                <span
                  v-for="tag in card.tags.split(', ')"
                  :key="tag"
                  class="px-1.5 py-0.5 text-xs rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400"
                >
                  {{ tag }}
                </span>
              </div>
            </div>
            <span class="text-xs text-gray-400 dark:text-gray-500 flex-shrink-0">#{{ card.index + 1 }}</span>
          </div>
          <div v-if="filteredCards.length === 0" class="p-8 text-center text-gray-400 dark:text-gray-500 text-sm">
            没有匹配的卡片
          </div>
        </div>
      </div>

      <!-- 底部操作 -->
      <div class="flex items-center gap-3 mt-6">
        <button
          class="px-4 py-2 text-sm font-medium rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
          @click="resetToUpload"
        >
          返回重新选择文件
        </button>
        <button
          :disabled="selectedCount === 0 || importing"
          class="px-6 py-2 text-sm font-medium rounded-lg bg-primary-600 text-white hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
          @click="confirmImport"
        >
          <div
            v-if="importing"
            class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"
          ></div>
          {{ importing ? '导入中...' : `导入 ${selectedCount} 张卡片` }}
        </button>
      </div>
    </div>

    <!-- 步骤3：导入结果 -->
    <div v-if="currentStep === 2" class="text-center py-12">
      <div v-if="importResult" class="max-w-md mx-auto">
        <!-- 成功图标 -->
        <div
          class="w-20 h-20 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center mx-auto mb-6"
        >
          <svg
            class="w-10 h-10 text-green-600 dark:text-green-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <h2 class="text-xl font-bold text-gray-900 dark:text-gray-100 mb-2">导入成功！</h2>
        <p class="text-gray-500 dark:text-gray-400 mb-6">
          成功导入
          <span class="font-bold text-primary-600 dark:text-primary-400">{{ importResult.imported }}</span>
          张卡片到 StudyForge
          <span v-if="importResult.total !== importResult.imported" class="text-amber-500">
            （{{ importResult.total - importResult.imported }} 张跳过）
          </span>
        </p>
        <div class="flex items-center justify-center gap-3">
          <router-link
            to="/cards"
            class="px-6 py-2.5 text-sm font-medium rounded-lg bg-primary-600 text-white hover:bg-primary-700 transition-colors"
          >
            查看卡片
          </router-link>
          <button
            class="px-6 py-2.5 text-sm font-medium rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
            @click="resetToUpload"
          >
            继续导入
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { importAnkiPreview, importAnkiConfirm } from '../api/client'
import { useToast } from '../composables/useToast'

const _router = useRouter()
const toast = useToast()

const steps = ['上传文件', '选择卡片', '导入完成']
const currentStep = ref(0)
const isDragging = ref(false)
const parsing = ref(false)
const importing = ref(false)
const errorMsg = ref('')
const fileName = ref('')
const previewData = ref(null)
const importResult = ref(null)
const filterQuery = ref('')
const selectedSet = ref(new Set())
const fileInput = ref(null)

// 过滤后的卡片
const filteredCards = computed(() => {
  if (!previewData.value) return []
  const q = filterQuery.value.toLowerCase().trim()
  if (!q) return previewData.value.cards
  return previewData.value.cards.filter(
    (c) =>
      (c.concept || '').toLowerCase().includes(q) ||
      (c.detail || '').toLowerCase().includes(q) ||
      (c.tags || '').toLowerCase().includes(q)
  )
})

// 已选数量
const selectedCount = computed(() => selectedSet.value.size)

// 文件选择处理
function handleFileSelect(e) {
  const file = e.target.files?.[0]
  if (file) processFile(file)
}

function handleDrop(e) {
  isDragging.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file) processFile(file)
}

async function processFile(file) {
  errorMsg.value = ''

  // 客户端验证
  if (!file.name.toLowerCase().endsWith('.apkg')) {
    errorMsg.value = '仅支持 .apkg 格式的 Anki 牌组文件'
    return
  }
  if (file.size > 50 * 1024 * 1024) {
    errorMsg.value = '文件过大，最大支持 50MB'
    return
  }
  if (file.size === 0) {
    errorMsg.value = '文件为空'
    return
  }

  fileName.value = file.name
  parsing.value = true
  currentStep.value = 0

  try {
    const res = await importAnkiPreview(file)
    previewData.value = res.data
    // 默认全选
    selectedSet.value = new Set(res.data.cards.map((c) => c.index))
    currentStep.value = 1
  } catch (err) {
    errorMsg.value = err.response?.data?.error || '解析文件失败，请确认文件格式正确'
  } finally {
    parsing.value = false
  }
}

// 选择操作
function selectAll() {
  if (!previewData.value) return
  selectedSet.value = new Set(previewData.value.cards.map((c) => c.index))
}

function selectNone() {
  selectedSet.value = new Set()
}

function invertSelection() {
  if (!previewData.value) return
  const newSet = new Set()
  for (const card of previewData.value.cards) {
    if (!selectedSet.value.has(card.index)) {
      newSet.add(card.index)
    }
  }
  selectedSet.value = newSet
}

function toggleCard(index) {
  const newSet = new Set(selectedSet.value)
  if (newSet.has(index)) {
    newSet.delete(index)
  } else {
    newSet.add(index)
  }
  selectedSet.value = newSet
}

// 确认导入
async function confirmImport() {
  if (!previewData.value || selectedCount.value === 0) return

  importing.value = true
  try {
    const selectedCards = previewData.value.cards.filter((c) => selectedSet.value.has(c.index))
    const res = await importAnkiConfirm({ cards: selectedCards })
    importResult.value = res.data
    currentStep.value = 2
    toast.success(`成功导入 ${res.data.imported} 张卡片！`)
  } catch (err) {
    toast.error(err.response?.data?.error || '导入失败')
  } finally {
    importing.value = false
  }
}

// 重置
function resetToUpload() {
  currentStep.value = 0
  previewData.value = null
  importResult.value = null
  errorMsg.value = ''
  fileName.value = ''
  filterQuery.value = ''
  selectedSet.value = new Set()
  if (fileInput.value) fileInput.value.value = ''
}
</script>

<style scoped>
:deep(.custom-scroll)::-webkit-scrollbar {
  width: 6px;
}
:deep(.custom-scroll)::-webkit-scrollbar-track {
  background: transparent;
}
:deep(.custom-scroll)::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.4);
  border-radius: 4px;
}
:deep(.custom-scroll)::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.6);
}
</style>
