import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import { vPermission } from './directives/permission'
import { vTooltip } from './directives/tooltip'
import { setupErrorHandler } from './utils/errorHandler'
import { setupRipple } from './utils/ripple'
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
app.directive('tooltip', vTooltip)

setupErrorHandler(app)

// 全局按钮点击涟漪（微交互语言：事件委托，无需逐个按钮接入）
setupRipple()

app.mount('#app')
