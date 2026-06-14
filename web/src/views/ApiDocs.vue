<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-5xl mx-auto">
    <!-- 头部 -->
    <div class="mb-6">
      <div class="flex items-center gap-3 mb-2">
        <button @click="$router.back()" class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
        </button>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">API 接口文档</h1>
        <span class="px-2 py-0.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-medium rounded-full">v1.0</span>
      </div>
      <p class="text-sm text-gray-500 dark:text-gray-400 ml-8">StudyForge Pro 全部 REST API 端点参考。所有认证端点需在 Header 携带 <code class="text-xs bg-gray-100 dark:bg-gray-700 px-1 rounded">Authorization: Bearer &lt;token&gt;</code></p>
    </div>

    <!-- 搜索框 -->
    <div class="mb-6">
      <div class="relative">
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
        <input
          v-model="search"
          type="text"
          placeholder="搜索端点路径或描述..."
          class="w-full pl-10 pr-4 py-2.5 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none"
        />
      </div>
    </div>

    <!-- 目录 -->
    <div class="flex flex-wrap gap-2 mb-6">
      <button
        v-for="cat in categories"
        :key="cat.id"
        @click="activeCategory = activeCategory === cat.id ? '' : cat.id"
        class="px-3 py-1.5 rounded-full text-xs font-medium transition-colors"
        :class="activeCategory === cat.id
          ? 'bg-primary-600 text-white'
          : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700'"
      >
        {{ cat.label }}
        <span class="ml-1 opacity-60">{{ cat.count }}</span>
      </button>
    </div>

    <!-- 端点列表 -->
    <div class="space-y-3">
      <template v-for="group in filteredGroups" :key="group.category">
        <h2 class="text-lg font-semibold text-gray-800 dark:text-gray-200 pt-2 flex items-center gap-2">
          <span class="w-2 h-2 rounded-full" :class="group.color"></span>
          {{ group.label }}
        </h2>
        <div
          v-for="ep in group.endpoints"
          :key="ep.method + ep.path"
          class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden bg-white dark:bg-gray-800"
        >
          <!-- 端点头部（可点击展开） -->
          <button
            @click="toggle(ep.id)"
            class="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-gray-50 dark:hover:bg-gray-750 transition-colors"
          >
            <span
              class="px-2 py-0.5 rounded text-xs font-bold uppercase min-w-[56px] text-center shrink-0"
              :class="methodClass(ep.method)"
            >{{ ep.method }}</span>
            <code class="text-sm font-mono text-gray-800 dark:text-gray-200 flex-1">{{ ep.path }}</code>
            <span class="text-sm text-gray-500 dark:text-gray-400 hidden sm:inline">{{ ep.summary }}</span>
            <svg
              class="w-4 h-4 text-gray-400 transition-transform shrink-0"
              :class="{ 'rotate-180': openSet.has(ep.id) }"
              fill="none" stroke="currentColor" viewBox="0 0 24 24"
            ><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
          </button>
          <!-- 展开详情 -->
          <div v-if="openSet.has(ep.id)" class="border-t border-gray-100 dark:border-gray-700 px-4 py-3 space-y-3 bg-gray-50/50 dark:bg-gray-800/50">
            <p class="text-sm text-gray-600 dark:text-gray-400">{{ ep.description }}</p>
            <!-- 认证标记 -->
            <div v-if="ep.auth" class="inline-flex items-center gap-1 px-2 py-0.5 bg-amber-50 dark:bg-amber-900/20 text-amber-700 dark:text-amber-400 text-xs rounded-full">
              <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd"/></svg>
              需要认证
            </div>
            <span v-else class="inline-flex items-center gap-1 px-2 py-0.5 bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 text-xs rounded-full">公开</span>
            <!-- 参数 -->
            <div v-if="ep.params && ep.params.length">
              <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">参数</h4>
              <div class="overflow-x-auto">
                <table class="w-full text-sm">
                  <thead><tr class="text-left text-xs text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700">
                    <th class="pb-1 pr-4">名称</th><th class="pb-1 pr-4">位置</th><th class="pb-1 pr-4">类型</th><th class="pb-1">说明</th>
                  </tr></thead>
                  <tbody>
                    <tr v-for="p in ep.params" :key="p.name" class="border-b border-gray-100 dark:border-gray-700/50">
                      <td class="py-1.5 pr-4"><code class="text-xs bg-gray-100 dark:bg-gray-700 px-1 rounded text-primary-600 dark:text-primary-400">{{ p.name }}</code></td>
                      <td class="py-1.5 pr-4 text-xs text-gray-500 dark:text-gray-400">{{ p.in }}</td>
                      <td class="py-1.5 pr-4 text-xs text-gray-500 dark:text-gray-400">{{ p.type }}</td>
                      <td class="py-1.5 text-gray-600 dark:text-gray-400">{{ p.desc }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
            <!-- 响应 -->
            <div v-if="ep.response">
              <h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">响应示例</h4>
              <pre class="text-xs bg-gray-900 dark:bg-gray-950 text-green-400 p-3 rounded-lg overflow-x-auto leading-relaxed">{{ ep.response }}</pre>
            </div>
          </div>
        </div>
      </template>

      <!-- 空状态 -->
      <div v-if="filteredGroups.length === 0" class="text-center py-12 text-gray-400 dark:text-gray-500">
        <svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
        <p>没有找到匹配的端点</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, reactive } from 'vue'

const search = ref('')
const activeCategory = ref('')
const openSet = reactive(new Set())

function toggle(id) {
  openSet.has(id) ? openSet.delete(id) : openSet.add(id)
}

function methodClass(m) {
  const map = {
    GET: 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400',
    POST: 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400',
    PUT: 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400',
    DELETE: 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400',
  }
  return map[m] || 'bg-gray-100 text-gray-600'
}

const endpointGroups = [
  {
    id: 'auth', category: 'auth', label: '认证', color: 'bg-green-500',
    endpoints: [
      { id: 'register', method: 'POST', path: '/api/register', summary: '用户注册', auth: false,
        description: '注册新用户账号，密码使用 bcrypt 加密存储。',
        params: [
          { name: 'username', in: 'body', type: 'string', desc: '用户名（必填）' },
          { name: 'email', in: 'body', type: 'string', desc: '邮箱（必填）' },
          { name: 'password', in: 'body', type: 'string', desc: '密码（必填）' },
        ],
        response: '{ "token": "eyJhbGciOi...", "user": { "id": "uuid", "username": "demo", "email": "demo@example.com" } }'
      },
      { id: 'login', method: 'POST', path: '/api/login', summary: '用户登录', auth: false,
        description: '验证用户名密码，返回 JWT Token。',
        params: [
          { name: 'username', in: 'body', type: 'string', desc: '用户名' },
          { name: 'password', in: 'body', type: 'string', desc: '密码' },
        ],
        response: '{ "token": "eyJhbGciOi...", "user": { "id": "uuid", "username": "demo" } }'
      },
      { id: 'health', method: 'GET', path: '/api/health', summary: '健康检查', auth: false,
        description: '检查服务运行状态。',
        response: '{ "status": "ok", "service": "StudyForge Pro" }'
      },
    ]
  },
  {
    id: 'materials', category: 'materials', label: '材料管理', color: 'bg-blue-500',
    endpoints: [
      { id: 'upload-material', method: 'POST', path: '/api/materials', summary: '创建文本材料', auth: true,
        description: '创建一条文本或 URL 类型的学习材料。',
        params: [
          { name: 'title', in: 'body', type: 'string', desc: '材料标题（必填）' },
          { name: 'content_type', in: 'body', type: 'string', desc: '类型：text 或 url' },
          { name: 'content', in: 'body', type: 'string', desc: '文本内容' },
          { name: 'source_url', in: 'body', type: 'string', desc: '来源 URL' },
        ],
        response: '{ "id": "uuid", "title": "材料标题", "status": "pending", ... }'
      },
      { id: 'upload-file', method: 'POST', path: '/api/materials/upload', summary: '上传文件材料', auth: true,
        description: '上传 PDF/DOCX/MD/TXT 文件（最大 20MB），自动提取文本内容。',
        params: [
          { name: 'file', in: 'form-data', type: 'file', desc: '上传文件' },
          { name: 'title', in: 'form-data', type: 'string', desc: '材料标题' },
        ],
      },
      { id: 'list-materials', method: 'GET', path: '/api/materials', summary: '材料列表', auth: true,
        description: '分页获取当前用户的学习材料列表。',
        params: [
          { name: 'limit', in: 'query', type: 'int', desc: '每页条数（默认 20，上限 200）' },
          { name: 'offset', in: 'query', type: 'int', desc: '偏移量（默认 0）' },
        ],
        response: '{ "data": [...], "total": 10, "limit": 20, "offset": 0 }'
      },
      { id: 'get-material', method: 'GET', path: '/api/materials/:id', summary: '材料详情', auth: true,
        description: '获取单个材料详情，含关联的知识卡片和练习题。',
        params: [{ name: 'id', in: 'path', type: 'string', desc: '材料 ID' }],
        response: '{ "id": "uuid", "title": "...", "cards": [...], "quizzes": [...], "analysis_data": "..." }'
      },
      { id: 'delete-material', method: 'DELETE', path: '/api/materials/:id', summary: '删除材料', auth: true,
        description: '删除材料及其关联的卡片、练习题和 RAG 向量数据。',
        params: [{ name: 'id', in: 'path', type: 'string', desc: '材料 ID' }],
        response: '{ "message": "材料已删除" }'
      },
      { id: 'analyze-material', method: 'POST', path: '/api/materials/:id/analyze', summary: '触发 AI 分析', auth: true,
        description: '触发 4 个 Agent 并发分析材料（分析师/出题师/卡片师/图谱师），通过 WebSocket 推送进度。',
        params: [{ name: 'id', in: 'path', type: 'string', desc: '材料 ID' }],
        response: '{ "message": "分析已启动" }'
      },
      { id: 'material-status', method: 'GET', path: '/api/materials/:id/status', summary: '分析进度轮询', auth: true,
        description: '轮询获取材料分析进度（WebSocket 断线时的 fallback 方案）。',
        params: [{ name: 'id', in: 'path', type: 'string', desc: '材料 ID' }],
        response: '{ "agents": [{ "name": "Analyst", "status": "done" }, ...], "started": "..." }'
      },
    ]
  },
  {
    id: 'cards', category: 'cards', label: '知识卡片', color: 'bg-purple-500',
    endpoints: [
      { id: 'list-cards', method: 'GET', path: '/api/cards', summary: '卡片列表', auth: true,
        description: '分页获取知识卡片，支持按材料、难度、待复习状态过滤。',
        params: [
          { name: 'limit', in: 'query', type: 'int', desc: '每页条数（默认 20）' },
          { name: 'offset', in: 'query', type: 'int', desc: '偏移量' },
          { name: 'material_id', in: 'query', type: 'string', desc: '按材料 ID 过滤' },
          { name: 'difficulty', in: 'query', type: 'string', desc: '按难度过滤：easy/medium/hard' },
          { name: 'due', in: 'query', type: 'bool', desc: 'true 时只返回待复习卡片' },
        ],
        response: '{ "data": [...], "total": 25, "limit": 20, "offset": 0 }'
      },
      { id: 'get-card', method: 'GET', path: '/api/cards/:id', summary: '卡片详情', auth: true,
        description: '获取单张知识卡片详情。',
        params: [{ name: 'id', in: 'path', type: 'string', desc: '卡片 ID' }],
      },
      { id: 'review-card', method: 'POST', path: '/api/cards/:id/review', summary: '复习卡片（SM-2）', auth: true,
        description: '提交复习结果，基于 SM-2 算法更新间隔天数和难度因子。',
        params: [
          { name: 'id', in: 'path', type: 'string', desc: '卡片 ID' },
          { name: 'result', in: 'body', type: 'string', desc: '"mastered"（已掌握）或 "review"（再复习）' },
        ],
        response: '{ "id": "uuid", "review_count": 3, "ease_factor": 2.6, "interval_days": 7, "next_review_at": "..." }'
      },
      { id: 'export-cards', method: 'GET', path: '/api/cards/export', summary: '导出 Anki CSV', auth: true,
        description: '导出知识卡片为 Anki 兼容的 tab-separated CSV 文件（含 UTF-8 BOM）。',
        params: [
          { name: 'material_id', in: 'query', type: 'string', desc: '按材料过滤' },
          { name: 'difficulty', in: 'query', type: 'string', desc: '按难度过滤' },
        ],
        response: '(二进制文件流，Content-Disposition: attachment)'
      },
    ]
  },
  {
    id: 'quizzes', category: 'quizzes', label: '练习题', color: 'bg-orange-500',
    endpoints: [
      { id: 'list-quizzes', method: 'GET', path: '/api/quizzes', summary: '题目列表', auth: true,
        description: '分页获取练习题，支持按材料和难度过滤。',
        params: [
          { name: 'limit', in: 'query', type: 'int', desc: '每页条数（默认 20）' },
          { name: 'offset', in: 'query', type: 'int', desc: '偏移量' },
          { name: 'material_id', in: 'query', type: 'string', desc: '按材料 ID 过滤' },
          { name: 'difficulty', in: 'query', type: 'string', desc: '按难度过滤' },
        ],
        response: '{ "data": [...], "total": 15, "limit": 20, "offset": 0 }'
      },
      { id: 'answer-quiz', method: 'POST', path: '/api/quizzes/:id/answer', summary: '提交答案', auth: true,
        description: '提交练习题答案，自动判断正确性并记录答题记录。',
        params: [
          { name: 'id', in: 'path', type: 'string', desc: '题目 ID' },
          { name: 'answer', in: 'body', type: 'string', desc: '用户答案（必填）' },
        ],
        response: '{ "is_correct": true, "correct_answer": "..." }'
      },
    ]
  },
  {
    id: 'chat', category: 'chat', label: 'AI 对话', color: 'bg-indigo-500',
    endpoints: [
      { id: 'chat', method: 'POST', path: '/api/chat', summary: '对话（非流式）', auth: true,
        description: '发送消息并获取 AI 回复，支持 Function Calling 工具调用。',
        params: [
          { name: 'message', in: 'body', type: 'string', desc: '用户消息' },
          { name: 'conversation_id', in: 'body', type: 'string', desc: '会话 ID（不传则自动创建）' },
        ],
        response: '{ "reply": "AI 回复内容", "conversation_id": "uuid" }'
      },
      { id: 'chat-stream', method: 'GET', path: '/api/chat/stream', summary: '对话（SSE 流式）', auth: true,
        description: 'SSE 流式对话，打字机效果逐字输出。支持 [TOOL_CALL]、[TOOL_RESULT]、[CONV_ID] 等特殊事件。',
        params: [
          { name: 'message', in: 'query', type: 'string', desc: '用户消息' },
          { name: 'conversation_id', in: 'query', type: 'string', desc: '会话 ID' },
        ],
        response: 'text/event-stream\ndata: [CONV_ID]{"conversation_id":"..."}\ndata: 你\ndata: 好\ndata: [DONE]'
      },
    ]
  },
  {
    id: 'conversations', category: 'conversations', label: '会话管理', color: 'bg-teal-500',
    endpoints: [
      { id: 'list-convs', method: 'GET', path: '/api/conversations', summary: '会话列表', auth: true,
        description: '分页获取对话列表，含消息数量统计。',
        params: [
          { name: 'limit', in: 'query', type: 'int', desc: '每页条数（默认 20）' },
          { name: 'offset', in: 'query', type: 'int', desc: '偏移量' },
        ],
        response: '{ "data": [{ "id": "uuid", "title": "...", "message_count": 5 }], "total": 3 }'
      },
      { id: 'get-conv', method: 'GET', path: '/api/conversations/:id', summary: '会话详情', auth: true,
        description: '获取单个会话详情，含完整消息列表。',
        params: [{ name: 'id', in: 'path', type: 'string', desc: '会话 ID' }],
      },
      { id: 'create-conv', method: 'POST', path: '/api/conversations', summary: '创建会话', auth: true,
        description: '创建一个新的空对话会话。',
        params: [{ name: 'title', in: 'body', type: 'string', desc: '会话标题（可选）' }],
      },
      { id: 'update-conv', method: 'PUT', path: '/api/conversations/:id', summary: '修改会话', auth: true,
        description: '修改会话标题。',
        params: [
          { name: 'id', in: 'path', type: 'string', desc: '会话 ID' },
          { name: 'title', in: 'body', type: 'string', desc: '新标题' },
        ],
      },
      { id: 'delete-conv', method: 'DELETE', path: '/api/conversations/:id', summary: '删除会话', auth: true,
        description: '删除会话及其所有消息，并清理内存记忆。',
        params: [{ name: 'id', in: 'path', type: 'string', desc: '会话 ID' }],
      },
    ]
  },
  {
    id: 'misc', category: 'misc', label: '其他', color: 'bg-gray-500',
    endpoints: [
      { id: 'search', method: 'GET', path: '/api/search', summary: '全局搜索', auth: true,
        description: '跨材料、卡片、练习题的全文模糊搜索，每类最多返回 8 条。',
        params: [{ name: 'q', in: 'query', type: 'string', desc: '搜索关键词' }],
        response: '{ "results": [{ "type": "card", "title": "...", "subtitle": "...", "material_id": "..." }] }'
      },
      { id: 'graph', method: 'GET', path: '/api/graph/:material_id', summary: '知识图谱', auth: true,
        description: '获取材料的知识图谱数据（ECharts Graph 格式）。',
        params: [{ name: 'material_id', in: 'path', type: 'string', desc: '材料 ID' }],
        response: '{ "nodes": [{ "name": "...", "category": 0, "symbolSize": 50 }], "edges": [...] }'
      },
      { id: 'stats', method: 'GET', path: '/api/stats', summary: '学习统计', auth: true,
        description: '获取用户的学习统计数据（材料数、卡片数、答题数等）。',
        response: '{ "materials": 5, "cards": 30, "quizzes": 20, "conversations": 3 }'
      },
      { id: 'achievements', method: 'GET', path: '/api/achievements', summary: '学习成就', auth: true,
        description: '获取 18 种学习成就的进度和解锁状态。',
        response: '{ "achievements": [{ "id": "...", "name": "...", "progress": 5, "unlocked": true }] }'
      },
      { id: 'metrics', method: 'GET', path: '/api/dashboard/metrics', summary: 'LLM 指标', auth: true,
        description: '获取 LLM 调用的汇总指标（总调用数、Token 用量、平均延迟等）。',
      },
      { id: 'activity', method: 'GET', path: '/api/dashboard/activity', summary: '每日活跃', auth: true,
        description: '获取最近 30 天的每日学习活跃数据。',
      },
      { id: 'traces', method: 'GET', path: '/api/dashboard/traces', summary: '调用追踪', auth: true,
        description: '分页获取 LLM 调用追踪记录。',
        params: [
          { name: 'limit', in: 'query', type: 'int', desc: '每页条数（默认 50，上限 500）' },
          { name: 'offset', in: 'query', type: 'int', desc: '偏移量' },
        ],
      },
      { id: 'seed', method: 'POST', path: '/api/seed', summary: '生成演示数据', auth: true,
        description: '一键生成示例学习材料（Go 并发编程 + 机器学习基础），含完整分析结果、卡片和练习题。幂等操作。',
        response: '{ "message": "示例数据生成成功！", "seeded": true, "materials": 2, "cards": 9, "quizzes": 8 }'
      },
      { id: 'ws', method: 'GET', path: '/ws', summary: 'WebSocket', auth: false,
        description: 'WebSocket 端点，用于实时接收材料分析进度推送。连接时需携带 ?token=<JWT> 查询参数进行认证。',
        params: [{ name: 'token', in: 'query', type: 'string', desc: 'JWT Token' }],
      },
    ]
  },
]

const categories = computed(() => {
  return endpointGroups.map(g => ({
    id: g.id,
    label: g.label,
    count: g.endpoints.length,
  }))
})

const filteredGroups = computed(() => {
  const q = search.value.toLowerCase().trim()
  let groups = endpointGroups
  if (activeCategory.value) {
    groups = groups.filter(g => g.id === activeCategory.value)
  }
  if (!q) return groups
  return groups
    .map(g => ({
      ...g,
      endpoints: g.endpoints.filter(ep =>
        ep.path.toLowerCase().includes(q) ||
        ep.summary.toLowerCase().includes(q) ||
        (ep.description && ep.description.toLowerCase().includes(q)) ||
        ep.method.toLowerCase().includes(q)
      ),
    }))
    .filter(g => g.endpoints.length > 0)
})
</script>
