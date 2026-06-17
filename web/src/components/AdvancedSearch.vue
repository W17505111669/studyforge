<template>
  <div class="border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/30">
    <!-- 折叠切换按钮 -->
    <button
      class="w-full flex items-center justify-between px-5 py-2.5 text-xs font-medium text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 transition-colors"
      @click="expanded = !expanded"
    >
      <span class="flex items-center gap-1.5">
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"
          />
        </svg>
        高级搜索
        <span
          v-if="activeFilterCount > 0"
          class="ml-1 px-1.5 py-0.5 bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 rounded-full text-[10px] font-bold"
        >
          {{ activeFilterCount }}
        </span>
      </span>
      <svg
        :class="expanded ? 'rotate-180' : ''"
        class="w-3.5 h-3.5 transition-transform duration-200"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <!-- 过滤面板 -->
    <Transition name="panel-slide">
      <div v-if="expanded" class="px-5 pb-4 space-y-3">
        <!-- 类型多选 -->
        <div>
          <label class="block text-[11px] font-medium text-gray-500 dark:text-gray-400 mb-1.5">搜索类型</label>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="opt in typeOptions"
              :key="opt.value"
              :class="
                isTypeSelected(opt.value)
                  ? 'bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-400 border-primary-300 dark:border-primary-700'
                  : 'bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-300 border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
              "
              class="px-2.5 py-1 text-xs rounded-md border transition-colors"
              @click="toggleType(opt.value)"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>

        <!-- 时间范围 -->
        <div>
          <label class="block text-[11px] font-medium text-gray-500 dark:text-gray-400 mb-1.5">时间范围</label>
          <div class="flex flex-wrap gap-1.5 mb-1.5">
            <button
              v-for="preset in datePresets"
              :key="preset.value"
              :class="
                datePreset === preset.value
                  ? 'bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-400 border-primary-300 dark:border-primary-700'
                  : 'bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-300 border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
              "
              class="px-2.5 py-1 text-xs rounded-md border transition-colors"
              @click="setDatePreset(preset.value)"
            >
              {{ preset.label }}
            </button>
          </div>
          <div v-if="datePreset === 'custom'" class="flex items-center gap-2">
            <input
              v-model="customDateFrom"
              type="date"
              class="flex-1 px-2 py-1 text-xs bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 outline-none focus:border-primary-400"
            />
            <span class="text-gray-400 text-xs">至</span>
            <input
              v-model="customDateTo"
              type="date"
              class="flex-1 px-2 py-1 text-xs bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 outline-none focus:border-primary-400"
            />
          </div>
        </div>

        <!-- 标签多选 -->
        <div>
          <label class="block text-[11px] font-medium text-gray-500 dark:text-gray-400 mb-1.5">标签过滤</label>
          <div ref="tagDropdownRef" class="relative">
            <div
              class="flex flex-wrap items-center gap-1 min-h-[30px] px-2 py-1 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md cursor-text"
              @click="showTagDropdown = true"
            >
              <span
                v-for="tag in selectedTags"
                :key="tag"
                class="inline-flex items-center gap-0.5 px-1.5 py-0.5 bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-400 rounded text-[10px] font-medium"
              >
                {{ tag }}
                <button class="hover:text-primary-900 dark:hover:text-primary-200" @click.stop="removeTag(tag)">
                  ×
                </button>
              </span>
              <span v-if="selectedTags.length === 0" class="text-xs text-gray-400 dark:text-gray-500">
                点击选择标签...
              </span>
            </div>
            <Transition name="dropdown-fade">
              <div
                v-if="showTagDropdown && availableTags.length > 0"
                class="absolute z-10 w-full mt-1 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg max-h-32 overflow-y-auto custom-scroll"
              >
                <button
                  v-for="tag in filteredTags"
                  :key="tag.name"
                  class="w-full text-left px-3 py-1.5 text-xs text-gray-700 dark:text-gray-300 hover:bg-primary-50 dark:hover:bg-primary-900/20 flex items-center justify-between transition-colors"
                  @click.stop="toggleTag(tag.name)"
                >
                  <span>{{ tag.name }}</span>
                  <span class="text-[10px] text-gray-400 dark:text-gray-500">{{ tag.count }}</span>
                </button>
              </div>
            </Transition>
          </div>
        </div>

        <!-- 难度选择（仅题目类型） -->
        <div v-if="showDifficulty" class="flex items-center gap-3">
          <div class="flex-1">
            <label class="block text-[11px] font-medium text-gray-500 dark:text-gray-400 mb-1.5">难度</label>
            <div class="flex gap-1.5">
              <button
                v-for="d in difficultyOptions"
                :key="d.value"
                :class="
                  selectedDifficulty === d.value
                    ? d.activeClass
                    : 'bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-300 border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
                "
                class="px-2.5 py-1 text-xs rounded-md border transition-colors"
                @click="selectedDifficulty = selectedDifficulty === d.value ? '' : d.value"
              >
                {{ d.label }}
              </button>
            </div>
          </div>
        </div>

        <!-- 材料状态 -->
        <div v-if="showStatus">
          <label class="block text-[11px] font-medium text-gray-500 dark:text-gray-400 mb-1.5">材料状态</label>
          <div class="flex gap-1.5">
            <button
              v-for="s in statusOptions"
              :key="s.value"
              :class="
                selectedStatus === s.value
                  ? s.activeClass
                  : 'bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-300 border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
              "
              class="px-2.5 py-1 text-xs rounded-md border transition-colors"
              @click="selectedStatus = selectedStatus === s.value ? '' : s.value"
            >
              {{ s.label }}
            </button>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="flex items-center gap-2 pt-1">
          <button
            class="flex-1 px-3 py-1.5 text-xs font-medium text-white bg-primary-600 hover:bg-primary-700 dark:bg-primary-600 dark:hover:bg-primary-700 rounded-md transition-colors"
            @click="$emit('search')"
          >
            搜索
          </button>
          <button
            class="px-3 py-1.5 text-xs font-medium text-gray-600 dark:text-gray-400 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600 rounded-md transition-colors"
            @click="resetFilters"
          >
            重置
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { getTags } from '../api/client'

const emit = defineEmits(['update:filters', 'search'])

const expanded = ref(false)
const showTagDropdown = ref(false)
const tagDropdownRef = ref(null)

// 类型选项
const typeOptions = [
  { value: 'material', label: '材料' },
  { value: 'card', label: '卡片' },
  { value: 'quiz', label: '练习题' }
]

// 时间预设
const datePresets = [
  { value: '', label: '不限' },
  { value: 'today', label: '今天' },
  { value: 'week', label: '本周' },
  { value: 'month', label: '本月' },
  { value: 'custom', label: '自定义' }
]

// 难度选项
const difficultyOptions = [
  {
    value: 'easy',
    label: '简单',
    activeClass:
      'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 border-green-300 dark:border-green-700'
  },
  {
    value: 'medium',
    label: '中等',
    activeClass:
      'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 border-amber-300 dark:border-amber-700'
  },
  {
    value: 'hard',
    label: '困难',
    activeClass: 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 border-red-300 dark:border-red-700'
  }
]

// 状态选项
const statusOptions = [
  {
    value: 'pending',
    label: '待分析',
    activeClass: 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 border-gray-300 dark:border-gray-500'
  },
  {
    value: 'completed',
    label: '已完成',
    activeClass:
      'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 border-green-300 dark:border-green-700'
  },
  {
    value: 'failed',
    label: '失败',
    activeClass: 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 border-red-300 dark:border-red-700'
  }
]

// 状态
const selectedTypes = ref([])
const datePreset = ref('')
const customDateFrom = ref('')
const customDateTo = ref('')
const selectedTags = ref([])
const selectedDifficulty = ref('')
const selectedStatus = ref('')
const availableTags = ref([])

// 计算属性
const showDifficulty = computed(() => {
  return selectedTypes.value.length === 0 || selectedTypes.value.includes('quiz')
})

const showStatus = computed(() => {
  return selectedTypes.value.length === 0 || selectedTypes.value.includes('material')
})

const activeFilterCount = computed(() => {
  let count = 0
  if (selectedTypes.value.length > 0) count++
  if (datePreset.value) count++
  if (selectedTags.value.length > 0) count++
  if (selectedDifficulty.value) count++
  if (selectedStatus.value) count++
  return count
})

const filteredTags = computed(() => {
  return availableTags.value.filter((t) => !selectedTags.value.includes(t.name)).slice(0, 20)
})

// 计算日期范围
function computeDateRange() {
  if (datePreset.value === 'today') {
    const today = new Date().toISOString().slice(0, 10)
    return { date_from: today, date_to: today }
  }
  if (datePreset.value === 'week') {
    const now = new Date()
    const day = now.getDay() || 7
    const monday = new Date(now)
    monday.setDate(now.getDate() - day + 1)
    const sunday = new Date(monday)
    sunday.setDate(monday.getDate() + 6)
    return {
      date_from: monday.toISOString().slice(0, 10),
      date_to: sunday.toISOString().slice(0, 10)
    }
  }
  if (datePreset.value === 'month') {
    const now = new Date()
    const firstDay = new Date(now.getFullYear(), now.getMonth(), 1)
    const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0)
    return {
      date_from: firstDay.toISOString().slice(0, 10),
      date_to: lastDay.toISOString().slice(0, 10)
    }
  }
  if (datePreset.value === 'custom') {
    return {
      date_from: customDateFrom.value || '',
      date_to: customDateTo.value || ''
    }
  }
  return {}
}

// 导出当前过滤器状态
function buildFilters() {
  const filters = {}
  if (selectedTypes.value.length > 0) {
    filters.type = selectedTypes.value.join(',')
  }
  const dateRange = computeDateRange()
  if (dateRange.date_from) filters.date_from = dateRange.date_from
  if (dateRange.date_to) filters.date_to = dateRange.date_to
  if (selectedTags.value.length > 0) {
    filters.tags = selectedTags.value.join(',')
  }
  if (selectedDifficulty.value) filters.difficulty = selectedDifficulty.value
  if (selectedStatus.value) filters.status = selectedStatus.value
  return filters
}

// 操作函数
function toggleType(type) {
  const idx = selectedTypes.value.indexOf(type)
  if (idx >= 0) {
    selectedTypes.value.splice(idx, 1)
  } else {
    selectedTypes.value.push(type)
  }
  emitFilters()
}

function isTypeSelected(type) {
  return selectedTypes.value.includes(type)
}

function setDatePreset(value) {
  datePreset.value = value
  if (value !== 'custom') {
    customDateFrom.value = ''
    customDateTo.value = ''
  }
  emitFilters()
}

function toggleTag(tag) {
  const idx = selectedTags.value.indexOf(tag)
  if (idx >= 0) {
    selectedTags.value.splice(idx, 1)
  } else {
    selectedTags.value.push(tag)
  }
  emitFilters()
}

function removeTag(tag) {
  const idx = selectedTags.value.indexOf(tag)
  if (idx >= 0) {
    selectedTags.value.splice(idx, 1)
  }
  emitFilters()
}

function resetFilters() {
  selectedTypes.value = []
  datePreset.value = ''
  customDateFrom.value = ''
  customDateTo.value = ''
  selectedTags.value = []
  selectedDifficulty.value = ''
  selectedStatus.value = ''
  emitFilters()
}

function emitFilters() {
  emit('update:filters', buildFilters())
}

// 加载标签
async function loadTags() {
  try {
    const { data } = await getTags()
    availableTags.value = Array.isArray(data) ? data : data.data || []
  } catch {
    availableTags.value = []
  }
}

// 点击外部关闭标签下拉
function handleClickOutside(e) {
  if (tagDropdownRef.value && !tagDropdownRef.value.contains(e.target)) {
    showTagDropdown.value = false
  }
}

onMounted(() => {
  loadTags()
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

// 自定义日期变更时自动更新
watch([customDateFrom, customDateTo], () => {
  if (datePreset.value === 'custom') {
    emitFilters()
  }
})

defineExpose({ buildFilters, resetFilters })
</script>

<style scoped>
.panel-slide-enter-active,
.panel-slide-leave-active {
  transition: all 0.2s ease;
  max-height: 400px;
  overflow: hidden;
}
.panel-slide-enter-from,
.panel-slide-leave-to {
  max-height: 0;
  opacity: 0;
}
.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition: opacity 0.12s ease;
}
.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
}
.custom-scroll::-webkit-scrollbar {
  width: 3px;
}
.custom-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroll::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}
:deep(.dark) .custom-scroll::-webkit-scrollbar-thumb {
  background: #4b5563;
}
</style>
