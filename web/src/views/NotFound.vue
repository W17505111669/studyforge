<template>
  <div class="flex-1 flex items-center justify-center p-6 min-h-full">
    <div class="text-center max-w-lg">
      <!-- 404 动画数字 -->
      <div class="relative mb-6">
        <h1 class="text-[10rem] font-black leading-none tracking-tighter select-none"
            style="background: linear-gradient(135deg, #6366f1, #8b5cf6, #a78bfa);
                   -webkit-background-clip: text; -webkit-text-fill-color: transparent;
                   background-clip: text;">
          404
        </h1>
        <!-- 装饰圆点 -->
        <div class="absolute -top-2 -left-4 w-4 h-4 rounded-full bg-primary-400 opacity-60 animate-bounce" style="animation-delay: 0.1s"></div>
        <div class="absolute top-8 -right-2 w-3 h-3 rounded-full bg-purple-400 opacity-50 animate-bounce" style="animation-delay: 0.4s"></div>
        <div class="absolute bottom-12 -left-8 w-2 h-2 rounded-full bg-indigo-400 opacity-40 animate-bounce" style="animation-delay: 0.7s"></div>
      </div>

      <!-- 文案 -->
      <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-200 mb-3">
        页面走丢了
      </h2>
      <p class="text-gray-500 dark:text-gray-400 mb-8 leading-relaxed">
        你访问的页面不存在或已被移动。<br/>
        试试回到首页，或者从侧边栏导航到其他页面。
      </p>

      <!-- 路径提示 -->
      <div v-if="currentPath" class="mb-8 inline-flex items-center gap-2 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-sm px-4 py-2 rounded-lg font-mono">
        <svg class="w-4 h-4 text-gray-400 dark:text-gray-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/>
        </svg>
        <span class="truncate max-w-xs">{{ currentPath }}</span>
      </div>

      <!-- 操作按钮 -->
      <div class="flex items-center justify-center gap-4">
        <button
          @click="goHome"
          class="px-6 py-3 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg
                 transition-all duration-200 shadow-lg shadow-primary-600/20 hover:shadow-primary-600/30
                 flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
          </svg>
          回到首页
        </button>
        <button
          @click="goBack"
          class="px-6 py-3 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 font-medium rounded-lg
                 border border-gray-200 dark:border-gray-600 transition-all duration-200
                 flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
          </svg>
          返回上页
        </button>
      </div>

      <!-- 快捷入口 -->
      <div class="mt-10 pt-8 border-t border-gray-100 dark:border-gray-700">
        <p class="text-xs text-gray-400 dark:text-gray-500 mb-4">快捷入口</p>
        <div class="flex items-center justify-center gap-3 flex-wrap">
          <router-link
            v-for="link in quickLinks"
            :key="link.path"
            :to="link.path"
            class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-600 dark:text-gray-400
                   bg-gray-50 dark:bg-gray-700 hover:bg-primary-50 dark:hover:bg-primary-900/20 hover:text-primary-600 dark:hover:text-primary-400 rounded-lg
                   transition-colors duration-200"
          >
            <span v-html="link.icon" class="w-4 h-4 flex items-center justify-center"></span>
            {{ link.label }}
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const currentPath = computed(() => {
  const p = route.params.pathMatch
  if (!p) return ''
  const path = '/' + (Array.isArray(p) ? p.join('/') : p)
  return path === '/' ? '' : path
})

function goHome() {
  router.push('/')
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/')
  }
}

const quickLinks = [
  {
    path: '/upload',
    label: '上传材料',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/></svg>'
  },
  {
    path: '/cards',
    label: '知识卡片',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/></svg>'
  },
  {
    path: '/quiz',
    label: '练习测验',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg>'
  },
  {
    path: '/chat',
    label: 'AI 对话',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"/></svg>'
  },
]
</script>
