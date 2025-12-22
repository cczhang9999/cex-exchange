import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    username: localStorage.getItem('username') || '',
    userInfo: null,
  }),
  actions: {
    setToken(token) {
      this.token = token
      localStorage.setItem('token', token)
      
      // 自动从 token 中解析用户名（如果是 JWT）
      if (token) {
        try {
          const parts = token.split('.')
          if (parts.length === 3) {
            const payload = JSON.parse(atob(parts[1]))
            const username = payload.username || payload.sub || ''
            if (username) {
              this.setUsername(username)
            }
          }
        } catch (error) {
          console.error('解析 token 失败:', error)
        }
      }
    },
    setUsername(username) {
      this.username = username
      localStorage.setItem('username', username)
    },
    setUserInfo(info) {
      this.userInfo = info
    },
    logout() {
      this.token = ''
      this.username = ''
      this.userInfo = null
      localStorage.removeItem('token')
      localStorage.removeItem('username')
    }
  }
}) 