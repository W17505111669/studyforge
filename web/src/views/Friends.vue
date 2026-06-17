<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-4xl mx-auto">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">好友</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">与好友一起学习，互相激励</p>
      </div>
      <button
        class="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors text-sm font-medium"
        @click="activeTab = 'add'"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
        </svg>
        添加好友
      </button>
    </div>

    <!-- 标签切换 -->
    <div class="flex gap-1 mb-6 bg-gray-100 dark:bg-gray-800 rounded-lg p-1">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="flex-1 px-4 py-2 rounded-md text-sm font-medium transition-all duration-200"
        :class="
          activeTab === tab.key
            ? 'bg-white dark:bg-gray-700 text-primary-600 dark:text-primary-400 shadow-sm'
            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
        "
        @click="activeTab = tab.key"
      >
        {{ tab.label }}
        <span
          v-if="tab.key === 'requests' && pendingRequests.length > 0"
          class="ml-1.5 inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 rounded-full bg-blue-500 text-white text-xs font-bold"
        >
          {{ pendingRequests.length }}
        </span>
        <span
          v-if="tab.key === 'friends' && friends.length > 0"
          class="ml-1.5 text-xs text-gray-400 dark:text-gray-500"
        >
          {{ friends.length }}
        </span>
      </button>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="space-y-4">
      <div
        v-for="i in 3"
        :key="i"
        class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 animate-pulse"
      >
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-gray-200 dark:bg-gray-700 rounded-full"></div>
          <div class="flex-1">
            <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-24 mb-2"></div>
            <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-40"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- ========== 好友列表 ========== -->
    <div v-else-if="activeTab === 'friends'">
      <div v-if="friends.length === 0" class="text-center py-16">
        <svg
          class="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
          />
        </svg>
        <h3 class="text-lg font-medium text-gray-600 dark:text-gray-400 mb-2">还没有好友</h3>
        <p class="text-sm text-gray-400 dark:text-gray-500 mb-4">搜索用户名添加好友，一起学习更有动力！</p>
        <button
          class="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors text-sm"
          @click="activeTab = 'add'"
        >
          添加好友
        </button>
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="friend in friends"
          :key="friend.friend_id"
          class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 hover:shadow-md dark:hover:shadow-gray-900/50 transition-shadow"
        >
          <div class="flex items-center gap-4">
            <!-- 头像 -->
            <div
              class="w-12 h-12 rounded-full flex items-center justify-center text-white font-bold text-lg shrink-0"
              :style="{ backgroundColor: avatarColor(friend.friend_username) }"
            >
              {{ (friend.friend_nickname || friend.friend_username || '?').charAt(0).toUpperCase() }}
            </div>

            <!-- 信息 -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <h3 class="font-semibold text-gray-900 dark:text-white truncate">
                  {{ friend.friend_nickname || friend.friend_username }}
                </h3>
                <span class="text-xs text-gray-400 dark:text-gray-500">@{{ friend.friend_username }}</span>
              </div>
              <div class="flex items-center gap-3 mt-1 text-xs text-gray-500 dark:text-gray-400">
                <span
                  class="flex items-center gap-1"
                  :class="friend.weekly_cards > 0 ? 'text-emerald-600 dark:text-emerald-400' : ''"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                    />
                  </svg>
                  复习 {{ friend.weekly_cards }}
                </span>
                <span
                  class="flex items-center gap-1"
                  :class="friend.weekly_quizzes > 0 ? 'text-blue-600 dark:text-blue-400' : ''"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907"
                    />
                  </svg>
                  做题 {{ friend.weekly_quizzes }}
                </span>
                <span
                  class="flex items-center gap-1"
                  :class="friend.weekly_streak > 0 ? 'text-orange-500 dark:text-orange-400' : ''"
                >
                  <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
                    <path
                      fill-rule="evenodd"
                      d="M12.395 2.553a1 1 0 00-1.45-.385c-.345.23-.614.558-.822.88-.214.33-.403.713-.57 1.116-.334.804-.614 1.768-.84 2.734a31.365 31.365 0 00-.613 3.58 2.64 2.64 0 01-.945-1.067c-.328-.68-.398-1.534-.398-2.654A1 1 0 005.05 6.05 6.981 6.981 0 003 11a7 7 0 1011.95-4.95c-.592-.591-.98-.985-1.348-1.467-.363-.476-.724-1.063-1.207-2.03zM12.12 15.12A3 3 0 017 13s.879.5 2.5.5c0-1 .5-2 1-3 .5 1 1 1 1 3a3 3 0 01.62 1.62z"
                      clip-rule="evenodd"
                    />
                  </svg>
                  {{ friend.weekly_streak }}天
                </span>
              </div>
            </div>

            <!-- 操作 -->
            <button
              class="text-gray-400 hover:text-red-500 dark:text-gray-500 dark:hover:text-red-400 transition-colors p-2 shrink-0"
              title="删除好友"
              @click="handleRemoveFriend(friend.friend_id, friend.friend_username)"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13 7a4 4 0 11-8 0 4 4 0 018 0zM9 14a6 6 0 00-6 6v1h12v-1a6 6 0 00-6-6zM21 12h-6"
                />
              </svg>
            </button>
          </div>

          <!-- 本周学习摘要进度条 -->
          <div class="mt-3 flex gap-2">
            <div class="flex-1 h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
              <div
                class="h-full bg-emerald-500 rounded-full transition-all duration-500"
                :style="{ width: Math.min(friend.weekly_cards * 5, 100) + '%' }"
              ></div>
            </div>
            <div class="flex-1 h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
              <div
                class="h-full bg-blue-500 rounded-full transition-all duration-500"
                :style="{ width: Math.min(friend.weekly_quizzes * 10, 100) + '%' }"
              ></div>
            </div>
            <div class="flex-1 h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
              <div
                class="h-full bg-orange-500 rounded-full transition-all duration-500"
                :style="{ width: Math.min(friend.weekly_streak * 14.3, 100) + '%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ========== 待处理请求 ========== -->
    <div v-else-if="activeTab === 'requests'">
      <div v-if="pendingRequests.length === 0" class="text-center py-16">
        <svg
          class="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
          />
        </svg>
        <h3 class="text-lg font-medium text-gray-600 dark:text-gray-400 mb-2">没有待处理的好友请求</h3>
        <p class="text-sm text-gray-400 dark:text-gray-500">当有人添加你为好友时，请求会显示在这里</p>
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="req in pendingRequests"
          :key="req.id"
          class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4"
        >
          <div class="flex items-center gap-4">
            <div
              class="w-12 h-12 rounded-full flex items-center justify-center text-white font-bold text-lg shrink-0"
              :style="{ backgroundColor: avatarColor(req.username) }"
            >
              {{ (req.nickname || req.username || '?').charAt(0).toUpperCase() }}
            </div>
            <div class="flex-1 min-w-0">
              <h3 class="font-semibold text-gray-900 dark:text-white truncate">
                {{ req.nickname || req.username }}
              </h3>
              <p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">
                {{ formatTime(req.created_at) }}
              </p>
            </div>
            <div class="flex gap-2 shrink-0">
              <button
                :disabled="req._loading"
                class="px-3 py-1.5 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors text-sm font-medium disabled:opacity-50"
                @click="handleAccept(req)"
              >
                <span v-if="req._loading" class="inline-flex items-center gap-1">
                  <svg class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                </span>
                <span v-else>接受</span>
              </button>
              <button
                :disabled="req._loading"
                class="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors text-sm font-medium disabled:opacity-50"
                @click="handleReject(req)"
              >
                拒绝
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ========== 添加好友 ========== -->
    <div v-else-if="activeTab === 'add'">
      <!-- 搜索框 -->
      <div class="mb-6">
        <div class="relative">
          <svg
            class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索用户名或昵称..."
            class="w-full pl-10 pr-4 py-3 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all text-sm"
            @input="debounceSearch"
          />
          <button
            v-if="searchQuery"
            class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
            @click="clearSearch()"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <!-- 搜索结果 -->
      <div v-if="searchResults.length > 0" class="space-y-3">
        <div
          v-for="user in searchResults"
          :key="user.id"
          class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4"
        >
          <div class="flex items-center gap-4">
            <div
              class="w-12 h-12 rounded-full flex items-center justify-center text-white font-bold text-lg shrink-0"
              :style="{ backgroundColor: avatarColor(user.username) }"
            >
              {{ (user.nickname || user.username || '?').charAt(0).toUpperCase() }}
            </div>
            <div class="flex-1 min-w-0">
              <h3 class="font-semibold text-gray-900 dark:text-white truncate">
                {{ user.nickname || user.username }}
              </h3>
              <p class="text-xs text-gray-400 dark:text-gray-500">@{{ user.username }}</p>
            </div>
            <div class="shrink-0">
              <!-- 已是好友 -->
              <span
                v-if="user.friend_status === 'accepted'"
                class="px-3 py-1.5 bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400 rounded-lg text-sm font-medium"
              >
                已是好友
              </span>
              <!-- 待处理 -->
              <span
                v-else-if="user.friend_status === 'pending'"
                class="px-3 py-1.5 bg-amber-50 dark:bg-amber-900/20 text-amber-600 dark:text-amber-400 rounded-lg text-sm font-medium"
              >
                请求中
              </span>
              <!-- 可添加 -->
              <button
                v-else
                :disabled="user._loading"
                class="px-3 py-1.5 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors text-sm font-medium disabled:opacity-50 flex items-center gap-1"
                @click="handleAddFriend(user)"
              >
                <svg v-if="user._loading" class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                  />
                </svg>
                添加
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 搜索无结果 -->
      <div v-else-if="searchQuery && !searchLoading" class="text-center py-12">
        <svg
          class="w-12 h-12 mx-auto text-gray-300 dark:text-gray-600 mb-3"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
        <p class="text-gray-500 dark:text-gray-400">没有找到匹配的用户</p>
      </div>

      <!-- 搜索中 -->
      <div v-else-if="searchLoading" class="text-center py-12">
        <svg class="w-8 h-8 mx-auto text-primary-500 animate-spin mb-3" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        <p class="text-sm text-gray-400 dark:text-gray-500">搜索中...</p>
      </div>

      <!-- 初始状态提示 -->
      <div v-else class="text-center py-12">
        <svg
          class="w-12 h-12 mx-auto text-gray-300 dark:text-gray-600 mb-3"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
          />
        </svg>
        <p class="text-gray-500 dark:text-gray-400">输入用户名搜索并添加好友</p>
        <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">支持按用户名或昵称模糊搜索</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useToast, useConfirm } from '../composables/useToast'
import {
  listFriends,
  getFriendRequests,
  sendFriendRequest,
  acceptFriendRequest,
  rejectFriendRequest,
  removeFriend,
  searchUsers
} from '../api/client'

const { toast } = useToast()
const { confirm } = useConfirm()

const loading = ref(true)
const activeTab = ref('friends')

const tabs = [
  { key: 'friends', label: '好友列表' },
  { key: 'requests', label: '好友请求' },
  { key: 'add', label: '添加好友' }
]

// ========== 好友列表 ==========
const friends = ref([])

async function loadFriends() {
  loading.value = true
  try {
    const res = await listFriends()
    friends.value = res.data.friends || []
  } catch {
    friends.value = []
  } finally {
    loading.value = false
  }
}

async function handleRemoveFriend(friendId, username) {
  const ok = await confirm(`确定要删除好友 @${username} 吗？`)
  if (!ok) return
  try {
    await removeFriend(friendId)
    toast.success('已删除好友')
    friends.value = friends.value.filter((f) => f.friend_id !== friendId)
  } catch {
    toast.error('删除好友失败')
  }
}

// ========== 待处理请求 ==========
const pendingRequests = ref([])

async function loadPendingRequests() {
  try {
    const res = await getFriendRequests()
    pendingRequests.value = (res.data.requests || []).map((r) => ({ ...r, _loading: false }))
  } catch {
    pendingRequests.value = []
  }
}

async function handleAccept(req) {
  req._loading = true
  try {
    await acceptFriendRequest(req.id)
    toast.success(`已添加 ${req.nickname || req.username} 为好友`)
    pendingRequests.value = pendingRequests.value.filter((r) => r.id !== req.id)
    // 刷新好友列表
    loadFriends()
  } catch {
    toast.error('接受请求失败')
    req._loading = false
  }
}

async function handleReject(req) {
  req._loading = true
  try {
    await rejectFriendRequest(req.id)
    toast.success('已拒绝好友请求')
    pendingRequests.value = pendingRequests.value.filter((r) => r.id !== req.id)
  } catch {
    toast.error('拒绝请求失败')
    req._loading = false
  }
}

// ========== 搜索添加好友 ==========
const searchQuery = ref('')
const searchResults = ref([])
const searchLoading = ref(false)
let searchTimer = null

function debounceSearch() {
  clearTimeout(searchTimer)
  const q = searchQuery.value.trim()
  if (!q) {
    searchResults.value = []
    searchLoading.value = false
    return
  }
  searchLoading.value = true
  searchTimer = setTimeout(async () => {
    try {
      const res = await searchUsers(q)
      searchResults.value = (res.data.users || []).map((u) => ({ ...u, _loading: false }))
    } catch {
      searchResults.value = []
    } finally {
      searchLoading.value = false
    }
  }, 400)
}

function clearSearch() {
  searchQuery.value = ''
  searchResults.value = []
}

async function handleAddFriend(user) {
  user._loading = true
  try {
    await sendFriendRequest(user.username)
    toast.success(`已向 @${user.username} 发送好友请求`)
    user.friend_status = 'pending'
  } catch (err) {
    const msg = err.response?.data?.error || '发送请求失败'
    toast.error(msg)
  } finally {
    user._loading = false
  }
}

// ========== 辅助函数 ==========
const avatarColors = [
  '#6366f1',
  '#8b5cf6',
  '#a855f7',
  '#d946ef',
  '#ec4899',
  '#f43f5e',
  '#ef4444',
  '#f97316',
  '#eab308',
  '#84cc16',
  '#22c55e',
  '#14b8a6',
  '#06b6d4',
  '#3b82f6',
  '#6366f1'
]

function avatarColor(name) {
  if (!name) return avatarColors[0]
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return avatarColors[Math.abs(hash) % avatarColors.length]
}

function formatTime(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diffMs = now - d
  const diffMin = Math.floor(diffMs / 60000)
  if (diffMin < 1) return '刚刚'
  if (diffMin < 60) return `${diffMin} 分钟前`
  const diffHour = Math.floor(diffMin / 60)
  if (diffHour < 24) return `${diffHour} 小时前`
  const diffDay = Math.floor(diffHour / 24)
  if (diffDay < 7) return `${diffDay} 天前`
  return d.toLocaleDateString('zh-CN')
}

// ========== 初始化 ==========
onMounted(async () => {
  await Promise.all([loadFriends(), loadPendingRequests()])
  loading.value = false
})
</script>
