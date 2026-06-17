<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto">
    <!-- 欢迎区域 -->
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">
        {{ greeting }}，{{ auth.user?.username || '同学' }}
      </h1>
      <p class="text-gray-500 dark:text-gray-400 mt-1">今天也要高效学习哦</p>
    </div>

    <!-- 离线缓存数据提示 -->
    <div
      v-if="!loading && fromCache"
      class="mb-4 flex items-center gap-2 px-3 py-2 rounded-lg text-xs font-medium bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-900/20 dark:text-amber-400 dark:border-amber-800/30"
    >
      <svg class="w-3.5 h-3.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
      <span>正在查看缓存数据，联网后将自动更新</span>
    </div>

    <!-- ========== 学习建议横幅 ========== -->
    <div
      v-if="!loading && recommendations.length > 0"
      class="mb-6 bg-gradient-to-r from-indigo-50 via-purple-50 to-pink-50 dark:from-indigo-950/40 dark:via-purple-950/30 dark:to-pink-950/20 rounded-xl border border-indigo-100 dark:border-indigo-800/40 p-4 sm:p-5"
    >
      <div class="flex items-center gap-2 mb-3">
        <span class="text-lg">💡</span>
        <h2 class="text-sm font-semibold text-indigo-700 dark:text-indigo-300">今日学习建议</h2>
        <span class="ml-auto text-xs text-indigo-400 dark:text-indigo-500">{{ recommendations.length }} 条建议</span>
      </div>
      <div class="flex flex-col sm:flex-row gap-2 sm:gap-3">
        <router-link
          v-for="rec in recommendations"
          :key="rec.type + rec.title"
          :to="rec.action_url"
          class="flex items-start gap-3 p-3 rounded-lg bg-white/70 dark:bg-gray-800/60 hover:bg-white dark:hover:bg-gray-800 hover:shadow-sm transition-all border border-transparent hover:border-indigo-200 dark:hover:border-indigo-700/50 group flex-1 min-w-0"
        >
          <span class="text-xl shrink-0 mt-0.5">{{ rec.icon }}</span>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span
                class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors"
              >
                {{ rec.title }}
              </span>
              <span
                v-if="rec.badge"
                class="shrink-0 px-1.5 py-0.5 rounded-full text-[10px] font-bold bg-red-500 text-white leading-none"
              >
                {{ rec.badge }}
              </span>
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 line-clamp-2 leading-relaxed">
              {{ rec.description }}
            </p>
          </div>
          <svg
            class="w-4 h-4 text-gray-300 dark:text-gray-600 shrink-0 mt-1 group-hover:text-indigo-400 transition-colors"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
          </svg>
        </router-link>
      </div>
    </div>

    <!-- ========== 骨架屏状态 ========== -->
    <template v-if="loading">
      <!-- 统计卡片骨架 -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5 mb-8">
        <div
          v-for="i in 4"
          :key="i"
          class="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-100 dark:border-gray-700"
        >
          <div class="flex items-center justify-between mb-3">
            <div class="skeleton skeleton-text w-20 h-4"></div>
            <div class="skeleton skeleton-box w-10 h-10 rounded-lg"></div>
          </div>
          <div class="skeleton skeleton-text w-16 h-8 mb-2"></div>
          <div class="skeleton skeleton-text w-24 h-3"></div>
        </div>
      </div>

      <!-- 图表骨架：第一行 -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-5 mb-5">
        <div
          class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6"
        >
          <div class="skeleton skeleton-text w-32 h-5 mb-4"></div>
          <div class="skeleton skeleton-box w-full rounded-lg" style="height: 280px"></div>
        </div>
        <div
          class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6 lg:col-span-2"
        >
          <div class="skeleton skeleton-text w-32 h-5 mb-4"></div>
          <div class="skeleton skeleton-box w-full rounded-lg" style="height: 280px"></div>
        </div>
      </div>

      <!-- 图表骨架：第二行 -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-5 mb-5">
        <div
          class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6 lg:col-span-2"
        >
          <div class="skeleton skeleton-text w-40 h-5 mb-4"></div>
          <div class="skeleton skeleton-box w-full rounded-lg" style="height: 200px"></div>
        </div>
        <div
          class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6"
        >
          <div class="skeleton skeleton-text w-28 h-5 mb-4"></div>
          <div class="skeleton skeleton-box w-full rounded-lg" style="height: 200px"></div>
          <div class="skeleton skeleton-text w-16 h-6 mt-3 mx-auto"></div>
        </div>
      </div>

      <!-- 材料列表骨架 -->
      <div
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6 mb-5"
      >
        <div class="flex items-center justify-between mb-5">
          <div class="skeleton skeleton-text w-28 h-5"></div>
          <div class="skeleton skeleton-text w-20 h-4"></div>
        </div>
        <div class="space-y-3">
          <div
            v-for="i in 3"
            :key="i"
            class="rounded-lg border border-gray-100 dark:border-gray-700 md:flex md:items-center md:justify-between md:p-4"
          >
            <div class="flex-1 p-4 md:p-0">
              <div class="skeleton skeleton-text w-48 h-4 mb-2"></div>
              <div class="skeleton skeleton-text w-24 h-3"></div>
            </div>
            <div class="hidden md:block skeleton skeleton-box w-16 h-6 rounded-full ml-4"></div>
          </div>
        </div>
      </div>

      <!-- 成就骨架 -->
      <div
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6 mb-5"
      >
        <div class="flex items-center justify-between mb-4">
          <div class="skeleton skeleton-text w-32 h-5"></div>
          <div class="skeleton skeleton-box w-24 h-2 rounded-full"></div>
        </div>
        <div class="flex gap-1.5 mb-4">
          <div v-for="i in 5" :key="i" class="skeleton skeleton-box w-12 h-6 rounded-full"></div>
        </div>
        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
          <div v-for="i in 6" :key="i" class="rounded-xl p-3 text-center border border-gray-100 dark:border-gray-700">
            <div class="skeleton skeleton-box w-8 h-8 rounded-full mx-auto mb-2"></div>
            <div class="skeleton skeleton-text w-16 h-3 mx-auto mb-1"></div>
            <div class="skeleton skeleton-text w-20 h-2 mx-auto"></div>
          </div>
        </div>
      </div>

      <!-- AI 调用概览骨架 -->
      <div
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6"
      >
        <div class="skeleton skeleton-text w-28 h-5 mb-5"></div>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div v-for="i in 4" :key="i" class="text-center p-4 rounded-lg bg-gray-50 dark:bg-gray-700">
            <div class="skeleton skeleton-text w-12 h-6 mx-auto mb-2"></div>
            <div class="skeleton skeleton-text w-16 h-3 mx-auto"></div>
          </div>
        </div>
      </div>
    </template>

    <!-- ========== 正常内容状态 ========== -->
    <template v-else>
      <!-- 统计卡片 -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5 mb-5">
        <div
          v-for="stat in stats"
          :key="stat.label"
          class="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-100 dark:border-gray-700 hover:shadow-md transition-shadow"
        >
          <div class="flex items-center justify-between mb-3">
            <span class="text-sm text-gray-500 dark:text-gray-400">{{ stat.label }}</span>
            <span class="w-10 h-10 rounded-lg flex items-center justify-center text-lg" :class="stat.bgColor">
              {{ stat.icon }}
            </span>
          </div>
          <div class="text-3xl font-bold" :class="stat.color">{{ stat.value }}</div>
          <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">{{ stat.desc }}</p>
        </div>
      </div>

      <!-- 待复习卡片入口 — 可点击跳转 /review -->
      <router-link
        v-if="dueCardCount > 0"
        to="/review"
        class="block rounded-xl p-5 mb-5 border transition-all hover:shadow-md group"
        :class="
          isDark
            ? 'bg-gradient-to-r from-red-900/30 to-orange-900/20 border-gray-700 hover:border-red-700/50'
            : 'bg-gradient-to-r from-red-50 to-orange-50 border-gray-100 hover:border-red-200'
        "
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div
              class="w-12 h-12 rounded-xl flex items-center justify-center"
              :class="isDark ? 'bg-red-500/20' : 'bg-red-100'"
            >
              <svg
                class="w-6 h-6"
                :class="isDark ? 'text-red-400' : 'text-red-600'"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" />
                <path
                  fill-rule="evenodd"
                  d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm9.707 5.707a1 1 0 00-1.414-1.414L9 12.586l-1.293-1.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                  clip-rule="evenodd"
                />
              </svg>
            </div>
            <div>
              <p class="text-xs font-medium mb-1" :class="isDark ? 'text-gray-400' : 'text-gray-500'">待复习卡片</p>
              <div class="flex items-center gap-3">
                <span class="text-2xl font-bold" :class="isDark ? 'text-red-400' : 'text-red-600'">
                  {{ dueCardCount }}
                </span>
                <span class="text-sm" :class="isDark ? 'text-gray-400' : 'text-gray-500'">张卡片等待复习</span>
              </div>
            </div>
          </div>
          <span
            class="text-sm font-medium px-4 py-2 rounded-lg transition-colors"
            :class="isDark ? 'text-red-400 group-hover:bg-red-500/10' : 'text-red-600 group-hover:bg-red-50'"
          >
            开始复习 →
          </span>
        </div>
      </router-link>

      <!-- 今日专注统计 -->
      <div
        class="rounded-xl p-5 mb-5 border flex items-center justify-between"
        :class="
          isDark
            ? 'bg-gradient-to-r from-indigo-900/30 to-purple-900/20 border-gray-700'
            : 'bg-gradient-to-r from-indigo-50 to-purple-50 border-gray-100'
        "
      >
        <div class="flex items-center gap-5">
          <div
            class="w-12 h-12 rounded-xl flex items-center justify-center"
            :class="isDark ? 'bg-primary-500/20' : 'bg-primary-100'"
          >
            <svg
              class="w-6 h-6"
              :class="isDark ? 'text-primary-400' : 'text-primary-600'"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fill-rule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z"
                clip-rule="evenodd"
              />
            </svg>
          </div>
          <div>
            <p class="text-xs font-medium mb-1" :class="isDark ? 'text-gray-400' : 'text-gray-500'">今日专注</p>
            <div class="flex items-center gap-4">
              <span class="text-lg font-bold" :class="isDark ? 'text-gray-100' : 'text-gray-900'">
                {{ pomodoroStats.completed_count }}
                <span class="text-xs font-normal" :class="isDark ? 'text-gray-400' : 'text-gray-500'">番茄</span>
              </span>
              <span class="text-lg font-bold" :class="isDark ? 'text-gray-100' : 'text-gray-900'">
                {{ formatPomodoroMinutes(pomodoroStats.total_minutes) }}
              </span>
            </div>
          </div>
        </div>
        <router-link
          to="/pomodoro"
          class="text-sm font-medium px-4 py-2 rounded-lg transition-colors"
          :class="isDark ? 'text-primary-400 hover:bg-primary-500/10' : 'text-primary-600 hover:bg-primary-50'"
        >
          开始专注 →
        </router-link>
      </div>

      <!-- 本周目标迷你卡片 -->
      <div
        v-if="weeklyGoals.length > 0"
        class="rounded-xl p-5 mb-5 border"
        :class="
          isDark
            ? 'bg-gradient-to-r from-emerald-900/20 to-teal-900/15 border-gray-700'
            : 'bg-gradient-to-r from-emerald-50 to-teal-50 border-gray-100'
        "
      >
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-2">
            <span class="text-lg">🎯</span>
            <h3 class="text-sm font-semibold" :class="isDark ? 'text-emerald-300' : 'text-emerald-700'">本周目标</h3>
          </div>
          <router-link
            to="/goals"
            class="text-xs font-medium transition-colors"
            :class="isDark ? 'text-emerald-400 hover:text-emerald-300' : 'text-emerald-600 hover:text-emerald-700'"
          >
            查看全部 →
          </router-link>
        </div>
        <div class="space-y-3">
          <div v-for="goal in weeklyGoals" :key="goal.id" class="flex items-center gap-3">
            <span class="text-base shrink-0">{{ goalTypeIcon(goal.type) }}</span>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-1">
                <span class="text-xs font-medium truncate" :class="isDark ? 'text-gray-300' : 'text-gray-700'">
                  {{ goal.type_label }}
                </span>
                <span class="text-xs shrink-0 ml-2" :class="isDark ? 'text-gray-400' : 'text-gray-500'">
                  {{ goal.current_value }}/{{ goal.target_value }}
                </span>
              </div>
              <div class="h-1.5 rounded-full overflow-hidden" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :style="{ width: Math.min(100, goal.percent) + '%' }"
                  :class="goal.percent >= 80 ? 'bg-emerald-500' : goal.percent >= 50 ? 'bg-yellow-500' : 'bg-amber-500'"
                ></div>
              </div>
            </div>
            <span
              class="text-xs font-semibold shrink-0 w-12 text-right"
              :class="
                goal.percent >= 80
                  ? isDark
                    ? 'text-emerald-400'
                    : 'text-emerald-600'
                  : goal.percent >= 50
                    ? isDark
                      ? 'text-yellow-400'
                      : 'text-yellow-600'
                    : isDark
                      ? 'text-amber-400'
                      : 'text-amber-600'
              "
            >
              {{ goal.percent.toFixed(0) }}%
            </span>
          </div>
        </div>
      </div>

      <!-- 学习连续打卡 -->
      <div
        class="rounded-xl p-5 mb-5 border"
        :class="
          isDark
            ? 'bg-gradient-to-r from-orange-900/20 to-amber-900/15 border-gray-700'
            : 'bg-gradient-to-r from-orange-50 to-amber-50 border-gray-100'
        "
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div
              class="w-12 h-12 rounded-xl flex items-center justify-center"
              :class="isDark ? 'bg-orange-500/20' : 'bg-orange-100'"
            >
              <span class="text-2xl">🔥</span>
            </div>
            <div>
              <p class="text-xs font-medium mb-1" :class="isDark ? 'text-gray-400' : 'text-gray-500'">学习连续打卡</p>
              <div class="flex items-baseline gap-2">
                <span class="text-2xl font-bold" :class="isDark ? 'text-orange-400' : 'text-orange-600'">
                  {{ streakInfo.current_streak }}
                </span>
                <span class="text-sm" :class="isDark ? 'text-gray-400' : 'text-gray-500'">天</span>
                <span
                  v-if="streakInfo.longest_streak > 0"
                  class="text-xs ml-2"
                  :class="isDark ? 'text-gray-500' : 'text-gray-400'"
                >
                  最长 {{ streakInfo.longest_streak }} 天
                </span>
              </div>
            </div>
          </div>
          <!-- 7 天圆点图 -->
          <div class="flex items-center gap-1.5">
            <div
              v-for="day in streakInfo.last_7_days || []"
              :key="day.date"
              class="w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-medium transition-colors"
              :class="
                day.active
                  ? isDark
                    ? 'bg-orange-500/30 text-orange-400 border border-orange-500/50'
                    : 'bg-orange-500 text-white'
                  : isDark
                    ? 'bg-gray-700 text-gray-500 border border-gray-600'
                    : 'bg-gray-200 text-gray-400 border border-gray-300'
              "
              :title="day.date"
            >
              {{ day.date.slice(-2) }}
            </div>
          </div>
        </div>
        <!-- 里程碑 -->
        <div
          v-if="streakInfo.streak_milestones"
          class="flex items-center gap-3 mt-3 pt-3 border-t"
          :class="isDark ? 'border-gray-700' : 'border-gray-200'"
        >
          <span class="text-xs" :class="isDark ? 'text-gray-500' : 'text-gray-400'">里程碑</span>
          <div
            v-for="m in streakInfo.streak_milestones"
            :key="m.days"
            class="text-xs px-2 py-0.5 rounded-full font-medium"
            :class="
              m.achieved
                ? isDark
                  ? 'bg-orange-500/20 text-orange-400'
                  : 'bg-orange-100 text-orange-700'
                : isDark
                  ? 'bg-gray-700 text-gray-500'
                  : 'bg-gray-100 text-gray-400'
            "
          >
            {{ m.days }}天
          </div>
        </div>
      </div>

      <!-- 今日待办 -->
      <div
        v-if="dailyTasks.length > 0"
        class="rounded-xl p-5 mb-5 border"
        :class="
          isDark
            ? 'bg-gradient-to-r from-sky-900/20 to-cyan-900/15 border-gray-700'
            : 'bg-gradient-to-r from-sky-50 to-cyan-50 border-gray-100'
        "
      >
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-2">
            <span class="text-lg">📋</span>
            <h3 class="text-sm font-semibold" :class="isDark ? 'text-sky-300' : 'text-sky-700'">今日待办</h3>
            <span
              v-if="dailyTasksAllCompleted"
              class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/20 text-emerald-500"
            >
              全部完成
            </span>
          </div>
          <span class="text-xs" :class="isDark ? 'text-gray-500' : 'text-gray-400'">
            {{ dailyTasksCompleted }}/{{ dailyTasks.length }}
          </span>
        </div>
        <div class="space-y-2.5">
          <div v-for="task in dailyTasks" :key="task.id" class="flex items-center gap-3">
            <!-- 完成状态图标 -->
            <span
              class="w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0"
              :class="
                task.is_completed
                  ? 'bg-emerald-500'
                  : isDark
                    ? 'bg-gray-700 border border-gray-600'
                    : 'bg-gray-200 border border-gray-300'
              "
            >
              <svg v-if="task.is_completed" class="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                  clip-rule="evenodd"
                />
              </svg>
            </span>
            <!-- 任务描述 -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-0.5">
                <span
                  class="text-xs font-medium truncate"
                  :class="
                    task.is_completed
                      ? 'line-through text-gray-400 dark:text-gray-500'
                      : isDark
                        ? 'text-gray-300'
                        : 'text-gray-700'
                  "
                >
                  {{ dailyTaskLabel(task) }}
                </span>
                <span class="text-[10px] shrink-0 ml-2" :class="isDark ? 'text-gray-500' : 'text-gray-400'">
                  {{ task.completed_count }}/{{ task.target_count }}
                </span>
              </div>
              <div class="h-1 rounded-full overflow-hidden" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :style="{ width: dailyTaskPercent(task) + '%' }"
                  :class="
                    task.is_completed ? 'bg-emerald-500' : dailyTaskPercent(task) >= 50 ? 'bg-yellow-500' : 'bg-sky-500'
                  "
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 好友动态 -->
      <div
        v-if="friendList.length > 0"
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6 mb-5"
      >
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-2">
            <svg class="w-5 h-5 text-indigo-500" fill="currentColor" viewBox="0 0 20 20">
              <path
                d="M9 6a3 3 0 11-6 0 3 3 0 016 0zM17 6a3 3 0 11-6 0 3 3 0 016 0zM12.93 17c.046-.327.07-.66.07-1a6.97 6.97 0 00-1.5-4.33A5 5 0 0119 16v1h-6.07zM6 11a5 5 0 015 5v1H1v-1a5 5 0 015-5z"
              />
            </svg>
            <h3 class="text-sm font-semibold text-gray-800 dark:text-gray-200">好友本周动态</h3>
          </div>
          <router-link
            to="/friends"
            class="text-xs text-primary-500 hover:text-primary-600 dark:text-primary-400 transition-colors"
          >
            查看全部 →
          </router-link>
        </div>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          <div
            v-for="friend in friendList"
            :key="friend.friend_id"
            class="flex items-center gap-3 p-3 rounded-lg"
            :class="isDark ? 'bg-gray-700/50 hover:bg-gray-700' : 'bg-gray-50 hover:bg-gray-100'"
          >
            <div
              class="w-9 h-9 rounded-full flex items-center justify-center text-white font-bold text-sm shrink-0"
              :style="{ backgroundColor: friendAvatarColor(friend.friend_username) }"
            >
              {{ (friend.friend_nickname || friend.friend_username || '?').charAt(0).toUpperCase() }}
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-xs font-medium truncate text-gray-800 dark:text-gray-200">
                {{ friend.friend_nickname || friend.friend_username }}
              </p>
              <div
                class="flex items-center gap-2 mt-0.5 text-[10px]"
                :class="isDark ? 'text-gray-400' : 'text-gray-500'"
              >
                <span v-if="friend.weekly_cards > 0" class="text-emerald-500">卡{{ friend.weekly_cards }}</span>
                <span v-if="friend.weekly_quizzes > 0" class="text-blue-500">题{{ friend.weekly_quizzes }}</span>
                <span v-if="friend.weekly_streak > 0" class="text-orange-500">🔥{{ friend.weekly_streak }}天</span>
                <span v-if="!friend.weekly_cards && !friend.weekly_quizzes && !friend.weekly_streak">本周暂无活动</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 本周排名 -->
      <div
        v-if="weeklyRank"
        class="rounded-xl p-5 mb-5 border transition-all"
        :class="
          isDark
            ? 'bg-gradient-to-r from-rose-900/25 to-pink-900/15 border-gray-700'
            : 'bg-gradient-to-r from-rose-50 to-pink-50 border-gray-100'
        "
      >
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-2">
            <svg
              class="w-5 h-5"
              :class="isDark ? 'text-rose-400' : 'text-rose-500'"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fill-rule="evenodd"
                d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H8V3a1 1 0 00-2 0zm0 4a1 1 0 012 0v1h4V6a1 1 0 112 0v4.5a3.5 3.5 0 01-3.28 2.49L12 15h1a1 1 0 110 2H7a1 1 0 110-2h1l.28-2.01A3.5 3.5 0 015 10.5V6z"
                clip-rule="evenodd"
              />
            </svg>
            <h3 class="text-sm font-semibold" :class="isDark ? 'text-rose-300' : 'text-rose-700'">本周排名</h3>
            <span
              class="text-xs px-2 py-0.5 rounded-full"
              :class="isDark ? 'bg-rose-900/30 text-rose-400' : 'bg-rose-100 text-rose-600'"
            >
              第 {{ weeklyRank.rank }} 名
            </span>
          </div>
          <router-link
            to="/leaderboard"
            class="text-xs font-medium transition-colors"
            :class="isDark ? 'text-rose-400 hover:text-rose-300' : 'text-rose-500 hover:text-rose-600'"
          >
            查看排行 →
          </router-link>
        </div>

        <!-- 积分概览 -->
        <div class="flex items-center gap-4 mb-4">
          <div class="text-center">
            <div class="text-2xl font-bold" :class="isDark ? 'text-rose-300' : 'text-rose-600'">
              {{ weeklyRank.score }}
            </div>
            <div class="text-[10px]" :class="isDark ? 'text-gray-400' : 'text-gray-500'">本周积分</div>
          </div>
          <div class="h-8 w-px" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'"></div>
          <div class="text-center">
            <div class="text-lg font-semibold" :class="isDark ? 'text-gray-200' : 'text-gray-700'">
              {{ weeklyRank.totalUsers }}
            </div>
            <div class="text-[10px]" :class="isDark ? 'text-gray-400' : 'text-gray-500'">参与人数</div>
          </div>
          <div
            v-if="weeklyRank.nextRankScore > weeklyRank.score"
            class="h-8 w-px"
            :class="isDark ? 'bg-gray-700' : 'bg-gray-200'"
          ></div>
          <div v-if="weeklyRank.nextRankScore > weeklyRank.score" class="text-center">
            <div class="text-lg font-semibold" :class="isDark ? 'text-amber-300' : 'text-amber-600'">
              +{{ weeklyRank.nextRankScore - weeklyRank.score }}
            </div>
            <div class="text-[10px]" :class="isDark ? 'text-gray-400' : 'text-gray-500'">距上一名</div>
          </div>
        </div>

        <!-- 附近排名列表 -->
        <div class="space-y-1.5">
          <div
            v-for="u in weeklyRank.nearby"
            :key="u.rank"
            class="flex items-center gap-3 px-3 py-2 rounded-lg transition-colors"
            :class="[
              u.isMe
                ? isDark
                  ? 'bg-rose-900/30 border border-rose-700/50'
                  : 'bg-rose-100/80 border border-rose-200'
                : isDark
                  ? 'bg-gray-800/50 hover:bg-gray-700/50'
                  : 'bg-white/60 hover:bg-white/80'
            ]"
          >
            <!-- 排名数字 -->
            <span
              class="w-6 text-center text-sm font-bold shrink-0"
              :class="
                u.rank === 1
                  ? 'text-yellow-500'
                  : u.rank === 2
                    ? 'text-gray-400'
                    : u.rank === 3
                      ? 'text-amber-600'
                      : isDark
                        ? 'text-gray-400'
                        : 'text-gray-500'
              "
            >
              {{ u.rank <= 3 ? ['🥇', '🥈', '🥉'][u.rank - 1] : u.rank }}
            </span>
            <!-- 头像 -->
            <div
              class="w-7 h-7 rounded-full flex items-center justify-center text-white font-bold text-xs shrink-0"
              :style="{ backgroundColor: friendAvatarColor(u.name) }"
            >
              {{ u.name.charAt(0).toUpperCase() }}
            </div>
            <!-- 名字 -->
            <span
              class="flex-1 min-w-0 text-sm truncate"
              :class="
                u.isMe
                  ? isDark
                    ? 'text-rose-300 font-semibold'
                    : 'text-rose-700 font-semibold'
                  : isDark
                    ? 'text-gray-300'
                    : 'text-gray-700'
              "
            >
              {{ u.name }}
              <span
                v-if="u.isMe"
                class="text-[10px] ml-1 px-1.5 py-0.5 rounded-full"
                :class="isDark ? 'bg-rose-800/50 text-rose-300' : 'bg-rose-200 text-rose-600'"
              >
                我
              </span>
            </span>
            <!-- 积分 -->
            <span
              class="text-sm font-medium tabular-nums"
              :class="
                u.isMe ? (isDark ? 'text-rose-300' : 'text-rose-600') : isDark ? 'text-gray-400' : 'text-gray-500'
              "
            >
              {{ u.score }} 分
            </span>
          </div>
        </div>
      </div>

      <!-- 知识诊断入口 -->
      <router-link
        to="/diagnosis"
        class="block rounded-xl p-5 mb-5 border transition-all hover:shadow-md group"
        :class="
          isDark
            ? 'bg-gradient-to-r from-violet-900/25 to-fuchsia-900/15 border-gray-700 hover:border-violet-700/50'
            : 'bg-gradient-to-r from-violet-50 to-fuchsia-50 border-gray-100 hover:border-violet-200'
        "
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div
              class="w-12 h-12 rounded-xl flex items-center justify-center"
              :class="isDark ? 'bg-violet-500/20' : 'bg-violet-100'"
            >
              <svg
                class="w-6 h-6"
                :class="isDark ? 'text-violet-400' : 'text-violet-600'"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path
                  fill-rule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-11a1 1 0 10-2 0v2H7a1 1 0 100 2h2v2a1 1 0 102 0v-2h2a1 1 0 100-2h-2V7z"
                  clip-rule="evenodd"
                />
              </svg>
            </div>
            <div>
              <p class="text-xs font-medium mb-1" :class="isDark ? 'text-gray-400' : 'text-gray-500'">学情诊断</p>
              <p class="text-sm font-semibold" :class="isDark ? 'text-gray-200' : 'text-gray-800'">
                多维度分析学习弱点
              </p>
              <p class="text-xs mt-0.5" :class="isDark ? 'text-gray-500' : 'text-gray-400'">
                雷达图 · 薄弱知识点 · 题型分析
              </p>
            </div>
          </div>
          <span
            class="text-sm font-medium px-4 py-2 rounded-lg transition-colors"
            :class="
              isDark ? 'text-violet-400 group-hover:bg-violet-500/10' : 'text-violet-600 group-hover:bg-violet-50'
            "
          >
            开始诊断 →
          </span>
        </div>
      </router-link>

      <!-- 知识洞察入口 -->
      <router-link
        to="/insights"
        class="block rounded-xl p-5 mb-5 border transition-all hover:shadow-md group"
        :class="
          isDark
            ? 'bg-gradient-to-r from-cyan-900/25 to-teal-900/15 border-gray-700 hover:border-cyan-700/50'
            : 'bg-gradient-to-r from-cyan-50 to-teal-50 border-gray-100 hover:border-cyan-200'
        "
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div
              class="w-12 h-12 rounded-xl flex items-center justify-center"
              :class="isDark ? 'bg-cyan-500/20' : 'bg-cyan-100'"
            >
              <svg
                class="w-6 h-6"
                :class="isDark ? 'text-cyan-400' : 'text-cyan-600'"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path
                  d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1z"
                />
              </svg>
            </div>
            <div>
              <p class="text-xs font-medium mb-1" :class="isDark ? 'text-gray-400' : 'text-gray-500'">知识洞察</p>
              <p class="text-sm font-semibold" :class="isDark ? 'text-gray-200' : 'text-gray-800'">
                发现跨材料知识交叉点
              </p>
              <p class="text-xs mt-0.5" :class="isDark ? 'text-gray-500' : 'text-gray-400'">
                关联图 · 交叉概念 · 深入学习
              </p>
            </div>
          </div>
          <span
            class="text-sm font-medium px-4 py-2 rounded-lg transition-colors"
            :class="isDark ? 'text-cyan-400 group-hover:bg-cyan-500/10' : 'text-cyan-600 group-hover:bg-cyan-50'"
          >
            探索关联 →
          </span>
        </div>
      </router-link>

      <!-- 图表区域：第一行 -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-5 mb-5">
        <!-- Agent 调用分布 - 环形图 -->
        <div
          class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6"
        >
          <h2 class="text-lg font-semibold dark:text-gray-100 mb-4">Agent 调用分布</h2>
          <div ref="donutChart" class="w-full" style="height: 280px"></div>
          <div v-if="!hasAgentData" class="text-center text-gray-400 dark:text-gray-500 py-12">
            <p class="text-3xl mb-2">🤖</p>
            <p class="text-sm">分析材料后即可查看</p>
          </div>
        </div>

        <!-- Agent 性能对比 - 柱状图 -->
        <div
          class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6 lg:col-span-2"
        >
          <h2 class="text-lg font-semibold dark:text-gray-100 mb-4">Agent 性能对比</h2>
          <div ref="barChart" class="w-full" style="height: 280px"></div>
          <div v-if="!hasAgentData" class="text-center text-gray-400 dark:text-gray-500 py-12">
            <p class="text-3xl mb-2">📊</p>
            <p class="text-sm">分析材料后即可查看</p>
          </div>
        </div>
      </div>

      <!-- 图表区域：第二行 -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-5 mb-5">
        <!-- 全年学习日历热力图 -->
        <div
          class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6 lg:col-span-2"
        >
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-semibold dark:text-gray-100">学习活跃度</h2>
            <div class="flex items-center gap-2">
              <!-- 统计摘要 -->
              <div class="hidden sm:flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500 mr-2">
                <span>
                  活跃
                  <strong class="text-gray-700 dark:text-gray-300">{{ calendarStats.activeDays }}</strong>
                  天
                </span>
                <span>
                  共
                  <strong class="text-gray-700 dark:text-gray-300">{{ calendarStats.totalActivities }}</strong>
                  次
                </span>
                <span v-if="calendarStats.streak > 0">
                  连续
                  <strong class="text-emerald-600 dark:text-emerald-400">{{ calendarStats.streak }}</strong>
                  天
                </span>
              </div>
              <!-- 年份选择器 -->
              <select
                v-model.number="calendarYear"
                class="text-sm rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 px-2 py-1 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
                @change="loadCalendarData"
              >
                <option v-for="y in yearOptions" :key="y" :value="y">{{ y }}</option>
              </select>
            </div>
          </div>
          <!-- 移动端统计摘要 -->
          <div class="flex sm:hidden items-center gap-3 text-xs text-gray-400 dark:text-gray-500 mb-3">
            <span>
              活跃
              <strong class="text-gray-700 dark:text-gray-300">{{ calendarStats.activeDays }}</strong>
              天
            </span>
            <span>
              共
              <strong class="text-gray-700 dark:text-gray-300">{{ calendarStats.totalActivities }}</strong>
              次
            </span>
            <span v-if="calendarStats.streak > 0">
              连续
              <strong class="text-emerald-600 dark:text-emerald-400">{{ calendarStats.streak }}</strong>
              天
            </span>
          </div>
          <div ref="heatmapChart" class="w-full" style="height: 180px"></div>
          <!-- 图例 -->
          <div class="flex items-center justify-end gap-1.5 mt-2 text-xs text-gray-400 dark:text-gray-500">
            <span>少</span>
            <span class="w-3 h-3 rounded-sm" :style="{ background: isDark ? '#1e293b' : '#ebedf0' }"></span>
            <span class="w-3 h-3 rounded-sm" :style="{ background: isDark ? '#0e4429' : '#9be9a8' }"></span>
            <span class="w-3 h-3 rounded-sm" :style="{ background: isDark ? '#006d32' : '#40c463' }"></span>
            <span class="w-3 h-3 rounded-sm" :style="{ background: isDark ? '#26a641' : '#30a14e' }"></span>
            <span class="w-3 h-3 rounded-sm" :style="{ background: isDark ? '#39d353' : '#216e39' }"></span>
            <span>多</span>
          </div>
        </div>

        <!-- 质量评分仪表盘 + 质量分布 + Judge 评语 -->
        <div
          class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6"
        >
          <h2 class="text-lg font-semibold dark:text-gray-100 mb-4">AI 质量评分</h2>
          <div ref="gaugeChart" class="w-full" style="height: 160px"></div>
          <div class="text-center mt-1 mb-3">
            <span class="text-3xl font-bold text-primary-600">{{ qualityDisplay }}</span>
            <span class="text-gray-400 dark:text-gray-500 text-sm ml-1">/ 10</span>
          </div>
          <!-- 质量分布条 -->
          <div v-if="qualityDistTotal > 0" class="mb-3">
            <div class="flex items-center justify-between text-[10px] text-gray-400 dark:text-gray-500 mb-1">
              <span>质量分布</span>
              <span>{{ qualityDistTotal }} 次评估</span>
            </div>
            <div class="flex h-2 rounded-full overflow-hidden bg-gray-100 dark:bg-gray-700">
              <div
                class="bg-red-400 dark:bg-red-500 transition-all"
                :style="{ width: qualityDistPct('0-3') + '%' }"
              ></div>
              <div
                class="bg-amber-400 dark:bg-amber-500 transition-all"
                :style="{ width: qualityDistPct('3-6') + '%' }"
              ></div>
              <div
                class="bg-emerald-400 dark:bg-emerald-500 transition-all"
                :style="{ width: qualityDistPct('6-8') + '%' }"
              ></div>
              <div
                class="bg-indigo-500 dark:bg-indigo-400 transition-all"
                :style="{ width: qualityDistPct('8-10') + '%' }"
              ></div>
            </div>
            <div class="flex items-center gap-2 mt-1.5 text-[10px] text-gray-400 dark:text-gray-500">
              <span class="flex items-center gap-0.5">
                <span class="w-2 h-2 rounded-sm bg-red-400 dark:bg-red-500"></span>
                0-3
              </span>
              <span class="flex items-center gap-0.5">
                <span class="w-2 h-2 rounded-sm bg-amber-400 dark:bg-amber-500"></span>
                3-6
              </span>
              <span class="flex items-center gap-0.5">
                <span class="w-2 h-2 rounded-sm bg-emerald-400 dark:bg-emerald-500"></span>
                6-8
              </span>
              <span class="flex items-center gap-0.5">
                <span class="w-2 h-2 rounded-sm bg-indigo-500 dark:bg-indigo-400"></span>
                8-10
              </span>
            </div>
          </div>
          <!-- 最近 Judge 评语 -->
          <div v-if="judgeComments.length > 0" class="border-t border-gray-100 dark:border-gray-700 pt-3">
            <div class="text-[10px] text-gray-400 dark:text-gray-500 mb-2">最近评语</div>
            <div class="space-y-2 max-h-36 overflow-y-auto custom-scroll">
              <div
                v-for="(c, i) in judgeComments"
                :key="i"
                class="text-xs p-2 rounded-lg bg-gray-50 dark:bg-gray-700/50 border border-gray-100 dark:border-gray-600"
              >
                <div class="flex items-center gap-1.5 mb-0.5">
                  <span class="font-medium text-gray-700 dark:text-gray-300">
                    {{ agentLabels[c.agent_name] || c.agent_name }}
                  </span>
                  <span class="text-[10px] px-1.5 py-0.5 rounded-full font-medium" :class="scoreColorClass(c.score)">
                    {{ c.score.toFixed(1) }}
                  </span>
                </div>
                <div class="text-gray-500 dark:text-gray-400 leading-relaxed line-clamp-2">{{ c.comment }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 最近学习材料 -->
      <div
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6 mb-5"
      >
        <div class="flex items-center justify-between mb-5">
          <h2 class="text-lg font-semibold dark:text-gray-100">最近学习材料</h2>
          <router-link to="/upload" class="text-sm text-primary-600 hover:text-primary-700 font-medium">
            上传新材料 →
          </router-link>
        </div>

        <div v-if="materials.length === 0" class="text-center py-12 text-gray-400 dark:text-gray-500">
          <p class="text-4xl mb-3">📄</p>
          <p>还没有上传过学习材料</p>
          <router-link to="/upload" class="text-primary-600 hover:underline text-sm mt-2 inline-block">
            去上传第一份材料
          </router-link>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="m in materials"
            :key="m.id"
            class="rounded-lg border border-gray-100 dark:border-gray-700 hover:border-primary-200 hover:bg-primary-50/30 dark:hover:bg-gray-700 transition-all cursor-pointer md:flex md:items-center md:justify-between md:p-4"
          >
            <!-- 移动端卡片布局 -->
            <div class="md:hidden p-4 pb-3">
              <router-link
                :to="`/materials/${m.id}`"
                class="font-medium text-gray-900 dark:text-gray-100 block truncate"
              >
                {{ m.title }}
              </router-link>
              <div class="flex items-center justify-between mt-2">
                <p class="text-sm text-gray-400 dark:text-gray-500">{{ formatDate(m.created_at) }}</p>
                <span class="px-2 py-1 rounded-full text-xs font-medium" :class="statusClass(m.status)">
                  {{ statusLabel(m.status) }}
                </span>
              </div>
            </div>

            <!-- 桌面端横排布局 -->
            <div class="hidden md:block flex-1 min-w-0">
              <h3 class="font-medium text-gray-900 dark:text-gray-100 truncate">{{ m.title }}</h3>
              <p class="text-sm text-gray-400 dark:text-gray-500 mt-0.5">
                {{ formatDate(m.created_at) }}
              </p>
            </div>
            <span
              class="hidden md:inline-block px-3 py-1 rounded-full text-xs font-medium ml-4"
              :class="statusClass(m.status)"
            >
              {{ statusLabel(m.status) }}
            </span>
          </div>
        </div>
      </div>

      <!-- ========== 学习成就 ========== -->
      <div
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6 mb-5"
      >
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-3">
            <h2 class="text-lg font-semibold dark:text-gray-100">学习成就</h2>
            <span class="text-sm text-gray-400 dark:text-gray-500">
              {{ achievementUnlocked }} / {{ achievementTotal }} 已解锁
            </span>
          </div>
          <div class="flex items-center gap-1">
            <span class="text-xs text-gray-400 dark:text-gray-500 mr-1">
              {{ achievementTotal > 0 ? Math.round((achievementUnlocked / achievementTotal) * 100) : 0 }}%
            </span>
            <div class="w-24 h-2 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
              <div
                class="h-full rounded-full transition-all duration-500"
                :class="
                  achievementUnlocked === achievementTotal && achievementTotal > 0 ? 'bg-amber-500' : 'bg-primary-500'
                "
                :style="{ width: (achievementTotal > 0 ? (achievementUnlocked / achievementTotal) * 100 : 0) + '%' }"
              ></div>
            </div>
          </div>
        </div>

        <!-- 最近解锁 -->
        <div v-if="recentUnlocks.length > 0" class="mb-4 flex flex-wrap gap-2">
          <div
            v-for="a in recentUnlocks"
            :key="'recent-' + a.id"
            class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-amber-50 dark:bg-amber-900/20 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-800/50"
          >
            <span>{{ a.icon }}</span>
            <span>{{ a.name }}</span>
            <span class="text-amber-400 dark:text-amber-600">·</span>
            <span class="text-amber-500/70 dark:text-amber-500/60">{{ formatAchievementDate(a.unlocked_at) }}</span>
          </div>
        </div>

        <!-- 分类过滤 -->
        <div class="flex flex-wrap gap-1.5 mb-4">
          <button
            v-for="(label, key) in categoryLabels"
            :key="key"
            class="px-3 py-1 rounded-full text-xs font-medium transition-all"
            :class="
              activeCategory === key
                ? 'bg-primary-500 text-white shadow-sm'
                : 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600'
            "
            @click="activeCategory = key"
          >
            {{ label }}
          </button>
        </div>

        <!-- 成就网格 -->
        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
          <div
            v-for="a in filteredAchievements"
            :key="a.id"
            class="relative group rounded-xl p-3 text-center transition-all duration-200 border"
            :class="
              a.unlocked
                ? 'bg-gradient-to-b from-amber-50 to-white dark:from-amber-900/20 dark:to-gray-800 border-amber-200 dark:border-amber-800/50 hover:shadow-md'
                : 'bg-gray-50 dark:bg-gray-800/50 border-gray-200 dark:border-gray-700 opacity-70 hover:opacity-100'
            "
          >
            <!-- 图标 -->
            <div
              class="text-2xl mb-1.5 transition-transform group-hover:scale-110"
              :class="{ 'grayscale opacity-50': !a.unlocked }"
            >
              {{ a.icon }}
            </div>
            <!-- 名称 -->
            <div
              class="text-xs font-semibold mb-0.5 truncate"
              :class="a.unlocked ? 'text-gray-900 dark:text-gray-100' : 'text-gray-500 dark:text-gray-400'"
            >
              {{ a.name }}
            </div>
            <!-- 描述（tooltip） -->
            <div class="text-[10px] leading-tight text-gray-400 dark:text-gray-500 mb-1.5 line-clamp-2 min-h-[24px]">
              {{ a.description }}
            </div>
            <!-- 进度条（未解锁时显示） -->
            <div v-if="!a.unlocked" class="mt-1">
              <div class="w-full h-1 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                <div
                  class="h-full bg-primary-400 rounded-full transition-all duration-300"
                  :style="{ width: a.progress_pct + '%' }"
                ></div>
              </div>
              <div class="text-[10px] text-gray-400 dark:text-gray-500 mt-0.5">{{ a.progress }} / {{ a.target }}</div>
            </div>
            <!-- 解锁标记 -->
            <div v-if="a.unlocked" class="absolute top-1.5 right-1.5">
              <span class="text-[10px] text-amber-500">✓</span>
            </div>
          </div>
        </div>
      </div>

      <!-- AI 调用概览 -->
      <div
        class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6"
      >
        <h2 class="text-lg font-semibold dark:text-gray-100 mb-5">AI 调用概览</h2>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div
            v-for="metric in metrics"
            :key="metric.label"
            class="text-center p-4 rounded-lg bg-gray-50 dark:bg-gray-700"
          >
            <div class="text-2xl font-bold text-primary-600">{{ metric.value }}</div>
            <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ metric.label }}</div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import * as echarts from 'echarts'
import { useAuthStore } from '../stores/auth'
import {
  getUserStats,
  listMaterials,
  getMetrics,
  getAchievements,
  getRecommendations,
  getCalendarHeatmap,
  getPomodoroStats,
  getGoalProgress,
  getStreaks,
  getDailyTasks,
  listFriends,
  getLeaderboard
} from '../api/client'
import { useDarkMode } from '../composables/useDarkMode'
import { useNetworkStatus } from '../composables/useNetworkStatus'

const auth = useAuthStore()
const { isDark } = useDarkMode()
const { isOnline, isResponseFromCache } = useNetworkStatus()

// ===== 加载状态 =====
const loading = ref(true)
const fromCache = ref(false)

// ===== 学习建议 =====
const recommendations = ref([])

// ===== 响应式数据 =====
const stats = ref([
  { label: '已分析材料', value: '0', icon: '📚', bgColor: 'bg-blue-50', color: 'text-blue-600', desc: '累计上传分析' },
  { label: '知识卡片', value: '0', icon: '🃏', bgColor: 'bg-purple-50', color: 'text-purple-600', desc: '已生成卡片' },
  { label: '练习题', value: '0', icon: '✏️', bgColor: 'bg-amber-50', color: 'text-amber-600', desc: '已完成题目' },
  { label: '正确率', value: '--', icon: '🎯', bgColor: 'bg-green-50', color: 'text-green-600', desc: '答题正确率' }
])

// ===== 番茄钟专注统计 =====
const pomodoroStats = ref({ completed_count: 0, total_minutes: 0 })

// ===== 待复习卡片数 =====
const dueCardCount = ref(0)
function formatPomodoroMinutes(m) {
  if (m < 60) return m + '分钟'
  const h = Math.floor(m / 60)
  const r = m % 60
  return r > 0 ? h + '时' + r + '分' : h + '小时'
}

// ===== 本周学习目标 =====
const weeklyGoals = ref([])

// ===== 学习连续打卡 =====
const streakInfo = ref({ current_streak: 0, longest_streak: 0, total_days: 0, last_7_days: [], streak_milestones: [] })

// ===== 今日待办任务 =====
const dailyTasks = ref([])
const dailyTasksCompleted = computed(() => dailyTasks.value.filter((t) => t.is_completed).length)
const dailyTasksAllCompleted = computed(
  () => dailyTasks.value.length > 0 && dailyTasksCompleted.value === dailyTasks.value.length
)

// ===== 好友动态 =====
const friendList = ref([])

// ===== 本周排名 =====
const weeklyRank = ref(null) // { rank, score, nearby: [{rank, name, score, isMe}], nextRankScore }

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
function friendAvatarColor(name) {
  if (!name) return avatarColors[0]
  let h = 0
  for (let i = 0; i < name.length; i++) h = name.charCodeAt(i) + ((h << 5) - h)
  return avatarColors[Math.abs(h) % avatarColors.length]
}

function dailyTaskLabel(task) {
  const labels = {
    review_due_cards: '复习到期卡片',
    complete_n_quizzes: '完成练习题',
    study_n_minutes: '学习时长(分钟)',
    read_material: '阅读材料',
    upload_material: '上传材料'
  }
  return labels[task.type] || task.type
}

function dailyTaskPercent(task) {
  if (!task.target_count || task.target_count <= 0) return 0
  return Math.min(100, Math.round((task.completed_count / task.target_count) * 100))
}

const materials = ref([])
const metrics = ref([
  { label: 'LLM 调用次数', value: '0' },
  { label: '平均延迟', value: '0ms' },
  { label: 'Token 消耗', value: '0' },
  { label: '平均质量分', value: '--' }
])

let metricsData = null
const hasAgentData = ref(false)
const qualityDisplay = ref('--')

// ===== 质量分布 + Judge 评语 =====
const qualityDistribution = ref({})
const judgeComments = computed(() => {
  if (!metricsData?.judge_comments) return []
  const all = []
  for (const [agentName, comments] of Object.entries(metricsData.judge_comments)) {
    for (const c of comments) {
      all.push({ ...c, agent_name: agentName })
    }
  }
  // 按时间倒序，取最近 5 条
  return all.sort((a, b) => new Date(b.created_at) - new Date(a.created_at)).slice(0, 5)
})

const qualityDistTotal = computed(() => {
  const d = qualityDistribution.value
  return (d['0-3'] || 0) + (d['3-6'] || 0) + (d['6-8'] || 0) + (d['8-10'] || 0)
})

function qualityDistPct(range) {
  const total = qualityDistTotal.value
  if (total === 0) return 0
  return (((qualityDistribution.value[range] || 0) / total) * 100).toFixed(1)
}

function scoreColorClass(score) {
  if (score >= 8) return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
  if (score >= 6) return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
  return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
}

// ===== 学习日历热力图 =====
const currentYear = new Date().getFullYear()
const calendarYear = ref(currentYear)
const calendarData = ref(null)
const calendarStats = computed(() => {
  if (!calendarData.value) return { activeDays: 0, totalActivities: 0, streak: 0 }
  const data = calendarData.value
  const activeDays = data.filter((d) => d.count > 0).length
  const totalActivities = data.reduce((sum, d) => sum + d.count, 0)
  // 计算当前连续活跃天数（从今天往前推）
  let streak = 0
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  for (let i = 0; i < data.length; i++) {
    const d = new Date(today.getTime() - i * 86400000)
    const dateStr = d.toISOString().slice(0, 10)
    const entry = data.find((e) => e.date === dateStr)
    if (entry && entry.count > 0) streak++
    else break
  }
  return { activeDays, totalActivities, streak }
})

// ===== 成就系统 =====
const achievements = ref([])
const achievementTotal = ref(0)
const achievementUnlocked = ref(0)
const activeCategory = ref('all')
const categoryLabels = {
  all: '全部',
  learning: '学习',
  practice: '练习',
  review: '复习',
  exploration: '探索',
  special: '特殊'
}

const filteredAchievements = computed(() => {
  if (activeCategory.value === 'all') return achievements.value
  return achievements.value.filter((a) => a.category === activeCategory.value)
})

const recentUnlocks = computed(() => {
  return achievements.value
    .filter((a) => a.unlocked)
    .sort((a, b) => new Date(b.unlocked_at) - new Date(a.unlocked_at))
    .slice(0, 4)
})

// ===== 图表实例引用 =====
const donutChart = ref(null)
const barChart = ref(null)
const heatmapChart = ref(null)
const gaugeChart = ref(null)

let donutInstance = null
let barInstance = null
let heatmapInstance = null
let gaugeInstance = null

// ===== 计算属性 =====
const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '夜深了'
  if (h < 12) return '早上好'
  if (h < 18) return '下午好'
  return '晚上好'
})

// ===== Agent 配色 =====
const agentColors = {
  Analyst: '#6366f1', // primary
  QuizMaster: '#f59e0b', // amber
  CardMaker: '#10b981', // emerald
  MapBuilder: '#3b82f6', // blue
  Judge: '#ef4444' // red
}

const agentLabels = {
  Analyst: '分析师',
  QuizMaster: '出题官',
  CardMaker: '卡片师',
  MapBuilder: '图谱师',
  Judge: '评审官'
}

// ===== 工具函数 =====
function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function statusClass(s) {
  const map = {
    pending: 'bg-gray-100 text-gray-600',
    analyzing: 'bg-blue-100 text-blue-600',
    completed: 'bg-green-100 text-green-600',
    partial: 'bg-amber-100 text-amber-600',
    failed: 'bg-red-100 text-red-600'
  }
  return map[s] || 'bg-gray-100 text-gray-600'
}

function statusLabel(s) {
  const map = { pending: '待分析', analyzing: '分析中', completed: '已完成', partial: '部分完成', failed: '失败' }
  return map[s] || s
}

function formatAchievementDate(d) {
  if (!d) return ''
  const date = new Date(d)
  const now = new Date()
  const diff = now - date
  const days = Math.floor(diff / 86400000)
  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return days + ' 天前'
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

function goalTypeIcon(type) {
  const icons = { review_cards: '🃏', complete_quizzes: '✏️', study_minutes: '⏱️', upload_materials: '📚' }
  return icons[type] || '🎯'
}

// ===== 学习日历年份选项 =====
const yearOptions = computed(() => {
  const y = new Date().getFullYear()
  const years = []
  for (let i = y; i >= y - 3; i--) years.push(i)
  return years
})

// ===== 加载日历数据 =====
async function loadCalendarData() {
  try {
    const res = await getCalendarHeatmap(calendarYear.value)
    calendarData.value = res.data
    await nextTick()
    initHeatmapChart(calendarData.value)
  } catch (e) {
    console.error('日历数据加载失败:', e)
  }
}

// ===== 图表初始化 =====
function initDonutChart(data) {
  if (!donutChart.value) return
  const breakdown = data?.agent_breakdown || {}
  const names = Object.keys(breakdown)
  if (names.length === 0) {
    hasAgentData.value = false
    return
  }
  hasAgentData.value = true

  donutInstance = echarts.init(donutChart.value)
  const seriesData = names.map((name) => ({
    name: agentLabels[name] || name,
    value: breakdown[name],
    itemStyle: { color: agentColors[name] || '#94a3b8' }
  }))

  donutInstance.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} 次 ({d}%)',
      backgroundColor: isDark.value ? 'rgba(30,41,59,0.95)' : 'rgba(255,255,255,0.95)',
      textStyle: { color: isDark.value ? '#d1d5db' : '#374151' },
      borderColor: isDark.value ? '#374151' : '#e5e7eb'
    },
    legend: {
      bottom: 0,
      itemWidth: 12,
      itemHeight: 12,
      textStyle: { fontSize: 12, color: isDark.value ? '#9ca3af' : '#6b7280' }
    },
    series: [
      {
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['50%', '45%'],
        avoidLabelOverlap: false,
        label: { show: false },
        emphasis: {
          label: { show: true, fontSize: 14, fontWeight: 'bold' }
        },
        data: seriesData
      }
    ]
  })
}

function initBarChart(data) {
  if (!barChart.value) return
  const breakdown = data?.agent_breakdown || {}
  const quality = data?.agent_quality || {}
  const latency = data?.agent_latency || {}
  const names = Object.keys(breakdown)
  if (names.length === 0) return

  barInstance = echarts.init(barChart.value)
  const labels = names.map((n) => agentLabels[n] || n)
  const qualityData = names.map((n) => +(quality[n] || 0).toFixed(1))
  const latencyData = names.map((n) => +(latency[n] || 0).toFixed(0))
  const colors = names.map((n) => agentColors[n] || '#94a3b8')

  barInstance.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: isDark.value ? 'rgba(30,41,59,0.95)' : 'rgba(255,255,255,0.95)',
      textStyle: { color: isDark.value ? '#d1d5db' : '#374151', fontSize: 12 },
      borderColor: isDark.value ? '#374151' : '#e5e7eb',
      formatter: (params) => {
        if (!params || params.length === 0) return ''
        const agentKey = names[params[0].dataIndex]
        const label = agentLabels[agentKey] || agentKey
        let html = `<div style="font-weight:600;margin-bottom:4px">${label}</div>`
        for (const p of params) {
          html += `<div style="display:flex;align-items:center;gap:6px;margin:2px 0">
            ${p.marker}<span>${p.seriesName}:</span>
            <strong>${p.seriesName === '质量评分' ? p.value.toFixed(1) + '/10' : p.value + 'ms'}</strong>
          </div>`
        }
        // 显示最近 Judge 评语
        const comments = metricsData?.judge_comments?.[agentKey]
        if (comments && comments.length > 0) {
          html += `<div style="margin-top:6px;padding-top:6px;border-top:1px solid ${isDark.value ? '#4b5563' : '#e5e7eb'}">
            <div style="font-size:11px;color:${isDark.value ? '#f59e0b' : '#d97706'};margin-bottom:3px">💬 Judge 评语</div>
            <div style="font-size:11px;color:${isDark.value ? '#9ca3af' : '#6b7280'};max-width:240px">${comments[0].comment || '暂无评语'}</div>
          </div>`
        }
        return html
      }
    },
    legend: {
      bottom: 0,
      itemWidth: 12,
      itemHeight: 12,
      textStyle: { fontSize: 12, color: isDark.value ? '#9ca3af' : '#6b7280' }
    },
    grid: { left: 60, right: 60, top: 20, bottom: 40 },
    xAxis: {
      type: 'category',
      data: labels,
      axisLabel: { color: isDark.value ? '#9ca3af' : '#6b7280' },
      axisTick: { show: false },
      axisLine: { lineStyle: { color: isDark.value ? '#4b5563' : '#e5e7eb' } }
    },
    yAxis: [
      {
        type: 'value',
        name: '质量分',
        nameTextStyle: { color: isDark.value ? '#9ca3af' : '#6b7280', fontSize: 11 },
        max: 10,
        axisLabel: { color: isDark.value ? '#9ca3af' : '#6b7280' },
        splitLine: { lineStyle: { color: isDark.value ? '#374151' : '#f3f4f6' } }
      },
      {
        type: 'value',
        name: '延迟 (ms)',
        nameTextStyle: { color: isDark.value ? '#9ca3af' : '#6b7280', fontSize: 11 },
        axisLabel: { color: isDark.value ? '#9ca3af' : '#6b7280' },
        splitLine: { show: false }
      }
    ],
    series: [
      {
        name: '质量评分',
        type: 'bar',
        barWidth: 28,
        data: qualityData.map((v, i) => ({
          value: v,
          itemStyle: { color: colors[i], borderRadius: [4, 4, 0, 0] }
        }))
      },
      {
        name: '平均延迟',
        type: 'bar',
        yAxisIndex: 1,
        barWidth: 28,
        data: latencyData.map((v, i) => ({
          value: v,
          itemStyle: { color: colors[i], opacity: 0.4, borderRadius: [4, 4, 0, 0] }
        }))
      }
    ]
  })
}

function initHeatmapChart(data) {
  if (!heatmapChart.value || !data || data.length === 0) return

  if (heatmapInstance) {
    heatmapInstance.dispose()
  }
  heatmapInstance = echarts.init(heatmapChart.value)

  const year = calendarYear.value
  const startDate = `${year}-01-01`
  const endDate = `${year}-12-31`

  // 构建 ECharts calendar 数据格式: [date, count]
  const chartData = data.map((d) => [d.date, d.count])
  const maxCount = Math.max(...data.map((d) => d.count), 1)

  // GitHub 风格绿色渐变（亮色/暗色分别配色）
  const colorRange = isDark.value
    ? ['#1e293b', '#0e4429', '#006d32', '#26a641', '#39d353']
    : ['#ebedf0', '#9be9a8', '#40c463', '#30a14e', '#216e39']

  heatmapInstance.setOption({
    tooltip: {
      formatter: (params) => {
        const date = params.data[0]
        const count = params.data[1]
        const d = new Date(date)
        const dayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
        const dayName = dayNames[d.getDay()]
        const dateLabel = `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日 ${dayName}`
        if (count === 0) return `${dateLabel}<br/><span style="color:#9ca3af">无学习活动</span>`
        return `${dateLabel}<br/>学习活动: <strong>${count}</strong> 次`
      },
      backgroundColor: isDark.value ? 'rgba(30,41,59,0.95)' : 'rgba(255,255,255,0.95)',
      textStyle: { color: isDark.value ? '#d1d5db' : '#374151', fontSize: 12 },
      borderColor: isDark.value ? '#374151' : '#e5e7eb',
      padding: [8, 12]
    },
    visualMap: {
      show: false,
      min: 0,
      max: Math.max(maxCount, 10),
      type: 'piecewise',
      pieces: [
        { value: 0, color: colorRange[0] },
        { min: 1, max: 2, color: colorRange[1] },
        { min: 3, max: 5, color: colorRange[2] },
        { min: 6, max: 9, color: colorRange[3] },
        { min: 10, color: colorRange[4] }
      ]
    },
    calendar: {
      range: [startDate, endDate],
      cellSize: ['auto', 14],
      top: 30,
      left: 40,
      right: 10,
      orient: 'horizontal',
      splitLine: {
        show: true,
        lineStyle: {
          color: isDark.value ? '#374151' : '#e5e7eb',
          width: 0.5
        }
      },
      itemStyle: {
        color: colorRange[0],
        borderColor: isDark.value ? '#1f2937' : '#fff',
        borderWidth: 2,
        borderRadius: 3
      },
      dayLabel: {
        firstDay: 1, // 周一起始
        nameMap: isDark.value ? 'ZH' : 'ZH',
        color: isDark.value ? '#9ca3af' : '#767676',
        fontSize: 11,
        margin: 8
      },
      monthLabel: {
        nameMap: 'ZH',
        color: isDark.value ? '#9ca3af' : '#767676',
        fontSize: 12,
        margin: 8
      },
      yearLabel: { show: false }
    },
    series: [
      {
        type: 'heatmap',
        coordinateSystem: 'calendar',
        calendarIndex: 0,
        data: chartData,
        emphasis: {
          itemStyle: {
            borderColor: isDark.value ? '#6366f1' : '#4f46e5',
            borderWidth: 1.5,
            shadowBlur: 6,
            shadowColor: 'rgba(99,102,241,0.3)'
          }
        }
      }
    ]
  })
}

function initGaugeChart(score) {
  if (!gaugeChart.value) return

  gaugeInstance = echarts.init(gaugeChart.value)
  const val = score > 0 ? score : 0
  qualityDisplay.value = score > 0 ? score.toFixed(1) : '--'

  gaugeInstance.setOption({
    series: [
      {
        type: 'gauge',
        startAngle: 200,
        endAngle: -20,
        min: 0,
        max: 10,
        splitNumber: 5,
        pointer: {
          show: true,
          length: '60%',
          width: 4,
          itemStyle: { color: '#4f46e5' }
        },
        axisLine: {
          lineStyle: {
            width: 16,
            color: [
              [0.3, '#ef4444'],
              [0.6, '#f59e0b'],
              [0.8, '#10b981'],
              [1, '#6366f1']
            ]
          }
        },
        axisTick: { show: false },
        splitLine: {
          distance: -18,
          length: 12,
          lineStyle: { color: isDark.value ? '#4b5563' : '#e5e7eb', width: 2 }
        },
        axisLabel: {
          distance: 22,
          color: isDark.value ? '#9ca3af' : '#9ca3af',
          fontSize: 11
        },
        detail: { show: false },
        data: [{ value: val }]
      }
    ]
  })
}

// ===== 窗口 resize 处理 =====
function handleResize() {
  donutInstance?.resize()
  barInstance?.resize()
  heatmapInstance?.resize()
  gaugeInstance?.resize()
}

// ===== 生命周期 =====
onMounted(async () => {
  try {
    const [
      statsRes,
      matRes,
      metricsRes,
      calendarRes,
      achieveRes,
      recRes,
      pomodoroRes,
      goalsRes,
      streaksRes,
      tasksRes,
      friendsRes,
      leaderboardRes
    ] = await Promise.allSettled([
      getUserStats(),
      listMaterials(),
      getMetrics(),
      getCalendarHeatmap(calendarYear.value),
      getAchievements(),
      getRecommendations(),
      getPomodoroStats(),
      getGoalProgress(),
      getStreaks(),
      getDailyTasks(),
      listFriends(),
      getLeaderboard('weekly')
    ])

    if (statsRes.status === 'fulfilled') {
      const d = statsRes.value.data
      stats.value[0].value = String(d.material_count || 0)
      stats.value[1].value = String(d.card_count || 0)
      stats.value[2].value = String(d.quiz_count || 0)
      stats.value[3].value = d.accuracy ? `${Math.round(d.accuracy)}%` : '--'
      dueCardCount.value = d.due_card_count || 0
    }

    if (matRes.status === 'fulfilled') {
      materials.value = (matRes.value.data.data || []).slice(0, 5)
    }

    if (metricsRes.status === 'fulfilled') {
      metricsData = metricsRes.value.data
      const m = metricsData
      metrics.value[0].value = String(m.total_calls || 0)
      metrics.value[1].value = `${(m.avg_latency_ms || 0).toFixed(0)}ms`
      metrics.value[2].value = String(m.total_tokens || 0)
      metrics.value[3].value = m.avg_quality_score ? `${m.avg_quality_score.toFixed(1)}/10` : '--'
      qualityDistribution.value = m.quality_distribution || {}
    }

    if (calendarRes.status === 'fulfilled') {
      calendarData.value = calendarRes.value.data
    }

    if (achieveRes.status === 'fulfilled') {
      const ad = achieveRes.value.data
      achievements.value = ad.achievements || []
      achievementTotal.value = ad.total || 0
      achievementUnlocked.value = ad.unlocked_count || 0
    }

    if (recRes.status === 'fulfilled') {
      recommendations.value = recRes.value.data.recommendations || []
    }

    if (pomodoroRes.status === 'fulfilled') {
      const pd = pomodoroRes.value.data
      pomodoroStats.value = pd.today || { completed_count: 0, total_minutes: 0 }
    }

    if (goalsRes.status === 'fulfilled') {
      const gd = goalsRes.value.data
      // 只取前 2 个活跃目标在 Dashboard 显示
      weeklyGoals.value = (gd.goals || []).slice(0, 2)
    }

    if (streaksRes.status === 'fulfilled') {
      streakInfo.value = streaksRes.value.data || {
        current_streak: 0,
        longest_streak: 0,
        total_days: 0,
        last_7_days: [],
        streak_milestones: []
      }
    }

    if (tasksRes.status === 'fulfilled') {
      // 最多显示 5 项任务
      dailyTasks.value = (tasksRes.value.data.tasks || []).slice(0, 5)
    }

    if (friendsRes.status === 'fulfilled') {
      friendList.value = (friendsRes.value.data.friends || []).slice(0, 5)
    }

    if (leaderboardRes.status === 'fulfilled') {
      const ld = leaderboardRes.value.data
      const cu = ld.current_user
      if (cu && cu.score > 0) {
        // 构建附近排名：当前用户前后各 2 名
        const allUsers = ld.users || []
        const myIdx = allUsers.findIndex((u) => u.rank === cu.rank)
        const nearby = []
        const startIdx = Math.max(0, myIdx - 2)
        const endIdx = Math.min(allUsers.length - 1, myIdx + 2)
        for (let i = startIdx; i <= endIdx; i++) {
          const u = allUsers[i]
          nearby.push({
            rank: u.rank,
            name: u.nickname || u.username || '用户',
            score: u.score,
            isMe: u.rank === cu.rank
          })
        }
        // 如果当前用户不在 top 50 列表中，仍然显示自己
        if (myIdx === -1) {
          nearby.push({ rank: cu.rank, name: '我', score: cu.score, isMe: true })
        }
        weeklyRank.value = {
          rank: cu.rank,
          score: cu.score,
          nearby,
          nextRankScore: ld.next_rank_score || cu.score,
          totalUsers: ld.total_users || 0
        }
      }
    }

    // 检测离线缓存: 检查是否有响应来自 SW 缓存
    const allResults = [
      statsRes,
      matRes,
      metricsRes,
      calendarRes,
      achieveRes,
      recRes,
      pomodoroRes,
      goalsRes,
      streaksRes,
      tasksRes,
      friendsRes,
      leaderboardRes
    ]
    fromCache.value =
      !isOnline.value && allResults.some((r) => r.status === 'fulfilled' && r.value && isResponseFromCache(r.value))

    // 关闭骨架屏，切换到真实内容
    loading.value = false

    // 等待 DOM 更新后初始化图表
    await nextTick()
    if (metricsData) {
      initDonutChart(metricsData)
      initBarChart(metricsData)
      initGaugeChart(metricsData.avg_quality_score || 0)
    }
    if (calendarData.value) {
      initHeatmapChart(calendarData.value)
    }

    window.addEventListener('resize', handleResize)
  } catch (e) {
    console.error('Dashboard 数据加载失败:', e)
    loading.value = false
  }
})

// 暗色模式切换时重绘热力图
watch(isDark, async () => {
  if (calendarData.value) {
    await nextTick()
    initHeatmapChart(calendarData.value)
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  donutInstance?.dispose()
  barInstance?.dispose()
  heatmapInstance?.dispose()
  gaugeInstance?.dispose()
})
</script>

<style scoped>
.skeleton {
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 37%, #f1f5f9 63%);
  background-size: 400% 100%;
  animation: shimmer 1.4s ease infinite;
  border-radius: 6px;
}

.skeleton-box {
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 37%, #f1f5f9 63%);
  background-size: 400% 100%;
  animation: shimmer 1.4s ease infinite;
  border-radius: 8px;
}

.dark .skeleton {
  background: linear-gradient(90deg, #1e293b 25%, #334155 37%, #1e293b 63%);
  background-size: 400% 100%;
}

.dark .skeleton-box {
  background: linear-gradient(90deg, #1e293b 25%, #334155 37%, #1e293b 63%);
  background-size: 400% 100%;
}

@keyframes shimmer {
  0% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0 50%;
  }
}

.custom-scroll::-webkit-scrollbar {
  width: 3px;
}
.custom-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroll::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}
.dark .custom-scroll::-webkit-scrollbar-thumb {
  background: #4b5563;
}
</style>
