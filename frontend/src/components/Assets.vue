<template>
  <div class="assets-page">
    <el-card class="glass-panel content-card">
      <template #header>
        <div class="flex-between">
          <h3 class="page-title-sm">资产管理</h3>
          <el-tag type="info" effect="dark" class="glass-tag">Accounts</el-tag>
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
    <el-dialog v-model="showDepositAddress" title="充币地址" width="500px" class="glass-dialog">
      <div v-if="depositAddressData.address" class="deposit-address-container">
        <div class="deposit-asset-info">
          <span class="deposit-asset-name">{{ depositAddressData.asset }}</span>
          <el-tag type="success" size="small">主网</el-tag>
        </div>
        
        <div class="deposit-qr-container">
          <div class="qr-code-wrapper">
            <!-- 简化的二维码展示 -->
            <div class="qr-placeholder">
              <svg viewBox="0 0 100 100" class="qr-svg">
                <rect x="0" y="0" width="30" height="30" fill="#000"/>
                <rect x="35" y="0" width="10" height="10" fill="#000"/>
                <rect x="50" y="0" width="15" height="15" fill="#000"/>
                <rect x="70" y="0" width="30" height="30" fill="#000"/>
                <rect x="0" y="35" width="10" height="10" fill="#000"/>
                <rect x="15" y="40" width="5" height="5" fill="#000"/>
                <rect x="25" y="35" width="15" height="15" fill="#000"/>
                <rect x="0" y="70" width="30" height="30" fill="#000"/>
                <rect x="35" y="70" width="10" height="10" fill="#000"/>
                <rect x="50" y="65" width="15" height="15" fill="#000"/>
                <rect x="70" y="70" width="10" height="10" fill="#000"/>
                <rect x="85" y="70" width="15" height="15" fill="#000"/>
                <rect x="70" y="85" width="15" height="15" fill="#000"/>
                <rect x="85" y="85" width="10" height="10" fill="#000"/>
              </svg>
            </div>
          </div>
          <div class="address-info">
            <div class="address-label">充币地址</div>
            <div class="address-value">
              <span class="mono">{{ depositAddressData.address }}</span>
              <el-button 
                type="primary" 
                size="small" 
                @click="copyAddress"
                class="copy-btn"
              >
                <el-icon><CopyDocument /></el-icon>
              </el-button>
            </div>
            <div class="address-hint">
              <el-icon><Warning /></el-icon>
              请确认充币网络为 <strong>{{ depositAddressData.network || 'TRC-20' }}</strong>，以免造成资产丢失
            </div>
          </div>
        </div>
        
        <div class="deposit-min-amount">
          <span class="text-muted">最小充值金额：</span>
          <span class="mono">{{ depositAddressData.minDeposit || '10' }} {{ depositAddressData.asset }}</span>
        </div>
        
        <div class="deposit-note">
          <el-alert type="info" :closable="false" show-icon>
            <template #title>
              充值说明：
              <ul class="deposit-note-list">
                <li>请勿向上述地址充值任何非 {{ depositAddressData.asset }} 资产，否则资产将无法恢复</li>
                <li>充值后需要网络确认才能到账，预计确认时间为 1-3 分钟</li>
                <li>请务必确认充值网络与您提现网络一致</li>
              </ul>
            </template>
          </el-alert>
        </div>
      </div>
      <div v-else class="deposit-loading">
        <el-icon class="is-loading"><Loading /></el-icon>
        正在获取充币地址...
      </div>
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
  deposit, getAccounts, withdraw
} from '../api/api'
import { 
  Download, Upload, Wallet, CopyDocument, 
  Warning, Loading, Plus, Delete 
} from '@element-plus/icons-vue'

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

const openDeposit = (row) => {
  depositForm.value.asset = row.asset
  depositForm.value.amount = ''
  showDeposit.value = true
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

// 打开充币地址对话框
const openDepositAddressDialog = async (row) => {
  depositAddressData.value = {}
  showDepositAddress.value = true
  
  // 模拟获取充币地址（实际项目中应调用API）
  // 后端需要实现: GET /api/deposit/address?asset=xxx
  setTimeout(() => {
    depositAddressData.value = {
      asset: row.asset,
      address: row.asset === 'USDT' 
        ? 'TXqwerty' + Math.random().toString(36).substring(2, 15) + 'abcdef'
        : row.asset === 'BTC'
        ? '1' + Math.random().toString(36).substring(2, 15) + 'ghijkl'
        : '0x' + Math.random().toString(36).substring(2, 18),
      network: row.asset === 'USDT' ? 'TRC-20' : row.asset === 'ETH' ? 'ERC-20' : 'BEP-20',
      minDeposit: row.asset === 'USDT' ? '10' : row.asset === 'BTC' ? '0.001' : '0.01'
    }
  }, 500)
}

// 修改 openDeposit 函数以打开充币地址对话框
const originalOpenDeposit = openDeposit
const doOpenDeposit = (row) => {
  openDeposit(row)
  // 在充值对话框中添加查看充币地址的选项
  showDeposit.value = false
  openDepositAddressDialog(row)
}

// 覆盖 openDeposit 为新版本（同时打开地址和充值）
const openDepositNew = (row) => {
  depositForm.value.asset = row.asset
  depositForm.value.amount = ''
  openDepositAddressDialog(row)
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
  max-width: 1200px;
  margin: 0 auto;
}

.content-card {
  background: var(--bg-card);
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

.page-title-sm {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 600;
  color: var(--text-main);
}

.glass-dialog :global(.el-dialog__header) {
  background: transparent !important;
  border-bottom: 1px solid var(--border-color) !important;
  color: var(--text-main) !important;
}
.glass-dialog :global(.el-dialog__body) { padding: 20px !important; }

/* 充币地址样式 */
.deposit-address-container {
  padding: 8px 0;
}

.deposit-asset-info {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
}

.deposit-asset-name {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-main);
}

.deposit-qr-container {
  display: flex;
  gap: 24px;
  align-items: flex-start;
  margin-bottom: 20px;
  padding: 20px;
  background: color-mix(in srgb, var(--bg-elevated), transparent 85%);
  border-radius: var(--radius-base);
}

.qr-code-wrapper {
  flex-shrink: 0;
}

.qr-placeholder {
  width: 140px;
  height: 140px;
  background: #fff;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 10px;
}

.qr-svg {
  width: 100%;
  height: 100%;
}

.address-info {
  flex: 1;
}

.address-label {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 8px;
}

.address-value {
  display: flex;
  align-items: center;
  gap: 8px;
  word-break: break-all;
  margin-bottom: 12px;
}

.address-value .mono {
  font-size: 12px;
  color: var(--text-main);
}

.copy-btn {
  flex-shrink: 0;
}

.address-hint {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 12px;
  color: var(--warning);
  background: color-mix(in srgb, var(--warning), transparent 90%);
  padding: 8px 12px;
  border-radius: 6px;
}

.address-hint strong {
  color: var(--warning);
}

.deposit-min-amount {
  margin-bottom: 16px;
  font-size: 13px;
}

.deposit-note-list {
  margin: 8px 0 0 16px;
  padding-left: 0;
}

.deposit-note-list li {
  margin-bottom: 4px;
  color: var(--text-muted);
}

.deposit-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 40px;
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
.text-small { font-size: 12px; }
</style>