<template>
  <div class="h-full flex flex-col p-4 sm:p-6 lg:p-8">
    <!-- 离线缓存数据提示 -->
    <div
      v-if="!notesLoading && fromCache"
      class="mb-3 flex items-center gap-2 px-3 py-2 rounded-lg text-xs font-medium bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-900/20 dark:text-amber-400 dark:border-amber-800/30 flex-shrink-0"
    >
      <svg class="w-3.5 h-3.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
      <span>正在查看缓存数据，编辑操作将排队等待联网同步</span>
    </div>

    <!-- 页头 -->
    <div class="flex items-center justify-between mb-4 flex-shrink-0">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">知识笔记本</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Markdown 笔记 · 知识整理利器</p>
      </div>
      <!-- 移动端：汉堡菜单展开文件夹树 -->
      <button
        class="lg:hidden inline-flex items-center gap-2 px-3 py-2 rounded-lg text-sm border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
        @click="showMobileFolders = !showMobileFolders"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
          />
        </svg>
        文件夹
      </button>
    </div>

    <!-- 三栏布局 -->
    <div class="flex-1 flex gap-4 min-h-0 overflow-hidden">
      <!-- ====== 左侧：文件夹树 ====== -->
      <aside
        class="flex-shrink-0 w-52 flex flex-col bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden transition-all fixed lg:static inset-0 z-40 lg:z-auto"
        :class="showMobileFolders ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'"
      >
        <!-- 文件夹头部 -->
        <div class="flex items-center justify-between px-3 py-2.5 border-b border-gray-100 dark:border-gray-700">
          <span class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">文件夹</span>
          <button
            class="text-gray-400 hover:text-primary-600 dark:hover:text-primary-400 transition-colors"
            title="新建文件夹"
            @click="showNewFolderForm = !showNewFolderForm"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
            </svg>
          </button>
        </div>

        <!-- 新建文件夹表单 -->
        <div
          v-if="showNewFolderForm"
          class="px-3 py-2 border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-750"
        >
          <input
            v-model="newFolderName"
            placeholder="文件夹名称"
            class="w-full text-sm rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-200 px-2.5 py-1.5 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none mb-2"
            @keyup.enter="handleCreateFolder"
          />
          <div class="flex gap-1 mb-2 flex-wrap">
            <button
              v-for="color in folderColors"
              :key="color"
              class="w-5 h-5 rounded-full border-2 transition-transform"
              :class="newFolderColor === color ? 'border-gray-800 dark:border-white scale-125' : 'border-transparent'"
              :style="{ backgroundColor: color }"
              @click="newFolderColor = color"
            />
          </div>
          <div class="flex gap-2">
            <button
              :disabled="!newFolderName.trim()"
              class="flex-1 text-xs py-1.5 rounded-lg bg-primary-600 text-white font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              @click="handleCreateFolder"
            >
              创建
            </button>
            <button
              class="flex-1 text-xs py-1.5 rounded-lg border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
              @click="cancelNewFolder()"
            >
              取消
            </button>
          </div>
        </div>

        <!-- 文件夹列表 -->
        <nav class="flex-1 overflow-y-auto py-1.5 custom-scroll">
          <!-- 全部笔记 -->
          <button
            class="w-full flex items-center gap-2 px-3 py-2 text-sm transition-colors"
            :class="
              activeFolderId === null && activeFolderSpecial === 'all'
                ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400 font-medium'
                : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
            "
            @click="selectFolder(null)"
          >
            <svg class="w-4 h-4 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
              />
            </svg>
            <span class="truncate">全部笔记</span>
            <span class="ml-auto text-xs opacity-60">{{ allNotesCount }}</span>
          </button>

          <!-- 无文件夹 -->
          <button
            class="w-full flex items-center gap-2 px-3 py-2 text-sm transition-colors"
            :class="
              activeFolderId === '__none__'
                ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400 font-medium'
                : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
            "
            @click="selectFolder('__none__')"
          >
            <svg
              class="w-4 h-4 flex-shrink-0 opacity-50"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"
              />
            </svg>
            <span class="truncate">未分类</span>
            <span class="ml-auto text-xs opacity-60">{{ noFolderCount }}</span>
          </button>

          <div class="my-1.5 mx-3 border-t border-gray-100 dark:border-gray-700"></div>

          <!-- 自定义文件夹 -->
          <div v-for="folder in folders" :key="folder.id" class="group relative">
            <button
              class="w-full flex items-center gap-2 px-3 py-2 text-sm transition-colors"
              :class="
                activeFolderId === folder.id
                  ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400 font-medium'
                  : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
              "
              @click="selectFolder(folder.id)"
            >
              <svg class="w-4 h-4 flex-shrink-0" fill="currentColor" :style="{ color: folder.color }">
                <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
              </svg>
              <span class="truncate">{{ folder.name }}</span>
              <span class="ml-auto text-xs opacity-60">{{ folder.note_count }}</span>
            </button>
            <!-- 右键/长按操作 -->
            <div
              class="absolute right-1 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 flex gap-0.5 transition-opacity"
            >
              <button
                class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                title="重命名"
                @click.stop="startRenameFolder(folder)"
              >
                <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
              </button>
              <button
                class="p-1 rounded hover:bg-red-100 dark:hover:bg-red-900/30 text-gray-400 hover:text-red-500"
                title="删除"
                @click.stop="handleDeleteFolder(folder)"
              >
                <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
              </button>
            </div>
          </div>
        </nav>

        <!-- 移动端关闭按钮 -->
        <button
          class="lg:hidden flex-shrink-0 py-2.5 text-sm text-gray-500 dark:text-gray-400 border-t border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          @click="showMobileFolders = false"
        >
          关闭
        </button>
      </aside>

      <!-- 移动端遮罩 -->
      <div
        v-if="showMobileFolders"
        class="lg:hidden fixed inset-0 z-30 bg-black/40"
        @click="showMobileFolders = false"
      ></div>

      <!-- ====== 中间：笔记列表 ====== -->
      <div
        class="w-full sm:w-64 lg:w-72 flex-shrink-0 flex flex-col bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden"
      >
        <!-- 搜索栏 -->
        <div class="px-3 py-2.5 border-b border-gray-100 dark:border-gray-700">
          <div class="relative">
            <svg
              class="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              v-model="searchQuery"
              placeholder="搜索笔记..."
              class="w-full pl-8 pr-3 py-1.5 text-sm rounded-lg border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 text-gray-800 dark:text-gray-200 placeholder-gray-400 dark:placeholder-gray-500 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition"
              @input="handleSearch"
            />
          </div>
        </div>

        <!-- 新建笔记按钮 -->
        <div class="px-3 py-2 border-b border-gray-100 dark:border-gray-700">
          <button
            class="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg text-sm font-medium bg-primary-600 text-white hover:bg-primary-700 dark:bg-primary-500 dark:hover:bg-primary-600 transition-colors"
            @click="handleCreateNote"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
            </svg>
            新建笔记
          </button>
        </div>

        <!-- 笔记列表 -->
        <div class="flex-1 overflow-y-auto custom-scroll">
          <!-- 加载骨架 -->
          <div v-if="notesLoading" class="p-3 space-y-2">
            <div v-for="i in 4" :key="i" class="animate-pulse">
              <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-2"></div>
              <div class="h-2 bg-gray-100 dark:bg-gray-700/50 rounded w-1/2"></div>
            </div>
          </div>

          <!-- 空状态 -->
          <div v-else-if="notes.length === 0" class="flex flex-col items-center justify-center h-48 px-4">
            <svg
              class="w-12 h-12 text-gray-300 dark:text-gray-600 mb-3"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="1.5"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
              />
            </svg>
            <p class="text-sm text-gray-400 dark:text-gray-500 text-center">
              {{ searchQuery ? '没有找到匹配的笔记' : '这里还没有笔记' }}
            </p>
            <button
              v-if="!searchQuery"
              class="mt-2 text-xs text-primary-600 dark:text-primary-400 hover:underline"
              @click="handleCreateNote"
            >
              创建第一条笔记
            </button>
          </div>

          <!-- 笔记卡片 -->
          <button
            v-for="note in notes"
            :key="note.id"
            class="w-full text-left px-3 py-2.5 border-b border-gray-50 dark:border-gray-700/50 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
            :class="
              activeNoteId === note.id ? 'bg-primary-50 dark:bg-primary-900/20 border-l-2 border-l-primary-500' : ''
            "
            @click="selectNote(note)"
          >
            <div class="flex items-start gap-2">
              <svg
                v-if="note.pinned"
                class="w-3 h-3 text-amber-500 mt-0.5 flex-shrink-0"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path d="M9.828 4.172a4 4 0 015.657 5.657l-5.657 5.657-5.657-5.657a4 4 0 015.657-5.657z" />
              </svg>
              <div class="flex-1 min-w-0">
                <h3 class="text-sm font-medium text-gray-800 dark:text-gray-200 truncate">
                  {{ note.title || '无标题笔记' }}
                </h3>
                <p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5 line-clamp-2">
                  {{ notePreview(note.content) }}
                </p>
                <p class="text-xs text-gray-300 dark:text-gray-600 mt-1">
                  {{ formatTime(note.updated_at) }}
                </p>
              </div>
            </div>
          </button>
        </div>
      </div>

      <!-- ====== 右侧：编辑器 + 预览 ====== -->
      <div
        class="flex-1 flex flex-col bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden min-w-0"
      >
        <!-- 未选择笔记状态 -->
        <div v-if="!activeNote" class="flex-1 flex flex-col items-center justify-center">
          <svg
            class="w-16 h-16 text-gray-200 dark:text-gray-700 mb-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="1"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
            />
          </svg>
          <p class="text-gray-400 dark:text-gray-500 text-sm">选择一条笔记开始编辑</p>
          <p class="text-gray-300 dark:text-gray-600 text-xs mt-1">或创建一条新笔记</p>
        </div>

        <!-- 编辑器 -->
        <template v-else>
          <!-- 工具栏 -->
          <div class="flex items-center gap-2 px-4 py-2.5 border-b border-gray-100 dark:border-gray-700 flex-shrink-0">
            <!-- 标题输入 -->
            <input
              v-model="activeNote.title"
              placeholder="笔记标题"
              class="flex-1 text-base font-semibold bg-transparent text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 outline-none border-none"
              @input="debounceSave"
            />
            <!-- 固定按钮 -->
            <button
              class="p-1.5 rounded-lg transition-colors"
              :class="
                activeNote.pinned
                  ? 'text-amber-500 bg-amber-50 dark:bg-amber-900/30'
                  : 'text-gray-400 hover:text-amber-500 hover:bg-gray-100 dark:hover:bg-gray-700'
              "
              :title="activeNote.pinned ? '取消固定' : '固定笔记'"
              @click="togglePin"
            >
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path d="M9.828 4.172a4 4 0 015.657 5.657l-5.657 5.657-5.657-5.657a4 4 0 015.657-5.657z" />
              </svg>
            </button>
            <!-- 文件夹选择 -->
            <select
              v-model="activeNote.folder_id"
              class="text-xs rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-400 px-2 py-1.5 focus:ring-2 focus:ring-primary-500 outline-none"
              @change="handleMoveToFolder"
            >
              <option value="">未分类</option>
              <option v-for="f in folders" :key="f.id" :value="f.id">{{ f.name }}</option>
            </select>
            <!-- 预览切换 -->
            <button
              class="p-1.5 rounded-lg transition-colors"
              :class="
                showPreview
                  ? 'text-primary-600 bg-primary-50 dark:bg-primary-900/30 dark:text-primary-400'
                  : 'text-gray-400 hover:text-primary-600 hover:bg-gray-100 dark:hover:bg-gray-700'
              "
              title="预览/编辑"
              @click="showPreview = !showPreview"
            >
              <svg
                v-if="!showPreview"
                class="w-4 h-4"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                />
              </svg>
              <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                />
              </svg>
            </button>
            <!-- 删除 -->
            <button
              class="p-1.5 rounded-lg text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/30 transition-colors"
              title="删除笔记"
              @click="handleDeleteNote"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                />
              </svg>
            </button>
          </div>

          <!-- 编辑区 / 预览区 -->
          <div class="flex-1 flex min-h-0 overflow-hidden">
            <!-- Markdown 编辑器 -->
            <div class="flex-1 flex flex-col min-w-0" :class="showPreview ? 'hidden sm:flex sm:w-1/2' : 'flex'">
              <textarea
                ref="editorRef"
                v-model="activeNote.content"
                placeholder="开始用 Markdown 写下你的想法..."
                class="flex-1 w-full resize-none bg-transparent text-sm text-gray-800 dark:text-gray-200 placeholder-gray-400 dark:placeholder-gray-500 outline-none p-4 font-mono leading-relaxed custom-scroll"
                @input="debounceSave"
              ></textarea>
            </div>

            <!-- Markdown 预览 -->
            <div
              v-if="showPreview"
              class="flex-1 overflow-y-auto p-4 border-l border-gray-100 dark:border-gray-700 custom-scroll prose prose-sm dark:prose-invert max-w-none"
              v-html="renderedMarkdown"
            ></div>
          </div>

          <!-- 状态栏 -->
          <div
            class="flex items-center justify-between px-4 py-1.5 border-t border-gray-100 dark:border-gray-700 text-xs text-gray-400 dark:text-gray-500 flex-shrink-0"
          >
            <span>{{ charCount }} 字</span>
            <span v-if="saveStatus">{{ saveStatus }}</span>
          </div>
        </template>
      </div>
    </div>

    <!-- 重命名弹窗 -->
    <Transition name="fade">
      <div
        v-if="renameState.show"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
        @click.self="renameState.show = false"
      >
        <div
          class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-5 w-80 border border-gray-200 dark:border-gray-700"
        >
          <h3 class="text-sm font-semibold text-gray-800 dark:text-gray-200 mb-3">重命名文件夹</h3>
          <input
            v-model="renameState.name"
            class="w-full text-sm rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-200 px-3 py-2 focus:ring-2 focus:ring-primary-500 outline-none mb-3"
            @keyup.enter="handleRenameFolder"
          />
          <div class="flex gap-2 justify-end">
            <button
              class="px-3 py-1.5 text-sm rounded-lg border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700"
              @click="renameState.show = false"
            >
              取消
            </button>
            <button
              class="px-3 py-1.5 text-sm rounded-lg bg-primary-600 text-white hover:bg-primary-700"
              @click="handleRenameFolder"
            >
              确认
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import {
  listNotes,
  getNote,
  createNote,
  updateNote,
  deleteNote,
  searchNotes,
  listNoteFolders,
  createNoteFolder,
  updateNoteFolder,
  deleteNoteFolder
} from '../api/client.js'
import { useToast, useConfirm } from '../composables/useToast.js'
import { useNetworkStatus } from '../composables/useNetworkStatus.js'

const toast = useToast()
const { confirm } = useConfirm()
const { isOnline: _isOnline, isResponseFromCache } = useNetworkStatus()

// ========== 状态 ==========
const folders = ref([])
const notes = ref([])
const notesLoading = ref(false)
const fromCache = ref(false)
const activeFolderId = ref(null) // null = all, '__none__' = uncategorized, or folder ID
const activeFolderSpecial = ref('all') // 'all' when showing all
const activeNoteId = ref(null)
const activeNote = ref(null)
const searchQuery = ref('')
const showPreview = ref(true)
const showMobileFolders = ref(false)
const showNewFolderForm = ref(false)
const newFolderName = ref('')
const newFolderColor = ref('#6366f1')
const allNotesCount = ref(0)
const noFolderCount = ref(0)
const saveStatus = ref('')
const editorRef = ref(null)

const renameState = reactive({ show: false, folderId: '', name: '' })

const folderColors = ['#6366f1', '#ec4899', '#10b981', '#f59e0b', '#3b82f6', '#8b5cf6', '#ef4444', '#06b6d4']

let saveTimer = null
let searchTimer = null

// ========== 计算属性 ==========
const charCount = computed(() => (activeNote.value?.content || '').length)

const renderedMarkdown = computed(() => {
  if (!activeNote.value?.content)
    return '<p class="text-gray-400 dark:text-gray-500 italic">预览区域 — 编辑笔记后这里会实时渲染 Markdown</p>'
  try {
    const html = marked.parse(activeNote.value.content, { breaks: true, gfm: true })
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
        'hr',
        'ul',
        'ol',
        'li',
        'a',
        'strong',
        'em',
        'del',
        'code',
        'pre',
        'blockquote',
        'table',
        'thead',
        'tbody',
        'tr',
        'th',
        'td',
        'img',
        'span',
        'div',
        'input',
        'sup',
        'sub'
      ],
      ALLOWED_ATTR: ['href', 'target', 'rel', 'src', 'alt', 'class', 'id', 'type', 'checked', 'disabled'],
      ALLOW_DATA_ATTR: false
    })
  } catch {
    return '<p class="text-red-500">Markdown 渲染失败</p>'
  }
})

// ========== 数据加载 ==========
async function loadFolders() {
  try {
    const res = await listNoteFolders()
    folders.value = res.data.folders || []
    noFolderCount.value = res.data.no_folder_count || 0
  } catch {
    folders.value = []
  }
}

async function loadNotes() {
  notesLoading.value = true
  try {
    const params = { limit: 100, offset: 0 }
    if (activeFolderId.value !== null || activeFolderSpecial.value !== 'all') {
      if (activeFolderId.value) params.folder_id = activeFolderId.value
    }

    let res
    if (searchQuery.value.trim()) {
      res = await searchNotes(searchQuery.value.trim())
      notes.value = res.data.data || []
    } else {
      res = await listNotes(params)
      notes.value = res.data.data || []
    }

    // 刷新全部计数
    if (!searchQuery.value.trim() && activeFolderSpecial.value === 'all') {
      allNotesCount.value = res.data.total || notes.value.length
    }

    fromCache.value = isResponseFromCache(res)
  } catch {
    notes.value = []
  } finally {
    notesLoading.value = false
  }
}

async function refreshAll() {
  await Promise.all([loadFolders(), loadNotes()])
}

// ========== 文件夹操作 ==========
function selectFolder(folderId) {
  activeFolderId.value = folderId
  activeFolderSpecial.value = folderId === null ? 'all' : ''
  activeNoteId.value = null
  activeNote.value = null
  searchQuery.value = ''
  showMobileFolders.value = false
  loadNotes()
}

async function handleCreateFolder() {
  const name = newFolderName.value.trim()
  if (!name) return
  try {
    await createNoteFolder({ name, color: newFolderColor.value })
    toast.success('文件夹创建成功')
    newFolderName.value = ''
    newFolderColor.value = '#6366f1'
    showNewFolderForm.value = false
    await loadFolders()
  } catch {
    toast.error('创建文件夹失败')
  }
}

function cancelNewFolder() {
  showNewFolderForm.value = false
  newFolderName.value = ''
}

function startRenameFolder(folder) {
  renameState.folderId = folder.id
  renameState.name = folder.name
  renameState.show = true
}

async function handleRenameFolder() {
  const name = renameState.name.trim()
  if (!name) return
  try {
    await updateNoteFolder(renameState.folderId, { name })
    toast.success('重命名成功')
    renameState.show = false
    await loadFolders()
  } catch {
    toast.error('重命名失败')
  }
}

async function handleDeleteFolder(folder) {
  const ok = await confirm(`确定删除文件夹「${folder.name}」？\n其中的笔记将移至"未分类"。`)
  if (!ok) return
  try {
    await deleteNoteFolder(folder.id)
    toast.success('文件夹已删除')
    if (activeFolderId.value === folder.id) {
      selectFolder(null)
    }
    await refreshAll()
  } catch {
    toast.error('删除文件夹失败')
  }
}

// ========== 笔记操作 ==========
async function selectNote(note) {
  activeNoteId.value = note.id
  try {
    const res = await getNote(note.id)
    activeNote.value = res.data
    await nextTick()
    if (editorRef.value) editorRef.value.focus()
  } catch {
    toast.error('加载笔记失败')
  }
}

async function handleCreateNote() {
  const folderId = activeFolderId.value && activeFolderId.value !== '__none__' ? activeFolderId.value : ''
  try {
    const res = await createNote({
      title: '',
      content: '',
      folder_id: folderId
    })
    toast.success('笔记已创建')
    await loadNotes()
    await selectNote(res.data)
  } catch {
    toast.error('创建笔记失败')
  }
}

function debounceSave() {
  if (saveTimer) clearTimeout(saveTimer)
  saveStatus.value = '编辑中...'
  saveTimer = setTimeout(() => saveActiveNote(), 800)
}

async function saveActiveNote() {
  if (!activeNote.value) return
  try {
    await updateNote(activeNote.value.id, {
      title: activeNote.value.title,
      content: activeNote.value.content
    })
    saveStatus.value = '已保存'
    setTimeout(() => {
      if (saveStatus.value === '已保存') saveStatus.value = ''
    }, 2000)
    // 更新列表中的该条目
    const idx = notes.value.findIndex((n) => n.id === activeNote.value.id)
    if (idx !== -1) {
      notes.value[idx].title = activeNote.value.title
      notes.value[idx].content = activeNote.value.content
      notes.value[idx].updated_at = new Date().toISOString()
    }
  } catch {
    saveStatus.value = '保存失败'
    toast.error('保存笔记失败')
  }
}

async function togglePin() {
  if (!activeNote.value) return
  const newPinned = !activeNote.value.pinned
  try {
    await updateNote(activeNote.value.id, { pinned: newPinned })
    activeNote.value.pinned = newPinned
    toast.success(newPinned ? '已固定' : '已取消固定')
    await loadNotes()
  } catch {
    toast.error('操作失败')
  }
}

async function handleMoveToFolder() {
  if (!activeNote.value) return
  try {
    await updateNote(activeNote.value.id, { folder_id: activeNote.value.folder_id || '' })
    toast.success('已移动')
    await Promise.all([loadFolders(), loadNotes()])
  } catch {
    toast.error('移动失败')
  }
}

async function handleDeleteNote() {
  if (!activeNote.value) return
  const ok = await confirm(`确定删除笔记「${activeNote.value.title || '无标题笔记'}」？`)
  if (!ok) return
  try {
    await deleteNote(activeNote.value.id)
    toast.success('笔记已删除')
    activeNoteId.value = null
    activeNote.value = null
    await refreshAll()
  } catch {
    toast.error('删除笔记失败')
  }
}

// ========== 搜索 ==========
function handleSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadNotes(), 300)
}

// ========== 辅助 ==========
function notePreview(content) {
  if (!content) return '空笔记'
  // eslint-disable-next-line no-useless-escape
  return content.replace(/[#*`>\-\[\]()!]/g, '').slice(0, 80)
}

function formatTime(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diffMs = now - d
  const diffMin = Math.floor(diffMs / 60000)
  if (diffMin < 1) return '刚刚'
  if (diffMin < 60) return `${diffMin} 分钟前`
  const diffHr = Math.floor(diffMin / 60)
  if (diffHr < 24) return `${diffHr} 小时前`
  const diffDay = Math.floor(diffHr / 24)
  if (diffDay < 7) return `${diffDay} 天前`
  return d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

// ========== 生命周期 ==========
onMounted(() => {
  refreshAll()
})
</script>

<style scoped>
/* 细滚动条 */
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
.dark .custom-scroll::-webkit-scrollbar-thumb {
  background: #4b5563;
}

/* Markdown 预览样式 */
.prose :deep(h1) {
  font-size: 1.5rem;
  font-weight: 700;
  margin-top: 1.5rem;
  margin-bottom: 0.75rem;
}
.prose :deep(h2) {
  font-size: 1.25rem;
  font-weight: 600;
  margin-top: 1.25rem;
  margin-bottom: 0.5rem;
}
.prose :deep(h3) {
  font-size: 1.1rem;
  font-weight: 600;
  margin-top: 1rem;
  margin-bottom: 0.5rem;
}
.prose :deep(p) {
  margin-bottom: 0.75rem;
  line-height: 1.7;
}
.prose :deep(ul),
.prose :deep(ol) {
  padding-left: 1.5rem;
  margin-bottom: 0.75rem;
}
.prose :deep(li) {
  margin-bottom: 0.25rem;
}
.prose :deep(code) {
  background: rgba(99, 102, 241, 0.08);
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  font-size: 0.85em;
}
.dark .prose :deep(code) {
  background: rgba(99, 102, 241, 0.15);
}
.prose :deep(pre) {
  background: #1e1e2e;
  color: #cdd6f4;
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-x: auto;
  margin-bottom: 1rem;
}
.prose :deep(pre code) {
  background: none;
  padding: 0;
  color: inherit;
}
.prose :deep(blockquote) {
  border-left: 3px solid #6366f1;
  padding-left: 1rem;
  color: #6b7280;
  margin-bottom: 0.75rem;
}
.dark .prose :deep(blockquote) {
  color: #9ca3af;
}
.prose :deep(a) {
  color: #6366f1;
  text-decoration: underline;
}
.prose :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 1rem;
}
.prose :deep(th),
.prose :deep(td) {
  border: 1px solid #e5e7eb;
  padding: 0.5rem;
  text-align: left;
}
.dark .prose :deep(th),
.dark .prose :deep(td) {
  border-color: #374151;
}
.prose :deep(hr) {
  border: none;
  border-top: 1px solid #e5e7eb;
  margin: 1.5rem 0;
}
.dark .prose :deep(hr) {
  border-color: #374151;
}
.prose :deep(img) {
  max-width: 100%;
  border-radius: 0.5rem;
}

/* 过渡 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* line-clamp polyfill */
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
