<template>
  <!-- Toast 通知 -->
  <div class="fixed top-4 right-4 z-50 flex flex-col gap-2 pointer-events-none">
    <div
      v-for="toast in toasts"
      :key="toast.id"
      class="pointer-events-auto px-4 py-3 rounded-lg shadow-lg border min-w-[280px] max-w-[400px] transition-all duration-300 flex items-start gap-3"
      :class="[toast.visible ? 'opacity-100 translate-x-0' : 'opacity-0 translate-x-8', toastClasses(toast.type)]"
    >
      <span class="text-lg mt-0.5">{{ toastIcon(toast.type) }}</span>
      <span class="text-sm font-medium flex-1">{{ toast.message }}</span>
      <button class="text-current opacity-40 hover:opacity-80 text-xs mt-0.5" @click="remove(toast.id)">✕</button>
    </div>
  </div>

  <!-- Confirm 弹窗 -->
  <Teleport to="body">
    <div
      v-if="confirmState.show"
      class="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 backdrop-blur-sm"
      @click.self="onCancel"
    >
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl p-6 max-w-sm w-full mx-4 animate-fade-in-up">
        <p class="text-gray-800 dark:text-gray-200 font-medium mb-6">{{ confirmState.message }}</p>
        <div class="flex gap-3 justify-end">
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-gray-600 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            @click="onCancel"
          >
            取消
          </button>
          <button
            class="px-4 py-2 rounded-lg text-sm font-medium text-white bg-red-500 hover:bg-red-600 transition-colors"
            @click="onConfirm"
          >
            确定
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import { useToast, useConfirm } from '../composables/useToast'

const { toasts, remove } = useToast()
const { confirmState, onConfirm, onCancel } = useConfirm()

// Esc 关闭确认弹窗
function handleEscKeydown(e) {
  if (e.key === 'Escape' && confirmState.show) {
    onCancel()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleEscKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleEscKeydown)
})

function toastClasses(type) {
  const map = {
    success:
      'bg-green-50 dark:bg-green-900/30 border-green-200 dark:border-green-700 text-green-800 dark:text-green-300',
    error: 'bg-red-50 dark:bg-red-900/30 border-red-200 dark:border-red-700 text-red-800 dark:text-red-300',
    warning:
      'bg-amber-50 dark:bg-amber-900/30 border-amber-200 dark:border-amber-700 text-amber-800 dark:text-amber-300',
    info: 'bg-blue-50 dark:bg-blue-900/30 border-blue-200 dark:border-blue-700 text-blue-800 dark:text-blue-300'
  }
  return map[type] || map.info
}

function toastIcon(type) {
  const map = { success: '✓', error: '✕', warning: '⚠', info: 'ℹ' }
  return map[type] || 'ℹ'
}
</script>
