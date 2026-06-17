<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-6xl mx-auto">
    <!-- 顶部导航 -->
    <div class="flex items-start gap-3 mb-6">
      <button
        class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition text-gray-500 dark:text-gray-400 shrink-0"
        @click="$router.back()"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
      <div class="flex-1 min-w-0">
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-gray-100">
          {{ material?.title || '加载中...' }}
        </h1>
        <div class="flex items-center gap-2 mt-1 flex-wrap">
          <span class="text-xs px-2 py-0.5 rounded-full" :class="statusClass">{{ statusLabel }}</span>
          <span class="text-xs text-gray-400 dark:text-gray-500">{{ formatDate(material?.created_at) }}</span>
          <span v-if="material?.content_type" class="text-xs text-gray-400 dark:text-gray-500">· {{ typeLabel }}</span>
        </div>
      </div>
      <div class="flex items-center gap-2 shrink-0 flex-wrap justify-end">
        <router-link
          v-if="material?.graph_data"
          :to="`/graph/${material.id}`"
          class="px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-600 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 transition flex items-center gap-1.5"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101"
            />
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M10.172 13.828a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.102 1.101"
            />
          </svg>
          知识图谱
        </router-link>
        <button
          :disabled="material?.status === 'analyzing'"
          class="px-3 py-2 rounded-lg text-sm font-medium transition disabled:opacity-50 disabled:cursor-not-allowed"
          :class="
            material?.status === 'analyzing'
              ? 'bg-gray-100 dark:bg-gray-700 text-gray-400 dark:text-gray-500'
              : 'bg-primary-600 text-white hover:bg-primary-700'
          "
          @click="triggerAnalyze"
        >
          {{ material?.status === 'analyzing' ? '分析中...' : '重新分析' }}
        </button>
      </div>
    </div>

    <!-- 加载骨架 -->
    <div v-if="loading" class="space-y-4">
      <div class="bg-white dark:bg-gray-800 rounded-xl p-6 animate-pulse">
        <div class="h-4 bg-gray-200 rounded w-1/3 mb-4"></div>
        <div class="h-3 bg-gray-100 rounded w-full mb-2"></div>
        <div class="h-3 bg-gray-100 rounded w-5/6 mb-2"></div>
        <div class="h-3 bg-gray-100 rounded w-4/6"></div>
      </div>
    </div>

    <!-- 错误状态 -->
    <div
      v-else-if="error"
      class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-red-100 dark:border-red-800 p-12 text-center"
    >
      <p class="text-4xl mb-4">😵</p>
      <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">{{ error }}</h3>
      <button class="text-primary-600 hover:underline text-sm" @click="loadDetail">重试</button>
    </div>

    <template v-else-if="material">
      <!-- 分析摘要 -->
      <div
        v-if="analysis"
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-6 mb-6"
      >
        <div class="flex items-center gap-2 mb-4">
          <span
            class="w-8 h-8 rounded-lg bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
          </span>
          <h2 class="text-lg font-bold text-gray-900 dark:text-gray-100">Analyst 分析摘要</h2>
        </div>

        <!-- 摘要 -->
        <div v-if="analysis.summary" class="mb-5">
          <p class="text-gray-700 dark:text-gray-300 leading-relaxed" v-html="renderMath(analysis.summary)"></p>
        </div>

        <!-- 知识点 -->
        <div v-if="analysis.key_points?.length" class="mb-5">
          <h3 class="text-sm font-semibold text-gray-600 dark:text-gray-400 mb-3 flex items-center gap-1.5">
            <svg class="w-4 h-4 text-amber-500" fill="currentColor" viewBox="0 0 20 20">
              <path
                d="M11 3a1 1 0 10-2 0v1a1 1 0 102 0V3zM15.657 5.757a1 1 0 00-1.414-1.414l-.707.707a1 1 0 001.414 1.414l.707-.707zM18 10a1 1 0 01-1 1h-1a1 1 0 110-2h1a1 1 0 011 1zM5.05 6.464A1 1 0 106.464 5.05l-.707-.707a1 1 0 00-1.414 1.414l.707.707zM4 11a1 1 0 100-2H3a1 1 0 000 2h1zM10 18a1 1 0 001-1v-1a1 1 0 10-2 0v1a1 1 0 001 1z"
              />
              <path
                fill-rule="evenodd"
                d="M10 2a8 8 0 100 16 8 8 0 000-16zm0 14a6 6 0 110-12 6 6 0 010 12z"
                clip-rule="evenodd"
              />
            </svg>
            核心知识点（{{ analysis.key_points.length }}）
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
            <div
              v-for="(point, idx) in analysis.key_points"
              :key="idx"
              class="flex items-start gap-2 p-3 rounded-lg bg-gray-50 dark:bg-gray-700 border border-gray-100 dark:border-gray-700"
            >
              <span
                class="w-5 h-5 rounded-full bg-primary-100 text-primary-700 text-xs flex items-center justify-center font-medium mt-0.5 shrink-0"
              >
                {{ idx + 1 }}
              </span>
              <span
                class="text-sm text-gray-700 dark:text-gray-300"
                v-html="
                  renderMath(typeof point === 'string' ? point : point.title || point.name || JSON.stringify(point))
                "
              ></span>
            </div>
          </div>
        </div>

        <!-- 关系 -->
        <div v-if="analysis.relationships?.length">
          <h3 class="text-sm font-semibold text-gray-600 dark:text-gray-400 mb-3 flex items-center gap-1.5">
            <svg class="w-4 h-4 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101"
              />
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M10.172 13.828a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.102 1.101"
              />
            </svg>
            概念关系（{{ analysis.relationships.length }}）
          </h3>
          <div class="space-y-2">
            <div
              v-for="(rel, idx) in analysis.relationships"
              :key="idx"
              class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 p-2 rounded bg-purple-50 dark:bg-purple-900/20 border border-purple-100 dark:border-purple-800"
            >
              <span class="font-medium text-purple-700">{{ rel.source || rel.from }}</span>
              <svg class="w-4 h-4 text-purple-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
              </svg>
              <span class="font-medium text-purple-700">{{ rel.target || rel.to }}</span>
              <span v-if="rel.label || rel.relation" class="text-xs text-purple-500 ml-1">
                ({{ rel.label || rel.relation }})
              </span>
            </div>
          </div>
        </div>

        <!-- 重要性 -->
        <div
          v-if="analysis.importance"
          class="mt-5 p-3 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-100 dark:border-emerald-800"
        >
          <h3 class="text-sm font-semibold text-emerald-700 mb-1">重要性评估</h3>
          <p class="text-sm text-emerald-600" v-html="renderMath(analysis.importance)"></p>
        </div>
      </div>

      <!-- 无分析数据 -->
      <div
        v-else-if="material.status === 'completed'"
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-8 mb-6 text-center"
      >
        <p class="text-4xl mb-3">📊</p>
        <p class="text-gray-500 dark:text-gray-400">分析数据暂不可用</p>
      </div>

      <!-- 关联卡片 -->
      <div v-if="material.cards?.length" class="mb-6">
        <h2 class="text-lg font-bold text-gray-900 dark:text-gray-100 mb-4 flex items-center gap-2">
          <span
            class="w-8 h-8 rounded-lg bg-green-50 dark:bg-green-900/30 flex items-center justify-center text-green-600 dark:text-green-400"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01"
              />
            </svg>
          </span>
          知识卡片（{{ material.cards.length }}）
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div
            v-for="card in material.cards"
            :key="card.id"
            class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-5 hover:shadow-md transition"
          >
            <div class="flex items-start justify-between mb-2">
              <span class="px-2 py-0.5 rounded text-xs font-medium" :class="diffClass(card.difficulty)">
                {{ diffLabel(card.difficulty) }}
              </span>
              <span v-if="card.review_count > 0" class="text-xs text-emerald-600">复习 {{ card.review_count }}次</span>
            </div>
            <h4 class="font-bold text-gray-900 dark:text-gray-100 mb-1" v-html="renderMath(card.concept)"></h4>
            <p class="text-sm text-gray-500 dark:text-gray-400 line-clamp-2 mb-2" v-html="renderMath(card.detail)"></p>
            <div class="flex flex-wrap gap-1">
              <span
                v-for="tag in parseTags(card.tags)"
                :key="tag"
                class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 rounded text-xs"
              >
                {{ tag }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 关联练习题 -->
      <div v-if="material.quizzes?.length">
        <h2 class="text-lg font-bold text-gray-900 dark:text-gray-100 mb-4 flex items-center gap-2">
          <span
            class="w-8 h-8 rounded-lg bg-orange-50 dark:bg-orange-900/30 flex items-center justify-center text-orange-600 dark:text-orange-400"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"
              />
            </svg>
          </span>
          练习题（{{ material.quizzes.length }}）
        </h2>
        <div class="space-y-3">
          <div
            v-for="(quiz, idx) in material.quizzes"
            :key="quiz.id"
            class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-5"
          >
            <div class="flex items-start gap-3">
              <span
                class="w-7 h-7 rounded-full bg-orange-50 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400 text-sm flex items-center justify-center font-medium shrink-0"
              >
                {{ idx + 1 }}
              </span>
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-1">
                  <span
                    class="text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400"
                  >
                    {{ quizTypeLabel(quiz.question_type) }}
                  </span>
                  <span class="text-xs px-1.5 py-0.5 rounded" :class="diffClass(quiz.difficulty)">
                    {{ diffLabel(quiz.difficulty) }}
                  </span>
                </div>
                <p class="text-sm text-gray-800 dark:text-gray-200" v-html="renderMath(quiz.question)"></p>
                <p v-if="quiz.explanation" class="text-xs text-gray-400 dark:text-gray-500 mt-2">
                  解析:
                  <span v-html="renderMath(quiz.explanation)"></span>
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getMaterial, analyzeMaterial } from '../api/client'
import { useToast } from '../composables/useToast'
import { renderMath } from '../composables/useMathRender'

const route = useRoute()
const _router = useRouter()
const { success: toastSuccess, error: toastError } = useToast()

const material = ref(null)
const loading = ref(true)
const error = ref('')

// 解析分析数据
const analysis = computed(() => {
  if (!material.value?.analysis_data) return null
  try {
    const data = JSON.parse(material.value.analysis_data)
    return data
  } catch {
    return null
  }
})

const statusClass = computed(() => {
  const s = material.value?.status
  return (
    {
      pending: 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400',
      analyzing: 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400',
      completed: 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400',
      failed: 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400',
      partial: 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400'
    }[s] || 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
  )
})

const statusLabel = computed(() => {
  return (
    { pending: '待分析', analyzing: '分析中', completed: '已完成', failed: '失败', partial: '部分完成' }[
      material.value?.status
    ] ||
    material.value?.status ||
    ''
  )
})

const typeLabel = computed(() => {
  return { text: '文本', pdf: 'PDF', url: '网页' }[material.value?.content_type] || material.value?.content_type || ''
})

function formatDate(d) {
  if (!d) return ''
  const date = new Date(d)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function diffClass(d) {
  return (
    {
      easy: 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400',
      medium: 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400',
      hard: 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400'
    }[d] || 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
  )
}

function diffLabel(d) {
  return { easy: '简单', medium: '中等', hard: '困难' }[d] || '未分级'
}

function quizTypeLabel(t) {
  return { choice: '选择题', fill: '填空题', short_answer: '简答题' }[t] || t || '未知'
}

function parseTags(tags) {
  if (Array.isArray(tags)) return tags
  if (typeof tags === 'string' && tags) return tags.split(',').map((t) => t.trim())
  return []
}

async function loadDetail() {
  loading.value = true
  error.value = ''
  try {
    const res = await getMaterial(route.params.id)
    material.value = res.data
  } catch (e) {
    error.value = '材料加载失败'
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function triggerAnalyze() {
  if (!material.value) return
  try {
    await analyzeMaterial(material.value.id)
    material.value.status = 'analyzing'
    toastSuccess('分析任务已启动，请等待完成')
  } catch (e) {
    toastError(e?.response?.data?.error || '触发分析失败')
  }
}

onMounted(loadDetail)
</script>

<style scoped>
/* KaTeX 暗色模式适配 */
:deep(.katex) {
  color: inherit;
}

/* 卡片内嵌图片样式 */
:deep(.card-image) {
  max-width: 100%;
  height: auto;
  border-radius: 0.5rem;
  margin: 0.5rem 0;
  border: 1px solid #e5e7eb;
}
:deep(.dark .card-image) {
  border-color: #4b5563;
  box-shadow: 0 0 0 1px rgba(75, 85, 99, 0.3);
}
</style>
