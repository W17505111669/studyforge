<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-6xl mx-auto">
    <!-- 标题区 -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-6 sm:mb-8 gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">错题本</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">自动收集答错的题目，针对性复习提升</p>
      </div>
      <div class="flex items-center gap-2 sm:gap-3 flex-wrap">
        <!-- 错题重练按钮 -->
        <button
          v-if="(stats?.unreviewed || 0) > 0 && !retryMode && !consolidateMode"
          :disabled="retryLoading"
          class="flex items-center gap-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white text-sm font-medium rounded-lg transition disabled:opacity-50"
          @click="startRetry"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
            />
          </svg>
          错题重练 ({{ stats?.unreviewed || 0 }})
        </button>
        <!-- 巩固强化按钮 -->
        <button
          v-if="(stats?.total || 0) > 0 && !retryMode && !consolidateMode"
          :disabled="consolidateLoading"
          class="flex items-center gap-2 px-4 py-2 bg-violet-600 hover:bg-violet-700 text-white text-sm font-medium rounded-lg transition disabled:opacity-50"
          @click="startConsolidate"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          {{ consolidateLoading ? '分析中...' : '巩固强化' }}
        </button>
        <!-- 退出重练 -->
        <button
          v-if="retryMode || consolidateMode"
          class="flex items-center gap-2 px-4 py-2 bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 text-sm font-medium rounded-lg transition"
          @click="retryMode ? exitRetry() : exitConsolidate()"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          {{ retryMode ? '退出重练' : '退出巩固' }}
        </button>
      </div>
    </div>

    <!-- ========== 错题重练模式 ========== -->
    <div v-if="retryMode" class="space-y-6">
      <!-- 重练进度 -->
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 sm:p-6">
        <div class="flex items-center justify-between mb-3">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            重练进度：{{ retryIndex + 1 }} / {{ retryQuizzes.length }}
          </span>
          <span class="text-sm text-gray-500 dark:text-gray-400">
            正确 {{ retryCorrect }} 题 · 错误 {{ retryWrong }} 题
          </span>
        </div>
        <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
          <div
            class="bg-primary-500 h-2 rounded-full transition-all duration-300"
            :style="{ width: `${((retryIndex + (retryAnswered ? 1 : 0)) / retryQuizzes.length) * 100}%` }"
          ></div>
        </div>
      </div>

      <!-- 当前题目 -->
      <div
        v-if="retryQuizzes.length > 0 && retryIndex < retryQuizzes.length"
        class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 sm:p-6"
      >
        <div class="flex items-center gap-2 mb-4">
          <span
            class="px-2 py-0.5 rounded text-xs font-medium"
            :class="
              retryQuiz.difficulty === 'easy'
                ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                : retryQuiz.difficulty === 'hard'
                  ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                  : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
            "
          >
            {{ retryQuiz.difficulty === 'easy' ? '简单' : retryQuiz.difficulty === 'hard' ? '困难' : '中等' }}
          </span>
          <span
            class="px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
          >
            {{ typeLabel(retryQuiz.type) }}
          </span>
          <span v-if="currentMistakeItem" class="text-xs text-gray-400 dark:text-gray-500 ml-auto">
            上次错误答案：
            <span class="text-red-500 dark:text-red-400 font-medium">{{ currentMistakeItem.last_wrong_answer }}</span>
          </span>
        </div>

        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-6">{{ retryQuiz.question }}</h3>

        <!-- 选择题 -->
        <div v-if="retryQuiz.type === 'choice' && retryOptions.length" class="space-y-3 mb-6">
          <button
            v-for="(opt, idx) in retryOptions"
            :key="idx"
            class="w-full text-left px-4 py-3 rounded-lg border-2 transition text-sm"
            :class="retryAnswerClass(idx)"
            :disabled="retryAnswered"
            @click="!retryAnswered && selectRetryAnswer(String.fromCharCode(65 + idx))"
          >
            <span class="font-medium mr-2">{{ String.fromCharCode(65 + idx) }}.</span>
            {{ opt }}
            <svg
              v-if="retryAnswered && String.fromCharCode(65 + idx) === retryQuiz.answer"
              class="inline w-4 h-4 ml-2 text-green-500"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fill-rule="evenodd"
                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                clip-rule="evenodd"
              />
            </svg>
          </button>
        </div>

        <!-- 填空/简答 -->
        <div v-else class="mb-6">
          <textarea
            v-model="retryTextAnswer"
            :disabled="retryAnswered"
            rows="3"
            class="w-full rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 p-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary-300 disabled:opacity-60"
            placeholder="输入你的答案..."
          ></textarea>
          <button
            v-if="!retryAnswered"
            class="mt-3 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white text-sm rounded-lg transition"
            @click="submitRetryTextAnswer"
          >
            提交答案
          </button>
        </div>

        <!-- 答题结果 -->
        <div v-if="retryAnswered" class="space-y-4">
          <div
            class="p-4 rounded-lg"
            :class="
              retryIsCorrect
                ? 'bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800'
                : 'bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800'
            "
          >
            <div class="flex items-center gap-2 mb-2">
              <svg v-if="retryIsCorrect" class="w-5 h-5 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                  clip-rule="evenodd"
                />
              </svg>
              <svg v-else class="w-5 h-5 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                  clip-rule="evenodd"
                />
              </svg>
              <span
                class="font-medium"
                :class="retryIsCorrect ? 'text-green-700 dark:text-green-400' : 'text-red-700 dark:text-red-400'"
              >
                {{ retryIsCorrect ? '回答正确！' : '回答错误' }}
              </span>
            </div>
            <div v-if="!retryIsCorrect" class="text-sm space-y-1">
              <p>
                <span class="text-gray-500 dark:text-gray-400">你的答案：</span>
                <span class="text-red-600 dark:text-red-400">{{ retrySelectedAnswer || retryTextAnswer }}</span>
              </p>
              <p>
                <span class="text-gray-500 dark:text-gray-400">正确答案：</span>
                <span class="text-green-600 dark:text-green-400">{{ retryQuiz.answer }}</span>
              </p>
            </div>
            <p v-if="retryQuiz.explanation" class="text-sm text-gray-600 dark:text-gray-400 mt-2">
              {{ retryQuiz.explanation }}
            </p>
          </div>

          <!-- 下一题按钮 -->
          <div class="flex justify-end">
            <button
              class="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white text-sm rounded-lg transition"
              @click="nextRetryQuestion"
            >
              {{ retryIndex + 1 >= retryQuizzes.length ? '完成重练' : '下一题 →' }}
            </button>
          </div>
        </div>
      </div>

      <!-- 重练完成 -->
      <div
        v-if="retryIndex >= retryQuizzes.length"
        class="bg-white dark:bg-gray-800 rounded-xl border border-green-200 dark:border-green-800 p-6 text-center"
      >
        <div
          class="w-16 h-16 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center mx-auto mb-4"
        >
          <svg class="w-8 h-8 text-green-500" fill="currentColor" viewBox="0 0 20 20">
            <path
              fill-rule="evenodd"
              d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
              clip-rule="evenodd"
            />
          </svg>
        </div>
        <h3 class="text-lg font-bold text-gray-900 dark:text-gray-100 mb-2">重练完成！</h3>
        <p class="text-gray-500 dark:text-gray-400 mb-4">
          共 {{ retryQuizzes.length }} 题，正确 {{ retryCorrect }} 题，错误 {{ retryWrong }} 题
          <span v-if="retryQuizzes.length > 0" class="font-medium">
            ({{ Math.round((retryCorrect / retryQuizzes.length) * 100) }}%)
          </span>
        </p>
        <div class="flex justify-center gap-3">
          <button
            class="px-4 py-2 bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 text-sm rounded-lg transition"
            @click="exitRetry"
          >
            返回错题本
          </button>
        </div>
      </div>
    </div>

    <!-- ========== 巩固强化模式 ========== -->
    <div v-else-if="consolidateMode" class="space-y-6">
      <!-- 薄弱点分析 -->
      <div
        v-if="weakAreas.length > 0"
        class="bg-white dark:bg-gray-800 rounded-xl border border-violet-200 dark:border-violet-800 p-4 sm:p-6"
      >
        <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3 flex items-center gap-2">
          <svg class="w-4 h-4 text-violet-500" fill="currentColor" viewBox="0 0 20 20">
            <path
              fill-rule="evenodd"
              d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
              clip-rule="evenodd"
            />
          </svg>
          薄弱点分析
        </h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          <div
            v-for="area in weakAreas"
            :key="area.material_id"
            class="px-3 py-2 rounded-lg bg-violet-50 dark:bg-violet-900/20 border border-violet-100 dark:border-violet-800"
          >
            <p class="text-xs font-medium text-violet-700 dark:text-violet-300 truncate">
              {{ area.material_title || '未知材料' }}
            </p>
            <div class="flex items-center gap-2 mt-1 text-xs text-gray-600 dark:text-gray-400">
              <span class="text-red-600 dark:text-red-400 font-medium">错 {{ area.mistake_count }} 题</span>
              <span
                v-if="area.main_type"
                class="px-1.5 py-0.5 rounded bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
              >
                {{ typeLabel(area.main_type) }}
              </span>
              <span
                v-if="area.main_difficulty"
                class="px-1.5 py-0.5 rounded"
                :class="
                  area.main_difficulty === 'easy'
                    ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                    : area.main_difficulty === 'hard'
                      ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                      : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                "
              >
                {{ area.main_difficulty === 'easy' ? '简单' : area.main_difficulty === 'hard' ? '困难' : '中等' }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 巩固进度 -->
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 sm:p-6">
        <div class="flex items-center justify-between mb-3">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            巩固进度：{{ consolidateIndex + 1 }} / {{ consolidateQuizzes.length }}
          </span>
          <span class="text-sm text-gray-500 dark:text-gray-400">
            正确 {{ consolidateCorrectCount }} 题 · 错误 {{ consolidateWrongCount }} 题
          </span>
        </div>
        <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
          <div
            class="bg-violet-500 h-2 rounded-full transition-all duration-300"
            :style="{
              width: `${((consolidateIndex + (consolidateAnswered ? 1 : 0)) / consolidateQuizzes.length) * 100}%`
            }"
          ></div>
        </div>
      </div>

      <!-- 当前题目 -->
      <div
        v-if="consolidateQuizzes.length > 0 && consolidateIndex < consolidateQuizzes.length"
        class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 sm:p-6"
      >
        <div class="flex items-center gap-2 mb-4 flex-wrap">
          <span
            class="px-2 py-0.5 rounded text-xs font-medium"
            :class="
              cQuiz.difficulty === 'easy'
                ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                : cQuiz.difficulty === 'hard'
                  ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                  : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
            "
          >
            {{ cQuiz.difficulty === 'easy' ? '简单' : cQuiz.difficulty === 'hard' ? '困难' : '中等' }}
          </span>
          <span
            class="px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
          >
            {{ typeLabel(cQuiz.type) }}
          </span>
          <span v-if="cQuiz.material_title" class="text-xs text-gray-400 dark:text-gray-500 truncate max-w-[200px]">
            {{ cQuiz.material_title }}
          </span>
        </div>

        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-6">{{ cQuiz.question }}</h3>

        <!-- 选择题 -->
        <div v-if="cQuiz.type === 'choice' && cOptions.length" class="space-y-3 mb-6">
          <button
            v-for="(opt, idx) in cOptions"
            :key="idx"
            class="w-full text-left px-4 py-3 rounded-lg border-2 transition text-sm"
            :class="cAnswerClass(idx)"
            :disabled="consolidateAnswered"
            @click="!consolidateAnswered && selectConsolidateAnswer(String.fromCharCode(65 + idx))"
          >
            <span class="font-medium mr-2">{{ String.fromCharCode(65 + idx) }}.</span>
            {{ opt }}
            <svg
              v-if="consolidateAnswered && String.fromCharCode(65 + idx) === cQuiz.answer"
              class="inline w-4 h-4 ml-2 text-green-500"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fill-rule="evenodd"
                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                clip-rule="evenodd"
              />
            </svg>
          </button>
        </div>

        <!-- 填空/简答 -->
        <div v-else class="mb-6">
          <textarea
            v-model="consolidateTextAnswer"
            :disabled="consolidateAnswered"
            rows="3"
            class="w-full rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 p-3 text-sm focus:outline-none focus:ring-2 focus:ring-violet-300 disabled:opacity-60"
            placeholder="输入你的答案..."
          ></textarea>
          <button
            v-if="!consolidateAnswered"
            class="mt-3 px-4 py-2 bg-violet-600 hover:bg-violet-700 text-white text-sm rounded-lg transition"
            @click="submitConsolidateTextAnswer"
          >
            提交答案
          </button>
        </div>

        <!-- 答题结果 -->
        <div v-if="consolidateAnswered" class="space-y-4">
          <div
            class="p-4 rounded-lg"
            :class="
              consolidateCorrect
                ? 'bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800'
                : 'bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800'
            "
          >
            <div class="flex items-center gap-2 mb-2">
              <svg v-if="consolidateCorrect" class="w-5 h-5 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                  clip-rule="evenodd"
                />
              </svg>
              <svg v-else class="w-5 h-5 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                  clip-rule="evenodd"
                />
              </svg>
              <span
                class="font-medium"
                :class="consolidateCorrect ? 'text-green-700 dark:text-green-400' : 'text-red-700 dark:text-red-400'"
              >
                {{ consolidateCorrect ? '回答正确！' : '回答错误' }}
              </span>
            </div>
            <div v-if="!consolidateCorrect" class="text-sm space-y-1">
              <p>
                <span class="text-gray-500 dark:text-gray-400">你的答案：</span>
                <span class="text-red-600 dark:text-red-400">{{ consolidateAnswer || consolidateTextAnswer }}</span>
              </p>
              <p>
                <span class="text-gray-500 dark:text-gray-400">正确答案：</span>
                <span class="text-green-600 dark:text-green-400">{{ cQuiz.answer }}</span>
              </p>
            </div>
            <p v-if="cQuiz.explanation" class="text-sm text-gray-600 dark:text-gray-400 mt-2">
              {{ cQuiz.explanation }}
            </p>
          </div>

          <!-- 下一题按钮 -->
          <div class="flex justify-end">
            <button
              class="px-4 py-2 bg-violet-600 hover:bg-violet-700 text-white text-sm rounded-lg transition"
              @click="nextConsolidateQuestion"
            >
              {{ consolidateIndex + 1 >= consolidateQuizzes.length ? '完成巩固' : '下一题 →' }}
            </button>
          </div>
        </div>
      </div>

      <!-- 巩固完成 -->
      <div
        v-if="consolidateIndex >= consolidateQuizzes.length"
        class="bg-white dark:bg-gray-800 rounded-xl border border-violet-200 dark:border-violet-800 p-6 text-center"
      >
        <div
          class="w-16 h-16 bg-violet-100 dark:bg-violet-900/30 rounded-full flex items-center justify-center mx-auto mb-4"
        >
          <svg class="w-8 h-8 text-violet-500" fill="currentColor" viewBox="0 0 20 20">
            <path
              fill-rule="evenodd"
              d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
              clip-rule="evenodd"
            />
          </svg>
        </div>
        <h3 class="text-lg font-bold text-gray-900 dark:text-gray-100 mb-2">巩固练习完成！</h3>
        <p class="text-gray-500 dark:text-gray-400 mb-4">
          共 {{ consolidateQuizzes.length }} 题，正确 {{ consolidateCorrectCount }} 题，错误
          {{ consolidateWrongCount }} 题
          <span v-if="consolidateQuizzes.length > 0" class="font-medium">
            ({{ Math.round((consolidateCorrectCount / consolidateQuizzes.length) * 100) }}%)
          </span>
        </p>
        <div class="flex justify-center gap-3">
          <button
            class="px-4 py-2 bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 text-sm rounded-lg transition"
            @click="exitConsolidate"
          >
            返回错题本
          </button>
        </div>
      </div>
    </div>

    <!-- ========== 错题列表模式 ========== -->
    <div v-else>
      <!-- 统计卡片 -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4 mb-6">
        <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">总错题</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-gray-100">{{ stats?.total || 0 }}</p>
        </div>
        <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">待复习</p>
          <p class="text-2xl font-bold text-red-600 dark:text-red-400">{{ stats?.unreviewed || 0 }}</p>
        </div>
        <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">已复习</p>
          <p class="text-2xl font-bold text-green-600 dark:text-green-400">{{ stats?.reviewed || 0 }}</p>
        </div>
        <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">复习率</p>
          <p class="text-2xl font-bold text-primary-600 dark:text-primary-400">
            {{ (stats?.total || 0) > 0 ? Math.round(((stats?.reviewed || 0) / (stats?.total || 1)) * 100) : 0 }}%
          </p>
        </div>
      </div>

      <!-- 过滤标签 + 批量操作 -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-4 gap-3">
        <div class="flex items-center gap-2">
          <button
            v-for="tab in filterTabs"
            :key="tab.value"
            class="px-3 py-1.5 rounded-lg text-sm font-medium transition"
            :class="
              activeFilter === tab.value
                ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
                : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
            "
            @click="activeFilter = tab.value"
          >
            {{ tab.label }}
            <span v-if="tab.count !== undefined" class="ml-1 text-xs">({{ tab.count }})</span>
          </button>
        </div>
        <div class="flex items-center gap-2">
          <!-- 全选 -->
          <label class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 cursor-pointer">
            <input
              type="checkbox"
              :checked="isAllSelected"
              class="rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500"
              @change="toggleSelectAll"
            />
            全选
            <span v-if="selectedIds.size > 0" class="text-primary-600 dark:text-primary-400">
              已选 {{ selectedIds.size }}
            </span>
          </label>
          <!-- 批量标记复习 -->
          <button
            v-if="selectedIds.size > 0"
            :disabled="batchLoading"
            class="px-3 py-1.5 bg-green-600 hover:bg-green-700 text-white text-xs rounded-lg transition disabled:opacity-50"
            @click="handleBatchReview"
          >
            {{ batchLoading ? '处理中...' : '批量标记复习' }}
          </button>
        </div>
      </div>

      <!-- 骨架屏加载 -->
      <div v-if="loading" class="space-y-4">
        <ListSkeleton :count="6" type="mistake" />
      </div>

      <div
        v-else-if="filteredMistakes.length === 0"
        class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-12 text-center"
      >
        <div
          class="w-16 h-16 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center mx-auto mb-4"
        >
          <svg class="w-8 h-8 text-green-500" fill="currentColor" viewBox="0 0 20 20">
            <path
              fill-rule="evenodd"
              d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
              clip-rule="evenodd"
            />
          </svg>
        </div>
        <h3 class="text-lg font-bold text-gray-900 dark:text-gray-100 mb-2">
          {{
            activeFilter === 'unreviewed'
              ? '没有待复习的错题'
              : activeFilter === 'reviewed'
                ? '还没有复习过任何错题'
                : '错题本是空的'
          }}
        </h3>
        <p class="text-gray-500 dark:text-gray-400 text-sm">
          {{ activeFilter === 'unreviewed' ? '继续保持，去做些练习吧！' : '去练习场答题，答错的题目会自动收集到这里' }}
        </p>
        <router-link
          to="/quiz"
          class="inline-flex items-center gap-2 mt-4 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white text-sm rounded-lg transition"
        >
          去练习场
        </router-link>
      </div>

      <div v-else ref="mistakeListRef" class="space-y-4 max-h-[70vh] overflow-y-auto custom-scroll">
        <!-- 虚拟滚动：顶部占位 -->
        <div v-if="mistVirtualized" :style="{ height: mistTopSpacer + 'px' }" aria-hidden="true"></div>
        <div
          v-for="item in visibleMistakes"
          :key="item.id"
          class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden transition hover:border-gray-300 dark:hover:border-gray-600"
        >
          <!-- 题目头部 -->
          <div class="p-4 sm:p-5">
            <div class="flex items-start gap-3">
              <!-- 选择框 -->
              <input
                type="checkbox"
                :checked="selectedIds.has(item.id)"
                class="mt-1 rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500 shrink-0"
                @change="toggleSelect(item.id)"
              />
              <div class="flex-1 min-w-0">
                <!-- 标签行 -->
                <div class="flex items-center gap-2 flex-wrap mb-2">
                  <span
                    class="px-2 py-0.5 rounded text-xs font-medium"
                    :class="
                      item.difficulty === 'easy'
                        ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                        : item.difficulty === 'hard'
                          ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                          : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                    "
                  >
                    {{ item.difficulty === 'easy' ? '简单' : item.difficulty === 'hard' ? '困难' : '中等' }}
                  </span>
                  <span
                    class="px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
                  >
                    {{ typeLabel(item.quiz_type) }}
                  </span>
                  <span
                    v-if="item.material_title"
                    class="text-xs text-gray-400 dark:text-gray-500 truncate max-w-[200px]"
                  >
                    {{ item.material_title }}
                  </span>
                  <span
                    v-if="item.reviewed"
                    class="px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"
                  >
                    已复习
                  </span>
                  <span class="text-xs text-gray-400 dark:text-gray-500 ml-auto shrink-0">
                    {{ formatTime(item.mistake_at) }}
                  </span>
                </div>

                <!-- 题目 -->
                <h3 class="text-sm sm:text-base font-medium text-gray-900 dark:text-gray-100 mb-3">
                  {{ item.question }}
                </h3>

                <!-- 选择题选项展示 -->
                <div v-if="item.quiz_type === 'choice' && item.options" class="mb-3">
                  <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                    <div
                      v-for="(opt, idx) in parseOptions(item.options)"
                      :key="idx"
                      class="px-3 py-2 rounded-lg text-xs sm:text-sm border transition"
                      :class="optionClass(item, String.fromCharCode(65 + idx), opt)"
                    >
                      <span class="font-medium mr-1">{{ String.fromCharCode(65 + idx) }}.</span>
                      {{ opt }}
                    </div>
                  </div>
                </div>

                <!-- 答案对比（非选择题） -->
                <div v-if="item.quiz_type !== 'choice'" class="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-3">
                  <div
                    class="px-3 py-2 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800"
                  >
                    <p class="text-xs text-red-500 dark:text-red-400 mb-1">你的答案</p>
                    <p class="text-sm text-red-700 dark:text-red-300 font-medium">{{ item.user_answer || '未作答' }}</p>
                  </div>
                  <div
                    class="px-3 py-2 rounded-lg bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800"
                  >
                    <p class="text-xs text-green-500 dark:text-green-400 mb-1">正确答案</p>
                    <p class="text-sm text-green-700 dark:text-green-300 font-medium">{{ item.correct_answer }}</p>
                  </div>
                </div>

                <!-- 解析 -->
                <div
                  v-if="item.explanation"
                  class="text-sm text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-700/50 rounded-lg p-3"
                >
                  <span class="font-medium text-gray-700 dark:text-gray-300">解析：</span>
                  {{ item.explanation }}
                </div>
              </div>
            </div>
          </div>

          <!-- 操作栏 -->
          <div
            class="px-4 sm:px-5 py-3 bg-gray-50 dark:bg-gray-800/50 border-t border-gray-100 dark:border-gray-700 flex items-center justify-end gap-2"
          >
            <button
              v-if="!item.reviewed"
              class="px-3 py-1.5 bg-green-600 hover:bg-green-700 text-white text-xs rounded-lg transition"
              @click="handleReview(item)"
            >
              标记已复习
            </button>
            <button
              class="px-3 py-1.5 text-gray-500 hover:text-red-600 dark:text-gray-400 dark:hover:text-red-400 text-xs transition"
              @click="handleDelete(item)"
            >
              移除
            </button>
          </div>
        </div>
        <!-- 虚拟滚动：底部占位 -->
        <div v-if="mistVirtualized" :style="{ height: mistBottomSpacer + 'px' }" aria-hidden="true"></div>
        <div v-if="mistVirtualized" class="text-center py-2 text-xs text-gray-400 dark:text-gray-500">
          已启用虚拟滚动 · 显示 {{ mistEndIdx - mistStartIdx }} / {{ filteredMistakes.length }} 条
        </div>
        <!-- 无限滚动：哨兵元素 -->
        <div ref="mistSentinelRef" class="h-1" aria-hidden="true"></div>
        <!-- 无限滚动：底部状态 -->
        <InfiniteScrollFooter
          :loading="mistLoadMore"
          :has-more="mistHasMore"
          :error="mistLoadError"
          :total-count="mistakes.length"
          @retry="loadMoreMistakes"
        />
      </div>
      <ScrollToTop :show="showMistScrollTop" @click="scrollToMistTop" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, reactive, nextTick } from 'vue'
import {
  listMistakes,
  getMistakeStats,
  reviewMistake,
  batchReviewMistakes,
  deleteMistake,
  retryMistakes,
  consolidatePractice
} from '../api/client'
import { useToast, useConfirm } from '../composables/useToast'
import { useVirtualScroll } from '../composables/useVirtualScroll'
import { useInfiniteScroll, useScrollToTop } from '../composables/useInfiniteScroll'
import ListSkeleton from '../components/ListSkeleton.vue'
import InfiniteScrollFooter from '../components/InfiniteScrollFooter.vue'
import ScrollToTop from '../components/ScrollToTop.vue'

const toast = useToast()
const { confirm } = useConfirm()

// ========== 状态 ==========
const loading = ref(true)
const mistakes = ref([])
const stats = reactive({ total: 0, reviewed: 0, unreviewed: 0 })
const activeFilter = ref('all')
const selectedIds = ref(new Set())

// 无限滚动
const mistSentinelRef = ref(null)
const mistakeListRef = ref(null)
const {
  loading: mistLoadMore,
  hasMore: mistHasMore,
  error: mistLoadError,
  total,
  loadNext: mistLoadNext,
  reset: mistResetScroll,
  retry: _mistRetry,
  init: mistInitScroll
} = useInfiniteScroll({
  limit: 20,
  rootMargin: '300px',
  onLoad: async (offset, limit) => {
    const res = await listMistakes({ limit, offset })
    return { items: res.data.data || [], total: res.data.total || 0 }
  },
  onItems: (items) => {
    mistakes.value.push(...items)
  }
})

// 返回顶部
const {
  showButton: showMistScrollTop,
  scrollToTop: scrollToMistTop,
  initScrollListener: initMistScrollListener,
  destroyScrollListener: destroyMistScrollListener
} = useScrollToTop(mistakeListRef)

// 重练状态
const retryMode = ref(false)
const retryLoading = ref(false)
const retryQuizzes = ref([])
const retryIndex = ref(0)
const retrySelectedAnswer = ref('')
const retryTextAnswer = ref('')
const retryAnswered = ref(false)
const retryIsCorrect = ref(false)
const retryCorrect = ref(0)
const retryWrong = ref(0)

// 巩固强化状态
const consolidateMode = ref(false)
const consolidateLoading = ref(false)
const consolidateQuizzes = ref([])
const consolidateIndex = ref(0)
const consolidateAnswer = ref('')
const consolidateTextAnswer = ref('')
const consolidateAnswered = ref(false)
const consolidateCorrect = ref(false)
const consolidateCorrectCount = ref(0)
const consolidateWrongCount = ref(0)
const weakAreas = ref([])
const _consolidateTitle = ref('巩固强化练习')

// ========== 计算属性 ==========
const filterTabs = computed(() => [
  { label: '全部', value: 'all', count: stats?.total || 0 },
  { label: '待复习', value: 'unreviewed', count: stats?.unreviewed || 0 },
  { label: '已复习', value: 'reviewed', count: stats?.reviewed || 0 }
])

const filteredMistakes = computed(() => {
  if (activeFilter.value === 'unreviewed') return mistakes.value.filter((m) => !m.reviewed)
  if (activeFilter.value === 'reviewed') return mistakes.value.filter((m) => m.reviewed)
  return mistakes.value
})

// 虚拟滚动：错题列表 >100 条时启用
const {
  startIndex: mistStartIdx,
  endIndex: mistEndIdx,
  topSpacerHeight: mistTopSpacer,
  bottomSpacerHeight: mistBottomSpacer,
  shouldVirtualize: mistVirtualized
} = useVirtualScroll(
  mistakeListRef,
  computed(() => filteredMistakes.value.length),
  {
    itemHeight: 200,
    buffer: 5,
    threshold: 100
  }
)

const visibleMistakes = computed(() => {
  if (!mistVirtualized.value) return filteredMistakes.value
  return filteredMistakes.value.slice(mistStartIdx.value, mistEndIdx.value)
})

const isAllSelected = computed(() => {
  return filteredMistakes.value.length > 0 && filteredMistakes.value.every((m) => selectedIds.value.has(m.id))
})

const retryQuiz = computed(() => retryQuizzes.value[retryIndex.value] || {})
const currentMistakeItem = computed(() => retryQuizzes.value[retryIndex.value] || null)

const retryOptions = computed(() => {
  if (!retryQuiz.value.options) return []
  try {
    return JSON.parse(retryQuiz.value.options)
  } catch {
    return []
  }
})

// 巩固强化计算属性
const cQuiz = computed(() => consolidateQuizzes.value[consolidateIndex.value] || {})
const cOptions = computed(() => {
  if (!cQuiz.value.options) return []
  try {
    return JSON.parse(cQuiz.value.options)
  } catch {
    return []
  }
})

// ========== 方法 ==========
async function loadMistakes() {
  loading.value = true
  mistResetScroll()
  mistakes.value = []
  try {
    const newItems = await mistLoadNext()
    mistakes.value.push(...newItems)
  } catch (e) {
    toast.error('加载错题失败')
  } finally {
    loading.value = false
  }
}

async function loadMoreMistakes() {
  try {
    const newItems = await mistLoadNext()
    mistakes.value.push(...newItems)
  } catch (e) {
    toast.error('加载更多错题失败')
  }
}

function initMistInfiniteScroll() {
  if (mistSentinelRef.value && mistakeListRef.value) {
    mistInitScroll(mistSentinelRef.value, mistakeListRef.value)
    initMistScrollListener()
  }
}

async function loadStats() {
  try {
    const res = await getMistakeStats()
    const data = res?.data || {}
    stats.total = data.total || 0
    stats.reviewed = data.reviewed || 0
    stats.unreviewed = data.unreviewed || 0
  } catch (e) {
    console.warn('加载错题统计失败:', e)
  }
}

async function handleReview(item) {
  try {
    await reviewMistake(item.id)
    item.reviewed = true
    stats.reviewed++
    stats.unreviewed = Math.max(0, stats.unreviewed - 1)
    toast.success('已标记为已复习')
  } catch {
    toast.error('标记失败')
  }
}

async function handleDelete(item) {
  const ok = await confirm('确定移除这条错题记录？')
  if (!ok) return
  try {
    await deleteMistake(item.id)
    mistakes.value = mistakes.value.filter((m) => m.id !== item.id)
    total.value = Math.max(0, total.value - 1)
    if (item.reviewed) stats.reviewed = Math.max(0, stats.reviewed - 1)
    else stats.unreviewed = Math.max(0, stats.unreviewed - 1)
    stats.total = Math.max(0, stats.total - 1)
    selectedIds.value.delete(item.id)
    toast.success('已移除')
  } catch {
    toast.error('移除失败')
  }
}

async function handleBatchReview() {
  const ids = [...selectedIds.value]
  if (ids.length === 0) return
  const ok = await confirm(`确定将 ${ids.length} 条错题标记为已复习？`)
  if (!ok) return
  try {
    await batchReviewMistakes(ids)
    mistakes.value.forEach((m) => {
      if (ids.includes(m.id) && !m.reviewed) {
        m.reviewed = true
        stats.reviewed++
        stats.unreviewed = Math.max(0, stats.unreviewed - 1)
      }
    })
    selectedIds.value.clear()
    toast.success(`已标记 ${ids.length} 条为已复习`)
  } catch {
    toast.error('批量标记失败')
  }
}

function toggleSelect(id) {
  const newSet = new Set(selectedIds.value)
  if (newSet.has(id)) newSet.delete(id)
  else newSet.add(id)
  selectedIds.value = newSet
}

function toggleSelectAll() {
  if (isAllSelected.value) {
    selectedIds.value = new Set()
  } else {
    selectedIds.value = new Set(filteredMistakes.value.map((m) => m.id))
  }
}

// ========== 重练模式 ==========
async function startRetry() {
  retryLoading.value = true
  try {
    const res = await retryMistakes()
    if (res.data.quizzes && res.data.quizzes.length > 0) {
      retryQuizzes.value = res.data.quizzes
      retryMode.value = true
      retryIndex.value = 0
      retryCorrect.value = 0
      retryWrong.value = 0
      resetRetryAnswer()
    } else {
      toast.info('没有待复习的错题')
    }
  } catch {
    toast.error('加载重练数据失败')
  } finally {
    retryLoading.value = false
  }
}

function exitRetry() {
  retryMode.value = false
  retryQuizzes.value = []
  retryIndex.value = 0
  loadStats()
  loadMistakes()
}

function selectRetryAnswer(letter) {
  if (retryAnswered.value) return
  retrySelectedAnswer.value = letter
  retryAnswered.value = true
  // 检查是否正确
  const correct = retryQuiz.value.answer
  retryIsCorrect.value = letter === correct || letter === correct?.charAt(0)
  if (retryIsCorrect.value) {
    retryCorrect.value++
    // 答对了，标记该错题为已复习
    autoReviewMistake()
  } else {
    retryWrong.value++
  }
}

function submitRetryTextAnswer() {
  if (!retryTextAnswer.value.trim() || retryAnswered.value) return
  retrySelectedAnswer.value = retryTextAnswer.value
  retryAnswered.value = true
  const correct = retryQuiz.value.answer?.trim().toLowerCase()
  const user = retryTextAnswer.value.trim().toLowerCase()
  retryIsCorrect.value = user === correct
  if (retryIsCorrect.value) {
    retryCorrect.value++
    autoReviewMistake()
  } else {
    retryWrong.value++
  }
}

async function autoReviewMistake() {
  const item = currentMistakeItem.value
  if (item?.mistake_id) {
    try {
      await reviewMistake(item.mistake_id)
    } catch {}
  }
}

function nextRetryQuestion() {
  if (retryIndex.value + 1 >= retryQuizzes.value.length) {
    // 重练完成
    retryIndex.value = retryQuizzes.value.length
    toast.success(
      `重练完成！正确率 ${retryQuizzes.value.length > 0 ? Math.round((retryCorrect.value / retryQuizzes.value.length) * 100) : 0}%`
    )
    return
  }
  retryIndex.value++
  resetRetryAnswer()
}

function resetRetryAnswer() {
  retrySelectedAnswer.value = ''
  retryTextAnswer.value = ''
  retryAnswered.value = false
  retryIsCorrect.value = false
}

function retryAnswerClass(idx) {
  const letter = String.fromCharCode(65 + idx)
  const correct = retryQuiz.value.answer
  if (!retryAnswered.value) {
    return 'border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:border-primary-400 dark:hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/20 cursor-pointer'
  }
  if (letter === correct || letter === correct?.charAt(0)) {
    return 'border-green-400 bg-green-50 dark:border-green-600 dark:bg-green-900/30 text-green-700 dark:text-green-400'
  }
  if (letter === retrySelectedAnswer.value && !retryIsCorrect.value) {
    return 'border-red-400 bg-red-50 dark:border-red-600 dark:bg-red-900/30 text-red-700 dark:text-red-400'
  }
  return 'border-gray-200 dark:border-gray-600 text-gray-400 dark:text-gray-500 opacity-50'
}

// ========== 巩固强化模式 ==========
async function startConsolidate() {
  consolidateLoading.value = true
  try {
    const res = await consolidatePractice()
    if (res.data.quizzes && res.data.quizzes.length > 0) {
      consolidateQuizzes.value = res.data.quizzes
      weakAreas.value = res.data.weak_areas || []
      consolidateMode.value = true
      consolidateIndex.value = 0
      consolidateCorrectCount.value = 0
      consolidateWrongCount.value = 0
      resetConsolidateAnswer()
    } else {
      toast.info(res.data.message || '暂无同类题目可供练习')
    }
  } catch (e) {
    const msg = e.response?.data?.error || e.response?.data?.message || e.message || '请求失败'
    const status = e.response?.status
    if (status === 404) {
      toast.error('巩固强化接口不可用，请重启后端服务')
    } else if (status === 500) {
      toast.error(`服务端错误：${msg}`)
    } else {
      toast.error(`加载巩固练习失败：${msg}`)
    }
    console.error('[Consolidate] API error:', e.response?.status, e.response?.data, e.message)
  } finally {
    consolidateLoading.value = false
  }
}

function exitConsolidate() {
  consolidateMode.value = false
  consolidateQuizzes.value = []
  consolidateIndex.value = 0
  weakAreas.value = []
  loadStats()
  loadMistakes()
}

function selectConsolidateAnswer(letter) {
  if (consolidateAnswered.value) return
  consolidateAnswer.value = letter
  consolidateAnswered.value = true
  const correct = cQuiz.value.answer
  consolidateCorrect.value = letter === correct || letter === correct?.charAt(0)
  if (consolidateCorrect.value) consolidateCorrectCount.value++
  else consolidateWrongCount.value++
}

function submitConsolidateTextAnswer() {
  if (!consolidateTextAnswer.value.trim() || consolidateAnswered.value) return
  consolidateAnswer.value = consolidateTextAnswer.value
  consolidateAnswered.value = true
  const correct = cQuiz.value.answer?.trim().toLowerCase()
  const user = consolidateTextAnswer.value.trim().toLowerCase()
  consolidateCorrect.value = user === correct
  if (consolidateCorrect.value) consolidateCorrectCount.value++
  else consolidateWrongCount.value++
}

function nextConsolidateQuestion() {
  if (consolidateIndex.value + 1 >= consolidateQuizzes.value.length) {
    consolidateIndex.value = consolidateQuizzes.value.length
    const pct =
      consolidateQuizzes.value.length > 0
        ? Math.round((consolidateCorrectCount.value / consolidateQuizzes.value.length) * 100)
        : 0
    toast.success(`巩固练习完成！正确率 ${pct}%`)
    return
  }
  consolidateIndex.value++
  resetConsolidateAnswer()
}

function resetConsolidateAnswer() {
  consolidateAnswer.value = ''
  consolidateTextAnswer.value = ''
  consolidateAnswered.value = false
  consolidateCorrect.value = false
}

function cAnswerClass(idx) {
  const letter = String.fromCharCode(65 + idx)
  const correct = cQuiz.value.answer
  if (!consolidateAnswered.value) {
    return 'border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:border-primary-400 dark:hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/20 cursor-pointer'
  }
  if (letter === correct || letter === correct?.charAt(0)) {
    return 'border-green-400 bg-green-50 dark:border-green-600 dark:bg-green-900/30 text-green-700 dark:text-green-400'
  }
  if (letter === consolidateAnswer.value && !consolidateCorrect.value) {
    return 'border-red-400 bg-red-50 dark:border-red-600 dark:bg-red-900/30 text-red-700 dark:text-red-400'
  }
  return 'border-gray-200 dark:border-gray-600 text-gray-400 dark:text-gray-500 opacity-50'
}

// ========== 辅助函数 ==========
function typeLabel(type) {
  const map = { choice: '选择题', fill: '填空题', short_answer: '简答题', judge: '判断题' }
  return map[type] || type
}

function parseOptions(optStr) {
  try {
    return JSON.parse(optStr)
  } catch {
    return []
  }
}

function optionClass(item, letter, optText) {
  const correct = item.correct_answer
  const userAns = item.user_answer
  const isCorrectLetter = letter === correct || letter === correct?.charAt(0) || optText === correct
  const isUserLetter = letter === userAns || letter === userAns?.charAt(0) || optText === userAns
  if (isCorrectLetter) {
    return 'border-green-400 bg-green-50 dark:border-green-600 dark:bg-green-900/30 text-green-700 dark:text-green-400 font-medium'
  }
  if (isUserLetter && !isCorrectLetter) {
    return 'border-red-400 bg-red-50 dark:border-red-600 dark:bg-red-900/30 text-red-700 dark:text-red-400 line-through'
  }
  return 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400'
}

function formatTime(t) {
  if (!t) return ''
  const d = new Date(t)
  const now = new Date()
  const diff = now - d
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  if (diff < 604800000) return `${Math.floor(diff / 86400000)} 天前`
  return d.toLocaleDateString('zh-CN')
}

// ========== 初始化 ==========
onMounted(async () => {
  try {
    await Promise.all([loadStats(), loadMistakes()])
  } catch (e) {
    console.warn('错题本初始化加载失败:', e)
  } finally {
    loading.value = false
  }
  await nextTick()
  initMistInfiniteScroll()
})
onBeforeUnmount(() => {
  destroyMistScrollListener()
})
</script>

<style scoped>
.custom-scroll::-webkit-scrollbar {
  width: 6px;
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
