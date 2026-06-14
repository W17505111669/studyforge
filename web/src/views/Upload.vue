<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-4xl mx-auto">
    <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">上传学习材料</h1>
    <p class="text-gray-500 dark:text-gray-400 mb-8">上传材料后，4 个 AI Agent 将并行分析，实时生成知识卡片、练习题和知识图谱</p>

    <!-- 上传表单 -->
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6 mb-6 sm:mb-8">
      <form @submit.prevent="handleUpload" class="space-y-4">

        <!-- 文件上传区域 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">上传文件</label>
          <div
            class="border-2 border-dashed rounded-xl p-8 text-center transition-all cursor-pointer"
            :class="dragOver ? 'border-primary-500 bg-primary-50' : 'border-gray-200 dark:border-gray-600 hover:border-primary-300 hover:bg-gray-50 dark:hover:bg-gray-700'"
            @dragover.prevent="dragOver = true"
            @dragleave.prevent="dragOver = false"
            @drop.prevent="handleDrop"
            @click="triggerFileInput"
          >
            <input
              ref="fileInput"
              type="file"
              accept=".pdf,.docx,.md,.txt,.markdown"
              class="hidden"
              @change="handleFileSelect"
            />
            <div v-if="!fileInfo" class="space-y-2">
              <p class="text-4xl">📄</p>
              <p class="text-sm text-gray-600 dark:text-gray-400 font-medium">拖拽文件到这里，或点击选择文件</p>
              <p class="text-xs text-gray-400 dark:text-gray-500">支持 PDF、DOCX、MD、TXT（最大 20MB）</p>
            </div>
            <div v-else class="flex items-center justify-center gap-3">
              <!-- PDF 文件图标 -->
              <svg v-if="fileType === 'pdf'" class="w-10 h-10 shrink-0" viewBox="0 0 40 40" fill="none">
                <rect x="4" y="2" width="32" height="36" rx="4" fill="#FEE2E2" stroke="#EF4444" stroke-width="1.5"/>
                <rect x="10" y="16" width="20" height="12" rx="2" fill="#EF4444"/>
                <text x="20" y="25" text-anchor="middle" fill="white" font-size="7" font-weight="bold">PDF</text>
                <line x1="10" y1="8" x2="24" y2="8" stroke="#FCA5A5" stroke-width="1.5" stroke-linecap="round"/>
                <line x1="10" y1="11" x2="20" y2="11" stroke="#FCA5A5" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
              <!-- Word 文件图标 -->
              <svg v-else-if="fileType === 'docx' || fileType === 'doc'" class="w-10 h-10 shrink-0" viewBox="0 0 40 40" fill="none">
                <rect x="4" y="2" width="32" height="36" rx="4" fill="#DBEAFE" stroke="#3B82F6" stroke-width="1.5"/>
                <rect x="10" y="16" width="20" height="12" rx="2" fill="#3B82F6"/>
                <text x="20" y="25" text-anchor="middle" fill="white" font-size="6.5" font-weight="bold">DOCX</text>
                <line x1="10" y1="8" x2="24" y2="8" stroke="#93C5FD" stroke-width="1.5" stroke-linecap="round"/>
                <line x1="10" y1="11" x2="20" y2="11" stroke="#93C5FD" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
              <!-- Markdown 文件图标 -->
              <svg v-else-if="fileType === 'md' || fileType === 'markdown'" class="w-10 h-10 shrink-0" viewBox="0 0 40 40" fill="none">
                <rect x="4" y="2" width="32" height="36" rx="4" fill="#D1FAE5" stroke="#10B981" stroke-width="1.5"/>
                <rect x="8" y="14" width="24" height="14" rx="2" fill="#10B981"/>
                <text x="20" y="24" text-anchor="middle" fill="white" font-size="6" font-weight="bold">MD</text>
                <path d="M12 8h6M12 11h10" stroke="#6EE7B7" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
              <!-- 通用文本文件图标 -->
              <svg v-else class="w-10 h-10 shrink-0" viewBox="0 0 40 40" fill="none">
                <rect x="4" y="2" width="32" height="36" rx="4" fill="#F3F4F6" stroke="#9CA3AF" stroke-width="1.5"/>
                <line x1="10" y1="10" x2="30" y2="10" stroke="#D1D5DB" stroke-width="1.5" stroke-linecap="round"/>
                <line x1="10" y1="14" x2="26" y2="14" stroke="#D1D5DB" stroke-width="1.5" stroke-linecap="round"/>
                <line x1="10" y1="18" x2="28" y2="18" stroke="#D1D5DB" stroke-width="1.5" stroke-linecap="round"/>
                <line x1="10" y1="22" x2="22" y2="22" stroke="#D1D5DB" stroke-width="1.5" stroke-linecap="round"/>
                <line x1="10" y1="26" x2="30" y2="26" stroke="#D1D5DB" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
              <div class="text-left">
                <p class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ fileInfo.filename }}</p>
                <p class="text-xs text-gray-400 dark:text-gray-500">{{ formatSize(fileInfo.size) }} · 已提取 {{ form.content.length }} 字</p>
              </div>
              <button type="button" @click.stop="clearFile" class="ml-4 text-gray-400 dark:text-gray-500 hover:text-red-500 text-sm">清除</button>
            </div>
          </div>
          <!-- 上传进度条 -->
          <div v-if="fileUploading" class="mt-2">
            <div class="flex items-center justify-between text-xs mb-1">
              <span class="text-blue-600 dark:text-blue-400 font-medium">
                {{ uploadProgress > 0 ? `上传中 ${uploadProgress}%` : '正在提取文件内容...' }}
              </span>
              <span class="text-gray-400 dark:text-gray-500">{{ fileInfo ? '' : '处理中' }}</span>
            </div>
            <div class="h-1.5 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden">
              <div
                class="h-full rounded-full bg-blue-500 dark:bg-blue-400 transition-all duration-300 ease-out"
                :style="{ width: (uploadProgress > 0 ? uploadProgress : 15) + '%' }"
                :class="uploadProgress === 0 ? 'animate-pulse' : ''"
              ></div>
            </div>
          </div>
          <p v-if="fileSizeWarning" class="text-xs text-amber-500 dark:text-amber-400 mt-1 flex items-center gap-1">
            <svg class="w-3.5 h-3.5 shrink-0" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
            </svg>
            {{ fileSizeWarning }}
          </p>
          <p v-if="fileError" class="text-xs text-red-500 mt-1">{{ fileError }}</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">材料标题</label>
          <input
            v-model="form.title"
            type="text"
            required
            class="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 dark:text-gray-100 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-all"
            placeholder="例：计算机网络第三章 - 数据链路层"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">材料内容</label>
          <textarea
            v-model="form.content"
            required
            rows="10"
            class="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 dark:text-gray-100 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-all resize-y"
            placeholder="粘贴你的学习笔记、课件内容、论文段落..."
          ></textarea>
          <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">支持直接粘贴文本，建议 500-10000 字</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">分类标签（可选）</label>
          <div class="relative">
            <input
              v-model="form.tags"
              type="text"
              class="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 dark:text-gray-100 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-all"
              placeholder="例：计算机网络, 期末考试"
              @focus="showTagSuggestions = true"
              @blur="hideSuggestions"
              @input="filterTagSuggestions"
            />
            <!-- 已有标签快捷建议 -->
            <Transition name="tag-suggest">
              <div
                v-if="showTagSuggestions && filteredExistingTags.length > 0"
                class="absolute z-20 top-full left-0 right-0 mt-1 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-600 max-h-40 overflow-y-auto custom-scroll"
              >
                <button
                  v-for="tag in filteredExistingTags"
                  :key="tag.name"
                  type="button"
                  @mousedown.prevent="addTagSuggestion(tag.name)"
                  class="w-full text-left px-3 py-2 text-sm hover:bg-primary-50 dark:hover:bg-primary-900/20 flex items-center justify-between transition-colors"
                >
                  <span class="text-gray-700 dark:text-gray-300">{{ tag.name }}</span>
                  <span class="text-xs text-gray-400 dark:text-gray-500">{{ tag.count }} 次使用</span>
                </button>
              </div>
            </Transition>
          </div>
          <!-- 已输入的标签 pills -->
          <div v-if="parsedFormTags.length > 0" class="flex flex-wrap gap-1.5 mt-2">
            <span
              v-for="(tag, idx) in parsedFormTags"
              :key="idx"
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400"
            >
              {{ tag }}
              <button type="button" @click="removeFormTag(idx)" class="hover:text-red-500 transition-colors">×</button>
            </span>
          </div>
        </div>

        <button
          type="submit"
          :disabled="uploading"
          class="w-full py-3 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
        >
          {{ uploading ? '上传中...' : '上传并开始分析' }}
        </button>
      </form>
    </div>

    <!-- Agent 并发可视化 -->
    <AgentFlow v-if="analyzing || agents.some(a => a.done)" :agents="agents" :analyzing="analyzing" :timeline-start="analyzeStartTime" :agent-finish-times="agentFinishTimes" />
    <div v-if="isPolling" class="mb-4 px-4 py-2 bg-amber-50 dark:bg-gray-700 border border-amber-200 dark:border-gray-600 rounded-lg text-xs text-amber-600 dark:text-gray-300 flex items-center gap-2">
      <span class="inline-block w-2 h-2 bg-amber-400 rounded-full animate-pulse"></span>
      WebSocket 已断线，正在通过轮询同步分析进度...
    </div>

    <!-- 历史材料列表 -->
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 p-4 sm:p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold dark:text-gray-100">已上传材料</h2>

        <!-- 全选 checkbox -->
        <label v-if="materials.length > 0" class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 cursor-pointer select-none">
          <input
            type="checkbox"
            :checked="isAllSelected"
            :indeterminate="isPartialSelected"
            @change="toggleSelectAll"
            class="w-4 h-4 rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500 accent-primary-600"
          />
          <span>全选</span>
          <span v-if="selectedIds.size > 0" class="text-primary-600 dark:text-primary-400 font-medium">（{{ selectedIds.size }}）</span>
        </label>
      </div>

      <!-- 批量操作栏 -->
      <Transition name="batch-bar">
        <div v-if="selectedIds.size > 0" class="mb-4 flex flex-col sm:flex-row items-start sm:items-center gap-3 p-3 rounded-lg bg-primary-50 dark:bg-primary-900/20 border border-primary-200 dark:border-primary-800">
          <span class="text-sm font-medium text-primary-700 dark:text-primary-300">
            已选 {{ selectedIds.size }} 项
          </span>
          <div class="flex items-center gap-2 flex-wrap">
            <button
              @click="handleBatchAnalyze"
              :disabled="batchAnalyzing"
              class="px-3 py-1.5 text-xs font-medium rounded-lg bg-primary-600 hover:bg-primary-700 text-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ batchAnalyzing ? `分析中 ${batchProgress.completed}/${batchProgress.total}` : '批量分析' }}
            </button>
            <button
              @click="handleBatchDelete"
              class="px-3 py-1.5 text-xs font-medium rounded-lg bg-red-500 hover:bg-red-600 text-white transition-colors"
            >
              批量删除
            </button>
            <button
              @click="selectedIds = new Set()"
              class="px-3 py-1.5 text-xs font-medium rounded-lg text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
            >
              取消选择
            </button>
          </div>
          <!-- 批量分析进度条 -->
          <div v-if="batchAnalyzing && batchProgress.total > 0" class="w-full sm:w-auto sm:flex-1">
            <div class="h-2 rounded-full bg-primary-200 dark:bg-primary-800 overflow-hidden">
              <div
                class="h-full rounded-full bg-primary-600 dark:bg-primary-400 transition-all duration-500"
                :style="{ width: Math.round(batchProgress.completed / batchProgress.total * 100) + '%' }"
              ></div>
            </div>
            <p class="text-xs text-primary-600 dark:text-primary-400 mt-1">
              {{ batchProgress.completed }} / {{ batchProgress.total }} 完成
            </p>
          </div>
        </div>
      </Transition>

      <!-- 标签过滤栏 -->
      <div v-if="existingTags.length > 0" class="mb-4 flex flex-wrap items-center gap-2">
        <span class="text-xs text-gray-500 dark:text-gray-400 shrink-0">标签过滤：</span>
        <button
          @click="activeTagFilter = ''"
          class="px-2.5 py-1 rounded-full text-xs font-medium transition-all border"
          :class="activeTagFilter === '' ? 'bg-primary-50 text-primary-700 border-primary-200 dark:bg-primary-900/30 dark:text-primary-400 dark:border-primary-700' : 'bg-gray-50 text-gray-500 border-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-600 hover:border-gray-300'"
        >
          全部
        </button>
        <button
          v-for="tag in existingTags"
          :key="tag.name"
          @click="activeTagFilter = activeTagFilter === tag.name ? '' : tag.name"
          class="px-2.5 py-1 rounded-full text-xs font-medium transition-all border"
          :class="activeTagFilter === tag.name ? 'bg-primary-50 text-primary-700 border-primary-200 dark:bg-primary-900/30 dark:text-primary-400 dark:border-primary-700' : 'bg-gray-50 text-gray-500 border-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-600 hover:border-gray-300'"
        >
          {{ tag.name }}
          <span class="ml-0.5 opacity-60">{{ tag.count }}</span>
        </button>
      </div>

      <div v-if="materials.length === 0" class="text-center py-8 text-gray-400 dark:text-gray-500">
        暂无材料，上传你的第一份学习材料吧
      </div>

      <div v-else class="space-y-3">
        <div v-for="m in materials" :key="m.id"
          class="relative rounded-lg border transition-all"
          :class="isSelected(m.id) ? 'border-primary-400 dark:border-primary-600 bg-primary-50/50 dark:bg-primary-900/10' : 'border-gray-100 dark:border-gray-700 hover:border-primary-200 dark:hover:bg-gray-700'"
          @mouseenter="handleCardMouseenter($event, m)"
          @mouseleave="handleCardMouseleave"
        >
          <!-- 移动端卡片布局 -->
          <div class="md:hidden p-4 pb-3">
            <div class="flex items-start gap-3">
              <input
                type="checkbox"
                :checked="isSelected(m.id)"
                @click.stop="toggleSelection(m.id)"
                class="w-4 h-4 mt-0.5 rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500 accent-primary-600 shrink-0"
              />
              <div class="flex-1 min-w-0">
                <router-link :to="`/materials/${m.id}`" class="font-medium dark:text-gray-100 hover:text-primary-600 transition block truncate">{{ m.title }}</router-link>
                <div v-if="m.tags" class="flex flex-wrap gap-1 mt-1">
                  <span v-for="t in m.tags.split(',')" :key="t" class="px-1.5 py-0.5 rounded text-[10px] bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400">{{ t.trim() }}</span>
                </div>
                <div class="flex items-center justify-between mt-2">
                  <p class="text-xs text-gray-400 dark:text-gray-500">{{ formatDate(m.created_at) }}</p>
                  <span class="px-2 py-1 rounded text-xs font-medium" :class="statusClass(m.status)">
                    {{ statusLabel(m.status) }}
                  </span>
                </div>
              </div>
            </div>
          </div>
          <div class="md:hidden flex items-center gap-3 px-4 py-3 border-t border-gray-50 dark:border-gray-700/50 bg-gray-50/50 dark:bg-gray-800/50 rounded-b-lg pl-11">
            <router-link
              v-if="m.status === 'completed' || m.status === 'partial'"
              :to="`/materials/${m.id}`"
              class="text-xs text-primary-600 hover:text-primary-700 font-medium"
            >
              查看详情
            </router-link>
            <button
              v-if="m.status === 'pending'"
              @click="handleAnalyze(m.id)"
              class="text-xs text-primary-600 hover:text-primary-700 font-medium"
            >
              开始分析
            </button>
            <button
              @click="handleDelete(m.id)"
              class="text-xs text-red-400 hover:text-red-600 ml-auto"
            >
              删除
            </button>
          </div>

          <!-- 桌面端横排布局 -->
          <div class="hidden md:flex items-center gap-3 p-4">
            <input
              type="checkbox"
              :checked="isSelected(m.id)"
              @click.stop="toggleSelection(m.id)"
              class="w-4 h-4 rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500 accent-primary-600 shrink-0"
            />
            <div class="flex-1 min-w-0">
              <router-link :to="`/materials/${m.id}`" class="font-medium dark:text-gray-100 truncate hover:text-primary-600 transition">{{ m.title }}</router-link>
              <div v-if="m.tags" class="flex flex-wrap gap-1 mt-1">
                <span v-for="t in m.tags.split(',')" :key="t" class="px-1.5 py-0.5 rounded text-[10px] bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400">{{ t.trim() }}</span>
              </div>
              <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">{{ formatDate(m.created_at) }}</p>
            </div>
            <div class="flex items-center gap-3">
              <span class="px-2 py-1 rounded text-xs font-medium" :class="statusClass(m.status)">
                {{ statusLabel(m.status) }}
              </span>
              <router-link
                v-if="m.status === 'completed' || m.status === 'partial'"
                :to="`/materials/${m.id}`"
                class="text-xs text-primary-600 hover:text-primary-700 font-medium"
              >
                查看
              </router-link>
              <button
                v-if="m.status === 'pending'"
                @click="handleAnalyze(m.id)"
                class="text-xs text-primary-600 hover:text-primary-700 font-medium"
              >
                开始分析
              </button>
              <button
                @click="handleDelete(m.id)"
                class="text-xs text-red-400 hover:text-red-600"
              >
                删除
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 分析摘要预览浮层 -->
      <Teleport to="body">
        <Transition name="preview">
          <div
            v-if="hoveredMaterial && hoverPreview"
            class="fixed z-[9999] w-96 bg-white dark:bg-gray-800 rounded-xl shadow-xl border border-gray-200 dark:border-gray-600 p-5 pointer-events-none"
            :style="{ top: popoverPos.top + 'px', left: popoverPos.left + 'px' }"
          >
            <div class="flex items-center gap-2 mb-3 pb-2.5 border-b border-gray-100 dark:border-gray-700">
              <span class="w-7 h-7 rounded-lg bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center text-blue-500 dark:text-blue-400 text-xs font-bold">Ai</span>
              <span class="text-sm font-semibold text-gray-800 dark:text-gray-200">分析摘要预览</span>
              <span class="ml-auto text-[10px] text-gray-400 dark:text-gray-500">hover 触发</span>
            </div>

            <div class="max-h-64 overflow-y-auto pr-1 custom-scroll">
              <!-- 摘要文本 -->
              <p v-if="hoverPreview.summary" class="text-sm text-gray-600 dark:text-gray-300 leading-relaxed mb-3 line-clamp-4">
                {{ hoverPreview.summary }}
              </p>
              <p v-else class="text-sm text-gray-400 dark:text-gray-500 italic mb-3">暂无文字摘要</p>

              <!-- 知识点预览 -->
              <div v-if="hoverPreview.keyPoints?.length" class="mb-3">
                <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-2 flex items-center gap-1">
                  <svg class="w-3 h-3 text-amber-500" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M11 3a1 1 0 10-2 0v1a1 1 0 102 0V3zM15.657 5.757a1 1 0 00-1.414-1.414l-.707.707a1 1 0 001.414 1.414l.707-.707zM18 10a1 1 0 01-1 1h-1a1 1 0 110-2h1a1 1 0 011 1zM5.05 6.464A1 1 0 106.464 5.05l-.707-.707a1 1 0 00-1.414 1.414l.707.707zM4 11a1 1 0 100-2H3a1 1 0 000 2h1zM10 18a1 1 0 001-1v-1a1 1 0 10-2 0v1a1 1 0 001 1z" />
                    <path fill-rule="evenodd" d="M10 2a8 8 0 100 16 8 8 0 000-16zm0 14a6 6 0 110-12 6 6 0 010 12z" clip-rule="evenodd" />
                  </svg>
                  知识点（{{ hoverPreview.keyPoints.length }}）
                </h4>
                <div class="space-y-1">
                  <div
                    v-for="(point, idx) in hoverPreview.keyPoints.slice(0, 4)"
                    :key="idx"
                    class="flex items-start gap-1.5 text-xs text-gray-600 dark:text-gray-400"
                  >
                    <span class="w-4 h-4 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-400 text-[10px] flex items-center justify-center font-medium mt-0.5 shrink-0">{{ idx + 1 }}</span>
                    <span class="line-clamp-1">{{ typeof point === 'string' ? point : point.title || point.name || '' }}</span>
                  </div>
                  <p v-if="hoverPreview.keyPoints.length > 4" class="text-[10px] text-gray-400 dark:text-gray-500 ml-5">
                    还有 {{ hoverPreview.keyPoints.length - 4 }} 个知识点...
                  </p>
                </div>
              </div>

              <!-- 关系预览 -->
              <div v-if="hoverPreview.relationships?.length" class="mb-3">
                <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-2 flex items-center gap-1">
                  <svg class="w-3 h-3 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.172 13.828a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.102 1.101" />
                  </svg>
                  概念关系（{{ hoverPreview.relationships.length }}）
                </h4>
                <div class="space-y-1">
                  <div
                    v-for="(rel, idx) in hoverPreview.relationships.slice(0, 3)"
                    :key="idx"
                    class="text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1"
                  >
                    <span class="font-medium text-purple-600 dark:text-purple-400 truncate max-w-[100px]">{{ rel.source || rel.from }}</span>
                    <span class="text-purple-400 shrink-0">→</span>
                    <span class="font-medium text-purple-600 dark:text-purple-400 truncate max-w-[100px]">{{ rel.target || rel.to }}</span>
                  </div>
                  <p v-if="hoverPreview.relationships.length > 3" class="text-[10px] text-gray-400 dark:text-gray-500">
                    还有 {{ hoverPreview.relationships.length - 3 }} 条关系...
                  </p>
                </div>
              </div>

              <!-- 重要性 -->
              <div v-if="hoverPreview.importance" class="p-2 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-100 dark:border-emerald-800">
                <p class="text-xs text-emerald-700 dark:text-emerald-400 font-medium mb-0.5">重要性</p>
                <p class="text-xs text-emerald-600 dark:text-emerald-500 line-clamp-2">{{ hoverPreview.importance }}</p>
              </div>
            </div>

            <div class="mt-3 pt-2.5 border-t border-gray-100 dark:border-gray-700">
              <p class="text-[10px] text-gray-400 dark:text-gray-500 flex items-center gap-1">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                点击"查看"或标题进入详情页查看完整分析
              </p>
            </div>
          </div>
        </Transition>
      </Teleport>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted, onUnmounted } from 'vue'
import {
  uploadMaterial, uploadFile, listMaterials, analyzeMaterial, deleteMaterial, getMaterialStatus,
  batchAnalyzeMaterials, batchDeleteMaterials, getTags
} from '../api/client'
import AgentFlow from '../components/AgentFlow.vue'
import { useToast, useConfirm } from '../composables/useToast'

const toast = useToast()
const { confirm } = useConfirm()

const form = reactive({ title: '', content: '', tags: '' })
const uploading = ref(false)
const analyzing = ref(false)
const materials = ref([])
let ws = null
let wsRetryCount = 0
const MAX_WS_RETRIES = 5
const currentAnalyzingId = ref(null)
let pollTimer = null
const isPolling = ref(false)

// 批量操作
const selectedIds = ref(new Set())
const batchAnalyzing = ref(false)
const batchProgress = ref({ completed: 0, total: 0 })

const isAllSelected = computed(() => {
  return materials.value.length > 0 && selectedIds.value.size === materials.value.length
})

const isPartialSelected = computed(() => {
  return selectedIds.value.size > 0 && selectedIds.value.size < materials.value.length
})

function isSelected(id) {
  return selectedIds.value.has(id)
}

function toggleSelection(id) {
  const newSet = new Set(selectedIds.value)
  if (newSet.has(id)) {
    newSet.delete(id)
  } else {
    newSet.add(id)
  }
  selectedIds.value = newSet
}

function toggleSelectAll() {
  if (isAllSelected.value) {
    selectedIds.value = new Set()
  } else {
    selectedIds.value = new Set(materials.value.map(m => m.id))
  }
}

async function handleBatchAnalyze() {
  const ids = Array.from(selectedIds.value)
  if (ids.length === 0) return

  const ok = await confirm(`确定批量分析 ${ids.length} 个材料？分析将通过后台并发执行。`)
  if (!ok) return

  batchAnalyzing.value = true
  batchProgress.value = { completed: 0, total: ids.length }

  // 连接 WebSocket 接收进度
  connectWebSocket()

  try {
    const res = await batchAnalyzeMaterials(ids)
    toast.success(res.data.message || '批量分析已启动')
    // 更新本地材料状态为 analyzing
    materials.value = materials.value.map(m => {
      if (selectedIds.value.has(m.id) && m.status === 'pending') {
        return { ...m, status: 'analyzing' }
      }
      return m
    })
    batchProgress.value.total = res.data.started || ids.length
  } catch (err) {
    toast.error('批量分析启动失败: ' + (err.response?.data?.error || err.message))
    batchAnalyzing.value = false
  }
}

async function handleBatchDelete() {
  const ids = Array.from(selectedIds.value)
  if (ids.length === 0) return

  const ok = await confirm(`确定批量删除 ${ids.length} 个材料？此操作不可撤销。`)
  if (!ok) return

  try {
    const res = await batchDeleteMaterials(ids)
    const deleted = res.data.deleted || 0
    toast.success(`成功删除 ${deleted} 个材料`)
    selectedIds.value = new Set()
    await loadMaterials()
  } catch (err) {
    toast.error('批量删除失败: ' + (err.response?.data?.error || err.message))
  }
}

// 文件上传相关
const fileInput = ref(null)
const fileInfo = ref(null)
const fileUploading = ref(false)
const fileError = ref('')
const dragOver = ref(false)
const uploadProgress = ref(0) // 0-100 upload progress percentage
const fileSizeWarning = ref('') // warning for files > 10MB

// Hover 预览浮层
const hoveredMaterial = ref(null)
const hoverPreview = ref(null)
const popoverPos = ref({ top: 0, left: 0 })
let hoverTimer = null

// ========== 标签系统 ==========
const existingTags = ref([]) // [{ name, count }]
const activeTagFilter = ref('')
const showTagSuggestions = ref(false)
const tagSuggestionQuery = ref('')

// 解析表单中已输入的标签
const parsedFormTags = computed(() => {
  if (!form.tags) return []
  return form.tags.split(',').map(t => t.trim()).filter(Boolean)
})

// 过滤后的已有标签建议（排除已输入的）
const filteredExistingTags = computed(() => {
  const entered = new Set(parsedFormTags.value.map(t => t.toLowerCase()))
  let tags = existingTags.value.filter(t => !entered.has(t.name.toLowerCase()))
  if (tagSuggestionQuery.value) {
    const q = tagSuggestionQuery.value.toLowerCase()
    tags = tags.filter(t => t.name.toLowerCase().includes(q))
  }
  return tags.slice(0, 8)
})

function filterTagSuggestions() {
  // 取最后一个逗号后的文字作为搜索词
  const parts = form.tags.split(',')
  tagSuggestionQuery.value = (parts[parts.length - 1] || '').trim()
}

function addTagSuggestion(tagName) {
  const parts = parsedFormTags.value
  // 如果已有标签，替换最后一个未完成的；否则追加
  const rawParts = form.tags.split(',')
  const lastPart = rawParts[rawParts.length - 1]?.trim()
  if (lastPart && !parts.includes(lastPart)) {
    rawParts[rawParts.length - 1] = ' ' + tagName
  } else {
    rawParts.push(' ' + tagName)
  }
  form.tags = rawParts.join(',').trim()
  showTagSuggestions.value = false
  tagSuggestionQuery.value = ''
}

function removeFormTag(idx) {
  const parts = parsedFormTags.value
  parts.splice(idx, 1)
  form.tags = parts.join(', ')
}

function hideSuggestions() {
  setTimeout(() => { showTagSuggestions.value = false }, 150)
}

async function loadTags() {
  try {
    const res = await getTags()
    existingTags.value = res.data.tags || []
  } catch (e) {
    console.error('标签加载失败:', e)
  }
}

// 标签过滤变化时重新加载材料
watch(activeTagFilter, () => { loadMaterials() })

const fileType = computed(() => {
  if (!fileInfo.value) return ''
  const ext = fileInfo.value.filename.split('.').pop().toLowerCase()
  return ext
})

function formatSize(bytes) {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function triggerFileInput() {
  fileInput.value?.click()
}

function clearFile() {
  fileInfo.value = null
  fileError.value = ''
  fileSizeWarning.value = ''
  uploadProgress.value = 0
  if (fileInput.value) fileInput.value.value = ''
}

async function handleFileSelect(event) {
  const file = event.target.files?.[0]
  if (file) await processFile(file)
}

async function handleDrop(event) {
  dragOver.value = false
  const file = event.dataTransfer?.files?.[0]
  if (file) await processFile(file)
}

const ALLOWED_EXTENSIONS = ['.pdf', '.docx', '.md', '.txt', '.markdown']
const MAX_FILE_SIZE = 20 * 1024 * 1024 // 20MB

async function processFile(file) {
  fileError.value = ''
  fileSizeWarning.value = ''
  uploadProgress.value = 0

  // 客户端文件类型验证（拖拽可绕过 accept 属性）
  const ext = file.name.substring(file.name.lastIndexOf('.')).toLowerCase()
  if (!ALLOWED_EXTENSIONS.includes(ext)) {
    fileError.value = `不支持的文件类型: ${ext}，请上传 PDF/DOCX/MD/TXT 文件`
    return
  }

  // 客户端文件大小验证（避免上传大文件后才报错）
  if (file.size > MAX_FILE_SIZE) {
    fileError.value = `文件大小 ${(file.size / 1024 / 1024).toFixed(1)}MB 超过 20MB 限制`
    return
  }

  // >10MB 文件大小警告
  if (file.size > 10 * 1024 * 1024) {
    fileSizeWarning.value = `文件较大（${(file.size / 1024 / 1024).toFixed(1)}MB），上传可能需要较长时间`
  }

  fileUploading.value = true

  try {
    const res = await uploadFile(file, (percent) => {
      uploadProgress.value = percent
    })
    fileInfo.value = { filename: res.data.filename, size: res.data.size }
    form.content = res.data.text
    form.title = res.data.title || form.title
    fileSizeWarning.value = '' // 上传成功后清除警告
  } catch (err) {
    fileError.value = err.response?.data?.error || '文件处理失败'
    fileInfo.value = null
  } finally {
    fileUploading.value = false
    uploadProgress.value = 0
  }
}

const agents = ref([
  { name: 'Analyst 分析师', icon: '🔍', output: '', done: false, error: false, qualityScore: 0, judgeComment: '', duration: 0, finishedAt: 0 },
  { name: 'QuizMaster 出题官', icon: '✏️', output: '', done: false, error: false, qualityScore: 0, judgeComment: '', duration: 0, finishedAt: 0 },
  { name: 'CardMaker 卡片师', icon: '🃏', output: '', done: false, error: false, qualityScore: 0, judgeComment: '', duration: 0, finishedAt: 0 },
  { name: 'MapBuilder 图谱师', icon: '🗺️', output: '', done: false, error: false, qualityScore: 0, judgeComment: '', duration: 0, finishedAt: 0 },
])

const analyzeStartTime = ref(0)

// Computed: map agent key → absolute finish timestamp (ms)
const agentFinishTimes = computed(() => {
  const keyMap = { Analyst: 0, QuizMaster: 1, CardMaker: 2, MapBuilder: 3 }
  const result = {}
  for (const [key, idx] of Object.entries(keyMap)) {
    if (agents.value[idx].finishedAt > 0) {
      result[key] = agents.value[idx].finishedAt
    }
  }
  return result
})

function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function statusClass(s) {
  return {
    pending: 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400',
    analyzing: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
    completed: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
    partial: 'bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400',
    failed: 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400'
  }[s] || 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
}

function statusLabel(s) {
  return { pending: '待分析', analyzing: '分析中', completed: '已完成', partial: '部分完成', failed: '失败' }[s] || s
}

// ========== Hover 预览浮层 ==========

function handleCardMouseenter(event, m) {
  if (!m.analysis_data || (m.status !== 'completed' && m.status !== 'partial')) return

  if (hoverTimer) clearTimeout(hoverTimer)

  hoverTimer = setTimeout(() => {
    const parsed = parseAnalysis(m.analysis_data)
    if (!parsed) return
    hoveredMaterial.value = m
    hoverPreview.value = parsed

    // 计算浮层位置（卡片右侧，视口自适应）
    const rect = event.currentTarget.getBoundingClientRect()
    let top = rect.top
    let left = rect.right + 12

    // 右侧空间不足时放到左侧
    if (left + 384 > window.innerWidth) {
      left = rect.left - 396
    }
    // 底部溢出时上移
    if (top + 420 > window.innerHeight) {
      top = Math.max(8, window.innerHeight - 420)
    }

    popoverPos.value = { top, left }
  }, 250)
}

function handleCardMouseleave() {
  if (hoverTimer) {
    clearTimeout(hoverTimer)
    hoverTimer = null
  }
  hoveredMaterial.value = null
  hoverPreview.value = null
}

function parseAnalysis(data) {
  if (!data) return null
  try {
    const obj = JSON.parse(data)
    return {
      summary: obj.summary || '',
      keyPoints: Array.isArray(obj.key_points) ? obj.key_points : [],
      relationships: Array.isArray(obj.relationships) ? obj.relationships : [],
      importance: obj.importance || ''
    }
  } catch {
    return null
  }
}

async function handleUpload() {
  uploading.value = true
  try {
    const res = await uploadMaterial({
      title: form.title,
      content: form.content,
      content_type: 'text',
      tags: form.tags || '',
    })
    form.title = ''
    form.content = ''
    form.tags = ''
    await Promise.all([loadMaterials(), loadTags()])

    toast.success('上传成功，正在分析...')

    // 自动触发分析
    await handleAnalyze(res.data.id)
  } catch (err) {
    toast.error('上传失败: ' + (err.response?.data?.error || err.message))
  } finally {
    uploading.value = false
  }
}

async function handleAnalyze(id) {
  analyzing.value = true
  currentAnalyzingId.value = id
  analyzeStartTime.value = Date.now()
  agents.value.forEach(a => { a.output = ''; a.done = false; a.error = false; a.qualityScore = 0; a.judgeComment = ''; a.duration = 0; a.finishedAt = 0 })

  // 连接 WebSocket 接收 Agent 实时输出
  connectWebSocket()

  try {
    await analyzeMaterial(id)
  } catch (err) {
    toast.error('分析触发失败: ' + (err.response?.data?.error || err.message))
    analyzing.value = false
    currentAnalyzingId.value = null
  }
}

function connectWebSocket() {
  // Close previous WebSocket to prevent orphaned connections
  if (ws) {
    ws.onclose = null
    ws.onmessage = null
    ws.close()
    ws = null
  }
  const token = localStorage.getItem('token')
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws?token=${token}`
  ws = new WebSocket(wsUrl)

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)

      // 处理 Agent 实时输出
      if (data.type === 'agent_output') {
        // 后端发送 agent_name: "Analyst" / "QuizMaster" / "CardMaker" / "MapBuilder"
        const agentMap = {
          analyst: 0, quizmaster: 1, cardmaker: 2, mapbuilder: 3,
        }
        const name = (data.agent_name || '').toLowerCase()
        const idx = agentMap[name]
        if (idx !== undefined) {
          // content 是 AgentResult 对象（含 data, status, error, quality_score 等）
          const result = typeof data.content === 'object' ? data.content : {}

          // 提取耗时（duration_ms 是 Go time.Duration 的毫秒值）
          if (result.duration_ms) {
            agents.value[idx].duration = result.duration_ms
          }

          // 记录完成时间（相对于分析开始的时间戳）
          agents.value[idx].finishedAt = Date.now()

          if (result.status === 'success') {
            // 质量评分
            if (result.quality_score) {
              agents.value[idx].qualityScore = result.quality_score
            }
            // Judge 评语
            if (result.judge_comment) {
              agents.value[idx].judgeComment = result.judge_comment
            }
            // 格式化输出：展示摘要信息
            const agentData = result.data
            let summary = ''
            try {
              const parsed = typeof agentData === 'string' ? JSON.parse(agentData) : agentData
              if (parsed.cards) summary = `生成了 ${parsed.cards.length} 张知识卡片`
              else if (parsed.quizzes) summary = `生成了 ${parsed.quizzes.length} 道练习题`
              else if (parsed.nodes) summary = `构建了 ${parsed.nodes.length} 个知识节点，${(parsed.edges || []).length} 条关系边`
              else if (parsed.key_points) summary = `提取了 ${parsed.key_points.length} 个关键知识点`
              else summary = '分析完成'
            } catch {
              summary = '分析完成'
            }
            agents.value[idx].output = summary
            agents.value[idx].done = true
            agents.value[idx].error = false
          } else if (result.status === 'error') {
            agents.value[idx].output = `失败: ${result.error || '未知错误'}`
            agents.value[idx].done = true
            agents.value[idx].error = true
          }
        }
      }

      // 处理分析完成通知
      if (data.type === 'analysis_complete') {
        analyzing.value = false
        currentAnalyzingId.value = null
        if (isPolling.value) stopPolling()
        // Close WebSocket to free resources after analysis
        if (ws) { ws.onclose = null; ws.close(); ws = null }
        loadMaterials()
        loadTags()
      }

      // 处理批量分析进度
      if (data.type === 'batch_analyze_progress') {
        const content = typeof data.content === 'object' ? data.content : {}
        batchProgress.value.completed = (batchProgress.value.completed || 0) + 1
        // 更新对应材料的本地状态
        if (content.status === 'completed' || content.status === 'failed') {
          materials.value = materials.value.map(m => {
            if (m.id === content.material_id) {
              return { ...m, status: content.status === 'completed' ? 'completed' : 'failed' }
            }
            return m
          })
        }
      }

      // 处理批量分析全部完成
      if (data.type === 'batch_analyze_complete') {
        batchAnalyzing.value = false
        batchProgress.value = { completed: 0, total: 0 }
        selectedIds.value = new Set()
        // Close WebSocket to free resources
        if (ws) { ws.onclose = null; ws.close(); ws = null }
        loadMaterials()
        toast.success('批量分析全部完成')
      }

      // 处理错误通知
      if (data.type === 'error') {
        console.error('分析错误:', data.content)
        analyzing.value = false
        currentAnalyzingId.value = null
        if (isPolling.value) stopPolling()
        // Close WebSocket to free resources
        if (ws) { ws.onclose = null; ws.close(); ws = null }
        loadMaterials()
      }
    } catch (e) {
      console.error('WebSocket 消息解析失败:', e)
    }
  }

  ws.onopen = () => {
    wsRetryCount = 0 // 连接成功，重置重试计数
    // WebSocket 重连成功，停止轮询 fallback
    if (isPolling.value) {
      stopPolling()
    }
  }

  ws.onclose = () => {
    console.log('WebSocket 连接关闭')
    // 分析进行中且未超过最大重试次数时，自动重连
    if (analyzing.value && wsRetryCount < MAX_WS_RETRIES) {
      wsRetryCount++
      const delay = Math.min(1000 * Math.pow(2, wsRetryCount - 1), 10000)
      console.log(`WebSocket 将在 ${delay}ms 后重连 (第 ${wsRetryCount} 次)`)
      // WebSocket 断线，启动轮询 fallback 获取 Agent 进度
      if (currentAnalyzingId.value && !isPolling.value) {
        startPolling(currentAnalyzingId.value)
      }
      setTimeout(() => {
        if (analyzing.value) connectWebSocket()
      }, delay)
    } else if (analyzing.value && wsRetryCount >= MAX_WS_RETRIES) {
      // 超过最大重试次数，仅依赖轮询
      console.log('WebSocket 重连次数已达上限，切换到轮询模式')
      if (currentAnalyzingId.value && !isPolling.value) {
        startPolling(currentAnalyzingId.value)
      }
    }
  }
}

// ========== 轮询 Fallback ==========

function startPolling(materialId) {
  if (pollTimer) clearInterval(pollTimer)
  isPolling.value = true
  console.log(`启动轮询 fallback，材料 ID: ${materialId}`)

  // 立即执行一次
  pollMaterialStatus(materialId)

  // 每 3 秒轮询一次
  pollTimer = setInterval(() => {
    pollMaterialStatus(materialId)
  }, 3000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  isPolling.value = false
  console.log('轮询 fallback 已停止')
}

async function pollMaterialStatus(materialId) {
  try {
    const res = await getMaterialStatus(materialId)
    const data = res.data

    // 更新各 Agent 的进度状态
    if (data.agents && Array.isArray(data.agents)) {
      const agentMap = {
        analyst: 0, quizmaster: 1, cardmaker: 2, mapbuilder: 3,
      }
      for (const agent of data.agents) {
        const name = (agent.name || '').toLowerCase()
        const idx = agentMap[name]
        if (idx !== undefined) {
          if (agent.status === 'done' && !agents.value[idx].done) {
            agents.value[idx].done = true
            agents.value[idx].error = false
            agents.value[idx].output = '分析完成（轮询同步）'
            agents.value[idx].finishedAt = Date.now()
          } else if (agent.status === 'error' && !agents.value[idx].done) {
            agents.value[idx].done = true
            agents.value[idx].error = true
            agents.value[idx].output = '分析失败（轮询同步）'
            agents.value[idx].finishedAt = Date.now()
          }
        }
      }
    }

    // 分析完成，停止轮询并刷新材料列表
    if (data.completed) {
      analyzing.value = false
      currentAnalyzingId.value = null
      stopPolling()
      loadMaterials()
      toast.success('材料分析完成')
    }
  } catch (err) {
    console.error('轮询材料状态失败:', err)
  }
}

async function handleDelete(id) {
  const ok = await confirm('确定删除这份材料？')
  if (!ok) return
  try {
    await deleteMaterial(id)
    await loadMaterials()
    toast.success('材料已删除')
  } catch (err) {
    toast.error('删除失败')
  }
}

async function loadMaterials() {
  try {
    const params = { limit: 200 }
    if (activeTagFilter.value) params.tag = activeTagFilter.value
    const res = await listMaterials(params)
    materials.value = res.data.data || []
  } catch (e) {
    console.error('材料列表加载失败:', e)
  }
}

onMounted(() => {
  loadMaterials()
  loadTags()
})
onUnmounted(() => {
  if (ws) ws.close()
  if (isPolling.value) stopPolling()
  if (hoverTimer) clearTimeout(hoverTimer)
})
</script>

<style scoped>
.preview-enter-active,
.preview-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}
.preview-enter-from,
.preview-leave-to {
  opacity: 0;
  transform: translateX(-6px);
}
.batch-bar-enter-active,
.batch-bar-leave-active {
  transition: all 0.25s ease;
}
.batch-bar-enter-from,
.batch-bar-leave-to {
  opacity: 0;
  transform: translateY(-8px);
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
.tag-suggest-enter-active,
.tag-suggest-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.tag-suggest-enter-from,
.tag-suggest-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
