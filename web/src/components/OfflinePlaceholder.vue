<template>
  <div class="flex flex-col items-center justify-center py-16 px-6 text-center">
    <!-- 图标 -->
    <div
      class="w-20 h-20 rounded-2xl flex items-center justify-center mb-5"
      :class="isDark ? 'bg-gray-800' : 'bg-gray-100'"
    >
      <svg
        class="w-10 h-10"
        :class="isDark ? 'text-gray-500' : 'text-gray-400'"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M18.364 5.636a9 9 0 010 12.728M5.636 5.636a9 9 0 000 12.728"
        />
        <line x1="1" y1="1" x2="23" y2="23" stroke-width="2" stroke-linecap="round" />
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M8.53 16.11a6 6 0 016.95 0M12 20h.01"
        />
      </svg>
    </div>

    <!-- 标题 -->
    <h3 class="text-lg font-semibold mb-2" :class="isDark ? 'text-gray-200' : 'text-gray-800'">{{ title }}</h3>

    <!-- 描述 -->
    <p class="text-sm max-w-md leading-relaxed mb-6" :class="isDark ? 'text-gray-400' : 'text-gray-500'">
      {{ description }}
    </p>

    <!-- 操作按钮 -->
    <div class="flex gap-3">
      <button
        v-if="showRetry"
        class="px-4 py-2 text-sm font-medium rounded-lg text-white bg-gradient-to-r from-primary-600 to-indigo-600 hover:from-primary-700 hover:to-indigo-700 transition-colors"
        @click="$emit('retry')"
      >
        重新连接后重试
      </button>
      <router-link
        v-if="fallbackRoute"
        :to="fallbackRoute"
        class="px-4 py-2 text-sm font-medium rounded-lg border transition-colors"
        :class="
          isDark ? 'border-gray-600 text-gray-300 hover:bg-gray-800' : 'border-gray-300 text-gray-600 hover:bg-gray-50'
        "
      >
        {{ fallbackLabel }}
      </router-link>
    </div>

    <!-- 离线队列信息 -->
    <div
      v-if="pendingCount > 0"
      class="mt-6 px-4 py-2.5 rounded-lg text-xs"
      :class="
        isDark
          ? 'bg-amber-900/20 text-amber-400 border border-amber-800/30'
          : 'bg-amber-50 text-amber-700 border border-amber-200'
      "
    >
      <div class="flex items-center gap-2">
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span>{{ pendingCount }} 项操作正在排队，联网后将自动同步</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useDarkMode } from '../composables/useDarkMode'
import { useOfflineQueue } from '../composables/useOfflineQueue'

defineProps({
  title: { type: String, default: '此功能需要网络连接' },
  description: { type: String, default: '请检查您的网络设置后重试。部分学习功能在离线状态下不可用。' },
  showRetry: { type: Boolean, default: true },
  fallbackRoute: { type: String, default: '' },
  fallbackLabel: { type: String, default: '返回首页' }
})

defineEmits(['retry'])

const { isDark } = useDarkMode()
const { pendingCount } = useOfflineQueue()
</script>
