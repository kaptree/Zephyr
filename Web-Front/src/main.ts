import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import { vPermission } from './directives/permission'
import { setupErrorHandler } from './utils/errorHandler'
import './style.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.directive('permission', vPermission)

setupErrorHandler(app)

app.mount('#app')
