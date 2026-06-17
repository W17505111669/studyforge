<template>
  <div
    v-if="error"
    class="error-fallback flex items-center justify-center p-8 min-h-[200px]"
  >
    <div
      class="text-center p-6 rounded-xl border max-w-md w-full bg-red-50 border-red-200 dark:bg-red-950/30 dark:border-red-800/40"
    >
      <svg
        class="error-icon w-12 h-12 mx-auto mb-3 text-red-500 dark:text-red-400"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
      <h3
        class="error-title text-lg font-semibold mb-2 text-red-800 dark:text-red-300"
      >
        组件出错了
      </h3>
      <p
        class="error-message text-sm mb-4 leading-relaxed break-words text-red-600 dark:text-red-400/80"
      >
        {{ errorMessage }}
      </p>
      <button
        class="retry-btn inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all duration-150 bg-white border border-red-300 text-red-600 hover:bg-red-50 hover:border-red-400 dark:bg-red-900/40 dark:border-red-700/50 dark:text-red-300 dark:hover:bg-red-900/60 dark:hover:border-red-600"
        @click="reset"
      >
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          class="w-4 h-4"
        >
          <polyline points="23 4 23 10 17 10" />
          <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10" />
        </svg>
        重试
      </button>
    </div>
  </div>
  <slot v-else />
</template>

<script setup>
import { ref, onErrorCaptured, computed } from 'vue'

const error = ref(null)

const errorMessage = computed(() => {
  if (!error.value) return ''
  return error.value.message || '发生了未知错误'
})

onErrorCaptured((err) => {
  error.value = err
  console.error('[ErrorBoundary] 组件渲染错误:', err)
  return false
})

function reset() {
  error.value = null
}
</script>
