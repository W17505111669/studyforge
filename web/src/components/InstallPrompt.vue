<template>
  <Teleport to="body">
    <Transition name="install-fade">
      <div v-if="showInstall" class="fixed bottom-20 md:bottom-6 right-4 md:right-6 z-[9990] max-w-sm">
        <div
          class="rounded-2xl border shadow-lg p-4 transition-colors"
          :class="
            isDark ? 'bg-gray-800 border-gray-700 shadow-black/30' : 'bg-white border-gray-200 shadow-gray-200/60'
          "
        >
          <!-- 顶部: 图标 + 标题 + 关闭 -->
          <div class="flex items-start gap-3">
            <div
              class="w-10 h-10 rounded-xl bg-gradient-to-br from-primary-500 to-indigo-600 flex items-center justify-center flex-shrink-0"
            >
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                />
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <h3 class="font-semibold text-sm" :class="isDark ? 'text-gray-100' : 'text-gray-900'">
                安装 StudyForge Pro
              </h3>
              <p class="text-xs mt-0.5 leading-relaxed" :class="isDark ? 'text-gray-400' : 'text-gray-500'">
                添加到桌面，获得更好的学习体验。支持离线使用。
              </p>
            </div>
            <button
              class="p-1 rounded-lg transition-colors flex-shrink-0"
              :class="
                isDark
                  ? 'text-gray-500 hover:text-gray-300 hover:bg-gray-700'
                  : 'text-gray-400 hover:text-gray-600 hover:bg-gray-100'
              "
              aria-label="关闭"
              @click="dismissInstall"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- 底部: 按钮 -->
          <div class="flex gap-2 mt-3">
            <button
              class="flex-1 px-3 py-2 text-xs font-medium rounded-lg border transition-colors"
              :class="
                isDark
                  ? 'border-gray-600 text-gray-300 hover:bg-gray-700'
                  : 'border-gray-300 text-gray-600 hover:bg-gray-50'
              "
              @click="dismissInstall"
            >
              稍后再说
            </button>
            <button
              class="flex-1 px-3 py-2 text-xs font-semibold rounded-lg text-white transition-colors bg-gradient-to-r from-primary-600 to-indigo-600 hover:from-primary-700 hover:to-indigo-700"
              @click="installApp"
            >
              立即安装
            </button>
          </div>

          <!-- 特性标签 -->
          <div class="flex flex-wrap gap-1.5 mt-2.5">
            <span
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium"
              :class="isDark ? 'bg-emerald-900/30 text-emerald-400' : 'bg-emerald-50 text-emerald-700'"
            >
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              离线可用
            </span>
            <span
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium"
              :class="isDark ? 'bg-blue-900/30 text-blue-400' : 'bg-blue-50 text-blue-700'"
            >
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              快速启动
            </span>
            <span
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium"
              :class="isDark ? 'bg-purple-900/30 text-purple-400' : 'bg-purple-50 text-purple-700'"
            >
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z"
                />
              </svg>
              原生体验
            </span>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useDarkMode } from '../composables/useDarkMode'

const { isDark } = useDarkMode()

const showInstall = ref(false)
let deferredPrompt = null

const DISMISS_KEY = 'studyforge-install-dismissed'

onMounted(() => {
  // 如果用户已拒绝过, 不再显示
  if (localStorage.getItem(DISMISS_KEY)) return

  // 已经安装, 不再显示
  if (window.matchMedia('(display-mode: standalone)').matches) return

  // 捕获 beforeinstallprompt 事件
  window.addEventListener('beforeinstallprompt', (event) => {
    event.preventDefault()
    deferredPrompt = event

    // 延迟 30s 再显示, 避免打扰用户
    setTimeout(() => {
      if (!localStorage.getItem(DISMISS_KEY)) {
        showInstall.value = true
      }
    }, 30000)
  })
})

async function installApp() {
  if (!deferredPrompt) {
    // 没有 beforeinstallprompt, 可能是 Safari 或不支持
    showInstall.value = false
    return
  }

  deferredPrompt.prompt()
  const { outcome } = await deferredPrompt.userChoice

  if (outcome === 'accepted') {
    showInstall.value = false
  } else {
    // 用户拒绝
    dismissInstall()
  }

  deferredPrompt = null
}

function dismissInstall() {
  showInstall.value = false
  // 30 天内不再显示
  localStorage.setItem(DISMISS_KEY, Date.now().toString())
}
</script>

<style scoped>
.install-fade-enter-active {
  transition:
    transform 0.4s ease,
    opacity 0.4s ease;
}
.install-fade-leave-active {
  transition:
    transform 0.25s ease,
    opacity 0.25s ease;
}
.install-fade-enter-from {
  transform: translateY(20px) scale(0.95);
  opacity: 0;
}
.install-fade-leave-to {
  transform: translateY(10px) scale(0.98);
  opacity: 0;
}
</style>
