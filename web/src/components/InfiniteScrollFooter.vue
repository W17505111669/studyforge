<template>
  <!-- 加载中 -->
  <div v-if="loading" class="flex items-center justify-center py-6 gap-2">
    <svg
      class="animate-spin h-5 w-5 text-primary-500"
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
    >
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
      <path
        class="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      ></path>
    </svg>
    <span class="text-sm text-gray-500 dark:text-gray-400">加载中...</span>
  </div>

  <!-- 加载失败 -->
  <div v-else-if="error" class="flex flex-col items-center justify-center py-6 gap-2">
    <span class="text-sm text-red-500 dark:text-red-400">{{ error }}</span>
    <button
      class="px-4 py-1.5 text-sm font-medium rounded-lg bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/40 border border-red-200 dark:border-red-800 transition-colors"
      @click="$emit('retry')"
    >
      重试
    </button>
  </div>

  <!-- 已经到底了 -->
  <div v-else-if="showEnd && !hasMore && totalCount > 0" class="flex items-center justify-center py-6 gap-2">
    <div class="h-px w-8 bg-gray-200 dark:bg-gray-700"></div>
    <span class="text-xs text-gray-400 dark:text-gray-500">已经到底了</span>
    <div class="h-px w-8 bg-gray-200 dark:bg-gray-700"></div>
  </div>
</template>

<script setup>
defineProps({
  loading: { type: Boolean, default: false },
  hasMore: { type: Boolean, default: true },
  error: { type: String, default: null },
  totalCount: { type: Number, default: 0 },
  showEnd: { type: Boolean, default: true }
})
defineEmits(['retry'])
</script>
