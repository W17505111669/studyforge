<template>
  <div class="border-t border-dark-700">
    <!-- 折叠标题栏 -->
    <button
      class="w-full flex items-center justify-between px-4 py-3 text-sm font-medium transition-colors"
      :class="
        isDark ? 'text-gray-300 hover:text-white hover:bg-dark-800' : 'text-gray-400 hover:text-white hover:bg-dark-800'
      "
      @click="collapsed = !collapsed"
    >
      <div class="flex items-center gap-2">
        <svg class="w-4 h-4" viewBox="0 0 20 20" fill="currentColor">
          <path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" />
          <path
            fill-rule="evenodd"
            d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z"
            clip-rule="evenodd"
          />
        </svg>
        <span>今日任务</span>
        <span
          v-if="tasks.length > 0"
          class="px-1.5 py-0.5 rounded-full text-[10px] font-bold leading-none"
          :class="allCompleted ? 'bg-emerald-500/20 text-emerald-400' : 'bg-primary-500/20 text-primary-400'"
        >
          {{ completedCount }}/{{ tasks.length }}
        </span>
      </div>
      <svg
        class="w-4 h-4 transition-transform duration-200"
        :class="{ 'rotate-180': !collapsed }"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <!-- 任务列表（可折叠） -->
    <Transition name="task-collapse">
      <div v-if="!collapsed" class="px-3 pb-3 space-y-1.5">
        <!-- 全部完成徽章 -->
        <div
          v-if="allCompleted && tasks.length > 0"
          class="text-center py-2 px-3 rounded-lg mb-2"
          :class="
            isDark ? 'bg-emerald-900/30 border border-emerald-700/50' : 'bg-emerald-900/20 border border-emerald-700/40'
          "
        >
          <span class="text-xs font-semibold text-emerald-400">🎉 今日全部完成！</span>
        </div>

        <!-- 每项任务 -->
        <div
          v-for="task in tasks"
          :key="task.id"
          class="flex items-start gap-2 px-2 py-2 rounded-lg transition-colors"
          :class="[isDark ? 'hover:bg-dark-800' : 'hover:bg-dark-800', task.is_completed ? 'opacity-60' : '']"
        >
          <!-- Checkbox -->
          <button
            class="mt-0.5 w-4 h-4 rounded border flex items-center justify-center flex-shrink-0 transition-all"
            :class="
              task.is_completed
                ? 'bg-emerald-500 border-emerald-500'
                : isDark
                  ? 'border-gray-500 hover:border-primary-400'
                  : 'border-gray-500 hover:border-primary-400'
            "
            @click="handleToggle(task)"
          >
            <svg v-if="task.is_completed" class="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
              <path
                fill-rule="evenodd"
                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                clip-rule="evenodd"
              />
            </svg>
          </button>

          <!-- 任务描述 + 进度条 -->
          <div class="flex-1 min-w-0">
            <p
              class="text-xs leading-tight"
              :class="task.is_completed ? 'line-through text-gray-500' : isDark ? 'text-gray-300' : 'text-gray-300'"
            >
              {{ taskLabel(task) }}
            </p>
            <!-- 进度条 -->
            <div class="mt-1 h-1 rounded-full overflow-hidden" :class="isDark ? 'bg-gray-700' : 'bg-gray-600'">
              <div
                class="h-full rounded-full transition-all duration-500"
                :style="{ width: taskPercent(task) + '%' }"
                :class="
                  taskPercent(task) >= 100
                    ? 'bg-emerald-500'
                    : taskPercent(task) >= 50
                      ? 'bg-yellow-500'
                      : 'bg-primary-500'
                "
              ></div>
            </div>
            <p class="text-[10px] mt-0.5" :class="isDark ? 'text-gray-500' : 'text-gray-500'">
              {{ task.completed_count }}/{{ task.target_count }}
            </p>
          </div>
        </div>

        <!-- 加载中 -->
        <div v-if="loading" class="text-center py-2">
          <div class="w-4 h-4 border-2 border-primary-500 border-t-transparent rounded-full animate-spin mx-auto"></div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useDarkMode } from '../composables/useDarkMode'
import { getDailyTasks, toggleDailyTask } from '../api/client'

const { isDark } = useDarkMode()

const tasks = ref([])
const loading = ref(false)
const collapsed = ref(false)

const completedCount = computed(() => tasks.value.filter((t) => t.is_completed).length)
const allCompleted = computed(() => tasks.value.length > 0 && completedCount.value === tasks.value.length)

function taskLabel(task) {
  const labels = {
    review_due_cards: '复习到期卡片',
    complete_n_quizzes: '完成练习题',
    study_n_minutes: '学习时长(分钟)',
    read_material: '阅读材料',
    upload_material: '上传材料'
  }
  return labels[task.type] || task.type
}

function taskPercent(task) {
  if (!task.target_count || task.target_count <= 0) return 0
  return Math.min(100, Math.round((task.completed_count / task.target_count) * 100))
}

async function loadTasks() {
  loading.value = true
  try {
    const res = await getDailyTasks()
    tasks.value = res.data.tasks || []
  } catch {
    tasks.value = []
  } finally {
    loading.value = false
  }
}

async function handleToggle(task) {
  try {
    const res = await toggleDailyTask(task.id)
    // 更新本地状态
    const idx = tasks.value.findIndex((t) => t.id === task.id)
    if (idx !== -1) {
      tasks.value[idx] = res.data
    }
  } catch {}
}

// 暴露刷新方法给父组件
defineExpose({ loadTasks })

onMounted(() => {
  loadTasks()
})
</script>

<style scoped>
.task-collapse-enter-active,
.task-collapse-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}
.task-collapse-enter-from,
.task-collapse-leave-to {
  opacity: 0;
  max-height: 0;
}
.task-collapse-enter-to,
.task-collapse-leave-from {
  opacity: 1;
  max-height: 400px;
}
</style>
