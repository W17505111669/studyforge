<template>
  <Teleport to="body">
    <Transition name="help-modal">
      <div
        v-if="showHelpModal"
        class="fixed inset-0 z-[70] flex items-center justify-center p-4"
        @click.self="closeHelp"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/50 dark:bg-black/60"></div>

        <!-- Modal -->
        <div
          class="relative w-full max-w-lg rounded-2xl border overflow-hidden flex flex-col max-h-[80vh]"
          :class="
            isDark
              ? 'bg-gray-800 border-gray-700 shadow-2xl shadow-black/50'
              : 'bg-white border-gray-200 shadow-2xl shadow-gray-300/50'
          "
        >
          <!-- Header -->
          <div
            class="flex items-center justify-between px-6 py-4 border-b"
            :class="isDark ? 'border-gray-700' : 'border-gray-200'"
          >
            <div class="flex items-center gap-3">
              <div
                class="w-9 h-9 rounded-lg flex items-center justify-center"
                :class="isDark ? 'bg-primary-900/30 text-primary-400' : 'bg-primary-100 text-primary-600'"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"
                  />
                </svg>
              </div>
              <div>
                <h2 class="text-lg font-semibold" :class="isDark ? 'text-gray-100' : 'text-gray-900'">键盘快捷键</h2>
                <p class="text-xs" :class="isDark ? 'text-gray-500' : 'text-gray-500'">
                  使用快捷键高效操作 · 按
                  <kbd
                    class="inline-block px-1 py-0.5 rounded text-[10px] font-mono border"
                    :class="
                      isDark ? 'bg-gray-700 border-gray-600 text-gray-300' : 'bg-gray-100 border-gray-300 text-gray-600'
                    "
                  >
                    ?
                  </kbd>
                  关闭
                </p>
              </div>
            </div>
            <button
              class="p-1.5 rounded-lg transition-colors"
              :class="
                isDark
                  ? 'text-gray-400 hover:text-gray-200 hover:bg-gray-700'
                  : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
              "
              @click="closeHelp"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Body -->
          <div class="flex-1 overflow-y-auto px-6 py-4 space-y-6 custom-scroll">
            <div v-for="group in allShortcuts" :key="group.category">
              <!-- Category header -->
              <div class="flex items-center gap-2 mb-3">
                <span :class="isDark ? 'text-primary-400' : 'text-primary-600'" v-html="group.icon"></span>
                <h3
                  class="text-sm font-semibold uppercase tracking-wide"
                  :class="isDark ? 'text-gray-300' : 'text-gray-700'"
                >
                  {{ group.category }}
                </h3>
                <div class="flex-1 h-px" :class="isDark ? 'bg-gray-700' : 'bg-gray-200'"></div>
              </div>

              <!-- Shortcuts grid -->
              <div class="space-y-2">
                <div
                  v-for="shortcut in group.shortcuts"
                  :key="shortcut.desc"
                  class="flex items-center justify-between py-1.5 px-2 rounded-lg transition-colors"
                  :class="isDark ? 'hover:bg-gray-700/50' : 'hover:bg-gray-50'"
                >
                  <span class="text-sm" :class="isDark ? 'text-gray-300' : 'text-gray-700'">
                    {{ shortcut.desc }}
                  </span>
                  <div class="flex items-center gap-1 ml-4 shrink-0">
                    <template v-for="(key, idx) in shortcut.keys" :key="idx">
                      <span v-if="idx > 0" class="text-[10px] text-gray-400 mx-0.5">+</span>
                      <kbd
                        class="inline-flex items-center justify-center min-w-[24px] px-1.5 py-1 rounded-md text-xs font-mono font-medium border shadow-sm"
                        :class="
                          isDark
                            ? 'bg-gray-700 border-gray-600 text-gray-200 shadow-gray-900/50'
                            : 'bg-gray-50 border-gray-300 text-gray-700 shadow-gray-200/50'
                        "
                      >
                        {{ key }}
                      </kbd>
                    </template>
                  </div>
                </div>
              </div>
            </div>

            <!-- Navigation prefix hint -->
            <div
              class="rounded-xl p-4 border"
              :class="isDark ? 'bg-indigo-900/20 border-indigo-800/30' : 'bg-indigo-50 border-indigo-200'"
            >
              <div class="flex items-start gap-3">
                <div
                  class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 mt-0.5"
                  :class="isDark ? 'bg-indigo-800/40 text-indigo-300' : 'bg-indigo-100 text-indigo-600'"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                </div>
                <div>
                  <p class="text-sm font-medium" :class="isDark ? 'text-indigo-300' : 'text-indigo-800'">导航前缀</p>
                  <p class="text-xs mt-1" :class="isDark ? 'text-indigo-400' : 'text-indigo-600'">
                    按
                    <kbd
                      class="inline-block px-1 py-0.5 rounded text-[10px] font-mono border mx-0.5"
                      :class="
                        isDark
                          ? 'bg-indigo-800/50 border-indigo-700 text-indigo-300'
                          : 'bg-indigo-100 border-indigo-300 text-indigo-700'
                      "
                    >
                      g
                    </kbd>
                    后 1.5 秒内按第二个字母键跳转对应页面。例如
                    <kbd
                      class="inline-block px-1 py-0.5 rounded text-[10px] font-mono border mx-0.5"
                      :class="
                        isDark
                          ? 'bg-indigo-800/50 border-indigo-700 text-indigo-300'
                          : 'bg-indigo-100 border-indigo-300 text-indigo-700'
                      "
                    >
                      g
                    </kbd>
                    <span class="mx-0.5">→</span>
                    <kbd
                      class="inline-block px-1 py-0.5 rounded text-[10px] font-mono border"
                      :class="
                        isDark
                          ? 'bg-indigo-800/50 border-indigo-700 text-indigo-300'
                          : 'bg-indigo-100 border-indigo-300 text-indigo-700'
                      "
                    >
                      d
                    </kbd>
                    跳转仪表盘
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div
            class="px-6 py-3 border-t flex items-center justify-between"
            :class="isDark ? 'border-gray-700 bg-gray-800/50' : 'border-gray-200 bg-gray-50/50'"
          >
            <p class="text-xs" :class="isDark ? 'text-gray-500' : 'text-gray-400'">在输入框中快捷键自动禁用</p>
            <button
              class="px-4 py-1.5 rounded-lg text-sm font-medium transition-colors"
              :class="
                isDark ? 'bg-gray-700 text-gray-300 hover:bg-gray-600' : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
              "
              @click="closeHelp"
            >
              关闭
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { useDarkMode } from '../composables/useDarkMode'
import { useShortcuts } from '../composables/useShortcuts'

const { isDark } = useDarkMode()
const { showHelpModal, closeHelp, allShortcuts } = useShortcuts()
</script>

<style scoped>
/* Modal transition */
.help-modal-enter-active,
.help-modal-leave-active {
  transition: opacity 0.2s ease;
}
.help-modal-enter-active .relative,
.help-modal-leave-active .relative {
  transition:
    transform 0.2s ease,
    opacity 0.2s ease;
}
.help-modal-enter-from,
.help-modal-leave-to {
  opacity: 0;
}
.help-modal-enter-from .relative,
.help-modal-leave-to .relative {
  transform: scale(0.95) translateY(8px);
  opacity: 0;
}
.help-modal-enter-to .relative,
.help-modal-leave-from .relative {
  transform: scale(1) translateY(0);
  opacity: 1;
}

/* Custom scrollbar */
.custom-scroll::-webkit-scrollbar {
  width: 4px;
}
.custom-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroll::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.3);
  border-radius: 4px;
}
.custom-scroll::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.5);
}
:global(.dark) .custom-scroll::-webkit-scrollbar-thumb {
  background: rgba(75, 85, 99, 0.4);
}
:global(.dark) .custom-scroll::-webkit-scrollbar-thumb:hover {
  background: rgba(75, 85, 99, 0.6);
}
</style>
