<template>
  <el-card shadow="hover">
    <template #header>
      <div style="display: flex; align-items: center;">
        <el-icon style="margin-right: 8px;"><Setting /></el-icon>
        <span>后台管理</span>
      </div>
    </template>
    
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <el-tab-pane label="用户管理" name="users">
        <div style="margin-bottom: 20px;">
          <el-input 
            v-model="userSearch" 
            placeholder="搜索用户名或邮箱" 
            style="width: 300px; margin-right: 10px;"
            @input="searchUsers"
          />
          <el-button type="primary" @click="fetchUsers">刷新</el-button>
        </div>
        
        <el-table :data="users" v-loading="userLoading" style="width: 100%">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户名" />
          <el-table-column prop="email" label="邮箱" />
          <el-table-column prop="phone" label="手机号" />
          <el-table-column prop="is_blocked" label="状态">
            <template #default="{ row }">
              <el-tag :type="row.is_blocked ? 'danger' : 'success'">
                {{ row.is_blocked ? '已封禁' : '正常' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="注册时间" />
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button 
                :type="row.is_blocked ? 'success' : 'danger'" 
                size="small"
                @click="toggleUserStatus(row)"
              >
                {{ row.is_blocked ? '解封' : '封禁' }}
              </el-button>
              <el-button type="primary" size="small" @click="viewUserDetail(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
        
        <el-pagination 
          :total="userTotal" 
          :page-size="pageSize" 
          v-model:current-page="userPage" 
          @current-change="fetchUsers"
          style="margin-top: 20px; text-align: center;"
        />
      </el-tab-pane>

      <el-tab-pane label="订单管理" name="orders">
        <div style="margin-bottom: 20px;">
          <el-select v-model="orderStatusFilter" placeholder="订单状态" style="width: 120px; margin-right: 10px;">
            <el-option label="全部" value="" />
            <el-option label="开放" value="open" />
            <el-option label="已成交" value="filled" />
            <el-option label="已取消" value="cancelled" />
          </el-select>
          <el-input 
            v-model="orderSymbolFilter" 
            placeholder="交易对" 
            style="width: 150px; margin-right: 10px;"
          />
          <el-button type="primary" @click="fetchOrders">查询</el-button>
          <el-button @click="resetOrderFilters">重置</el-button>
        </div>
        
        <el-table :data="orders" v-loading="orderLoading" style="width: 100%">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="user_id" label="用户ID" width="100" />
          <el-table-column prop="symbol" label="交易对" />
          <el-table-column prop="side" label="方向">
            <template #default="{ row }">
              <el-tag :type="row.side === 'buy' ? 'success' : 'danger'">
                {{ row.side === 'buy' ? '买入' : '卖出' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="type" label="类型" />
          <el-table-column prop="price" label="价格" />
          <el-table-column prop="amount" label="数量" />
          <el-table-column prop="filled" label="已成交" />
          <el-table-column prop="status" label="状态">
            <template #default="{ row }">
              <el-tag :type="getOrderStatusType(row.status)">
                {{ getOrderStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" />
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button 
                v-if="row.status === 'open'" 
                type="danger" 
                size="small"
                @click="cancelOrder(row)"
              >
                取消
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        
        <el-pagination 
          :total="orderTotal" 
          :page-size="pageSize" 
          v-model:current-page="orderPage" 
          @current-change="fetchOrders"
          style="margin-top: 20px; text-align: center;"
        />
      </el-tab-pane>

      <el-tab-pane label="资金账户管理" name="accounts">
        <div style="margin-bottom: 20px;">
          <el-input 
            v-model="accountUserFilter" 
            placeholder="用户ID" 
            style="width: 150px; margin-right: 10px;"
          />
          <el-select v-model="accountAssetFilter" placeholder="币种" style="width: 120px; margin-right: 10px;">
            <el-option label="全部" value="" />
            <el-option label="BTC" value="BTC" />
            <el-option label="ETH" value="ETH" />
            <el-option label="USDT" value="USDT" />
          </el-select>
          <el-button type="primary" @click="fetchAccounts">查询</el-button>
          <el-button @click="resetAccountFilters">重置</el-button>
          <el-button type="success" @click="openAddFunds">添加资金</el-button>
        </div>
        
        <el-table :data="accounts" v-loading="accountLoading" style="width: 100%">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="user_id" label="用户ID" width="100" />
          <el-table-column prop="asset" label="币种" />
          <el-table-column prop="balance" label="可用余额" />
          <el-table-column prop="frozen" label="冻结余额" />
          <el-table-column label="总余额">
            <template #default="{ row }">
              {{ (parseFloat(row.balance) + parseFloat(row.frozen)).toFixed(8) }}
            </template>
          </el-table-column>
          <el-table-column prop="updated_at" label="更新时间" />
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="adjustBalance(row)">调整余额</el-button>
              <el-button type="success" size="small" @click="addFundsToAccount(row)">添加资金</el-button>
            </template>
          </el-table-column>
        </el-table>
        
        <el-pagination 
          :total="accountTotal" 
          :page-size="pageSize" 
          v-model:current-page="accountPage" 
          @current-change="fetchAccounts"
          style="margin-top: 20px; text-align: center;"
        />
      </el-tab-pane>
    </el-tabs>

    <!-- 用户详情对话框 -->
    <el-dialog v-model="showUserDetail" title="用户详情" width="600px">
      <div v-if="selectedUser">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="用户ID">{{ selectedUser.id }}</el-descriptions-item>
          <el-descriptions-item label="用户名">{{ selectedUser.username }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ selectedUser.email }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ selectedUser.phone || '未设置' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="selectedUser.is_blocked ? 'danger' : 'success'">
              {{ selectedUser.is_blocked ? '已封禁' : '正常' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="注册时间">{{ selectedUser.created_at }}</el-descriptions-item>
          <el-descriptions-item label="最后登录">{{ selectedUser.last_login || '从未登录' }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>

    <!-- 调整余额对话框 -->
    <el-dialog v-model="showAdjustBalance" title="调整余额" width="400px">
      <el-form :model="adjustForm" label-width="100px">
        <el-form-item label="用户ID">
          <el-input v-model="adjustForm.user_id" disabled />
        </el-form-item>
        <el-form-item label="币种">
          <el-input v-model="adjustForm.asset" disabled />
        </el-form-item>
        <el-form-item label="当前余额">
          <el-input v-model="adjustForm.current_balance" disabled />
        </el-form-item>
        <el-form-item label="调整类型">
          <el-radio-group v-model="adjustForm.type">
            <el-radio label="add">增加</el-radio>
            <el-radio label="subtract">减少</el-radio>
            <el-radio label="set">设置为</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="金额">
          <el-input v-model="adjustForm.amount" placeholder="请输入金额" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="adjustForm.remark" placeholder="调整原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAdjustBalance = false">取消</el-button>
        <el-button type="primary" @click="confirmAdjustBalance">确认</el-button>
      </template>
    </el-dialog>

    <!-- 添加资金对话框 -->
    <el-dialog v-model="showAddFunds" title="添加资金" width="500px">
      <el-form :model="addFundsForm" :rules="addFundsRules" ref="addFundsFormRef" label-width="100px">
        <el-form-item label="用户ID" prop="user_id">
          <el-input v-model="addFundsForm.user_id" placeholder="请输入用户ID" />
        </el-form-item>
        <el-form-item label="币种" prop="asset">
          <el-select v-model="addFundsForm.asset" placeholder="请选择币种" style="width: 100%;">
            <el-option label="BTC" value="BTC" />
            <el-option label="ETH" value="ETH" />
            <el-option label="USDT" value="USDT" />
            <el-option label="BNB" value="BNB" />
            <el-option label="ADA" value="ADA" />
            <el-option label="DOT" value="DOT" />
          </el-select>
        </el-form-item>
        <el-form-item label="添加金额" prop="amount">
          <el-input v-model="addFundsForm.amount" placeholder="请输入添加金额" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input 
            v-model="addFundsForm.remark" 
            type="textarea" 
            :rows="3"
            placeholder="请输入备注信息"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddFunds = false">取消</el-button>
        <el-button type="primary" @click="confirmAddFunds">确认添加</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Setting } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { getUsers } from '../api/api'

// 当前活跃的标签页
const activeTab = ref('users')

// 分页配置
const pageSize = 20

// 用户管理相关
const users = ref([])
const userPage = ref(1)
const userTotal = ref(0)
const userLoading = ref(false)
const userSearch = ref('')
const showUserDetail = ref(false)
const selectedUser = ref(null)

// 订单管理相关
const orders = ref([])
const orderPage = ref(1)
const orderTotal = ref(0)
const orderLoading = ref(false)
const orderStatusFilter = ref('')
const orderSymbolFilter = ref('')

// 账户管理相关
const accounts = ref([])
const accountPage = ref(1)
const accountTotal = ref(0)
const accountLoading = ref(false)
const accountUserFilter = ref('')
const accountAssetFilter = ref('')
const showAdjustBalance = ref(false)
const adjustForm = ref({
  user_id: '',
  asset: '',
  current_balance: '',
  type: 'add',
  amount: '',
  remark: ''
})

// 添加资金相关
const showAddFunds = ref(false)
const addFundsForm = ref({
  user_id: '',
  asset: '',
  amount: '',
  remark: ''
})
const addFundsFormRef = ref(null)

// 添加资金表单验证规则
const addFundsRules = {
  user_id: [
    { required: true, message: '请输入用户ID', trigger: 'blur' },
    { pattern: /^\d+$/, message: '用户ID必须是数字', trigger: 'blur' }
  ],
  asset: [
    { required: true, message: '请选择币种', trigger: 'change' }
  ],
  amount: [
    { required: true, message: '请输入添加金额', trigger: 'blur' },
    { pattern: /^\d+(\.\d+)?$/, message: '金额格式不正确', trigger: 'blur' },
    { 
      validator: (rule, value, callback) => {
        if (parseFloat(value) <= 0) {
          callback(new Error('金额必须大于0'))
        } else {
          callback()
        }
      }, 
      trigger: 'blur' 
    }
  ],
  remark: [
    { required: true, message: '请输入备注信息', trigger: 'blur' }
  ]
}

// 获取用户列表
const fetchUsers = async () => {
  userLoading.value = true
  try {
    const params = {
      page: userPage.value,
      limit: pageSize
    }
    if (userSearch.value) {
      params.search = userSearch.value
    }
  
    //const { data } = await axios.get('/api/admin/users', { params })
    const { data } = await getUsers({ params})
    users.value = data.data.users || data
    userTotal.value = data.data.total || 0
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  } finally {
    userLoading.value = false
  }
}

// 搜索用户
const searchUsers = () => {
  userPage.value = 1
  fetchUsers()
}

// 切换用户状态
const toggleUserStatus = async (user) => {
  try {
    const action = user.is_blocked ? 'unblock' : 'block'
    const actionText = user.is_blocked ? '解封' : '封禁'
    
    await ElMessageBox.confirm(`确定要${actionText}用户 ${user.username} 吗？`, '确认操作')
    
    await axios.post(`/api/admin/users/${user.id}/${action}`)
    ElMessage.success(`${actionText}成功`)
    fetchUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 查看用户详情
const viewUserDetail = (user) => {
  selectedUser.value = user
  showUserDetail.value = true
}

// 获取订单列表
const fetchOrders = async () => {
  orderLoading.value = true
  try {
    const params = {
      page: orderPage.value,
      limit: pageSize
    }
    if (orderStatusFilter.value) params.status = orderStatusFilter.value
    if (orderSymbolFilter.value) params.symbol = orderSymbolFilter.value
    
    const { data } = await axios.get('/api/admin/orders', { params })
    orders.value = data.orders || data
    orderTotal.value = data.total || 0
  } catch (error) {
    ElMessage.error('获取订单列表失败')
  } finally {
    orderLoading.value = false
  }
}

// 重置订单过滤器
const resetOrderFilters = () => {
  orderStatusFilter.value = ''
  orderSymbolFilter.value = ''
  orderPage.value = 1
  fetchOrders()
}

// 取消订单
const cancelOrder = async (order) => {
  try {
    await ElMessageBox.confirm(`确定要取消订单 ${order.id} 吗？`, '确认操作')
    await axios.post(`/api/admin/orders/${order.id}/cancel`)
    ElMessage.success('订单取消成功')
    fetchOrders()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('取消订单失败')
    }
  }
}

// 获取账户列表
const fetchAccounts = async () => {
  accountLoading.value = true
  try {
    const params = {
      page: accountPage.value,
      limit: pageSize
    }
    if (accountUserFilter.value) params.user_id = accountUserFilter.value
    if (accountAssetFilter.value) params.asset = accountAssetFilter.value
    
    const { data } = await axios.get('/api/admin/accounts', { params })
    accounts.value = data.accounts || data
    accountTotal.value = data.total || 0
  } catch (error) {
    ElMessage.error('获取账户列表失败')
  } finally {
    accountLoading.value = false
  }
}

// 重置账户过滤器
const resetAccountFilters = () => {
  accountUserFilter.value = ''
  accountAssetFilter.value = ''
  accountPage.value = 1
  fetchAccounts()
}

// 调整余额
const adjustBalance = (account) => {
  adjustForm.value = {
    user_id: account.user_id,
    asset: account.asset,
    current_balance: account.balance,
    type: 'add',
    amount: '',
    remark: ''
  }
  showAdjustBalance.value = true
}

// 确认调整余额
const confirmAdjustBalance = async () => {
  try {
    await axios.post('/api/admin/accounts/adjust', adjustForm.value)
    ElMessage.success('余额调整成功')
    showAdjustBalance.value = false
    fetchAccounts()
  } catch (error) {
    ElMessage.error('余额调整失败')
  }
}

// 标签页切换
const handleTabChange = (tabName) => {
  switch (tabName) {
    case 'users':
      fetchUsers()
      break
    case 'orders':
      fetchOrders()
      break
    case 'accounts':
      fetchAccounts()
      break
  }
}

// 订单状态相关方法
const getOrderStatusType = (status) => {
  const statusMap = {
    'open': 'warning',
    'filled': 'success',
    'cancelled': 'danger',
    'partially_filled': 'primary'
  }
  return statusMap[status] || 'info'
}

const getOrderStatusText = (status) => {
  const statusMap = {
    'open': '开放',
    'filled': '已成交',
    'cancelled': '已取消',
    'partially_filled': '部分成交'
  }
  return statusMap[status] || status
}

// 打开添加资金对话框
const openAddFunds = () => {
  addFundsForm.value = {
    user_id: '',
    asset: '',
    amount: '',
    remark: ''
  }
  showAddFunds.value = true
}

// 为特定账户添加资金
const addFundsToAccount = (account) => {
  addFundsForm.value = {
    user_id: account.user_id.toString(),
    asset: account.asset,
    amount: '',
    remark: ''
  }
  showAddFunds.value = true
}

// 确认添加资金
const confirmAddFunds = async () => {
  try {
    // 表单验证
    await addFundsFormRef.value.validate()
    
    await axios.post('/api/admin/accounts/add-funds', addFundsForm.value)
    ElMessage.success('资金添加成功')
    showAddFunds.value = false
    fetchAccounts()
  } catch (error) {
    if (error.response) {
      ElMessage.error(error.response.data.error || '资金添加失败')
    } else {
      console.error('表单验证失败:', error)
    }
  }
}

// 组件挂载时加载用户数据
onMounted(() => {
  fetchUsers()
})
</script> 
