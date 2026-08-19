import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import { vPermission } from './directives/permission'
import { setupErrorHandler } from './utils/errorHandler'
// 本地字体（Inter，替代 Google Fonts 在线加载，避免内网访问卡顿）
import '@fontsource/inter/400.css'
import '@fontsource/inter/500.css'
import '@fontsource/inter/600.css'
import '@fontsource/inter/700.css'
import './style.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.directive('permission', vPermission)

setupErrorHandler(app)

app.mount('#app')
