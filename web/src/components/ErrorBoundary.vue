<template>
  <div v-if="error" class="error-fallback">
    <div class="error-fallback-inner">
      <svg class="error-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
      <h3 class="error-title">组件出错了</h3>
      <p class="error-message">{{ errorMessage }}</p>
      <button class="retry-btn" @click="reset">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="retry-icon">
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
  // 返回 false 阻止错误继续向上传播
  return false
})

function reset() {
  error.value = null
}
</script>

<style scoped>
.error-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  min-height: 200px;
}
.error-fallback-inner {
  text-align: center;
  padding: 2rem;
  border-radius: 12px;
  border: 1px solid #fecaca;
  background-color: #fef2f2;
  max-width: 400px;
  width: 100%;
}
.error-icon {
  width: 48px;
  height: 48px;
  color: #ef4444;
  margin: 0 auto 1rem;
}
.error-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #991b1b;
  margin: 0 0 0.5rem;
}
.error-message {
  font-size: 0.875rem;
  color: #b91c1c;
  margin: 0 0 1.25rem;
  line-height: 1.5;
  word-break: break-word;
}
.retry-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1.25rem;
  border-radius: 8px;
  border: 1px solid #fca5a5;
  background: white;
  color: #dc2626;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}
.retry-btn:hover {
  background: #fef2f2;
  border-color: #f87171;
}
.retry-icon {
  width: 16px;
  height: 16px;
}

/* 暗色模式 */
:global(.dark) .error-fallback-inner {
  background-color: rgba(127, 29, 29, 0.3);
  border-color: rgba(239, 68, 68, 0.3);
}
:global(.dark) .error-icon {
  color: #f87171;
}
:global(.dark) .error-title {
  color: #fca5a5;
}
:global(.dark) .error-message {
  color: #fca5a5;
}
:global(.dark) .retry-btn {
  background: rgba(127, 29, 29, 0.4);
  border-color: rgba(239, 68, 68, 0.4);
  color: #fca5a5;
}
:global(.dark) .retry-btn:hover {
  background: rgba(127, 29, 29, 0.6);
  border-color: rgba(239, 68, 68, 0.6);
}
</style>
