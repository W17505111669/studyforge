<template>
  <div class="list-skeleton">
    <div v-for="i in count" :key="i" class="skeleton-item" :class="`skeleton-${type}`">
      <!-- 列表型骨架（Upload/Mistakes） -->
      <template v-if="type === 'list'">
        <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 animate-pulse">
          <div class="flex items-start gap-3">
            <div class="w-4 h-4 rounded bg-gray-200 dark:bg-gray-700 shrink-0 mt-1"></div>
            <div class="flex-1 space-y-3">
              <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-2/3"></div>
              <div class="flex gap-2">
                <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded-full w-12"></div>
                <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded-full w-16"></div>
              </div>
              <div class="flex items-center justify-between">
                <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded w-24"></div>
                <div class="h-5 bg-gray-100 dark:bg-gray-700 rounded w-16"></div>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- 卡片型骨架（Cards） -->
      <template v-else-if="type === 'card'">
        <div
          class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6 h-80 animate-pulse flex flex-col"
        >
          <div class="flex items-center justify-between mb-4">
            <div class="h-5 bg-gray-200 dark:bg-gray-700 rounded-full w-12"></div>
            <div class="h-4 w-4 bg-gray-100 dark:bg-gray-700 rounded"></div>
          </div>
          <div class="h-5 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-3"></div>
          <div class="h-4 bg-gray-100 dark:bg-gray-700 rounded w-full mb-2"></div>
          <div class="h-4 bg-gray-100 dark:bg-gray-700 rounded w-5/6 mb-2"></div>
          <div class="h-4 bg-gray-100 dark:bg-gray-700 rounded w-2/3"></div>
          <div class="flex-1"></div>
          <div class="flex gap-2 mt-4 pt-4 border-t border-gray-100 dark:border-gray-700">
            <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded-full w-10"></div>
            <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded-full w-14"></div>
            <div class="h-3 bg-gray-100 dark:bg-gray-700 rounded-full w-8"></div>
          </div>
        </div>
      </template>

      <!-- 聊天型骨架（Chat） -->
      <template v-else-if="type === 'chat'">
        <div class="flex" :class="i % 2 === 0 ? 'justify-end' : 'justify-start'">
          <div
            class="max-w-[75%] rounded-2xl px-4 py-3 animate-pulse"
            :class="
              i % 2 === 0
                ? 'bg-primary-100 dark:bg-primary-900/30 rounded-br-sm'
                : 'bg-gray-100 dark:bg-gray-700 rounded-bl-sm'
            "
          >
            <div
              class="h-4 rounded w-48 mb-2"
              :class="i % 2 === 0 ? 'bg-primary-200 dark:bg-primary-800/40' : 'bg-gray-200 dark:bg-gray-600'"
            ></div>
            <div
              class="h-4 rounded w-32"
              :class="i % 2 === 0 ? 'bg-primary-200 dark:bg-primary-800/40' : 'bg-gray-200 dark:bg-gray-600'"
            ></div>
          </div>
        </div>
      </template>

      <!-- 错题型骨架（Mistakes） -->
      <template v-else-if="type === 'mistake'">
        <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 animate-pulse">
          <div class="flex items-start gap-3">
            <div class="w-4 h-4 rounded bg-gray-200 dark:bg-gray-700 shrink-0 mt-1"></div>
            <div class="flex-1 space-y-3">
              <div class="flex gap-2">
                <div class="h-5 bg-gray-200 dark:bg-gray-700 rounded-full w-12"></div>
                <div class="h-5 bg-gray-100 dark:bg-gray-700 rounded-full w-16"></div>
                <div class="h-5 bg-gray-100 dark:bg-gray-700 rounded-full w-20"></div>
              </div>
              <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"></div>
              <div class="h-4 bg-gray-100 dark:bg-gray-700 rounded w-4/5"></div>
              <div class="grid grid-cols-2 gap-2 mt-2">
                <div class="h-8 bg-gray-100 dark:bg-gray-700 rounded-lg"></div>
                <div class="h-8 bg-gray-100 dark:bg-gray-700 rounded-lg"></div>
              </div>
            </div>
          </div>
          <div class="flex gap-2 mt-4 pt-3 border-t border-gray-100 dark:border-gray-700">
            <div class="h-7 bg-gray-100 dark:bg-gray-700 rounded w-20"></div>
            <div class="h-7 bg-gray-100 dark:bg-gray-700 rounded w-16"></div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
defineProps({
  count: { type: Number, default: 6 },
  type: { type: String, default: 'list', validator: (v) => ['list', 'card', 'chat', 'mistake'].includes(v) }
})
</script>

<style scoped>
.list-skeleton {
  display: flex;
  flex-direction: column;
}
.skeleton-list,
.skeleton-mistake {
  margin-bottom: 0.75rem;
}
.skeleton-card {
  /* grid gap handles spacing */
}
.skeleton-chat {
  margin-bottom: 1rem;
}
</style>
