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
          <h2 class="auth-title">创建账号</h2>
          <p class="auth-subtitle">填写信息以开始您的交易之旅</p>
        </div>
      </template>

      <el-form
        :model="form"
        :rules="rules"
        ref="formRef"
        label-width="100px"
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

        <el-form-item label="邮箱" prop="email">
          <el-input
            v-model="form.email"
            placeholder="请输入邮箱"
            :prefix-icon="Message"
            class="auth-input"
          />
        </el-form-item>

        <el-form-item label="手机号" prop="phone">
          <el-input
            v-model="form.phone"
            placeholder="请输入手机号"
            :prefix-icon="Phone"
            class="auth-input"
          />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码 (6-20 位)"
            :prefix-icon="Lock"
            show-password
            class="auth-input"
          />
        </el-form-item>

        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            placeholder="请确认密码"
            :prefix-icon="Lock"
            show-password
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
            {{ loading ? '注册中...' : '立即注册' }}
          </el-button>
        </el-form-item>

        <el-form-item>
          <el-button
            type="text"
            @click="goToLogin"
            class="auth-link"
          >
            已有账号？立即登录
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { User, Message, Phone, Lock } from '@element-plus/icons-vue'
import { register } from '../api/api'

const formRef = ref()
const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
  email: '',
  phone: '',
  confirmPassword: ''
})

// 表单验证规则
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号格式', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== form.password) {
          callback(new Error('两次输入密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

const onSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    loading.value = true

    const { data } = await register({
      username: form.username,
      password: form.password,
      email: form.email,
      phone: form.phone
    })

    // 保存token - 后端返回的数据在 data.data 中
    userStore.setToken(data.data.token)

    ElMessage.success('注册成功！')

    // 跳转到首页
    router.push('/')

  } catch (error) {
    console.error('注册失败:', error)

    if (error.response?.data?.error) {
      ElMessage.error(error.response.data.error)
    } else if (error.message) {
      ElMessage.error(error.message)
    } else {
      ElMessage.error('注册失败，请稍后重试')
    }
  } finally {
    loading.value = false
  }
}

const goToLogin = () => {
  router.push('/login')
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

.auth-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.55;
  animation: float 30s ease-in-out infinite;
  z-index: 0;
}

.auth-orb-1 {
  width: 460px;
  height: 460px;
  background: radial-gradient(circle, rgba(59, 130, 246, 0.5) 0%, transparent 60%);
  top: 8%;
  left: -10%;
}

.auth-orb-2 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(16, 185, 129, 0.45) 0%, transparent 60%);
  bottom: 6%;
  right: -8%;
  animation-duration: 36s;
}

@keyframes float {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(5%, -5%) scale(1.03); }
  66% { transform: translate(-5%, 6%) scale(0.97); }
}

.auth-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 470px;
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

@media (max-width: 768px) {
  .auth-container { padding: 16px; }
  .auth-card { max-width: 100%; padding: 4px; }
}
</style>
