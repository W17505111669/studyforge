import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { guest: true },
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
      { path: 'quiz', name: 'Quiz', component: () => import('../views/Quiz.vue') },
      { path: 'mistakes', name: 'Mistakes', component: () => import('../views/Mistakes.vue') },
      { path: 'graph/:materialId?', name: 'Graph', component: () => import('../views/Graph.vue') },
      { path: 'learning-path', name: 'LearningPath', component: () => import('../views/LearningPath.vue') },
      { path: 'debate', name: 'Debate', component: () => import('../views/Debate.vue') },
      { path: 'chat', name: 'Chat', component: () => import('../views/Chat.vue') },
      { path: 'api-docs', name: 'ApiDocs', component: () => import('../views/ApiDocs.vue') },
      // Catch-all 404 — 必须放在最后
      { path: ':pathMatch(.*)*', name: 'NotFound', component: () => import('../views/NotFound.vue') },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
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
