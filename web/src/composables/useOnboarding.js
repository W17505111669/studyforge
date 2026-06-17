import { ref, computed } from 'vue'

const STORAGE_KEY = 'studyforge-onboarding-done'

// 模块级单例状态，保证所有组件共享同一份状态
const isActive = ref(false)
const currentStep = ref(0)

const steps = [
  {
    target: '[data-onboarding="upload"]',
    title: '上传学习材料',
    description: '点击这里上传 PDF、文档或网页链接，AI 将自动分析并提取知识点、生成卡片和练习题。',
    position: 'right'
  },
  {
    target: '[data-onboarding="cards"]',
    title: '复习知识卡片',
    description: '系统自动生成的智能卡片，支持间隔重复复习（SM-2 算法），帮你高效记忆。还能搜索和导出 Anki。',
    position: 'right'
  },
  {
    target: '[data-onboarding="chat"]',
    title: 'AI 智能对话',
    description: '基于你的学习材料，与 AI 进行深度对话。支持工具调用，可以实时查询知识图谱和卡片。',
    position: 'right'
  }
]

const totalSteps = computed(() => steps.length)
const isFirstStep = computed(() => currentStep.value === 0)
const isLastStep = computed(() => currentStep.value === steps.length - 1)
const currentStepData = computed(() => steps[currentStep.value])

function startOnboarding() {
  if (localStorage.getItem(STORAGE_KEY)) return false
  currentStep.value = 0
  isActive.value = true
  return true
}

function nextStep() {
  if (isLastStep.value) {
    finishOnboarding()
  } else {
    currentStep.value++
  }
}

function prevStep() {
  if (!isFirstStep.value) {
    currentStep.value--
  }
}

function finishOnboarding() {
  isActive.value = false
  localStorage.setItem(STORAGE_KEY, '1')
}

function skipOnboarding() {
  finishOnboarding()
}

function resetOnboarding() {
  localStorage.removeItem(STORAGE_KEY)
  currentStep.value = 0
  isActive.value = true
}

export function useOnboarding() {
  return {
    isActive,
    currentStep,
    steps,
    totalSteps,
    isFirstStep,
    isLastStep,
    currentStepData,
    startOnboarding,
    nextStep,
    prevStep,
    finishOnboarding,
    skipOnboarding,
    resetOnboarding
  }
}
