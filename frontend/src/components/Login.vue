<template>
  <div class="auth-container">
    <div class="auth-orb auth-orb-1"></div>
    <div class="auth-orb auth-orb-2"></div>

    <el-card class="auth-card glass-card" shadow="hover">
      <template #header>
        <div class="auth-header">
          <div class="brand">
            <span class="brand-dot"></span>
            <span class="brand-text">CEX Exchange</span>
          </div>
          <h2 class="auth-title">欢迎回来</h2>
          <p class="auth-subtitle">请输入您的账号信息以登录</p>
        </div>
      </template>

      <el-form
        :model="form"
        :rules="rules"
        ref="formRef"
        label-width="80px"
        label-position="top"
        @submit.prevent="onSubmit"
        class="auth-form"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
            :prefix-icon="User"
            class="auth-input"
          />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            :prefix-icon="Lock"
            show-password
            @keyup.enter="onSubmit"
            class="auth-input"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            @click="onSubmit"
            :loading="loading"
            class="auth-btn"
          >
            {{ loading ? '登录中...' : '立即登录' }}
          </el-button>
        </el-form-item>

        <el-form-item>
          <el-button
            type="text"
            @click="goToRegister"
            class="auth-link"
          >
            没有账号？立即注册
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <p class="auth-footer">© 2024 CEX Exchange · 专业安全的数字资产交易平台</p>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { login } from '../api/api'

const formRef = ref()
const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

// 表单验证规则
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

const onSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    loading.value = true

    const { data } = await login(form)

    // 保存token - 后端返回的数据在 data.data 中
    userStore.setToken(data.data.token)

    // 保存用户名
    userStore.setUsername(form.username)

    ElMessage.success('登录成功！')

    // 跳转到首页
    router.push('/')

  } catch (error) {
    console.error('登录失败:', error)

    if (error.response?.data?.error) {
      ElMessage.error(error.response.data.error)
    } else if (error.message) {
      ElMessage.error(error.message)
    } else {
      ElMessage.error('登录失败，请稍后重试')
    }
  } finally {
    loading.value = false
  }
}

const goToRegister = () => {
  router.push('/register')
}
</script>

<style scoped>
.auth-container {
  position: fixed;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--bg-app-gradient);
  overflow: hidden;
  padding: 24px;
}

/* Decorative orbs */
.auth-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.55;
  animation: float 28s ease-in-out infinite;
  z-index: 0;
}

.auth-orb-1 {
  width: 420px;
  height: 420px;
  background: radial-gradient(circle, rgba(59, 130, 246, 0.5) 0%, transparent 60%);
  top: 12%;
  left: -8%;
}

.auth-orb-2 {
  width: 360px;
  height: 360px;
  background: radial-gradient(circle, rgba(16, 185, 129, 0.45) 0%, transparent 60%);
  bottom: 10%;
  right: -6%;
  animation-duration: 34s;
}

@keyframes float {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(6%, -4%) scale(1.04); }
  66% { transform: translate(-4%, 6%) scale(0.98); }
}

.auth-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 410px;
  padding: 8px;
  box-shadow: var(--shadow-card-hover);
}

.auth-card :global(.el-card__header) {
  background: transparent !important;
  border-bottom: 1px solid var(--border-color) !important;
  padding: 0 12px 16px !important;
}

.auth-card :global(.el-card__body) { padding: 16px !important; }

.auth-header { text-align: center; }

.brand {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-weight: 800;
  font-size: 1.1rem;
  color: var(--text-main);
  margin-bottom: 4px;
}

.brand-dot {
  display: inline-block;
  width: 26px;
  height: 26px;
  border-radius: 6px;
  background: linear-gradient(135deg, var(--primary) 0%, var(--info) 100%);
  box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.5);
  animation: pulse-dot 2.2s ease-in-out infinite;
}

@keyframes pulse-dot {
  0% { box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.5); }
  70% { box-shadow: 0 0 0 10px rgba(59, 130, 246, 0); }
  100% { box-shadow: 0 0 0 0 rgba(59, 130, 246, 0); }
}

.auth-title {
  margin: 4px 0 2px;
  font-size: 1.6rem;
  font-weight: 700;
  color: var(--text-main);
}

.auth-subtitle {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-muted);
}

.auth-form { margin-top: 8px; }

.auth-form :global(.el-form-item) {
  margin-bottom: 20px;
}

.auth-form :global(.el-form-item__label) {
  padding-bottom: 6px;
  color: var(--text-muted);
  font-weight: 600;
}

.auth-input :global(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.8) !important;
  border-radius: var(--radius-input) !important;
  box-shadow: 0 0 0 1px var(--border-color) inset !important;
}

.auth-input :deep(.el-input__wrapper.is-focus) {
  background: #fff !important;
  box-shadow: 0 0 0 1px var(--primary) inset !important;
}

:root.dark .auth-input :global(.el-input__wrapper) {
  background: rgba(15, 23, 42, 0.48) !important;
}
:root.dark .auth-input :global(.el-input__wrapper.is-focus) {
  background: rgba(15, 23, 42, 0.82) !important;
}

.auth-btn {
  width: 100%;
  height: 44px;
  font-weight: 600;
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-emphasis) 100%);
  border: none;
  box-shadow: 0 6px 18px var(--primary-glow);
}

.auth-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 20px var(--primary-glow);
}

.auth-link {
  width: 100%;
  color: var(--primary) !important;
  font-weight: 500;
}

.auth-link:hover {
  color: var(--primary-emphasis) !important;
  background: var(--primary-glow) !important;
}

.auth-footer {
  position: relative;
  z-index: 1;
  margin-top: 24px;
  font-size: 0.78rem;
  color: var(--text-muted);
  text-align: center;
}

/* Responsive */
@media (max-width: 768px) {
  .auth-container { padding: 16px; }
  .auth-card { max-width: 100%; padding: 4px; }
}
</style>
