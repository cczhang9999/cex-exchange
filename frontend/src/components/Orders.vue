<template>
  <el-card>
    <h3>我的订单</h3>
    <el-table :data="orders">
      <el-table-column prop="symbol" label="交易对" />
      <el-table-column prop="side" label="方向" />
      <el-table-column prop="type" label="类型" />
      <el-table-column prop="price" label="价格" />
      <el-table-column prop="amount" label="数量" />
      <el-table-column prop="filled" label="已成交" />
      <el-table-column prop="status" label="状态" />
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button v-if="row.status==='open'" size="small" @click="cancelOrder(row)">撤单</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
  <el-card style="margin-top:20px">
    <h3>我的成交</h3>
    <el-table :data="trades">
      <el-table-column prop="symbol" label="交易对" />
      <el-table-column prop="price" label="价格" />
      <el-table-column prop="amount" label="数量" />
    </el-table>
  </el-card>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const orders = ref([])
const trades = ref([])

const fetchOrders = async () => {
  const { data } = await axios.get('/api/my_orders')
  orders.value = data
}
const fetchTrades = async () => {
  const { data } = await axios.get('/api/my_trades')
  trades.value = data
}
const cancelOrder = async (row) => {
  await axios.post(`/api/order/cancel/${row.id}`)
  fetchOrders()
}
onMounted(() => {
  fetchOrders()
  fetchTrades()
})
</script> 