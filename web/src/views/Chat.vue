<template>
  <div class="flex h-[calc(100vh-2rem)] p-4 max-w-6xl mx-auto gap-4">
    <!-- 左侧：对话历史列表 -->
    <div
      class="w-64 flex-shrink-0 flex flex-col bg-white rounded-xl border border-gray-200 overflow-hidden dark:bg-gray-800 dark:border-gray-700"
    >
      <div class="p-3 border-b border-gray-100 dark:border-gray-700">
        <button
          class="w-full flex items-center justify-center gap-2 px-3 py-2.5 bg-primary-600 hover:bg-primary-700 text-white text-sm font-medium rounded-lg transition-colors"
          @click="startNewConversation"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          新对话
        </button>
      </div>
      <div class="flex-1 overflow-y-auto">
        <div v-if="conversations.length === 0" class="p-4 text-center text-gray-400 dark:text-gray-500 text-xs">
          暂无对话历史
        </div>
        <div
          v-for="conv in conversations"
          :key="conv.id"
          class="group flex items-center gap-2 px-3 py-2.5 cursor-pointer border-b border-gray-50 transition-colors dark:border-gray-700"
          :class="
            conv.id === currentConvId
              ? 'bg-primary-50 border-l-2 border-l-primary-500 dark:bg-primary-900/20'
              : 'hover:bg-gray-50 dark:hover:bg-gray-700'
          "
          @click="switchConversation(conv.id)"
        >
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-gray-800 dark:text-gray-200 truncate">{{ conv.title || '新对话' }}</p>
            <p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">{{ conv.message_count }} 条消息</p>
          </div>
          <button
            class="opacity-0 group-hover:opacity-100 p-1 text-gray-400 hover:text-red-500 rounded transition-all"
            title="删除对话"
            @click.stop="confirmDeleteConversation(conv.id)"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- 右侧：对话主区域 -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- 头部 -->
      <div class="flex items-center justify-between mb-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">AI 学习助手</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">流式输出 · 多轮对话 · RAG 检索 · 工具调用</p>
          <div class="flex items-center gap-2 mt-1.5 text-[11px] text-gray-400 dark:text-gray-500">
            <span class="flex items-center gap-0.5">
              <kbd class="px-1 py-0.5 rounded border border-gray-200 dark:border-gray-600 font-mono text-[10px]">/</kbd>
              聚焦
            </span>
            <span class="flex items-center gap-0.5">
              <kbd class="px-1 py-0.5 rounded border border-gray-200 dark:border-gray-600 font-mono text-[10px]">
                Enter
              </kbd>
              发送
            </span>
            <span class="flex items-center gap-0.5">
              <kbd class="px-1 py-0.5 rounded border border-gray-200 dark:border-gray-600 font-mono text-[10px]">
                Ctrl+K
              </kbd>
              搜索
            </span>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <div v-if="materials.length > 0" class="flex items-center gap-2">
            <label class="text-xs text-gray-500 dark:text-gray-400">关联材料：</label>
            <select
              v-model="selectedMaterialId"
              class="text-sm border rounded-lg px-3 py-1.5 text-gray-700 focus:border-primary-500 outline-none dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200"
            >
              <option value="">不关联</option>
              <option v-for="m in materials" :key="m.id" :value="m.id">{{ m.title }}</option>
            </select>
          </div>
          <button
            v-if="messages.length > 0 && !isStreaming"
            class="text-xs text-gray-400 hover:text-red-500 transition-colors flex items-center gap-1 px-2 py-1.5 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20"
            title="清空对话"
            @click="clearConversation"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
            清空
          </button>
        </div>
      </div>

      <!-- 消息区域 -->
      <div ref="messagesContainer" class="flex-1 overflow-y-auto space-y-4 mb-4 scroll-smooth" @click="handleCopyCode">
        <div
          v-if="messages.length === 0"
          class="flex flex-col items-center justify-center h-full text-gray-400 dark:text-gray-500"
        >
          <p class="text-5xl mb-4">🤖</p>
          <p class="text-lg font-medium">你好！我是你的 AI 学习助手</p>
          <p class="text-sm mt-2">上传学习材料后，我可以帮你解答问题、生成练习题、制定复习计划</p>
          <div class="flex flex-wrap gap-2 mt-6 justify-center">
            <button
              v-for="q in quickQuestions"
              :key="q"
              class="px-4 py-2 text-sm bg-gray-100 hover:bg-primary-50 hover:text-primary-600 rounded-full transition-colors dark:bg-gray-700 dark:text-gray-300"
              @click="sendQuickQuestion(q)"
            >
              {{ q }}
            </button>
          </div>
        </div>

        <!-- 虚拟滚动：顶部占位 -->
        <div v-if="msgVirtualized" :style="{ height: msgTopSpacer + 'px' }" aria-hidden="true"></div>

        <div
          v-for="(msg, idx) in visibleMessages"
          :key="msgStartIdx + idx"
          class="flex"
          :class="msg.role === 'user' ? 'justify-end' : 'justify-start'"
        >
          <div
            class="max-w-[80%] rounded-2xl px-4 py-3 text-sm leading-relaxed"
            :class="
              msg.role === 'user'
                ? 'bg-primary-600 text-white rounded-br-sm'
                : 'bg-white border border-gray-200 text-gray-800 rounded-bl-sm shadow-sm dark:bg-gray-800 dark:border-gray-700 dark:text-gray-200'
            "
          >
            <p v-if="msg.role === 'user'" class="whitespace-pre-wrap">{{ msg.content }}</p>
            <template v-else>
              <!-- 工具调用状态面板 -->
              <div v-if="msg.toolEvents && msg.toolEvents.length > 0" class="mb-2 space-y-1">
                <div
                  v-for="(evt, eidx) in msg.toolEvents"
                  :key="eidx"
                  class="flex items-center gap-2 text-xs px-3 py-1.5 rounded-lg bg-gray-50 text-gray-600 border border-gray-100 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-400"
                >
                  <svg
                    v-if="!evt.done"
                    class="w-3.5 h-3.5 text-primary-500 animate-spin flex-shrink-0"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                  </svg>
                  <svg
                    v-else
                    class="w-3.5 h-3.5 text-green-500 flex-shrink-0"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                  <span class="font-medium">{{ getToolLabel(evt.name) }}</span>
                  <span class="text-gray-400 truncate max-w-[160px]">{{ evt.argsStr || '' }}</span>
                </div>
              </div>
              <div
                class="prose prose-sm max-w-none prose-headings:mt-3 prose-headings:mb-1.5 prose-p:my-1.5 prose-pre:bg-gray-900 prose-pre:text-gray-100 prose-pre:rounded-lg prose-pre:text-xs prose-code:text-primary-600 prose-code:before:content-[''] prose-code:after:content-[''] prose-ul:my-1.5 prose-ol:my-1.5 prose-li:my-0.5 dark:text-gray-200"
                v-html="renderMarkdown(msg.content)"
              ></div>
            </template>
            <div
              v-if="msg.streaming"
              class="inline-block w-1 h-4 bg-primary-600 animate-pulse ml-0.5 align-middle"
            ></div>
          </div>
        </div>

        <!-- 虚拟滚动：底部占位 -->
        <div v-if="msgVirtualized" :style="{ height: msgBottomSpacer + 'px' }" aria-hidden="true"></div>

        <!-- 正在输入的指示器 -->
        <div v-if="isThinking" class="flex justify-start">
          <div
            class="bg-white border border-gray-200 rounded-2xl rounded-bl-sm px-4 py-3 shadow-sm dark:bg-gray-800 dark:border-gray-700"
          >
            <div class="flex gap-1">
              <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0ms"></span>
              <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 150ms"></span>
              <span class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 300ms"></span>
            </div>
          </div>
        </div>
      </div>

      <!-- 离线提示 -->
      <div
        v-if="!isOnline"
        class="flex items-center gap-2 px-3 py-2 rounded-lg text-xs font-medium bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-900/20 dark:text-amber-400 dark:border-amber-800/30 mb-2"
      >
        <svg class="w-3.5 h-3.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M18.364 5.636a9 9 0 010 12.728M5.636 5.636a9 9 0 000 12.728"
          />
          <line x1="1" y1="1" x2="23" y2="23" stroke-width="2" stroke-linecap="round" />
        </svg>
        <span>AI 对话需要网络连接，请恢复网络后使用</span>
      </div>

      <!-- 输入区域 -->
      <div class="flex gap-3">
        <input
          ref="inputRef"
          v-model="inputMessage"
          :disabled="isStreaming || !isOnline"
          type="text"
          class="flex-1 px-4 py-3 rounded-xl border border-gray-200 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-all text-sm disabled:opacity-50 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200"
          :placeholder="isOnline ? '输入你的问题... (/ 聚焦 · Ctrl+Enter 发送)' : '当前离线，无法发送消息'"
          @keydown.enter="sendMessage"
        />
        <button
          v-if="!isStreaming"
          :disabled="!inputMessage.trim()"
          class="px-6 py-3 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-xl transition-colors disabled:opacity-40 text-sm"
          @click="sendMessage"
        >
          发送
        </button>
        <button
          v-else
          class="px-6 py-3 bg-red-500 hover:bg-red-600 text-white font-medium rounded-xl transition-colors text-sm"
          @click="stopStreaming"
        >
          停止
        </button>
      </div>
    </div>
  </div>

  <!-- AI 解释浮动按钮（选中文字时出现） -->
  <Teleport to="body">
    <button
      v-if="buttonPos.visible"
      class="explain-float-btn fixed z-50 flex items-center gap-1.5 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-medium rounded-lg shadow-lg transition-all animate-in"
      :style="{ left: buttonPos.x + 'px', top: buttonPos.y + 'px' }"
      @click="explainSelection"
    >
      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
        />
      </svg>
      AI 解释
    </button>
  </Teleport>

  <!-- AI 解释侧边抽屉 -->
  <Teleport to="body">
    <Transition name="drawer-slide">
      <div
        v-if="explainDrawer.visible"
        class="explain-drawer fixed inset-0 z-50 flex justify-end"
        @click.self="closeExplainDrawer"
      >
        <!-- 遮罩 -->
        <div class="absolute inset-0 bg-black/30 dark:bg-black/50" @click="closeExplainDrawer"></div>
        <!-- 抽屉面板 -->
        <div class="relative w-full max-w-md h-full bg-white dark:bg-gray-800 shadow-2xl overflow-y-auto">
          <!-- 头部 -->
          <div
            class="sticky top-0 z-10 flex items-center justify-between px-5 py-4 bg-white/95 dark:bg-gray-800/95 backdrop-blur border-b border-gray-100 dark:border-gray-700"
          >
            <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
              <svg class="w-5 h-5 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                />
              </svg>
              AI 概念解释
            </h3>
            <button
              class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              @click="closeExplainDrawer"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- 加载中 -->
          <div v-if="explainDrawer.loading" class="flex flex-col items-center justify-center py-20 px-5">
            <div
              class="w-10 h-10 border-3 border-indigo-200 border-t-indigo-600 rounded-full animate-spin dark:border-indigo-800 dark:border-t-indigo-400"
            ></div>
            <p class="mt-4 text-sm text-gray-500 dark:text-gray-400">AI 正在分析概念...</p>
          </div>

          <!-- 错误 -->
          <div v-else-if="explainDrawer.error" class="px-5 py-10 text-center">
            <div
              class="w-12 h-12 mx-auto mb-3 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center"
            >
              <svg class="w-6 h-6 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <p class="text-sm text-red-600 dark:text-red-400">{{ explainDrawer.error }}</p>
            <button
              class="mt-4 px-4 py-2 text-sm bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg transition"
              @click="requestExplain(selectedText)"
            >
              重试
            </button>
          </div>

          <!-- 解释结果 -->
          <div v-else-if="explainDrawer.result" class="px-5 py-5 space-y-5">
            <!-- 概念标题 -->
            <div class="flex items-center gap-2">
              <span
                class="px-2.5 py-1 bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300 text-xs font-medium rounded-full"
              >
                概念
              </span>
              <h4 class="text-base font-semibold text-gray-900 dark:text-gray-100">
                {{ explainDrawer.result.concept }}
              </h4>
              <span
                v-if="explainDrawer.result.cached"
                class="ml-auto text-[10px] text-gray-400 dark:text-gray-500 bg-gray-100 dark:bg-gray-700 px-1.5 py-0.5 rounded"
              >
                缓存
              </span>
            </div>

            <!-- 通俗解释 -->
            <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-700/50 border border-gray-100 dark:border-gray-600">
              <h5 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                通俗解释
              </h5>
              <p class="text-sm text-gray-800 dark:text-gray-200 leading-relaxed whitespace-pre-wrap">
                {{ explainDrawer.result.explanation }}
              </p>
            </div>

            <!-- 生活类比 -->
            <div
              v-if="explainDrawer.result.analogy"
              class="p-4 rounded-xl bg-amber-50 dark:bg-amber-900/10 border border-amber-100 dark:border-amber-800/30"
            >
              <h5
                class="text-xs font-semibold text-amber-600 dark:text-amber-400 uppercase tracking-wider mb-2 flex items-center gap-1.5"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z"
                  />
                </svg>
                生活类比
              </h5>
              <p class="text-sm text-gray-800 dark:text-gray-200 leading-relaxed whitespace-pre-wrap">
                {{ explainDrawer.result.analogy }}
              </p>
            </div>

            <!-- 具体例子 -->
            <div
              v-if="explainDrawer.result.example"
              class="p-4 rounded-xl bg-emerald-50 dark:bg-emerald-900/10 border border-emerald-100 dark:border-emerald-800/30"
            >
              <h5
                class="text-xs font-semibold text-emerald-600 dark:text-emerald-400 uppercase tracking-wider mb-2 flex items-center gap-1.5"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"
                  />
                </svg>
                具体例子
              </h5>
              <p class="text-sm text-gray-800 dark:text-gray-200 leading-relaxed whitespace-pre-wrap">
                {{ explainDrawer.result.example }}
              </p>
            </div>

            <!-- 关联知识点 -->
            <div
              v-if="explainDrawer.result.related_concepts?.length"
              class="p-4 rounded-xl bg-sky-50 dark:bg-sky-900/10 border border-sky-100 dark:border-sky-800/30"
            >
              <h5
                class="text-xs font-semibold text-sky-600 dark:text-sky-400 uppercase tracking-wider mb-2 flex items-center gap-1.5"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
                  />
                </svg>
                关联知识点
              </h5>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="concept in explainDrawer.result.related_concepts"
                  :key="concept"
                  class="px-2.5 py-1 bg-sky-100 dark:bg-sky-800/30 text-sky-700 dark:text-sky-300 text-xs rounded-full font-medium cursor-pointer hover:bg-sky-200 dark:hover:bg-sky-700/40 transition"
                  @click="requestExplain(concept)"
                >
                  {{ concept }}
                </span>
              </div>
            </div>

            <!-- 操作按钮 -->
            <div class="flex gap-2 pt-2">
              <button
                class="flex-1 px-3 py-2 text-xs font-medium bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition"
                @click="requestExplain(explainDrawer.result.concept)"
              >
                重新生成
              </button>
              <button
                class="flex-1 px-3 py-2 text-xs font-medium bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition"
                @click="closeExplainDrawer"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import katex from 'katex'
import 'katex/dist/katex.min.css'
import { listMaterials, chatStream, listConversations, getConversation, deleteConversation } from '../api/client'
import { useToast, useConfirm } from '../composables/useToast'
import { useExplain } from '../composables/useExplain'
import { useVirtualScroll } from '../composables/useVirtualScroll'
import { useNetworkStatus } from '../composables/useNetworkStatus'

// 语言标签映射
const langLabels = {
  javascript: 'JavaScript',
  js: 'JavaScript',
  jsx: 'JSX',
  typescript: 'TypeScript',
  ts: 'TypeScript',
  tsx: 'TSX',
  python: 'Python',
  py: 'Python',
  go: 'Go',
  golang: 'Go',
  java: 'Java',
  cpp: 'C++',
  'c++': 'C++',
  c: 'C',
  'c#': 'C#',
  csharp: 'C#',
  rust: 'Rust',
  rs: 'Rust',
  ruby: 'Ruby',
  rb: 'Ruby',
  php: 'PHP',
  swift: 'Swift',
  kotlin: 'Kotlin',
  kt: 'Kotlin',
  html: 'HTML',
  xml: 'XML',
  css: 'CSS',
  scss: 'SCSS',
  sass: 'Sass',
  less: 'Less',
  sql: 'SQL',
  mysql: 'MySQL',
  postgresql: 'PostgreSQL',
  bash: 'Bash',
  sh: 'Shell',
  shell: 'Shell',
  zsh: 'Zsh',
  json: 'JSON',
  yaml: 'YAML',
  yml: 'YAML',
  toml: 'TOML',
  markdown: 'Markdown',
  md: 'Markdown',
  dockerfile: 'Dockerfile',
  docker: 'Dockerfile',
  lua: 'Lua',
  perl: 'Perl',
  r: 'R',
  matlab: 'MATLAB',
  scala: 'Scala',
  haskell: 'Haskell',
  elixir: 'Elixir',
  vue: 'Vue',
  svelte: 'Svelte'
}

// 自定义 marked 渲染器 — 代码块增加复制按钮+语言标签
const markedRenderer = new marked.Renderer()
markedRenderer.code = function (tokenOrCode, langArg) {
  // marked v18+: code() receives { text, lang, escaped } token object
  // fallback for older API: positional (code, lang)
  const code = typeof tokenOrCode === 'object' ? tokenOrCode.text : tokenOrCode
  const lang = typeof tokenOrCode === 'object' ? tokenOrCode.lang : langArg
  // code() 仅处理围栏代码块（```code```），行内代码由 codespan() 处理
  const langStr = (lang || '').trim().split(/\s+/)[0].toLowerCase()
  const displayLang = langLabels[langStr] || (langStr ? langStr.charAt(0).toUpperCase() + langStr.slice(1) : '')
  const langHtml = displayLang ? `<span class="code-lang">${displayLang}</span>` : ''
  const encoded = encodeURIComponent(code)
  return `<div class="code-block-wrapper">${langHtml}<button class="code-copy-btn" data-copy-code="${encoded}" title="复制代码"><svg class="code-copy-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><rect x="9" y="9" width="13" height="13" rx="2" ry="2" stroke-width="2"></rect><path stroke-width="2" d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"></path></svg><span>复制</span></button><pre><code class="${langStr ? 'language-' + langStr : ''}">${code}</code></pre></div>`
}

// 配置 marked
marked.setOptions({
  breaks: true,
  gfm: true,
  renderer: markedRenderer
})

// 渲染 LaTeX 数学公式（在 Markdown 解析前处理，避免 KaTeX HTML 被 marked 转义）
function renderLatex(text) {
  if (!text) return text
  // 块级 $$...$$ 和 \[...\]
  let result = text.replace(/\$\$([\s\S]*?)\$\$/g, (_, f) => {
    try {
      return katex.renderToString(f.trim(), { displayMode: true, throwOnError: false })
    } catch {
      return f
    }
  })
  result = result.replace(/\\\[([\s\S]*?)\\\]/g, (_, f) => {
    try {
      return katex.renderToString(f.trim(), { displayMode: true, throwOnError: false })
    } catch {
      return f
    }
  })
  // 行内 $...$ 和 \(...\)
  result = result.replace(/\$([^$\n]+?)\$/g, (_, f) => {
    try {
      return katex.renderToString(f.trim(), { displayMode: false, throwOnError: false })
    } catch {
      return f
    }
  })
  result = result.replace(/\\\((.+?)\\\)/g, (_, f) => {
    try {
      return katex.renderToString(f.trim(), { displayMode: false, throwOnError: false })
    } catch {
      return f
    }
  })
  return result
}

function renderMarkdown(text) {
  if (!text) return ''
  // 1. 先渲染 LaTeX（KaTeX 生成带 class 的 span/math 标签）
  const withMath = renderLatex(text)
  // 2. 解析 Markdown
  const html = marked.parse(withMath)
  // 3. DOMPurify 清除 XSS，保留 KaTeX 输出的标签和代码块复制按钮属性
  const purified = DOMPurify.sanitize(html, {
    allowlist: {
      h1: true,
      h2: true,
      h3: true,
      h4: true,
      h5: true,
      h6: true,
      p: true,
      br: true,
      strong: true,
      em: true,
      u: true,
      s: true,
      del: true,
      ul: true,
      ol: true,
      li: true,
      a: true,
      code: true,
      pre: true,
      blockquote: true,
      table: true,
      thead: true,
      tbody: true,
      tr: true,
      th: true,
      td: true,
      hr: true,
      span: true,
      div: true,
      sup: true,
      sub: true,
      img: true,
      button: true,
      svg: true,
      rect: true,
      path: true,
      line: true,
      circle: true,
      g: true,
      math: true,
      semantics: true,
      annotation: true,
      mrow: true,
      mi: true,
      mo: true,
      mn: true,
      msup: true,
      msub: true,
      mfrac: true,
      msqrt: true,
      mover: true,
      munder: true,
      mtable: true,
      mtr: true,
      mtd: true,
      mtext: true,
      mspace: true,
      menclose: true,
      mpadded: true,
      mphantom: true,
      mglyph: true
    },
    ALLOW_DATA_ATTR: false,
    ADD_ATTR: [
      'href',
      'target',
      'rel',
      'src',
      'alt',
      'title',
      'class',
      'style',
      'xmlns',
      'encoding',
      'mathvariant',
      'displaystyle',
      'scriptlevel',
      'aria-hidden',
      'width',
      'height',
      'viewBox',
      'd',
      'fill',
      'stroke',
      'stroke-width',
      'transform',
      'x',
      'y',
      'dx',
      'dy',
      'cx',
      'cy',
      'r',
      'data-copy-code'
    ]
  })
  // 4. 后处理：为代码块添加复制按钮（DOMPurify 会移除 onclick，改用事件委托）
  return addCodeCopyButtons(purified)
}

// 代码块复制按钮后处理（DOMPurify 移除 onclick 后的补偿）
function addCodeCopyButtons(html) {
  return html
}

// 代码块复制事件委托（绑定在消息容器上，避免内联事件被 DOMPurify 移除）
function handleCopyCode(e) {
  const btn = e.target.closest('[data-copy-code]')
  if (!btn) return
  e.preventDefault()
  const code = decodeURIComponent(btn.getAttribute('data-copy-code') || '')
  navigator.clipboard
    .writeText(code)
    .then(() => {
      const spanEl = btn.querySelector('span')
      const svgEl = btn.querySelector('svg')
      if (spanEl) spanEl.textContent = '已复制'
      btn.classList.add('copied')
      if (svgEl)
        svgEl.innerHTML = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>'
      setTimeout(() => {
        if (spanEl) spanEl.textContent = '复制'
        btn.classList.remove('copied')
        if (svgEl)
          svgEl.innerHTML =
            '<rect x="9" y="9" width="13" height="13" rx="2" ry="2" stroke-width="2"></rect><path stroke-width="2" d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"></path>'
      }, 2000)
    })
    .catch(() => {
      // fallback: 创建 textarea 复制
      const ta = document.createElement('textarea')
      ta.value = code
      ta.style.position = 'fixed'
      ta.style.opacity = '0'
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
      const spanEl = btn.querySelector('span')
      if (spanEl) {
        spanEl.textContent = '已复制'
        setTimeout(() => {
          spanEl.textContent = '复制'
        }, 2000)
      }
    })
}

const toast = useToast()
const { confirm } = useConfirm()
const { isOnline } = useNetworkStatus()

// AI 概念解释器
const {
  selectedText,
  buttonPos,
  drawer: explainDrawer,
  hideButton: _hideButton,
  initExplainListener,
  cleanupExplainListener,
  requestExplain,
  explainSelection,
  closeDrawer: closeExplainDrawer
} = useExplain()

const messages = ref([])
const inputMessage = ref('')
const isStreaming = ref(false)
const isThinking = ref(false)
const selectedMaterialId = ref('')
const materials = ref([])
const messagesContainer = ref(null)
const inputRef = ref(null)
let currentController = null

// 虚拟滚动：消息列表 >100 条时启用
const {
  startIndex: msgStartIdx,
  endIndex: msgEndIdx,
  topSpacerHeight: msgTopSpacer,
  bottomSpacerHeight: msgBottomSpacer,
  shouldVirtualize: msgVirtualized
} = useVirtualScroll(
  messagesContainer,
  computed(() => messages.value.length),
  {
    itemHeight: 80,
    buffer: 10,
    threshold: 100
  }
)

const visibleMessages = computed(() => {
  if (!msgVirtualized.value) return messages.value
  return messages.value.slice(msgStartIdx.value, msgEndIdx.value)
})

// 对话会话相关
const conversations = ref([])
const currentConvId = ref('')

const quickQuestions = [
  '帮我解释一下 TCP 三次握手',
  '什么是深度学习？',
  '帮我出一道关于二叉树的题',
  '制定一个期末复习计划'
]

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const toolLabels = {
  search_materials: '搜索学习材料',
  create_quiz: '生成练习题',
  generate_card: '生成知识卡片',
  get_user_stats: '查看学习统计',
  recommend_study_plan: '制定复习计划'
}

function getToolLabel(name) {
  return toolLabels[name] || name
}

function formatToolArgs(name, argsStr) {
  try {
    const args = JSON.parse(argsStr)
    if (args.query) return `「${args.query}」`
    if (args.topic) return `「${args.topic}」${args.difficulty ? ' · ' + args.difficulty : ''}`
    if (args.concept) return `「${args.concept}」`
    return ''
  } catch {
    return ''
  }
}

async function loadMaterials() {
  try {
    const res = await listMaterials()
    materials.value = (res.data.data || []).filter((m) => m.status === 'completed' || m.status === 'partial')
  } catch (e) {
    console.error('加载材料列表失败:', e)
  }
}

async function loadConversations() {
  try {
    const res = await listConversations()
    conversations.value = res.data.data || []
  } catch (e) {
    console.error('加载对话列表失败:', e)
  }
}

async function switchConversation(convId) {
  if (convId === currentConvId.value) return
  if (isStreaming.value) return

  try {
    const res = await getConversation(convId)
    const data = res.data
    currentConvId.value = convId
    messages.value = (data.messages || []).map((m) => ({
      role: m.role,
      content: m.content,
      toolEvents: []
    }))
    scrollToBottom()
  } catch (e) {
    toast.error('加载对话失败')
  }
}

function startNewConversation() {
  if (isStreaming.value) return
  currentConvId.value = ''
  messages.value = []
  inputRef.value?.focus()
}

async function confirmDeleteConversation(convId) {
  const convTitle = conversations.value.find((c) => c.id === convId)?.title || '此对话'
  const ok = await confirm(`确定要删除「${convTitle}」吗？删除后不可恢复。`, '删除对话')
  if (!ok) return

  try {
    await deleteConversation(convId)
    // 如果删除的是当前对话，清空消息区域
    if (convId === currentConvId.value) {
      currentConvId.value = ''
      messages.value = []
    }
    // 刷新列表
    await loadConversations()
    toast.success('对话已删除')
  } catch (e) {
    toast.error('删除失败')
  }
}

function sendQuickQuestion(q) {
  inputMessage.value = q
  sendMessage()
}

function sendMessage() {
  const message = inputMessage.value.trim()
  if (!message || isStreaming.value) return

  // 离线时禁止发送
  if (!isOnline.value) {
    toast.warning('AI 对话需要网络连接，请检查网络后重试')
    return
  }

  // /explain 命令：AI 概念解释
  if (message.startsWith('/explain ')) {
    const concept = message.slice(9).trim()
    if (!concept) {
      toast.warning('请输入要解释的概念，如：/explain TCP三次握手')
      return
    }
    inputMessage.value = ''
    messages.value.push({ role: 'user', content: `请解释：${concept}` })
    isThinking.value = true
    const aiMsgIndex = messages.value.length
    messages.value.push({ role: 'assistant', content: '', streaming: true, toolEvents: [] })
    scrollToBottom()

    requestExplain(concept).then(() => {
      isThinking.value = false
      if (explainDrawer.error) {
        messages.value[aiMsgIndex].content = `解释生成失败：${explainDrawer.error}`
      } else if (explainDrawer.result) {
        const r = explainDrawer.result
        let md = `## ${r.concept}\n\n`
        md += `**通俗解释**\n${r.explanation}\n\n`
        if (r.analogy) md += `**生活类比**\n${r.analogy}\n\n`
        if (r.example) md += `**具体例子**\n${r.example}\n\n`
        if (r.related_concepts?.length) md += `**关联知识点**\n${r.related_concepts.map((c) => `- ${c}`).join('\n')}\n`
        if (r.cached) md += `\n> 来自缓存`
        messages.value[aiMsgIndex].content = md
      }
      messages.value[aiMsgIndex].streaming = false
      scrollToBottom()
    })
    return
  }

  // 添加用户消息
  messages.value.push({ role: 'user', content: message })
  inputMessage.value = ''
  scrollToBottom()

  // 添加 AI 回复占位（流式填充）
  isThinking.value = true
  const aiMsgIndex = messages.value.length
  messages.value.push({ role: 'assistant', content: '', streaming: true, toolEvents: [] })
  scrollToBottom()

  // 启动 SSE 流式请求
  isStreaming.value = true

  currentController = chatStream(
    message,
    selectedMaterialId.value || '',
    // onToken: 每收到一个 token
    (token) => {
      isThinking.value = false // 收到第一个 token 时关闭思考指示器
      messages.value[aiMsgIndex].content += token
      scrollToBottom()
    },
    // onDone: 流结束
    (_fullText) => {
      isThinking.value = false
      messages.value[aiMsgIndex].streaming = false
      isStreaming.value = false
      currentController = null
      scrollToBottom()
      inputRef.value?.focus()
      // 刷新对话列表（更新时间和消息数）
      loadConversations()
    },
    // onError: 错误
    (err) => {
      isThinking.value = false
      messages.value[aiMsgIndex].content = `请求失败: ${err.message}`
      messages.value[aiMsgIndex].streaming = false
      isStreaming.value = false
      currentController = null
    },
    // onToolEvent: 工具调用事件（Function Calling）
    (evt) => {
      isThinking.value = false
      if (evt.type === 'tool_call') {
        messages.value[aiMsgIndex].toolEvents.push({
          name: evt.name,
          done: false,
          argsStr: formatToolArgs(evt.name, evt.args)
        })
      } else if (evt.type === 'tool_result') {
        const lastEvt = messages.value[aiMsgIndex].toolEvents.findLast((e) => e.name === evt.name && !e.done)
        if (lastEvt) lastEvt.done = true
      }
      scrollToBottom()
    },
    // conversationId: 当前对话 ID
    currentConvId.value || '',
    // onConvId: 收到后端返回的 conversation_id
    (convId) => {
      if (!currentConvId.value) {
        currentConvId.value = convId
        // 刷新对话列表，让新对话出现在侧栏
        loadConversations()
      }
    }
  )
}

function stopStreaming() {
  if (currentController) {
    currentController.abort()
    currentController = null
  }
  // 标记当前消息完成
  const lastMsg = messages.value[messages.value.length - 1]
  if (lastMsg && lastMsg.streaming) {
    lastMsg.streaming = false
    lastMsg.content += '\n\n[已停止生成]'
  }
  isStreaming.value = false
}

async function clearConversation() {
  if (currentConvId.value) {
    // 有当前对话 → 删除后新建
    await confirmDeleteConversation(currentConvId.value)
  } else {
    // 无当前对话 → 仅清空消息
    messages.value = []
  }
}

// ==================== 全局键盘快捷键 ====================

function handleGlobalKeydown(e) {
  const tag = e.target?.tagName?.toLowerCase()
  const isEditable = tag === 'input' || tag === 'textarea' || tag === 'select' || e.target?.isContentEditable

  // '/' 聚焦聊天输入框（不在其他输入框中时）
  if (e.key === '/' && !isEditable) {
    e.preventDefault()
    inputRef.value?.focus()
    return
  }

  // Ctrl+Enter 发送消息
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
    e.preventDefault()
    sendMessage()
    return
  }
}

onMounted(() => {
  loadMaterials()
  loadConversations()
  inputRef.value?.focus()
  scrollToBottom()
  document.addEventListener('keydown', handleGlobalKeydown)
  initExplainListener()
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleGlobalKeydown)
  cleanupExplainListener()
})
</script>

<style scoped>
/* KaTeX 暗色模式适配 */
:deep(.katex) {
  color: inherit;
}

/* ========== 代码块增强样式 ========== */
:deep(.code-block-wrapper) {
  position: relative;
  margin: 0.75rem 0;
  border-radius: 0.5rem;
  overflow: hidden;
  border: 1px solid #e5e7eb;
}

:deep(.code-block-wrapper .code-lang) {
  position: absolute;
  top: 0;
  left: 0.75rem;
  font-size: 0.7rem;
  color: #9ca3af;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  pointer-events: none;
  line-height: 2.25rem;
  z-index: 1;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

:deep(.code-block-wrapper .code-copy-btn) {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  font-size: 0.7rem;
  color: #9ca3af;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 0.375rem;
  cursor: pointer;
  opacity: 0;
  transition:
    opacity 0.2s ease,
    color 0.2s ease,
    background 0.2s ease,
    border-color 0.2s ease;
  z-index: 2;
  line-height: 1;
  font-family: inherit;
}

:deep(.code-block-wrapper:hover .code-copy-btn) {
  opacity: 1;
}

:deep(.code-block-wrapper .code-copy-btn:hover) {
  color: #e5e7eb;
  background: rgba(255, 255, 255, 0.15);
  border-color: rgba(255, 255, 255, 0.3);
}

:deep(.code-block-wrapper .code-copy-btn.copied) {
  color: #34d399;
  background: rgba(52, 211, 153, 0.12);
  border-color: rgba(52, 211, 153, 0.3);
  opacity: 1;
}

:deep(.code-block-wrapper .code-copy-icon) {
  width: 0.875rem;
  height: 0.875rem;
  flex-shrink: 0;
}

/* 代码块 pre 覆盖 prose 默认样式 */
:deep(.code-block-wrapper pre) {
  margin: 0;
  padding: 2rem 1rem 1rem;
  background: #1e293b;
  color: #e2e8f0;
  font-size: 0.8rem;
  line-height: 1.65;
  overflow-x: auto;
  border-radius: 0;
}

:deep(.code-block-wrapper pre code) {
  color: inherit;
  background: transparent;
  padding: 0;
  font-size: inherit;
  font-weight: normal;
}

/* 行内 code 保持原样 */
:deep(code:not(pre code)) {
  background: #f1f5f9;
  color: #c026d3;
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  font-size: 0.85em;
}

/* ===== 暗色模式 ===== */
:deep(.dark .code-block-wrapper) {
  border-color: #374151;
}

:deep(.dark .code-block-wrapper .code-lang) {
  color: #6b7280;
}

:deep(.dark .code-block-wrapper .code-copy-btn) {
  color: #6b7280;
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.1);
}

:deep(.dark .code-block-wrapper:hover .code-copy-btn) {
  opacity: 1;
}

:deep(.dark .code-block-wrapper .code-copy-btn:hover) {
  color: #d1d5db;
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.2);
}

:deep(.dark .code-block-wrapper .code-copy-btn.copied) {
  color: #34d399;
  background: rgba(52, 211, 153, 0.1);
  border-color: rgba(52, 211, 153, 0.25);
}

:deep(.dark .code-block-wrapper pre) {
  background: #0f172a;
  color: #cbd5e1;
}

:deep(.dark code:not(pre code)) {
  background: #374151;
  color: #e879f9;
}

/* ===== 自定义滚动条 ===== */
:deep(.code-block-wrapper pre::-webkit-scrollbar) {
  height: 4px;
}

:deep(.code-block-wrapper pre::-webkit-scrollbar-track) {
  background: transparent;
}

:deep(.code-block-wrapper pre::-webkit-scrollbar-thumb) {
  background: #475569;
  border-radius: 2px;
}

:deep(.code-block-wrapper pre::-webkit-scrollbar-thumb:hover) {
  background: #64748b;
}

/* ===== AI 解释浮动按钮+抽屉 ===== */
.explain-float-btn {
  animation: explain-pop 0.15s ease-out;
}

@keyframes explain-pop {
  from {
    opacity: 0;
    transform: scale(0.85);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.drawer-slide-enter-active,
.drawer-slide-leave-active {
  transition: opacity 0.2s ease;
}
.drawer-slide-enter-active > div:last-child,
.drawer-slide-leave-active > div:last-child {
  transition: transform 0.25s ease;
}
.drawer-slide-enter-from,
.drawer-slide-leave-to {
  opacity: 0;
}
.drawer-slide-enter-from > div:last-child,
.drawer-slide-leave-to > div:last-child {
  transform: translateX(100%);
}
</style>
