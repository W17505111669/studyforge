import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, register as apiRegister } from '../api/client'

/**
 * 解码 JWT payload 并检查是否过期
 * @param {string} token - JWT 字符串
 * @returns {{ valid: boolean, payload: object|null }}
 */
function decodeAndCheckToken(token) {
  if (!token) return { valid: false, payload: null }
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return { valid: false, payload: null }
    const payload = JSON.parse(atob(parts[1]))
    // 检查 exp 字段（秒级时间戳），预留 30s 缓冲
    if (payload.exp && payload.exp * 1000 < Date.now() + 30000) {
      return { valid: false, payload }
    }
    return { valid: true, payload }
  } catch {
    return { valid: false, payload: null }
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')

  // 安全解析 localStorage 中的 user 数据（防止 JSON 格式错误导致应用崩溃）
  let storedUser = null
  try {
    storedUser = JSON.parse(localStorage.getItem('user') || 'null')
  } catch {
    localStorage.removeItem('user')
  }
  const user = ref(storedUser)

  // 解码 JWT 并检查过期，而非仅判断 token 是否存在
  const isLoggedIn = computed(() => {
    const { valid } = decodeAndCheckToken(token.value)
    if (!valid && token.value) {
      // Token 已过期或格式异常，自动清理
      token.value = ''
      user.value = null
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    }
    return valid
  })

  async function login(username, password) {
    const res = await apiLogin({ username, password })
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data.user))
  }

  async function register(username, password, email) {
    const res = await apiRegister({ username, password, email })
    // 后端 register 接口已直接返回 token + user，无需再用密码二次登录
    token.value = res.data.token
    user.value = res.data.user
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('user', JSON.stringify(res.data.user))
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return { token, user, isLoggedIn, login, register, logout }
})
