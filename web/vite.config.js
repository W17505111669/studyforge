import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'jsdom',
    include: ['src/**/*.{test,spec}.{js,ts}'],
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://127.0.0.1:8080',
        ws: true,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            // ECharts 体积大，单独拆包
            if (id.includes('echarts') || id.includes('zrender')) {
              return 'echarts'
            }
            // Markdown 渲染库单独拆包
            if (id.includes('marked')) {
              return 'marked'
            }
            // Vue 生态核心库合并为一个 vendor 包
            if (id.includes('vue') || id.includes('pinia') || id.includes('axios')) {
              return 'vendor'
            }
          }
        },
      },
    },
  },
})
