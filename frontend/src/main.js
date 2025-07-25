import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import axios from 'axios'

// 自动携带JWT token
import { useUserStore } from './stores'
axios.interceptors.request.use(config => {
  // 兼容pinia和localStorage
  let token = ''
  try {
    const userStore = useUserStore()
    token = userStore.token
  } catch {}
  if (!token) {
    token = localStorage.getItem('token') || ''
  }
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

axios.interceptors.response.use(
  res => res,
  err => {
    if (err.response && err.response.status === 401) {
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

const app = createApp(App)

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(router)
app.use(createPinia())
app.use(ElementPlus)
app.mount('#app')
