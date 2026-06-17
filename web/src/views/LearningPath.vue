<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-4xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-6">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
          <svg class="w-6 h-6 text-primary-500" viewBox="0 0 20 20" fill="currentColor">
            <path
              fill-rule="evenodd"
              d="M3 4a1 1 0 011-1h4a1 1 0 010 2H6.414l2.293 2.293a1 1 0 01-1.414 1.414L5 6.414V8a1 1 0 01-2 0V4zm9 1a1 1 0 110-2h4a1 1 0 011 1v4a1 1 0 11-2 0V6.414l-2.293 2.293a1 1 0 11-1.414-1.414L13.586 5H12zm-9 7a1 1 0 112 0v1.586l2.293-2.293a1 1 0 011.414 1.414L6.414 15H8a1 1 0 110 2H4a1 1 0 01-1-1v-4zm13-1a1 1 0 10-2 0v1.586l-2.293-2.293a1 1 0 00-1.414 1.414L13.586 15H12a1 1 0 100 2h4a1 1 0 001-1v-4z"
              clip-rule="evenodd"
            />
          </svg>
          学习路径
        </h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">AI 根据你的材料智能规划最优学习路线</p>
      </div>
      <button
        v-if="!loading && steps.length > 0"
        :disabled="generating"
        class="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-primary-600 text-white hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors shrink-0"
        @click="generatePath(true)"
      >
        <svg v-if="generating" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
          />
        </svg>
        {{ generating ? '生成中...' : '重新生成' }}
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="space-y-6">
      <div class="animate-pulse">
        <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-4"></div>
        <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mb-8"></div>
      </div>
      <div v-for="i in 4" :key="i" class="flex gap-4 animate-pulse">
        <div class="flex flex-col items-center">
          <div class="w-10 h-10 bg-gray-200 dark:bg-gray-700 rounded-full"></div>
          <div class="w-0.5 h-16 bg-gray-200 dark:bg-gray-700 mt-2"></div>
        </div>
        <div class="flex-1 pb-6">
          <div class="bg-white dark:bg-gray-800 rounded-xl p-4 sm:p-5 border border-gray-200 dark:border-gray-700">
            <div class="h-5 bg-gray-200 dark:bg-gray-700 rounded w-2/3 mb-3"></div>
            <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-full mb-2"></div>
            <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-4/5"></div>
          </div>
        </div>
      </div>
      <p class="text-xs text-gray-400 dark:text-gray-500 mt-6 text-center animate-pulse">
        AI 正在分析你的学习材料并规划最优学习路线，通常需要 30-90 秒...
      </p>
    </div>

    <!-- Empty State -->
    <div v-else-if="empty" class="text-center py-16">
      <div class="w-20 h-20 mx-auto mb-6 bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center">
        <svg class="w-10 h-10 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7"
          />
        </svg>
      </div>
      <h3 class="text-lg font-semibold text-gray-700 dark:text-gray-300 mb-2">暂无学习材料</h3>
      <p class="text-sm text-gray-500 dark:text-gray-400 mb-6 max-w-sm mx-auto">
        请先上传并分析至少一份学习材料，AI 将为你规划最优学习路线。
      </p>
      <router-link
        to="/upload"
        class="inline-flex items-center gap-2 px-5 py-2.5 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors text-sm font-medium"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        上传材料
      </router-link>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="text-center py-16">
      <div class="w-20 h-20 mx-auto mb-6 bg-red-50 dark:bg-red-900/20 rounded-full flex items-center justify-center">
        <svg class="w-10 h-10 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="1.5"
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
      </div>
      <h3 class="text-lg font-semibold text-gray-700 dark:text-gray-300 mb-2">生成路径失败</h3>
      <p class="text-sm text-gray-500 dark:text-gray-400 mb-6">{{ error }}</p>
      <button
        class="inline-flex items-center gap-2 px-5 py-2.5 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors text-sm font-medium"
        @click="generatePath(true)"
      >
        重试
      </button>
    </div>

    <!-- Path Content -->
    <div v-else-if="steps.length > 0">
      <!-- Overview Card -->
      <div
        class="bg-gradient-to-r from-primary-50 to-indigo-50 dark:from-gray-800 dark:to-gray-800 border border-primary-100 dark:border-gray-700 rounded-xl p-4 sm:p-5 mb-8"
      >
        <div class="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-6 mb-3">
          <div class="flex items-center gap-2">
            <svg class="w-5 h-5 text-primary-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
              />
            </svg>
            <span class="text-sm font-semibold text-primary-700 dark:text-primary-400">路径概览</span>
          </div>
          <div class="flex flex-wrap gap-3 sm:gap-4 text-sm">
            <span class="text-gray-600 dark:text-gray-400">
              <strong class="text-gray-900 dark:text-white">{{ steps.length }}</strong>
              个学习步骤
            </span>
            <span class="text-gray-600 dark:text-gray-400">
              <strong class="text-gray-900 dark:text-white">{{ totalMinutes }}</strong>
              分钟
            </span>
            <span class="text-gray-600 dark:text-gray-400">
              <strong class="text-gray-900 dark:text-white">{{ materialCount }}</strong>
              份材料
            </span>
          </div>
        </div>
        <p class="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">{{ overview }}</p>
      </div>

      <!-- Timeline -->
      <div class="relative">
        <div v-for="(step, idx) in steps" :key="idx" class="flex gap-3 sm:gap-4 mb-2 last:mb-0">
          <!-- Timeline Rail -->
          <div class="flex flex-col items-center shrink-0">
            <!-- Node -->
            <button
              class="relative z-10 w-10 h-10 sm:w-12 sm:h-12 rounded-full flex items-center justify-center text-sm font-bold transition-all duration-200 border-2"
              :class="nodeClass(step, idx)"
              @click="toggleStep(idx)"
            >
              <span v-if="stepProgress(step) >= 100" class="text-green-500">
                <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fill-rule="evenodd"
                    d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                    clip-rule="evenodd"
                  />
                </svg>
              </span>
              <span v-else>{{ idx + 1 }}</span>
            </button>
            <!-- Connector Line -->
            <div
              v-if="idx < steps.length - 1"
              class="w-0.5 flex-1 min-h-[24px] transition-colors"
              :class="stepProgress(step) >= 100 ? 'bg-green-400 dark:bg-green-500' : 'bg-gray-200 dark:bg-gray-700'"
            ></div>
          </div>

          <!-- Step Card -->
          <div class="flex-1 pb-6 last:pb-0">
            <div
              class="bg-white dark:bg-gray-800 rounded-xl border transition-all duration-200 cursor-pointer"
              :class="
                expandedStep === idx
                  ? 'border-primary-300 dark:border-primary-600 shadow-md shadow-primary-100 dark:shadow-primary-900/20'
                  : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
              "
              @click="toggleStep(idx)"
            >
              <!-- Header -->
              <div class="p-4 sm:p-5">
                <div class="flex items-start justify-between gap-3 mb-2">
                  <h3 class="font-semibold text-gray-900 dark:text-white leading-snug">{{ step.title }}</h3>
                  <svg
                    class="w-5 h-5 text-gray-400 dark:text-gray-500 shrink-0 mt-0.5 transition-transform duration-200"
                    :class="{ 'rotate-180': expandedStep === idx }"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                  </svg>
                </div>

                <!-- Meta Tags -->
                <div class="flex flex-wrap gap-2 mb-2">
                  <!-- Difficulty Badge -->
                  <span
                    class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
                    :class="difficultyClass(step.difficulty)"
                  >
                    {{ difficultyLabel(step.difficulty) }}
                  </span>
                  <!-- Duration -->
                  <span
                    class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400"
                  >
                    <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    {{ step.estimated_minutes }} 分钟
                  </span>
                  <!-- Card Progress -->
                  <span
                    v-if="step.card_count > 0"
                    class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium"
                    :class="
                      stepProgress(step) >= 100
                        ? 'bg-green-50 dark:bg-green-900/30 text-green-600 dark:text-green-400'
                        : 'bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400'
                    "
                  >
                    <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                      <path
                        d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
                      />
                    </svg>
                    {{ step.reviewed_count }}/{{ step.card_count }} 卡片
                  </span>
                </div>

                <!-- Prerequisites -->
                <div v-if="step.prerequisites && step.prerequisites.length > 0" class="flex items-center gap-1.5 mb-2">
                  <svg class="w-3.5 h-3.5 text-amber-500" fill="currentColor" viewBox="0 0 20 20">
                    <path
                      fill-rule="evenodd"
                      d="M5 2a1 1 0 011 1v1h1a1 1 0 010 2H6v1a1 1 0 01-2 0V6H3a1 1 0 010-2h1V3a1 1 0 011-1zm0 10a1 1 0 011 1v1h1a1 1 0 110 2H6v1a1 1 0 11-2 0v-1H3a1 1 0 110-2h1v-1a1 1 0 011-1zM12 2a1 1 0 01.967.744L14.146 7.2 17.5 9.134a1 1 0 010 1.732l-3.354 1.935-1.18 4.455a1 1 0 01-1.933 0L9.854 12.8 6.5 10.866a1 1 0 010-1.732l3.354-1.935 1.18-4.455A1 1 0 0112 2z"
                      clip-rule="evenodd"
                    />
                  </svg>
                  <span class="text-xs text-amber-600 dark:text-amber-400">
                    前置：{{ step.prerequisites.map((p) => steps[p]?.title || `步骤${p + 1}`).join('、') }}
                  </span>
                </div>

                <!-- Description (always visible, truncated) -->
                <p
                  v-if="expandedStep !== idx"
                  class="text-sm text-gray-500 dark:text-gray-400 leading-relaxed line-clamp-2"
                >
                  {{ step.description }}
                </p>
              </div>

              <!-- Expanded Details -->
              <div
                v-if="expandedStep === idx"
                class="px-4 sm:px-5 pb-4 sm:pb-5 border-t border-gray-100 dark:border-gray-700 pt-4"
              >
                <p class="text-sm text-gray-600 dark:text-gray-400 leading-relaxed mb-4">{{ step.description }}</p>

                <!-- Progress Bar -->
                <div v-if="step.card_count > 0" class="mb-4">
                  <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 mb-1.5">
                    <span>学习进度</span>
                    <span>{{ stepProgress(step) }}%</span>
                  </div>
                  <div class="h-2 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                    <div
                      class="h-full rounded-full transition-all duration-500"
                      :class="stepProgress(step) >= 100 ? 'bg-green-500' : 'bg-primary-500'"
                      :style="{ width: stepProgress(step) + '%' }"
                    ></div>
                  </div>
                </div>

                <!-- Associated Materials -->
                <div v-if="step.material_titles && step.material_titles.length > 0">
                  <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                    关联材料
                  </h4>
                  <div class="space-y-2">
                    <router-link
                      v-for="(title, mIdx) in step.material_titles"
                      :key="mIdx"
                      :to="`/materials/${step.material_ids[mIdx]}`"
                      class="flex items-center gap-2 px-3 py-2 rounded-lg bg-gray-50 dark:bg-gray-700/50 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-sm group"
                    >
                      <svg class="w-4 h-4 text-primary-500 shrink-0" fill="currentColor" viewBox="0 0 20 20">
                        <path
                          d="M9 4.804A7.968 7.968 0 005.5 4c-1.255 0-2.443.29-3.5.804v10A7.969 7.969 0 015.5 14c1.669 0 3.218.51 4.5 1.385A7.962 7.962 0 0114.5 14c1.255 0 2.443.29 3.5.804v-10A7.968 7.968 0 0014.5 4c-1.255 0-2.443.29-3.5.804V12a1 1 0 11-2 0V4.804z"
                        />
                      </svg>
                      <span
                        class="text-gray-700 dark:text-gray-300 truncate group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors"
                      >
                        {{ title }}
                      </span>
                      <svg
                        class="w-3.5 h-3.5 text-gray-400 dark:text-gray-500 ml-auto shrink-0 group-hover:text-primary-500 transition-colors"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                      </svg>
                    </router-link>
                  </div>
                </div>

                <!-- Action Buttons -->
                <div class="flex flex-wrap gap-2 mt-4">
                  <router-link
                    v-if="step.material_ids && step.material_ids.length === 1"
                    :to="`/materials/${step.material_ids[0]}`"
                    class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 hover:bg-primary-100 dark:hover:bg-primary-900/50 transition-colors"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
                      />
                    </svg>
                    查看材料
                  </router-link>
                  <router-link
                    v-if="step.card_count > 0 && step.reviewed_count < step.card_count"
                    to="/study"
                    class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg bg-emerald-50 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-100 dark:hover:bg-emerald-900/50 transition-colors"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                      />
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    开始学习
                  </router-link>
                  <router-link
                    v-if="step.card_count > 0"
                    to="/cards"
                    class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg bg-gray-50 dark:bg-gray-700/50 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                  >
                    <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
                      <path
                        d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
                      />
                    </svg>
                    查看卡片
                  </router-link>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Completion Banner -->
      <div
        v-if="overallProgress >= 100"
        class="mt-8 bg-gradient-to-r from-green-50 to-emerald-50 dark:from-green-900/20 dark:to-emerald-900/20 border border-green-200 dark:border-green-800 rounded-xl p-4 sm:p-5 text-center"
      >
        <div
          class="w-12 h-12 mx-auto mb-3 bg-green-100 dark:bg-green-900/40 rounded-full flex items-center justify-center"
        >
          <svg class="w-6 h-6 text-green-500" fill="currentColor" viewBox="0 0 20 20">
            <path
              fill-rule="evenodd"
              d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
              clip-rule="evenodd"
            />
          </svg>
        </div>
        <h3 class="text-lg font-semibold text-green-700 dark:text-green-400 mb-1">学习路径已全部完成！</h3>
        <p class="text-sm text-green-600 dark:text-green-500">恭喜你已经完成了所有学习步骤，继续保持学习习惯吧！</p>
      </div>

      <!-- Overall Progress Footer -->
      <div
        v-else-if="steps.length > 0"
        class="mt-8 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl p-4 sm:p-5"
      >
        <div class="flex items-center justify-between text-sm mb-2">
          <span class="text-gray-600 dark:text-gray-400">总体学习进度</span>
          <span class="font-semibold text-gray-900 dark:text-white">{{ overallProgress }}%</span>
        </div>
        <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
          <div
            class="h-full bg-gradient-to-r from-primary-500 to-indigo-500 rounded-full transition-all duration-500"
            :style="{ width: overallProgress + '%' }"
          ></div>
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
          已复习 {{ totalReviewed }} / {{ totalCards }} 张卡片
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getLearningPath } from '../api/client'
import { useToast } from '../composables/useToast'

const toast = useToast()

const loading = ref(true)
const generating = ref(false)
const empty = ref(false)
const error = ref('')
const overview = ref('')
const totalHours = ref(0)
const steps = ref([])
const materialCount = ref(0)
const expandedStep = ref(null)

const totalMinutes = computed(() => {
  return steps.value.reduce((sum, s) => sum + (s.estimated_minutes || 0), 0)
})

const totalCards = computed(() => {
  return steps.value.reduce((sum, s) => sum + (s.card_count || 0), 0)
})

const totalReviewed = computed(() => {
  return steps.value.reduce((sum, s) => sum + (s.reviewed_count || 0), 0)
})

const overallProgress = computed(() => {
  if (totalCards.value === 0) return 0
  return Math.round((totalReviewed.value / totalCards.value) * 100)
})

function stepProgress(step) {
  if (!step.card_count || step.card_count === 0) return 0
  return Math.round(((step.reviewed_count || 0) / step.card_count) * 100)
}

function toggleStep(idx) {
  expandedStep.value = expandedStep.value === idx ? null : idx
}

function nodeClass(step, idx) {
  const progress = stepProgress(step)
  if (progress >= 100) {
    return 'bg-green-50 dark:bg-green-900/30 border-green-400 dark:border-green-500 text-green-600 dark:text-green-400'
  }
  if (expandedStep.value === idx) {
    return 'bg-primary-50 dark:bg-primary-900/30 border-primary-400 dark:border-primary-500 text-primary-600 dark:text-primary-400'
  }
  if (progress > 0) {
    return 'bg-blue-50 dark:bg-blue-900/30 border-blue-400 dark:border-blue-500 text-blue-600 dark:text-blue-400'
  }
  return 'bg-gray-50 dark:bg-gray-800 border-gray-300 dark:border-gray-600 text-gray-500 dark:text-gray-400'
}

function difficultyClass(diff) {
  const map = {
    easy: 'bg-green-50 dark:bg-green-900/30 text-green-600 dark:text-green-400',
    medium: 'bg-amber-50 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400',
    hard: 'bg-red-50 dark:bg-red-900/30 text-red-600 dark:text-red-400'
  }
  return map[diff] || map.medium
}

function difficultyLabel(diff) {
  const map = { easy: '入门', medium: '进阶', hard: '高级' }
  return map[diff] || '进阶'
}

async function generatePath(force = false) {
  generating.value = true
  error.value = ''
  if (force) {
    loading.value = true
    steps.value = []
  }
  try {
    const res = await getLearningPath(force)
    const data = res.data
    overview.value = data.overview || ''
    totalHours.value = data.total_hours || 0
    steps.value = data.ordered_steps || []
    materialCount.value = data.material_count || 0
    expandedStep.value = steps.value.length > 0 ? 0 : null

    if (data.material_count === 0 || !data.ordered_steps || data.ordered_steps.length === 0) {
      empty.value = true
    }
  } catch (e) {
    error.value = e.response?.data?.error || '生成学习路径失败，请稍后重试'
    toast.error(error.value)
  } finally {
    if (loadingTimer) {
      clearTimeout(loadingTimer)
      loadingTimer = null
    }
    generating.value = false
    loading.value = false
  }
}

let loadingTimer = null

onMounted(() => {
  generatePath()
  loadingTimer = setTimeout(() => {
    if (loading.value) {
      loading.value = false
      if (!error.value && steps.value.length === 0) {
        error.value = 'AI 生成超时，学习路径规划需要较长时间，请点击重试'
      }
    }
  }, 120000)
})
</script>
