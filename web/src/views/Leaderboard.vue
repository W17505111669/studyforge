<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-5xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
      <div>
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white">学习排行榜</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">与他人比拼学习积分，共同进步</p>
      </div>
      <div class="flex items-center gap-2">
        <button
          :class="
            period === 'week'
              ? 'bg-indigo-600 text-white shadow-sm'
              : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
          "
          class="px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          @click="period = 'week'"
        >
          本周
        </button>
        <button
          :class="
            period === 'month'
              ? 'bg-indigo-600 text-white shadow-sm'
              : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
          "
          class="px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          @click="period = 'month'"
        >
          本月
        </button>
      </div>
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading" class="space-y-6">
      <div class="flex justify-center gap-4 sm:gap-8">
        <div
          v-for="i in 3"
          :key="i"
          class="skeleton-box rounded-xl"
          :style="{ width: i === 2 ? '140px' : '120px', height: i === 2 ? '160px' : '130px' }"
        ></div>
      </div>
      <div class="space-y-3">
        <div v-for="i in 5" :key="i" class="skeleton-box h-14 rounded-lg"></div>
      </div>
    </div>

    <template v-else>
      <!-- Top 3 podium -->
      <div v-if="users.length >= 3" class="flex justify-center items-end gap-3 sm:gap-6 mb-8">
        <!-- 2nd place -->
        <div class="text-center">
          <div
            class="w-14 h-14 sm:w-16 sm:h-16 rounded-full mx-auto mb-2 flex items-center justify-center text-white text-lg font-bold shadow-md"
            :style="{ background: avatarColors[1 % avatarColors.length] }"
          >
            {{ getInitial(users[1]?.username, users[1]?.nickname) }}
          </div>
          <div
            class="text-xs sm:text-sm font-semibold text-gray-800 dark:text-gray-200 truncate max-w-[80px] sm:max-w-[120px] mx-auto"
          >
            {{ users[1]?.nickname || users[1]?.username }}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ users[1]?.score }} 分</div>
          <div
            class="bg-gray-200 dark:bg-gray-700 rounded-t-lg mt-2 flex items-end justify-center"
            style="height: 70px"
          >
            <span class="text-2xl font-bold text-gray-400 dark:text-gray-500">2</span>
          </div>
        </div>
        <!-- 1st place -->
        <div class="text-center">
          <div class="relative">
            <svg class="w-8 h-8 sm:w-10 sm:h-10 text-amber-400 mx-auto -mb-2" fill="currentColor" viewBox="0 0 24 24">
              <path d="M5 16L3 5l5.5 5L12 4l3.5 6L21 5l-2 11H5zm14 3c0 .6-.4 1-1 1H6c-.6 0-1-.4-1-1v-1h14v1z" />
            </svg>
          </div>
          <div
            class="w-16 h-16 sm:w-20 sm:h-20 rounded-full mx-auto mb-2 flex items-center justify-center text-white text-xl font-bold shadow-lg ring-4 ring-amber-300 dark:ring-amber-500"
            :style="{ background: avatarColors[0 % avatarColors.length] }"
          >
            {{ getInitial(users[0]?.username, users[0]?.nickname) }}
          </div>
          <div
            class="text-sm sm:text-base font-bold text-gray-900 dark:text-white truncate max-w-[100px] sm:max-w-[140px] mx-auto"
          >
            {{ users[0]?.nickname || users[0]?.username }}
          </div>
          <div class="text-sm text-amber-600 dark:text-amber-400 font-semibold mt-0.5">{{ users[0]?.score }} 分</div>
          <div
            class="bg-gradient-to-t from-amber-100 to-amber-50 dark:from-amber-900/40 dark:to-amber-800/20 rounded-t-lg mt-2 flex items-end justify-center"
            style="height: 100px"
          >
            <span class="text-3xl font-bold text-amber-500 dark:text-amber-400">1</span>
          </div>
        </div>
        <!-- 3rd place -->
        <div class="text-center">
          <div
            class="w-14 h-14 sm:w-16 sm:h-16 rounded-full mx-auto mb-2 flex items-center justify-center text-white text-lg font-bold shadow-md"
            :style="{ background: avatarColors[2 % avatarColors.length] }"
          >
            {{ getInitial(users[2]?.username, users[2]?.nickname) }}
          </div>
          <div
            class="text-xs sm:text-sm font-semibold text-gray-800 dark:text-gray-200 truncate max-w-[80px] sm:max-w-[120px] mx-auto"
          >
            {{ users[2]?.nickname || users[2]?.username }}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ users[2]?.score }} 分</div>
          <div
            class="bg-orange-100 dark:bg-orange-900/30 rounded-t-lg mt-2 flex items-end justify-center"
            style="height: 50px"
          >
            <span class="text-2xl font-bold text-orange-400 dark:text-orange-500">3</span>
          </div>
        </div>
      </div>

      <!-- Current user stats card -->
      <div
        v-if="currentUser"
        class="bg-gradient-to-r from-indigo-500 to-purple-600 rounded-xl p-4 sm:p-5 mb-6 text-white shadow-sm"
      >
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <div class="text-sm text-indigo-100 mb-1">
              {{ isMe(users[0]) ? '🎉 你是排行榜第一名！' : '我的排名' }}
            </div>
            <div class="flex items-baseline gap-3">
              <span class="text-3xl font-bold">{{ currentUser.rank }}</span>
              <span class="text-lg text-indigo-100">/ {{ totalUsers }} 人</span>
            </div>
          </div>
          <div class="flex items-center gap-4">
            <div class="text-center">
              <div class="text-2xl font-bold">{{ currentUser.score }}</div>
              <div class="text-xs text-indigo-200">总积分</div>
            </div>
            <div v-if="nextRankScore > currentUser.score" class="text-center border-l border-indigo-300/40 pl-4">
              <div class="text-lg font-semibold text-indigo-100">+{{ nextRankScore - currentUser.score }}</div>
              <div class="text-xs text-indigo-200">距上一名</div>
            </div>
          </div>
        </div>
        <!-- Score breakdown pills -->
        <div class="flex flex-wrap gap-2 mt-3 pt-3 border-t border-indigo-300/30">
          <span v-if="currentUser.cards_reviewed > 0" class="bg-white/20 rounded-full px-2.5 py-0.5 text-xs">
            卡片 {{ currentUser.cards_reviewed }}×1
          </span>
          <span v-if="currentUser.quizzes_completed > 0" class="bg-white/20 rounded-full px-2.5 py-0.5 text-xs">
            练习 {{ currentUser.quizzes_completed }}×3
          </span>
          <span v-if="currentUser.correct_answers > 0" class="bg-white/20 rounded-full px-2.5 py-0.5 text-xs">
            答对 {{ currentUser.correct_answers }}×1
          </span>
          <span v-if="currentUser.materials_uploaded > 0" class="bg-white/20 rounded-full px-2.5 py-0.5 text-xs">
            材料 {{ currentUser.materials_uploaded }}×8
          </span>
          <span v-if="currentUser.pomodoros_completed > 0" class="bg-white/20 rounded-full px-2.5 py-0.5 text-xs">
            番茄 {{ currentUser.pomodoros_completed }}×5
          </span>
          <span v-if="currentUser.notes_created > 0" class="bg-white/20 rounded-full px-2.5 py-0.5 text-xs">
            笔记 {{ currentUser.notes_created }}×2
          </span>
        </div>
      </div>

      <!-- Scoring rules -->
      <div class="bg-white dark:bg-gray-800 rounded-xl p-4 mb-6 border border-gray-200 dark:border-gray-700">
        <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">积分规则</h3>
        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-2">
          <div class="flex items-center gap-1.5 text-xs text-gray-600 dark:text-gray-400">
            <span class="w-2 h-2 rounded-full bg-green-500"></span>
            卡片复习 1分
          </div>
          <div class="flex items-center gap-1.5 text-xs text-gray-600 dark:text-gray-400">
            <span class="w-2 h-2 rounded-full bg-blue-500"></span>
            练习题作答 3分
          </div>
          <div class="flex items-center gap-1.5 text-xs text-gray-600 dark:text-gray-400">
            <span class="w-2 h-2 rounded-full bg-cyan-500"></span>
            答对奖励 1分
          </div>
          <div class="flex items-center gap-1.5 text-xs text-gray-600 dark:text-gray-400">
            <span class="w-2 h-2 rounded-full bg-purple-500"></span>
            上传材料 8分
          </div>
          <div class="flex items-center gap-1.5 text-xs text-gray-600 dark:text-gray-400">
            <span class="w-2 h-2 rounded-full bg-amber-500"></span>
            番茄专注 5分
          </div>
          <div class="flex items-center gap-1.5 text-xs text-gray-600 dark:text-gray-400">
            <span class="w-2 h-2 rounded-full bg-pink-500"></span>
            创建笔记 2分
          </div>
        </div>
      </div>

      <!-- Full ranking table -->
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
        <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
          <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">完整排名</h3>
        </div>
        <div class="divide-y divide-gray-100 dark:divide-gray-700/50 max-h-[500px] overflow-y-auto custom-scroll">
          <div
            v-for="(user, idx) in users"
            :key="user.user_id"
            :class="isMe(user) ? 'bg-indigo-50 dark:bg-indigo-900/20' : ''"
            class="flex items-center px-4 py-3 gap-3 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
          >
            <!-- Rank -->
            <div class="w-8 text-center shrink-0">
              <span v-if="idx < 3" class="text-lg" :class="idx === 0 ? '🥇' : idx === 1 ? '🥈' : '🥉'">
                {{ idx === 0 ? '🥇' : idx === 1 ? '🥈' : '🥉' }}
              </span>
              <span v-else class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ idx + 1 }}</span>
            </div>
            <!-- Avatar -->
            <div
              class="w-9 h-9 rounded-full flex items-center justify-center text-white text-sm font-bold shrink-0"
              :style="{ background: avatarColors[idx % avatarColors.length] }"
            >
              {{ getInitial(user.username, user.nickname) }}
            </div>
            <!-- Name -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium text-gray-900 dark:text-white truncate">
                  {{ user.nickname || user.username }}
                </span>
                <span
                  v-if="isMe(user)"
                  class="bg-indigo-100 dark:bg-indigo-900/50 text-indigo-600 dark:text-indigo-400 text-xs px-1.5 py-0.5 rounded font-medium"
                >
                  我
                </span>
              </div>
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 flex flex-wrap gap-x-2">
                <span v-if="user.cards_reviewed">卡片 {{ user.cards_reviewed }}</span>
                <span v-if="user.quizzes_completed">练习 {{ user.quizzes_completed }}</span>
                <span v-if="user.correct_answers">答对 {{ user.correct_answers }}</span>
                <span v-if="user.materials_uploaded">材料 {{ user.materials_uploaded }}</span>
                <span v-if="user.pomodoros_completed">番茄 {{ user.pomodoros_completed }}</span>
                <span v-if="user.notes_created">笔记 {{ user.notes_created }}</span>
              </div>
            </div>
            <!-- Score -->
            <div class="text-right shrink-0">
              <span class="text-lg font-bold text-gray-900 dark:text-white">{{ user.score }}</span>
              <span class="text-xs text-gray-500 dark:text-gray-400 ml-1">分</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty state -->
      <div v-if="users.length === 0" class="text-center py-16">
        <svg
          class="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M16.5 18.75h-9m9 0a3 3 0 013 3h-15a3 3 0 013-3m9 0v-4.5A3.375 3.375 0 0012.75 10.875h-1.5A3.375 3.375 0 007.875 14.25v4.5m4.125-12a2.625 2.625 0 100-5.25 2.625 2.625 0 000 5.25z"
          />
        </svg>
        <h3 class="text-lg font-semibold text-gray-600 dark:text-gray-400 mb-2">暂无排行数据</h3>
        <p class="text-sm text-gray-500 dark:text-gray-500">
          {{ period === 'week' ? '本周还没有人获得积分' : '本月还没有人获得积分' }}
        </p>
        <p class="text-sm text-gray-500 dark:text-gray-500 mt-1">开始学习来获得积分吧！</p>
      </div>

      <!-- Period info -->
      <div class="text-center text-xs text-gray-400 dark:text-gray-500 mt-4">
        {{ periodLabel }}（{{ start }} 至 {{ end }}）
      </div>
    </template>

    <!-- Reload button -->
    <div v-if="!loading && users.length > 0" class="text-center mt-4">
      <button class="text-sm text-indigo-600 dark:text-indigo-400 hover:underline" @click="loadLeaderboard">
        刷新排行
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { getLeaderboard } from '../api/client'

const period = ref('week')
const loading = ref(false)
const users = ref([])
const currentUser = ref(null)
const totalUsers = ref(0)
const nextRankScore = ref(0)
const start = ref('')
const end = ref('')

const avatarColors = [
  '#6366f1',
  '#ec4899',
  '#10b981',
  '#f59e0b',
  '#3b82f6',
  '#8b5cf6',
  '#ef4444',
  '#06b6d4',
  '#84cc16',
  '#f97316',
  '#14b8a6',
  '#a855f7',
  '#0ea5e9',
  '#e11d48',
  '#22c55e'
]

const periodLabel = computed(() => {
  return period.value === 'week' ? '本周排行' : '本月排行'
})

let myUserId = ''
try {
  const auth = JSON.parse(localStorage.getItem('user') || '{}')
  myUserId = auth.id || ''
} catch {
  myUserId = ''
}

function isMe(user) {
  return user && user.user_id === myUserId
}

function getInitial(username, nickname) {
  const name = nickname || username || '?'
  return name.charAt(0).toUpperCase()
}

async function loadLeaderboard() {
  loading.value = true
  try {
    const res = await getLeaderboard(period.value)
    const data = res.data
    users.value = data.users || []
    currentUser.value = data.current_user || null
    totalUsers.value = data.total_users || 0
    nextRankScore.value = data.next_rank_score || 0
    start.value = data.start || ''
    end.value = data.end || ''
  } catch (err) {
    console.error('Failed to load leaderboard:', err)
    users.value = []
    currentUser.value = null
  } finally {
    loading.value = false
  }
}

watch(period, () => {
  loadLeaderboard()
})

onMounted(() => {
  loadLeaderboard()
})
</script>

<style scoped>
.skeleton-box {
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}
:global(.dark) .skeleton-box {
  background: linear-gradient(90deg, #374151 25%, #4b5563 50%, #374151 75%);
  background-size: 200% 100%;
}
@keyframes shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}
.custom-scroll::-webkit-scrollbar {
  width: 4px;
}
.custom-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroll::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.4);
  border-radius: 4px;
}
.custom-scroll::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.6);
}
</style>
