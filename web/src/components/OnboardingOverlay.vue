<template>
  <Teleport to="body">
    <Transition name="onboarding-fade">
      <div v-if="isActive" class="fixed inset-0 z-[9999]">
        <!-- 聚光灯遮罩 (box-shadow 挖孔方案，高亮区域点击可穿透) -->
        <div
          class="absolute pointer-events-none transition-all duration-500 ease-out"
          :style="{
            left: spotlightRect.x + 'px',
            top: spotlightRect.y + 'px',
            width: spotlightRect.width + 'px',
            height: spotlightRect.height + 'px',
            borderRadius: spotlightRect.radius + 'px',
            boxShadow: '0 0 0 9999px rgba(0, 0, 0, 0.6)',
          }"
        ></div>

        <!-- 高亮边框（发光脉冲效果） -->
        <div
          class="absolute rounded-xl border-2 border-primary-400 pointer-events-none transition-all duration-500 ease-out"
          :style="{
            left: spotlightRect.x - 4 + 'px',
            top: spotlightRect.y - 4 + 'px',
            width: spotlightRect.width + 8 + 'px',
            height: spotlightRect.height + 8 + 'px',
            boxShadow: '0 0 0 4px rgba(99, 102, 241, 0.2), 0 0 20px rgba(99, 102, 241, 0.3)',
            animation: 'onboarding-pulse 2s ease-in-out infinite',
          }"
        ></div>

        <!-- 四周可点击区域（点击跳过引导） -->
        <!-- 上方 -->
        <div
          class="absolute left-0 right-0 top-0 cursor-pointer"
          :style="{ height: Math.max(0, spotlightRect.y - 4) + 'px' }"
          @click="skipOnboarding"
        ></div>
        <!-- 下方 -->
        <div
          class="absolute left-0 right-0 bottom-0 cursor-pointer"
          :style="{ top: (spotlightRect.y + spotlightRect.height + 4) + 'px' }"
          @click="skipOnboarding"
        ></div>
        <!-- 左侧 -->
        <div
          class="absolute left-0 cursor-pointer"
          :style="{
            top: (spotlightRect.y - 4) + 'px',
            width: Math.max(0, spotlightRect.x - 4) + 'px',
            height: (spotlightRect.height + 8) + 'px',
          }"
          @click="skipOnboarding"
        ></div>
        <!-- 右侧 -->
        <div
          class="absolute right-0 cursor-pointer"
          :style="{
            top: (spotlightRect.y - 4) + 'px',
            left: (spotlightRect.x + spotlightRect.width + 4) + 'px',
            height: (spotlightRect.height + 8) + 'px',
          }"
          @click="skipOnboarding"
        ></div>

        <!-- 工具提示卡片 -->
        <Transition :name="tooltipTransition" mode="out-in">
          <div
            :key="currentStep"
            class="absolute z-10 w-[360px] max-w-[calc(100vw-32px)] transition-all duration-500"
            :style="tooltipStyle"
          >
            <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden">
              <!-- 顶部装饰条 -->
              <div class="h-1 bg-gradient-to-r from-primary-500 via-purple-500 to-pink-500"></div>

              <div class="p-5">
                <!-- 步骤指示器 -->
                <div class="flex items-center gap-2 mb-3">
                  <span
                    v-for="(s, i) in totalSteps"
                    :key="i"
                    class="h-1.5 rounded-full transition-all duration-300"
                    :class="i === currentStep
                      ? 'bg-primary-500 w-6'
                      : i < currentStep
                        ? 'bg-primary-300 dark:bg-primary-700 w-3'
                        : 'bg-gray-300 dark:bg-gray-600 w-3'"
                  ></span>
                  <span class="ml-auto text-xs text-gray-400 dark:text-gray-500 font-mono">
                    {{ currentStep + 1 }}/{{ totalSteps }}
                  </span>
                </div>

                <!-- 步骤图标 + 标题 -->
                <div class="flex items-start gap-3 mb-3">
                  <div class="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0"
                    :class="stepIconClass">
                    <span class="text-lg" v-html="stepIcon"></span>
                  </div>
                  <div>
                    <h3 class="text-base font-bold text-gray-900 dark:text-white leading-tight">
                      {{ currentStepData?.title }}
                    </h3>
                  </div>
                </div>

                <!-- 描述文字 -->
                <p class="text-sm text-gray-600 dark:text-gray-400 leading-relaxed mb-5">
                  {{ currentStepData?.description }}
                </p>

                <!-- 操作按钮 -->
                <div class="flex items-center gap-2">
                  <button
                    @click="skipOnboarding"
                    class="text-xs text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                  >
                    跳过引导
                  </button>
                  <div class="flex-1"></div>
                  <button
                    v-if="!isFirstStep"
                    @click="prevStep"
                    class="px-3 py-1.5 text-sm font-medium text-gray-600 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-lg transition-colors"
                  >
                    上一步
                  </button>
                  <button
                    @click="nextStep"
                    class="px-4 py-1.5 text-sm font-medium text-white rounded-lg transition-all"
                    :class="isLastStep
                      ? 'bg-gradient-to-r from-green-500 to-emerald-500 hover:from-green-600 hover:to-emerald-600 shadow-md shadow-green-500/20'
                      : 'bg-gradient-to-r from-primary-500 to-primary-600 hover:from-primary-600 hover:to-primary-700 shadow-md shadow-primary-500/20'"
                  >
                    {{ isLastStep ? '开始学习' : '下一步' }}
                    <span v-if="!isLastStep" class="ml-1">&rarr;</span>
                    <span v-else class="ml-1">&#10003;</span>
                  </button>
                </div>
              </div>
            </div>

            <!-- 箭头指示器 -->
            <div
              class="absolute w-3 h-3 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 transform rotate-45"
              :style="arrowStyle"
            ></div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useOnboarding } from '../composables/useOnboarding'

const {
  isActive,
  currentStep,
  currentStepData,
  totalSteps,
  isFirstStep,
  isLastStep,
  nextStep,
  prevStep,
  skipOnboarding,
} = useOnboarding()

const spotlightRect = ref({ x: 0, y: 0, width: 0, height: 0, radius: 8 })
const tooltipPos = ref({ x: 0, y: 0 })
const arrowDirection = ref('left')

const stepIcons = [
  '<svg viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5"><path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM6.293 6.707a1 1 0 010-1.414l3-3a1 1 0 011.414 0l3 3a1 1 0 01-1.414 1.414L11 5.414V13a1 1 0 11-2 0V5.414L7.707 6.707a1 1 0 01-1.414 0z" clip-rule="evenodd"/></svg>',
  '<svg viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5"><path d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zm-2 4a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"/></svg>',
  '<svg viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5"><path fill-rule="evenodd" d="M18 10c0 3.866-3.582 7-8 7a8.841 8.841 0 01-4.083-.98L2 17l1.338-3.123C2.493 12.767 2 11.434 2 10c0-3.866 3.582-7 8-7s8 3.134 8 7zM7 9H5v2h2V9zm8 0h-2v2h2V9zm-4 0H9v2h2V9z" clip-rule="evenodd"/></svg>',
]

const stepIconClasses = [
  'bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400',
  'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400',
  'bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400',
]

const stepIcon = computed(() => stepIcons[currentStep.value] || stepIcons[0])
const stepIconClass = computed(() => stepIconClasses[currentStep.value] || stepIconClasses[0])

const tooltipTransition = computed(() => {
  if (arrowDirection.value === 'left' || arrowDirection.value === 'right') {
    return 'onboarding-slide-y'
  }
  return 'onboarding-slide-x'
})

const tooltipStyle = computed(() => ({
  left: tooltipPos.value.x + 'px',
  top: tooltipPos.value.y + 'px',
}))

const arrowStyle = computed(() => {
  const base = {}
  switch (arrowDirection.value) {
    case 'left':
      base.left = '-7px'
      base.top = '20px'
      base.borderRight = 'none'
      base.borderTop = 'none'
      break
    case 'right':
      base.right = '-7px'
      base.top = '20px'
      base.borderLeft = 'none'
      base.borderBottom = 'none'
      break
    case 'top':
      base.top = '-7px'
      base.left = '24px'
      base.borderBottom = 'none'
      base.borderRight = 'none'
      break
    case 'bottom':
      base.bottom = '-7px'
      base.left = '24px'
      base.borderTop = 'none'
      base.borderLeft = 'none'
      break
  }
  return base
})

function updatePosition() {
  if (!isActive.value || !currentStepData.value) return

  const selector = currentStepData.value.target
  const el = document.querySelector(selector)

  if (!el) {
    // Fallback: center of screen
    spotlightRect.value = {
      x: window.innerWidth / 2 - 100,
      y: window.innerHeight / 2 - 30,
      width: 200,
      height: 60,
      radius: 8,
    }
    tooltipPos.value = {
      x: window.innerWidth / 2 - 180,
      y: window.innerHeight / 2 + 50,
    }
    arrowDirection.value = 'top'
    return
  }

  const rect = el.getBoundingClientRect()
  const padding = 8
  const tooltipWidth = 360
  const tooltipHeight = 260

  spotlightRect.value = {
    x: rect.left - padding,
    y: rect.top - padding,
    width: rect.width + padding * 2,
    height: rect.height + padding * 2,
    radius: 10,
  }

  // Determine tooltip position based on preferred direction
  const preferred = currentStepData.value.position || 'right'
  const vw = window.innerWidth
  const vh = window.innerHeight

  if (preferred === 'right' && rect.right + tooltipWidth + 24 < vw) {
    tooltipPos.value = {
      x: rect.right + padding + 12,
      y: Math.max(16, Math.min(rect.top - padding, vh - tooltipHeight - 16)),
    }
    arrowDirection.value = 'left'
  } else if (preferred === 'left' && rect.left - tooltipWidth - 24 > 0) {
    tooltipPos.value = {
      x: rect.left - padding - 12 - tooltipWidth,
      y: Math.max(16, Math.min(rect.top - padding, vh - tooltipHeight - 16)),
    }
    arrowDirection.value = 'right'
  } else if (rect.bottom + tooltipHeight + 24 < vh) {
    tooltipPos.value = {
      x: Math.max(16, Math.min(rect.left - 20, vw - tooltipWidth - 16)),
      y: rect.bottom + padding + 12,
    }
    arrowDirection.value = 'top'
  } else {
    tooltipPos.value = {
      x: Math.max(16, Math.min(rect.left - 20, vw - tooltipWidth - 16)),
      y: Math.max(16, rect.top - padding - 12 - tooltipHeight),
    }
    arrowDirection.value = 'bottom'
  }
}

watch([isActive, currentStep], () => {
  nextTick(() => {
    updatePosition()
  })
})

function handleResize() {
  updatePosition()
}

function handleKeydown(e) {
  if (!isActive.value) return
  if (e.key === 'Escape') {
    skipOnboarding()
  } else if (e.key === 'ArrowRight' || e.key === 'Enter') {
    nextStep()
  } else if (e.key === 'ArrowLeft') {
    prevStep()
  }
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
  document.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  document.removeEventListener('keydown', handleKeydown)
})

// Expose for parent to trigger start
defineExpose({ updatePosition })
</script>

<style scoped>
.onboarding-fade-enter-active,
.onboarding-fade-leave-active {
  transition: opacity 0.3s ease;
}
.onboarding-fade-enter-from,
.onboarding-fade-leave-to {
  opacity: 0;
}

.onboarding-slide-y-enter-active,
.onboarding-slide-y-leave-active {
  transition: all 0.3s ease;
}
.onboarding-slide-y-enter-from {
  opacity: 0;
  transform: translateY(12px);
}
.onboarding-slide-y-leave-to {
  opacity: 0;
  transform: translateY(-12px);
}

.onboarding-slide-x-enter-active,
.onboarding-slide-x-leave-active {
  transition: all 0.3s ease;
}
.onboarding-slide-x-enter-from {
  opacity: 0;
  transform: translateX(12px);
}
.onboarding-slide-x-leave-to {
  opacity: 0;
  transform: translateX(-12px);
}
</style>

<style>
@keyframes onboarding-pulse {
  0%, 100% {
    box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.2), 0 0 20px rgba(99, 102, 241, 0.15);
  }
  50% {
    box-shadow: 0 0 0 6px rgba(99, 102, 241, 0.3), 0 0 30px rgba(99, 102, 241, 0.25);
  }
}
</style>
