import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    // dev 启动时预转换常用页面模块，消除首次切换路由的按需编译等待
    warmup: {
      clientFiles: ['./src/pages/**/*.vue', './src/layouts/*.vue'],
    },
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8090',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://127.0.0.1:8090',
        ws: true,
      },
      '/uploads': {
        target: 'http://127.0.0.1:8090',
        changeOrigin: true,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        // 稳定 vendor 分包：框架/网络/文本处理库独立成块，
        // 业务页面更新时不影响 vendor 缓存，多页面共享同一份库代码
        // （Vite 8 / Rolldown 仅支持函数形式）
        manualChunks(id: string): string | undefined {
          const p = id.replace(/\\/g, '/')
          if (!p.includes('/node_modules/')) return undefined
          if (/[\\/]node_modules[\\/](vue|@vue|vue-router|pinia)[\\/]/.test(p)) return 'vendor-vue'
          if (/[\\/]node_modules[\\/](axios|socket\.io|@socket\.io|engine\.io)[\\/]/.test(p)) return 'vendor-net'
          if (/[\\/]node_modules[\\/](marked|turndown|pinyin-pro|@pinyin-pro)[\\/]/.test(p)) return 'vendor-text'
          return undefined
        },
      },
    },
  },
})
