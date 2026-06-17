<template>
  <div class="p-4 sm:p-6 lg:p-8 max-w-5xl mx-auto">
    <!-- Header -->
    <div class="mb-6">
      <div class="flex items-center gap-3 mb-2">
        <button
          class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
          @click="$router.back()"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">API 接口文档</h1>
        <span
          class="px-2 py-0.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 text-xs font-medium rounded-full"
        >
          OpenAPI 3.0
        </span>
      </div>
      <p class="text-sm text-gray-500 dark:text-gray-400 ml-8">
        从 Swagger 自动生成。所有认证端点需在 Header 携带
        <code class="text-xs bg-gray-100 dark:bg-gray-700 px-1 rounded">Authorization: Bearer &lt;token&gt;</code>
      </p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-4">
      <div v-for="i in 6" :key="i" class="animate-pulse">
        <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/4 mb-2"></div>
        <div class="h-12 bg-gray-100 dark:bg-gray-800 rounded-lg"></div>
      </div>
    </div>

    <!-- Error state -->
    <div
      v-else-if="loadError"
      class="text-center py-12 bg-red-50 dark:bg-red-900/10 rounded-xl border border-red-200 dark:border-red-800"
    >
      <svg class="w-12 h-12 mx-auto mb-3 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"
        />
      </svg>
      <p class="text-red-600 dark:text-red-400 font-medium">加载 OpenAPI 规范失败</p>
      <p class="text-sm text-red-500 dark:text-red-500 mt-1">{{ loadError }}</p>
      <button
        class="mt-4 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 text-sm"
        @click="loadSwaggerSpec"
      >
        重试
      </button>
    </div>

    <!-- Main content -->
    <template v-else>
      <!-- Search -->
      <div class="mb-6">
        <div class="relative">
          <svg
            class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
          <input
            v-model="search"
            type="text"
            placeholder="搜索端点路径或描述..."
            class="w-full pl-10 pr-4 py-2.5 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none"
          />
        </div>
      </div>

      <!-- Stats bar -->
      <div class="flex flex-wrap items-center gap-3 mb-4 text-xs text-gray-500 dark:text-gray-400">
        <span class="font-medium">{{ totalEndpoints }}</span> 个端点
        <span class="text-gray-300 dark:text-gray-600">|</span>
        <span class="px-2 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 rounded-full font-medium">
          GET {{ methodCount.GET }}
        </span>
        <span class="px-2 py-0.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 rounded-full font-medium">
          POST {{ methodCount.POST }}
        </span>
        <span class="px-2 py-0.5 bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 rounded-full font-medium">
          PUT {{ methodCount.PUT }}
        </span>
        <span class="px-2 py-0.5 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 rounded-full font-medium">
          DELETE {{ methodCount.DELETE }}
        </span>
        <a
          href="/api/swagger/index.html"
          target="_blank"
          class="ml-auto flex items-center gap-1 text-primary-600 dark:text-primary-400 hover:underline"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
            />
          </svg>
          Swagger UI
        </a>
      </div>

      <!-- Category filter -->
      <div class="flex flex-wrap gap-2 mb-6">
        <button
          v-for="cat in tagList"
          :key="cat.name"
          class="px-3 py-1.5 rounded-full text-xs font-medium transition-colors"
          :class="
            activeTag === cat.name
              ? 'bg-primary-600 text-white'
              : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700'
          "
          @click="activeTag = activeTag === cat.name ? '' : cat.name"
        >
          {{ cat.name }}
          <span class="ml-1 opacity-60">{{ cat.count }}</span>
        </button>
      </div>

      <!-- Endpoints -->
      <div class="space-y-3">
        <template v-for="group in filteredGroups" :key="group.tag">
          <h2 class="text-lg font-semibold text-gray-800 dark:text-gray-200 pt-2 flex items-center gap-2">
            <span class="w-2 h-2 rounded-full" :class="tagColor(group.tag)"></span>
            {{ group.tag }}
          </h2>
          <div
            v-for="ep in group.endpoints"
            :key="ep.id"
            class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden bg-white dark:bg-gray-800"
          >
            <!-- Endpoint header -->
            <button
              class="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-gray-50 dark:hover:bg-gray-750 transition-colors"
              @click="toggleEndpoint(ep.id)"
            >
              <span
                class="px-2 py-0.5 rounded text-xs font-bold uppercase min-w-[56px] text-center shrink-0"
                :class="methodClass(ep.method)"
              >
                {{ ep.method }}
              </span>
              <code class="text-sm font-mono text-gray-800 dark:text-gray-200 flex-1 truncate">{{ ep.path }}</code>
              <span class="text-sm text-gray-500 dark:text-gray-400 hidden sm:inline truncate max-w-[200px]">{{
                ep.summary
              }}</span>
              <div
                v-if="ep.secured"
                class="shrink-0 w-4 h-4 text-amber-500 dark:text-amber-400"
                title="需要认证"
              >
                <svg fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fill-rule="evenodd"
                    d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z"
                    clip-rule="evenodd"
                  />
                </svg>
              </div>
              <svg
                class="w-4 h-4 text-gray-400 transition-transform shrink-0"
                :class="{ 'rotate-180': openEndpoints.has(ep.id) }"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M19 9l-7 7-7-7"
                />
              </svg>
            </button>

            <!-- Expanded details -->
            <div
              v-if="openEndpoints.has(ep.id)"
              class="border-t border-gray-100 dark:border-gray-700 px-4 py-3 space-y-3 bg-gray-50/50 dark:bg-gray-800/50"
            >
              <p class="text-sm text-gray-600 dark:text-gray-400">{{ ep.description }}</p>

              <!-- Params -->
              <div v-if="ep.params && ep.params.length">
                <h4
                  class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1"
                >
                  参数
                </h4>
                <div class="overflow-x-auto">
                  <table class="w-full text-sm">
                    <thead>
                      <tr
                        class="text-left text-xs text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700"
                      >
                        <th class="pb-1 pr-4">名称</th>
                        <th class="pb-1 pr-4">位置</th>
                        <th class="pb-1 pr-4">类型</th>
                        <th class="pb-1 pr-4">必填</th>
                        <th class="pb-1">说明</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="p in ep.params"
                        :key="p.name"
                        class="border-b border-gray-100 dark:border-gray-700/50"
                      >
                        <td class="py-1.5 pr-4">
                          <code
                            class="text-xs bg-gray-100 dark:bg-gray-700 px-1 rounded text-primary-600 dark:text-primary-400"
                          >{{ p.name }}</code>
                        </td>
                        <td class="py-1.5 pr-4 text-xs text-gray-500 dark:text-gray-400">{{ p.in }}</td>
                        <td class="py-1.5 pr-4 text-xs text-gray-500 dark:text-gray-400">{{ p.type }}</td>
                        <td class="py-1.5 pr-4">
                          <span
                            v-if="p.required"
                            class="text-xs text-red-500 dark:text-red-400 font-medium"
                          >是</span>
                          <span v-else class="text-xs text-gray-400">否</span>
                        </td>
                        <td class="py-1.5 text-gray-600 dark:text-gray-400">{{ p.description }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <!-- Try it out -->
              <div class="flex items-center gap-2">
                <button
                  class="px-3 py-1.5 bg-primary-600 hover:bg-primary-700 text-white text-xs font-medium rounded-lg transition-colors flex items-center gap-1.5"
                  @click.stop="tryItOut(ep)"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                    />
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                  Try it out
                </button>
                <span v-if="ep.secured" class="text-xs text-gray-400 dark:text-gray-500">
                  需要 Bearer Token
                </span>
              </div>

              <!-- Try it out result -->
              <div
                v-if="tryResults[ep.id]"
                class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden"
              >
                <div
                  class="flex items-center gap-2 px-3 py-2 bg-gray-100 dark:bg-gray-900 text-xs"
                >
                  <span
                    class="px-2 py-0.5 rounded font-bold"
                    :class="statusClass(tryResults[ep.id].status)"
                  >
                    {{ tryResults[ep.id].status }}
                  </span>
                  <span class="text-gray-500 dark:text-gray-400">{{ tryResults[ep.id].time }}ms</span>
                  <button
                    class="ml-auto text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                    @click.stop="delete tryResults[ep.id]"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M6 18L18 6M6 6l12 12"
                      />
                    </svg>
                  </button>
                </div>
                <pre
                  class="text-xs bg-gray-900 dark:bg-gray-950 text-green-400 p-3 overflow-x-auto max-h-64 leading-relaxed"
                >{{ tryResults[ep.id].body }}</pre>
              </div>
            </div>
          </div>
        </template>

        <!-- Empty state -->
        <div v-if="filteredGroups.length === 0" class="text-center py-12 text-gray-400 dark:text-gray-500">
          <svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.5"
              d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <p>没有找到匹配的端点</p>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted } from 'vue'

const loading = ref(true)
const loadError = ref('')
const search = ref('')
const activeTag = ref('')
const openEndpoints = reactive(new Set())
const tryResults = reactive({})
const endpointGroups = ref([])
const tagColors = [
  'bg-blue-500',
  'bg-purple-500',
  'bg-green-500',
  'bg-orange-500',
  'bg-indigo-500',
  'bg-teal-500',
  'bg-red-500',
  'bg-pink-500',
  'bg-cyan-500',
  'bg-amber-500'
]

// Parse OpenAPI spec from swagger JSON
async function loadSwaggerSpec() {
  loading.value = true
  loadError.value = ''
  try {
    const token = localStorage.getItem('token')
    const res = await fetch('/api/swagger/doc.json', {
      headers: token ? { Authorization: `Bearer ${token}` } : {}
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const spec = await res.json()
    parseOpenAPISpec(spec)
  } catch (e) {
    loadError.value = e.message
  } finally {
    loading.value = false
  }
}

function parseOpenAPISpec(spec) {
  const paths = spec.paths || {}
  const groups = {}

  for (const [path, methods] of Object.entries(paths)) {
    for (const [method, details] of Object.entries(methods)) {
      if (['get', 'post', 'put', 'delete', 'patch'].indexOf(method) === -1) continue

      const tag = (details.tags && details.tags[0]) || '其他'
      if (!groups[tag]) {
        groups[tag] = []
      }

      // Build params
      const params = []
      if (details.parameters) {
        for (const p of details.parameters) {
          params.push({
            name: p.name,
            in: p.in,
            type: p.schema ? p.schema.type || 'string' : 'string',
            required: !!p.required,
            description: p.description || ''
          })
        }
      }
      // Add body params from requestBody
      if (details.requestBody) {
        const content = details.requestBody.content
        if (content && content['application/json'] && content['application/json'].schema) {
          const schema = content['application/json'].schema
          if (schema.$ref) {
            // Resolve ref
            const refName = schema.$ref.split('/').pop()
            const def = spec.definitions && spec.definitions[refName]
            if (def && def.properties) {
              for (const [propName, propDef] of Object.entries(def.properties)) {
                params.push({
                  name: propName,
                  in: 'body',
                  type: propDef.type || 'string',
                  required: def.required ? def.required.includes(propName) : false,
                  description: propDef.description || ''
                })
              }
            } else {
              params.push({
                name: 'request',
                in: 'body',
                type: refName,
                required: true,
                description: ''
              })
            }
          } else if (schema.properties) {
            for (const [propName, propDef] of Object.entries(schema.properties)) {
              params.push({
                name: propName,
                in: 'body',
                type: propDef.type || 'string',
                required: schema.required ? schema.required.includes(propName) : false,
                description: propDef.description || ''
              })
            }
          }
        }
        if (content && content['multipart/form-data']) {
          const fdSchema = content['multipart/form-data'].schema
          if (fdSchema && fdSchema.properties) {
            for (const [propName, propDef] of Object.entries(fdSchema.properties)) {
              params.push({
                name: propName,
                in: 'form-data',
                type: propDef.type || 'file',
                required: fdSchema.required ? fdSchema.required.includes(propName) : false,
                description: propDef.description || ''
              })
            }
          }
        }
      }

      const secured = !!(details.security && details.security.length > 0)

      groups[tag].push({
        id: `${method}-${path}`,
        method: method.toUpperCase(),
        path: `/api${path}`,
        summary: details.summary || '',
        description: details.description || '',
        params,
        secured,
        tag
      })
    }
  }

  // Convert to sorted array
  const tagOrder = ['认证', '材料管理', '知识卡片', '练习题', 'AI 对话', '搜索', '学习统计']
  endpointGroups.value = Object.entries(groups)
    .map(([tag, endpoints]) => ({ tag, endpoints }))
    .sort((a, b) => {
      const ai = tagOrder.indexOf(a.tag)
      const bi = tagOrder.indexOf(b.tag)
      if (ai !== -1 && bi !== -1) return ai - bi
      if (ai !== -1) return -1
      if (bi !== -1) return 1
      return a.tag.localeCompare(b.tag)
    })
}

function toggleEndpoint(id) {
  openEndpoints.has(id) ? openEndpoints.delete(id) : openEndpoints.add(id)
}

function methodClass(m) {
  const map = {
    GET: 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400',
    POST: 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400',
    PUT: 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400',
    DELETE: 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400',
    PATCH: 'bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-400'
  }
  return map[m] || 'bg-gray-100 text-gray-600'
}

function tagColor(tag) {
  const idx = Math.abs(tag.split('').reduce((a, c) => a + c.charCodeAt(0), 0)) % tagColors.length
  return tagColors[idx]
}

function statusClass(status) {
  if (status >= 200 && status < 300) return 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
  if (status >= 400 && status < 500) return 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400'
  if (status >= 500) return 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400'
  return 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
}

async function tryItOut(ep) {
  const token = localStorage.getItem('token')
  const startTime = Date.now()
  try {
    const headers = { 'Content-Type': 'application/json' }
    if (token && ep.secured) {
      headers['Authorization'] = `Bearer ${token}`
    }
    const fetchOptions = {
      method: ep.method,
      headers
    }
    // For GET requests with path params, replace them
    let url = ep.path
    const pathParams = ep.params.filter((p) => p.in === 'path')
    for (const pp of pathParams) {
      url = url.replace(`{${pp.name}}`, 'test-id')
    }

    const res = await fetch(url, fetchOptions)
    const elapsed = Date.now() - startTime
    let body = ''
    const ct = res.headers.get('content-type') || ''
    if (ct.includes('json')) {
      const json = await res.json()
      body = JSON.stringify(json, null, 2)
    } else {
      body = await res.text()
      if (body.length > 2000) body = body.substring(0, 2000) + '\n... (truncated)'
    }
    tryResults[ep.id] = { status: res.status, time: elapsed, body }
  } catch (e) {
    tryResults[ep.id] = { status: 0, time: Date.now() - startTime, body: `Error: ${e.message}` }
  }
}

// Computed
const tagList = computed(() => {
  return endpointGroups.value.map((g) => ({
    name: g.tag,
    count: g.endpoints.length
  }))
})

const totalEndpoints = computed(() => {
  return endpointGroups.value.reduce((sum, g) => sum + g.endpoints.length, 0)
})

const methodCount = computed(() => {
  const counts = { GET: 0, POST: 0, PUT: 0, DELETE: 0, PATCH: 0 }
  for (const g of endpointGroups.value) {
    for (const ep of g.endpoints) {
      if (counts[ep.method] !== undefined) counts[ep.method]++
    }
  }
  return counts
})

const filteredGroups = computed(() => {
  const q = search.value.toLowerCase().trim()
  let groups = endpointGroups.value
  if (activeTag.value) {
    groups = groups.filter((g) => g.tag === activeTag.value)
  }
  if (!q) return groups
  return groups
    .map((g) => ({
      ...g,
      endpoints: g.endpoints.filter(
        (ep) =>
          ep.path.toLowerCase().includes(q) ||
          ep.summary.toLowerCase().includes(q) ||
          (ep.description && ep.description.toLowerCase().includes(q)) ||
          ep.method.toLowerCase().includes(q)
      )
    }))
    .filter((g) => g.endpoints.length > 0)
})

onMounted(() => {
  loadSwaggerSpec()
})
</script>
