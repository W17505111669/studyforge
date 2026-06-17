<template>
  <div
    class="min-h-screen bg-gradient-to-br from-dark-900 via-dark-800 to-dark-900 flex items-center justify-center p-4"
  >
    <div class="w-full max-w-md">
      <!-- Logo -->
      <div class="text-center mb-8">
        <div
          class="inline-flex items-center justify-center w-16 h-16 bg-primary-500 rounded-2xl mb-4 shadow-lg shadow-primary-500/30"
        >
          <span class="text-2xl font-bold text-white">SF</span>
        </div>
        <h1 class="text-3xl font-bold text-white">StudyForge Pro</h1>
        <p class="text-gray-400 mt-2">AI 智能学习平台 · 多 Agent 并行分析</p>
      </div>

      <!-- 表单卡片 -->
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-8">
        <h2 class="text-xl font-bold text-gray-900 dark:text-gray-100 mb-6">
          {{ isLogin ? '欢迎回来' : '创建账户' }}
        </h2>

        <form class="space-y-4" @submit.prevent="handleSubmit">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">用户名</label>
            <input
              v-model="form.username"
              type="text"
              required
              class="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 dark:placeholder-gray-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-all"
              placeholder="输入用户名"
            />
          </div>

          <div v-if="!isLogin">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">邮箱</label>
            <input
              v-model="form.email"
              type="email"
              required
              class="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 dark:placeholder-gray-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-all"
              placeholder="your@email.com"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">密码</label>
            <input
              v-model="form.password"
              type="password"
              required
              class="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200 dark:placeholder-gray-400 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-all"
              placeholder="至少 6 位"
              minlength="6"
            />
          </div>

          <div v-if="error" class="text-red-500 text-sm bg-red-50 dark:bg-red-900/20 p-3 rounded-lg">
            {{ error }}
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="w-full py-3 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ loading ? '处理中...' : isLogin ? '登 录' : '注 册' }}
          </button>
        </form>

        <div class="mt-6 text-center text-sm text-gray-500 dark:text-gray-400">
          <span v-if="isLogin">还没有账户？</span>
          <span v-else>已有账户？</span>
          <button class="text-primary-600 hover:text-primary-700 font-medium ml-1" @click="toggleMode">
            {{ isLogin ? '立即注册' : '去登录' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

const isLogin = ref(true)
const loading = ref(false)
const error = ref('')
const form = reactive({
  username: '',
  password: '',
  email: ''
})

function toggleMode() {
  isLogin.value = !isLogin.value
  error.value = ''
  form.username = ''
  form.password = ''
  form.email = ''
}

async function handleSubmit() {
  loading.value = true
  error.value = ''

  try {
    if (isLogin.value) {
      await auth.login(form.username, form.password)
    } else {
      await auth.register(form.username, form.password, form.email)
    }
    // 登录/注册成功后立即清除表单中的敏感数据
    form.password = ''
    form.email = ''
    router.push('/')
  } catch (err) {
    error.value = err.response?.data?.error || '操作失败，请重试'
  } finally {
    loading.value = false
  }
}
</script>
