<template>
  <ErrorBoundary>
    <router-view v-slot="{ Component }">
      <transition name="fade-slide" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </ErrorBoundary>
  <Toast />
</template>

<script setup>
import { useAuthStore } from './stores/auth'
import Toast from './components/Toast.vue'
import ErrorBoundary from './components/ErrorBoundary.vue'
import { useDarkMode } from './composables/useDarkMode'

// 初始化时从 localStorage 恢复登录状态
const _auth = useAuthStore()

// 初始化暗色模式（从 localStorage 恢复或跟随系统）
const { isDark: _isDark } = useDarkMode()
</script>

<style>
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}
</style>
