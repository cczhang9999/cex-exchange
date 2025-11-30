import axios from 'axios'
import { useUserStore } from '../stores'

const request = axios.create({
  timeout: 30000 // 30秒超时
})

// Request interceptor
request.interceptors.request.use(
  config => {
    // Read token directly from localStorage to avoid Pinia initialization issues
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// Response interceptor
request.interceptors.response.use(
  response => {
    return response
  },
  error => {
    if (error.response && error.response.status === 401) {
      const userStore = useUserStore()
      userStore.logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default request
