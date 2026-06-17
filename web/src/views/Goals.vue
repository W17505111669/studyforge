<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-5xl mx-auto">
    <!-- 页头 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">学习目标</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">设定周/月目标，追踪学习进度</p>
      </div>
      <button
        class="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium bg-primary-600 text-white hover:bg-primary-700 transition-colors shadow-sm dark:bg-primary-500 dark:hover:bg-primary-600"
        @click="showCreateForm = !showCreateForm"
      >
        <svg
          class="w-4 h-4"
          :class="showCreateForm ? 'rotate-45' : ''"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
        </svg>
        {{ showCreateForm ? '取消' : '新建目标' }}
      </button>
    </div>

    <!-- ========== 创建目标表单 ========== -->
    <Transition name="slide-down">
      <div
        v-if="showCreateForm"
        class="mb-6 bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 shadow-sm"
      >
        <h2 class="text-base font-semibold text-gray-800 dark:text-gray-200 mb-4">创建新目标</h2>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <!-- 目标类型 -->
          <div>
            <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1.5">目标类型</label>
            <select
              v-model="form.type"
              class="w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm text-gray-800 dark:text-gray-200 px-3 py-2.5 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition"
            >
              <option value="review_cards">复习卡片</option>
              <option value="complete_quizzes">完成练习</option>
              <option value="study_minutes">学习时长(分钟)</option>
              <option value="upload_materials">上传材料</option>
            </select>
          </div>
          <!-- 目标值 -->
          <div>
            <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1.5">目标值</label>
            <input
              v-model.number="form.target_value"
              type="number"
              min="1"
              max="99999"
              class="w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm text-gray-800 dark:text-gray-200 px-3 py-2.5 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition"
              placeholder="如: 50"
            />
          </div>
          <!-- 周期 -->
          <div>
            <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1.5">周期</label>
            <div class="flex gap-2">
              <button
                v-for="p in ['weekly', 'monthly']"
                :key="p"
                class="flex-1 rounded-lg border text-sm py-2.5 font-medium transition-colors"
                :class="
                  form.period === p
                    ? 'border-primary-500 bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400 dark:border-primary-500'
                    : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700'
                "
                @click="form.period = p"
              >
                {{ p === 'weekly' ? '每周' : '每月' }}
              </button>
            </div>
          </div>
          <!-- 创建按钮 -->
          <div class="flex items-end">
            <button
              :disabled="!form.target_value || form.target_value < 1 || creating"
              class="w-full rounded-lg bg-primary-600 text-white text-sm font-medium py-2.5 hover:bg-primary-700 dark:bg-primary-500 dark:hover:bg-primary-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              @click="handleCreate"
            >
              {{ creating ? '创建中...' : '创建目标' }}
            </button>
          </div>
        </div>
        <!-- 推荐目标提示 -->
        <div class="mt-3 flex flex-wrap gap-2">
          <span class="text-xs text-gray-400 dark:text-gray-500">推荐:</span>
          <button
            v-for="preset in presets"
            :key="preset.label"
            class="text-xs px-2.5 py-1 rounded-full border border-gray-200 dark:border-gray-600 text-gray-500 dark:text-gray-400 hover:border-primary-400 hover:text-primary-600 dark:hover:border-primary-500 dark:hover:text-primary-400 transition-colors"
            @click="applyPreset(preset)"
          >
            {{ preset.label }}
          </button>
        </div>
      </div>
    </Transition>

    <!-- ========== 活跃目标 ========== -->
    <div v-if="loading" class="space-y-4">
      <div
        v-for="i in 3"
        :key="i"
        class="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 p-5 animate-pulse"
      >
        <div class="flex items-center justify-between mb-3">
          <div class="h-4 w-24 bg-gray-200 dark:bg-gray-700 rounded"></div>
          <div class="h-4 w-16 bg-gray-200 dark:bg-gray-700 rounded"></div>
        </div>
        <div class="h-3 w-full bg-gray-200 dark:bg-gray-700 rounded-full mb-2"></div>
        <div class="h-3 w-32 bg-gray-200 dark:bg-gray-700 rounded"></div>
      </div>
    </div>

    <template v-else>
      <!-- 活跃目标卡片 -->
      <div v-if="activeGoals.length > 0" class="mb-8">
        <h2 class="text-sm font-semibold text-gray-600 dark:text-gray-400 uppercase tracking-wider mb-3">
          进行中的目标
          <span class="ml-1 text-primary-600 dark:text-primary-400">({{ activeGoals.length }})</span>
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div
            v-for="goal in activeGoals"
            :key="goal.id"
            class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 shadow-sm hover:shadow-md transition-shadow group"
          >
            <!-- 头部：类型 + 剩余天数 -->
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2">
                <span class="text-lg">{{ typeIcon(goal.type) }}</span>
                <span class="text-sm font-semibold text-gray-800 dark:text-gray-200">
                  {{ goal.type_label || typeLabel(goal.type) }}
                </span>
              </div>
              <span
                class="text-xs px-2 py-0.5 rounded-full"
                :class="
                  goal.remaining_days <= 1
                    ? 'bg-red-50 text-red-600 dark:bg-red-900/30 dark:text-red-400'
                    : goal.remaining_days <= 3
                      ? 'bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400'
                      : 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
                "
              >
                剩余 {{ goal.remaining_days }} 天
              </span>
            </div>

            <!-- 进度条 -->
            <div class="relative h-3 rounded-full bg-gray-100 dark:bg-gray-700 mb-2 overflow-hidden">
              <div
                class="absolute inset-y-0 left-0 rounded-full transition-all duration-500 ease-out"
                :style="{ width: Math.min(100, goal.percent) + '%' }"
                :class="progressBarClass(goal.percent)"
              ></div>
            </div>

            <!-- 数值 -->
            <div class="flex items-center justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">
                <span class="font-semibold text-gray-800 dark:text-gray-200">{{ goal.current_value }}</span>
                / {{ goal.target_value }}
              </span>
              <span class="font-semibold" :class="percentColor(goal.percent)">{{ goal.percent.toFixed(1) }}%</span>
            </div>

            <!-- 周期标签 -->
            <div class="flex items-center justify-between mt-3 pt-3 border-t border-gray-100 dark:border-gray-700">
              <span class="text-xs text-gray-400 dark:text-gray-500">
                {{ goal.period === 'weekly' ? '每周' : '每月' }}目标 · {{ formatDate(goal.start_date) }} ~
                {{ formatDate(goal.end_date) }}
              </span>
              <button
                class="opacity-0 group-hover:opacity-100 text-xs text-gray-400 dark:text-gray-500 hover:text-red-500 dark:hover:text-red-400 transition-all"
                @click="handleDelete(goal.id)"
              >
                删除
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div
        v-else-if="historyGoals.length === 0"
        class="text-center py-16 bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700"
      >
        <div class="text-5xl mb-4">🎯</div>
        <h3 class="text-lg font-semibold text-gray-700 dark:text-gray-300 mb-2">还没有学习目标</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">设定一个目标，开始你的学习之旅吧</p>
        <button
          class="px-5 py-2.5 rounded-lg bg-primary-600 text-white text-sm font-medium hover:bg-primary-700 dark:bg-primary-500 dark:hover:bg-primary-600 transition-colors"
          @click="showCreateForm = true"
        >
          创建第一个目标
        </button>
      </div>

      <!-- ========== 历史目标 ========== -->
      <div v-if="historyGoals.length > 0">
        <h2 class="text-sm font-semibold text-gray-600 dark:text-gray-400 uppercase tracking-wider mb-3 mt-6">
          历史目标
          <span class="ml-1 text-gray-400 dark:text-gray-500">({{ historyGoals.length }})</span>
        </h2>
        <div class="space-y-2">
          <div
            v-for="goal in historyGoals"
            :key="goal.id"
            class="bg-white dark:bg-gray-800 rounded-lg border border-gray-100 dark:border-gray-700 p-4 flex items-center justify-between"
          >
            <div class="flex items-center gap-3">
              <span class="text-lg">{{ typeIcon(goal.type) }}</span>
              <div>
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ goal.type_label || typeLabel(goal.type) }}
                </span>
                <span class="text-xs text-gray-400 dark:text-gray-500 ml-2">
                  {{ goal.current_value }}/{{ goal.target_value }} · {{ goal.period === 'weekly' ? '每周' : '每月' }}
                </span>
              </div>
            </div>
            <span
              class="text-xs font-medium px-2.5 py-1 rounded-full"
              :class="
                goal.status === 'completed'
                  ? 'bg-green-50 text-green-600 dark:bg-green-900/30 dark:text-green-400'
                  : 'bg-red-50 text-red-500 dark:bg-red-900/30 dark:text-red-400'
              "
            >
              {{ goal.status === 'completed' ? '已完成' : '未完成' }}
            </span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { listGoals, getGoalProgress, createGoal, deleteGoal } from '../api/client'
import { useToast } from '../composables/useToast'

const { success, error: toastError } = useToast()

// ===== 状态 =====
const loading = ref(true)
const creating = ref(false)
const showCreateForm = ref(false)

const activeGoals = ref([])
const historyGoals = ref([])

const form = ref({
  type: 'review_cards',
  target_value: 20,
  period: 'weekly'
})

// 推荐预设
const presets = [
  { label: '每周复习 20 张卡片', type: 'review_cards', value: 20, period: 'weekly' },
  { label: '每周完成 30 道题', type: 'complete_quizzes', value: 30, period: 'weekly' },
  { label: '每月学习 600 分钟', type: 'study_minutes', value: 600, period: 'monthly' },
  { label: '每月上传 5 份材料', type: 'upload_materials', value: 5, period: 'monthly' }
]

// ===== 工具函数 =====
function typeIcon(type) {
  const icons = {
    review_cards: '🃏',
    complete_quizzes: '✏️',
    study_minutes: '⏱️',
    upload_materials: '📚'
  }
  return icons[type] || '🎯'
}

function typeLabel(type) {
  const labels = {
    review_cards: '复习卡片',
    complete_quizzes: '完成练习',
    study_minutes: '学习时长(分钟)',
    upload_materials: '上传材料'
  }
  return labels[type] || type
}

function progressBarClass(percent) {
  if (percent >= 80) return 'bg-gradient-to-r from-green-400 to-green-500 dark:from-green-500 dark:to-green-400'
  if (percent >= 50) return 'bg-gradient-to-r from-yellow-400 to-green-400 dark:from-yellow-500 dark:to-green-500'
  if (percent >= 25) return 'bg-gradient-to-r from-amber-400 to-yellow-400 dark:from-amber-500 dark:to-yellow-500'
  return 'bg-gradient-to-r from-red-400 to-amber-400 dark:from-red-500 dark:to-amber-500'
}

function percentColor(percent) {
  if (percent >= 80) return 'text-green-600 dark:text-green-400'
  if (percent >= 50) return 'text-yellow-600 dark:text-yellow-400'
  if (percent >= 25) return 'text-amber-600 dark:text-amber-400'
  return 'text-red-500 dark:text-red-400'
}

function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

// ===== 应用推荐预设 =====
function applyPreset(preset) {
  form.value.type = preset.type
  form.value.target_value = preset.value
  form.value.period = preset.period
}

// ===== 数据加载 =====
async function loadData() {
  loading.value = true
  try {
    const [progressRes, historyRes] = await Promise.allSettled([getGoalProgress(), listGoals('all')])

    if (progressRes.status === 'fulfilled') {
      activeGoals.value = progressRes.value.data.goals || []
    }

    if (historyRes.status === 'fulfilled') {
      const allGoals = historyRes.value.data.goals || []
      // 过滤出非活跃的历史目标
      historyGoals.value = allGoals
        .filter((g) => g.status === 'completed' || g.status === 'failed')
        .sort((a, b) => new Date(b.updated_at || b.created_at) - new Date(a.updated_at || a.created_at))
    }
  } catch (e) {
    console.error('加载目标失败:', e)
    toastError('加载目标失败')
  } finally {
    loading.value = false
  }
}

// ===== 创建目标 =====
async function handleCreate() {
  if (!form.value.target_value || form.value.target_value < 1) return
  creating.value = true
  try {
    await createGoal({
      type: form.value.type,
      target_value: form.value.target_value,
      period: form.value.period
    })
    success('目标创建成功！')
    showCreateForm.value = false
    form.value = { type: 'review_cards', target_value: 20, period: 'weekly' }
    await loadData()
  } catch (e) {
    console.error('创建目标失败:', e)
    toastError('创建目标失败')
  } finally {
    creating.value = false
  }
}

// ===== 删除目标 =====
async function handleDelete(id) {
  try {
    await deleteGoal(id)
    success('目标已删除')
    await loadData()
  } catch (e) {
    toastError('删除失败')
  }
}

// ===== 生命周期 =====
onMounted(() => {
  loadData()
})
</script>

<style scoped>
.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.25s ease;
  overflow: hidden;
}
.slide-down-enter-from,
.slide-down-leave-to {
  opacity: 0;
  transform: translateY(-12px);
  max-height: 0;
}
.slide-down-enter-to,
.slide-down-leave-from {
  opacity: 1;
  transform: translateY(0);
  max-height: 400px;
}
</style>
