<template>
  <!-- 移动端底部导航栏 -->
  <div class="md:hidden fixed bottom-0 left-0 right-0 z-30">
    <!-- FAB 展开菜单遮罩 -->
    <Transition name="fab-overlay">
      <div
        v-if="showFabMenu"
        class="fixed inset-0 bg-black/30 dark:bg-black/50 z-40"
        @click="showFabMenu = false"
      ></div>
    </Transition>

    <!-- FAB 展开菜单 -->
    <Transition name="fab-menu">
      <div v-if="showFabMenu" class="fixed bottom-20 left-1/2 -translate-x-1/2 z-50 flex flex-col items-center gap-3">
        <button
          v-for="(item, idx) in fabItems"
          :key="idx"
          class="flex items-center gap-3 px-5 py-3 rounded-2xl border shadow-lg backdrop-blur-sm transition-all duration-200 hover:scale-105 active:scale-95 min-w-[160px]"
          :class="
            isDark
              ? 'bg-gray-800/95 border-gray-700 text-gray-200 hover:bg-gray-700/95'
              : 'bg-white/95 border-gray-200 text-gray-700 hover:bg-gray-50/95'
          "
          :style="{ transitionDelay: `${idx * 50}ms` }"
          @click="handleFabClick(item)"
        >
          <span class="w-9 h-9 rounded-full flex items-center justify-center flex-shrink-0" :class="item.bgClass">
            <span class="w-5 h-5" v-html="item.icon"></span>
          </span>
          <span class="text-sm font-medium">{{ item.label }}</span>
        </button>
      </div>
    </Transition>

    <!-- 导航栏主体 -->
    <div
      class="relative flex items-end justify-around border-t safe-area-bottom"
      :class="
        isDark ? 'bg-gray-900/98 border-gray-700 backdrop-blur-lg' : 'bg-white/98 border-gray-200 backdrop-blur-lg'
      "
      style="min-height: 56px"
    >
      <!-- 左侧两项 -->
      <button
        v-for="item in leftItems"
        :key="item.path"
        class="flex flex-col items-center justify-center flex-1 py-2 gap-1 transition-all duration-200"
        :class="isActive(item.path) ? navActiveClass : navInactiveClass"
        @click="navigateTo(item.path)"
      >
        <span
          class="w-6 h-6 transition-transform duration-200"
          :class="{ 'scale-110': isActive(item.path) }"
          v-html="item.icon"
        ></span>
        <span class="text-[10px] font-medium leading-tight">{{ item.label }}</span>
        <!-- 活跃指示器小圆点 -->
        <span v-if="isActive(item.path)" class="w-1 h-1 rounded-full bg-primary-500 -mt-0.5"></span>
      </button>

      <!-- 中间 FAB 按钮 -->
      <div class="flex flex-col items-center flex-1 relative">
        <button
          class="relative -top-4 w-14 h-14 rounded-full flex items-center justify-center shadow-lg transition-all duration-300 active:scale-90"
          :class="[
            showFabMenu
              ? 'bg-red-500 dark:bg-red-600 rotate-45'
              : 'bg-primary-500 dark:bg-primary-600 hover:bg-primary-600 dark:hover:bg-primary-500'
          ]"
          @click.stop="toggleFabMenu"
        >
          <!-- 加号图标（展开时旋转为 X） -->
          <svg
            class="w-7 h-7 text-white transition-transform duration-300"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" />
          </svg>
        </button>
        <span
          class="text-[10px] font-medium leading-tight -mt-1 transition-colors duration-200"
          :class="isDark ? 'text-gray-500' : 'text-gray-400'"
        >
          快捷
        </span>
      </div>

      <!-- 右侧两项 -->
      <button
        v-for="item in rightItems"
        :key="item.path"
        class="flex flex-col items-center justify-center flex-1 py-2 gap-1 transition-all duration-200"
        :class="isActive(item.path) ? navActiveClass : navInactiveClass"
        @click="navigateTo(item.path)"
      >
        <span
          class="w-6 h-6 transition-transform duration-200"
          :class="{ 'scale-110': isActive(item.path) }"
          v-html="item.icon"
        ></span>
        <span class="text-[10px] font-medium leading-tight">{{ item.label }}</span>
        <span v-if="isActive(item.path)" class="w-1 h-1 rounded-full bg-primary-500 -mt-0.5"></span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDarkMode } from '../composables/useDarkMode'

const route = useRoute()
const router = useRouter()
const { isDark } = useDarkMode()

const showFabMenu = ref(false)

// 路由变化时关闭 FAB 菜单
watch(
  () => route.path,
  () => {
    showFabMenu.value = false
  }
)

// 导航项目定义
const leftItems = [
  {
    path: '/',
    label: '首页',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z"/></svg>'
  },
  {
    path: '/upload',
    label: '上传',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM6.293 6.707a1 1 0 010-1.414l3-3a1 1 0 011.414 0l3 3a1 1 0 01-1.414 1.414L11 5.414V13a1 1 0 11-2 0V5.414L7.707 6.707a1 1 0 01-1.414 0z" clip-rule="evenodd"/></svg>'
  }
]

const rightItems = [
  {
    path: '/chat',
    label: '对话',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M18 10c0 3.866-3.582 7-8 7a8.841 8.841 0 01-4.083-.98L2 17l1.338-3.123C2.493 12.767 2 11.434 2 10c0-3.866 3.582-7 8-7s8 3.134 8 7zM7 9H5v2h2V9zm8 0h-2v2h2V9zm-4 0H9v2h2V9z" clip-rule="evenodd"/></svg>'
  },
  {
    path: '/cards',
    label: '卡片',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"/></svg>'
  }
]

// FAB 展开菜单项
const fabItems = [
  {
    path: '/review',
    label: '开始复习',
    bgClass: 'bg-amber-100 text-amber-600 dark:bg-amber-900/40 dark:text-amber-400',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z"/><path fill-rule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm9.707 5.707a1 1 0 00-1.414-1.414L9 12.586l-1.293-1.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/></svg>'
  },
  {
    path: '/quiz',
    label: '快速练习',
    bgClass: 'bg-blue-100 text-blue-600 dark:bg-blue-900/40 dark:text-blue-400',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z"/><path fill-rule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clip-rule="evenodd"/></svg>'
  },
  {
    path: '/pomodoro',
    label: '开始专注',
    bgClass: 'bg-emerald-100 text-emerald-600 dark:bg-emerald-900/40 dark:text-emerald-400',
    icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd"/></svg>'
  }
]

// 活跃状态样式
const navActiveClass = 'text-primary-500 dark:text-primary-400'
const navInactiveClass = computed(() =>
  isDark.value ? 'text-gray-500 hover:text-gray-300' : 'text-gray-400 hover:text-gray-600'
)

function isActive(path) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

function navigateTo(path) {
  if (route.path !== path) {
    router.push(path)
  }
  showFabMenu.value = false
}

function toggleFabMenu() {
  showFabMenu.value = !showFabMenu.value
}

function handleFabClick(item) {
  showFabMenu.value = false
  router.push(item.path)
}
</script>

<style scoped>
/* iPhone 安全区域适配 */
.safe-area-bottom {
  padding-bottom: env(safe-area-inset-bottom, 0px);
}

/* FAB 菜单过渡动画 */
.fab-menu-enter-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.fab-menu-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}
.fab-menu-enter-from {
  opacity: 0;
  transform: translate(-50%, 20px) scale(0.8);
}
.fab-menu-leave-to {
  opacity: 0;
  transform: translate(-50%, 10px) scale(0.9);
}
.fab-menu-enter-to,
.fab-menu-leave-from {
  opacity: 1;
  transform: translate(-50%, 0) scale(1);
}

/* FAB 遮罩过渡 */
.fab-overlay-enter-active {
  transition: opacity 0.25s ease;
}
.fab-overlay-leave-active {
  transition: opacity 0.15s ease;
}
.fab-overlay-enter-from,
.fab-overlay-leave-to {
  opacity: 0;
}
.fab-overlay-enter-to,
.fab-overlay-leave-from {
  opacity: 1;
}

/* FAB 子项入场延迟动画 */
.fab-menu-enter-active button {
  transition:
    opacity 0.2s ease,
    transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}
</style>
