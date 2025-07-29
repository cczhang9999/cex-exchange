<template>
  <!-- 交易中心全屏模式 -->
  <div v-if="currentContent === 'trade'" style="height: 100vh; background: #f0f2f5; position: relative;">
    <el-button type="primary" @click="showContent('home')" style="position: absolute; top: 20px; left: 20px; z-index: 10;">
      返回首页
    </el-button>
    <Trade />
  </div>
  
  <!-- 主布局模式 -->
  <el-container v-else style="height: 100vh;">
    <el-header style="height: 60px; padding: 0; background: #fff; border-bottom: 1px solid #e4e7ed;">
      <Header @nav="showContent" />
    </el-header>

    <el-container style="height: calc(100vh - 60px);">
      <el-aside width="200px" style="background: #fff; border-right: 1px solid #e4e7ed;">
        <SideBar :activeMenu="activeMenu" @menu="handleMenuSelect" />
      </el-aside>

      <el-main style="background: #f0f2f5; padding: 20px; overflow-y: auto;">
        <!-- 首页概览 -->
        <div v-if="currentContent === 'home'" style="height: 100%; overflow-y: auto;">
          <el-row :gutter="20">
            <el-col :span="8">
              <el-card shadow="hover" style="height: 200px;">
                <template #header>
                  <div style="display: flex; align-items: center;">
                    <el-icon style="margin-right: 8px;"><TrendCharts /></el-icon>
                    <span>市场概览</span>
                  </div>
                </template>
                <div style="text-align: center; padding: 20px;">
                  <h3 style="color: #409EFF;">BTC/USDT</h3>
                  <p style="font-size: 24px; color: #67C23A;">$45,123.45</p>
                  <p style="color: #909399;">+2.34%</p>
                </div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card shadow="hover" style="height: 200px;">
                <template #header>
                  <div style="display: flex; align-items: center;">
                    <el-icon style="margin-right: 8px;"><Wallet /></el-icon>
                    <span>资产总览</span>
                  </div>
                </template>
                <div style="text-align: center; padding: 20px;">
                  <h3 style="color: #E6A23C;">总资产</h3>
                  <p style="font-size: 24px; color: #67C23A;">$12,345.67</p>
                  <p style="color: #909399;">24小时变化</p>
                </div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card shadow="hover" style="height: 200px;">
                <template #header>
                  <div style="display: flex; align-items: center;">
                    <el-icon style="margin-right: 8px;"><Document /></el-icon>
                    <span>交易统计</span>
                  </div>
                </template>
                <div style="text-align: center; padding: 20px;">
                  <h3 style="color: #F56C6C;">今日交易</h3>
                  <p style="font-size: 24px; color: #67C23A;">156</p>
                  <p style="color: #909399;">笔订单</p>
                </div>
              </el-card>
            </el-col>
          </el-row>

          <el-row :gutter="20" style="margin-top: 20px;">
            <el-col :span="12">
              <el-card shadow="hover">
                <template #header>
                  <div style="display: flex; align-items: center; justify-content: space-between;">
                    <span>快速操作</span>
                  </div>
                </template>
                <el-row :gutter="10">
                  <el-col :span="12">
                    <el-button type="primary" style="width: 100%; margin-bottom: 10px;" @click="showContent('trade')">
                      <el-icon style="margin-right: 5px;"><TrendCharts /></el-icon>
                      开始交易
                    </el-button>
                  </el-col>
                  <el-col :span="12">
                    <el-button type="success" style="width: 100%; margin-bottom: 10px;" @click="showContent('assets')">
                      <el-icon style="margin-right: 5px;"><Wallet /></el-icon>
                      资产管理
                    </el-button>
                  </el-col>
                  <el-col :span="12">
                    <el-button type="warning" style="width: 100%; margin-bottom: 10px;" @click="showContent('orders')">
                      <el-icon style="margin-right: 5px;"><Document /></el-icon>
                      订单管理
                    </el-button>
                  </el-col>
                  <el-col :span="12">
                    <el-button type="info" style="width: 100%; margin-bottom: 10px;" @click="showContent('admin')">
                      <el-icon style="margin-right: 5px;"><Setting /></el-icon>
                      后台管理
                    </el-button>
                  </el-col>
                </el-row>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card shadow="hover">
                <template #header>
                  <div style="display: flex; align-items: center; justify-content: space-between;">
                    <span>系统公告</span>
                  </div>
                </template>
                <div style="padding: 10px;">
                  <p style="color: #606266; margin: 5px 0;">• 系统维护通知：2024年1月15日 02:00-04:00</p>
                  <p style="color: #606266; margin: 5px 0;">• 新增ETH交易对，欢迎体验</p>
                  <p style="color: #606266; margin: 5px 0;">• 手续费优惠活动进行中</p>
                  <p style="color: #606266; margin: 5px 0;">• 安全提醒：请妥善保管您的账户信息</p>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </div>

        <!-- 用户登录 -->
        <div v-else-if="currentContent === 'login'">
          <el-card shadow="hover" style="max-width: 400px; margin: 0 auto;">
            <template #header>
              <div style="text-align: center;">
                <h3>用户登录</h3>
              </div>
            </template>
            <el-form :model="loginForm" label-width="80px">
              <el-form-item label="用户名">
                <el-input v-model="loginForm.username" placeholder="请输入用户名"></el-input>
              </el-form-item>
              <el-form-item label="密码">
                <el-input v-model="loginForm.password" type="password" placeholder="请输入密码"></el-input>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" style="width: 100%;">登录</el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </div>

        <!-- 用户注册 -->
        <div v-else-if="currentContent === 'register'">
          <el-card shadow="hover" style="max-width: 400px; margin: 0 auto;">
            <template #header>
              <div style="text-align: center;">
                <h3>用户注册</h3>
              </div>
            </template>
            <el-form :model="registerForm" label-width="80px">
              <el-form-item label="用户名">
                <el-input v-model="registerForm.username" placeholder="请输入用户名"></el-input>
              </el-form-item>
              <el-form-item label="邮箱">
                <el-input v-model="registerForm.email" placeholder="请输入邮箱"></el-input>
              </el-form-item>
              <el-form-item label="密码">
                <el-input v-model="registerForm.password" type="password" placeholder="请输入密码"></el-input>
              </el-form-item>
              <el-form-item label="确认密码">
                <el-input v-model="registerForm.confirmPassword" type="password" placeholder="请确认密码"></el-input>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" style="width: 100%;">注册</el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </div>

        <!-- 资产管理 -->
        <div v-else-if="currentContent === 'assets'">
          <Assets />
        </div>

        <!-- 订单管理 -->
        <div v-else-if="currentContent === 'orders'">
          <el-card shadow="hover">
            <template #header>
              <div style="display: flex; align-items: center; justify-content: space-between;">
                <span>订单管理</span>
                <el-button type="primary" size="small">刷新</el-button>
              </div>
            </template>
            <el-table :data="ordersData" style="width: 100%">
              <el-table-column prop="id" label="订单ID" width="100" />
              <el-table-column prop="symbol" label="交易对" width="120" />
              <el-table-column prop="side" label="方向" width="80">
                <template #default="scope">
                  <el-tag :type="scope.row.side === 'buy' ? 'success' : 'danger'">
                    {{ scope.row.side === 'buy' ? '买入' : '卖出' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="price" label="价格" width="120" />
              <el-table-column prop="amount" label="数量" width="120" />
              <el-table-column prop="filled" label="已成交" width="120" />
              <el-table-column prop="status" label="状态" width="120">
                <template #default="scope">
                  <el-tag 
                    :type="scope.row.status === 'filled' ? 'success' : 
                           scope.row.status === 'partially_filled' ? 'warning' : 'info'"
                  >
                    {{ scope.row.status === 'filled' ? '已完成' : 
                       scope.row.status === 'partially_filled' ? '部分成交' : '待成交' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="createdAt" label="创建时间" width="180" />
              <el-table-column label="操作" width="120">
                <template #default="scope">
                  <el-button 
                    v-if="scope.row.status !== 'filled'" 
                    type="danger" 
                    size="small"
                  >
                    取消
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
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

// 当前显示的内容
const currentContent = ref('home')
const activeMenu = ref('home')

// 表单数据
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
.el-header {
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.el-aside {
  box-shadow: 2px 0 4px rgba(0, 0, 0, 0.1);
}

.el-card {
  transition: all 0.3s;
}

.el-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.el-button {
  transition: all 0.3s;
}

.el-button:hover {
  transform: translateY(-1px);
}
</style> 
