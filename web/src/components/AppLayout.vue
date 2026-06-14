<template>
  <div class="flex h-screen bg-gray-50 dark:bg-gray-900 transition-colors duration-300">
    <!-- 移动端遮罩 -->
    <div v-if="sidebarOpen"
      class="fixed inset-0 bg-black/40 z-30 lg:hidden"
      @click="sidebarOpen = false"
    ></div>

    <!-- 侧边栏 -->
    <aside
      class="w-64 bg-dark-900 text-white flex flex-col flex-shrink-0 z-40 transition-transform duration-300 fixed lg:relative h-full"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'"
    >
      <!-- Logo -->
      <div class="p-6 border-b border-dark-700 flex items-center justify-between">
        <div>
          <h1 class="text-xl font-bold flex items-center gap-2">
            <span class="w-8 h-8 bg-primary-500 rounded-lg flex items-center justify-center text-sm">SF</span>
            StudyForge Pro
          </h1>
          <p class="text-xs text-gray-400 mt-1">AI 智能学习平台</p>
        </div>
        <button @click="sidebarOpen = false" class="lg:hidden text-gray-400 hover:text-white p-1">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <!-- 导航菜单 -->
      <nav class="flex-1 p-4 space-y-1">
        <!-- 全局搜索入口 -->
        <button
          @click="openSearch"
          class="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium text-gray-400 hover:text-white hover:bg-dark-800 transition-all duration-200 mb-2"
        >
          <span class="w-5 h-5 flex items-center justify-center">
            <svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd"/></svg>
          </span>
          搜索
          <span class="ml-auto text-xs text-gray-500 hidden sm:inline">
            <kbd class="px-1.5 py-0.5 rounded border border-gray-600 text-[10px] font-mono">Ctrl+K</kbd>
          </span>
        </button>

        <!-- 通知铃铛 -->
        <button
          @click="toggleNotifPanel"
          class="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium text-gray-400 hover:text-white hover:bg-dark-800 transition-all duration-200 mb-2 relative"
        >
          <span class="w-5 h-5 flex items-center justify-center">
            <svg viewBox="0 0 20 20" fill="currentColor"><path d="M10 2a6 6 0 00-6 6v3.586l-.707.707A1 1 0 004 14h12a1 1 0 00.707-1.707L16 11.586V8a6 6 0 00-6-6zM10 18a3 3 0 01-3-3h6a3 3 0 01-3 3z"/></svg>
          </span>
          通知
          <span
            v-if="unreadCount > 0"
            class="ml-auto min-w-[20px] h-5 px-1.5 flex items-center justify-center rounded-full bg-red-500 text-white text-xs font-bold"
          >
            {{ unreadCount > 99 ? '99+' : unreadCount }}
          </span>
        </button>

        <div class="border-t border-dark-700 mb-2"></div>

        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          @click="sidebarOpen = false"
          :data-onboarding="item.onboardingId"
          class="flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-all duration-200"
          :class="isActive(item.path)
            ? 'bg-primary-600 text-white shadow-lg shadow-primary-600/20'
            : 'text-gray-400 hover:text-white hover:bg-dark-800'"
        >
          <span class="w-5 h-5 flex items-center justify-center" v-html="item.icon"></span>
          {{ item.label }}
        </router-link>
      </nav>

      <!-- 用户信息 + 暗色模式切换 -->
      <div class="p-4 border-t border-dark-700">
        <div class="flex items-center gap-3 px-2">
          <div class="w-8 h-8 bg-primary-500/20 rounded-full flex items-center justify-center text-primary-400 text-sm font-bold">
            {{ auth.user?.username?.charAt(0)?.toUpperCase() || 'U' }}
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium truncate">{{ auth.user?.username || '用户' }}</p>
            <p class="text-xs text-gray-500 truncate">{{ auth.user?.email || '' }}</p>
          </div>
          <!-- 重新引导按钮 -->
          <button
            @click="resetOnboarding"
            class="text-gray-500 hover:text-primary-400 transition-colors p-1"
            title="重新观看引导"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
          </button>
          <!-- 暗色模式切换按钮 -->
          <button
            @click="toggleDarkMode"
            class="text-gray-500 hover:text-yellow-400 transition-colors p-1"
            :title="isDark ? '切换到亮色模式' : '切换到暗色模式'"
          >
            <!-- 月亮图标（亮色模式下显示，点击切换到暗色） -->
            <svg v-if="!isDark" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"/>
            </svg>
            <!-- 太阳图标（暗色模式下显示，点击切换到亮色） -->
            <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"/>
            </svg>
          </button>
          <button
            @click="handleLogout"
            class="text-gray-500 hover:text-red-400 transition-colors p-1"
            title="退出登录"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
            </svg>
          </button>
        </div>
      </div>
    </aside>

    <!-- 主内容区 -->
    <main class="flex-1 overflow-y-auto">
      <!-- 移动端顶部导航栏 -->
      <div class="lg:hidden sticky top-0 z-20 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-4 py-3 flex items-center gap-3">
        <button @click="sidebarOpen = true" class="text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white p-1">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>
          </svg>
        </button>
        <span class="text-sm font-semibold text-gray-800 dark:text-gray-200">{{ currentPageLabel }}</span>
        <!-- 移动端搜索按钮 -->
        <button @click="openSearch" class="text-gray-500 dark:text-gray-400 hover:text-primary-500 p-1">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
          </svg>
        </button>
        <!-- 移动端通知铃铛 -->
        <button @click="toggleNotifPanel" class="relative text-gray-500 dark:text-gray-400 hover:text-primary-500 p-1">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"/>
          </svg>
          <span
            v-if="unreadCount > 0"
            class="absolute -top-1 -right-1 min-w-[16px] h-4 px-1 flex items-center justify-center rounded-full bg-red-500 text-white text-[10px] font-bold"
          >
            {{ unreadCount > 99 ? '99+' : unreadCount }}
          </span>
        </button>
        <!-- 移动端暗色模式切换 -->
        <button @click="toggleDarkMode" class="ml-auto text-gray-500 dark:text-gray-400 hover:text-primary-500 p-1">
          <svg v-if="!isDark" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"/>
          </svg>
          <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"/>
          </svg>
        </button>
      </div>
      <router-view v-slot="{ Component }">
        <transition name="fade-slide" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>

    <!-- 通知面板 -->
    <Teleport to="body">
      <Transition name="notif-fade">
        <div v-if="showNotifPanel" class="fixed inset-0 z-50" @click.self="showNotifPanel = false">
          <!-- 背景遮罩（透明，仅用于点击关闭） -->
          <div class="absolute inset-0" @click="showNotifPanel = false"></div>
          <!-- 面板主体 -->
          <div
            class="absolute top-14 left-4 lg:left-72 w-[calc(100vw-2rem)] sm:w-96 max-h-[70vh] rounded-xl border overflow-hidden flex flex-col"
            :class="isDark
              ? 'bg-gray-800 border-gray-700 shadow-2xl shadow-black/40'
              : 'bg-white border-gray-200 shadow-2xl shadow-gray-300/50'"
          >
            <!-- 头部 -->
            <div
              class="flex items-center justify-between px-4 py-3 border-b"
              :class="isDark ? 'border-gray-700' : 'border-gray-200'"
            >
              <h3 class="text-sm font-semibold" :class="isDark ? 'text-gray-100' : 'text-gray-900'">
                通知
                <span v-if="unreadCount > 0" class="ml-2 px-2 py-0.5 text-xs rounded-full bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400">
                  {{ unreadCount }} 未读
                </span>
              </h3>
              <button
                v-if="notifications.length > 0"
                @click="handleReadAll"
                class="text-xs text-primary-500 hover:text-primary-600 dark:hover:text-primary-400 transition-colors"
              >
                全部已读
              </button>
            </div>
            <!-- 通知列表 -->
            <div class="flex-1 overflow-y-auto custom-scroll">
              <div v-if="notifLoading" class="p-8 text-center">
                <div class="w-6 h-6 border-2 border-primary-500 border-t-transparent rounded-full animate-spin mx-auto"></div>
                <p class="text-xs mt-2" :class="isDark ? 'text-gray-400' : 'text-gray-500'">加载中...</p>
              </div>
              <div v-else-if="notifications.length === 0" class="p-8 text-center">
                <svg class="w-12 h-12 mx-auto mb-3" :class="isDark ? 'text-gray-600' : 'text-gray-300'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"/>
                </svg>
                <p class="text-sm" :class="isDark ? 'text-gray-400' : 'text-gray-500'">暂无通知</p>
              </div>
              <div v-else>
                <button
                  v-for="notif in notifications"
                  :key="notif.id"
                  @click="handleNotifClick(notif)"
                  class="w-full text-left px-4 py-3 border-b transition-colors"
                  :class="[
                    isDark ? 'border-gray-700/50' : 'border-gray-100',
                    notif.read_at
                      ? (isDark ? 'bg-gray-800 hover:bg-gray-700/50' : 'bg-white hover:bg-gray-50')
                      : (isDark ? 'bg-indigo-900/20 hover:bg-indigo-900/30' : 'bg-indigo-50/50 hover:bg-indigo-50')
                  ]"
                >
                  <div class="flex items-start gap-3">
                    <!-- 类型图标 -->
                    <span
                      class="w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5"
                      :class="notifIconClass(notif.type)"
                    >
                      <span v-html="notifIcon(notif.type)"></span>
                    </span>
                    <div class="flex-1 min-w-0">
                      <div class="flex items-center gap-2">
                        <p
                          class="text-sm font-medium truncate"
                          :class="notif.read_at
                            ? (isDark ? 'text-gray-400' : 'text-gray-600')
                            : (isDark ? 'text-gray-100' : 'text-gray-900')"
                        >
                          {{ notif.title }}
                        </p>
                        <span
                          v-if="!notif.read_at"
                          class="w-2 h-2 rounded-full bg-primary-500 flex-shrink-0"
                        ></span>
                      </div>
                      <p
                        class="text-xs mt-0.5 line-clamp-2"
                        :class="isDark ? 'text-gray-500' : 'text-gray-500'"
                      >
                        {{ notif.body }}
                      </p>
                      <p
                        class="text-xs mt-1"
                        :class="isDark ? 'text-gray-600' : 'text-gray-400'"
                      >
                        {{ formatNotifTime(notif.created_at) }}
                      </p>
                    </div>
                  </div>
                </button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 全局搜索模态框 -->
    <GlobalSearch ref="globalSearchRef" />

    <!-- Onboarding 引导浮层 -->
    <OnboardingOverlay ref="onboardingRef" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useDarkMode } from '../composables/useDarkMode'
import { useOnboarding } from '../composables/useOnboarding'
import {
  getNotifications,
  readNotification,
  readAllNotifications
} from '../api/client'
import GlobalSearch from './GlobalSearch.vue'
import OnboardingOverlay from './OnboardingOverlay.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const { isDark, toggleDarkMode } = useDarkMode()
const { startOnboarding, resetOnboarding, isActive: onboardingActive } = useOnboarding()
const sidebarOpen = ref(false)
const globalSearchRef = ref(null)
const onboardingRef = ref(null)

// ========== 通知系统 ==========
const notifications = ref([])
const unreadCount = ref(0)
const showNotifPanel = ref(false)
const notifLoading = ref(false)
let notifPollTimer = null

async function loadUnreadCount() {
  try {
    const res = await getNotifications()
    unreadCount.value = res.data.unread_count || 0
  } catch {}
}

async function toggleNotifPanel() {
  showNotifPanel.value = !showNotifPanel.value
  if (showNotifPanel.value) {
    await loadNotifications()
  }
}

async function loadNotifications() {
  notifLoading.value = true
  try {
    const res = await getNotifications()
    notifications.value = res.data.data || []
    unreadCount.value = res.data.unread_count || 0
  } catch {
    notifications.value = []
  } finally {
    notifLoading.value = false
  }
}

async function handleNotifClick(notif) {
  if (!notif.read_at) {
    try {
      await readNotification(notif.id)
      notif.read_at = new Date().toISOString()
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    } catch {}
  }
  showNotifPanel.value = false
  if (notif.action_url) {
    router.push(notif.action_url)
  }
}

async function handleReadAll() {
  try {
    await readAllNotifications()
    notifications.value.forEach(n => { n.read_at = n.read_at || new Date().toISOString() })
    unreadCount.value = 0
  } catch {}
}

function notifIcon(type) {
  const icons = {
    review_reminder: '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>',
    analysis_complete: '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>',
    achievement_unlocked: '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z"/></svg>'
  }
  return icons[type] || icons.review_reminder
}

function notifIconClass(type) {
  const classes = {
    review_reminder: 'bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400',
    analysis_complete: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
    achievement_unlocked: 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400'
  }
  return classes[type] || classes.review_reminder
}

function formatNotifTime(dateStr) {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diff = Math.floor((now - date) / 1000)
  if (diff < 60) return '刚刚'
  if (diff < 3600) return Math.floor(diff / 60) + ' 分钟前'
  if (diff < 86400) return Math.floor(diff / 3600) + ' 小时前'
  return Math.floor(diff / 86400) + ' 天前'
}

onMounted(() => {
  // 首次登录自动启动引导
  setTimeout(() => {
    if (startOnboarding()) {
      onboardingRef.value?.updatePosition()
    }
  }, 800) // 等页面渲染完毕后再启动

  // 初始化通知：加载未读计数 + 60 秒轮询
  loadUnreadCount()
  notifPollTimer = setInterval(loadUnreadCount, 60000)

  // Esc 关闭通知面板
  document.addEventListener('keydown', handleNotifEsc)
})

onUnmounted(() => {
  if (notifPollTimer) clearInterval(notifPollTimer)
  document.removeEventListener('keydown', handleNotifEsc)
})

function handleNotifEsc(e) {
  if (e.key === 'Escape' && showNotifPanel.value) {
    showNotifPanel.value = false
  }
}

function openSearch() {
  globalSearchRef.value?.openSearch()
}

const navItems = [
  { path: '/', label: '仪表盘', onboardingId: '', icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 6a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zm10 0a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z"/></svg>' },
  { path: '/upload', label: '上传分析', onboardingId: 'upload', icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM6.293 6.707a1 1 0 010-1.414l3-3a1 1 0 011.414 0l3 3a1 1 0 01-1.414 1.414L11 5.414V13a1 1 0 11-2 0V5.414L7.707 6.707a1 1 0 01-1.414 0z" clip-rule="evenodd"/></svg>' },
  { path: '/cards', label: '知识卡片', onboardingId: 'cards', icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"/></svg>' },
  { path: '/study', label: '学习模式', onboardingId: '', icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path d="M10.394 2.08a1 1 0 00-.788 0l-7 3a1 1 0 000 1.84L5.25 8.051a.999.999 0 01.356-.257l4-1.714a1 1 0 11.788 1.838L7.667 9.088l1.94.831a1 1 0 00.787 0l7-3a1 1 0 000-1.838l-7-3zM3.31 9.397L5 10.12v4.102a8.969 8.969 0 00-1.05-.174 1 1 0 01-.89-.89 11.115 11.115 0 01.25-3.762zM9.3 16.573A9.026 9.026 0 007 14.935v-3.957l1.818.78a3 3 0 002.364 0l5.508-2.361a11.026 11.026 0 01.25 3.762 1 1 0 01-.89.89 8.968 8.968 0 00-5.35 2.524 1 1 0 01-1.4 0zM6 18a1 1 0 001-1v-2.065a8.935 8.935 0 00-2-.712V17a1 1 0 001 1z"/></svg>' },
  { path: '/quiz', label: '练习场', onboardingId: '', icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z"/><path fill-rule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clip-rule="evenodd"/></svg>' },
  { path: '/mistakes', label: '错题本', onboardingId: '', icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M4 2a2 2 0 00-2 2v11a3 3 0 003 3h10a3 3 0 003-3V4a2 2 0 00-2-2H4zm1 3a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1zm0 4a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1zm1 3a1 1 0 100 2h4a1 1 0 100-2H6z" clip-rule="evenodd"/></svg>' },
  { path: '/graph', label: '知识图谱', onboardingId: '', icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path d="M13 7a3 3 0 11-6 0 3 3 0 016 0zM3 15a6 6 0 0112 0v1H3v-1z"/><path d="M16 7a1 1 0 10-2 0v1h-1a1 1 0 100 2h1v1a1 1 0 102 0v-1h1a1 1 0 100-2h-1V7z"/></svg>' },
  { path: '/learning-path', label: '学习路径', onboardingId: '', icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M3 4a1 1 0 011-1h4a1 1 0 010 2H6.414l2.293 2.293a1 1 0 01-1.414 1.414L5 6.414V8a1 1 0 01-2 0V4zm9 1a1 1 0 110-2h4a1 1 0 011 1v4a1 1 0 11-2 0V6.414l-2.293 2.293a1 1 0 11-1.414-1.414L13.586 5H12zm-9 7a1 1 0 112 0v1.586l2.293-2.293a1 1 0 011.414 1.414L6.414 15H8a1 1 0 110 2H4a1 1 0 01-1-1v-4zm13-1a1 1 0 10-2 0v1.586l-2.293-2.293a1 1 0 00-1.414 1.414L13.586 15H12a1 1 0 100 2h4a1 1 0 001-1v-4z" clip-rule="evenodd"/></svg>' },
  { path: '/debate', label: 'Agent 辩论', onboardingId: '', icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path d="M2 5a2 2 0 012-2h7a2 2 0 012 2v4a2 2 0 01-2 2H9l-3 3v-3H4a2 2 0 01-2-2V5z"/><path d="M15 7v2a4 4 0 01-4 4H9.828l-1.703 1.703A1.998 1.998 0 009.828 15H11l3 3v-3h1a2 2 0 002-2V9a2 2 0 00-2-2h-0z"/></svg>' },
  { path: '/chat', label: 'AI 对话', onboardingId: 'chat', icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M18 10c0 3.866-3.582 7-8 7a8.841 8.841 0 01-4.083-.98L2 17l1.338-3.123C2.493 12.767 2 11.434 2 10c0-3.866 3.582-7 8-7s8 3.134 8 7zM7 9H5v2h2V9zm8 0h-2v2h2V9zm-4 0H9v2h2V9z" clip-rule="evenodd"/></svg>' },
  { path: '/api-docs', label: 'API 文档', onboardingId: '', icon: '<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z" clip-rule="evenodd"/></svg>' },
]

const currentPageLabel = computed(() => {
  const item = navItems.find(n => isActive(n.path))
  return item?.label || 'StudyForge Pro'
})

function isActive(path) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
/* 通知面板过渡动画 */
.notif-fade-enter-active,
.notif-fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.notif-fade-enter-from,
.notif-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
.notif-fade-enter-to,
.notif-fade-leave-from {
  opacity: 1;
  transform: translateY(0);
}

/* 通知列表自定义滚动条 */
:deep(.custom-scroll)::-webkit-scrollbar {
  width: 4px;
}
:deep(.custom-scroll)::-webkit-scrollbar-track {
  background: transparent;
}
:deep(.custom-scroll)::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.4);
  border-radius: 4px;
}
:deep(.custom-scroll)::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.6);
}
</style>
