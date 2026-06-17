<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-5xl mx-auto">
    <!-- 头部 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-6">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-gray-100">学习小组</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">组队学习，共同进步</p>
      </div>
      <button
        v-if="!selectedGroup"
        class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium flex items-center gap-2 shrink-0"
        @click="showCreateModal = true"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        创建小组
      </button>
      <button
        v-else
        class="px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors text-sm font-medium flex items-center gap-2 shrink-0"
        @click="backToList"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        返回列表
      </button>
    </div>

    <!-- ========== 小组列表视图 ========== -->
    <div v-if="!selectedGroup">
      <!-- 过滤标签 -->
      <div class="flex gap-2 mb-4">
        <button
          v-for="t in filterTabs"
          :key="t.key"
          :class="
            activeFilter === t.key
              ? 'bg-indigo-100 dark:bg-indigo-900/40 text-indigo-700 dark:text-indigo-300'
              : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700'
          "
          class="px-3 py-1.5 rounded-full text-sm font-medium transition-colors"
          @click="selectFilter(t.key)"
        >
          {{ t.label }}
        </button>
      </div>

      <!-- 骨架屏 -->
      <div v-if="loading" class="space-y-3">
        <div v-for="i in 4" :key="i" class="bg-white dark:bg-gray-800 rounded-xl p-4 animate-pulse">
          <div class="h-5 bg-gray-200 dark:bg-gray-700 rounded w-1/3 mb-3"></div>
          <div class="h-4 bg-gray-100 dark:bg-gray-700 rounded w-2/3 mb-2"></div>
          <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded w-1/4"></div>
        </div>
      </div>

      <!-- 小组列表 -->
      <div v-else-if="groups.length > 0" class="grid gap-3 sm:grid-cols-2">
        <div
          v-for="g in groups"
          :key="g.id"
          class="bg-white dark:bg-gray-800 rounded-xl p-4 border border-gray-200 dark:border-gray-700 hover:border-indigo-300 dark:hover:border-indigo-600 hover:shadow-md transition-all cursor-pointer group"
          @click="openGroup(g)"
        >
          <div class="flex items-start justify-between mb-2">
            <h3
              class="font-semibold text-gray-900 dark:text-gray-100 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors line-clamp-1"
            >
              {{ g.name }}
            </h3>
            <div class="flex items-center gap-1.5 shrink-0 ml-2">
              <span
                v-if="g.is_owner"
                class="px-1.5 py-0.5 bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 text-xs rounded"
              >
                组长
              </span>
              <span
                v-if="g.is_member"
                class="px-1.5 py-0.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs rounded"
              >
                已加入
              </span>
              <span
                v-if="g.is_public"
                class="px-1.5 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 text-xs rounded"
              >
                公开
              </span>
            </div>
          </div>
          <p v-if="g.description" class="text-sm text-gray-500 dark:text-gray-400 line-clamp-2 mb-3">
            {{ g.description }}
          </p>
          <div class="flex items-center gap-4 text-xs text-gray-400 dark:text-gray-500">
            <span class="flex items-center gap-1">
              <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
                <path
                  d="M9 6a3 3 0 11-6 0 3 3 0 016 0zM17 6a3 3 0 11-6 0 3 3 0 016 0zM12.93 17c.046-.327.07-.66.07-1a6.97 6.97 0 00-1.5-4.33A5 5 0 0119 16v1h-6.07zM6 11a5 5 0 015 5v1H1v-1a5 5 0 015-5z"
                />
              </svg>
              {{ g.member_count }}/{{ g.max_members }} 人
            </span>
            <span>{{ formatTime(g.created_at) }}</span>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-else class="text-center py-16">
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
            d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"
          />
        </svg>
        <p class="text-gray-500 dark:text-gray-400 mb-2">暂无{{ activeFilter === 'my' ? '已加入的' : '' }}小组</p>
        <p class="text-sm text-gray-400 dark:text-gray-500">
          {{ activeFilter === 'my' ? '加入或创建一个小组，开始组队学习吧！' : '创建第一个学习小组吧！' }}
        </p>
      </div>
    </div>

    <!-- ========== 小组详情视图 ========== -->
    <div v-else>
      <!-- 小组信息头 -->
      <div class="bg-white dark:bg-gray-800 rounded-xl p-5 border border-gray-200 dark:border-gray-700 mb-6">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-3">
          <h2 class="text-lg font-bold text-gray-900 dark:text-gray-100">{{ selectedGroup.name }}</h2>
          <div class="flex items-center gap-2">
            <template v-if="selectedGroup.is_member">
              <button
                v-if="!selectedGroup.is_owner"
                class="px-3 py-1.5 text-sm bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/40 transition-colors"
                @click="handleLeave"
              >
                离开小组
              </button>
              <button
                v-else
                class="px-3 py-1.5 text-sm bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/40 transition-colors"
                @click="handleDelete"
              >
                删除小组
              </button>
            </template>
            <button
              v-else
              class="px-3 py-1.5 text-sm bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
              @click="handleJoin"
            >
              加入小组
            </button>
          </div>
        </div>
        <p v-if="selectedGroup.description" class="text-sm text-gray-600 dark:text-gray-400 mb-3">
          {{ selectedGroup.description }}
        </p>
        <div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
          <span class="flex items-center gap-1">
            <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
              <path
                d="M9 6a3 3 0 11-6 0 3 3 0 016 0zM17 6a3 3 0 11-6 0 3 3 0 016 0zM12.93 17c.046-.327.07-.66.07-1a6.97 6.97 0 00-1.5-4.33A5 5 0 0119 16v1h-6.07zM6 11a5 5 0 015 5v1H1v-1a5 5 0 015-5z"
              />
            </svg>
            {{ selectedGroup.member_count }}/{{ selectedGroup.max_members }} 人
          </span>
          <span v-if="selectedGroup.is_public" class="flex items-center gap-1">
            <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
              <path
                fill-rule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zM4.332 8.027a6.012 6.012 0 011.912-2.706C6.512 5.73 6.974 6 7.5 6A1.5 1.5 0 019 7.5V8a2 2 0 004 0 2 2 0 012-2 6.01 6.01 0 01.748 12.18A6.01 6.01 0 0112 14v-1a2 2 0 00-4 0v.5A2.5 2.5 0 015.5 16c-.76 0-1.47-.25-2.05-.67A6.003 6.003 0 014.332 8.027z"
                clip-rule="evenodd"
              />
            </svg>
            公开小组
          </span>
          <span>创建于 {{ formatTime(selectedGroup.created_at) }}</span>
        </div>
      </div>

      <!-- 详情标签 -->
      <div class="flex gap-2 mb-4">
        <button
          v-for="t in detailTabs"
          :key="t.key"
          :class="
            activeDetailTab === t.key
              ? 'bg-indigo-100 dark:bg-indigo-900/40 text-indigo-700 dark:text-indigo-300'
              : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700'
          "
          class="px-3 py-1.5 rounded-full text-sm font-medium transition-colors"
          @click="selectDetailTab(t.key)"
        >
          {{ t.label }}
        </button>
      </div>

      <!-- 成员列表 -->
      <div v-if="activeDetailTab === 'members'" class="space-y-3">
        <div
          v-for="m in members"
          :key="m.user_id"
          class="bg-white dark:bg-gray-800 rounded-xl p-4 border border-gray-200 dark:border-gray-700 flex items-center gap-4"
        >
          <div
            :style="{ backgroundColor: avatarColor(m.user_id) }"
            class="w-10 h-10 rounded-full flex items-center justify-center text-white font-bold text-sm shrink-0"
          >
            {{ (m.nickname || m.username || '?')[0].toUpperCase() }}
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="font-medium text-gray-900 dark:text-gray-100 truncate">{{ m.nickname || m.username }}</span>
              <span
                v-if="m.role === 'owner'"
                class="px-1.5 py-0.5 bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 text-xs rounded"
              >
                组长
              </span>
            </div>
            <div class="flex items-center gap-3 mt-1 text-xs text-gray-500 dark:text-gray-400">
              <span class="flex items-center gap-1">
                <svg class="w-3 h-3 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1z" />
                </svg>
                本周 {{ m.weekly_cards }} 卡
              </span>
              <span class="flex items-center gap-1">
                <svg class="w-3 h-3 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" />
                  <path
                    fill-rule="evenodd"
                    d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5z"
                    clip-rule="evenodd"
                  />
                </svg>
                本周 {{ m.weekly_quizzes }} 题
              </span>
            </div>
          </div>
        </div>
        <div v-if="members.length === 0 && !detailLoading" class="text-center py-8 text-gray-500 dark:text-gray-400">
          暂无成员
        </div>
      </div>

      <!-- 小组目标 -->
      <div v-if="activeDetailTab === 'goals'">
        <!-- 创建目标（仅组长） -->
        <div
          v-if="selectedGroup.is_owner"
          class="bg-white dark:bg-gray-800 rounded-xl p-4 border border-gray-200 dark:border-gray-700 mb-4"
        >
          <h3 class="text-sm font-medium text-gray-900 dark:text-gray-100 mb-3">设定新目标</h3>
          <div class="flex flex-col sm:flex-row gap-3">
            <select
              v-model="newGoal.type"
              class="px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm"
            >
              <option value="review_cards">复习卡片</option>
              <option value="complete_quizzes">完成练习</option>
              <option value="study_minutes">专注学习(分钟)</option>
            </select>
            <input
              v-model.number="newGoal.target_value"
              type="number"
              min="1"
              max="10000"
              placeholder="目标数量"
              class="px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm w-28"
            />
            <input
              v-model="newGoal.deadline"
              type="date"
              class="px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm"
            />
            <button
              :disabled="goalCreating"
              class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors text-sm disabled:opacity-50 shrink-0"
              @click="handleCreateGoal"
            >
              {{ goalCreating ? '创建中...' : '创建目标' }}
            </button>
          </div>
        </div>

        <!-- 目标列表 -->
        <div v-if="detailLoading" class="space-y-3">
          <div v-for="i in 2" :key="i" class="bg-white dark:bg-gray-800 rounded-xl p-4 animate-pulse">
            <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/3 mb-2"></div>
            <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded w-full mb-2"></div>
          </div>
        </div>
        <div v-else-if="goals.length > 0" class="space-y-3">
          <div
            v-for="g in goals"
            :key="g.id"
            class="bg-white dark:bg-gray-800 rounded-xl p-4 border border-gray-200 dark:border-gray-700"
          >
            <div class="flex items-center justify-between mb-2">
              <div class="flex items-center gap-2">
                <span :class="goalIconClass(g.type)" class="w-8 h-8 rounded-lg flex items-center justify-center">
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path
                      v-if="g.type === 'review_cards'"
                      d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
                    />
                    <path v-else-if="g.type === 'complete_quizzes'" d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" />
                    <path
                      v-if="g.type === 'complete_quizzes'"
                      fill-rule="evenodd"
                      d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3z"
                      clip-rule="evenodd"
                    />
                    <path
                      v-if="g.type === 'study_minutes'"
                      fill-rule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z"
                      clip-rule="evenodd"
                    />
                  </svg>
                </span>
                <div>
                  <span class="font-medium text-gray-900 dark:text-gray-100 text-sm">{{ g.type_label }}</span>
                  <span class="text-xs text-gray-400 dark:text-gray-500 ml-2">截止 {{ formatDate(g.deadline) }}</span>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <span
                  class="text-sm font-semibold"
                  :class="
                    g.percent >= 100 ? 'text-green-600 dark:text-green-400' : 'text-indigo-600 dark:text-indigo-400'
                  "
                >
                  {{ g.current_value }}/{{ g.target_value }}
                </span>
                <button
                  v-if="selectedGroup.is_owner"
                  class="text-gray-400 hover:text-red-500 dark:hover:text-red-400 transition-colors"
                  @click="handleDeleteGoal(g.id)"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            <!-- 进度条 -->
            <div class="w-full h-2 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
              <div
                class="h-full rounded-full transition-all duration-500"
                :class="g.percent >= 100 ? 'bg-green-500' : g.percent >= 50 ? 'bg-indigo-500' : 'bg-amber-500'"
                :style="{ width: Math.min(g.percent, 100) + '%' }"
              ></div>
            </div>
            <div class="text-right mt-1">
              <span
                class="text-xs"
                :class="g.percent >= 100 ? 'text-green-600 dark:text-green-400' : 'text-gray-400 dark:text-gray-500'"
              >
                {{ g.percent }}%
              </span>
            </div>
          </div>
        </div>
        <div v-else class="text-center py-8 text-gray-500 dark:text-gray-400">
          {{ selectedGroup.is_owner ? '设定一个小组目标，激励成员一起学习！' : '组长还没有设定目标' }}
        </div>
      </div>

      <!-- 成员贡献排行 -->
      <div v-if="activeDetailTab === 'contributions'">
        <div v-if="detailLoading" class="space-y-3">
          <div v-for="i in 3" :key="i" class="bg-white dark:bg-gray-800 rounded-xl p-4 animate-pulse">
            <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mb-2"></div>
            <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded w-3/4"></div>
          </div>
        </div>
        <div v-else-if="contributions.length > 0" class="space-y-3">
          <div
            v-for="(c, idx) in contributions"
            :key="c.user_id"
            class="bg-white dark:bg-gray-800 rounded-xl p-4 border border-gray-200 dark:border-gray-700 flex items-center gap-4"
          >
            <!-- 排名 -->
            <div
              class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold shrink-0"
              :class="
                idx === 0
                  ? 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400'
                  : idx === 1
                    ? 'bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-300'
                    : idx === 2
                      ? 'bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-400'
                      : 'bg-gray-100 dark:bg-gray-800 text-gray-400 dark:text-gray-500'
              "
            >
              {{ idx === 0 ? '🥇' : idx === 1 ? '🥈' : idx === 2 ? '🥉' : idx + 1 }}
            </div>
            <div
              :style="{ backgroundColor: avatarColor(c.user_id) }"
              class="w-10 h-10 rounded-full flex items-center justify-center text-white font-bold text-sm shrink-0"
            >
              {{ (c.nickname || c.username || '?')[0].toUpperCase() }}
            </div>
            <div class="flex-1 min-w-0">
              <span class="font-medium text-gray-900 dark:text-gray-100">{{ c.nickname || c.username }}</span>
              <div class="flex items-center gap-3 mt-1 text-xs text-gray-500 dark:text-gray-400">
                <span class="flex items-center gap-1">
                  <span class="inline-block w-2 h-2 rounded-full bg-green-500"></span>
                  {{ c.cards }} 卡片
                </span>
                <span class="flex items-center gap-1">
                  <span class="inline-block w-2 h-2 rounded-full bg-blue-500"></span>
                  {{ c.quizzes }} 练习
                </span>
              </div>
            </div>
            <div class="text-right shrink-0">
              <span class="text-lg font-bold text-indigo-600 dark:text-indigo-400">{{ c.total }}</span>
              <span class="text-xs text-gray-400 dark:text-gray-500 block">总贡献</span>
            </div>
          </div>
        </div>
        <div v-else class="text-center py-8 text-gray-500 dark:text-gray-400">暂无贡献数据</div>
      </div>
    </div>

    <!-- ========== 创建小组弹窗 ========== -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showCreateModal"
          class="fixed inset-0 z-50 flex items-center justify-center p-4"
          @click.self="showCreateModal = false"
        >
          <div class="fixed inset-0 bg-black/50" @click="showCreateModal = false"></div>
          <div class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-xl w-full max-w-md p-6 z-10">
            <h2 class="text-lg font-bold text-gray-900 dark:text-gray-100 mb-4">创建学习小组</h2>
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">小组名称 *</label>
                <input
                  v-model="createForm.name"
                  type="text"
                  maxlength="100"
                  placeholder="例如：Go 语言学习组"
                  class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">描述</label>
                <textarea
                  v-model="createForm.description"
                  maxlength="500"
                  rows="3"
                  placeholder="描述一下小组的学习目标..."
                  class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
                ></textarea>
              </div>
              <div class="flex gap-4">
                <div class="flex-1">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">最大人数</label>
                  <input
                    v-model.number="createForm.max_members"
                    type="number"
                    min="2"
                    max="100"
                    class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  />
                </div>
                <div class="flex-1 flex items-end pb-1">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input
                      v-model="createForm.is_public"
                      type="checkbox"
                      class="w-4 h-4 text-indigo-600 rounded border-gray-300 dark:border-gray-600 focus:ring-indigo-500"
                    />
                    <span class="text-sm text-gray-700 dark:text-gray-300">公开小组</span>
                  </label>
                </div>
              </div>
            </div>
            <div class="flex gap-3 mt-6">
              <button
                class="flex-1 px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors text-sm"
                @click="showCreateModal = false"
              >
                取消
              </button>
              <button
                :disabled="creating"
                class="flex-1 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors text-sm disabled:opacity-50"
                @click="handleCreate"
              >
                {{ creating ? '创建中...' : '创建' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useToast, useConfirm } from '../composables/useToast'
import {
  listGroups,
  getGroup,
  createGroup,
  deleteGroup,
  joinGroup,
  leaveGroup,
  getGroupMembers,
  getGroupProgress,
  getGroupGoals,
  createGroupGoal,
  deleteGroupGoal
} from '../api/client'

const { toast } = useToast()
const { confirm } = useConfirm()

const loading = ref(true)
const activeFilter = ref('all')
const groups = ref([])
const selectedGroup = ref(null)

// 列表过滤标签
const filterTabs = [
  { key: 'all', label: '全部小组' },
  { key: 'my', label: '我的小组' },
  { key: 'public', label: '公开小组' }
]

// 详情标签
const detailTabs = [
  { key: 'members', label: '成员' },
  { key: 'goals', label: '目标' },
  { key: 'contributions', label: '贡献排行' }
]
const activeDetailTab = ref('members')
const detailLoading = ref(false)

// 详情数据
const members = ref([])
const goals = ref([])
const contributions = ref([])

// 创建小组
const showCreateModal = ref(false)
const creating = ref(false)
const createForm = reactive({
  name: '',
  description: '',
  max_members: 20,
  is_public: true
})

// 创建目标
const goalCreating = ref(false)
const newGoal = reactive({
  type: 'review_cards',
  target_value: 100,
  deadline: ''
})

// ========== 数据加载 ==========

async function loadGroups() {
  loading.value = true
  try {
    const res = await listGroups(activeFilter.value)
    groups.value = res.data.groups || []
  } catch (err) {
    console.error('加载小组失败:', err)
    toast.error('加载小组列表失败')
  } finally {
    loading.value = false
  }
}

function selectFilter(key) {
  activeFilter.value = key
  loadGroups()
}

function selectDetailTab(key) {
  activeDetailTab.value = key
  loadDetailTab()
}

async function openGroup(group) {
  try {
    const res = await getGroup(group.id)
    selectedGroup.value = res.data.group
    activeDetailTab.value = 'members'
    await loadDetailTab()
  } catch (err) {
    console.error('加载小组详情失败:', err)
    toast.error('加载小组详情失败')
  }
}

function backToList() {
  selectedGroup.value = null
  members.value = []
  goals.value = []
  contributions.value = []
  loadGroups()
}

async function loadDetailTab() {
  if (!selectedGroup.value) return
  detailLoading.value = true
  try {
    const groupId = selectedGroup.value.id
    if (activeDetailTab.value === 'members') {
      const res = await getGroupMembers(groupId)
      members.value = res.data.members || []
    } else if (activeDetailTab.value === 'goals') {
      const res = await getGroupGoals(groupId)
      goals.value = res.data.goals || []
    } else if (activeDetailTab.value === 'contributions') {
      const res = await getGroupProgress(groupId)
      contributions.value = res.data.contributions || []
      goals.value = res.data.goals || []
    }
  } catch (err) {
    console.error('加载详情失败:', err)
    toast.error('加载数据失败')
  } finally {
    detailLoading.value = false
  }
}

// ========== 操作 ==========

async function handleCreate() {
  if (!createForm.name.trim()) {
    toast.warning('请填写小组名称')
    return
  }
  creating.value = true
  try {
    await createGroup({
      name: createForm.name.trim(),
      description: createForm.description.trim(),
      max_members: createForm.max_members || 20,
      is_public: createForm.is_public
    })
    toast.success('小组创建成功')
    showCreateModal.value = false
    createForm.name = ''
    createForm.description = ''
    createForm.max_members = 20
    createForm.is_public = true
    await loadGroups()
  } catch (err) {
    toast.error(err.response?.data?.error || '创建失败')
  } finally {
    creating.value = false
  }
}

async function handleJoin() {
  if (!selectedGroup.value) return
  try {
    await joinGroup(selectedGroup.value.id)
    toast.success('成功加入小组')
    const res = await getGroup(selectedGroup.value.id)
    selectedGroup.value = res.data.group
    await loadDetailTab()
  } catch (err) {
    toast.error(err.response?.data?.error || '加入失败')
  }
}

async function handleLeave() {
  const ok = await confirm('确定要离开这个小组吗？')
  if (!ok) return
  try {
    await leaveGroup(selectedGroup.value.id)
    toast.success('已离开小组')
    backToList()
  } catch (err) {
    toast.error(err.response?.data?.error || '离开失败')
  }
}

async function handleDelete() {
  const ok = await confirm('确定要删除这个小组吗？此操作不可恢复。')
  if (!ok) return
  try {
    await deleteGroup(selectedGroup.value.id)
    toast.success('小组已删除')
    backToList()
  } catch (err) {
    toast.error(err.response?.data?.error || '删除失败')
  }
}

async function handleCreateGoal() {
  if (!newGoal.target_value || newGoal.target_value < 1) {
    toast.warning('请设置有效的目标数量')
    return
  }
  if (!newGoal.deadline) {
    toast.warning('请选择截止日期')
    return
  }
  goalCreating.value = true
  try {
    await createGroupGoal(selectedGroup.value.id, {
      type: newGoal.type,
      target_value: newGoal.target_value,
      deadline: newGoal.deadline
    })
    toast.success('目标已创建')
    newGoal.target_value = 100
    newGoal.deadline = ''
    await loadDetailTab()
  } catch (err) {
    toast.error(err.response?.data?.error || '创建目标失败')
  } finally {
    goalCreating.value = false
  }
}

async function handleDeleteGoal(goalId) {
  const ok = await confirm('确定删除这个目标吗？')
  if (!ok) return
  try {
    await deleteGroupGoal(selectedGroup.value.id, goalId)
    toast.success('目标已删除')
    await loadDetailTab()
  } catch (err) {
    toast.error(err.response?.data?.error || '删除目标失败')
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

function avatarColor(id) {
  if (!id) return avatarColors[0]
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = (hash << 5) - hash + id.charCodeAt(i)
    hash = hash & hash
  }
  return avatarColors[Math.abs(hash) % avatarColors.length]
}

function goalIconClass(type) {
  switch (type) {
    case 'review_cards':
      return 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400'
    case 'complete_quizzes':
      return 'bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400'
    case 'study_minutes':
      return 'bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400'
    default:
      return 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
  }
}

function formatTime(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diff = now - d
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return Math.floor(diff / 60000) + ' 分钟前'
  if (diff < 86400000) return Math.floor(diff / 3600000) + ' 小时前'
  if (diff < 2592000000) return Math.floor(diff / 86400000) + ' 天前'
  return d.toLocaleDateString('zh-CN')
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

// ========== 初始化 ==========
onMounted(() => {
  // 默认截止日期为 7 天后
  const d = new Date()
  d.setDate(d.getDate() + 7)
  newGoal.deadline = d.toISOString().split('T')[0]

  loadGroups()
})
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.custom-scroll::-webkit-scrollbar {
  width: 4px;
}
.custom-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroll::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 2px;
}
:root.dark .custom-scroll::-webkit-scrollbar-thumb {
  background: #4b5563;
}
</style>
