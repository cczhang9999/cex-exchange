<template>
  <div class="assets-page page-container">
    <PageHeader title="资产管理" subtitle="账户余额、充值、提现与地址簿管理" icon="Wallet" tone="success" />

    <el-card class="glass-panel content-card">
      <template #header>
        <div class="flex-between">
          <h3 class="card-title">账户总览</h3>
          <span class="badge badge-primary">{{ assets.length }} 个币种</span>
        </div>
      </template>

      <el-table :data="assets" class="assets-table w-100" v-loading="!assets.length && !errored">
        <el-table-column prop="asset" label="币种" width="100" />
        <el-table-column prop="balance" label="可用余额" width="160">
          <template #default="{ row }">
            <span class="mono">{{ formatBalance(row.balance) }} {{ row.asset }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="frozen" label="冻结余额" width="160">
          <template #default="{ row }">
            <span class="mono text-muted">{{ formatBalance(row.frozen) }} {{ row.asset }}</span>
          </template>
        </el-table-column>
        <el-table-column label="总余额" width="160">
          <template #default="{ row }">
            <span class="mono font-bold">{{ formatBalance(getTotal(row)) }} {{ row.asset }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" @click="openDeposit(row)">
                <el-icon><Download /></el-icon> 充值
              </el-button>
              <el-button size="small" type="success" @click="openWithdraw(row)">
                <el-icon><Upload /></el-icon> 提现
              </el-button>
              <el-button size="small" type="warning" @click="openAddressBook(row)">
                <el-icon><Wallet /></el-icon> 地址
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 充值对话框 -->
    <el-dialog v-model="showDeposit" title="充值" width="480px" class="glass-dialog">
      <el-form :model="depositForm" label-width="80px" label-position="top">
        <el-form-item label="币种">
          <el-input v-model="depositForm.asset" disabled />
        </el-form-item>
        <el-form-item label="金额">
          <el-input v-model="depositForm.amount" placeholder="请输入充值金额">
            <template #suffix>{{ depositForm.asset }}</template>
          </el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDeposit = false">取消</el-button>
        <el-button type="primary" @click="doDeposit">确认充值</el-button>
      </template>
    </el-dialog>

    <!-- 提现对话框 -->
    <el-dialog v-model="showWithdraw" title="提现" width="480px" class="glass-dialog">
      <el-form :model="withdrawForm" label-width="80px" label-position="top">
        <el-form-item label="币种">
          <el-input v-model="withdrawForm.asset" disabled />
        </el-form-item>
        <el-form-item label="提现地址">
          <el-select 
            v-model="withdrawForm.address" 
            placeholder="请选择或输入提现地址"
            filterable
            allow-create
            default-first-option
            class="w-100"
          >
            <el-option
              v-for="addr in filteredAddresses"
              :key="addr.address"
              :label="addr.label || addr.address"
              :value="addr.address"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="金额">
          <el-input v-model="withdrawForm.amount" placeholder="请输入提现金额">
            <template #suffix>{{ withdrawForm.asset }}</template>
          </el-input>
        </el-form-item>
        <el-form-item label="手续费">
          <el-input :value="getWithdrawFee()" disabled>
            <template #suffix>{{ withdrawForm.asset }}</template>
          </el-input>
        </el-form-item>
        <el-form-item label="实际到账">
          <el-input :value="getActualAmount()" disabled>
            <template #suffix>{{ withdrawForm.asset }}</template>
          </el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showWithdraw = false">取消</el-button>
        <el-button type="success" @click="doWithdraw">确认提现</el-button>
      </template>
    </el-dialog>

    <!-- 充币地址对话框 -->
    <el-dialog
      v-model="showDepositAddress"
      :title="`充币 · ${depositAddressData.asset || ''}`"
      width="520px"
      class="glass-dialog"
    >
      <div v-if="depositAddressData.address" class="deposit-address-container">
        <div class="deposit-hero">
          <span class="deposit-avatar" :style="{ background: assetColor(depositAddressData.asset) }">
            {{ (depositAddressData.asset || '?').slice(0, 1) }}
          </span>
          <div class="deposit-hero-meta">
            <div class="deposit-asset-name">{{ depositAddressData.asset }}</div>
            <span class="badge badge-success">{{ depositAddressData.network }}</span>
          </div>
        </div>

        <div class="address-box">
          <div class="address-label">充币地址</div>
          <div class="address-value">
            <span class="mono">{{ depositAddressData.address }}</span>
            <el-button type="primary" size="small" @click="copyAddress">
              <el-icon><CopyDocument /></el-icon>
              复制
            </el-button>
          </div>
        </div>

        <div class="address-hint">
          <el-icon><Warning /></el-icon>
          <span>请确认充币网络为 <strong>{{ depositAddressData.network }}</strong>，否则资产将无法找回</span>
        </div>

        <div class="deposit-min-amount">
          <span class="text-muted">最小充值金额</span>
          <span class="mono font-semibold">{{ depositAddressData.minDeposit }} {{ depositAddressData.asset }}</span>
        </div>

        <el-alert type="info" :closable="false" show-icon title="充值说明">
          <ul class="deposit-note-list">
            <li>请勿向此地址充值任何非 {{ depositAddressData.asset }} 资产，否则无法恢复</li>
            <li>充值需网络确认后到账，预计 1–3 分钟</li>
            <li>请务必确认充值网络与提现网络一致</li>
          </ul>
        </el-alert>
      </div>

      <EmptyState
        v-else-if="depositAddressData.unavailable"
        icon="Warning"
        title="充币地址获取失败"
        description="后端暂未返回该币种的充币地址，请稍后重试或联系客服"
      />

      <div v-else class="deposit-loading">
        <el-icon class="is-loading"><Loading /></el-icon>
        正在获取充币地址...
      </div>

      <template #footer>
        <el-button text type="primary" @click="openManualDeposit({ asset: depositAddressData.asset })">
          手动入账（测试）
        </el-button>
        <el-button @click="showDepositAddress = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 地址簿对话框 -->
    <el-dialog v-model="showAddressBookDialog" title="提现地址管理" width="600px" class="glass-dialog">
      <div class="address-book-header">
        <el-button type="primary" @click="openAddAddressDialog">
          <el-icon><Plus /></el-icon> 添加地址
        </el-button>
      </div>
      
      <el-table :data="addressBook" class="address-book-table w-100 mt-16" v-loading="addressBookLoading">
        <el-table-column prop="label" label="备注名称" width="120" />
        <el-table-column prop="address" label="地址" min-width="200">
          <template #default="{ row }">
            <span class="mono text-small">{{ row.address }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="network" label="网络" width="100" />
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="danger" size="small" plain @click="deleteAddress(row)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 添加地址对话框 -->
      <el-dialog
        v-model="showAddAddressForm"
        title="添加提现地址"
        width="400px"
        append-to-body
        class="glass-dialog"
      >
        <el-form :model="newAddressForm" label-width="80px" label-position="top">
          <el-form-item label="币种">
            <el-select v-model="newAddressForm.asset" placeholder="请选择币种" class="w-100">
              <el-option label="USDT" value="USDT" />
              <el-option label="BTC" value="BTC" />
              <el-option label="ETH" value="ETH" />
              <el-option label="BNB" value="BNB" />
            </el-select>
          </el-form-item>
          <el-form-item label="网络">
            <el-select v-model="newAddressForm.network" placeholder="请选择网络" class="w-100">
              <el-option label="TRC-20 (TRON)" value="TRC-20" />
              <el-option label="ERC-20 (Ethereum)" value="ERC-20" />
              <el-option label="BEP-20 (BSC)" value="BEP-20" />
            </el-select>
          </el-form-item>
          <el-form-item label="地址">
            <el-input v-model="newAddressForm.address" placeholder="请输入钱包地址" />
          </el-form-item>
          <el-form-item label="备注名称">
            <el-input v-model="newAddressForm.label" placeholder="例如：我的钱包" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="showAddAddressForm = false">取消</el-button>
          <el-button type="primary" @click="addAddress">确认添加</el-button>
        </template>
      </el-dialog>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  deposit, getAccounts, withdraw, getDepositAddress
} from '../api/api'
import { 
  Download, Upload, Wallet, CopyDocument, 
  Warning, Loading, Plus, Delete 
} from '@element-plus/icons-vue'
import PageHeader from './ui/PageHeader.vue'
import EmptyState from './ui/EmptyState.vue'

// 币种配色
const ASSET_COLORS = {
  BTC: 'linear-gradient(135deg,#f7931a,#f2a900)',
  ETH: 'linear-gradient(135deg,#627eea,#8a92b2)',
  BNB: 'linear-gradient(135deg,#f3ba2f,#e8a90c)',
  USDT: 'linear-gradient(135deg,#26a17b,#1a7f5f)',
}
const assetColor = (asset) => ASSET_COLORS[asset] || 'linear-gradient(135deg,#3b82f6,#2563eb)'

const assets = ref([])
const errored = ref(false)
const showDeposit = ref(false)
const showWithdraw = ref(false)
const showDepositAddress = ref(false)
const showAddressBookDialog = ref(false)
const showAddAddressForm = ref(false)
const addressBookLoading = ref(false)

// 地址簿数据
const addressBook = ref([
  { id: 1, label: '我的钱包', address: 'TXqwertyuioplkjhgfdsazxcvbnm1234', network: 'TRC-20', asset: 'USDT' },
  { id: 2, label: '交易所', address: '0xabcd1234efgh5678ijkl9012mnop3456qrst', network: 'ERC-20', asset: 'ETH' },
])

const depositForm = ref({ asset: '', amount: '' })
const withdrawForm = ref({ asset: '', amount: '', address: '' })
const depositAddressData = ref({})
const newAddressForm = ref({ asset: '', network: '', address: '', label: '' })

// 根据选择的币种过滤地址簿
const filteredAddresses = computed(() => {
  if (!withdrawForm.value.asset) return addressBook.value
  return addressBook.value.filter(a => a.asset === withdrawForm.value.asset)
})

// 计算总余额
const getTotal = (row) => {
  return (parseFloat(row.balance) || 0) + (parseFloat(row.frozen) || 0)
}

// 格式化余额
const formatBalance = (balance) => {
  if (!balance || isNaN(balance)) return '0.00'
  return parseFloat(balance).toFixed(8)
}

// 获取提现手续费
const getWithdrawFee = () => {
  if (!withdrawForm.value.asset) return '0'
  const fees = { 'USDT': '1', 'BTC': '0.0005', 'ETH': '0.005', 'BNB': '0.02' }
  return fees[withdrawForm.value.asset] || '0'
}

// 计算实际到账金额
const getActualAmount = () => {
  const amount = parseFloat(withdrawForm.value.amount) || 0
  const fee = parseFloat(getWithdrawFee()) || 0
  return (amount - fee).toFixed(8)
}

const fetchAssets = async () => {
  errored.value = false
  try {
    const { data } = await getAccounts()
    assets.value = data.data
  } catch (error) {
    errored.value = true
    console.error('获取资产失败:', error)
    ElMessage.error('获取资产失败')
  }
}
fetchAssets()

const openDeposit = async (row) => {
  depositForm.value.asset = row.asset
  depositForm.value.amount = ''
  openDepositAddressDialog(row)
}

const openWithdraw = (row) => {
  withdrawForm.value.asset = row.asset
  withdrawForm.value.amount = ''
  withdrawForm.value.address = ''
  showWithdraw.value = true
}

const openAddressBook = async (row) => {
  withdrawForm.value.asset = row.asset
  showAddressBookDialog.value = true
}

const doDeposit = async () => {
  if (!depositForm.value.amount || isNaN(Number(depositForm.value.amount)) || Number(depositForm.value.amount) <= 0) {
    ElMessage.error('请输入有效的充值金额')
    return
  }
  try {
    const response = await deposit({
      asset: depositForm.value.asset,
      amount: depositForm.value.amount.toString()
    })
    ElMessage.success(response.data.message || '充值成功')
    showDeposit.value = false
    fetchAssets()
  } catch (error) {
    const message = error.response?.data?.error || '充值失败'
    ElMessage.error(message)
  }
}

const doWithdraw = async () => {
  if (!withdrawForm.value.amount || isNaN(Number(withdrawForm.value.amount)) || Number(withdrawForm.value.amount) <= 0) {
    ElMessage.error('请输入有效的提现金额')
    return
  }
  if (!withdrawForm.value.address) {
    ElMessage.error('请输入提现地址')
    return
  }
  try {
    const response = await withdraw({
      asset: withdrawForm.value.asset,
      amount: withdrawForm.value.amount.toString(),
      address: withdrawForm.value.address
    })
    ElMessage.success(response.data.message || '提现成功')
    showWithdraw.value = false
    fetchAssets()
  } catch (error) {
    const message = error.response?.data?.error || '提现失败'
    ElMessage.error(message)
  }
}

// 充币网络 / 最小充值额的默认值
const DEFAULT_NETWORK = { USDT: 'TRC-20', BTC: 'Bitcoin', ETH: 'ERC-20', BNB: 'BEP-20' }
const DEFAULT_MIN = { USDT: '10', BTC: '0.001', ETH: '0.01', BNB: '0.01' }

// 打开充币地址对话框（优先后端接口，失败降级）
const openDepositAddressDialog = async (row) => {
  depositAddressData.value = {}
  showDepositAddress.value = true

  try {
    const { data } = await getDepositAddress(row.asset)
    const d = data?.data || {}
    depositAddressData.value = {
      asset: row.asset,
      address: d.address || '',
      network: d.network || DEFAULT_NETWORK[row.asset] || 'TRC-20',
      minDeposit: d.min_deposit || d.minDeposit || DEFAULT_MIN[row.asset] || '10',
    }
  } catch (error) {
    depositAddressData.value = {
      asset: row.asset,
      address: '',
      network: DEFAULT_NETWORK[row.asset] || 'TRC-20',
      minDeposit: DEFAULT_MIN[row.asset] || '10',
      unavailable: true,
    }
  }
}

// 手动入账（测试用）：打开金额入账对话框
const openManualDeposit = (row) => {
  depositForm.value.asset = row.asset
  depositForm.value.amount = ''
  showDepositAddress.value = false
  showDeposit.value = true
}

// 复制地址
const copyAddress = async () => {
  try {
    await navigator.clipboard.writeText(depositAddressData.value.address)
    ElMessage.success('地址已复制')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

// 添加地址
const openAddAddressDialog = () => {
  newAddressForm.value = { asset: '', network: '', address: '', label: '' }
  showAddAddressForm.value = true
}

const addAddress = async () => {
  if (!newAddressForm.value.asset) {
    ElMessage.error('请选择币种')
    return
  }
  if (!newAddressForm.value.address) {
    ElMessage.error('请输入地址')
    return
  }
  if (!newAddressForm.value.network) {
    ElMessage.error('请选择网络')
    return
  }
  
  // 模拟添加地址（实际项目中应调用API）
  const newId = addressBook.value.length + 1
  addressBook.value.push({
    id: newId,
    ...newAddressForm.value
  })
  
  ElMessage.success('地址添加成功')
  showAddAddressForm.value = false
}

const deleteAddress = async (row) => {
  try {
    // 模拟删除地址（实际项目中应调用API）
    const index = addressBook.value.findIndex(a => a.id === row.id)
    if (index > -1) {
      addressBook.value.splice(index, 1)
    }
    ElMessage.success('地址已删除')
  } catch (error) {
    ElMessage.error('删除失败')
  }
}
</script>

<style scoped>
.assets-page {
  max-width: var(--content-max);
  margin: 0 auto;
}

.content-card {
  background: var(--surface-1);
  border-radius: var(--radius-card);
}

.assets-table :global(.el-table__row) {
  border-radius: 0 !important;
}

.action-buttons {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 6px;
  flex-wrap: nowrap;
}

.glass-dialog :global(.el-dialog__header) {
  background: transparent !important;
  border-bottom: 1px solid var(--border-color) !important;
  color: var(--text-main) !important;
}
.glass-dialog :global(.el-dialog__body) { padding: 20px !important; }

/* 充币地址样式 */
.deposit-address-container {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.deposit-hero {
  display: flex;
  align-items: center;
  gap: 12px;
}
.deposit-avatar {
  width: 44px;
  height: 44px;
  border-radius: 13px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 800;
  font-size: 1.1rem;
  box-shadow: var(--shadow-sm);
}
.deposit-hero-meta { display: flex; flex-direction: column; gap: 4px; align-items: flex-start; }
.deposit-asset-name {
  font-size: 1.05rem;
  font-weight: 750;
  color: var(--text-main);
  line-height: 1.1;
}

.address-box {
  padding: 14px 16px;
  border-radius: var(--radius-base);
  background: var(--surface-sunken);
  border: 1px solid var(--border-subtle);
}
.address-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-secondary);
  margin-bottom: 8px;
}
.address-value {
  display: flex;
  align-items: center;
  gap: 10px;
}
.address-value .mono {
  flex: 1;
  font-size: 0.82rem;
  color: var(--text-main);
  word-break: break-all;
  line-height: 1.5;
}

.address-hint {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 0.78rem;
  color: var(--warning-emphasis);
  background: var(--warning-glow);
  border: 1px solid color-mix(in srgb, var(--warning), transparent 70%);
  padding: 10px 12px;
  border-radius: var(--radius-base);
  line-height: 1.5;
}
.address-hint strong { color: var(--warning-emphasis); }

.deposit-min-amount {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.82rem;
  padding: 0 2px;
}

.deposit-note-list {
  margin: 6px 0 0;
  padding-left: 16px;
}
.deposit-note-list li {
  margin-bottom: 3px;
  color: var(--text-muted);
  font-size: 0.78rem;
  line-height: 1.6;
}

.deposit-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 44px;
  color: var(--text-muted);
}

/* 地址簿样式 */
.address-book-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}

.address-book-table {
  border-radius: var(--radius-base);
}

/* 共享样式 */
.w-100 { width: 100%; }
.mt-16 { margin-top: 16px; }
.mono { font-family: 'Roboto Mono', 'Fira Code', monospace; }
.text-muted { color: var(--text-muted); }
.font-bold { font-weight: 700; }
.font-semibold { font-weight: 600; }
.text-small { font-size: 12px; }
</style>