<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-4xl mx-auto">
    <div class="mb-6 sm:mb-8 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-gray-100">练习场</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1 text-sm sm:text-base">
          AI 根据你的学习材料生成的练习题，即时检验掌握程度
        </p>
      </div>
      <!-- 报告/答题 切换按钮 -->
      <div
        v-if="answeredCount === filteredQuizzes.length && filteredQuizzes.length > 0"
        class="flex items-center gap-2 shrink-0"
      >
        <button
          :class="
            !showReport
              ? 'bg-primary-600 text-white'
              : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'
          "
          class="px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          @click="showReport = false"
        >
          答题回顾
        </button>
        <button
          :class="
            showReport
              ? 'bg-primary-600 text-white'
              : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'
          "
          class="px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          @click="showReport = true"
        >
          练习报告
        </button>
      </div>
    </div>

    <!-- ==================== 练习报告视图 ==================== -->
    <div v-if="showReport && answeredCount === filteredQuizzes.length">
      <!-- 总分概览卡片 -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div
          class="bg-white rounded-xl shadow-sm border border-gray-100 p-5 text-center dark:bg-gray-800 dark:border-gray-700"
        >
          <p
            class="text-3xl font-bold"
            :class="accuracy >= 80 ? 'text-green-600' : accuracy >= 60 ? 'text-amber-600' : 'text-red-600'"
          >
            {{ accuracy }}%
          </p>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">正确率</p>
        </div>
        <div
          class="bg-white rounded-xl shadow-sm border border-gray-100 p-5 text-center dark:bg-gray-800 dark:border-gray-700"
        >
          <p class="text-3xl font-bold text-primary-600">{{ formatTime(reportTotalTime) }}</p>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">总用时</p>
        </div>
        <div
          class="bg-white rounded-xl shadow-sm border border-gray-100 p-5 text-center dark:bg-gray-800 dark:border-gray-700"
        >
          <p class="text-3xl font-bold text-green-600">{{ totalCorrect }}</p>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">答对</p>
        </div>
        <div
          class="bg-white rounded-xl shadow-sm border border-gray-100 p-5 text-center dark:bg-gray-800 dark:border-gray-700"
        >
          <p class="text-3xl font-bold text-red-500">{{ totalWrong }}</p>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">答错</p>
        </div>
      </div>

      <!-- 提示使用统计条 -->
      <div
        v-if="totalHintsUsed > 0"
        class="mb-4 p-3 rounded-lg bg-amber-50 border border-amber-100 flex items-center justify-between dark:bg-amber-900/20 dark:border-amber-800/30"
      >
        <div class="flex items-center gap-2 text-amber-700 dark:text-amber-400">
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path
              fill-rule="evenodd"
              d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
              clip-rule="evenodd"
            />
          </svg>
          <span class="text-sm font-medium">本轮共使用 {{ totalHintsUsed }} 次提示</span>
          <span class="text-xs text-amber-500 dark:text-amber-500">（使用提示不影响得分）</span>
        </div>
        <div class="flex items-center gap-2 text-xs text-amber-600 dark:text-amber-400">
          <span>{{ hintsUsedCount }} 题使用了提示</span>
        </div>
      </div>

      <!-- 雷达图 + 复习建议 双栏 -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
        <!-- 知识点掌握雷达图 -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 dark:bg-gray-800 dark:border-gray-700">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center gap-2">
            <svg class="w-5 h-5 text-primary-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
              />
            </svg>
            知识掌握雷达
          </h3>
          <div ref="radarChartEl" class="w-full" style="height: 280px"></div>
        </div>

        <!-- 复习建议 -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 dark:bg-gray-800 dark:border-gray-700">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center gap-2">
            <svg class="w-5 h-5 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
              />
            </svg>
            复习建议
          </h3>
          <div class="space-y-3">
            <div
              v-for="(tip, idx) in reviewSuggestions"
              :key="idx"
              class="flex items-start gap-3 p-3 rounded-lg bg-amber-50 border border-amber-100 dark:bg-amber-900/20 dark:border-amber-800/30"
            >
              <span
                class="flex-shrink-0 w-6 h-6 rounded-full bg-amber-100 dark:bg-amber-900/40 flex items-center justify-center text-amber-600 dark:text-amber-400 text-sm font-medium"
              >
                {{ idx + 1 }}
              </span>
              <p class="text-sm text-amber-800 dark:text-amber-300 leading-relaxed">{{ tip }}</p>
            </div>
            <div
              v-if="reviewSuggestions.length === 0"
              class="p-4 rounded-lg bg-green-50 border border-green-100 dark:bg-green-900/20 dark:border-green-800/30 text-center"
            >
              <p class="text-green-700 dark:text-green-400 font-medium">表现优异，继续保持！</p>
            </div>
          </div>

          <!-- 维度正确率小条 -->
          <div class="mt-6 space-y-2">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2">各维度正确率</p>
            <div v-for="stat in dimensionStats" :key="stat.label" class="flex items-center gap-3">
              <span class="text-xs text-gray-600 dark:text-gray-400 w-16 flex-shrink-0 text-right">
                {{ stat.label }}
              </span>
              <div class="flex-1 bg-gray-100 rounded-full h-2 dark:bg-gray-700 overflow-hidden">
                <div
                  class="h-2 rounded-full transition-all duration-500"
                  :class="stat.rate >= 80 ? 'bg-green-500' : stat.rate >= 60 ? 'bg-amber-500' : 'bg-red-500'"
                  :style="{ width: stat.rate + '%' }"
                ></div>
              </div>
              <span
                class="text-xs font-medium w-10"
                :class="stat.rate >= 80 ? 'text-green-600' : stat.rate >= 60 ? 'text-amber-600' : 'text-red-600'"
              >
                {{ stat.rate }}%
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 每题详细报告 -->
      <div
        class="bg-white rounded-xl shadow-sm border border-gray-100 dark:bg-gray-800 dark:border-gray-700 overflow-hidden"
      >
        <div class="p-6 border-b border-gray-100 dark:border-gray-700">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
            <svg class="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
              />
            </svg>
            逐题详情
          </h3>
        </div>
        <div class="divide-y divide-gray-50 dark:divide-gray-700">
          <div
            v-for="(q, idx) in filteredQuizzes"
            :key="q.id"
            class="p-5 hover:bg-gray-50/50 dark:hover:bg-gray-700/30 transition-colors"
          >
            <div class="flex items-start gap-4">
              <!-- 题号 + 对错 -->
              <div class="flex-shrink-0 flex flex-col items-center gap-1">
                <span
                  class="w-9 h-9 rounded-full flex items-center justify-center text-sm font-bold text-white"
                  :class="q._answerHistory?.isCorrect ? 'bg-green-500' : 'bg-red-500'"
                >
                  {{ idx + 1 }}
                </span>
                <svg
                  v-if="q._answerHistory?.isCorrect"
                  class="w-4 h-4 text-green-500"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fill-rule="evenodd"
                    d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                    clip-rule="evenodd"
                  />
                </svg>
                <svg v-else class="w-4 h-4 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fill-rule="evenodd"
                    d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                    clip-rule="evenodd"
                  />
                </svg>
              </div>

              <!-- 题目内容 -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-2 flex-wrap">
                  <span
                    class="px-2 py-0.5 bg-primary-50 text-primary-600 rounded text-xs font-medium dark:bg-primary-900/20"
                  >
                    {{ typeLabel(q.question_type) }}
                  </span>
                  <span class="px-2 py-0.5 rounded text-xs font-medium" :class="diffClass(q.difficulty)">
                    {{ diffLabel(q.difficulty) }}
                  </span>
                  <span class="text-xs text-gray-400 dark:text-gray-500 ml-auto flex items-center gap-1">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    {{ formatTime(q._answerHistory?.timeSpent || 0) }}
                  </span>
                  <span
                    v-if="q._answerHistory?.hintsUsed > 0"
                    class="text-xs text-amber-600 dark:text-amber-400 flex items-center gap-1"
                    :title="`使用了 ${q._answerHistory.hintsUsed} 级提示`"
                  >
                    <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
                      <path
                        fill-rule="evenodd"
                        d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                        clip-rule="evenodd"
                      />
                    </svg>
                    提示×{{ q._answerHistory.hintsUsed }}
                  </span>
                </div>
                <p
                  class="text-sm font-medium text-gray-900 dark:text-gray-100 mb-3 leading-relaxed"
                  v-html="renderMath(q.question)"
                ></p>

                <!-- 答题对比 -->
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 mb-2">
                  <div
                    class="p-2.5 rounded-lg text-sm"
                    :class="
                      q._answerHistory?.isCorrect ? 'bg-green-50 dark:bg-green-900/20' : 'bg-red-50 dark:bg-red-900/20'
                    "
                  >
                    <p class="text-xs text-gray-500 dark:text-gray-400 mb-0.5">你的答案</p>
                    <p
                      :class="
                        q._answerHistory?.isCorrect
                          ? 'text-green-700 dark:text-green-400'
                          : 'text-red-700 dark:text-red-400'
                      "
                      class="font-medium"
                      v-html="renderMath(q._answerHistory?.answer || '-')"
                    ></p>
                  </div>
                  <div class="p-2.5 rounded-lg bg-green-50 dark:bg-green-900/20 text-sm">
                    <p class="text-xs text-gray-500 dark:text-gray-400 mb-0.5">正确答案</p>
                    <p class="text-green-700 dark:text-green-400 font-medium" v-html="renderMath(q.answer)"></p>
                  </div>
                </div>

                <!-- 解析 -->
                <div v-if="q.explanation" class="p-2.5 rounded-lg bg-blue-50 dark:bg-blue-900/20 text-sm">
                  <p class="text-xs text-blue-600 dark:text-blue-400 font-medium mb-0.5">解析</p>
                  <p
                    class="text-blue-700 dark:text-blue-300 text-xs leading-relaxed"
                    v-html="renderMath(q.explanation)"
                  ></p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 底部操作 -->
      <div class="mt-6 flex items-center justify-center gap-4">
        <button
          class="px-5 py-2.5 text-sm text-gray-600 hover:text-gray-800 rounded-lg hover:bg-gray-100 transition dark:text-gray-400 dark:hover:text-gray-200 dark:hover:bg-gray-700"
          @click="showReport = false"
        >
          返回答题回顾
        </button>
        <button
          class="px-5 py-2.5 text-sm bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition font-medium"
          @click="resetQuiz"
        >
          重新练习
        </button>
      </div>
    </div>

    <!-- ==================== 答题视图（原有） ==================== -->
    <template v-if="!showReport">
      <!-- 模式切换：全部题目 / 智能推荐 -->
      <div class="flex items-center gap-2 mb-4">
        <button
          class="px-4 py-2 rounded-lg text-sm font-medium transition-all border"
          :class="
            studyMode === 'all'
              ? 'bg-primary-600 text-white border-primary-600'
              : 'bg-white text-gray-600 border-gray-200 hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-600 dark:hover:bg-gray-700'
          "
          @click="switchMode('all')"
        >
          全部题目
          <span class="ml-1 text-xs opacity-75">{{ quizzes.length }}</span>
        </button>
        <button
          class="px-4 py-2 rounded-lg text-sm font-medium transition-all border flex items-center gap-1.5"
          :class="
            studyMode === 'recommended'
              ? 'bg-gradient-to-r from-primary-600 to-indigo-600 text-white border-primary-600'
              : 'bg-white text-gray-600 border-gray-200 hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-600 dark:hover:bg-gray-700'
          "
          @click="switchMode('recommended')"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          智能推荐
        </button>
      </div>

      <!-- 智能推荐难度信息横幅 -->
      <div
        v-if="studyMode === 'recommended' && difficultyInfo"
        class="mb-4 p-4 rounded-xl border bg-gradient-to-r from-primary-50 to-indigo-50 dark:from-gray-800 dark:to-gray-750 dark:border-gray-700 border-primary-100"
      >
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
              <svg
                class="w-5 h-5 text-primary-600 dark:text-primary-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-900 dark:text-gray-100">{{ difficultyInfo.reason }}</p>
              <div class="flex items-center gap-1.5 mt-1">
                <span class="text-xs text-gray-500 dark:text-gray-400">推荐难度：</span>
                <span
                  v-for="d in difficultyInfo.recommended_difficulty"
                  :key="d"
                  class="px-2 py-0.5 rounded text-xs font-medium"
                  :class="diffBadgeClass(d)"
                >
                  {{ diffLabel(d) }}
                </span>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-3 text-sm">
            <div class="text-center">
              <p
                class="font-bold text-lg"
                :class="
                  difficultyInfo.accuracy >= 80
                    ? 'text-green-600'
                    : difficultyInfo.accuracy >= 50
                      ? 'text-amber-600'
                      : 'text-red-500'
                "
              >
                {{ Math.round(difficultyInfo.accuracy) }}%
              </p>
              <p class="text-xs text-gray-400 dark:text-gray-500">正确率</p>
            </div>
            <div class="w-px h-8 bg-gray-200 dark:bg-gray-600"></div>
            <div class="text-center">
              <p class="font-bold text-lg text-primary-600">{{ difficultyInfo.total_attempts }}</p>
              <p class="text-xs text-gray-400 dark:text-gray-500">已答题</p>
            </div>
          </div>
        </div>
        <!-- 各难度统计条 -->
        <div
          v-if="difficultyInfo.difficulty_stats && difficultyInfo.difficulty_stats.length > 0"
          class="mt-3 pt-3 border-t border-primary-100 dark:border-gray-700"
        >
          <div class="flex items-center gap-4 flex-wrap text-xs text-gray-500 dark:text-gray-400">
            <span v-for="ds in difficultyInfo.difficulty_stats" :key="ds.difficulty" class="flex items-center gap-1.5">
              <span class="px-1.5 py-0.5 rounded" :class="diffBadgeClass(ds.difficulty)">
                {{ diffLabel(ds.difficulty) }}
              </span>
              <span>{{ ds.correct }}/{{ ds.total }} 正确</span>
              <span
                class="font-medium"
                :class="
                  ds.total > 0
                    ? ds.correct / ds.total >= 0.8
                      ? 'text-green-600'
                      : ds.correct / ds.total >= 0.5
                        ? 'text-amber-600'
                        : 'text-red-500'
                    : ''
                "
              >
                ({{ ds.total > 0 ? Math.round((ds.correct / ds.total) * 100) : 0 }}%)
              </span>
            </span>
          </div>
        </div>
      </div>

      <!-- 题型过滤 -->
      <div v-if="quizzes.length > 0" class="flex flex-wrap items-center gap-2 mb-4">
        <span class="text-xs text-gray-500 dark:text-gray-400 mr-1">题型：</span>
        <button
          v-for="t in [
            { key: 'choice', label: '选择题' },
            { key: 'fill', label: '填空题' },
            { key: 'judge', label: '判断题' },
            { key: 'short_answer', label: '解答题' }
          ]"
          :key="t.key"
          class="px-3 py-1.5 rounded-full text-xs font-medium transition-all border"
          :class="
            selectedTypes.has(t.key)
              ? 'bg-primary-50 text-primary-700 border-primary-200 dark:bg-primary-900/30 dark:text-primary-400 dark:border-primary-700'
              : 'bg-gray-50 text-gray-400 border-gray-200 dark:bg-gray-800 dark:text-gray-500 dark:border-gray-600 line-through opacity-60'
          "
          @click="toggleType(t.key)"
        >
          {{ t.label }}
          <span class="ml-1 opacity-70">{{ typeCounts[t.key] }}</span>
        </button>
      </div>

      <!-- 进度条 + 题目导航 -->
      <div
        v-if="filteredQuizzes.length > 0"
        class="bg-white rounded-xl shadow-sm border border-gray-100 p-4 mb-6 dark:bg-gray-800 dark:border-gray-700"
      >
        <div class="flex items-center justify-between mb-3">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            进度：{{ answeredCount }} / {{ filteredQuizzes.length }}
          </span>
          <div class="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
            <span v-if="currentQuiz && !currentQuiz._answerHistory" class="flex items-center gap-1">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              {{ formatTime(elapsed) }}
            </span>
            <span class="text-green-600">正确 {{ correctCount }}</span>
            <span class="text-red-500">错误 {{ wrongCount }}</span>
          </div>
        </div>

        <!-- 可点击的题目导航点 -->
        <div class="flex items-center gap-1.5 flex-wrap mb-2">
          <button
            v-for="(q, idx) in filteredQuizzes"
            :key="q.id"
            class="w-8 h-8 rounded-lg text-xs font-medium transition-all border-2 flex items-center justify-center"
            :class="dotClass(idx)"
            :title="`第 ${idx + 1} 题${q._answerHistory ? (q._answerHistory.isCorrect ? ' ✓' : ' ✗') : ''}`"
            @click="goToQuestion(idx)"
          >
            {{ idx + 1 }}
          </button>
        </div>

        <!-- 进度条 -->
        <div class="w-full bg-gray-100 rounded-full h-1.5 dark:bg-gray-700">
          <div
            class="bg-primary-500 h-1.5 rounded-full transition-all duration-300"
            :style="{ width: `${(answeredCount / filteredQuizzes.length) * 100}%` }"
          ></div>
        </div>
      </div>

      <!-- 空状态 -->
      <div
        v-if="filteredQuizzes.length === 0 && !loading"
        class="bg-white rounded-xl shadow-sm border border-gray-100 p-16 text-center dark:bg-gray-800 dark:border-gray-700"
      >
        <p class="text-5xl mb-4">{{ studyMode === 'recommended' ? '🎯' : '✏️' }}</p>
        <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">
          {{ studyMode === 'recommended' ? '暂无推荐题目' : '还没有练习题' }}
        </h3>
        <p class="text-gray-400 dark:text-gray-500 mb-4">
          {{
            studyMode === 'recommended'
              ? '当前推荐难度下暂无题目，试试切换到「全部题目」或上传更多材料'
              : '上传学习材料并且分析后，AI 会自动生成练习题'
          }}
        </p>
        <div class="flex items-center justify-center gap-3">
          <button
            v-if="studyMode === 'recommended'"
            class="px-4 py-2 text-sm text-primary-600 hover:text-primary-700 border border-primary-200 hover:bg-primary-50 rounded-lg transition dark:text-primary-400 dark:border-primary-700 dark:hover:bg-primary-900/20"
            @click="switchMode('all')"
          >
            切换到全部题目
          </button>
          <router-link v-else to="/upload" class="text-primary-600 hover:underline text-sm">去上传材料</router-link>
        </div>
      </div>

      <div v-if="loading" class="text-center py-12 text-gray-400 dark:text-gray-500">加载中...</div>

      <!-- 当前题目 -->
      <div
        v-if="filteredQuizzes.length > 0 && currentQuiz"
        class="bg-white rounded-xl shadow-sm border border-gray-100 p-8 dark:bg-gray-800 dark:border-gray-700"
      >
        <!-- 题型标签 -->
        <div class="flex items-center gap-2 mb-4">
          <span class="px-2 py-1 bg-primary-50 text-primary-600 rounded text-xs font-medium dark:bg-primary-900/20">
            {{ typeLabel(currentQuiz.question_type) }}
          </span>
          <span class="px-2 py-1 rounded text-xs font-medium" :class="diffClass(currentQuiz.difficulty)">
            {{ diffLabel(currentQuiz.difficulty) }}
          </span>
          <!-- 材料标签 -->
          <template v-if="quizMaterialTagMap[currentQuiz.material_id]?.length">
            <span
              v-for="mt in quizMaterialTagMap[currentQuiz.material_id]"
              :key="'mt-' + mt"
              class="px-1.5 py-0.5 rounded text-[10px] bg-indigo-50 text-indigo-500 dark:bg-indigo-900/20 dark:text-indigo-400"
            >
              {{ mt }}
            </span>
          </template>
          <span v-if="currentQuiz._answerHistory" class="text-xs text-gray-400 dark:text-gray-500 ml-auto">
            已答 · 用时 {{ formatTime(currentQuiz._answerHistory.timeSpent) }}
          </span>
        </div>

        <!-- 题目 -->
        <h2
          class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-6 leading-relaxed"
          v-html="renderMath(currentQuiz.question)"
        ></h2>

        <!-- 已答题 - 显示历史答案 -->
        <div v-if="currentQuiz._answerHistory" class="space-y-4">
          <!-- 选择题：选项回顾 -->
          <div v-if="currentQuiz.type === 'choice'" class="space-y-3">
            <div
              v-for="(opt, idx) in currentQuiz.options"
              :key="idx"
              class="w-full text-left p-4 rounded-lg border-2"
              :class="reviewOptionClass(idx)"
            >
              <span class="font-medium mr-2">{{ String.fromCharCode(65 + idx) }}.</span>
              <span v-html="renderMath(opt)"></span>
            </div>
          </div>

          <!-- 判断题：正确/错误按钮回顾 -->
          <div v-else-if="currentQuiz.type === 'judge'" class="grid grid-cols-2 gap-3">
            <button
              v-for="val in ['正确', '错误']"
              :key="val"
              class="p-4 rounded-lg border-2 text-center font-semibold text-sm transition-all cursor-default"
              :class="judgeClass(val)"
            >
              <span class="text-lg">{{ val === '正确' ? '✓' : '✗' }}</span>
              <span class="ml-2">{{ val }}</span>
            </button>
          </div>

          <!-- 填空题/解答题：文本答案回顾 -->
          <div v-else class="p-4 bg-gray-50 rounded-lg border border-gray-200 dark:bg-gray-700 dark:border-gray-600">
            <p class="text-sm text-gray-500 dark:text-gray-400 mb-1">你的答案：</p>
            <p class="text-gray-800 dark:text-gray-200" v-html="renderMath(currentQuiz._answerHistory.answer)"></p>
          </div>

          <!-- 答题结果 -->
          <div
            class="p-4 rounded-lg"
            :class="
              currentQuiz._answerHistory.isCorrect
                ? 'bg-green-50 border border-green-200 dark:bg-green-900/20 dark:border-green-700'
                : 'bg-red-50 border border-red-200 dark:bg-red-900/20 dark:border-red-700'
            "
          >
            <p
              class="font-medium mb-1"
              :class="
                currentQuiz._answerHistory.isCorrect
                  ? 'text-green-700 dark:text-green-400'
                  : 'text-red-700 dark:text-red-400'
              "
            >
              {{ currentQuiz._answerHistory.isCorrect ? '✓ 回答正确！' : '✗ 回答错误' }}
            </p>
            <p class="text-sm" :class="currentQuiz._answerHistory.isCorrect ? 'text-green-600' : 'text-red-600'">
              正确答案：
              <span v-html="renderMath(currentQuiz.answer)"></span>
            </p>
            <!-- 使用提示数量标记 -->
            <p
              v-if="currentQuiz._answerHistory.hintsUsed > 0"
              class="text-xs mt-1.5 flex items-center gap-1 text-amber-600 dark:text-amber-400"
            >
              <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                  clip-rule="evenodd"
                />
              </svg>
              使用了 {{ currentQuiz._answerHistory.hintsUsed }} 级提示
            </p>
          </div>

          <!-- 答错时服务端返回的提示 -->
          <div
            v-if="currentQuiz._answerHistory.hintFromResponse && !currentQuiz._answerHistory.isCorrect"
            class="p-3 rounded-lg border border-amber-200 bg-amber-50 dark:border-amber-800/40 dark:bg-amber-900/20"
          >
            <p class="text-xs font-semibold text-amber-700 dark:text-amber-400 mb-1 flex items-center gap-1.5">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                />
              </svg>
              提示参考
            </p>
            <p
              class="text-sm text-amber-800 dark:text-amber-300"
              v-html="renderMath(currentQuiz._answerHistory.hintFromResponse)"
            ></p>
          </div>

          <!-- 解析 -->
          <div
            v-if="currentQuiz.explanation"
            class="p-4 bg-blue-50 rounded-lg border border-blue-100 dark:bg-blue-900/20"
          >
            <p class="text-sm font-medium text-blue-700 mb-1">💡 解析</p>
            <p class="text-sm text-blue-600" v-html="renderMath(currentQuiz.explanation)"></p>
          </div>
        </div>

        <!-- 未答题 - 正常答题界面 -->
        <div v-else>
          <!-- 选择题：选项按钮 -->
          <div v-if="currentQuiz.type === 'choice'" class="space-y-3 mb-6">
            <button
              v-for="(opt, idx) in currentQuiz.options"
              :key="idx"
              class="w-full text-left p-4 rounded-lg border-2 transition-all"
              :class="optionClass(idx)"
              @click="selectOption(idx)"
            >
              <span class="font-medium mr-2">{{ String.fromCharCode(65 + idx) }}.</span>
              <span v-html="renderMath(opt)"></span>
            </button>
          </div>

          <!-- 判断题：正确/错误按钮 -->
          <div v-else-if="currentQuiz.type === 'judge'" class="grid grid-cols-2 gap-4 mb-6">
            <button
              class="p-5 rounded-xl border-2 text-center font-semibold transition-all"
              :class="
                selectedJudge === '正确'
                  ? 'border-green-500 bg-green-50 text-green-700 dark:bg-green-900/20 dark:text-green-400 dark:border-green-600'
                  : 'border-gray-200 hover:border-green-300 hover:bg-green-50/50 dark:border-gray-600 dark:hover:border-green-700 dark:hover:bg-green-900/10'
              "
              @click="selectJudge('正确')"
            >
              <span class="text-2xl block mb-1">✓</span>
              <span class="text-sm">正确</span>
            </button>
            <button
              class="p-5 rounded-xl border-2 text-center font-semibold transition-all"
              :class="
                selectedJudge === '错误'
                  ? 'border-red-500 bg-red-50 text-red-700 dark:bg-red-900/20 dark:text-red-400 dark:border-red-600'
                  : 'border-gray-200 hover:border-red-300 hover:bg-red-50/50 dark:border-gray-600 dark:hover:border-red-700 dark:hover:bg-red-900/10'
              "
              @click="selectJudge('错误')"
            >
              <span class="text-2xl block mb-1">✗</span>
              <span class="text-sm">错误</span>
            </button>
          </div>

          <!-- 填空题：单行输入 -->
          <div v-else-if="currentQuiz.type === 'fill'" class="mb-6">
            <input
              v-model="textAnswer"
              type="text"
              class="w-full px-4 py-3 rounded-lg border border-gray-200 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition text-base dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200"
              placeholder="填写答案..."
            />
          </div>

          <!-- 解答题：多行文本 -->
          <div v-else class="mb-6">
            <textarea
              v-model="textAnswer"
              rows="5"
              class="w-full px-4 py-3 rounded-lg border border-gray-200 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none resize-y dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200"
              placeholder="输入你的解答..."
            ></textarea>
          </div>

          <!-- 提交 + 提示按钮行 -->
          <div class="flex items-center gap-3 flex-wrap">
            <button
              :disabled="!canSubmit"
              class="px-6 py-3 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
              @click="submitAnswer"
            >
              提交答案
            </button>

            <!-- 需要提示按钮 -->
            <button
              v-if="currentQuiz._hintLevel < 3"
              :disabled="currentQuiz._hintLoading"
              class="px-4 py-3 text-sm font-medium rounded-lg border-2 border-dashed transition-all flex items-center gap-2 disabled:opacity-50 border-amber-300 text-amber-700 hover:border-amber-400 hover:bg-amber-50 dark:border-amber-700 dark:text-amber-400 dark:hover:border-amber-600 dark:hover:bg-amber-900/20"
              @click="requestHint"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                />
              </svg>
              <template v-if="currentQuiz._hintLoading">加载中...</template>
              <template v-else-if="currentQuiz._hintLevel === 0">需要提示？</template>
              <template v-else-if="currentQuiz._hintLevel === 1">再给一点提示</template>
              <template v-else-if="currentQuiz._hintLevel === 2">最后提示</template>
            </button>

            <!-- 已使用提示数标记 -->
            <span
              v-if="currentQuiz._hintLevel > 0"
              class="text-xs text-amber-600 dark:text-amber-400 flex items-center gap-1"
            >
              <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                  clip-rule="evenodd"
                />
              </svg>
              已使用 {{ currentQuiz._hintLevel }} 级提示（不影响得分）
            </span>
          </div>

          <!-- 渐进式提示展示区 -->
          <div v-if="currentQuiz._hintLevel > 0" class="mt-4 space-y-2">
            <!-- 第1级：概念方向 -->
            <div
              v-if="currentQuiz._hints.hint_1"
              class="p-3 rounded-lg border border-amber-200 bg-amber-50 dark:border-amber-800/40 dark:bg-amber-900/20"
            >
              <p class="text-xs font-semibold text-amber-700 dark:text-amber-400 mb-1 flex items-center gap-1.5">
                <span
                  class="inline-flex w-5 h-5 items-center justify-center rounded-full bg-amber-200 dark:bg-amber-800/40 text-amber-800 dark:text-amber-300 text-xs font-bold"
                >
                  1
                </span>
                方向提示
              </p>
              <p
                class="text-sm text-amber-800 dark:text-amber-300 leading-relaxed"
                v-html="renderMath(currentQuiz._hints.hint_1)"
              ></p>
            </div>

            <!-- 第2级：关键线索 -->
            <div
              v-if="currentQuiz._hintLevel >= 2 && currentQuiz._hints.hint_2"
              class="p-3 rounded-lg border border-orange-200 bg-orange-50 dark:border-orange-800/40 dark:bg-orange-900/20"
            >
              <p class="text-xs font-semibold text-orange-700 dark:text-orange-400 mb-1 flex items-center gap-1.5">
                <span
                  class="inline-flex w-5 h-5 items-center justify-center rounded-full bg-orange-200 dark:bg-orange-800/40 text-orange-800 dark:text-orange-300 text-xs font-bold"
                >
                  2
                </span>
                关键线索
              </p>
              <p
                class="text-sm text-orange-800 dark:text-orange-300 leading-relaxed"
                v-html="renderMath(currentQuiz._hints.hint_2)"
              ></p>
            </div>

            <!-- 第3级：接近答案 -->
            <div
              v-if="currentQuiz._hintLevel >= 3 && currentQuiz._hints.hint_3"
              class="p-3 rounded-lg border border-red-200 bg-red-50 dark:border-red-800/40 dark:bg-red-900/20"
            >
              <p class="text-xs font-semibold text-red-700 dark:text-red-400 mb-1 flex items-center gap-1.5">
                <span
                  class="inline-flex w-5 h-5 items-center justify-center rounded-full bg-red-200 dark:bg-red-800/40 text-red-800 dark:text-red-300 text-xs font-bold"
                >
                  3
                </span>
                接近答案
              </p>
              <p
                class="text-sm text-red-800 dark:text-red-300 leading-relaxed"
                v-html="renderMath(currentQuiz._hints.hint_3)"
              ></p>
            </div>

            <!-- 某级别没有提示时的占位 -->
            <div
              v-if="currentQuiz._hintLevel >= 1 && !currentQuiz._hints.hint_1"
              class="p-3 rounded-lg border border-gray-200 bg-gray-50 text-center dark:border-gray-700 dark:bg-gray-700/30"
            >
              <p class="text-sm text-gray-400 dark:text-gray-500">此题暂无提示，加油思考！</p>
            </div>
          </div>
        </div>

        <!-- 导航按钮 -->
        <div class="flex items-center justify-between mt-6 pt-4 border-t border-gray-100 dark:border-gray-700">
          <button
            v-if="currentIndex > 0"
            class="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 rounded-lg hover:bg-gray-50 transition flex items-center gap-1 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:bg-gray-700"
            @click="goToQuestion(currentIndex - 1)"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            上一题
          </button>
          <span v-else></span>

          <!-- 全部完成 -->
          <div
            v-if="answeredCount === filteredQuizzes.length"
            class="flex flex-col sm:flex-row items-center gap-2 sm:gap-3"
          >
            <div
              class="p-3 bg-green-50 rounded-lg border border-green-200 text-center dark:bg-green-900/20 dark:border-green-700"
            >
              <p class="text-sm font-bold text-green-700 dark:text-green-400">
                🎉 全部完成！正确率 {{ accuracy }}%（用时 {{ formatTime(reportTotalTime) }}）
              </p>
            </div>
            <button
              class="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white text-sm font-medium rounded-lg transition-colors whitespace-nowrap"
              @click="showReport = true"
            >
              查看报告
            </button>
          </div>

          <button
            v-if="currentIndex < filteredQuizzes.length - 1"
            class="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 rounded-lg hover:bg-gray-50 transition flex items-center gap-1 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:bg-gray-700"
            @click="goToQuestion(currentIndex + 1)"
          >
            下一题
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>
          <span v-else></span>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { listQuizzes, answerQuiz, getQuizHint, listMaterials, getDifficultyLevel } from '../api/client'
import * as echarts from 'echarts'
import { useDarkMode } from '../composables/useDarkMode'
import { renderMath } from '../composables/useMathRender'

const { isDark } = useDarkMode()

const quizzes = ref([])
const loading = ref(true)
const currentIndex = ref(0)
const selectedOption = ref(-1)
const selectedJudge = ref('') // '正确' or '错误'
const textAnswer = ref('')
const correctCount = ref(0)
const wrongCount = ref(0)
const elapsed = ref(0)
const showReport = ref(false)
const radarChartEl = ref(null)
let radarChart = null
let timerInterval = null

// 智能推荐模式
const studyMode = ref('all') // 'all' | 'recommended'
const difficultyInfo = ref(null)

// 材料标签映射
const quizMaterialTagMap = ref({})

// 题型过滤：默认全部启用
const selectedTypes = ref(new Set(['choice', 'fill', 'judge', 'short_answer']))

// 过滤后的题目列表
const filteredQuizzes = computed(() => {
  if (selectedTypes.value.size === 4) return quizzes.value
  return quizzes.value.filter((q) => selectedTypes.value.has(q.type))
})

// 各题型计数
const typeCounts = computed(() => {
  const counts = { choice: 0, fill: 0, judge: 0, short_answer: 0 }
  for (const q of quizzes.value) {
    if (q.type in counts) counts[q.type]++
  }
  return counts
})

function toggleType(type) {
  const s = new Set(selectedTypes.value)
  if (s.has(type)) {
    if (s.size > 1) s.delete(type) // 至少保留一种
  } else {
    s.add(type)
  }
  selectedTypes.value = s
  // 过滤变化时重置导航
  if (currentIndex.value >= filteredQuizzes.value.length) {
    currentIndex.value = Math.max(0, filteredQuizzes.value.length - 1)
  }
  resetInputState()
}

const currentQuiz = computed(() => filteredQuizzes.value[currentIndex.value] || null)
const answeredCount = computed(() => filteredQuizzes.value.filter((q) => q._answerHistory).length)
const totalTime = computed(() => filteredQuizzes.value.reduce((sum, q) => sum + (q._answerHistory?.timeSpent || 0), 0))
const totalHintsUsed = computed(() =>
  filteredQuizzes.value.reduce((sum, q) => sum + (q._answerHistory?.hintsUsed || 0), 0)
)
const hintsUsedCount = computed(() => filteredQuizzes.value.filter((q) => q._answerHistory?.hintsUsed > 0).length)

const canSubmit = computed(() => {
  if (!currentQuiz.value || currentQuiz.value._answerHistory) return false
  const t = currentQuiz.value.type
  if (t === 'choice') return selectedOption.value >= 0
  if (t === 'judge') return selectedJudge.value !== ''
  return textAnswer.value.trim().length > 0
})

// ===== 报告数据 =====
const totalCorrect = computed(() => filteredQuizzes.value.filter((q) => q._answerHistory?.isCorrect).length)
const totalWrong = computed(
  () => filteredQuizzes.value.filter((q) => q._answerHistory && !q._answerHistory.isCorrect).length
)
const accuracy = computed(() => {
  if (filteredQuizzes.value.length === 0) return 0
  return Math.round((totalCorrect.value / filteredQuizzes.value.length) * 100)
})
const reportTotalTime = computed(() => totalTime.value)
const avgTimePerQuestion = computed(() => {
  if (filteredQuizzes.value.length === 0) return 0
  return Math.round(reportTotalTime.value / filteredQuizzes.value.length)
})

// 按难度统计
const difficultyStats = computed(() => {
  const levels = ['easy', 'medium', 'hard']
  return levels.map((level) => {
    const qs = filteredQuizzes.value.filter((q) => q.difficulty === level && q._answerHistory)
    if (qs.length === 0) return { level, total: 0, correct: 0, rate: 0 }
    const correct = qs.filter((q) => q._answerHistory.isCorrect).length
    return { level, total: qs.length, correct, rate: Math.round((correct / qs.length) * 100) }
  })
})

// 按题型统计（含判断题）
const typeStats = computed(() => {
  const types = ['choice', 'fill', 'judge', 'short_answer']
  return types.map((type) => {
    const qs = filteredQuizzes.value.filter((q) => q.type === type && q._answerHistory)
    if (qs.length === 0) return { type, total: 0, correct: 0, rate: 0 }
    const correct = qs.filter((q) => q._answerHistory.isCorrect).length
    return { type, total: qs.length, correct, rate: Math.round((correct / qs.length) * 100) }
  })
})

// 维度正确率（合并难度+题型，用于右侧进度条）
const dimensionStats = computed(() => {
  const stats = []
  const diffLabels = { easy: '简单', medium: '中等', hard: '困难' }
  const typeLabels = { choice: '选择题', fill: '填空题', judge: '判断题', short_answer: '简答' }

  for (const s of difficultyStats.value) {
    if (s.total > 0) stats.push({ label: diffLabels[s.level] || s.level, rate: s.rate })
  }
  for (const s of typeStats.value) {
    if (s.total > 0) stats.push({ label: typeLabels[s.type] || s.type, rate: s.rate })
  }
  return stats
})

// 复习建议
const reviewSuggestions = computed(() => {
  const suggestions = []

  if (accuracy.value < 50) {
    suggestions.push('整体正确率低于 50%，建议重新复习相关材料后再来挑战')
  } else if (accuracy.value < 70) {
    suggestions.push('正确率在 50%-70% 之间，部分知识点还需加强练习')
  } else if (accuracy.value >= 90) {
    suggestions.push('正确率高达 ' + accuracy.value + '%，掌握得很好！可以尝试更高难度的题目')
  }

  // 难度维度建议
  const diffLabels = { easy: '简单题', medium: '中等题', hard: '困难题' }
  for (const stat of difficultyStats.value) {
    if (stat.total > 0 && stat.rate < 60) {
      suggestions.push(`${diffLabels[stat.level]}正确率仅 ${stat.rate}%，建议先从基础概念入手巩固`)
    }
  }

  // 题型维度建议
  const typeLabels = { choice: '选择题', fill: '填空题', judge: '判断题', short_answer: '简答题' }
  for (const stat of typeStats.value) {
    if (stat.total > 0 && stat.rate < 60) {
      suggestions.push(`${typeLabels[stat.type]}表现不佳（${stat.rate}%），建议多做同类练习强化理解`)
    }
  }

  // 答题速度建议
  if (avgTimePerQuestion.value < 10 && filteredQuizzes.value.length > 2) {
    suggestions.push('平均每题用时不到 10 秒，答题可能过于仓促，建议仔细审题')
  }

  // 错题集中建议
  if (totalWrong.value > 0 && totalWrong.value <= 3) {
    suggestions.push(`仅有 ${totalWrong.value} 道错题，针对性复习这些题目涉及的知识点即可`)
  } else if (totalWrong.value > 3) {
    suggestions.push(`${totalWrong.value} 道错题较多，建议按题型和难度分类进行专项突破`)
  }

  return suggestions
})

// ===== 雷达图 =====
function initRadarChart() {
  if (!radarChartEl.value) return

  if (radarChart) {
    radarChart.dispose()
  }

  radarChart = echarts.init(radarChartEl.value, isDark.value ? 'dark' : undefined)

  // 构建雷达维度（只显示有题目的维度）
  const indicators = []
  const values = []
  const diffLabels = { easy: '简单', medium: '中等', hard: '困难' }
  const typeLabels = { choice: '选择题', fill: '填空题', judge: '判断题', short_answer: '简答题' }

  for (const stat of difficultyStats.value) {
    if (stat.total > 0) {
      indicators.push({ name: diffLabels[stat.level] || stat.level, max: 100 })
      values.push(stat.rate)
    }
  }
  for (const stat of typeStats.value) {
    if (stat.total > 0) {
      indicators.push({ name: typeLabels[stat.type] || stat.type, max: 100 })
      values.push(stat.rate)
    }
  }

  // 至少 3 个维度才适合雷达图
  if (indicators.length < 3) {
    indicators.push({ name: '综合', max: 100 })
    values.push(accuracy.value)
  }

  const darkAxisColor = 'rgba(255,255,255,0.4)'
  const darkSplitColor = 'rgba(255,255,255,0.08)'
  const lightAxisColor = 'rgba(0,0,0,0.35)'
  const lightSplitColor = 'rgba(0,0,0,0.06)'
  const axisColor = isDark.value ? darkAxisColor : lightAxisColor
  const splitColor = isDark.value ? darkSplitColor : lightSplitColor

  radarChart.setOption({
    radar: {
      indicator: indicators,
      shape: 'polygon',
      splitNumber: 4,
      name: {
        textStyle: { color: axisColor, fontSize: 12 }
      },
      splitLine: { lineStyle: { color: splitColor } },
      splitArea: { show: false },
      axisLine: { lineStyle: { color: splitColor } }
    },
    series: [
      {
        type: 'radar',
        data: [
          {
            value: values,
            name: '正确率',
            symbol: 'circle',
            symbolSize: 6,
            lineStyle: { color: '#6366f1', width: 2 },
            itemStyle: { color: '#6366f1' },
            areaStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: 'rgba(99, 102, 241, 0.35)' },
                { offset: 1, color: 'rgba(99, 102, 241, 0.05)' }
              ])
            },
            label: {
              show: true,
              formatter: '{c}%',
              fontSize: 10,
              color: axisColor
            }
          }
        ]
      }
    ],
    tooltip: {
      trigger: 'item',
      formatter(params) {
        const names = indicators.map((i) => i.name)
        let html = '<div style="font-size:13px">'
        params.value.forEach((v, i) => {
          const color = v >= 80 ? '#22c55e' : v >= 60 ? '#f59e0b' : '#ef4444'
          html += `<div>${names[i]}：<span style="color:${color};font-weight:bold">${v}%</span></div>`
        })
        html += '</div>'
        return html
      }
    }
  })
}

// 监听 showReport 变化，切换到报告视图时渲染图表
watch(showReport, (val) => {
  if (val) {
    nextTick(() => {
      initRadarChart()
    })
  }
})

// 监听暗色模式切换，重新渲染图表
watch(isDark, () => {
  if (showReport.value) {
    nextTick(() => {
      initRadarChart()
    })
  }
})

// 监听全部完成后自动渲染图表
watch(answeredCount, (val) => {
  if (val === filteredQuizzes.value.length && filteredQuizzes.value.length > 0 && showReport.value) {
    nextTick(() => {
      initRadarChart()
    })
  }
})

function formatTime(seconds) {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  if (m > 0) return `${m}分${String(s).padStart(2, '0')}秒`
  return `${s}秒`
}

// 启动当前题目计时
function startTimer() {
  if (currentQuiz.value && !currentQuiz.value._answerHistory) {
    currentQuiz.value._startTime = Math.floor(Date.now() / 1000)
  }
  elapsed.value = 0
  clearInterval(timerInterval)
  timerInterval = setInterval(() => {
    if (currentQuiz.value?._startTime && !currentQuiz.value._answerHistory) {
      elapsed.value = Math.floor(Date.now() / 1000) - currentQuiz.value._startTime
    }
  }, 1000)
}

function typeLabel(t) {
  return { choice: '选择题', fill: '填空题', judge: '判断题', short_answer: '解答题' }[t] || '题目'
}
function diffClass(d) {
  return (
    {
      easy: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
      medium: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
      hard: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    }[d] || 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
  )
}
function diffLabel(d) {
  return { easy: '简单', medium: '中等', hard: '困难' }[d] || ''
}

// 难度徽章样式
function diffBadgeClass(d) {
  return (
    {
      easy: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
      medium: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
      hard: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    }[d] || 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
  )
}

// 加载难度推荐信息
async function loadDifficultyInfo() {
  try {
    const res = await getDifficultyLevel()
    difficultyInfo.value = res.data
  } catch (e) {
    console.error('获取难度信息失败:', e)
  }
}

// 切换模式
async function switchMode(mode) {
  studyMode.value = mode
  showReport.value = false
  currentIndex.value = 0
  correctCount.value = 0
  wrongCount.value = 0
  resetInputState()
  loading.value = true

  try {
    if (mode === 'recommended') {
      await loadDifficultyInfo()
      const res = await listQuizzes({ recommended: 'true', limit: 50 })
      quizzes.value = (res.data.data || []).map((q) => ({
        ...q,
        question_type: q.type,
        options: typeof q.options === 'string' ? JSON.parse(q.options || '[]') : q.options || [],
        _answerHistory: null,
        _startTime: null,
        _hintLevel: 0,
        _hints: {},
        _hintLoading: false
      }))
    } else {
      const [quizRes, matRes] = await Promise.all([listQuizzes({ limit: 200 }), listMaterials({ limit: 200 })])
      quizzes.value = (quizRes.data.data || []).map((q) => ({
        ...q,
        question_type: q.type,
        options: typeof q.options === 'string' ? JSON.parse(q.options || '[]') : q.options || [],
        _answerHistory: null,
        _startTime: null,
        _hintLevel: 0,
        _hints: {},
        _hintLoading: false
      }))
      // 重建材料标签映射
      const map = {}
      for (const m of matRes.data.data || []) {
        if (m.tags) {
          map[m.id] = m.tags
            .split(',')
            .map((t) => t.trim())
            .filter(Boolean)
        }
      }
      quizMaterialTagMap.value = map
    }
  } catch (e) {
    console.error('题目加载失败:', e)
  } finally {
    loading.value = false
  }
  startTimer()
}

// 进度点样式
function dotClass(idx) {
  const q = filteredQuizzes.value[idx]
  if (idx === currentIndex.value) {
    if (q._answerHistory) {
      return q._answerHistory.isCorrect
        ? 'border-green-500 bg-green-50 text-green-700'
        : 'border-red-500 bg-red-50 text-red-700'
    }
    return 'border-primary-500 bg-primary-50 text-primary-700'
  }
  if (q._answerHistory) {
    return q._answerHistory.isCorrect
      ? 'border-green-300 bg-green-50 text-green-600'
      : 'border-red-300 bg-red-50 text-red-600'
  }
  return 'border-gray-200 text-gray-400 hover:border-gray-300'
}

// 未答题选项样式
function optionClass(idx) {
  return selectedOption.value === idx
    ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
    : 'border-gray-200 hover:border-gray-300 dark:border-gray-600 dark:hover:border-gray-500'
}

// 已答题选项样式（回看）
function reviewOptionClass(idx) {
  const correct = currentQuiz.value.options?.indexOf(currentQuiz.value.answer)
  const selected = currentQuiz.value._answerHistory?.selectedIndex
  if (idx === correct) return 'border-green-500 bg-green-50'
  if (idx === selected && idx !== correct) return 'border-red-500 bg-red-50'
  return 'border-gray-200 opacity-60 dark:border-gray-600'
}

function selectOption(idx) {
  if (currentQuiz.value?._answerHistory) return
  selectedOption.value = idx
}

function selectJudge(val) {
  if (currentQuiz.value?._answerHistory) return
  selectedJudge.value = val
}

// 判断题按钮样式（已答题回顾）
function judgeClass(val) {
  const correct = currentQuiz.value.answer
  const selected = currentQuiz.value._answerHistory?.answer
  if (val === correct && val === selected) return 'border-green-500 bg-green-50 dark:bg-green-900/20'
  if (val === correct) return 'border-green-500 bg-green-50 dark:bg-green-900/20'
  if (val === selected && val !== correct) return 'border-red-500 bg-red-50 dark:bg-red-900/20'
  return 'border-gray-200 opacity-60 dark:border-gray-600'
}

// 重置输入状态（切题/切类型时调用）
function resetInputState() {
  selectedOption.value = -1
  selectedJudge.value = ''
  textAnswer.value = ''
}

// 跳转到指定题目
function goToQuestion(idx) {
  if (idx < 0 || idx >= filteredQuizzes.value.length) return
  currentIndex.value = idx
  resetInputState()
  startTimer()
}

// 重置练习
function resetQuiz() {
  filteredQuizzes.value.forEach((q) => {
    q._answerHistory = null
    q._startTime = null
    q._hintLevel = 0
    q._hints = {}
    q._hintLoading = false
  })
  correctCount.value = 0
  wrongCount.value = 0
  currentIndex.value = 0
  resetInputState()
  showReport.value = false
  elapsed.value = 0
  clearInterval(timerInterval)
  startTimer()
}

// 请求渐进式提示
async function requestHint() {
  const quiz = currentQuiz.value
  if (!quiz || quiz._answerHistory || quiz._hintLoading) return
  if (quiz._hintLevel >= 3) return // 最多3级

  quiz._hintLoading = true
  const nextLevel = quiz._hintLevel + 1

  try {
    const res = await getQuizHint(quiz.id, nextLevel)
    quiz._hints = res.data
    quiz._hintLevel = nextLevel
  } catch (e) {
    console.error('获取提示失败:', e)
  } finally {
    quiz._hintLoading = false
  }
}

async function submitAnswer() {
  if (!canSubmit.value) return

  let answer = ''
  let selectedIndex = -1
  const quizType = currentQuiz.value.type

  if (quizType === 'choice') {
    answer = String.fromCharCode(65 + selectedOption.value)
    selectedIndex = selectedOption.value
  } else if (quizType === 'judge') {
    answer = selectedJudge.value
  } else {
    answer = textAnswer.value
  }

  // 计算用时
  const timeSpent = currentQuiz.value._startTime ? Math.floor(Date.now() / 1000) - currentQuiz.value._startTime : 0

  let isCorrectResult = false
  let responseHint1 = null // 答错时服务端返回的 hint_1

  try {
    const res = await answerQuiz(currentQuiz.value.id, {
      answer,
      hints_used: currentQuiz.value._hintLevel
    })
    isCorrectResult = res.data.is_correct
    // 答错时可能返回 hint_1
    if (!res.data.is_correct && res.data.hint_1) {
      responseHint1 = res.data.hint_1
    }
  } catch (e) {
    const correct = currentQuiz.value.answer
    if (quizType === 'choice') {
      isCorrectResult = answer === correct || answer === correct?.charAt(0)
    } else if (quizType === 'judge') {
      isCorrectResult = answer === correct
    } else {
      isCorrectResult = answer.toLowerCase().includes(correct?.toLowerCase().slice(0, 10))
    }
  }

  // 记录答题历史
  currentQuiz.value._answerHistory = {
    answer,
    selectedIndex,
    isCorrect: isCorrectResult,
    timeSpent,
    hintsUsed: currentQuiz.value._hintLevel,
    hintFromResponse: responseHint1
  }

  if (isCorrectResult) correctCount.value++
  else wrongCount.value++

  // 在推荐模式下，每次答题后异步刷新难度信息
  if (studyMode.value === 'recommended') {
    loadDifficultyInfo()
  }

  clearInterval(timerInterval)
  elapsed.value = 0

  // 全部答完后自动显示报告（延迟 2 秒让用户看完最后一题的解析）
  if (answeredCount.value >= filteredQuizzes.value.length) {
    setTimeout(() => {
      showReport.value = true
    }, 2000)
  }
}

onMounted(async () => {
  try {
    const [quizRes, matRes, diffRes] = await Promise.all([
      listQuizzes({ limit: 200 }),
      listMaterials({ limit: 200 }),
      getDifficultyLevel().catch(() => null)
    ])
    quizzes.value = (quizRes.data.data || []).map((q) => ({
      ...q,
      question_type: q.type,
      options: typeof q.options === 'string' ? JSON.parse(q.options || '[]') : q.options || [],
      _answerHistory: null,
      _startTime: null,
      _hintLevel: 0, // 当前已请求的提示级别（0=未请求）
      _hints: {}, // { hint_1, hint_2, hint_3, available: {1,2,3} }
      _hintLoading: false
    }))
    // 构建材料标签映射
    const map = {}
    for (const m of matRes.data.data || []) {
      if (m.tags) {
        map[m.id] = m.tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean)
      }
    }
    quizMaterialTagMap.value = map
    // 加载难度推荐信息
    if (diffRes) {
      difficultyInfo.value = diffRes.data
    }
  } catch (e) {
    console.error('题目加载失败:', e)
  } finally {
    loading.value = false
  }
  startTimer()
})

onUnmounted(() => {
  clearInterval(timerInterval)
  if (radarChart) {
    radarChart.dispose()
    radarChart = null
  }
})
</script>

<style scoped>
/* KaTeX 暗色模式适配 */
:deep(.katex) {
  color: inherit;
}
</style>
