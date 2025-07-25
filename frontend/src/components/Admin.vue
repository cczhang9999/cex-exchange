<template>
  <el-tabs v-model="tab">
    <el-tab-pane label="用户管理" name="users">
      <el-table :data="users">
        <el-table-column prop="id" label="ID" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="phone" label="手机号" />
        <el-table-column prop="status" label="状态" />
      </el-table>
      <el-pagination :total="userTotal" :page-size="pageSize" v-model:current-page="userPage" @current-change="fetchUsers" />
    </el-tab-pane>
    <el-tab-pane label="订单管理" name="orders">
      <el-table :data="orders">
        <el-table-column prop="id" label="ID" />
        <el-table-column prop="user_id" label="用户ID" />
        <el-table-column prop="symbol" label="交易对" />
        <el-table-column prop="side" label="方向" />
        <el-table-column prop="type" label="类型" />
        <el-table-column prop="price" label="价格" />
        <el-table-column prop="amount" label="数量" />
        <el-table-column prop="status" label="状态" />
      </el-table>
      <el-pagination :total="orderTotal" :page-size="pageSize" v-model:current-page="orderPage" @current-change="fetchOrders" />
    </el-tab-pane>
    <el-tab-pane label="资金账户管理" name="accounts">
      <el-table :data="accounts">
        <el-table-column prop="id" label="ID" />
        <el-table-column prop="user_id" label="用户ID" />
        <el-table-column prop="asset" label="币种" />
        <el-table-column prop="balance" label="可用余额" />
        <el-table-column prop="frozen" label="冻结余额" />
      </el-table>
      <el-pagination :total="accountTotal" :page-size="pageSize" v-model:current-page="accountPage" @current-change="fetchAccounts" />
    </el-tab-pane>
  </el-tabs>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const tab = ref('users')
const users = ref([])
const orders = ref([])
const accounts = ref([])
const userPage = ref(1)
const orderPage = ref(1)
const accountPage = ref(1)
const pageSize = 20
const userTotal = ref(0)
const orderTotal = ref(0)
const accountTotal = ref(0)

const fetchUsers = async () => {
  const { data } = await axios.get('/api/admin/users', { params: { page: userPage.value, limit: pageSize } })
  users.value = data
  // userTotal.value = data.total // 如后端返回total可用
}
const fetchOrders = async () => {
  const { data } = await axios.get('/api/admin/orders', { params: { page: orderPage.value, limit: pageSize } })
  orders.value = data
  // orderTotal.value = data.total
}
const fetchAccounts = async () => {
  const { data } = await axios.get('/api/admin/accounts', { params: { page: accountPage.value, limit: pageSize } })
  accounts.value = data
  // accountTotal.value = data.total
}
onMounted(() => {
  fetchUsers()
  fetchOrders()
  fetchAccounts()
})
</script> 