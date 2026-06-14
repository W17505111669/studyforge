import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

// 请求拦截器：自动附加 JWT Token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：401 自动清除凭证并跳登录
api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

// ========== Auth ==========
export const register = (data) => api.post('/register', data)
export const login = (data) => api.post('/login', data)

// ========== Materials ==========
export const uploadMaterial = (data) => api.post('/materials', data)
export const uploadFile = (file, onProgress) => {
  const formData = new FormData()
  formData.append('file', file)
  const config = {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 60000, // 大文件上传超时 60s
  }
  if (typeof onProgress === 'function') {
    config.onUploadProgress = (progressEvent) => {
      if (progressEvent.total) {
        onProgress(Math.round((progressEvent.loaded * 100) / progressEvent.total))
      }
    }
  }
  return api.post('/materials/upload', formData, config)
}
export const listMaterials = (params) => api.get('/materials', { params })
export const getMaterial = (id) => api.get(`/materials/${id}`)
export const deleteMaterial = (id) => api.delete(`/materials/${id}`)
export const analyzeMaterial = (id) => api.post(`/materials/${id}/analyze`)
export const getMaterialStatus = (id) => api.get(`/materials/${id}/status`)
export const batchAnalyzeMaterials = (ids) => api.post('/materials/batch-analyze', { ids })
export const batchDeleteMaterials = (ids) => api.delete('/materials/batch', { data: { ids } })
export const getTags = () => api.get('/tags')

// ========== Cards ==========
export const listCards = (params) => api.get('/cards', { params })
export const getCard = (id) => api.get(`/cards/${id}`)
export const reviewCard = (id, result) => api.post(`/cards/${id}/review`, { result })
export const toggleBookmark = (id) => api.put(`/cards/${id}/bookmark`)
export const updateCardNote = (id, note) => api.put(`/cards/${id}/note`, { note })

// ========== Quizzes ==========
export const listQuizzes = (params) => api.get('/quizzes', { params })
export const answerQuiz = (id, data) => api.post(`/quizzes/${id}/answer`, data)
export const getQuizHint = (id, level = 1) => api.get(`/quizzes/${id}/hint`, { params: { level } })
export const getDifficultyLevel = () => api.get('/quizzes/difficulty-level')

// ========== Mistakes (错题本) ==========
export const listMistakes = (params) => api.get('/mistakes', { params })
export const getMistakeStats = () => api.get('/mistakes/stats')
export const reviewMistake = (id) => api.post(`/mistakes/${id}/review`)
export const batchReviewMistakes = (ids) => api.post('/mistakes/batch-review', { ids })
export const deleteMistake = (id) => api.delete(`/mistakes/${id}`)
export const retryMistakes = () => api.post('/mistakes/retry')
export const consolidatePractice = () => api.post('/mistakes/consolidate', {}, { timeout: 60000 })

// ========== Chat ==========
export const chat = (data) => api.post('/chat', data)

// ========== Conversations ==========
export const listConversations = (params) => api.get('/conversations', { params })
export const getConversation = (id) => api.get(`/conversations/${id}`)
export const createConversation = (title) => api.post('/conversations', { title })
export const updateConversation = (id, title) => api.put(`/conversations/${id}`, { title })
export const deleteConversation = (id) => api.delete(`/conversations/${id}`)

/**
 * 流式对话（SSE 打字机效果 + Function Calling 工具事件）
 * @param {string} message - 用户消息
 * @param {string} materialId - 可选的材料 ID
 * @param {function} onToken - 每收到一个 token 时的回调 (token: string) => void
 * @param {function} onDone - 流结束时的回调 (fullText: string) => void
 * @param {function} onError - 错误回调 (error: Error) => void
 * @param {function} onToolEvent - 工具调用事件回调 ({type, name, args?, result?}) => void
 * @param {string} conversationId - 可选的对话会话 ID
 * @param {function} onConvId - 收到 conversation_id 时的回调 (convId: string) => void
 * @returns {AbortController} - 可用于取消请求的控制器
 */
export function chatStream(message, materialId, onToken, onDone, onError, onToolEvent, conversationId, onConvId) {
  const controller = new AbortController()
  const token = localStorage.getItem('token')
  const params = new URLSearchParams({ message })
  if (materialId) params.append('material_id', materialId)
  if (conversationId) params.append('conversation_id', conversationId)

  fetch(`/api/chat/stream?${params.toString()}`, {
    headers: { Authorization: `Bearer ${token}` },
    signal: controller.signal,
  })
    .then(async (response) => {
      if (!response.ok) {
        // SSE fetch 绕过 axios 拦截器，需手动处理 401 自动登出
        if (response.status === 401) {
          localStorage.removeItem('token')
          localStorage.removeItem('user')
          window.location.href = '/login'
          return
        }
        throw new Error(`HTTP ${response.status}`)
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let fullText = ''
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })

        // 按行解析 SSE 格式: "data: xxx\n\n"
        const lines = buffer.split('\n')
        buffer = lines.pop() // 保留未完成的行

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const content = line.slice(6)
            if (content === '[DONE]') {
              if (onDone) onDone(fullText)
              return
            }
            // 解析 conversation_id 事件
            if (content.startsWith('[CONV_ID]')) {
              if (onConvId) {
                try {
                  const data = JSON.parse(content.slice(9))
                  onConvId(data.conversation_id)
                } catch {}
              }
              continue
            }
            // 解析工具调用事件
            if (content.startsWith('[TOOL_CALL]')) {
              if (onToolEvent) {
                try { onToolEvent({ type: 'tool_call', ...JSON.parse(content.slice(11)) }) } catch {}
              }
              continue
            }
            if (content.startsWith('[TOOL_RESULT]')) {
              if (onToolEvent) {
                try { onToolEvent({ type: 'tool_result', ...JSON.parse(content.slice(13)) }) } catch {}
              }
              continue
            }
            fullText += content
            if (onToken) onToken(content)
          }
        }
      }

      if (onDone) onDone(fullText)
    })
    .catch((err) => {
      if (err.name !== 'AbortError' && onError) {
        onError(err)
      }
    })

  return controller
}

// ========== Search ==========
export const globalSearch = (q) => api.get('/search', { params: { q } })

// ========== Graph ==========
export const getKnowledgeGraph = (materialId) => api.get(`/graph/${materialId}`)
export const getAllGraphs = () => api.get('/graph/all')

// ========== Stats ==========
export const getUserStats = () => api.get('/stats')
export const getCalendarHeatmap = (year) => api.get('/stats/calendar', { params: { year } })

// ========== Achievements ==========
export const getAchievements = () => api.get('/achievements')

// ========== Recommendations ==========
export const getRecommendations = () => api.get('/recommendations')

// ========== Learning Path ==========
export const getLearningPath = (force = false) => api.get('/learning-path', {
  params: force ? { force: 'true' } : {},
  timeout: 120000,
})

// ========== Debate (多 Agent 辩论) ==========
export const startDebate = (concept) => api.post('/debate', { concept }, { timeout: 120000 })

// ========== Notifications ==========
export const getNotifications = () => api.get('/notifications')
export const getUnreadNotificationCount = () => api.get('/notifications/unread-count')
export const readNotification = (id) => api.post(`/notifications/${id}/read`)
export const readAllNotifications = () => api.post('/notifications/read-all')

// ========== Dashboard ==========
export const getMetrics = () => api.get('/dashboard/metrics')
export const getDailyActivity = () => api.get('/dashboard/activity')
export const listTraces = () => api.get('/dashboard/traces')

export default api
