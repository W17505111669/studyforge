import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' }
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
    timeout: 60000 // 大文件上传超时 60s
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

// ========== Market (材料市场) ==========
export const toggleShare = (id) => api.put(`/materials/${id}/share`)
export const listMarketMaterials = (params) => api.get('/market/materials', { params })
export const previewMarketMaterial = (shareCode) => api.get(`/market/materials/${shareCode}`)
export const collectMarketMaterial = (shareCode) => api.post(`/market/materials/${shareCode}/collect`)
export const getMarketTags = () => api.get('/market/tags')

// ========== Decks (卡片组/牌组) ==========
export const listDecks = () => api.get('/decks')
export const createDeck = (data) => api.post('/decks', data)
export const getDeck = (id) => api.get(`/decks/${id}`)
export const deleteDeck = (id) => api.delete(`/decks/${id}`)
export const toggleDeckShare = (id) => api.put(`/decks/${id}/share`)
export const listMarketDecks = (params) => api.get('/market/decks', { params })
export const previewMarketDeck = (shareCode) => api.get(`/market/decks/${shareCode}`)
export const collectMarketDeck = (shareCode) => api.post(`/market/decks/${shareCode}/collect`)
export const getMarketDeckTags = () => api.get('/market/decks/tags')

// ========== Cards ==========
export const listCards = (params) => api.get('/cards', { params })
export const getCard = (id) => api.get(`/cards/${id}`)
export const reviewCard = (id, result) => api.post(`/cards/${id}/review`, { result })
export const toggleBookmark = (id) => api.put(`/cards/${id}/bookmark`)
export const updateCardNote = (id, note) => api.put(`/cards/${id}/note`, { note })
export const getDueCards = (params) => api.get('/cards/due', { params })

// ========== Images (卡片图片) ==========
export const uploadImage = (file) => {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/images/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 30000
  })
}

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
    signal: controller.signal
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
                try {
                  onToolEvent({ type: 'tool_call', ...JSON.parse(content.slice(11)) })
                } catch {}
              }
              continue
            }
            if (content.startsWith('[TOOL_RESULT]')) {
              if (onToolEvent) {
                try {
                  onToolEvent({ type: 'tool_result', ...JSON.parse(content.slice(13)) })
                } catch {}
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
export const globalSearch = (q, filters = {}) => {
  const params = { q, ...filters }
  // 清理空值参数
  Object.keys(params).forEach((key) => {
    if (params[key] === '' || params[key] === null || params[key] === undefined) {
      delete params[key]
    }
  })
  return api.get('/search', { params })
}

// ========== Graph ==========
export const getKnowledgeGraph = (materialId) => api.get(`/graph/${materialId}`)
export const getAllGraphs = () => api.get('/graph/all')

// ========== Stats ==========
export const getUserStats = () => api.get('/stats')
export const getCalendarHeatmap = (year) => api.get('/stats/calendar', { params: { year } })

// ========== Export (数据导出) ==========
export const exportDataPreview = (params) => api.get('/export/preview', { params })
export const exportDataDownload = (params) => {
  const token = localStorage.getItem('token')
  const query = new URLSearchParams(params).toString()
  return fetch(`/api/export/data?${query}`, {
    headers: { Authorization: `Bearer ${token}` }
  }).then((res) => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    return res.blob()
  })
}

// ========== Achievements ==========
export const getAchievements = () => api.get('/achievements')

// ========== Recommendations ==========
export const getRecommendations = () => api.get('/recommendations')

// ========== Diagnosis (知识弱点诊断) ==========
export const getDiagnosis = () => api.get('/diagnosis', { timeout: 60000 })

// ========== Learning Path ==========
export const getLearningPath = (force = false) =>
  api.get('/learning-path', {
    params: force ? { force: 'true' } : {},
    timeout: 120000
  })

// ========== Debate (多 Agent 辩论) ==========
export const startDebate = (concept) => api.post('/debate', { concept }, { timeout: 120000 })

// ========== Notifications ==========
export const getNotifications = (params) => api.get('/notifications', { params })
export const getUnreadNotificationCount = () => api.get('/notifications/unread-count')
export const readNotification = (id) => api.post(`/notifications/${id}/read`)
export const readAllNotifications = () => api.post('/notifications/read-all')

// ========== Pomodoro (番茄钟) ==========
export const startPomodoro = (data) => api.post('/pomodoro/start', data)
export const endPomodoro = (data) => api.post('/pomodoro/end', data)
export const getPomodoroStats = () => api.get('/pomodoro/stats')

// ========== Goals (学习目标) ==========
export const listGoals = (status) => api.get('/goals', { params: status ? { status } : {} })
export const getGoalProgress = () => api.get('/goals/progress')
export const createGoal = (data) => api.post('/goals', data)
export const updateGoal = (id, data) => api.put(`/goals/${id}`, data)
export const deleteGoal = (id) => api.delete(`/goals/${id}`)

// ========== Reports (学习报告) ==========
export const getWeeklyReport = (date) => api.get('/reports/weekly', { params: date ? { date } : {} })
export const getMonthlyReport = (month) => api.get('/reports/monthly', { params: month ? { month } : {} })

// ========== Streaks (学习连续打卡) ==========
export const getStreaks = () => api.get('/streaks')

// ========== Leaderboard (学习排行榜) ==========
export const getLeaderboard = (period) => api.get('/leaderboard', { params: { period } })
export const getMyLeaderboardStats = (period) => api.get('/leaderboard/me', { params: { period } })

// ========== Friends (好友系统) ==========
export const listFriends = () => api.get('/friends')
export const sendFriendRequest = (username) => api.post('/friends/request', { username })
export const acceptFriendRequest = (id) => api.put(`/friends/request/${id}/accept`)
export const rejectFriendRequest = (id) => api.delete(`/friends/request/${id}`)
export const removeFriend = (id) => api.delete(`/friends/${id}`)
export const getFriendRequests = () => api.get('/friends/requests')
export const searchUsers = (q) => api.get('/friends/search', { params: { q } })
export const getFriendCount = () => api.get('/friends/count')

// ========== Daily Tasks (每日任务) ==========
export const getDailyTasks = (date) => api.get('/daily-tasks', { params: date ? { date } : {} })
export const createDailyTask = (data) => api.post('/daily-tasks', data)
export const updateDailyTask = (id, data) => api.put(`/daily-tasks/${id}`, data)
export const toggleDailyTask = (id) => api.put(`/daily-tasks/${id}/toggle`)
export const deleteDailyTask = (id) => api.delete(`/daily-tasks/${id}`)

// ========== Notes (知识笔记本) ==========
export const listNotes = (params = {}) => api.get('/notes', { params })
export const getNote = (id) => api.get(`/notes/${id}`)
export const createNote = (data) => api.post('/notes', data)
export const updateNote = (id, data) => api.put(`/notes/${id}`, data)
export const deleteNote = (id) => api.delete(`/notes/${id}`)
export const searchNotes = (q) => api.get('/notes/search', { params: { q } })
export const listNoteFolders = () => api.get('/notes/folders')
export const createNoteFolder = (data) => api.post('/notes/folders', data)
export const updateNoteFolder = (id, data) => api.put(`/notes/folders/${id}`, data)
export const deleteNoteFolder = (id) => api.delete(`/notes/folders/${id}`)

// ========== Anki Import (Anki 导入) ==========
export const importAnkiPreview = (file) => {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/import/anki', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 60000
  })
}
export const importAnkiConfirm = (data) => api.post('/import/anki/confirm', data, { timeout: 60000 })

// ========== Study Groups (学习小组) ==========
export const listGroups = (filter) => api.get('/groups', { params: filter ? { filter } : {} })
export const getGroup = (id) => api.get(`/groups/${id}`)
export const createGroup = (data) => api.post('/groups', data)
export const updateGroup = (id, data) => api.put(`/groups/${id}`, data)
export const deleteGroup = (id) => api.delete(`/groups/${id}`)
export const joinGroup = (id) => api.post(`/groups/${id}/join`)
export const leaveGroup = (id) => api.post(`/groups/${id}/leave`)
export const getGroupMembers = (id) => api.get(`/groups/${id}/members`)
export const getGroupProgress = (id) => api.get(`/groups/${id}/progress`)
export const getGroupGoals = (id) => api.get(`/groups/${id}/goals`)
export const createGroupGoal = (id, data) => api.post(`/groups/${id}/goals`, data)
export const deleteGroupGoal = (id, goalId) => api.delete(`/groups/${id}/goals/${goalId}`)

// ========== Exam (模拟考试) ==========
export const generateExam = (data) => api.post('/exam/generate', data, { timeout: 120000 })
export const submitExam = (id, answers) => api.post(`/exams/${id}/submit`, { answers }, { timeout: 30000 })
export const listExams = () => api.get('/exams')
export const getExam = (id) => api.get(`/exams/${id}`)

// ========== Explain (AI 概念解释器) ==========
export const explainConcept = (data) => api.post('/explain', data, { timeout: 60000 })
export const getExplainHistory = (params = {}) => api.get('/explain/history', { params })
export const deleteExplainCache = (id) => api.delete(`/explain/${id}`)

// ========== Insights (知识洞察) ==========
export const getConnections = () => api.get('/insights/connections')

// ========== Dashboard ==========
export const getMetrics = () => api.get('/dashboard/metrics')
export const getDailyActivity = () => api.get('/dashboard/activity')
export const listTraces = () => api.get('/dashboard/traces')

// ========== Admin (性能监控) ==========
export const getAPIMetrics = (rangeHours = 24) =>
  api.get('/admin/metrics', { params: { range: rangeHours } })
export const getDBStats = () => api.get('/admin/db-stats')

export default api
