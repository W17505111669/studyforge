<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-6xl mx-auto">
    <!-- 顶部标题 -->
    <div class="mb-6">
      <h1 class="text-xl sm:text-2xl font-bold text-gray-800 dark:text-gray-100 flex items-center gap-2">
        <svg class="w-6 h-6 text-emerald-500" viewBox="0 0 20 20" fill="currentColor">
          <path
            d="M3 1a1 1 0 000 2h1.22l.305 1.222a.997.997 0 00.01.042l1.358 5.43-.893.892C3.74 11.846 4.632 14 6.414 14H15a1 1 0 000-2H6.414l1-1H14a1 1 0 00.893-.553l3-6A1 1 0 0017 3H6.28l-.31-1.243A1 1 0 005 1H3zM16 16.5a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0zM6.5 18a1.5 1.5 0 100-3 1.5 1.5 0 000 3z"
          />
        </svg>
        学习市场
      </h1>
      <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">浏览和收藏其他用户分享的学习材料和卡片牌组</p>
    </div>

    <!-- Tab 切换 -->
    <div class="flex gap-1 mb-6 bg-gray-100 dark:bg-gray-800 p-1 rounded-lg w-fit">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        :class="[
          'px-4 py-2 rounded-md text-sm font-medium transition',
          activeTab === tab.value
            ? 'bg-white dark:bg-gray-700 text-emerald-600 dark:text-emerald-400 shadow-sm'
            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
        ]"
        @click="handleTabClick(tab.value)"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- ==================== 材料 Tab ==================== -->
    <div v-if="activeTab === 'materials'">
      <!-- 搜索栏 + 排序 -->
      <div class="flex flex-col sm:flex-row gap-3 mb-4">
        <div class="relative flex-1">
          <svg
            class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fill-rule="evenodd"
              d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
              clip-rule="evenodd"
            />
          </svg>
          <input
            v-model="matSearch"
            type="text"
            placeholder="搜索材料标题..."
            class="w-full pl-10 pr-4 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm dark:text-gray-200 focus:ring-2 focus:ring-emerald-500 focus:border-transparent outline-none transition"
            @input="debounceMatSearch"
          />
        </div>
        <div class="flex gap-2 shrink-0">
          <button
            v-for="s in matSortOptions"
            :key="s.value"
            :class="[
              'px-3 py-2 rounded-lg text-xs font-medium transition',
              matSort === s.value
                ? 'bg-emerald-500 text-white'
                : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600'
            ]"
            @click="setMatSort(s.value)"
          >
            {{ s.label }}
          </button>
        </div>
      </div>

      <!-- 标签过滤 -->
      <div v-if="matTags.length > 0" class="flex flex-wrap gap-1.5 mb-5">
        <button
          :class="[
            'px-2.5 py-1 rounded-full text-xs font-medium transition',
            !matActiveTag
              ? 'bg-emerald-500 text-white'
              : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600'
          ]"
          @click="clearMatTag()"
        >
          全部
        </button>
        <button
          v-for="tag in matTags.slice(0, 20)"
          :key="tag.name"
          :class="[
            'px-2.5 py-1 rounded-full text-xs font-medium transition',
            matActiveTag === tag.name
              ? 'bg-emerald-500 text-white'
              : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600'
          ]"
          @click="toggleMatTag(tag.name)"
        >
          {{ tag.name }}
          <span class="ml-0.5 opacity-60">{{ tag.count }}</span>
        </button>
      </div>

      <!-- 材料卡片网格 -->
      <div v-if="matLoading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="i in 6" :key="i" class="rounded-xl border border-gray-200 dark:border-gray-700 p-5">
          <div class="skeleton h-5 w-3/4 mb-3"></div>
          <div class="skeleton h-3 w-1/2 mb-2"></div>
          <div class="skeleton h-3 w-full mb-2"></div>
          <div class="flex gap-2 mt-4">
            <div class="skeleton h-6 w-16 rounded-full"></div>
            <div class="skeleton h-6 w-16 rounded-full"></div>
          </div>
        </div>
      </div>

      <div v-else-if="materials.length === 0" class="text-center py-16">
        <svg class="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-4" viewBox="0 0 20 20" fill="currentColor">
          <path
            d="M3 1a1 1 0 000 2h1.22l.305 1.222a.997.997 0 00.01.042l1.358 5.43-.893.892C3.74 11.846 4.632 14 6.414 14H15a1 1 0 000-2H6.414l1-1H14a1 1 0 00.893-.553l3-6A1 1 0 0017 3H6.28l-.31-1.243A1 1 0 005 1H3z"
          />
        </svg>
        <p class="text-gray-400 dark:text-gray-500 text-sm">暂无公开材料</p>
        <p class="text-gray-400 dark:text-gray-500 text-xs mt-1">去上传页分享你的学习材料吧！</p>
        <router-link
          to="/upload"
          class="inline-block mt-4 px-4 py-2 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600 transition"
        >
          上传材料
        </router-link>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="m in materials"
          :key="m.id"
          class="group rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-5 hover:border-emerald-300 dark:hover:border-emerald-600 transition cursor-pointer"
          @click="openMatPreview(m)"
        >
          <h3
            class="font-semibold text-gray-800 dark:text-gray-100 text-sm leading-snug line-clamp-2 mb-2 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition"
          >
            {{ m.title }}
          </h3>
          <div class="flex items-center gap-1.5 mb-2">
            <div
              class="w-5 h-5 rounded-full bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center text-emerald-600 dark:text-emerald-400 text-[10px] font-bold"
            >
              {{ (m.author_name || '匿')[0] }}
            </div>
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ m.author_name || '匿名用户' }}</span>
          </div>
          <div v-if="m.tags" class="flex flex-wrap gap-1 mb-3">
            <span
              v-for="t in m.tags.split(',').slice(0, 4)"
              :key="t"
              class="px-1.5 py-0.5 rounded text-[10px] bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400"
            >
              {{ t.trim() }}
            </span>
            <span v-if="m.tags.split(',').length > 4" class="text-[10px] text-gray-400">
              +{{ m.tags.split(',').length - 4 }}
            </span>
          </div>
          <div class="flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500">
            <span class="flex items-center gap-1">
              <svg class="w-3.5 h-3.5" viewBox="0 0 20 20" fill="currentColor">
                <path
                  d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
                />
              </svg>
              {{ m.card_count || 0 }} 卡片
            </span>
            <span>{{ formatDate(m.created_at) }}</span>
          </div>
        </div>
      </div>

      <!-- 材料分页 -->
      <div v-if="matTotal > matLimit" class="flex justify-center items-center gap-2 mt-8">
        <button
          :disabled="matOffset === 0"
          class="px-3 py-1.5 rounded-lg text-sm bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 disabled:opacity-40 hover:bg-gray-200 dark:hover:bg-gray-600 transition"
          @click="matPrevPage()"
        >
          上一页
        </button>
        <span class="text-sm text-gray-500 dark:text-gray-400">
          {{ Math.floor(matOffset / matLimit) + 1 }} / {{ Math.ceil(matTotal / matLimit) }}
        </span>
        <button
          :disabled="matOffset + matLimit >= matTotal"
          class="px-3 py-1.5 rounded-lg text-sm bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 disabled:opacity-40 hover:bg-gray-200 dark:hover:bg-gray-600 transition"
          @click="matNextPage()"
        >
          下一页
        </button>
      </div>
    </div>

    <!-- ==================== 牌组 Tab ==================== -->
    <div v-if="activeTab === 'decks'">
      <!-- 搜索栏 + 排序 -->
      <div class="flex flex-col sm:flex-row gap-3 mb-4">
        <div class="relative flex-1">
          <svg
            class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fill-rule="evenodd"
              d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"
              clip-rule="evenodd"
            />
          </svg>
          <input
            v-model="deckSearch"
            type="text"
            placeholder="搜索牌组名称..."
            class="w-full pl-10 pr-4 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm dark:text-gray-200 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
            @input="debounceDeckSearch"
          />
        </div>
        <div class="flex gap-2 shrink-0">
          <button
            v-for="s in deckSortOptions"
            :key="s.value"
            :class="[
              'px-3 py-2 rounded-lg text-xs font-medium transition',
              deckSort === s.value
                ? 'bg-indigo-500 text-white'
                : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600'
            ]"
            @click="setDeckSort(s.value)"
          >
            {{ s.label }}
          </button>
        </div>
      </div>

      <!-- 标签过滤 -->
      <div v-if="deckTags.length > 0" class="flex flex-wrap gap-1.5 mb-5">
        <button
          :class="[
            'px-2.5 py-1 rounded-full text-xs font-medium transition',
            !deckActiveTag
              ? 'bg-indigo-500 text-white'
              : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600'
          ]"
          @click="clearDeckTag()"
        >
          全部
        </button>
        <button
          v-for="tag in deckTags.slice(0, 20)"
          :key="tag.name"
          :class="[
            'px-2.5 py-1 rounded-full text-xs font-medium transition',
            deckActiveTag === tag.name
              ? 'bg-indigo-500 text-white'
              : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600'
          ]"
          @click="toggleDeckTag(tag.name)"
        >
          {{ tag.name }}
          <span class="ml-0.5 opacity-60">{{ tag.count }}</span>
        </button>
      </div>

      <!-- 牌组卡片网格 -->
      <div v-if="deckLoading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="i in 6" :key="i" class="rounded-xl border border-gray-200 dark:border-gray-700 p-5">
          <div class="skeleton h-5 w-3/4 mb-3"></div>
          <div class="skeleton h-3 w-full mb-2"></div>
          <div class="flex gap-2 mt-4">
            <div class="skeleton h-6 w-16 rounded-full"></div>
            <div class="skeleton h-6 w-20 rounded-full"></div>
          </div>
        </div>
      </div>

      <div v-else-if="decks.length === 0" class="text-center py-16">
        <svg class="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-4" viewBox="0 0 20 20" fill="currentColor">
          <path
            d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
          />
        </svg>
        <p class="text-gray-400 dark:text-gray-500 text-sm">暂无公开牌组</p>
        <p class="text-gray-400 dark:text-gray-500 text-xs mt-1">去卡片页创建并分享你的牌组吧！</p>
        <router-link
          to="/cards"
          class="inline-block mt-4 px-4 py-2 bg-indigo-500 text-white rounded-lg text-sm hover:bg-indigo-600 transition"
        >
          查看卡片
        </router-link>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="d in decks"
          :key="d.id"
          class="group rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-5 hover:border-indigo-300 dark:hover:border-indigo-600 transition cursor-pointer"
          @click="openDeckPreview(d)"
        >
          <h3
            class="font-semibold text-gray-800 dark:text-gray-100 text-sm leading-snug line-clamp-2 mb-2 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition"
          >
            {{ d.name }}
          </h3>
          <p v-if="d.description" class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mb-2">
            {{ d.description }}
          </p>
          <div class="flex items-center gap-1.5 mb-2">
            <div
              class="w-5 h-5 rounded-full bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400 text-[10px] font-bold"
            >
              {{ (d.author_name || '匿')[0] }}
            </div>
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ d.author_name || '匿名用户' }}</span>
          </div>
          <div v-if="d.tags" class="flex flex-wrap gap-1 mb-3">
            <span
              v-for="t in d.tags.split(',').slice(0, 4)"
              :key="t"
              class="px-1.5 py-0.5 rounded text-[10px] bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400"
            >
              {{ t.trim() }}
            </span>
          </div>
          <div class="flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500">
            <span class="flex items-center gap-1">
              <svg class="w-3.5 h-3.5" viewBox="0 0 20 20" fill="currentColor">
                <path
                  d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
                />
              </svg>
              {{ d.card_count || 0 }} 卡片
            </span>
            <span class="flex items-center gap-1">
              <svg class="w-3.5 h-3.5" viewBox="0 0 20 20" fill="currentColor">
                <path d="M5 4a2 2 0 012-2h6a2 2 0 012 2v14l-5-2.5L5 18V4z" />
              </svg>
              {{ d.collect_count || 0 }} 收藏
            </span>
            <span>{{ formatDate(d.created_at) }}</span>
          </div>
        </div>
      </div>

      <!-- 牌组分页 -->
      <div v-if="deckTotal > deckLimit" class="flex justify-center items-center gap-2 mt-8">
        <button
          :disabled="deckOffset === 0"
          class="px-3 py-1.5 rounded-lg text-sm bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 disabled:opacity-40 hover:bg-gray-200 dark:hover:bg-gray-600 transition"
          @click="deckPrevPage()"
        >
          上一页
        </button>
        <span class="text-sm text-gray-500 dark:text-gray-400">
          {{ Math.floor(deckOffset / deckLimit) + 1 }} / {{ Math.ceil(deckTotal / deckLimit) }}
        </span>
        <button
          :disabled="deckOffset + deckLimit >= deckTotal"
          class="px-3 py-1.5 rounded-lg text-sm bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 disabled:opacity-40 hover:bg-gray-200 dark:hover:bg-gray-600 transition"
          @click="deckNextPage()"
        >
          下一页
        </button>
      </div>
    </div>

    <!-- ==================== 材料预览弹窗 ==================== -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="matPreview"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
          @click.self="matPreview = null"
        >
          <div
            class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl max-w-lg w-full max-h-[85vh] overflow-hidden flex flex-col"
          >
            <div class="flex items-center gap-3 p-5 border-b border-gray-100 dark:border-gray-700">
              <div
                class="w-10 h-10 rounded-xl bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center text-emerald-500"
              >
                <svg class="w-5 h-5" viewBox="0 0 20 20" fill="currentColor">
                  <path
                    d="M9 4.804A7.968 7.968 0 005.5 4c-1.255 0-2.443.29-3.5.804v10A7.969 7.969 0 015.5 14c1.669 0 3.218.51 4.5 1.385A7.962 7.962 0 0114.5 14c1.255 0 2.443.29 3.5.804v-10A7.968 7.968 0 0014.5 4c-1.255 0-2.443.29-3.5.804V12a1 1 0 11-2 0V4.804z"
                  />
                </svg>
              </div>
              <div class="flex-1 min-w-0">
                <h2 class="font-bold text-gray-800 dark:text-gray-100 truncate">{{ matPreview.title }}</h2>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ matPreview.author_name }} · {{ matPreview.card_count }} 卡片 · {{ matPreview.quiz_count }} 题目
                </p>
              </div>
              <button
                class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition"
                @click="matPreview = null"
              >
                <svg class="w-5 h-5" viewBox="0 0 20 20" fill="currentColor">
                  <path
                    fill-rule="evenodd"
                    d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                    clip-rule="evenodd"
                  />
                </svg>
              </button>
            </div>
            <div class="flex-1 overflow-y-auto p-5 custom-scroll">
              <div v-if="matPreview.tags" class="flex flex-wrap gap-1.5 mb-4">
                <span
                  v-for="t in matPreview.tags.split(',')"
                  :key="t"
                  class="px-2 py-0.5 rounded-full text-xs bg-emerald-50 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400"
                >
                  {{ t.trim() }}
                </span>
              </div>
              <div v-if="matPreview.summary" class="mb-4">
                <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                  AI 分析摘要
                </h3>
                <p class="text-sm text-gray-700 dark:text-gray-300 leading-relaxed">{{ matPreview.summary }}</p>
              </div>
              <div v-if="matPreview.key_points && matPreview.key_points.length > 0" class="mb-4">
                <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                  核心知识点
                </h3>
                <div class="space-y-1.5">
                  <div v-for="(kp, idx) in matPreview.key_points" :key="idx" class="flex items-start gap-2 text-sm">
                    <span
                      class="w-5 h-5 rounded bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 text-xs flex items-center justify-center shrink-0 mt-0.5"
                    >
                      {{ idx + 1 }}
                    </span>
                    <span class="text-gray-700 dark:text-gray-300">{{ kp }}</span>
                  </div>
                </div>
              </div>
              <div class="grid grid-cols-3 gap-3 mt-4">
                <div class="text-center p-3 rounded-lg bg-gray-50 dark:bg-gray-700/50">
                  <div class="text-lg font-bold text-gray-800 dark:text-gray-100">{{ matPreview.card_count }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">知识卡片</div>
                </div>
                <div class="text-center p-3 rounded-lg bg-gray-50 dark:bg-gray-700/50">
                  <div class="text-lg font-bold text-gray-800 dark:text-gray-100">{{ matPreview.quiz_count }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">练习题</div>
                </div>
                <div class="text-center p-3 rounded-lg bg-gray-50 dark:bg-gray-700/50">
                  <div class="text-lg font-bold text-gray-800 dark:text-gray-100">
                    {{ formatDate(matPreview.created_at) }}
                  </div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">创建日期</div>
                </div>
              </div>
            </div>
            <div class="flex items-center gap-3 p-5 border-t border-gray-100 dark:border-gray-700">
              <button
                :disabled="matCollecting"
                class="flex-1 py-2.5 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white text-sm font-medium transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                @click="handleCollectMat(matPreview)"
              >
                <svg v-if="!matCollecting" class="w-4 h-4" viewBox="0 0 20 20" fill="currentColor">
                  <path d="M5 4a2 2 0 012-2h6a2 2 0 012 2v14l-5-2.5L5 18V4z" />
                </svg>
                <svg v-else class="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" opacity="0.25" />
                  <path d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" fill="currentColor" opacity="0.75" />
                </svg>
                {{ matCollecting ? '收藏中...' : '收藏到我的库' }}
              </button>
              <button
                class="px-4 py-2.5 rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-sm hover:bg-gray-200 dark:hover:bg-gray-600 transition"
                @click="matPreview = null"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- ==================== 牌组预览弹窗 ==================== -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="deckPreview"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
          @click.self="deckPreview = null"
        >
          <div
            class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl max-w-lg w-full max-h-[85vh] overflow-hidden flex flex-col"
          >
            <div class="flex items-center gap-3 p-5 border-b border-gray-100 dark:border-gray-700">
              <div
                class="w-10 h-10 rounded-xl bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-500"
              >
                <svg class="w-5 h-5" viewBox="0 0 20 20" fill="currentColor">
                  <path
                    d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
                  />
                </svg>
              </div>
              <div class="flex-1 min-w-0">
                <h2 class="font-bold text-gray-800 dark:text-gray-100 truncate">{{ deckPreview.name }}</h2>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ deckPreview.author_name }} · {{ deckPreview.card_count }} 卡片 ·
                  {{ deckPreview.collect_count }} 收藏
                </p>
              </div>
              <button
                class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition"
                @click="deckPreview = null"
              >
                <svg class="w-5 h-5" viewBox="0 0 20 20" fill="currentColor">
                  <path
                    fill-rule="evenodd"
                    d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                    clip-rule="evenodd"
                  />
                </svg>
              </button>
            </div>
            <div class="flex-1 overflow-y-auto p-5 custom-scroll">
              <!-- 描述 -->
              <div v-if="deckPreview.description" class="mb-4">
                <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                  牌组描述
                </h3>
                <p class="text-sm text-gray-700 dark:text-gray-300 leading-relaxed">{{ deckPreview.description }}</p>
              </div>
              <!-- 标签 -->
              <div v-if="deckPreview.tags" class="flex flex-wrap gap-1.5 mb-4">
                <span
                  v-for="t in deckPreview.tags.split(',')"
                  :key="t"
                  class="px-2 py-0.5 rounded-full text-xs bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400"
                >
                  {{ t.trim() }}
                </span>
              </div>
              <!-- 卡片预览 -->
              <div v-if="deckPreview.cards && deckPreview.cards.length > 0" class="mb-4">
                <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                  卡片预览
                </h3>
                <div class="space-y-2">
                  <div
                    v-for="(card, idx) in deckPreview.cards"
                    :key="card.id"
                    class="flex items-start gap-2 p-2 rounded-lg bg-gray-50 dark:bg-gray-700/50"
                  >
                    <span
                      class="w-5 h-5 rounded bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 text-xs flex items-center justify-center shrink-0 mt-0.5"
                    >
                      {{ idx + 1 }}
                    </span>
                    <div class="flex-1 min-w-0">
                      <p class="text-sm text-gray-800 dark:text-gray-200 truncate">{{ card.concept }}</p>
                      <div class="flex gap-1 mt-1">
                        <span
                          v-if="card.difficulty"
                          :class="[
                            'px-1.5 py-0.5 rounded text-[10px]',
                            card.difficulty === 'easy'
                              ? 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400'
                              : card.difficulty === 'hard'
                                ? 'bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400'
                                : 'bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400'
                          ]"
                        >
                          {{ card.difficulty === 'easy' ? '简单' : card.difficulty === 'hard' ? '困难' : '中等' }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <!-- 统计 -->
              <div class="grid grid-cols-3 gap-3 mt-4">
                <div class="text-center p-3 rounded-lg bg-gray-50 dark:bg-gray-700/50">
                  <div class="text-lg font-bold text-gray-800 dark:text-gray-100">{{ deckPreview.card_count }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">知识卡片</div>
                </div>
                <div class="text-center p-3 rounded-lg bg-gray-50 dark:bg-gray-700/50">
                  <div class="text-lg font-bold text-gray-800 dark:text-gray-100">{{ deckPreview.collect_count }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">收藏次数</div>
                </div>
                <div class="text-center p-3 rounded-lg bg-gray-50 dark:bg-gray-700/50">
                  <div class="text-lg font-bold text-gray-800 dark:text-gray-100">
                    {{ formatDate(deckPreview.created_at) }}
                  </div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">创建日期</div>
                </div>
              </div>
            </div>
            <div class="flex items-center gap-3 p-5 border-t border-gray-100 dark:border-gray-700">
              <button
                :disabled="deckCollecting"
                class="flex-1 py-2.5 rounded-lg bg-indigo-500 hover:bg-indigo-600 text-white text-sm font-medium transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                @click="handleCollectDeck(deckPreview)"
              >
                <svg v-if="!deckCollecting" class="w-4 h-4" viewBox="0 0 20 20" fill="currentColor">
                  <path d="M5 4a2 2 0 012-2h6a2 2 0 012 2v14l-5-2.5L5 18V4z" />
                </svg>
                <svg v-else class="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" opacity="0.25" />
                  <path d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" fill="currentColor" opacity="0.75" />
                </svg>
                {{ deckCollecting ? '收藏中...' : '收藏到我的库' }}
              </button>
              <button
                class="px-4 py-2.5 rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-sm hover:bg-gray-200 dark:hover:bg-gray-600 transition"
                @click="deckPreview = null"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import {
  listMarketMaterials,
  previewMarketMaterial,
  collectMarketMaterial,
  getMarketTags,
  listMarketDecks,
  previewMarketDeck,
  collectMarketDeck,
  getMarketDeckTags
} from '../api/client.js'
import { useToast } from '../composables/useToast.js'

const toast = useToast()

const tabs = [
  { label: '学习材料', value: 'materials' },
  { label: '卡片牌组', value: 'decks' }
]
const activeTab = ref('materials')

// ==================== 材料市场状态 ====================
const materials = ref([])
const matTags = ref([])
const matLoading = ref(true)
const matSearch = ref('')
const matActiveTag = ref('')
const matSort = ref('latest')
const matLimit = 12
const matOffset = ref(0)
const matTotal = ref(0)
const matPreview = ref(null)
const matCollecting = ref(false)

const matSortOptions = [
  { label: '最新', value: 'latest' },
  { label: '最多卡片', value: 'popular' }
]

// ==================== 牌组市场状态 ====================
const decks = ref([])
const deckTags = ref([])
const deckLoading = ref(true)
const deckSearch = ref('')
const deckActiveTag = ref('')
const deckSort = ref('latest')
const deckLimit = 12
const deckOffset = ref(0)
const deckTotal = ref(0)
const deckPreview = ref(null)
const deckCollecting = ref(false)
const deckLoaded = ref(false) // 延迟加载

const deckSortOptions = [
  { label: '最新', value: 'latest' },
  { label: '最多收藏', value: 'popular' }
]

// ==================== Tab 切换 ====================
function switchTab() {
  if (activeTab.value === 'materials' && materials.value.length === 0) {
    loadMaterials()
  } else if (activeTab.value === 'decks' && !deckLoaded.value) {
    loadDecks()
    loadDeckTags()
    deckLoaded.value = true
  }
}

function handleTabClick(value) {
  activeTab.value = value
  switchTab()
}

// ==================== 材料方法 ====================
let matTimer = null
function debounceMatSearch() {
  clearTimeout(matTimer)
  matTimer = setTimeout(() => {
    matOffset.value = 0
    loadMaterials()
  }, 300)
}

async function loadMaterials() {
  matLoading.value = true
  try {
    const params = { limit: matLimit, offset: matOffset.value, sort: matSort.value }
    if (matSearch.value) params.q = matSearch.value
    if (matActiveTag.value) params.tag = matActiveTag.value
    const res = await listMarketMaterials(params)
    materials.value = res.data.data || []
    matTotal.value = res.data.total || 0
  } catch (e) {
    console.error('市场材料加载失败:', e)
    toast.error('加载材料市场失败')
  } finally {
    matLoading.value = false
  }
}

async function loadMatTags() {
  try {
    const res = await getMarketTags()
    matTags.value = res.data.tags || []
  } catch (e) {
    console.error('市场标签加载失败:', e)
  }
}

function setMatSort(value) {
  matSort.value = value
  loadMaterials()
}

function clearMatTag() {
  matActiveTag.value = ''
  loadMaterials()
}

function toggleMatTag(name) {
  matActiveTag.value = matActiveTag.value === name ? '' : name
  loadMaterials()
}

function matPrevPage() {
  matOffset.value = Math.max(0, matOffset.value - matLimit)
  loadMaterials()
}

function matNextPage() {
  matOffset.value = matOffset.value + matLimit
  loadMaterials()
}

async function openMatPreview(m) {
  try {
    const res = await previewMarketMaterial(m.share_code)
    matPreview.value = { ...res.data, share_code: m.share_code }
  } catch (e) {
    toast.error('预览加载失败')
  }
}

async function handleCollectMat(item) {
  matCollecting.value = true
  try {
    const res = await collectMarketMaterial(item.share_code)
    toast.success(`收藏成功！${res.data.card_count} 张卡片和 ${res.data.quiz_count} 道题目已加入你的材料库`)
    matPreview.value = null
  } catch (e) {
    toast.error(e.response?.data?.error || '收藏失败')
  } finally {
    matCollecting.value = false
  }
}

// ==================== 牌组方法 ====================
let deckTimer = null
function debounceDeckSearch() {
  clearTimeout(deckTimer)
  deckTimer = setTimeout(() => {
    deckOffset.value = 0
    loadDecks()
  }, 300)
}

async function loadDecks() {
  deckLoading.value = true
  try {
    const params = { limit: deckLimit, offset: deckOffset.value, sort: deckSort.value }
    if (deckSearch.value) params.q = deckSearch.value
    if (deckActiveTag.value) params.tag = deckActiveTag.value
    const res = await listMarketDecks(params)
    decks.value = res.data.data || []
    deckTotal.value = res.data.total || 0
  } catch (e) {
    console.error('市场牌组加载失败:', e)
    toast.error('加载牌组市场失败')
  } finally {
    deckLoading.value = false
  }
}

async function loadDeckTags() {
  try {
    const res = await getMarketDeckTags()
    deckTags.value = res.data.tags || []
  } catch (e) {
    console.error('牌组标签加载失败:', e)
  }
}

function setDeckSort(value) {
  deckSort.value = value
  loadDecks()
}

function clearDeckTag() {
  deckActiveTag.value = ''
  loadDecks()
}

function toggleDeckTag(name) {
  deckActiveTag.value = deckActiveTag.value === name ? '' : name
  loadDecks()
}

function deckPrevPage() {
  deckOffset.value = Math.max(0, deckOffset.value - deckLimit)
  loadDecks()
}

function deckNextPage() {
  deckOffset.value = deckOffset.value + deckLimit
  loadDecks()
}

async function openDeckPreview(d) {
  try {
    const res = await previewMarketDeck(d.share_code)
    deckPreview.value = { ...res.data, share_code: d.share_code }
  } catch (e) {
    toast.error('预览加载失败')
  }
}

async function handleCollectDeck(item) {
  deckCollecting.value = true
  try {
    const res = await collectMarketDeck(item.share_code)
    toast.success(`牌组收藏成功！${res.data.card_count} 张卡片已加入你的牌组库`)
    deckPreview.value = null
  } catch (e) {
    toast.error(e.response?.data?.error || '收藏失败')
  } finally {
    deckCollecting.value = false
  }
}

// ==================== 通用 ====================
function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diffMs = now - d
  const diffDays = Math.floor(diffMs / 86400000)
  if (diffDays === 0) return '今天'
  if (diffDays === 1) return '昨天'
  if (diffDays < 7) return `${diffDays} 天前`
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} 周前`
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

onMounted(() => {
  loadMaterials()
  loadMatTags()
})
</script>

<style scoped>
.skeleton {
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
  border-radius: 4px;
}
.dark .skeleton {
  background: linear-gradient(90deg, #374151 25%, #4b5563 50%, #374151 75%);
  background-size: 200% 100%;
}
@keyframes shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-active > div,
.modal-leave-active > div {
  transition:
    transform 0.2s ease,
    opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-from > div {
  transform: scale(0.95) translateY(10px);
  opacity: 0;
}
.modal-leave-to > div {
  transform: scale(0.95) translateY(10px);
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
  border-radius: 4px;
}
.dark .custom-scroll::-webkit-scrollbar-thumb {
  background: #4b5563;
}
</style>
