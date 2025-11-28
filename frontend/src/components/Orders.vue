<template>
  <div class="orders-container">
    <el-card class="glass-panel mb-20">
      <template #header>
        <div class="flex-between">
          <h3>我的订单</h3>
          <el-tag type="info" effect="dark">Open Orders</el-tag>
        </div>
      </template>
      <el-table :data="orders" style="width: 100%">
        <el-table-column prop="symbol" label="交易对" />
        <el-table-column prop="side" label="方向">
          <template #default="{ row }">
            <span :class="row.side === 'buy' ? 'text-success' : 'text-danger'">
              {{ row.side === 'buy' ? '买入' : '卖出' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型">
          <template #default="{ row }">
            <el-tag size="small" effect="plain" class="glass-tag">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="price" label="价格" />
        <el-table-column prop="amount" label="数量" />
        <el-table-column prop="filled" label="已成交" />
        <el-table-column prop="status" label="状态">
          <template #default="{ row }">
            <el-tag :type="row.status === 'open' ? 'warning' : 'success'" effect="dark" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作">
          <template #default="{ row }">
            <el-button 
              v-if="row.status==='open'" 
              type="danger" 
              size="small" 
              plain
              @click="cancelOrder(row)"
            >
              撤单
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card class="glass-panel">
      <template #header>
        <div class="flex-between">
          <h3>我的成交</h3>
          <el-tag type="success" effect="dark">Trade History</el-tag>
        </div>
      </template>
      <el-table :data="trades" style="width: 100%">
        <el-table-column prop="symbol" label="交易对" />
        <el-table-column prop="price" label="价格" />
        <el-table-column prop="amount" label="数量" />
        <el-table-column prop="side" label="方向">
          <template #default="{ row }">
            <span :class="row.side === 'buy' ? 'text-success' : 'text-danger'">
              {{ row.side === 'buy' ? '买入' : '卖出' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间">
           <template #default="{ row }">
             <span class="text-muted">{{ new Date(row.created_at).toLocaleString() }}</span>
           </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import { getMyOrders, getMyTrades, cancelOrder as cancelOrderApi } from '../api/api'

const orders = ref([])
const trades = ref([])

const fetchOrders = async () => {
  const { data } = await getMyOrders()
  orders.value = data.data
}
const fetchTrades = async () => {
  const { data } = await getMyTrades()
  trades.value = data.data
}
const cancelOrder = async (row) => {
  await cancelOrderApi(row.id)
  fetchOrders()
}
onMounted(() => {
  fetchOrders()
  fetchTrades()
})
</script>
<style scoped>
.orders-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.mb-20 {
  margin-bottom: 20px;
}

.flex-between {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

h3 {
  margin: 0;
  font-weight: 600;
  color: var(--text-main);
}

.text-success {
  color: var(--success);
  font-weight: 600;
}

.text-danger {
  color: var(--danger);
  font-weight: 600;
}

.text-muted {
  color: var(--text-muted);
  font-size: 0.9em;
}

.glass-tag {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: var(--text-main);
}
</style> 