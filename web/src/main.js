import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './style.css'

const app = createApp(App)

// 全局错误处理：捕获组件内未处理的错误，防止白屏
app.config.errorHandler = (err, instance, info) => {
  console.error('[GlobalError] Vue 组件错误:', err)
  console.error('[GlobalError] 上下文:', info)
}

// 捕获未处理的 Promise rejection
window.addEventListener('unhandledrejection', (event) => {
  console.error('[GlobalError] 未处理的 Promise 拒绝:', event.reason)
})

app.use(createPinia())
app.use(router)
app.mount('#app')
