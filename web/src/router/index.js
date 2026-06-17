import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { guest: true }
  },
  {
    path: '/',
    component: () => import('../components/AppLayout.vue'),
    meta: { auth: true },
    children: [
      { path: '', name: 'Dashboard', component: () => import('../views/Dashboard.vue') },
      { path: 'upload', name: 'Upload', component: () => import('../views/Upload.vue') },
      { path: 'materials/:id', name: 'MaterialDetail', component: () => import('../views/MaterialDetail.vue') },
      { path: 'cards', name: 'Cards', component: () => import('../views/Cards.vue') },
      { path: 'study', name: 'CardStudy', component: () => import('../views/CardStudy.vue') },
      { path: 'review', name: 'Review', component: () => import('../views/Review.vue') },
      { path: 'quiz', name: 'Quiz', component: () => import('../views/Quiz.vue') },
      { path: 'mistakes', name: 'Mistakes', component: () => import('../views/Mistakes.vue') },
      { path: 'exam', name: 'Exam', component: () => import('../views/Exam.vue') },
      { path: 'diagnosis', name: 'Diagnosis', component: () => import('../views/Diagnosis.vue') },
      { path: 'insights', name: 'Insights', component: () => import('../views/Insights.vue') },
      { path: 'graph/:materialId?', name: 'Graph', component: () => import('../views/Graph.vue') },
      { path: 'learning-path', name: 'LearningPath', component: () => import('../views/LearningPath.vue') },
      { path: 'debate', name: 'Debate', component: () => import('../views/Debate.vue') },
      { path: 'pomodoro', name: 'Pomodoro', component: () => import('../views/Pomodoro.vue') },
      { path: 'goals', name: 'Goals', component: () => import('../views/Goals.vue') },
      { path: 'reports', name: 'Reports', component: () => import('../views/Reports.vue') },
      { path: 'leaderboard', name: 'Leaderboard', component: () => import('../views/Leaderboard.vue') },
      { path: 'friends', name: 'Friends', component: () => import('../views/Friends.vue') },
      { path: 'groups', name: 'Groups', component: () => import('../views/Groups.vue') },
      { path: 'import', name: 'Import', component: () => import('../views/Import.vue') },
      { path: 'notes', name: 'Notes', component: () => import('../views/Notes.vue') },
      { path: 'market', name: 'Market', component: () => import('../views/Market.vue') },
      { path: 'export', name: 'ExportData', component: () => import('../views/ExportData.vue') },
      { path: 'chat', name: 'Chat', component: () => import('../views/Chat.vue') },
      { path: 'admin/metrics', name: 'Metrics', component: () => import('../views/Metrics.vue') },
      { path: 'api-docs', name: 'ApiDocs', component: () => import('../views/ApiDocs.vue') },
      // Catch-all 404 — 必须放在最后
      { path: ':pathMatch(.*)*', name: 'NotFound', component: () => import('../views/NotFound.vue') }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()
  if (to.meta.auth && !auth.isLoggedIn) {
    next('/login')
  } else if (to.meta.guest && auth.isLoggedIn) {
    next('/')
  } else {
    next()
  }
})

export default router
