# StudyForge Pro

![CI](https://github.com/OWNER/studyforge-pro/actions/workflows/ci.yml/badge.svg)
![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go)
![Vue](https://img.shields.io/badge/Vue-3.4-4FC08D?style=flat-square&logo=vuedotjs)
![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)

StudyForge Pro 是一个 AI 驱动的智能学习平台，通过大语言模型自动将学习材料转化为知识卡片、练习题和知识图谱。平台集成了间隔重复算法（SM-2）、RAG 语义检索、多 Agent 协作分析、自适应出题等核心能力，帮助学习者高效掌握知识。

## 功能特性

**材料分析与知识提取**
- 上传 PDF、Word、Markdown、TXT 或 URL 材料，4 个 AI Agent 并发分析
- 两阶段流水线架构：Analyst 全文理解 → CardMaker / QuizMaster / MapBuilder 并发生成
- WebSocket 实时推送分析进度，Agent 甘特图时间线可视化
- 材料标签系统，支持自定义标签和按标签过滤

**知识卡片与间隔重复**
- SM-2 算法自动安排复习计划，支持"不熟 / 模糊 / 掌握"三级反馈
- 卡片滑动学习模式（移动端手势翻转 + 桌面端键盘快捷键）
- 书签与个人笔记系统，支持 Anki HTML 导出
- 知识卡片单击放大预览、双击翻转查看详情

**智能出题与练习**
- 4 种题型（选择 / 判断 / 填空 / 简答），学科感知出题策略
- 渐进式三级提示系统（方向提示 → 关键线索 → 接近答案）
- 雷达图练习报告 + 智能复习建议
- 自适应难度推荐，根据答题正确率动态调整
- 错题本自动收集 + 重练 + 巩固强化模式

**AI 对话与 RAG 检索**
- SSE 流式对话输出（打字机效果）
- Function Calling 工具调用（查卡片、查材料、创建练习等）
- RAG 语义检索增强，自动引用相关知识
- 多轮对话记忆（短期缓冲 + 长期摘要压缩）
- 对话历史持久化存储与管理

**知识图谱可视化**
- ECharts 交互式力导向图谱，圆角矩形卡片节点 + 分类着色
- 搜索过滤、缩放适应、节点高亮关联边、右键上下文菜单
- 跨材料知识图谱：合并所有材料，分组着色 + 跨材料关系高亮
- 边关系标签常驻显示（依赖 / 对比 / 关联等 8 种关系类型）

**学习路径规划**
- AI 根据所有材料生成学习路线图（竖向时间线）
- 步骤节点含预估时长、前置依赖、关联材料和复习进度

**多 Agent 辩论**
- 对同一知识点，3 个 Agent 从不同视角展开辩论
- Analyst（本质分析）→ QuizMaster（出题视角）→ CardMaker（记忆策略）→ 综合总结

**Dashboard 与成就系统**
- 数据看板：学习统计、Agent 质量评分（LLM-as-Judge 5 星级）、全年日历热力图
- 18 种成就徽章（学习 / 练习 / 复习 / 探索 / 特殊 5 大类）
- AI 学习建议：基于学习数据的智能推荐

**工程化与安全**
- PWA 支持：Service Worker 离线缓存，可安装为桌面应用
- 全面安全加固：JWT iss/aud 防重放、CORS 白名单、XSS DOMPurify 防护、IP 令牌桶限流、路径穿越防护、文件 DoS 防护、goroutine panic 恢复
- 环境变量 + YAML 双层配置，启动时自动校验必需项
- GitHub Actions CI/CD 流水线（Go Build & Test + Vue Build & Test + Go Lint）

## 技术架构

| 层级 | 技术栈 |
|------|--------|
| 后端框架 | Go 1.24 + Gin Web Framework |
| 前端框架 | Vue 3 + Vite 5 + TailwindCSS 3 |
| 数据库 | SQLite (GORM 自动迁移) |
| 向量存储 | 内存向量库（余弦相似度 + 支持 Qdrant 扩展） |
| AI/LLM | 通义千问 API（OpenAI 兼容协议，可替换） |
| 可视化 | ECharts 5（图表 / 图谱 / 热力图 / 雷达图） |
| 数学渲染 | KaTeX（4 种定界符，占位符防冲突策略） |
| 实时通信 | WebSocket（分析进度推送） + SSE（流式对话） |
| 状态管理 | Pinia + Vue Router |
| 测试 | Go test (149+ 用例) + Vitest + @vue/test-utils |
| 部署 | Docker 多阶段构建 + Docker Compose |

## 快速开始

### 环境要求

- **Go** 1.24+
- **Node.js** 20+ 和 npm
- **Qdrant** 向量数据库（可选，RAG 功能需要）

### 1. 克隆项目

```bash
git clone https://github.com/OWNER/studyforge-pro.git
cd studyforge-pro
```

### 2. 配置

复制配置文件模板并填入必需项：

```bash
cp config.yaml.example config.yaml
```

**必需配置**（二选一）：

| 配置项 | 环境变量 | config.yaml 字段 | 说明 |
|--------|----------|------------------|------|
| LLM API Key | `LLM_API_KEY` | `llm.api_key` | 通义千问 DashScope API Key |
| JWT Secret | `JWT_SECRET` | `jwt.secret` | JWT 签名密钥（建议随机强密钥） |

**可选环境变量**（完整列表见 `.env.example`）：

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `CONFIG_PATH` | `config.yaml` | 配置文件路径 |
| `SERVER_PORT` | `8080` | 服务监听端口 |
| `SERVER_HOST` | `0.0.0.0` | 服务监听地址 |
| `DB_PATH` | `./data/studyforge.db` | SQLite 数据库路径 |
| `LLM_BASE_URL` | DashScope 兼容模式 | LLM API 地址 |
| `LLM_MODEL` | `qwen-plus` | 对话模型 |
| `LLM_EMBEDDING_MODEL` | `text-embedding-v3` | 向量化模型 |
| `CORS_ORIGINS` | — | 额外允许的 CORS Origin（逗号分隔） |

> 配置优先级：环境变量 > config.yaml > 内置默认值

### 3. 启动后端

```bash
go run main.go
```

启动时会打印配置摘要（敏感值自动遮蔽），配置缺失时给出友好错误提示。

### 4. 启动前端（新终端）

```bash
cd web
npm install
npm run dev
```

访问 http://localhost:5173 即可使用。

### 5. 生成演示数据（可选）

首次使用可调用 Seed 接口自动生成 2 篇示例材料和配套数据：

```bash
# 先注册并获取 token
TOKEN=$(curl -s -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"demo","email":"demo@test.com","password":"demo123456"}' \
  | jq -r '.token')

# 生成演示数据
curl -X POST http://localhost:8080/api/seed \
  -H "Authorization: Bearer $TOKEN"
```

## Docker 部署

```bash
# 配置环境变量
export LLM_API_KEY=your-api-key
export JWT_SECRET=your-random-secret

# 一键启动（应用 + Qdrant 向量库）
docker-compose up -d
```

访问 http://localhost:8080 即可使用（后端直接提供前端静态文件）。

自定义配置可通过环境变量传入，详见 `.env.example`。数据持久化通过 Docker volumes 自动管理。

## API 概览

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/register` | 用户注册（返回 JWT token） |
| POST | `/api/login` | 用户登录 |
| GET | `/api/health` | 健康检查 |

### 材料管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/materials` | 材料列表（分页，支持 tag 过滤） |
| POST | `/api/materials` | 上传材料（URL / 文本） |
| POST | `/api/materials/upload` | 上传文件（PDF / Word / MD / TXT） |
| POST | `/api/materials/:id/analyze` | 触发 AI 分析（4 Agent 并发） |
| GET | `/api/materials/:id` | 材料详情（含卡片和题目） |
| DELETE | `/api/materials/:id` | 删除材料（级联清理） |
| POST | `/api/materials/batch-analyze` | 批量分析（并发限制 + WS 进度） |
| DELETE | `/api/materials/batch` | 批量删除 |
| GET | `/api/tags` | 所有标签及计数 |

### 知识卡片

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/cards` | 卡片列表（支持 material_id / difficulty / due / bookmarked 过滤） |
| GET | `/api/cards/export` | 导出 Anki HTML 格式 |
| POST | `/api/cards/:id/review` | 卡片复习（SM-2 算法） |
| PUT | `/api/cards/:id/bookmark` | 切换书签 |
| PUT | `/api/cards/:id/note` | 更新个人笔记 |

### 练习题

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/quizzes` | 题目列表（支持 recommended=true 自适应推荐） |
| GET | `/api/quizzes/difficulty-level` | 获取难度推荐分析 |
| GET | `/api/quizzes/:id/hint` | 渐进式提示（level=1/2/3） |
| POST | `/api/quizzes/:id/answer` | 提交答案 |

### 错题本

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/mistakes` | 错题列表（分页） |
| GET | `/api/mistakes/stats` | 错题统计 |
| POST | `/api/mistakes/retry` | 错题重练（最多 20 题） |
| POST | `/api/mistakes/consolidate` | 巩固强化（智能推荐同类题） |
| POST | `/api/mistakes/batch-review` | 批量标记已复习 |
| POST | `/api/mistakes/:id/review` | 标记单条已复习 |
| DELETE | `/api/mistakes/:id` | 删除错题记录 |

### AI 对话

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/chat` | 对话（非流式） |
| GET | `/api/chat/stream` | 对话（SSE 流式 + Function Calling） |
| GET | `/api/conversations` | 对话历史列表 |
| POST | `/api/conversations` | 创建对话 |
| DELETE | `/api/conversations/:id` | 删除对话 |

### 学习功能

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/learning-path` | 学习路径规划（AI 生成） |
| POST | `/api/debate` | 多 Agent 辩论 |
| GET | `/api/recommendations` | AI 学习建议 |
| GET | `/api/achievements` | 成就系统 |

### 知识图谱

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/graph/:material_id` | 单材料知识图谱 |
| GET | `/api/graph/all` | 跨材料知识图谱 |

### 其他

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/search` | 全局搜索（材料 / 卡片 / 题目 / 标签） |
| GET | `/api/notifications` | 通知列表 |
| GET | `/api/notifications/unread-count` | 未读通知计数 |
| POST | `/api/notifications/read-all` | 全部标记已读 |
| GET | `/api/stats` | 学习统计 |
| GET | `/api/stats/calendar` | 全年学习热力图 |
| GET | `/api/dashboard/metrics` | Dashboard 指标 |
| POST | `/api/seed` | 生成演示数据 |

> 除 `/api/register`、`/api/login`、`/api/health` 外，所有端点均需在请求头携带 `Authorization: Bearer <token>`。
> 应用内 `/api-docs` 页面提供完整的交互式 API 文档。

## 项目结构

```
studyforge-pro/
├── main.go                         # 应用入口 + 路由注册 + CORS
├── config.go                       # 配置加载、校验、环境变量覆盖
├── config.yaml.example             # 配置文件模板
├── .env.example                    # 环境变量参考文档
├── Dockerfile                      # 多阶段 Docker 构建
├── docker-compose.yml              # 编排应用 + Qdrant
├── .gitignore
├── .github/
│   └── workflows/
│       └── ci.yml                  # GitHub Actions CI/CD
│
├── internal/
│   ├── agent/                      # AI Agent 系统
│   │   ├── orchestrator.go         #   两阶段流水线编排
│   │   ├── analyst.go              #   材料分析 Agent
│   │   ├── cardmaker.go            #   知识卡片生成 Agent
│   │   ├── quizmaster.go           #   出题 Agent（学科感知 + 三级提示）
│   │   ├── mapbuilder.go           #   图谱构建 Agent
│   │   ├── pathplanner.go          #   学习路径规划 Agent
│   │   ├── debate.go               #   多 Agent 辩论协调器
│   │   ├── llm.go                  #   LLM 客户端（同步 + 流式）
│   │   └── tools.go                #   Function Calling 工具定义
│   ├── database/
│   │   └── database.go             #   SQLite 初始化 + GORM 自动迁移
│   ├── eval/
│   │   └── judge.go                #   LLM-as-Judge 质量评估
│   ├── handler/                    # HTTP 请求处理器
│   │   ├── auth.go                 #   注册 / 登录 / 记忆管理
│   │   ├── material.go             #   材料 CRUD + 批量操作
│   │   ├── study.go                #   卡片 / 题目 / 对话 / 搜索 / 导出
│   │   ├── conversation.go         #   对话历史 CRUD
│   │   ├── dashboard.go            #   统计 / 日历热力图 / 质量指标
│   │   ├── achievement.go          #   成就系统
│   │   ├── recommendation.go       #   AI 学习建议
│   │   ├── learning_path.go        #   学习路径规划
│   │   ├── debate.go               #   多 Agent 辩论
│   │   ├── mistake.go              #   错题本 + 巩固强化
│   │   ├── notification.go         #   通知系统（惰性生成）
│   │   ├── seed.go                 #   演示数据生成
│   │   ├── ws.go                   #   WebSocket 推送
│   │   ├── fileparser.go           #   文件解析（PDF / DOCX / MD / TXT）
│   │   ├── handler_test.go         #   Handler 层测试（16 用例）
│   │   └── testutil.go             #   测试辅助函数
│   ├── memory/
│   │   ├── memory.go               #   对话记忆（短期缓冲 + 长期摘要）
│   │   └── memory_test.go          #   Memory 测试（34 用例）
│   ├── middleware/
│   │   ├── jwt.go                  #   JWT 认证中间件
│   │   ├── ratelimit.go            #   IP 级令牌桶限流
│   │   └── middleware_test.go      #   Middleware 测试（26 用例）
│   ├── model/                      # 数据模型
│   │   ├── material.go             #   Material
│   │   ├── card.go                 #   Card + SM-2 算法
│   │   ├── quiz.go                 #   Quiz + QuizAttempt（含三级提示）
│   │   ├── user.go                 #   User
│   │   ├── conversation.go         #   Conversation + ChatMessage
│   │   ├── achievement.go          #   UserAchievement + 18 种成就定义
│   │   ├── mistake.go              #   QuizMistake
│   │   ├── notification.go         #   Notification
│   │   ├── trace.go                #   LLMTrace（含 Judge 评分）
│   │   └── model_test.go           #   Model 测试（35 用例）
│   └── rag/                        # RAG 检索增强
│       ├── retriever.go            #   内存向量存储 + 语义搜索
│       ├── retriever_test.go       #   RAG 测试（38 用例）
│       ├── chunker.go              #   文档切片
│       └── embedder.go             #   Embedding 封装
│
└── web/                            # Vue 3 前端
    ├── index.html                  #   入口 HTML + PWA 注册
    ├── vite.config.js              #   Vite 配置 + 手动分包
    ├── tailwind.config.js          #   TailwindCSS + 暗色模式
    ├── package.json
    ├── generate-icons.js           #   PWA 图标生成脚本
    ├── public/                     #   PWA 静态资源
    │   ├── manifest.json
    │   ├── sw.js                   #   Service Worker
    │   ├── offline.html            #   离线回退页
    │   └── icon-*.png              #   多尺寸图标
    └── src/
        ├── main.js                 #   Vue 入口 + 全局错误处理
        ├── App.vue                 #   根组件 + 路由过渡动画
        ├── style.css               #   全局样式
        ├── api/
        │   └── client.js           #   Axios API 客户端（401 自动登出）
        ├── components/
        │   ├── AppLayout.vue       #   应用布局 + 侧栏 + 通知铃铛
        │   ├── AgentFlow.vue       #   Agent 流程可视化 + 星级评分
        │   ├── AgentTimeline.vue   #   Agent 甘特图时间线
        │   ├── GlobalSearch.vue    #   Ctrl+K 全局搜索模态框
        │   ├── Toast.vue           #   Toast 通知 + Confirm 弹窗
        │   ├── ErrorBoundary.vue   #   错误边界
        │   ├── OnboardingOverlay.vue # 首次登录引导
        │   └── __tests__/          #   组件测试
        ├── composables/
        │   ├── useDarkMode.js      #   暗色模式切换
        │   ├── useToast.js         #   Toast / Confirm composable
        │   ├── useOnboarding.js    #   引导步骤状态
        │   └── useMathRender.js    #   KaTeX 数学公式渲染
        ├── router/
        │   └── index.js            #   Vue Router（含路由守卫）
        ├── stores/
        │   └── auth.js             #   Pinia 认证状态（JWT 过期检查）
        └── views/
            ├── Dashboard.vue       #   数据看板 + 成就 + 热力图 + 建议
            ├── Upload.vue          #   材料上传 + 分析 + 批量操作
            ├── Cards.vue           #   知识卡片 + 书签 + 笔记
            ├── CardStudy.vue       #   卡片滑动学习模式
            ├── Quiz.vue            #   练习 + 报告 + 自适应推荐
            ├── Mistakes.vue        #   错题本 + 重练 + 巩固强化
            ├── Chat.vue            #   AI 对话 + 代码块复制
            ├── Graph.vue           #   知识图谱（单材料 / 跨材料）
            ├── MaterialDetail.vue  #   材料详情页
            ├── LearningPath.vue    #   学习路径时间线
            ├── Debate.vue          #   多 Agent 辩论
            ├── ApiDocs.vue         #   API 文档（OpenAPI 风格）
            ├── Login.vue           #   登录 / 注册
            └── NotFound.vue        #   404 页面
```

## 截图

> 以下截图占位符将在项目部署后替换为实际截图。

| Dashboard 数据看板 | 材料上传与分析 |
|:---:|:---:|
| ![Dashboard](docs/screenshots/dashboard.png) | ![Upload](docs/screenshots/upload.png) |

| 知识卡片学习 | AI 对话 |
|:---:|:---:|
| ![Cards](docs/screenshots/cards.png) | ![Chat](docs/screenshots/chat.png) |

| 知识图谱可视化 | 练习题与报告 |
|:---:|:---:|
| ![Graph](docs/screenshots/graph.png) | ![Quiz](docs/screenshots/quiz.png) |

## 测试

```bash
# Go 后端测试（149+ 用例）
go test ./... -v -count=1

# 前端组件测试（25 用例）
cd web && npm run test

# 前端构建验证
cd web && npm run build
```

测试覆盖 5 个后端包（handler / model / rag / memory / middleware）和 2 个前端组件（Toast / ErrorBoundary），包含并发安全和边界条件测试。

## CI/CD

项目使用 GitHub Actions 自动执行 CI/CD 流水线（`.github/workflows/ci.yml`）：

- **Go Build & Test** — 编译检查、全量单元测试、go vet 静态分析
- **Vue Build & Test** — 前端构建、Vitest 组件测试
- **Go Lint** — golangci-lint 代码质量检查

每次推送到 `main` 分支或提交 Pull Request 时自动触发。

## 配置参考

完整的环境变量列表和说明请查看 [`.env.example`](.env.example)。

配置文件模板请查看 [`config.yaml.example`](config.yaml.example)。

配置优先级：**环境变量** > **config.yaml** > **内置默认值**

启动时自动校验必需配置（LLM API Key、JWT Secret 等），缺失时输出友好错误提示和修复建议。

## 贡献指南

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 确保测试通过 (`go test ./...` 和 `cd web && npm run test`)
4. 提交更改 (`git commit -m 'Add amazing feature'`)
5. 推送到分支 (`git push origin feature/amazing-feature`)
6. 创建 Pull Request

### 开发规范

- 所有数据库查询必须按 `user_id` 隔离
- 前端 Toast 通知使用 `composables/useToast.js`，禁止 `alert()` / `confirm()`
- 暗色模式 `dark:` 变体必须完整覆盖
- 响应式布局适配 `sm` / `md` / `lg` 断点
- 新增路由需在 `main.go` 注册，前端 API 在 `web/src/api/client.js` 添加

## 许可证

[MIT](LICENSE)
