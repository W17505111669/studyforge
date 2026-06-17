<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-6xl mx-auto">
    <!-- ===== 阶段 1：考试配置 ===== -->
    <div v-if="phase === 'config'" class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">模拟考试</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">选择材料，AI 自动生成全真模拟试卷</p>
        </div>
        <button
          v-if="examHistory.length > 0"
          class="px-4 py-2 text-sm font-medium text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-900/30 rounded-lg hover:bg-indigo-100 dark:hover:bg-indigo-900/50 transition-colors"
          @click="showHistory = !showHistory"
        >
          {{ showHistory ? '隐藏历史' : '考试历史' }}
        </button>
      </div>

      <!-- 历史考试列表 -->
      <div
        v-if="showHistory && examHistory.length > 0"
        class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden"
      >
        <div class="px-5 py-3 border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
          <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">历史考试</h3>
        </div>
        <div class="divide-y divide-gray-100 dark:divide-gray-700 max-h-64 overflow-y-auto custom-scroll">
          <div
            v-for="exam in examHistory"
            :key="exam.id"
            class="px-5 py-3 flex items-center justify-between hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer transition-colors"
            @click="viewExamReport(exam.id)"
          >
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-gray-900 dark:text-white truncate">
                {{ exam.material_names?.join('、') || '未知材料' }}
              </p>
              <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                {{ formatDate(exam.started_at) }} · {{ exam.time_limit }}分钟
              </p>
            </div>
            <div class="flex items-center gap-3 ml-4">
              <span
                v-if="exam.status === 'completed'"
                class="text-sm font-bold"
                :class="
                  exam.score >= 80
                    ? 'text-green-600 dark:text-green-400'
                    : exam.score >= 60
                      ? 'text-yellow-600 dark:text-yellow-400'
                      : 'text-red-600 dark:text-red-400'
                "
              >
                {{ exam.score }}分
              </span>
              <span v-else class="text-xs text-gray-400 dark:text-gray-500">未完成</span>
              <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </div>
          </div>
        </div>
      </div>

      <!-- 选择材料 -->
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 sm:p-6">
        <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-4">1. 选择学习材料</h3>
        <div v-if="loadingMaterials" class="space-y-3">
          <div v-for="i in 3" :key="i" class="h-16 bg-gray-100 dark:bg-gray-700 rounded-lg animate-pulse"></div>
        </div>
        <div v-else-if="materials.length === 0" class="text-center py-8 text-gray-400 dark:text-gray-500">
          <p>暂无已分析的材料，请先上传学习材料</p>
          <router-link to="/upload" class="text-indigo-500 hover:text-indigo-600 text-sm mt-2 inline-block">
            前往上传
          </router-link>
        </div>
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          <label
            v-for="m in materials"
            :key="m.id"
            class="relative flex items-start gap-3 p-3 rounded-lg border-2 cursor-pointer transition-all duration-200"
            :class="
              selectedMaterialIDs.includes(m.id)
                ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/30 dark:border-indigo-400'
                : 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
            "
          >
            <input v-model="selectedMaterialIDs" type="checkbox" :value="m.id" class="sr-only" />
            <div
              class="flex-shrink-0 w-5 h-5 rounded border-2 flex items-center justify-center mt-0.5"
              :class="
                selectedMaterialIDs.includes(m.id)
                  ? 'border-indigo-500 bg-indigo-500 dark:border-indigo-400 dark:bg-indigo-400'
                  : 'border-gray-300 dark:border-gray-500'
              "
            >
              <svg
                v-if="selectedMaterialIDs.includes(m.id)"
                class="w-3 h-3 text-white"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path
                  fill-rule="evenodd"
                  d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                  clip-rule="evenodd"
                />
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-gray-900 dark:text-white truncate">{{ m.title }}</p>
              <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                {{ m.status === 'completed' ? '已完成' : '部分完成' }}
              </p>
            </div>
          </label>
        </div>
      </div>

      <!-- 考试设置 -->
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 sm:p-6">
        <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-4">2. 考试设置</h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
          <div>
            <label class="block text-sm text-gray-600 dark:text-gray-400 mb-2">题目数量</label>
            <div class="flex gap-2">
              <button
                v-for="n in [10, 20, 30, 50]"
                :key="n"
                class="flex-1 py-2 px-3 text-sm font-medium rounded-lg transition-all duration-200"
                :class="
                  questionCount === n
                    ? 'bg-indigo-600 text-white shadow-md'
                    : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
                "
                @click="questionCount = n"
              >
                {{ n }}题
              </button>
            </div>
          </div>
          <div>
            <label class="block text-sm text-gray-600 dark:text-gray-400 mb-2">时间限制</label>
            <div class="flex gap-2">
              <button
                v-for="t in [15, 30, 60]"
                :key="t"
                class="flex-1 py-2 px-3 text-sm font-medium rounded-lg transition-all duration-200"
                :class="
                  timeLimit === t
                    ? 'bg-indigo-600 text-white shadow-md'
                    : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
                "
                @click="timeLimit = t"
              >
                {{ t }}分钟
              </button>
            </div>
          </div>
        </div>

        <!-- 题型分布预览 -->
        <div class="mt-5 p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-2">题型分布预览</p>
          <div class="flex flex-wrap gap-3 text-xs">
            <span class="px-2 py-1 rounded-full bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300">
              选择 {{ Math.round(questionCount * 0.4) }}题 × 2分
            </span>
            <span class="px-2 py-1 rounded-full bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-300">
              判断 {{ Math.round(questionCount * 0.2) }}题 × 2分
            </span>
            <span class="px-2 py-1 rounded-full bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300">
              填空 {{ Math.round(questionCount * 0.2) }}题 × 3分
            </span>
            <span
              class="px-2 py-1 rounded-full bg-purple-100 dark:bg-purple-900/40 text-purple-700 dark:text-purple-300"
            >
              简答
              {{
                questionCount -
                Math.round(questionCount * 0.4) -
                Math.round(questionCount * 0.2) -
                Math.round(questionCount * 0.2)
              }}题 × 5分
            </span>
          </div>
        </div>
      </div>

      <!-- 开始按钮 -->
      <button
        :disabled="selectedMaterialIDs.length === 0 || generating"
        class="w-full py-4 bg-gradient-to-r from-indigo-600 to-purple-600 text-white font-semibold rounded-xl hover:from-indigo-700 hover:to-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 flex items-center justify-center gap-2"
        @click="startExam"
      >
        <svg v-if="generating" class="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
        <span v-if="generating">AI 正在出题中...</span>
        <span v-else-if="selectedMaterialIDs.length === 0">请选择至少一个材料</span>
        <span v-else>开始考试</span>
      </button>
    </div>

    <!-- ===== 阶段 2：考试进行中 ===== -->
    <div v-else-if="phase === 'exam'" class="space-y-4">
      <!-- 顶部栏：倒计时 + 进度 -->
      <div
        class="sticky top-0 z-20 bg-white/95 dark:bg-gray-800/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-700 p-4 flex items-center justify-between"
      >
        <div class="flex items-center gap-4">
          <button
            class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
            @click="confirmExit"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">
              第 {{ currentIndex + 1 }} / {{ questions.length }} 题
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              已答 {{ answeredCount }} 题 · 标记 {{ markedCount }} 题
            </p>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <div class="text-right">
            <p
              class="text-lg font-mono font-bold"
              :class="
                remainingSeconds <= 300
                  ? 'text-red-600 dark:text-red-400 animate-pulse'
                  : 'text-gray-900 dark:text-white'
              "
            >
              {{ formatTime(remainingSeconds) }}
            </p>
            <p class="text-xs text-gray-400 dark:text-gray-500">剩余时间</p>
          </div>
          <button
            class="px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 transition-colors"
            @click="submitExamNow"
          >
            交卷
          </button>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-4 gap-4">
        <!-- 题目导航面板 -->
        <div class="lg:col-span-1 order-2 lg:order-1">
          <div
            class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 sticky top-24"
          >
            <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-3 uppercase tracking-wide">
              题目导航
            </h4>
            <div class="grid grid-cols-5 sm:grid-cols-8 lg:grid-cols-5 gap-2">
              <button
                v-for="(q, i) in questions"
                :key="i"
                class="w-9 h-9 rounded-lg text-xs font-bold flex items-center justify-center transition-all duration-200"
                :class="getNavButtonClass(i)"
                @click="currentIndex = i"
              >
                {{ i + 1 }}
              </button>
            </div>
            <div class="mt-3 flex flex-wrap gap-2 text-[10px] text-gray-500 dark:text-gray-400">
              <span class="flex items-center gap-1">
                <span class="w-2.5 h-2.5 rounded bg-green-500"></span>
                已答
              </span>
              <span class="flex items-center gap-1">
                <span class="w-2.5 h-2.5 rounded bg-gray-200 dark:bg-gray-600"></span>
                未答
              </span>
              <span class="flex items-center gap-1">
                <span class="w-2.5 h-2.5 rounded bg-amber-400"></span>
                标记
              </span>
            </div>
          </div>
        </div>

        <!-- 题目展示 -->
        <div class="lg:col-span-3 order-1 lg:order-2">
          <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 sm:p-6">
            <!-- 题目类型和标记 -->
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-center gap-2">
                <span class="px-2.5 py-1 rounded-full text-xs font-medium" :class="typeClass(currentQuestion.type)">
                  {{ typeLabel(currentQuestion.type) }}
                </span>
                <span class="px-2 py-0.5 rounded text-xs text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-700">
                  {{ currentQuestion.points }} 分
                </span>
                <span class="px-2 py-0.5 rounded text-xs" :class="diffClass(currentQuestion.difficulty)">
                  {{ diffLabel(currentQuestion.difficulty) }}
                </span>
              </div>
              <button
                class="flex items-center gap-1.5 text-sm transition-colors"
                :class="
                  isCurrentMarked
                    ? 'text-amber-500 hover:text-amber-600'
                    : 'text-gray-400 hover:text-gray-500 dark:hover:text-gray-300'
                "
                @click="toggleMark"
              >
                <svg
                  class="w-5 h-5"
                  :fill="isCurrentMarked ? 'currentColor' : 'none'"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
                  />
                </svg>
                {{ isCurrentMarked ? '已标记' : '标记' }}
              </button>
            </div>

            <!-- 题目内容 -->
            <h3
              class="text-lg font-semibold text-gray-900 dark:text-white mb-6 leading-relaxed"
              v-html="renderQuestion(currentQuestion.question)"
            ></h3>

            <!-- 选择题 -->
            <div v-if="currentQuestion.type === 'choice'" class="space-y-3">
              <label
                v-for="(opt, oi) in currentQuestion.options"
                :key="oi"
                class="flex items-start gap-3 p-3 rounded-lg border-2 cursor-pointer transition-all duration-200"
                :class="
                  currentAnswer === getOptionLetter(oi)
                    ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/30 dark:border-indigo-400'
                    : 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
                "
              >
                <input
                  v-model="answers[currentIndex]"
                  type="radio"
                  :name="'q-' + currentIndex"
                  :value="getOptionLetter(oi)"
                  class="sr-only"
                />
                <span
                  class="flex-shrink-0 w-7 h-7 rounded-full border-2 flex items-center justify-center text-xs font-bold"
                  :class="
                    currentAnswer === getOptionLetter(oi)
                      ? 'border-indigo-500 bg-indigo-500 text-white dark:border-indigo-400 dark:bg-indigo-400'
                      : 'border-gray-300 dark:border-gray-500 text-gray-500 dark:text-gray-400'
                  "
                >
                  {{ getOptionLetter(oi) }}
                </span>
                <span class="text-sm text-gray-700 dark:text-gray-300 pt-0.5">
                  {{ opt.replace(/^[A-D][.、]\s*/, '') }}
                </span>
              </label>
            </div>

            <!-- 判断题 -->
            <div v-else-if="currentQuestion.type === 'true_false'" class="flex gap-4">
              <button
                v-for="opt in ['正确', '错误']"
                :key="opt"
                class="flex-1 py-4 rounded-lg border-2 text-sm font-medium transition-all duration-200"
                :class="
                  currentAnswer === opt
                    ? opt === '正确'
                      ? 'border-green-500 bg-green-50 text-green-700 dark:bg-green-900/30 dark:border-green-400 dark:text-green-300'
                      : 'border-red-500 bg-red-50 text-red-700 dark:bg-red-900/30 dark:border-red-400 dark:text-red-300'
                    : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:border-gray-300 dark:hover:border-gray-500'
                "
                @click="answers[currentIndex] = opt"
              >
                {{ opt }}
              </button>
            </div>

            <!-- 填空题 -->
            <div v-else-if="currentQuestion.type === 'fill'">
              <input
                v-model="answers[currentIndex]"
                type="text"
                placeholder="请输入答案..."
                class="w-full px-4 py-3 border-2 border-gray-200 dark:border-gray-600 rounded-lg text-sm focus:border-indigo-500 dark:focus:border-indigo-400 focus:outline-none bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 transition-colors"
              />
            </div>

            <!-- 简答题 -->
            <div v-else-if="currentQuestion.type === 'short_answer'">
              <textarea
                v-model="answers[currentIndex]"
                rows="6"
                placeholder="请输入你的回答..."
                class="w-full px-4 py-3 border-2 border-gray-200 dark:border-gray-600 rounded-lg text-sm focus:border-indigo-500 dark:focus:border-indigo-400 focus:outline-none bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 resize-none transition-colors"
              ></textarea>
              <p class="text-xs text-gray-400 dark:text-gray-500 mt-1 text-right">
                {{ (currentAnswer || '').length }} 字
              </p>
            </div>

            <!-- 底部导航 -->
            <div class="flex items-center justify-between mt-8 pt-4 border-t border-gray-100 dark:border-gray-700">
              <button
                :disabled="currentIndex === 0"
                class="px-4 py-2 text-sm font-medium rounded-lg transition-colors disabled:opacity-40"
                :class="
                  currentIndex === 0 ? '' : 'text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                "
                @click="prevQuestion"
              >
                上一题
              </button>
              <button
                v-if="currentIndex < questions.length - 1"
                class="px-6 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
                @click="nextQuestion"
              >
                下一题
              </button>
              <button
                v-else
                class="px-6 py-2 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 transition-colors"
                @click="submitExamNow"
              >
                完成交卷
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ===== 阶段 3：考试报告 ===== -->
    <div v-else-if="phase === 'report' && report" class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">考试报告</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">{{ formatDate(examStartedAt) }}</p>
        </div>
        <button
          class="px-4 py-2 text-sm font-medium text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-900/30 rounded-lg hover:bg-indigo-100 dark:hover:bg-indigo-900/50 transition-colors"
          @click="backToConfig"
        >
          再考一次
        </button>
      </div>

      <!-- 总分卡片 -->
      <div class="bg-gradient-to-r from-indigo-600 to-purple-600 rounded-xl p-6 text-white">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-indigo-200">总分</p>
            <div class="flex items-baseline gap-2 mt-1">
              <span class="text-5xl font-bold">{{ report.total_score }}</span>
              <span class="text-xl text-indigo-200">/ {{ report.max_score }}</span>
            </div>
            <p class="text-sm text-indigo-200 mt-2">
              正确率 {{ report.percentage }}% · 用时 {{ formatTime(report.time_used) }}
            </p>
          </div>
          <div class="text-right">
            <div
              class="w-24 h-24 rounded-full border-4 flex items-center justify-center"
              :class="
                report.percentage >= 80
                  ? 'border-green-400'
                  : report.percentage >= 60
                    ? 'border-yellow-400'
                    : 'border-red-400'
              "
            >
              <div>
                <span class="text-3xl font-bold">{{ report.percentage }}</span>
                <span class="text-sm text-indigo-200 block">%</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 题型统计 -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div
          v-for="(stat, typeName) in report.type_stats"
          :key="typeName"
          class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 text-center"
        >
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ typeLabel(typeName) }}</p>
          <p
            class="text-2xl font-bold mt-1"
            :class="
              stat.rate >= 0.8
                ? 'text-green-600 dark:text-green-400'
                : stat.rate >= 0.6
                  ? 'text-yellow-600 dark:text-yellow-400'
                  : 'text-red-600 dark:text-red-400'
            "
          >
            {{ Math.round(stat.rate * 100) }}%
          </p>
          <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">{{ stat.correct }}/{{ stat.total }}</p>
          <div class="w-full h-1.5 bg-gray-100 dark:bg-gray-700 rounded-full mt-2 overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-500"
              :class="stat.rate >= 0.8 ? 'bg-green-500' : stat.rate >= 0.6 ? 'bg-yellow-500' : 'bg-red-500'"
              :style="{ width: Math.round(stat.rate * 100) + '%' }"
            ></div>
          </div>
        </div>
      </div>

      <!-- 薄弱知识点 -->
      <div
        v-if="report.weak_points && report.weak_points.length > 0"
        class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 sm:p-6"
      >
        <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 flex items-center gap-2">
          <svg class="w-4 h-4 text-red-500" fill="currentColor" viewBox="0 0 20 20">
            <path
              fill-rule="evenodd"
              d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
              clip-rule="evenodd"
            />
          </svg>
          薄弱知识点
        </h3>
        <div class="space-y-2">
          <div
            v-for="(wp, wi) in report.weak_points"
            :key="wi"
            class="flex items-start gap-3 p-3 bg-red-50 dark:bg-red-900/20 rounded-lg border border-red-100 dark:border-red-900/30"
          >
            <span
              class="flex-shrink-0 w-5 h-5 rounded-full bg-red-100 dark:bg-red-900/40 text-red-600 dark:text-red-400 text-xs font-bold flex items-center justify-center mt-0.5"
            >
              {{ wi + 1 }}
            </span>
            <div>
              <p class="text-sm font-medium text-gray-900 dark:text-white">{{ wp.concept }}</p>
              <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ wp.reason }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 逐题对答案 -->
      <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
        <div class="px-5 py-3 border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
          <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">逐题详解</h3>
        </div>
        <div class="divide-y divide-gray-100 dark:divide-gray-700">
          <div
            v-for="qr in report.question_results"
            :key="qr.index"
            class="p-5 hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
          >
            <div class="flex items-start gap-3">
              <span
                class="flex-shrink-0 w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold"
                :class="
                  qr.is_correct
                    ? 'bg-green-100 dark:bg-green-900/40 text-green-600 dark:text-green-400'
                    : 'bg-red-100 dark:bg-red-900/40 text-red-600 dark:text-red-400'
                "
              >
                {{ qr.index }}
              </span>
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-1">
                  <span class="px-2 py-0.5 rounded text-[10px] font-medium" :class="typeClass(qr.type)">
                    {{ typeLabel(qr.type) }}
                  </span>
                  <span v-if="qr.concept" class="text-xs text-gray-400 dark:text-gray-500">{{ qr.concept }}</span>
                </div>
                <!-- 答案对比 -->
                <div class="space-y-1.5 mt-2">
                  <div class="flex items-start gap-2">
                    <span class="text-xs text-gray-400 dark:text-gray-500 w-12 flex-shrink-0 pt-0.5">你的:</span>
                    <span
                      class="text-sm"
                      :class="qr.is_correct ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'"
                    >
                      {{ qr.user_answer || '(未作答)' }}
                    </span>
                  </div>
                  <div v-if="!qr.is_correct" class="flex items-start gap-2">
                    <span class="text-xs text-gray-400 dark:text-gray-500 w-12 flex-shrink-0 pt-0.5">正确:</span>
                    <span class="text-sm text-green-600 dark:text-green-400 font-medium">{{ qr.correct_answer }}</span>
                  </div>
                </div>
                <!-- 解析 -->
                <div v-if="qr.explanation" class="mt-3 p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                  <p class="text-xs text-gray-500 dark:text-gray-400 mb-1 font-medium">解析</p>
                  <p class="text-sm text-gray-700 dark:text-gray-300 leading-relaxed">{{ qr.explanation }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useToast, useConfirm } from '../composables/useToast'
import { listMaterials, generateExam, submitExam, listExams, getExam } from '../api/client'

const _router = useRouter()
const toast = useToast()
const { confirm } = useConfirm()

// ===== 状态 =====
const phase = ref('config') // config, exam, report
const generating = ref(false)
const showHistory = ref(false)
const loadingMaterials = ref(true)

// 配置
const materials = ref([])
const selectedMaterialIDs = ref([])
const questionCount = ref(20)
const timeLimit = ref(30)
const examHistory = ref([])

// 考试进行
const examId = ref('')
const questions = ref([])
const answers = ref({})
const markedSet = ref(new Set())
const currentIndex = ref(0)
const remainingSeconds = ref(0)
const examStartedAt = ref(null)
let timer = null

// 报告
const report = ref(null)

// ===== 计算 =====
const currentQuestion = computed(() => questions.value[currentIndex.value] || {})
const currentAnswer = computed(() => answers.value[currentIndex.value] || '')
const isCurrentMarked = computed(() => markedSet.value.has(currentIndex.value))
const answeredCount = computed(() => Object.values(answers.value).filter((a) => a && a.trim()).length)
const markedCount = computed(() => markedSet.value.size)

// ===== 生命周期 =====
onMounted(() => {
  loadMaterials()
  loadHistory()
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

// ===== 数据加载 =====
async function loadMaterials() {
  loadingMaterials.value = true
  try {
    const res = await listMaterials({ limit: 100 })
    materials.value = (res.data.data || []).filter((m) => m.status === 'completed' || m.status === 'partial')
  } catch {
    toast.error('加载材料失败')
  } finally {
    loadingMaterials.value = false
  }
}

async function loadHistory() {
  try {
    const res = await listExams()
    examHistory.value = res.data.data || []
  } catch {
    // ignore
  }
}

// ===== 考试流程 =====
async function startExam() {
  if (selectedMaterialIDs.value.length === 0) return
  generating.value = true
  try {
    const res = await generateExam({
      material_ids: selectedMaterialIDs.value,
      question_count: questionCount.value,
      time_limit: timeLimit.value
    })
    const data = res.data
    examId.value = data.id
    questions.value = data.questions
    answers.value = {}
    markedSet.value = new Set()
    currentIndex.value = 0
    remainingSeconds.value = data.time_limit * 60
    examStartedAt.value = data.started_at

    // 启动倒计时
    phase.value = 'exam'
    startTimer()
  } catch (err) {
    toast.error(err.response?.data?.error || '生成试卷失败')
  } finally {
    generating.value = false
  }
}

function startTimer() {
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    remainingSeconds.value--
    if (remainingSeconds.value <= 0) {
      clearInterval(timer)
      remainingSeconds.value = 0
      toast.warning('时间到，自动交卷')
      doSubmit()
    }
  }, 1000)
}

async function submitExamNow() {
  const unanswered = questions.value.length - answeredCount.value
  if (unanswered > 0) {
    const ok = await confirm(`还有 ${unanswered} 题未作答，确定要交卷吗？`)
    if (!ok) return
  }
  doSubmit()
}

async function doSubmit() {
  if (timer) clearInterval(timer)

  const answerList = questions.value.map((q, i) => ({
    index: q.index,
    answer: answers.value[i] || '',
    marked: markedSet.value.has(i)
  }))

  try {
    const res = await submitExam(examId.value, answerList)
    report.value = res.data.report
    phase.value = 'report'
    loadHistory()
  } catch (err) {
    toast.error(err.response?.data?.error || '提交失败')
  }
}

async function confirmExit() {
  const ok = await confirm('确定要退出考试吗？当前进度将丢失。')
  if (ok) {
    if (timer) clearInterval(timer)
    phase.value = 'config'
  }
}

async function viewExamReport(id) {
  try {
    const res = await getExam(id)
    const data = res.data
    if (data.report) {
      report.value = data.report
      examStartedAt.value = data.started_at
      phase.value = 'report'
    } else {
      toast.warning('该考试尚未完成')
    }
  } catch {
    toast.error('加载考试详情失败')
  }
}

function backToConfig() {
  phase.value = 'config'
  report.value = null
}

// ===== 题目操作 =====
function nextQuestion() {
  if (currentIndex.value < questions.value.length - 1) {
    currentIndex.value++
  }
}

function prevQuestion() {
  if (currentIndex.value > 0) {
    currentIndex.value--
  }
}

function toggleMark() {
  const idx = currentIndex.value
  if (markedSet.value.has(idx)) {
    markedSet.value.delete(idx)
  } else {
    markedSet.value.add(idx)
  }
  // 触发响应式
  markedSet.value = new Set(markedSet.value)
}

function getOptionLetter(index) {
  return ['A', 'B', 'C', 'D'][index] || String.fromCharCode(65 + index)
}

// ===== 样式辅助 =====
function getNavButtonClass(i) {
  const isAnswered = answers.value[i] && answers.value[i].toString().trim()
  const isMarked = markedSet.value.has(i)
  const isCurrent = i === currentIndex.value

  if (isCurrent) return 'bg-indigo-600 text-white ring-2 ring-indigo-300 dark:ring-indigo-500'
  if (isMarked) return 'bg-amber-400 text-white'
  if (isAnswered) return 'bg-green-500 text-white'
  return 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600'
}

function typeClass(type) {
  switch (type) {
    case 'choice':
      return 'bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300'
    case 'true_false':
      return 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-300'
    case 'fill':
      return 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300'
    case 'short_answer':
      return 'bg-purple-100 dark:bg-purple-900/40 text-purple-700 dark:text-purple-300'
    default:
      return 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
  }
}

function typeLabel(type) {
  switch (type) {
    case 'choice':
      return '选择题'
    case 'true_false':
      return '判断题'
    case 'fill':
      return '填空题'
    case 'short_answer':
      return '简答题'
    default:
      return '题目'
  }
}

function diffClass(diff) {
  switch (diff) {
    case 'easy':
      return 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-300'
    case 'medium':
      return 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300'
    case 'hard':
      return 'bg-red-100 dark:bg-red-900/40 text-red-700 dark:text-red-300'
    default:
      return 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400'
  }
}

function diffLabel(diff) {
  switch (diff) {
    case 'easy':
      return '简单'
    case 'medium':
      return '中等'
    case 'hard':
      return '困难'
    default:
      return ''
  }
}

function formatTime(seconds) {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

function formatDate(d) {
  if (!d) return ''
  const date = new Date(d)
  return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
}

function renderQuestion(text) {
  if (!text) return ''
  return text.replace(/\n/g, '<br>')
}
</script>

<style scoped>
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
