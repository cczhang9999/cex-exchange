<template>
  <div class="login-container">
    <el-card class="login-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <h2>用户登录</h2>
        </div>
      </template>
      
      <el-form 
        :model="form" 
        :rules="rules" 
        ref="formRef" 
        label-width="80px" 
        @submit.prevent="onSubmit"
        class="login-form"
      >
        <el-form-item label="用户名" prop="username">
          <el-input 
            v-model="form.username" 
            placeholder="请输入用户名"
            :prefix-icon="User"
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
          />
        </el-form-item>
        
        <el-form-item>
          <el-button 
            type="primary" 
            @click="onSubmit" 
            :loading="loading"
            style="width: 100%;"
          >
            {{ loading ? '登录中...' : '登录' }}
          </el-button>
        </el-form-item>
        
        <el-form-item>
          <el-button 
            type="text" 
            @click="goToRegister"
            style="width: 100%;"
          >
            没有账号？立即注册
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
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
    
    // 保存token
    userStore.setToken(data.token)
    
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
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.card-header {
  text-align: center;
}

.card-header h2 {
  margin: 0;
  color: #303133;
  font-weight: 600;
}

.login-form {
  margin-top: 20px;
}

.el-form-item {
  margin-bottom: 20px;
}

.el-input {
  border-radius: 8px;
}

.el-button {
  border-radius: 8px;
  font-weight: 500;
  height: 44px;
}

.el-button--primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
}

.el-button--primary:hover {
  background: linear-gradient(135deg, #5a6fd8 0%, #6a4190 100%);
  transform: translateY(-1px);
}

.el-button--text {
  color: #667eea;
}

.el-button--text:hover {
  color: #5a6fd8;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .login-container {
    padding: 10px;
  }
  
  .login-card {
    max-width: 100%;
  }
}
</style> 