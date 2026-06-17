<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-6xl mx-auto">
    <!-- 离线缓存数据提示 -->
    <div
      v-if="!loading && fromCache"
      class="mb-3 flex items-center gap-2 px-3 py-2 rounded-lg text-xs font-medium bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-900/20 dark:text-amber-400 dark:border-amber-800/30"
    >
      <svg class="w-3.5 h-3.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
      <span>正在查看缓存数据，复习操作将排队等待联网同步</span>
    </div>

    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-6 sm:mb-8 gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">知识卡片库</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">单击放大 · 双击翻转查看详情</p>
      </div>
      <div class="flex items-center gap-2 sm:gap-3 flex-wrap">
        <!-- 搜索框 -->
        <div class="relative">
          <svg
            class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 dark:text-gray-500"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索卡片..."
            class="pl-9 pr-8 py-2 rounded-lg border border-gray-200 text-sm w-full sm:w-56 focus:outline-none focus:ring-2 focus:ring-primary-300 focus:border-primary-400 transition dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200"
          />
          <button
            v-if="searchQuery"
            class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 transition dark:text-gray-500 dark:hover:text-gray-400"
            @click="searchQuery = ''"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <!-- 待复习过滤 -->
        <button
          class="px-3 py-2 rounded-lg border text-sm transition flex items-center gap-1.5"
          :class="
            filter.due
              ? 'bg-primary-50 border-primary-300 text-primary-700 dark:bg-primary-900/20'
              : 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-gray-600 dark:text-gray-400 dark:hover:border-gray-500'
          "
          @click="filter.due = !filter.due"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <span v-if="dueCount > 0">{{ dueCount }} 待复习</span>
          <span v-else>待复习</span>
        </button>
        <!-- 仅书签过滤 -->
        <button
          class="px-3 py-2 rounded-lg border text-sm transition flex items-center gap-1.5"
          :class="
            filter.bookmarked
              ? 'bg-amber-50 border-amber-300 text-amber-700 dark:bg-amber-900/20'
              : 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-gray-600 dark:text-gray-400 dark:hover:border-gray-500'
          "
          @click="filter.bookmarked = !filter.bookmarked"
        >
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
            <path d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
          </svg>
          <span v-if="bookmarkCount > 0">{{ bookmarkCount }} 书签</span>
          <span v-else>书签</span>
        </button>
        <select
          v-model="filter.difficulty"
          class="px-3 py-2 rounded-lg border border-gray-200 text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200"
        >
          <option value="">全部难度</option>
          <option value="easy">简单</option>
          <option value="medium">中等</option>
          <option value="hard">困难</option>
        </select>
        <select
          v-model="filter.sort"
          class="px-3 py-2 rounded-lg border border-gray-200 text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200"
        >
          <option value="created_at">创建时间</option>
          <option value="ai">AI 推荐</option>
          <option value="due_date">到期时间</option>
          <option value="difficulty">难度优先</option>
        </select>
        <!-- 选择模式 toggle -->
        <button
          class="px-3 py-2 rounded-lg border text-sm transition flex items-center gap-1.5"
          :class="
            selectMode
              ? 'bg-indigo-50 border-indigo-300 text-indigo-700 dark:bg-indigo-900/20'
              : 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-gray-600 dark:text-gray-400 dark:hover:border-gray-500'
          "
          @click="toggleSelectMode"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
            />
          </svg>
          <span>{{ selectMode ? `已选 ${selectedCardIds.size}` : '选择' }}</span>
        </button>
        <!-- 创建牌组按钮 -->
        <button
          v-if="selectMode && selectedCardIds.size > 0"
          class="px-3 py-2 rounded-lg bg-indigo-500 text-white text-sm transition hover:bg-indigo-600 flex items-center gap-1.5"
          @click="openDeckModal"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          创建牌组
        </button>
        <!-- 导出 Anki CSV -->
        <button
          :disabled="cards.length === 0 || exporting"
          class="px-3 py-2 rounded-lg border border-gray-200 text-sm text-gray-600 hover:border-primary-300 hover:text-primary-600 transition flex items-center gap-1.5 disabled:opacity-40 disabled:cursor-not-allowed dark:border-gray-600 dark:text-gray-400"
          title="导出为 Anki 兼容 CSV"
          @click="exportCards"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
          <span>{{ exporting ? '导出中...' : '导出 Anki' }}</span>
        </button>
      </div>
    </div>

    <!-- 搜索结果计数 -->
    <div v-if="searchQuery && !loading" class="mb-4 text-sm text-gray-500 dark:text-gray-400">
      找到
      <span class="font-medium text-primary-600">{{ filteredCards.length }}</span>
      张匹配卡片
      <span v-if="filteredCards.length < cards.length" class="text-gray-400 dark:text-gray-500">
        （共 {{ cards.length }} 张）
      </span>
    </div>

    <!-- 空状态 -->
    <div
      v-if="cards.length === 0 && !loading"
      class="bg-white rounded-xl shadow-sm border border-gray-100 p-16 text-center dark:bg-gray-800 dark:border-gray-700"
    >
      <p class="text-5xl mb-4">🃏</p>
      <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">还没有知识卡片</h3>
      <p class="text-gray-400 dark:text-gray-500 mb-4">上传学习材料后，AI 会自动为你生成精美卡片</p>
      <router-link to="/upload" class="text-primary-600 hover:underline text-sm">去上传材料</router-link>
    </div>

    <!-- 搜索无结果 -->
    <div
      v-else-if="filteredCards.length === 0 && !loading && searchQuery"
      class="bg-white rounded-xl shadow-sm border border-gray-100 p-16 text-center dark:bg-gray-800 dark:border-gray-700"
    >
      <p class="text-5xl mb-4">🔍</p>
      <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">没有找到匹配的卡片</h3>
      <p class="text-gray-400 dark:text-gray-500 mb-4">试试换个关键词，或清空搜索条件</p>
      <button class="text-primary-600 hover:underline text-sm" @click="searchQuery = ''">清空搜索</button>
    </div>

    <!-- 无待复习卡片 -->
    <div
      v-else-if="filteredCards.length === 0 && !loading && filter.due"
      class="bg-white rounded-xl shadow-sm border border-gray-100 p-16 text-center dark:bg-gray-800 dark:border-gray-700"
    >
      <p class="text-5xl mb-4">🎉</p>
      <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">暂无待复习卡片</h3>
      <p class="text-gray-400 dark:text-gray-500 mb-4">所有卡片都已掌握或未到复习时间</p>
      <button class="text-primary-600 hover:underline text-sm" @click="filter.due = false">查看全部卡片</button>
    </div>

    <!-- 无书签卡片 -->
    <div
      v-else-if="filteredCards.length === 0 && !loading && filter.bookmarked"
      class="bg-white rounded-xl shadow-sm border border-gray-100 p-16 text-center dark:bg-gray-800 dark:border-gray-700"
    >
      <p class="text-5xl mb-4">🔖</p>
      <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">暂无书签卡片</h3>
      <p class="text-gray-400 dark:text-gray-500 mb-4">点击卡片正面的书签图标来收藏重要卡片</p>
      <button class="text-primary-600 hover:underline text-sm" @click="filter.bookmarked = false">查看全部卡片</button>
    </div>

    <!-- 骨架屏加载 -->
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <ListSkeleton :count="6" type="card" />
    </div>

    <!-- 卡片网格（虚拟滚动） -->
    <div v-else ref="cardGridRef" class="max-h-[75vh] overflow-y-auto custom-scroll">
      <!-- 虚拟滚动：顶部占位 -->
      <div v-if="cardVirtualized" :style="{ height: cardTopSpacer + 'px' }" aria-hidden="true"></div>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
        <div
          v-for="card in visibleCards"
          :key="card.id"
          class="flip-card h-80 cursor-pointer"
          :class="{ flipped: card._flipped }"
          @click="onCardClick(card)"
          @dblclick="onCardDblClick(card)"
        >
          <div class="flip-card-inner relative w-full h-full">
            <!-- 正面 -->
            <div
              class="flip-card-front absolute inset-0 bg-white rounded-xl shadow-sm border border-gray-100 p-6 flex flex-col dark:bg-gray-800 dark:border-gray-700"
            >
              <div class="flex items-start justify-between mb-3">
                <div class="flex items-center gap-2">
                  <!-- 选择模式 checkbox -->
                  <button
                    v-if="selectMode"
                    class="w-5 h-5 rounded border-2 flex items-center justify-center transition shrink-0"
                    :class="
                      selectedCardIds.has(card.id)
                        ? 'bg-indigo-500 border-indigo-500'
                        : 'border-gray-300 dark:border-gray-600'
                    "
                    @click.stop="toggleCardSelect(card.id)"
                  >
                    <svg
                      v-if="selectedCardIds.has(card.id)"
                      class="w-3 h-3 text-white"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
                    </svg>
                  </button>
                  <span class="px-2 py-0.5 rounded text-xs font-medium" :class="diffClass(card.difficulty)">
                    {{ diffLabel(card.difficulty) }}
                  </span>
                </div>
                <div class="flex items-center gap-1.5">
                  <!-- 书签按钮 -->
                  <button
                    class="p-1 rounded transition hover:scale-110"
                    :class="
                      card.is_bookmarked
                        ? 'text-amber-500'
                        : 'text-gray-300 hover:text-amber-400 dark:text-gray-600 dark:hover:text-amber-500'
                    "
                    :title="card.is_bookmarked ? '移除书签' : '添加书签'"
                    @click.stop="toggleBookmarkCard(card)"
                  >
                    <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                    </svg>
                  </button>
                  <!-- 复习状态指示 -->
                  <span
                    v-if="card.review_count > 0"
                    class="text-xs px-1.5 py-0.5 bg-emerald-50 text-emerald-600 rounded dark:bg-emerald-900/30"
                  >
                    复习 {{ card.review_count }}次
                  </span>
                  <span v-else class="text-xs px-1.5 py-0.5 bg-blue-50 text-blue-500 rounded dark:bg-blue-900/30">
                    新卡片
                  </span>
                </div>
              </div>
              <h3
                class="text-lg font-bold text-gray-900 dark:text-gray-100 mb-2 line-clamp-2"
                v-html="highlight(card.concept)"
              ></h3>
              <p
                class="text-sm text-gray-500 dark:text-gray-400 flex-1 line-clamp-3"
                v-html="highlight(card.detail)"
              ></p>
              <div class="mt-3 flex flex-wrap gap-1">
                <span
                  v-for="tag in card.tags || []"
                  :key="tag"
                  class="px-2 py-0.5 rounded text-xs"
                  :class="
                    isTagMatch(tag)
                      ? 'bg-primary-100 text-primary-700 font-medium'
                      : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'
                  "
                  v-html="highlight(tag)"
                ></span>
                <!-- 材料标签 -->
                <template v-if="materialTagMap[card.material_id]?.length">
                  <span class="w-px h-3 bg-gray-200 dark:bg-gray-600 mx-0.5 self-center"></span>
                  <span
                    v-for="mt in materialTagMap[card.material_id]"
                    :key="'m-' + mt"
                    class="px-1.5 py-0.5 rounded text-[10px] bg-indigo-50 text-indigo-500 dark:bg-indigo-900/20 dark:text-indigo-400"
                  >
                    {{ mt }}
                  </span>
                </template>
              </div>
              <p class="text-xs text-gray-300 dark:text-gray-600 mt-2 text-center">单击放大 · 双击翻转</p>
            </div>

            <!-- 背面 -->
            <div
              class="flip-card-back absolute inset-0 bg-gradient-to-br from-primary-50 to-indigo-50 rounded-xl shadow-sm border border-primary-100 p-6 flex flex-col dark:from-gray-800 dark:to-gray-700 dark:border-gray-600"
            >
              <h3 class="text-base font-bold text-primary-700 mb-2" v-html="highlight(card.concept)"></h3>
              <p
                class="text-sm text-gray-700 dark:text-gray-300 flex-1 overflow-y-auto"
                v-html="highlight(card.detail)"
              ></p>
              <div
                v-if="card.formula"
                class="mt-2 p-2.5 bg-white rounded-lg border border-primary-100 dark:bg-gray-700 dark:border-gray-600"
              >
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">关键公式/代码</p>
                <code class="text-sm text-primary-600" v-html="renderMath(card.formula)"></code>
              </div>
              <div
                v-if="card.memory_tip"
                class="mt-2 p-2.5 bg-amber-50 rounded-lg border border-amber-100 dark:bg-amber-900/30 dark:border-amber-800"
              >
                <p class="text-xs text-amber-600 mb-1">💡 记忆技巧</p>
                <p class="text-sm text-amber-800 dark:text-amber-300" v-html="renderMath(card.memory_tip)"></p>
              </div>

              <!-- 个人笔记 -->
              <div class="mt-2" @click.stop>
                <div
                  class="flex items-center gap-1 mb-1 cursor-pointer"
                  @click="card._noteEditing = !card._noteEditing"
                >
                  <svg
                    class="w-3.5 h-3.5 text-gray-400 dark:text-gray-500"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                    />
                  </svg>
                  <span class="text-xs text-gray-400 dark:text-gray-500">笔记</span>
                </div>
                <textarea
                  v-if="card._noteEditing || card.user_note"
                  v-model="card.user_note"
                  placeholder="添加个人笔记... (可粘贴或拖拽图片)"
                  class="w-full text-xs p-2 rounded-lg border text-gray-700 resize-none focus:outline-none focus:ring-1 focus:ring-primary-300 transition"
                  :class="
                    dragOverNote === card.id
                      ? 'border-primary-400 bg-primary-50 dark:border-primary-500 dark:bg-primary-900/20'
                      : 'border-gray-200 bg-white dark:bg-gray-700 dark:border-gray-600'
                  "
                  rows="2"
                  @input="saveCardNote(card)"
                  @click.stop
                  @paste="handlePasteImage($event, card)"
                  @drop="handleDropImage($event, card)"
                  @dragover.prevent="dragOverNote = card.id"
                  @dragleave="dragOverNote = null"
                ></textarea>
              </div>

              <!-- 复习统计 + 按钮 -->
              <div class="mt-3 pt-3 border-t border-primary-100 dark:border-gray-600">
                <div class="flex items-center justify-between text-xs text-gray-400 dark:text-gray-500 mb-2">
                  <span>间隔 {{ card.interval_days || 0 }} 天 · 系数 {{ (card.ease_factor || 2.5).toFixed(1) }}</span>
                  <span v-if="card.next_review_at">下次: {{ formatDate(card.next_review_at) }}</span>
                  <span v-else>尚未复习</span>
                </div>
                <div class="flex gap-2" @click.stop>
                  <button
                    :disabled="card._reviewing"
                    class="flex-1 py-2 rounded-lg text-sm font-medium transition border bg-orange-50 border-orange-200 text-orange-700 hover:bg-orange-100 disabled:opacity-50 disabled:cursor-not-allowed"
                    @click="submitReview(card, 'review')"
                  >
                    🔄 再复习
                  </button>
                  <button
                    :disabled="card._reviewing"
                    class="flex-1 py-2 rounded-lg text-sm font-medium transition border bg-emerald-50 border-emerald-200 text-emerald-700 hover:bg-emerald-100 disabled:opacity-50 disabled:cursor-not-allowed"
                    @click="submitReview(card, 'mastered')"
                  >
                    ✅ 已掌握
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
        <!-- 虚拟滚动：底部占位 -->
        <div v-if="cardVirtualized" :style="{ height: cardBottomSpacer + 'px' }" aria-hidden="true"></div>
        <div v-if="cardVirtualized" class="text-center py-2 text-xs text-gray-400 dark:text-gray-500">
          已启用虚拟滚动 · 显示 {{ cardEndIdx - cardStartIdx }} / {{ filteredCards.length }} 张卡片
        </div>
        <!-- 无限滚动：哨兵元素 -->
        <div ref="cardSentinelRef" class="h-1" aria-hidden="true"></div>
        <!-- 无限滚动：底部状态 -->
        <InfiniteScrollFooter
          :loading="cardLoadingMore"
          :has-more="cardHasMore"
          :error="cardLoadError"
          :total-count="cards.length"
          @retry="loadMoreCards"
        />
      </div>
      <ScrollToTop :show="showCardScrollTop" @click="scrollToCardTop" />

      <!-- Enlarged card modal -->
      <Teleport to="body">
        <Transition name="modal-fade">
          <div
            v-if="enlargedCard"
            class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
            @click.self="closeEnlarged"
          >
            <!-- Close button -->
            <button
              class="absolute top-4 right-4 z-10 p-2 rounded-full bg-white/20 hover:bg-white/40 text-white transition"
              @click="closeEnlarged"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>

            <!-- Modal flip card -->
            <div
              class="flip-card w-full max-w-lg cursor-pointer"
              style="height: 520px"
              :class="{ flipped: modalFlipped }"
              @click="modalFlipped = !modalFlipped"
            >
              <div class="flip-card-inner relative w-full h-full">
                <!-- Front -->
                <div
                  class="flip-card-front absolute inset-0 bg-white rounded-2xl shadow-2xl border border-gray-100 p-8 flex flex-col dark:bg-gray-800 dark:border-gray-700 overflow-y-auto"
                >
                  <div class="flex items-start justify-between mb-4">
                    <span
                      class="px-3 py-1 rounded-lg text-sm font-semibold"
                      :class="diffClass(enlargedCard.difficulty)"
                    >
                      {{ diffLabel(enlargedCard.difficulty) }}
                    </span>
                    <div class="flex items-center gap-2">
                      <!-- 书签按钮 -->
                      <button
                        class="p-1.5 rounded-lg transition hover:scale-110"
                        :class="
                          enlargedCard.is_bookmarked
                            ? 'text-amber-500'
                            : 'text-gray-300 hover:text-amber-400 dark:text-gray-600 dark:hover:text-amber-500'
                        "
                        :title="enlargedCard.is_bookmarked ? '移除书签' : '添加书签'"
                        @click.stop="toggleBookmarkCard(enlargedCard)"
                      >
                        <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                          <path d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                        </svg>
                      </button>
                      <span
                        v-if="enlargedCard.review_count > 0"
                        class="text-xs px-2 py-1 bg-emerald-50 text-emerald-600 rounded-lg dark:bg-emerald-900/30"
                      >
                        复习 {{ enlargedCard.review_count }}次
                      </span>
                      <span v-else class="text-xs px-2 py-1 bg-blue-50 text-blue-500 rounded-lg dark:bg-blue-900/30">
                        新卡片
                      </span>
                    </div>
                  </div>
                  <h2
                    class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4 leading-relaxed"
                    v-html="renderMath(enlargedCard.concept)"
                  ></h2>
                  <p
                    class="text-base text-gray-600 dark:text-gray-300 flex-1 leading-relaxed whitespace-pre-wrap"
                    v-html="renderMath(enlargedCard.detail)"
                  ></p>
                  <div v-if="enlargedCard.tags && enlargedCard.tags.length" class="flex flex-wrap gap-1.5 mt-4">
                    <span
                      v-for="tag in enlargedCard.tags"
                      :key="tag"
                      class="px-2.5 py-1 rounded-lg text-xs bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400"
                    >
                      {{ tag }}
                    </span>
                  </div>
                  <p class="text-xs text-gray-300 dark:text-gray-600 mt-4 text-center">点击卡片翻转查看详情</p>
                </div>
                <!-- Back -->
                <div
                  class="flip-card-back absolute inset-0 bg-gradient-to-br from-primary-50 to-indigo-50 rounded-2xl shadow-2xl border border-primary-100 p-8 flex flex-col dark:from-gray-800 dark:to-gray-700 dark:border-gray-600 overflow-y-auto"
                >
                  <h2
                    class="text-xl font-bold text-primary-700 dark:text-primary-400 mb-4"
                    v-html="renderMath(enlargedCard.concept)"
                  ></h2>
                  <p
                    class="text-base text-gray-700 dark:text-gray-300 flex-1 leading-relaxed whitespace-pre-wrap"
                    v-html="renderMath(enlargedCard.detail)"
                  ></p>
                  <div
                    v-if="enlargedCard.formula"
                    class="mt-4 p-4 bg-white rounded-xl border border-primary-100 dark:bg-gray-700 dark:border-gray-600"
                  >
                    <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">关键公式/代码</p>
                    <code
                      class="text-sm text-primary-600 dark:text-primary-400 whitespace-pre-wrap break-all"
                      v-html="renderMath(enlargedCard.formula)"
                    ></code>
                  </div>
                  <div
                    v-if="enlargedCard.memory_tip"
                    class="mt-4 p-4 bg-amber-50 rounded-xl border border-amber-100 dark:bg-amber-900/30 dark:border-amber-800"
                  >
                    <p class="text-xs text-amber-600 dark:text-amber-400 mb-1">记忆技巧</p>
                    <p
                      class="text-sm text-amber-800 dark:text-amber-300 leading-relaxed"
                      v-html="renderMath(enlargedCard.memory_tip)"
                    ></p>
                  </div>
                  <!-- 个人笔记 -->
                  <div class="mt-4" @click.stop>
                    <div class="flex items-center gap-1.5 mb-2">
                      <svg
                        class="w-4 h-4 text-gray-400 dark:text-gray-500"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                        />
                      </svg>
                      <span class="text-sm text-gray-500 dark:text-gray-400 font-medium">个人笔记</span>
                    </div>
                    <textarea
                      v-model="enlargedCard.user_note"
                      placeholder="添加你的学习笔记... (可粘贴或拖拽图片)"
                      class="w-full text-sm p-3 rounded-xl border bg-white text-gray-700 resize-none focus:outline-none focus:ring-2 focus:ring-primary-300 transition"
                      :class="
                        dragOverNote === 'modal'
                          ? 'border-primary-400 bg-primary-50 dark:border-primary-500 dark:bg-primary-900/20'
                          : 'border-gray-200 dark:bg-gray-700 dark:border-gray-600'
                      "
                      rows="3"
                      @input="saveCardNote(enlargedCard)"
                      @click.stop
                      @paste="handlePasteImage($event, enlargedCard)"
                      @drop="handleDropImage($event, enlargedCard)"
                      @dragover.prevent="dragOverNote = 'modal'"
                      @dragleave="dragOverNote = null"
                    ></textarea>
                  </div>
                  <!-- Review stats + buttons -->
                  <div class="mt-4 pt-4 border-t border-primary-100 dark:border-gray-600" @click.stop>
                    <div class="flex items-center justify-between text-xs text-gray-400 dark:text-gray-500 mb-3">
                      <span>
                        间隔 {{ enlargedCard.interval_days || 0 }} 天 · 系数
                        {{ (enlargedCard.ease_factor || 2.5).toFixed(1) }}
                      </span>
                      <span v-if="enlargedCard.next_review_at">
                        下次: {{ formatDate(enlargedCard.next_review_at) }}
                      </span>
                      <span v-else>尚未复习</span>
                    </div>
                    <div class="flex gap-3">
                      <button
                        :disabled="enlargedCard._reviewing"
                        class="flex-1 py-2.5 rounded-lg text-sm font-medium transition border bg-orange-50 border-orange-200 text-orange-700 hover:bg-orange-100 disabled:opacity-50 disabled:cursor-not-allowed"
                        @click="submitReview(enlargedCard, 'review')"
                      >
                        再复习
                      </button>
                      <button
                        :disabled="enlargedCard._reviewing"
                        class="flex-1 py-2.5 rounded-lg text-sm font-medium transition border bg-emerald-50 border-emerald-200 text-emerald-700 hover:bg-emerald-100 disabled:opacity-50 disabled:cursor-not-allowed"
                        @click="submitReview(enlargedCard, 'mastered')"
                      >
                        已掌握
                      </button>
                    </div>
                  </div>
                  <p class="text-xs text-gray-300 dark:text-gray-600 mt-3 text-center">点击卡片翻转回正面</p>
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </Teleport>

      <!-- ==================== 我的牌组 ==================== -->
      <div v-if="myDecks.length > 0" class="mt-10 pt-8 border-t border-gray-200 dark:border-gray-700">
        <h2 class="text-lg font-bold text-gray-800 dark:text-gray-100 flex items-center gap-2 mb-4">
          <svg class="w-5 h-5 text-indigo-500" fill="currentColor" viewBox="0 0 20 20">
            <path
              d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
            />
          </svg>
          我的牌组
          <span class="text-sm font-normal text-gray-400 dark:text-gray-500">({{ myDecks.length }})</span>
        </h2>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          <div
            v-for="deck in myDecks"
            :key="deck.id"
            class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 hover:border-indigo-300 dark:hover:border-indigo-600 transition"
          >
            <div class="flex items-start justify-between mb-2">
              <h3 class="font-semibold text-gray-800 dark:text-gray-100 text-sm line-clamp-1 flex-1">
                {{ deck.name }}
              </h3>
              <div class="flex items-center gap-1 shrink-0 ml-2">
                <button
                  class="p-1.5 rounded-lg transition text-xs"
                  :class="
                    deck.is_public
                      ? 'text-green-500 bg-green-50 dark:bg-green-900/20'
                      : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
                  "
                  :title="deck.is_public ? '取消分享' : '分享牌组'"
                  @click="handleToggleDeckShare(deck)"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"
                    />
                  </svg>
                </button>
                <button
                  class="p-1.5 rounded-lg text-gray-400 hover:text-red-500 transition"
                  title="删除牌组"
                  @click="handleDeleteDeck(deck)"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
            <p v-if="deck.description" class="text-xs text-gray-500 dark:text-gray-400 line-clamp-1 mb-2">
              {{ deck.description }}
            </p>
            <div class="flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500">
              <span>{{ deck.card_count }} 卡片</span>
              <span>{{ deck.collect_count || 0 }} 收藏</span>
              <span v-if="deck.is_public" class="flex items-center gap-0.5 text-green-500">
                <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fill-rule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clip-rule="evenodd"
                  />
                </svg>
                已公开
              </span>
            </div>
            <div v-if="deck.tags" class="flex flex-wrap gap-1 mt-2">
              <span
                v-for="t in deck.tags.split(',').slice(0, 3)"
                :key="t"
                class="px-1.5 py-0.5 rounded text-[10px] bg-indigo-50 dark:bg-indigo-900/20 text-indigo-500 dark:text-indigo-400"
              >
                {{ t.trim() }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- ==================== 创建牌组弹窗 ==================== -->
      <Teleport to="body">
        <Transition name="modal-fade">
          <div
            v-if="showDeckModal"
            class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
            @click.self="showDeckModal = false"
          >
            <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl max-w-md w-full p-6">
              <h2 class="text-lg font-bold text-gray-800 dark:text-gray-100 mb-4 flex items-center gap-2">
                <svg class="w-5 h-5 text-indigo-500" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"
                  />
                </svg>
                创建牌组
                <span class="text-sm font-normal text-gray-400">({{ selectedCardIds.size }} 张卡片)</span>
              </h2>
              <div class="space-y-3">
                <div>
                  <label class="block text-sm text-gray-600 dark:text-gray-400 mb-1">牌组名称 *</label>
                  <input
                    v-model="deckForm.name"
                    type="text"
                    placeholder="如：Go 并发编程核心"
                    class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm dark:text-gray-200 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
                  />
                </div>
                <div>
                  <label class="block text-sm text-gray-600 dark:text-gray-400 mb-1">描述</label>
                  <textarea
                    v-model="deckForm.description"
                    placeholder="简要描述这个牌组的内容..."
                    rows="2"
                    class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm dark:text-gray-200 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none resize-none"
                  ></textarea>
                </div>
                <div>
                  <label class="block text-sm text-gray-600 dark:text-gray-400 mb-1">标签 (逗号分隔)</label>
                  <input
                    v-model="deckForm.tags"
                    type="text"
                    placeholder="如：Go,并发,goroutine"
                    class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm dark:text-gray-200 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
                  />
                </div>
              </div>
              <div class="flex gap-3 mt-5">
                <button
                  :disabled="deckCreating || !deckForm.name.trim()"
                  class="flex-1 py-2.5 rounded-lg bg-indigo-500 hover:bg-indigo-600 text-white text-sm font-medium transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                  @click="handleCreateDeck"
                >
                  <svg v-if="deckCreating" class="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none">
                    <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" opacity="0.25" />
                    <path d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" fill="currentColor" opacity="0.75" />
                  </svg>
                  {{ deckCreating ? '创建中...' : '创建牌组' }}
                </button>
                <button
                  class="px-4 py-2.5 rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-sm hover:bg-gray-200 dark:hover:bg-gray-600 transition"
                  @click="showDeckModal = false"
                >
                  取消
                </button>
              </div>
            </div>
          </div>
        </Transition>
      </Teleport>
    </div>

    <!-- AI 解释浮动按钮（选中文字时出现） -->
    <Teleport to="body">
      <button
        v-if="cardButtonPos.visible"
        class="explain-float-btn fixed z-50 flex items-center gap-1.5 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-medium rounded-lg shadow-lg transition-all"
        :style="{ left: cardButtonPos.x + 'px', top: cardButtonPos.y + 'px' }"
        @click="cardExplainSelection"
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
          v-if="cardExplainDrawer.visible"
          class="explain-drawer fixed inset-0 z-50 flex justify-end"
          @click.self="cardCloseExplainDrawer"
        >
          <div class="absolute inset-0 bg-black/30 dark:bg-black/50" @click="cardCloseExplainDrawer"></div>
          <div class="relative w-full max-w-md h-full bg-white dark:bg-gray-800 shadow-2xl overflow-y-auto">
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
                @click="cardCloseExplainDrawer"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div v-if="cardExplainDrawer.loading" class="flex flex-col items-center justify-center py-20 px-5">
              <div
                class="w-10 h-10 border-3 border-indigo-200 border-t-indigo-600 rounded-full animate-spin dark:border-indigo-800 dark:border-t-indigo-400"
              ></div>
              <p class="mt-4 text-sm text-gray-500 dark:text-gray-400">AI 正在分析概念...</p>
            </div>
            <div v-else-if="cardExplainDrawer.error" class="px-5 py-10 text-center">
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
              <p class="text-sm text-red-600 dark:text-red-400">{{ cardExplainDrawer.error }}</p>
            </div>
            <div v-else-if="cardExplainDrawer.result" class="px-5 py-5 space-y-5">
              <div class="flex items-center gap-2">
                <span
                  class="px-2.5 py-1 bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300 text-xs font-medium rounded-full"
                >
                  概念
                </span>
                <h4 class="text-base font-semibold text-gray-900 dark:text-gray-100">
                  {{ cardExplainDrawer.result.concept }}
                </h4>
                <span
                  v-if="cardExplainDrawer.result.cached"
                  class="ml-auto text-[10px] text-gray-400 dark:text-gray-500 bg-gray-100 dark:bg-gray-700 px-1.5 py-0.5 rounded"
                >
                  缓存
                </span>
              </div>
              <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-700/50 border border-gray-100 dark:border-gray-600">
                <h5 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                  通俗解释
                </h5>
                <p class="text-sm text-gray-800 dark:text-gray-200 leading-relaxed whitespace-pre-wrap">
                  {{ cardExplainDrawer.result.explanation }}
                </p>
              </div>
              <div
                v-if="cardExplainDrawer.result.analogy"
                class="p-4 rounded-xl bg-amber-50 dark:bg-amber-900/10 border border-amber-100 dark:border-amber-800/30"
              >
                <h5 class="text-xs font-semibold text-amber-600 dark:text-amber-400 uppercase tracking-wider mb-2">
                  生活类比
                </h5>
                <p class="text-sm text-gray-800 dark:text-gray-200 leading-relaxed whitespace-pre-wrap">
                  {{ cardExplainDrawer.result.analogy }}
                </p>
              </div>
              <div
                v-if="cardExplainDrawer.result.example"
                class="p-4 rounded-xl bg-emerald-50 dark:bg-emerald-900/10 border border-emerald-100 dark:border-emerald-800/30"
              >
                <h5 class="text-xs font-semibold text-emerald-600 dark:text-emerald-400 uppercase tracking-wider mb-2">
                  具体例子
                </h5>
                <p class="text-sm text-gray-800 dark:text-gray-200 leading-relaxed whitespace-pre-wrap">
                  {{ cardExplainDrawer.result.example }}
                </p>
              </div>
              <div
                v-if="cardExplainDrawer.result.related_concepts?.length"
                class="p-4 rounded-xl bg-sky-50 dark:bg-sky-900/10 border border-sky-100 dark:border-sky-800/30"
              >
                <h5 class="text-xs font-semibold text-sky-600 dark:text-sky-400 uppercase tracking-wider mb-2">
                  关联知识点
                </h5>
                <div class="flex flex-wrap gap-2">
                  <span
                    v-for="concept in cardExplainDrawer.result.related_concepts"
                    :key="concept"
                    class="px-2.5 py-1 bg-sky-100 dark:bg-sky-800/30 text-sky-700 dark:text-sky-300 text-xs rounded-full font-medium cursor-pointer hover:bg-sky-200 dark:hover:bg-sky-700/40 transition"
                    @click="cardRequestExplain(concept)"
                  >
                    {{ concept }}
                  </span>
                </div>
              </div>
              <div class="flex gap-2 pt-2">
                <button
                  class="flex-1 px-3 py-2 text-xs font-medium bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition"
                  @click="cardRequestExplain(cardExplainDrawer.result.concept)"
                >
                  重新生成
                </button>
                <button
                  class="flex-1 px-3 py-2 text-xs font-medium bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition"
                  @click="cardCloseExplainDrawer"
                >
                  关闭
                </button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import {
  listCards,
  reviewCard,
  listMaterials,
  toggleBookmark,
  updateCardNote,
  uploadImage,
  listDecks,
  createDeck,
  deleteDeck,
  toggleDeckShare
} from '../api/client'
import { useToast } from '../composables/useToast'
import { renderMath, renderMathWithHighlight } from '../composables/useMathRender'
import { useExplain } from '../composables/useExplain'
import { useGridVirtualScroll } from '../composables/useVirtualScroll'
import { useInfiniteScroll, useScrollToTop } from '../composables/useInfiniteScroll'
import { useNetworkStatus } from '../composables/useNetworkStatus'
import { useOfflineQueue } from '../composables/useOfflineQueue'
import ListSkeleton from '../components/ListSkeleton.vue'
import InfiniteScrollFooter from '../components/InfiniteScrollFooter.vue'
import ScrollToTop from '../components/ScrollToTop.vue'

const { success: toastSuccess } = useToast()
const { isOnline, isResponseFromCache: _isResponseFromCache } = useNetworkStatus()
const { enqueue, ACTION_TYPES } = useOfflineQueue()

// AI 概念解释器
const {
  selectedText: _cardSelectedText,
  buttonPos: cardButtonPos,
  drawer: cardExplainDrawer,
  hideButton: _cardHideButton,
  initExplainListener: cardInitExplain,
  cleanupExplainListener: cardCleanupExplain,
  requestExplain: cardRequestExplain,
  explainSelection: cardExplainSelection,
  closeDrawer: cardCloseExplainDrawer
} = useExplain()

const cards = ref([])
const loading = ref(true)
const fromCache = ref(false)
const searchQuery = ref('')
const exporting = ref(false)
const filter = reactive({ difficulty: '', due: false, bookmarked: false, sort: 'created_at' })

// 虚拟滚动：卡片网格 >200 张时启用
const cardGridRef = ref(null)
const gridCols = ref(window.innerWidth >= 1024 ? 3 : window.innerWidth >= 768 ? 2 : 1)

// 监听窗口宽度变化更新列数
function updateGridCols() {
  gridCols.value = window.innerWidth >= 1024 ? 3 : window.innerWidth >= 768 ? 2 : 1
}

let gridResizeHandler = null

const {
  startIndex: cardStartIdx,
  endIndex: cardEndIdx,
  topSpacerHeight: cardTopSpacer,
  bottomSpacerHeight: cardBottomSpacer,
  shouldVirtualize: cardVirtualized
} = useGridVirtualScroll(
  cardGridRef,
  computed(() => filteredCards.value.length),
  {
    rowHeight: 340,
    cols: gridCols,
    buffer: 3,
    threshold: 200
  }
)

// 无限滚动
const cardSentinelRef = ref(null)
const {
  loading: cardLoadingMore,
  hasMore: cardHasMore,
  error: cardLoadError,
  total: _cardTotal,
  loadNext: cardLoadNext,
  reset: cardResetScroll,
  retry: _cardRetryLoad,
  init: cardInitInfiniteScroll
} = useInfiniteScroll({
  limit: 20,
  rootMargin: '300px',
  onLoad: async (offset, limit) => {
    const params = { limit, offset }
    if (filter.difficulty) params.difficulty = filter.difficulty
    const res = await listCards(params)
    const items = (res.data.data || []).map((c) => ({
      ...c,
      tags: typeof c.tags === 'string' && c.tags ? c.tags.split(',').map((t) => t.trim()) : [],
      _flipped: false,
      _reviewing: false,
      _noteEditing: false
    }))
    return { items, total: res.data.total || 0 }
  },
  onItems: (items) => {
    cards.value.push(...items)
  }
})

// 返回顶部
const {
  showButton: showCardScrollTop,
  scrollToTop: scrollToCardTop,
  initScrollListener: initCardScrollListener,
  destroyScrollListener: destroyCardScrollListener
} = useScrollToTop(cardGridRef)

// Enlarged modal state
const enlargedCard = ref(null)
const modalFlipped = ref(false)
let cardClickTimer = null

// 材料标签映射 { material_id: ['tag1', 'tag2'] }
const materialTagMap = ref({})

// ==================== 牌组功能 ====================
const selectMode = ref(false)
const selectedCardIds = ref(new Set())
const showDeckModal = ref(false)
const myDecks = ref([])
const deckForm = reactive({ name: '', description: '', tags: '' })
const deckCreating = ref(false)
const toast = useToast()

function toggleSelectMode() {
  selectMode.value = !selectMode.value
  if (!selectMode.value) {
    selectedCardIds.value = new Set()
  }
}

function toggleCardSelect(cardId) {
  const newSet = new Set(selectedCardIds.value)
  if (newSet.has(cardId)) {
    newSet.delete(cardId)
  } else {
    newSet.add(cardId)
  }
  selectedCardIds.value = newSet
}

function openDeckModal() {
  if (selectedCardIds.value.size === 0) {
    toast.error('请先选择至少一张卡片')
    return
  }
  deckForm.name = ''
  deckForm.description = ''
  deckForm.tags = ''
  showDeckModal.value = true
}

async function handleCreateDeck() {
  if (!deckForm.name.trim()) {
    toast.error('请输入牌组名称')
    return
  }
  deckCreating.value = true
  try {
    await createDeck({
      name: deckForm.name.trim(),
      description: deckForm.description.trim(),
      tags: deckForm.tags.trim(),
      card_ids: Array.from(selectedCardIds.value)
    })
    toast.success(`牌组「${deckForm.name}」创建成功，包含 ${selectedCardIds.value.size} 张卡片`)
    showDeckModal.value = false
    selectMode.value = false
    selectedCardIds.value = new Set()
    loadMyDecks()
  } catch (e) {
    toast.error(e.response?.data?.error || '创建牌组失败')
  } finally {
    deckCreating.value = false
  }
}

async function loadMyDecks() {
  try {
    const res = await listDecks()
    myDecks.value = res.data.decks || []
  } catch (e) {
    console.error('牌组加载失败:', e)
  }
}

async function handleDeleteDeck(deck) {
  try {
    await deleteDeck(deck.id)
    myDecks.value = myDecks.value.filter((d) => d.id !== deck.id)
    toast.success('牌组已删除')
  } catch (e) {
    toast.error('删除失败')
  }
}

async function handleToggleDeckShare(deck) {
  try {
    const res = await toggleDeckShare(deck.id)
    deck.is_public = res.data.is_public
    deck.share_code = res.data.share_code
    toast.success(res.data.message)
  } catch (e) {
    toast.error('操作失败')
  }
}

function onCardClick(card) {
  // Delay single click to distinguish from dblclick
  if (cardClickTimer) clearTimeout(cardClickTimer)
  cardClickTimer = setTimeout(() => {
    enlargedCard.value = card
    modalFlipped.value = false
  }, 250)
}

function onCardDblClick(card) {
  // Cancel the pending single-click enlarge
  if (cardClickTimer) {
    clearTimeout(cardClickTimer)
    cardClickTimer = null
  }
  card._flipped = !card._flipped
}

function closeEnlarged() {
  enlargedCard.value = null
  modalFlipped.value = false
}

// 待复习计数
const dueCount = computed(() => {
  const now = new Date()
  return cards.value.filter((c) => {
    if (!c.next_review_at) return true // 新卡片也算
    return new Date(c.next_review_at) <= now
  }).length
})

// 书签计数
const bookmarkCount = computed(() => {
  return cards.value.filter((c) => c.is_bookmarked).length
})

// 模糊过滤：按概念名、标签、详情匹配
const filteredCards = computed(() => {
  let result = cards.value
  // 待复习过滤
  if (filter.due) {
    const now = new Date()
    result = result.filter((c) => {
      if (!c.next_review_at) return true
      return new Date(c.next_review_at) <= now
    })
  }
  // 书签过滤
  if (filter.bookmarked) {
    result = result.filter((c) => c.is_bookmarked)
  }
  // 搜索过滤
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    result = result.filter((card) => {
      if ((card.concept || '').toLowerCase().includes(q)) return true
      if ((card.detail || '').toLowerCase().includes(q)) return true
      if ((card.tags || []).some((tag) => tag.toLowerCase().includes(q))) return true
      if ((card.formula || '').toLowerCase().includes(q)) return true
      if ((card.memory_tip || '').toLowerCase().includes(q)) return true
      return false
    })
  }
  // 排序
  result = [...result] // 避免原地排序
  if (filter.sort === 'ai') {
    result.sort((a, b) => cardPriorityScore(b) - cardPriorityScore(a))
  } else if (filter.sort === 'due_date') {
    result.sort((a, b) => {
      if (!a.next_review_at && b.next_review_at) return -1
      if (a.next_review_at && !b.next_review_at) return 1
      if (a.next_review_at && b.next_review_at) return new Date(a.next_review_at) - new Date(b.next_review_at)
      return 0
    })
  } else if (filter.sort === 'difficulty') {
    result.sort((a, b) => (a.ease_factor || 2.5) - (b.ease_factor || 2.5))
  }
  // created_at (default) — already sorted by server
  return result
})

// 虚拟滚动可见卡片切片
const visibleCards = computed(() => {
  if (!cardVirtualized.value) return filteredCards.value
  return filteredCards.value.slice(cardStartIdx.value, cardEndIdx.value)
})

// 客户端优先级评分（4维，与 Review.vue 服务端评分对齐）
function cardPriorityScore(card) {
  const now = new Date()
  let score = 0
  // 过期分数 (0-30)
  if (!card.next_review_at) {
    score += 15
  } else {
    const overdueDays = Math.max(0, (now - new Date(card.next_review_at)) / (1000 * 60 * 60 * 24))
    score += Math.min(30, overdueDays * 3)
  }
  // 难度分数 (0-25)
  score += Math.max(0, Math.min(25, (2.5 - (card.ease_factor || 2.5)) * 20))
  // 新鲜度分数 (0-15)
  score += Math.max(0, 15 - (card.review_count || 0) * 2)
  // 书签加分 (0-10)
  if (card.is_bookmarked) score += 10
  return score
}

// 搜索关键词高亮 + LaTeX 数学公式渲染（先提取数学公式，对纯文本部分做转义+高亮，再还原 KaTeX）
function highlight(text) {
  const q = searchQuery.value.trim()
  const escapeAndHighlight = (t) => {
    if (!q) return escapeHtml(t)
    const escaped = escapeHtml(t)
    const escapedQ = escapeHtml(q)
    const regex = new RegExp(`(${escapeRegex(escapedQ)})`, 'gi')
    return escaped.replace(
      regex,
      '<mark class="bg-yellow-200 dark:bg-yellow-700/40 text-yellow-900 dark:text-yellow-200 rounded px-0.5">$1</mark>'
    )
  }
  return renderMathWithHighlight(text, escapeAndHighlight)
}

function isTagMatch(tag) {
  const q = searchQuery.value.trim().toLowerCase()
  return q && tag.toLowerCase().includes(q)
}

function escapeHtml(str) {
  const div = document.createElement('div')
  div.textContent = str
  return div.innerHTML
}

function escapeRegex(str) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
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
  return { easy: '简单', medium: '中等', hard: '困难' }[d] || '未分级'
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diffMs = d - now
  const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24))
  if (diffDays <= 0) return '今天'
  if (diffDays === 1) return '明天'
  if (diffDays <= 7) return `${diffDays} 天后`
  return `${d.getMonth() + 1}/${d.getDate()}`
}

// 提交复习结果
async function submitReview(card, result) {
  card._reviewing = true
  try {
    // 离线时排队操作
    if (!isOnline.value) {
      await enqueue(
        ACTION_TYPES.CARD_REVIEW,
        'POST',
        `/cards/${card.id}/review`,
        { result },
        `复习「${card.concept.slice(0, 15)}」(${result === 'mastered' ? '已掌握' : '再复习'})`
      )
      toastSuccess(`离线模式: 复习记录已排队，联网后自动同步`)
      return
    }

    const res = await reviewCard(card.id, result)
    const data = res.data
    // 更新卡片本地状态
    card.review_count = data.review_count
    card.interval_days = data.interval_days
    card.ease_factor = data.ease_factor
    card.next_review_at = data.next_review_at
    card.last_reviewed_at = new Date().toISOString()

    if (result === 'mastered') {
      toastSuccess(`「${card.concept.slice(0, 15)}」已掌握！${data.interval_days} 天后复习`)
    } else {
      toastSuccess(`「${card.concept.slice(0, 15)}」已标记，明天再复习`)
    }
  } catch (e) {
    console.error('复习提交失败:', e)
  } finally {
    card._reviewing = false
  }
}

// 导出 Anki CSV
async function exportCards() {
  exporting.value = true
  try {
    const params = new URLSearchParams()
    if (filter.difficulty) params.append('difficulty', filter.difficulty)
    const token = localStorage.getItem('token')
    const resp = await fetch(`/api/cards/export?${params.toString()}`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    if (!resp.ok) {
      const err = await resp.json().catch(() => ({}))
      toastSuccess(err.error || '导出失败')
      return
    }
    const blob = await resp.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `studyforge_cards_${new Date().toISOString().slice(0, 10)}.csv`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    toastSuccess(`已导出 ${cards.value.length} 张卡片为 Anki CSV`)
  } catch (e) {
    console.error('导出失败:', e)
    toastSuccess('导出失败，请重试')
  } finally {
    exporting.value = false
  }
}

// 切换书签
async function toggleBookmarkCard(card) {
  try {
    const res = await toggleBookmark(card.id)
    card.is_bookmarked = res.data.is_bookmarked
    toastSuccess(card.is_bookmarked ? '已添加书签' : '已移除书签')
  } catch (e) {
    console.error('书签切换失败:', e)
  }
}

// 保存笔记
let noteSaveTimer = null
function saveCardNote(card) {
  if (noteSaveTimer) clearTimeout(noteSaveTimer)
  noteSaveTimer = setTimeout(async () => {
    try {
      await updateCardNote(card.id, card.user_note || '')
    } catch (e) {
      console.error('笔记保存失败:', e)
    }
  }, 800)
}

// 图片粘贴/拖拽上传
const dragOverNote = ref(null) // 标记当前拖拽悬停的 textarea

const ALLOWED_IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/webp', 'image/svg+xml']

async function handlePasteImage(e, card) {
  const items = e.clipboardData?.items
  if (!items) return
  for (const item of items) {
    if (item.type.startsWith('image/')) {
      e.preventDefault()
      const file = item.getAsFile()
      if (file) await insertImageToNote(file, card, e.target)
      break
    }
  }
}

async function handleDropImage(e, card) {
  e.preventDefault()
  dragOverNote.value = null
  const files = e.dataTransfer?.files
  if (!files || files.length === 0) return
  const file = files[0]
  if (ALLOWED_IMAGE_TYPES.includes(file.type)) {
    await insertImageToNote(file, card, e.target)
  }
}

async function insertImageToNote(file, card, textarea) {
  if (file.size > 5 * 1024 * 1024) {
    useToast().error('图片大小不能超过 5MB')
    return
  }
  // 插入占位文本
  const placeholder = '\n![上传中...]()\n'
  const pos = textarea?.selectionStart ?? card.user_note?.length ?? 0
  const note = card.user_note || ''
  card.user_note = note.slice(0, pos) + placeholder + note.slice(pos)

  try {
    const res = await uploadImage(file)
    const url = res.data.url
    const markdown = `\n![image](${url})\n`
    card.user_note = card.user_note.replace(placeholder, markdown)
    saveCardNote(card)
  } catch (e) {
    card.user_note = card.user_note.replace(placeholder, '')
    console.error('图片上传失败:', e)
    useToast().error('图片上传失败')
  }
}

async function loadMaterialTags() {
  try {
    const res = await listMaterials({ limit: 200 })
    const mats = res.data.data || []
    const map = {}
    for (const m of mats) {
      if (m.tags) {
        map[m.id] = m.tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean)
      }
    }
    materialTagMap.value = map
  } catch (e) {
    // silent - material tags are supplementary
  }
}

async function loadCards() {
  loading.value = true
  cardResetScroll()
  cards.value = []
  try {
    const newItems = await cardLoadNext()
    cards.value.push(...newItems)
    fromCache.value = false
  } catch (e) {
    console.error('卡片加载失败:', e)
  } finally {
    loading.value = false
  }
}

async function loadMoreCards() {
  try {
    const newItems = await cardLoadNext()
    cards.value.push(...newItems)
  } catch (e) {
    console.error('加载更多卡片失败:', e)
  }
}

function setupCardInfiniteScroll() {
  if (cardSentinelRef.value && cardGridRef.value) {
    cardInitInfiniteScroll(cardSentinelRef.value, cardGridRef.value)
    initCardScrollListener()
  }
}

watch(
  () => filter.difficulty,
  async () => {
    await loadCards()
    await nextTick()
    setupCardInfiniteScroll()
  }
)
onMounted(async () => {
  await loadCards()
  loadMaterialTags()
  loadMyDecks()
  cardInitExplain()
  // 监听窗口大小变化以更新网格列数
  gridResizeHandler = () => updateGridCols()
  window.addEventListener('resize', gridResizeHandler)
  await nextTick()
  setupCardInfiniteScroll()
})
onUnmounted(() => {
  cardCleanupExplain()
  destroyCardScrollListener()
  if (gridResizeHandler) window.removeEventListener('resize', gridResizeHandler)
})
</script>

<style scoped>
/* KaTeX 暗色模式适配 */
:deep(.katex) {
  color: inherit;
}

/* 卡片内嵌图片样式 */
:deep(.card-image) {
  max-width: 100%;
  height: auto;
  border-radius: 0.5rem;
  margin: 0.5rem 0;
  border: 1px solid #e5e7eb;
}
:deep(.dark .card-image) {
  border-color: #4b5563;
  box-shadow: 0 0 0 1px rgba(75, 85, 99, 0.3);
}

/* Modal fade transition */
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
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
