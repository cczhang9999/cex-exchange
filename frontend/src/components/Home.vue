<template>
  <!-- 交易中心全屏模式 -->
  <div v-if="currentContent === 'trade'" style="height: 100vh; background: #f0f2f5; position: relative;">
    <el-button type="primary" @click="showContent('home')" style="position: absolute; top: 20px; left: 20px; z-index: 10;">
      返回首页
    </el-button>
    <Trade />
  </div>
  
  <!-- 主布局模式 -->
  <el-container v-else style="height: 100vh; width: 100%;">
    <el-header style="height: 60px; padding: 0; background: #fff; border-bottom: 1px solid #e4e7ed;">
      <Header @nav="showContent" />
    </el-header>

    <el-container style="height: calc(100vh - 60px); width: 100%;">
      <el-aside width="200px" style="background: #fff; border-right: 1px solid #e4e7ed;">
        <SideBar :activeMenu="activeMenu" @menu="handleMenuSelect" />
      </el-aside>

      <el-main style="background: #f0f2f5; padding: 20px; overflow-y: auto; width: 100%;">
        <!-- 首页概览 -->
        <div v-if="currentContent === 'home'" style="height: 100%; overflow-y: auto;">
          <el-row :gutter="20">
            <el-col :span="8">
              <el-card class="glass-panel dashboard-card" shadow="hover">
                <template #header>
                  <div class="flex-center">
                    <el-icon class="text-primary"><TrendCharts /></el-icon>
                    <span>Market Overview</span>
                  </div>
                </template>
                <div class="text-center p-20">
                  <h3 class="text-primary">BTC/USDT</h3>
                  <p class="stat-value text-success">$45,123.45</p>
                  <p class="text-muted">+2.34%</p>
                </div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card class="glass-panel dashboard-card" shadow="hover">
                <template #header>
                  <div class="flex-center">
                    <el-icon class="text-warning"><Wallet /></el-icon>
                    <span>Total Assets</span>
                  </div>
                </template>
                <div class="text-center p-20">
                  <h3 class="text-warning">Balance</h3>
                  <p class="stat-value text-success">$12,345.67</p>
                  <p class="text-muted">24h Change</p>
                </div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card class="glass-panel dashboard-card" shadow="hover">
                <template #header>
                  <div class="flex-center">
                    <el-icon class="text-danger"><Document /></el-icon>
                    <span>Trading Stats</span>
                  </div>
                </template>
                <div class="text-center p-20">
                  <h3 class="text-danger">Today's Orders</h3>
                  <p class="stat-value text-success">156</p>
                  <p class="text-muted">Orders Executed</p>
                </div>
              </el-card>
            </el-col>
          </el-row>

          <el-row :gutter="20" class="mt-20">
            <el-col :span="12">
              <el-card class="glass-panel">
                <template #header>
                  <div class="flex-between">
                    <span>Quick Actions</span>
                  </div>
                </template>
                <el-row :gutter="15">
                  <el-col :span="12">
                    <el-button type="primary" class="w-100 mb-10" @click="showContent('trade')">
                      <el-icon class="mr-10"><TrendCharts /></el-icon>
                      Trade Now
                    </el-button>
                  </el-col>
                  <el-col :span="12">
                    <el-button type="success" class="w-100 mb-10" @click="showContent('assets')">
                      <el-icon class="mr-10"><Wallet /></el-icon>
                      Assets
                    </el-button>
                  </el-col>
                  <el-col :span="12">
                    <el-button type="warning" class="w-100 mb-10" @click="showContent('orders')">
                      <el-icon class="mr-10"><Document /></el-icon>
                      Orders
                    </el-button>
                  </el-col>
                  <el-col :span="12">
                    <el-button type="info" class="w-100 mb-10" @click="showContent('admin')">
                      <el-icon class="mr-10"><Setting /></el-icon>
                      Admin
                    </el-button>
                  </el-col>
                </el-row>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card class="glass-panel">
                <template #header>
                  <div class="flex-between">
                    <span>Announcements</span>
                  </div>
                </template>
                <div class="p-10">
                  <p class="notice-item">• System Maintenance: Jan 15, 02:00-04:00 UTC</p>
                  <p class="notice-item">• New Listing: ETH/USDT is now live!</p>
                  <p class="notice-item">• Zero Fee Trading Promotion</p>
                  <p class="notice-item">• Security Alert: Enable 2FA for your account</p>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </div>

        <!-- 用户登录 -->
        <div v-else-if="currentContent === 'login'">
          <Login />
        </div>

        <!-- 用户注册 -->
        <div v-else-if="currentContent === 'register'">
          <Register />
        </div>

        <!-- 资产管理 -->
        <div v-else-if="currentContent === 'assets'">
          <Assets />
        </div>

        <!-- 订单管理 -->
        <div v-else-if="currentContent === 'orders'">
          <Orders />
        </div>

        <!-- 后台管理 -->
        <div v-else-if="currentContent === 'admin'">
          <Admin />
        </div>
        
        <!-- 实时行情监控 -->
        <div v-else-if="currentContent === 'realtime'">
          <RealTimeMarket />
        </div>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref } from 'vue'
import { House, User, UserFilled, Wallet, TrendCharts, Document, Setting, ArrowUp, ArrowDown } from '@element-plus/icons-vue'
import Trade from './Trade.vue'
import Header from './Header.vue'
import SideBar from './SideBar.vue'
import Assets from './Assets.vue'
import Admin from './Admin.vue'
import RealTimeMarket from './RealTimeMarket.vue'
import Login from './Login.vue'
import Register from './Register.vue'
import Orders from './Orders.vue'

// 当前显示的内容
const currentContent = ref('home')
const activeMenu = ref('home')

// 表单数据（保留用于其他功能）
const loginForm = ref({
  username: '',
  password: ''
})

const registerForm = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const tradeForm = ref({
  symbol: 'BTC/USDT',
  price: '',
  amount: ''
})

// 模拟数据
const assetsData = ref([
  { asset: 'BTC', balance: '0.1234', frozen: '0.0000', total: '0.1234', value: '5,678.90' },
  { asset: 'ETH', balance: '2.5678', frozen: '0.0000', total: '2.5678', value: '4,567.89' },
  { asset: 'USDT', balance: '1000.00', frozen: '0.00', total: '1000.00', value: '1000.00' }
])

const ordersData = ref([
  { id: '1001', symbol: 'BTC/USDT', side: 'buy', price: '45,123.45', amount: '0.001', filled: '0.001', status: 'filled', createdAt: '2024-01-15 10:30:00' },
  { id: '1002', symbol: 'ETH/USDT', side: 'sell', price: '2,345.67', amount: '0.1', filled: '0.05', status: 'partially_filled', createdAt: '2024-01-15 09:15:00' },
  { id: '1003', symbol: 'BTC/USDT', side: 'buy', price: '44,500.00', amount: '0.002', filled: '0.000', status: 'open', createdAt: '2024-01-15 08:45:00' }
])

// 处理菜单选择
const handleMenuSelect = (index) => {
  currentContent.value = index
  activeMenu.value = index
}

// 显示内容
const showContent = (content) => {
  currentContent.value = content
  activeMenu.value = content
}

// 获取状态类型
const getStatusType = (status) => {
  const types = {
    'open': 'warning',
    'partially_filled': 'info',
    'filled': 'success',
    'cancelled': 'danger'
  }
  return types[status] || 'info'
}

// 获取状态文本
const getStatusText = (status) => {
  const texts = {
    'open': '未成交',
    'partially_filled': '部分成交',
    'filled': '已成交',
    'cancelled': '已撤单'
  }
  return texts[status] || status
}

// 计算总额
const calculateTotal = () => {
  const price = parseFloat(tradeForm.value.price) || 0;
  const amount = parseFloat(tradeForm.value.amount) || 0;
  return (price * amount).toFixed(2);
};

// 设置数量为百分比
const setPercentage = (percentage) => {
  const price = parseFloat(tradeForm.value.price) || 0;
  const amount = (parseFloat(tradeForm.value.amount) || 0) * (percentage / 100);
  tradeForm.value.amount = amount.toFixed(4);
};

// 设置为市价
const setMarketPrice = () => {
  tradeForm.value.price = '0'; // 假设市价为0
  tradeForm.value.amount = '0';
};

// 清空表单
const clearForm = () => {
  tradeForm.value.price = '';
  tradeForm.value.amount = '';
};
</script>

<style scoped>
.dashboard-card {
  height: 220px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  margin: 10px 0;
}

.text-primary { color: var(--primary); }
.text-success { color: var(--success); }
.text-warning { color: var(--warning); }
.text-danger { color: var(--danger); }
.text-muted { color: var(--text-muted); }

.mt-20 { margin-top: 20px; }
.mb-10 { margin-bottom: 10px; }
.mr-10 { margin-right: 10px; }
.w-100 { width: 100%; }

.notice-item {
  color: var(--text-muted);
  margin: 8px 0;
  padding: 8px;
  border-radius: 6px;
  transition: background 0.3s;
}

.notice-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-main);
}

.el-header {
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.el-aside {
  box-shadow: 2px 0 4px rgba(0, 0, 0, 0.1);
}
</style> 
