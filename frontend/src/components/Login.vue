<template>
  <el-form :model="form" @submit.prevent="onSubmit" label-width="80px">
    <el-form-item label="用户名">
      <el-input v-model="form.username" />
    </el-form-item>
    <el-form-item label="密码">
      <el-input v-model="form.password" type="password" />
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="onSubmit">登录</el-button>
    </el-form-item>
  </el-form>
</template>
<script setup>
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores'
import { login } from '../api/api.js'
import { ElMessage } from 'element-plus'

const form = reactive({ username: '', password: '' })
const router = useRouter()
const userStore = useUserStore()

const onSubmit = async () => {
  try {
    const { data } = await login(form)
    userStore.setToken(data.token)
    // 可选：alert('登录成功')
    router.push('/') // 跳转到首页
  } catch (e) {
    ElMessage.error('登录失败')
  }
}
</script> 