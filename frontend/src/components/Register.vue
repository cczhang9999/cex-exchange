<template>
  <el-form :model="form" @submit.prevent="onSubmit" label-width="80px">
    <el-form-item label="用户名">
      <el-input v-model="form.username" />
    </el-form-item>
    <el-form-item label="密码">
      <el-input v-model="form.password" type="password" />
    </el-form-item>
    <el-form-item label="邮箱">
      <el-input v-model="form.email" />
    </el-form-item>
    <el-form-item label="手机号">
      <el-input v-model="form.phone" />
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="onSubmit">注册</el-button>
    </el-form-item>
  </el-form>
</template>
<script setup>
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores'
import axios from 'axios'

const form = reactive({ username: '', password: '', email: '', phone: '' })
const router = useRouter()
const userStore = useUserStore()

const onSubmit = async () => {
  try {
    const { data } = await axios.post('/api/register', form)
    userStore.setToken(data.token)
    router.push('/')
  } catch (e) {
    alert('注册失败')
  }
}
</script> 