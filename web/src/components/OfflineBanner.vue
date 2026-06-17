<template>
  <Teleport to="body">
    <Transition name="banner-slide">
      <div
        v-if="show"
        class="fixed top-0 left-0 right-0 z-[9999] px-4 py-2 text-center text-sm font-medium transition-colors"
        :class="bannerClass"
      >
        <div class="flex items-center justify-center gap-2 max-w-2xl mx-auto">
          <!-- 离线状态 -->
          <template v-if="!isOnline">
            <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M18.364 5.636a9 9 0 010 12.728M5.636 5.636a9 9 0 000 12.728M1.42 9a16 16 0 014.7-2.88M8.53 16.11a6 6 0 016.95 0"
              />
              <line x1="1" y1="1" x2="23" y2="23" stroke-width="2" stroke-linecap="round" />
            </svg>
            <span>当前处于离线模式</span>
            <span
              v-if="pendingCount > 0"
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-white/20"
            >
              {{ pendingCount }} 项操作待同步
            </span>
            <span class="text-xs opacity-75 hidden sm:inline">— 部分功能受限，已缓存的数据仍可查看</span>
          </template>

          <!-- 回连状态 + 同步中 -->
          <template v-else-if="syncing">
            <svg class="w-4 h-4 flex-shrink-0 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            <span>网络已恢复，正在同步 {{ pendingCount }} 项操作...</span>
          </template>

          <!-- 同步完成 -->
          <template v-else-if="showSyncResult">
            <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
            <span>网络已恢复</span>
            <span v-if="syncResult" class="text-xs opacity-80">
              — {{ syncResult.success }} 项操作已同步
              <template v-if="syncResult.failed > 0">, {{ syncResult.failed }} 项失败</template>
            </span>
          </template>

          <!-- 数据来自缓存提示 -->
          <template v-else-if="showCachedHint">
            <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 7v10c0 2 1 3 3 3h10c2 0 3-1 3-3V7c0-2-1-3-3-3H7C5 4 4 5 4 7z"
              />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6M12 9v6" />
            </svg>
            <span>正在查看缓存数据</span>
          </template>

          <!-- 关闭按钮 (非同步中) -->
          <button
            v-if="!syncing"
            class="ml-2 p-0.5 rounded hover:bg-white/20 transition-colors flex-shrink-0"
            aria-label="关闭"
            @click="dismiss"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useNetworkStatus } from '../composables/useNetworkStatus'
import { useOfflineQueue } from '../composables/useOfflineQueue'

const { isOnline, wasOffline } = useNetworkStatus()
const {
  pendingCount,
  syncing,
  lastSyncResult,
  init: initQueue,
  syncQueue: _syncQueue,
  onQueueEvent: _onQueueEvent
} = useOfflineQueue()

const dismissed = ref(false)
const showSyncResult = ref(false)
const showCachedHint = ref(false)
let syncTimer = null
let cachedTimer = null

// 初始化离线队列
onMounted(async () => {
  await initQueue()
})

onUnmounted(() => {
  if (syncTimer) clearTimeout(syncTimer)
  if (cachedTimer) clearTimeout(cachedTimer)
})

// 监听网络状态变化
watch(isOnline, (online) => {
  if (!online) {
    // 变为离线
    dismissed.value = false
    showSyncResult.value = false
    showCachedHint.value = false
  } else if (wasOffline.value) {
    // 从离线恢复
    dismissed.value = false
    showSyncResult.value = false
    // 如果有待同步操作, syncQueue 会被 online 事件自动触发
    if (pendingCount.value === 0) {
      // 没有待同步操作, 直接显示恢复提示
      showSyncResult.value = true
      syncTimer = setTimeout(() => {
        showSyncResult.value = false
      }, 4000)
    }
  }
})

// 监听同步完成
watch(lastSyncResult, (result) => {
  if (result && result.success > 0) {
    showSyncResult.value = true
    if (syncTimer) clearTimeout(syncTimer)
    syncTimer = setTimeout(() => {
      showSyncResult.value = false
    }, 5000)
  }
})

const syncResult = computed(() => lastSyncResult.value)

const show = computed(() => {
  if (dismissed.value) return false
  if (!isOnline.value) return true
  if (syncing.value) return true
  if (showSyncResult.value) return true
  if (showCachedHint.value) return true
  return false
})

const bannerClass = computed(() => {
  if (!isOnline.value) {
    return 'bg-amber-600 text-white'
  }
  if (syncing.value) {
    return 'bg-blue-600 text-white'
  }
  if (showSyncResult.value) {
    return 'bg-emerald-600 text-white'
  }
  if (showCachedHint.value) {
    return 'bg-sky-600 text-white'
  }
  return 'bg-gray-700 text-white'
})

function dismiss() {
  dismissed.value = true
}

/**
 * 外部调用: 显示"正在查看缓存数据"提示
 */
function showCacheHint() {
  showCachedHint.value = true
  if (cachedTimer) clearTimeout(cachedTimer)
  cachedTimer = setTimeout(() => {
    showCachedHint.value = false
  }, 4000)
}

function hideCacheHint() {
  showCachedHint.value = false
}

// 暴露给父组件
defineExpose({ showCacheHint, hideCacheHint })
</script>

<style scoped>
.banner-slide-enter-active,
.banner-slide-leave-active {
  transition:
    transform 0.3s ease,
    opacity 0.3s ease;
}
.banner-slide-enter-from,
.banner-slide-leave-to {
  transform: translateY(-100%);
  opacity: 0;
}
</style>
