<template>
  <div class="orders-container">
    <el-card class="glass-panel content-card mb-20">
      <template #header>
        <div class="flex-between">
          <h3 class="page-title-sm">我的订单</h3>
          <el-tag type="info" effect="dark" class="glass-tag">Open Orders</el-tag>
        </div>
      </template>
      <el-table :data="orders" class="orders-table w-100">
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
            <el-tag :type="row.status === 'open' ? 'warning' : 'success'" effect="dark" size="small" class="status-tag">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间">
          <template #default="{ row }">
            <span class="text-muted mono">{{ formatDate(row.created_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="110">
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

    <el-card class="glass-panel content-card">
      <template #header>
        <div class="flex-between">
          <h3 class="page-title-sm">我的成交</h3>
          <el-tag type="success" effect="dark" class="glass-tag">Trade History</el-tag>
        </div>
      </template>
      <el-table :data="trades" class="orders-table w-100">
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
             <span class="text-muted mono">{{ formatDate(row.created_at) }}</span>
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

// 格式化时间显示
const formatDate = (timestamp) => {
  if (!timestamp) return ''

  // 检查时间戳是否是毫秒格式（13位数字）
  // 如果是秒格式（10位数字），则乘以1000转换为毫秒
  let timestampMs = timestamp
  if (String(timestamp).length <= 10) {
    timestampMs = timestamp * 1000
  }

  const date = new Date(timestampMs)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

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
  padding: 8px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-title-sm {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 600;
  color: var(--text-main);
}

.orders-table {
  border-radius: var(--radius-base);
}

.orders-table :global(.el-table__row):hover {
  background: var(--primary-glow) !important;
}

.status-tag {
  font-weight: 700;
}

/* shared utilities (scoped fallback) */
.mb-20 { margin-bottom: 20px; }
.flex-between { display: flex; justify-content: space-between; align-items: center; }
.text-success { color: var(--success); font-weight: 600; }
.text-danger { color: var(--danger); font-weight: 600; }
.text-muted { color: var(--text-muted); }
.mono { font-family: 'Roboto Mono', 'Fira Code', monospace; }
.content-card { background: var(--bg-card); border-radius: var(--radius-card); }
</style>
