<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-6xl mx-auto">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-6 sm:mb-8 gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">知识卡片库</h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">单击放大 · 双击翻转查看详情</p>
      </div>
      <div class="flex items-center gap-2 sm:gap-3 flex-wrap">
        <!-- 搜索框 -->
        <div class="relative">
          <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索卡片..."
            class="pl-9 pr-8 py-2 rounded-lg border border-gray-200 text-sm w-full sm:w-56 focus:outline-none focus:ring-2 focus:ring-primary-300 focus:border-primary-400 transition dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200"
          />
          <button
            v-if="searchQuery"
            @click="searchQuery = ''"
            class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 transition dark:text-gray-500 dark:hover:text-gray-400"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <!-- 待复习过滤 -->
        <button
          @click="filter.due = !filter.due"
          class="px-3 py-2 rounded-lg border text-sm transition flex items-center gap-1.5"
          :class="filter.due ? 'bg-primary-50 border-primary-300 text-primary-700 dark:bg-primary-900/20' : 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-gray-600 dark:text-gray-400 dark:hover:border-gray-500'"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span v-if="dueCount > 0">{{ dueCount }} 待复习</span>
          <span v-else>待复习</span>
        </button>
        <!-- 仅书签过滤 -->
        <button
          @click="filter.bookmarked = !filter.bookmarked"
          class="px-3 py-2 rounded-lg border text-sm transition flex items-center gap-1.5"
          :class="filter.bookmarked ? 'bg-amber-50 border-amber-300 text-amber-700 dark:bg-amber-900/20' : 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-gray-600 dark:text-gray-400 dark:hover:border-gray-500'"
        >
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
            <path d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
          </svg>
          <span v-if="bookmarkCount > 0">{{ bookmarkCount }} 书签</span>
          <span v-else>书签</span>
        </button>
        <select v-model="filter.difficulty" class="px-3 py-2 rounded-lg border border-gray-200 text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200">
          <option value="">全部难度</option>
          <option value="easy">简单</option>
          <option value="medium">中等</option>
          <option value="hard">困难</option>
        </select>
        <!-- 导出 Anki CSV -->
        <button
          @click="exportCards"
          :disabled="cards.length === 0 || exporting"
          class="px-3 py-2 rounded-lg border border-gray-200 text-sm text-gray-600 hover:border-primary-300 hover:text-primary-600 transition flex items-center gap-1.5 disabled:opacity-40 disabled:cursor-not-allowed dark:border-gray-600 dark:text-gray-400"
          title="导出为 Anki 兼容 CSV"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <span>{{ exporting ? '导出中...' : '导出 Anki' }}</span>
        </button>
      </div>
    </div>

    <!-- 搜索结果计数 -->
    <div v-if="searchQuery && !loading" class="mb-4 text-sm text-gray-500 dark:text-gray-400">
      找到 <span class="font-medium text-primary-600">{{ filteredCards.length }}</span> 张匹配卡片
      <span v-if="filteredCards.length < cards.length" class="text-gray-400 dark:text-gray-500">（共 {{ cards.length }} 张）</span>
    </div>

    <!-- 空状态 -->
    <div v-if="cards.length === 0 && !loading" class="bg-white rounded-xl shadow-sm border border-gray-100 p-16 text-center dark:bg-gray-800 dark:border-gray-700">
      <p class="text-5xl mb-4">🃏</p>
      <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">还没有知识卡片</h3>
      <p class="text-gray-400 dark:text-gray-500 mb-4">上传学习材料后，AI 会自动为你生成精美卡片</p>
      <router-link to="/upload" class="text-primary-600 hover:underline text-sm">去上传材料</router-link>
    </div>

    <!-- 搜索无结果 -->
    <div v-else-if="filteredCards.length === 0 && !loading && searchQuery" class="bg-white rounded-xl shadow-sm border border-gray-100 p-16 text-center dark:bg-gray-800 dark:border-gray-700">
      <p class="text-5xl mb-4">🔍</p>
      <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">没有找到匹配的卡片</h3>
      <p class="text-gray-400 dark:text-gray-500 mb-4">试试换个关键词，或清空搜索条件</p>
      <button @click="searchQuery = ''" class="text-primary-600 hover:underline text-sm">清空搜索</button>
    </div>

    <!-- 无待复习卡片 -->
    <div v-else-if="filteredCards.length === 0 && !loading && filter.due" class="bg-white rounded-xl shadow-sm border border-gray-100 p-16 text-center dark:bg-gray-800 dark:border-gray-700">
      <p class="text-5xl mb-4">🎉</p>
      <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">暂无待复习卡片</h3>
      <p class="text-gray-400 dark:text-gray-500 mb-4">所有卡片都已掌握或未到复习时间</p>
      <button @click="filter.due = false" class="text-primary-600 hover:underline text-sm">查看全部卡片</button>
    </div>

    <!-- 无书签卡片 -->
    <div v-else-if="filteredCards.length === 0 && !loading && filter.bookmarked" class="bg-white rounded-xl shadow-sm border border-gray-100 p-16 text-center dark:bg-gray-800 dark:border-gray-700">
      <p class="text-5xl mb-4">🔖</p>
      <h3 class="text-lg font-medium text-gray-700 dark:text-gray-300 mb-2">暂无书签卡片</h3>
      <p class="text-gray-400 dark:text-gray-500 mb-4">点击卡片正面的书签图标来收藏重要卡片</p>
      <button @click="filter.bookmarked = false" class="text-primary-600 hover:underline text-sm">查看全部卡片</button>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="text-center py-12 text-gray-400 dark:text-gray-500">加载中...</div>

    <!-- 卡片瀑布流 -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div
        v-for="card in filteredCards"
        :key="card.id"
        class="flip-card h-80 cursor-pointer"
        :class="{ flipped: card._flipped }"
        @click="onCardClick(card)"
        @dblclick="onCardDblClick(card)"
      >
        <div class="flip-card-inner relative w-full h-full">
          <!-- 正面 -->
          <div class="flip-card-front absolute inset-0 bg-white rounded-xl shadow-sm border border-gray-100 p-6 flex flex-col dark:bg-gray-800 dark:border-gray-700">
            <div class="flex items-start justify-between mb-3">
              <span class="px-2 py-0.5 rounded text-xs font-medium" :class="diffClass(card.difficulty)">
                {{ diffLabel(card.difficulty) }}
              </span>
              <div class="flex items-center gap-1.5">
                <!-- 书签按钮 -->
                <button
                  @click.stop="toggleBookmarkCard(card)"
                  class="p-1 rounded transition hover:scale-110"
                  :class="card.is_bookmarked ? 'text-amber-500' : 'text-gray-300 hover:text-amber-400 dark:text-gray-600 dark:hover:text-amber-500'"
                  :title="card.is_bookmarked ? '移除书签' : '添加书签'"
                >
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                  </svg>
                </button>
                <!-- 复习状态指示 -->
                <span v-if="card.review_count > 0" class="text-xs px-1.5 py-0.5 bg-emerald-50 text-emerald-600 rounded dark:bg-emerald-900/30">
                  复习 {{ card.review_count }}次
                </span>
                <span v-else class="text-xs px-1.5 py-0.5 bg-blue-50 text-blue-500 rounded dark:bg-blue-900/30">新卡片</span>
              </div>
            </div>
            <h3 class="text-lg font-bold text-gray-900 dark:text-gray-100 mb-2 line-clamp-2" v-html="highlight(card.concept)"></h3>
            <p class="text-sm text-gray-500 dark:text-gray-400 flex-1 line-clamp-3" v-html="highlight(card.detail)"></p>
            <div class="mt-3 flex flex-wrap gap-1">
              <span v-for="tag in (card.tags || [])" :key="tag"
                class="px-2 py-0.5 rounded text-xs"
                :class="isTagMatch(tag) ? 'bg-primary-100 text-primary-700 font-medium' : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'"
                v-html="highlight(tag)"
              ></span>
              <!-- 材料标签 -->
              <template v-if="materialTagMap[card.material_id]?.length">
                <span class="w-px h-3 bg-gray-200 dark:bg-gray-600 mx-0.5 self-center"></span>
                <span v-for="mt in materialTagMap[card.material_id]" :key="'m-'+mt"
                  class="px-1.5 py-0.5 rounded text-[10px] bg-indigo-50 text-indigo-500 dark:bg-indigo-900/20 dark:text-indigo-400"
                >{{ mt }}</span>
              </template>
            </div>
            <p class="text-xs text-gray-300 dark:text-gray-600 mt-2 text-center">单击放大 · 双击翻转</p>
          </div>

          <!-- 背面 -->
          <div class="flip-card-back absolute inset-0 bg-gradient-to-br from-primary-50 to-indigo-50 rounded-xl shadow-sm border border-primary-100 p-6 flex flex-col dark:from-gray-800 dark:to-gray-700 dark:border-gray-600">
            <h3 class="text-base font-bold text-primary-700 mb-2" v-html="highlight(card.concept)"></h3>
            <p class="text-sm text-gray-700 dark:text-gray-300 flex-1 overflow-y-auto" v-html="highlight(card.detail)"></p>
            <div v-if="card.formula" class="mt-2 p-2.5 bg-white rounded-lg border border-primary-100 dark:bg-gray-700 dark:border-gray-600">
              <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">关键公式/代码</p>
              <code class="text-sm text-primary-600" v-html="renderMath(card.formula)"></code>
            </div>
            <div v-if="card.memory_tip" class="mt-2 p-2.5 bg-amber-50 rounded-lg border border-amber-100 dark:bg-amber-900/30 dark:border-amber-800">
              <p class="text-xs text-amber-600 mb-1">💡 记忆技巧</p>
              <p class="text-sm text-amber-800 dark:text-amber-300" v-html="renderMath(card.memory_tip)"></p>
            </div>

            <!-- 个人笔记 -->
            <div class="mt-2" @click.stop>
              <div class="flex items-center gap-1 mb-1 cursor-pointer" @click="card._noteEditing = !card._noteEditing">
                <svg class="w-3.5 h-3.5 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
                <span class="text-xs text-gray-400 dark:text-gray-500">笔记</span>
              </div>
              <textarea
                v-if="card._noteEditing || card.user_note"
                v-model="card.user_note"
                @input="saveCardNote(card)"
                @click.stop
                placeholder="添加个人笔记..."
                class="w-full text-xs p-2 rounded-lg border border-gray-200 bg-white text-gray-700 resize-none focus:outline-none focus:ring-1 focus:ring-primary-300 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-300"
                rows="2"
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
                  @click="submitReview(card, 'review')"
                  :disabled="card._reviewing"
                  class="flex-1 py-2 rounded-lg text-sm font-medium transition border
                    bg-orange-50 border-orange-200 text-orange-700 hover:bg-orange-100
                    disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  🔄 再复习
                </button>
                <button
                  @click="submitReview(card, 'mastered')"
                  :disabled="card._reviewing"
                  class="flex-1 py-2 rounded-lg text-sm font-medium transition border
                    bg-emerald-50 border-emerald-200 text-emerald-700 hover:bg-emerald-100
                    disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  ✅ 已掌握
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Enlarged card modal -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div v-if="enlargedCard"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
          @click.self="closeEnlarged"
        >
          <!-- Close button -->
          <button @click="closeEnlarged"
            class="absolute top-4 right-4 z-10 p-2 rounded-full bg-white/20 hover:bg-white/40 text-white transition"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>

          <!-- Modal flip card -->
          <div class="flip-card w-full max-w-lg cursor-pointer" style="height: 520px"
            :class="{ flipped: modalFlipped }"
            @click="modalFlipped = !modalFlipped"
          >
            <div class="flip-card-inner relative w-full h-full">
              <!-- Front -->
              <div class="flip-card-front absolute inset-0 bg-white rounded-2xl shadow-2xl border border-gray-100 p-8 flex flex-col dark:bg-gray-800 dark:border-gray-700 overflow-y-auto">
                <div class="flex items-start justify-between mb-4">
                  <span class="px-3 py-1 rounded-lg text-sm font-semibold" :class="diffClass(enlargedCard.difficulty)">
                    {{ diffLabel(enlargedCard.difficulty) }}
                  </span>
                  <div class="flex items-center gap-2">
                    <!-- 书签按钮 -->
                    <button
                      @click.stop="toggleBookmarkCard(enlargedCard)"
                      class="p-1.5 rounded-lg transition hover:scale-110"
                      :class="enlargedCard.is_bookmarked ? 'text-amber-500' : 'text-gray-300 hover:text-amber-400 dark:text-gray-600 dark:hover:text-amber-500'"
                      :title="enlargedCard.is_bookmarked ? '移除书签' : '添加书签'"
                    >
                      <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                        <path d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                      </svg>
                    </button>
                    <span v-if="enlargedCard.review_count > 0" class="text-xs px-2 py-1 bg-emerald-50 text-emerald-600 rounded-lg dark:bg-emerald-900/30">
                      复习 {{ enlargedCard.review_count }}次
                    </span>
                    <span v-else class="text-xs px-2 py-1 bg-blue-50 text-blue-500 rounded-lg dark:bg-blue-900/30">新卡片</span>
                  </div>
                </div>
                <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4 leading-relaxed" v-html="renderMath(enlargedCard.concept)"></h2>
                <p class="text-base text-gray-600 dark:text-gray-300 flex-1 leading-relaxed whitespace-pre-wrap" v-html="renderMath(enlargedCard.detail)"></p>
                <div v-if="enlargedCard.tags && enlargedCard.tags.length" class="flex flex-wrap gap-1.5 mt-4">
                  <span v-for="tag in enlargedCard.tags" :key="tag"
                    class="px-2.5 py-1 rounded-lg text-xs bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400"
                  >{{ tag }}</span>
                </div>
                <p class="text-xs text-gray-300 dark:text-gray-600 mt-4 text-center">点击卡片翻转查看详情</p>
              </div>
              <!-- Back -->
              <div class="flip-card-back absolute inset-0 bg-gradient-to-br from-primary-50 to-indigo-50 rounded-2xl shadow-2xl border border-primary-100 p-8 flex flex-col dark:from-gray-800 dark:to-gray-700 dark:border-gray-600 overflow-y-auto">
                <h2 class="text-xl font-bold text-primary-700 dark:text-primary-400 mb-4" v-html="renderMath(enlargedCard.concept)"></h2>
                <p class="text-base text-gray-700 dark:text-gray-300 flex-1 leading-relaxed whitespace-pre-wrap" v-html="renderMath(enlargedCard.detail)"></p>
                <div v-if="enlargedCard.formula" class="mt-4 p-4 bg-white rounded-xl border border-primary-100 dark:bg-gray-700 dark:border-gray-600">
                  <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">关键公式/代码</p>
                  <code class="text-sm text-primary-600 dark:text-primary-400 whitespace-pre-wrap break-all" v-html="renderMath(enlargedCard.formula)"></code>
                </div>
                <div v-if="enlargedCard.memory_tip" class="mt-4 p-4 bg-amber-50 rounded-xl border border-amber-100 dark:bg-amber-900/30 dark:border-amber-800">
                  <p class="text-xs text-amber-600 dark:text-amber-400 mb-1">记忆技巧</p>
                  <p class="text-sm text-amber-800 dark:text-amber-300 leading-relaxed" v-html="renderMath(enlargedCard.memory_tip)"></p>
                </div>
                <!-- 个人笔记 -->
                <div class="mt-4" @click.stop>
                  <div class="flex items-center gap-1.5 mb-2">
                    <svg class="w-4 h-4 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                    <span class="text-sm text-gray-500 dark:text-gray-400 font-medium">个人笔记</span>
                  </div>
                  <textarea
                    v-model="enlargedCard.user_note"
                    @input="saveCardNote(enlargedCard)"
                    @click.stop
                    placeholder="添加你的学习笔记..."
                    class="w-full text-sm p-3 rounded-xl border border-gray-200 bg-white text-gray-700 resize-none focus:outline-none focus:ring-2 focus:ring-primary-300 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-300"
                    rows="3"
                  ></textarea>
                </div>
                <!-- Review stats + buttons -->
                <div class="mt-4 pt-4 border-t border-primary-100 dark:border-gray-600" @click.stop>
                  <div class="flex items-center justify-between text-xs text-gray-400 dark:text-gray-500 mb-3">
                    <span>间隔 {{ enlargedCard.interval_days || 0 }} 天 · 系数 {{ (enlargedCard.ease_factor || 2.5).toFixed(1) }}</span>
                    <span v-if="enlargedCard.next_review_at">下次: {{ formatDate(enlargedCard.next_review_at) }}</span>
                    <span v-else>尚未复习</span>
                  </div>
                  <div class="flex gap-3">
                    <button @click="submitReview(enlargedCard, 'review')" :disabled="enlargedCard._reviewing"
                      class="flex-1 py-2.5 rounded-lg text-sm font-medium transition border bg-orange-50 border-orange-200 text-orange-700 hover:bg-orange-100 disabled:opacity-50 disabled:cursor-not-allowed"
                    >再复习</button>
                    <button @click="submitReview(enlargedCard, 'mastered')" :disabled="enlargedCard._reviewing"
                      class="flex-1 py-2.5 rounded-lg text-sm font-medium transition border bg-emerald-50 border-emerald-200 text-emerald-700 hover:bg-emerald-100 disabled:opacity-50 disabled:cursor-not-allowed"
                    >已掌握</button>
                  </div>
                </div>
                <p class="text-xs text-gray-300 dark:text-gray-600 mt-3 text-center">点击卡片翻转回正面</p>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { listCards, reviewCard, listMaterials, toggleBookmark, updateCardNote } from '../api/client'
import { useToast } from '../composables/useToast'
import { renderMath, renderMathWithHighlight } from '../composables/useMathRender'

const { success: toastSuccess } = useToast()

const cards = ref([])
const loading = ref(true)
const searchQuery = ref('')
const exporting = ref(false)
const filter = reactive({ difficulty: '', due: false, bookmarked: false })

// Enlarged modal state
const enlargedCard = ref(null)
const modalFlipped = ref(false)
let cardClickTimer = null

// 材料标签映射 { material_id: ['tag1', 'tag2'] }
const materialTagMap = ref({})

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
  if (cardClickTimer) { clearTimeout(cardClickTimer); cardClickTimer = null }
  card._flipped = !card._flipped
}

function closeEnlarged() {
  enlargedCard.value = null
  modalFlipped.value = false
}

// 待复习计数
const dueCount = computed(() => {
  const now = new Date()
  return cards.value.filter(c => {
    if (!c.next_review_at) return true // 新卡片也算
    return new Date(c.next_review_at) <= now
  }).length
})

// 书签计数
const bookmarkCount = computed(() => {
  return cards.value.filter(c => c.is_bookmarked).length
})

// 模糊过滤：按概念名、标签、详情匹配
const filteredCards = computed(() => {
  let result = cards.value
  // 待复习过滤
  if (filter.due) {
    const now = new Date()
    result = result.filter(c => {
      if (!c.next_review_at) return true
      return new Date(c.next_review_at) <= now
    })
  }
  // 书签过滤
  if (filter.bookmarked) {
    result = result.filter(c => c.is_bookmarked)
  }
  // 搜索过滤
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    result = result.filter(card => {
      if ((card.concept || '').toLowerCase().includes(q)) return true
      if ((card.detail || '').toLowerCase().includes(q)) return true
      if ((card.tags || []).some(tag => tag.toLowerCase().includes(q))) return true
      if ((card.formula || '').toLowerCase().includes(q)) return true
      if ((card.memory_tip || '').toLowerCase().includes(q)) return true
      return false
    })
  }
  return result
})

// 搜索关键词高亮 + LaTeX 数学公式渲染（先提取数学公式，对纯文本部分做转义+高亮，再还原 KaTeX）
function highlight(text) {
  const q = searchQuery.value.trim()
  const escapeAndHighlight = (t) => {
    if (!q) return escapeHtml(t)
    const escaped = escapeHtml(t)
    const escapedQ = escapeHtml(q)
    const regex = new RegExp(`(${escapeRegex(escapedQ)})`, 'gi')
    return escaped.replace(regex, '<mark class="bg-yellow-200 dark:bg-yellow-700/40 text-yellow-900 dark:text-yellow-200 rounded px-0.5">$1</mark>')
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
  return { easy: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400', medium: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400', hard: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' }[d] || 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
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
      headers: { Authorization: `Bearer ${token}` },
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

async function loadMaterialTags() {
  try {
    const res = await listMaterials({ limit: 200 })
    const mats = res.data.data || []
    const map = {}
    for (const m of mats) {
      if (m.tags) {
        map[m.id] = m.tags.split(',').map(t => t.trim()).filter(Boolean)
      }
    }
    materialTagMap.value = map
  } catch (e) {
    // silent - material tags are supplementary
  }
}

async function loadCards() {
  loading.value = true
  try {
    const params = { limit: 200 }
    if (filter.difficulty) params.difficulty = filter.difficulty
    const res = await listCards(params)
    cards.value = (res.data.data || []).map(c => ({
      ...c,
      tags: typeof c.tags === 'string' && c.tags ? c.tags.split(',').map(t => t.trim()) : [],
      _flipped: false,
      _reviewing: false,
      _noteEditing: false,
    }))
  } catch (e) {
    console.error('卡片加载失败:', e)
  } finally {
    loading.value = false
  }
}

watch(() => filter.difficulty, loadCards)
onMounted(() => {
  loadCards()
  loadMaterialTags()
})
</script>

<style scoped>
/* KaTeX 暗色模式适配 */
:deep(.katex) {
  color: inherit;
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
</style>
