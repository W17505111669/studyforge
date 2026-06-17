<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-4xl mx-auto h-[calc(100vh-2rem)] flex flex-col">
    <!-- 头部 -->
    <div class="flex items-center justify-between mb-4 flex-shrink-0">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-gray-100">Agent 辩论</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">多视角讨论知识点 — 分析师 · 出题官 · 记忆大师</p>
      </div>
      <button
        v-if="messages.length > 0"
        class="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-500 hover:text-red-500 dark:text-gray-400 dark:hover:text-red-400 transition-colors rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20"
        @click="resetDebate"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
          />
        </svg>
        新辩论
      </button>
    </div>

    <!-- 辩论消息区 -->
    <div
      ref="messagesContainer"
      class="flex-1 overflow-y-auto custom-scroll rounded-xl border p-4 sm:p-6 space-y-6"
      :class="isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'"
    >
      <!-- 空状态 -->
      <div
        v-if="messages.length === 0 && !loading"
        class="h-full flex flex-col items-center justify-center text-center py-12"
      >
        <div
          class="w-20 h-20 rounded-2xl flex items-center justify-center mb-6"
          :class="isDark ? 'bg-primary-900/30' : 'bg-primary-50'"
        >
          <svg
            class="w-10 h-10"
            :class="isDark ? 'text-primary-400' : 'text-primary-500'"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.5"
              d="M17 8h2a2 2 0 012 2v6a2 2 0 01-2 2h-2v4l-4-4H9a1.994 1.994 0 01-1.414-.586m0 0L11 14h4a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2v4l.586-.586z"
            />
          </svg>
        </div>
        <h3 class="text-lg font-semibold mb-2" :class="isDark ? 'text-gray-200' : 'text-gray-800'">多 Agent 辩论</h3>
        <p class="text-sm max-w-md mb-6" :class="isDark ? 'text-gray-400' : 'text-gray-500'">
          输入一个知识点或概念，三位 AI 专家将从不同角度展开讨论：
          <br />
          <span class="text-blue-500">分析师</span>
          解析本质、
          <span class="text-amber-500">出题官</span>
          分析考法、
          <span class="text-emerald-500">记忆大师</span>
          设计记忆策略
        </p>
        <!-- 推荐主题 -->
        <div class="flex flex-wrap gap-2 justify-center max-w-lg">
          <button
            v-for="topic in suggestedTopics"
            :key="topic"
            class="px-3 py-1.5 text-xs rounded-full border transition-colors"
            :class="
              isDark
                ? 'border-gray-600 text-gray-300 hover:bg-gray-700 hover:border-gray-500'
                : 'border-gray-300 text-gray-600 hover:bg-gray-50 hover:border-gray-400'
            "
            @click="selectTopic(topic)"
          >
            {{ topic }}
          </button>
        </div>
      </div>

      <!-- 辩论消息列表 -->
      <div v-for="(msg, idx) in messages" :key="idx" class="debate-message">
        <!-- 轮次分隔 -->
        <div v-if="msg.round !== messages[idx - 1]?.round" class="flex items-center gap-3 mb-4">
          <div class="flex-1 h-px" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'"></div>
          <span
            class="text-xs font-medium px-3 py-1 rounded-full"
            :class="isDark ? 'bg-gray-700 text-gray-400' : 'bg-gray-100 text-gray-500'"
          >
            {{ roundLabel(msg.round) }}
          </span>
          <div class="flex-1 h-px" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'"></div>
        </div>

        <!-- 消息卡片 -->
        <div class="rounded-xl border p-4 sm:p-5 transition-all" :class="messageCardClass(msg)">
          <!-- Agent 头部 -->
          <div class="flex items-center gap-3 mb-3">
            <div
              class="w-9 h-9 rounded-full flex items-center justify-center text-white text-sm font-bold flex-shrink-0"
              :class="avatarClass(msg)"
            >
              {{ avatarLabel(msg) }}
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <span class="text-sm font-semibold" :class="isDark ? 'text-gray-100' : 'text-gray-900'">
                  {{ msg.role }}
                </span>
                <span class="text-[10px] px-1.5 py-0.5 rounded-full font-medium" :class="badgeClass(msg)">
                  {{ msg.agent_name }}
                </span>
              </div>
              <span class="text-xs" :class="isDark ? 'text-gray-500' : 'text-gray-400'">
                {{ msg.duration_ms > 0 ? (msg.duration_ms / 1000).toFixed(1) + 's' : '' }}
              </span>
            </div>
          </div>
          <!-- 消息内容（Markdown 渲染） -->
          <div
            class="prose prose-sm max-w-none debate-content"
            :class="isDark ? 'dark' : ''"
            v-html="renderMarkdown(msg.content)"
          ></div>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="debate-message">
        <div class="flex items-center gap-3 mb-4">
          <div class="flex-1 h-px" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'"></div>
          <span
            class="text-xs font-medium px-3 py-1 rounded-full"
            :class="isDark ? 'bg-gray-700 text-gray-400' : 'bg-gray-100 text-gray-500'"
          >
            {{ loadingLabel }}
          </span>
          <div class="flex-1 h-px" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'"></div>
        </div>
        <div
          class="rounded-xl border p-5 flex items-center gap-4"
          :class="isDark ? 'bg-gray-800/50 border-gray-700' : 'bg-gray-50 border-gray-200'"
        >
          <div class="w-9 h-9 rounded-full flex items-center justify-center flex-shrink-0" :class="loadingAvatarClass">
            <div
              class="w-5 h-5 border-2 border-t-transparent rounded-full animate-spin"
              :class="loadingSpinnerClass"
            ></div>
          </div>
          <div>
            <p class="text-sm font-medium" :class="isDark ? 'text-gray-300' : 'text-gray-700'">
              {{ loadingRole }} 正在思考...
            </p>
            <p class="text-xs mt-0.5" :class="isDark ? 'text-gray-500' : 'text-gray-400'">{{ loadingHint }}</p>
          </div>
        </div>
      </div>

      <!-- 错误状态 -->
      <div
        v-if="error"
        class="rounded-xl border p-4 flex items-start gap-3"
        :class="isDark ? 'bg-red-900/20 border-red-800' : 'bg-red-50 border-red-200'"
      >
        <svg class="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <div>
          <p class="text-sm font-medium text-red-600 dark:text-red-400">辩论生成失败</p>
          <p class="text-xs mt-1" :class="isDark ? 'text-red-400/70' : 'text-red-500'">{{ error }}</p>
          <button class="mt-2 text-xs text-red-500 hover:text-red-600 underline" @click="startDebate()">重试</button>
        </div>
      </div>
    </div>

    <!-- 底部输入区 -->
    <div class="flex-shrink-0 mt-4">
      <form class="flex gap-3" @submit.prevent="startDebate()">
        <input
          v-model="concept"
          type="text"
          placeholder="输入一个知识点或概念，如：TCP 三次握手、二叉树遍历、牛顿第三定律..."
          class="flex-1 px-4 py-3 rounded-xl border text-sm outline-none transition-colors"
          :class="
            isDark
              ? 'bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500 focus:border-primary-500'
              : 'bg-white border-gray-300 text-gray-900 placeholder-gray-400 focus:border-primary-500'
          "
          :disabled="loading"
          maxlength="500"
        />
        <button
          type="submit"
          :disabled="loading || !concept.trim()"
          class="px-5 py-3 bg-primary-600 hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed text-white text-sm font-medium rounded-xl transition-colors flex items-center gap-2 flex-shrink-0"
        >
          <svg v-if="!loading" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <div v-else class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
          {{ loading ? '辩论中...' : '开始辩论' }}
        </button>
      </form>
      <p class="text-xs text-center mt-2" :class="isDark ? 'text-gray-600' : 'text-gray-400'">
        三位 AI 专家将依次从概念分析、考试应用、记忆策略三个角度展开讨论
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { startDebate as apiStartDebate } from '../api/client'
import { useToast } from '../composables/useToast'
import { useDarkMode } from '../composables/useDarkMode'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const route = useRoute()
const { isDark } = useDarkMode()
const toast = useToast()

const concept = ref('')
const messages = ref([])
const loading = ref(false)
const error = ref('')
const messagesContainer = ref(null)
const _currentRound = ref(0)

// 推荐辩论主题
const suggestedTopics = [
  'TCP 三次握手',
  '二叉搜索树',
  '牛顿第三定律',
  'HTTP 状态码',
  '递归与迭代',
  '贝叶斯定理',
  '操作系统进程调度',
  '深度学习反向传播'
]

// 加载状态提示
const loadingRound = ref(1)
const loadingAgentName = ref('Analyst')

const loadingLabel = computed(() => {
  const labels = { 1: '第一轮', 2: '第二轮', 3: '第三轮', 4: '总结' }
  return labels[loadingRound.value] || '生成中'
})

const loadingRole = computed(() => {
  const roles = { Analyst: '分析师', QuizMaster: '出题官', CardMaker: '记忆大师', Summary: '总结员' }
  return roles[loadingAgentName.value] || '专家'
})

const loadingHint = computed(() => {
  const hints = {
    Analyst: '正在深入分析概念的本质结构和核心原理...',
    QuizMaster: '正在从出题者角度分析易错点和考察重点...',
    CardMaker: '正在设计记忆策略和学习方法...',
    Summary: '正在综合三位专家的观点撰写总结...'
  }
  return hints[loadingAgentName.value] || '正在思考...'
})

const loadingAvatarClass = computed(() => {
  const classes = {
    Analyst: 'bg-blue-500',
    QuizMaster: 'bg-amber-500',
    CardMaker: 'bg-emerald-500',
    Summary: 'bg-purple-500'
  }
  return classes[loadingAgentName.value] || 'bg-gray-500'
})

const loadingSpinnerClass = computed(() => {
  const classes = {
    Analyst: 'border-blue-200',
    QuizMaster: 'border-amber-200',
    CardMaker: 'border-emerald-200',
    Summary: 'border-purple-200'
  }
  return classes[loadingAgentName.value] || 'border-gray-200'
})

// Agent 配置
const agentConfig = {
  Analyst: {
    role: '分析师',
    initials: '析',
    avatarBg: 'bg-blue-500',
    cardBg: 'bg-blue-50 dark:bg-blue-900/10',
    cardBorder: 'border-blue-200 dark:border-blue-800/50',
    badgeBg: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
  },
  QuizMaster: {
    role: '出题官',
    initials: '题',
    avatarBg: 'bg-amber-500',
    cardBg: 'bg-amber-50 dark:bg-amber-900/10',
    cardBorder: 'border-amber-200 dark:border-amber-800/50',
    badgeBg: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
  },
  CardMaker: {
    role: '记忆大师',
    initials: '记',
    avatarBg: 'bg-emerald-500',
    cardBg: 'bg-emerald-50 dark:bg-emerald-900/10',
    cardBorder: 'border-emerald-200 dark:border-emerald-800/50',
    badgeBg: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
  },
  Summary: {
    role: '总结员',
    initials: '总',
    avatarBg: 'bg-purple-500',
    cardBg: 'bg-purple-50 dark:bg-purple-900/10',
    cardBorder: 'border-purple-200 dark:border-purple-800/50',
    badgeBg: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400'
  }
}

function messageCardClass(msg) {
  const cfg = agentConfig[msg.agent_name]
  if (!cfg) return isDark.value ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'
  return `${cfg.cardBg} ${cfg.cardBorder}`
}

function avatarClass(msg) {
  const cfg = agentConfig[msg.agent_name]
  return cfg ? cfg.avatarBg : 'bg-gray-500'
}

function avatarLabel(msg) {
  const cfg = agentConfig[msg.agent_name]
  return cfg ? cfg.initials : '?'
}

function badgeClass(msg) {
  const cfg = agentConfig[msg.agent_name]
  return cfg ? cfg.badgeBg : isDark.value ? 'bg-gray-700 text-gray-400' : 'bg-gray-100 text-gray-600'
}

function roundLabel(round) {
  const labels = { 1: '第一轮 · 分析师', 2: '第二轮 · 出题官', 3: '第三轮 · 记忆大师', 4: '综合总结' }
  return labels[round] || `第 ${round} 轮`
}

// Markdown 渲染
function renderMarkdown(text) {
  if (!text) return ''
  const html = marked.parse(text)
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: [
      'h1',
      'h2',
      'h3',
      'h4',
      'h5',
      'h6',
      'p',
      'br',
      'strong',
      'em',
      'u',
      's',
      'del',
      'ul',
      'ol',
      'li',
      'a',
      'code',
      'pre',
      'blockquote',
      'table',
      'thead',
      'tbody',
      'tr',
      'th',
      'td',
      'span',
      'div',
      'hr',
      'sub',
      'sup'
    ],
    ALLOWED_ATTR: ['href', 'target', 'rel', 'class', 'id'],
    ALLOW_DATA_ATTR: false
  })
}

// 滚动到底部
function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

// 模拟逐轮加载状态
const _loadingAgents = ['Analyst', 'QuizMaster', 'CardMaker', 'Summary']

function simulateRoundProgress() {
  // 基于当前消息数量推断正在加载的轮次
  const msgCount = messages.value.length
  if (msgCount < 1) {
    loadingRound.value = 1
    loadingAgentName.value = 'Analyst'
  } else if (msgCount < 2) {
    loadingRound.value = 2
    loadingAgentName.value = 'QuizMaster'
  } else if (msgCount < 3) {
    loadingRound.value = 3
    loadingAgentName.value = 'CardMaker'
  } else {
    loadingRound.value = 4
    loadingAgentName.value = 'Summary'
  }
}

// 选择推荐主题并开始辩论
function selectTopic(topic) {
  concept.value = topic
  startDebate()
}

// 发起辩论
async function startDebate() {
  const text = concept.value.trim()
  if (!text || loading.value) return

  loading.value = true
  error.value = ''
  messages.value = []
  loadingRound.value = 1
  loadingAgentName.value = 'Analyst'
  scrollToBottom()

  // 模拟轮次进度（每隔一段时间更新 loading 提示）
  const roundTimer = setInterval(() => {
    simulateRoundProgress()
  }, 500)

  try {
    const res = await apiStartDebate(text)
    const data = res.data

    // 逐条添加消息，带延迟动画效果
    if (data.messages && data.messages.length > 0) {
      for (let i = 0; i < data.messages.length; i++) {
        messages.value.push(data.messages[i])
        loadingRound.value = data.messages[i].round
        loadingAgentName.value = data.messages[i].agent_name
        scrollToBottom()
        // 每条消息之间短暂延迟，营造逐步展开效果
        if (i < data.messages.length - 1) {
          await new Promise((r) => setTimeout(r, 300))
        }
      }
    }

    concept.value = ''
    toast.success('辩论完成')
  } catch (err) {
    console.error('辩论失败:', err)
    error.value = err.response?.data?.error || err.message || '未知错误'
    toast.error('辩论生成失败')
  } finally {
    clearInterval(roundTimer)
    loading.value = false
    scrollToBottom()
  }
}

// 重置辩论
function resetDebate() {
  messages.value = []
  error.value = ''
  concept.value = ''
}

// 监听消息变化自动滚动
watch(
  messages,
  () => {
    scrollToBottom()
  },
  { deep: true }
)

onMounted(() => {
  // 从 URL 查询参数读取概念（知识洞察"深入学习"跳转）
  if (route.query.concept) {
    concept.value = route.query.concept
  }
})
</script>

<style scoped>
/* 自定义滚动条 */
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

/* Markdown 内容样式 */
:deep(.debate-content) {
  color: inherit;
}
:deep(.debate-content h1),
:deep(.debate-content h2),
:deep(.debate-content h3) {
  margin-top: 0.75em;
  margin-bottom: 0.5em;
  font-weight: 600;
}
:deep(.debate-content h1) {
  font-size: 1.25em;
}
:deep(.debate-content h2) {
  font-size: 1.125em;
}
:deep(.debate-content h3) {
  font-size: 1em;
}
:deep(.debate-content p) {
  margin-bottom: 0.5em;
  line-height: 1.7;
}
:deep(.debate-content ul),
:deep(.debate-content ol) {
  padding-left: 1.5em;
  margin-bottom: 0.5em;
}
:deep(.debate-content li) {
  margin-bottom: 0.25em;
}
:deep(.debate-content code) {
  font-size: 0.85em;
  padding: 0.15em 0.4em;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.06);
}
:deep(.debate-content pre) {
  margin: 0.5em 0;
  padding: 0.75em 1em;
  border-radius: 8px;
  overflow-x: auto;
  background: #1e293b;
  color: #e2e8f0;
}
:deep(.debate-content pre code) {
  background: none;
  padding: 0;
  color: inherit;
}
:deep(.debate-content blockquote) {
  border-left: 3px solid;
  padding-left: 1em;
  margin: 0.5em 0;
  opacity: 0.85;
}
:deep(.debate-content strong) {
  font-weight: 600;
}

/* 暗色模式 Markdown */
:deep(.dark .debate-content code) {
  background: rgba(255, 255, 255, 0.08);
  color: #c4b5fd;
}
:deep(.dark .debate-content pre) {
  background: #0f172a;
}
:deep(.dark .debate-content blockquote) {
  border-color: rgba(255, 255, 255, 0.15);
}

/* 消息入场动画 */
.debate-message {
  animation: message-in 0.4s ease-out;
}
@keyframes message-in {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
