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
        <el-table-column prop="asset" label="币种" />
        <el-table-column prop="balance" label="可用余额" />
        <el-table-column prop="frozen" label="冻结余额" />
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="openDeposit(row)">充值</el-button>
            <el-button size="small" type="danger" @click="openWithdraw(row)">提现</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showDeposit" title="充值" width="420px" class="glass-dialog">
      <el-form :model="depositForm" label-width="80px" label-position="top">
        <el-form-item label="币种">
          <el-input v-model="depositForm.asset" disabled />
        </el-form-item>
        <el-form-item label="金额">
          <el-input v-model="depositForm.amount" placeholder="请输入充值金额">
            <template #suffix>USDT</template>
          </el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDeposit = false">取消</el-button>
        <el-button type="primary" @click="doDeposit">确认充值</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showWithdraw" title="提现" width="420px" class="glass-dialog">
      <el-form :model="withdrawForm" label-width="80px" label-position="top">
        <el-form-item label="币种">
          <el-input v-model="withdrawForm.asset" disabled />
        </el-form-item>
        <el-form-item label="金额">
          <el-input v-model="withdrawForm.amount" placeholder="请输入提现金额">
            <template #suffix>USDT</template>
          </el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showWithdraw = false">取消</el-button>
        <el-button type="danger" @click="doWithdraw">确认提现</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { deposit, getAccounts, withdraw } from '../api/api'

const assets = ref([])
const errored = ref(false)
const showDeposit = ref(false)
const showWithdraw = ref(false)
const depositForm = ref({ asset: '', amount: '' })
const withdrawForm = ref({ asset: '', amount: '' })

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
  showWithdraw.value = true
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
  try {
    const response = await withdraw({
      asset: withdrawForm.value.asset,
      amount: withdrawForm.value.amount.toString()
    })
    ElMessage.success(response.data.message || '提现成功')
    showWithdraw.value = false
    fetchAssets()
  } catch (error) {
    const message = error.response?.data?.error || '提现失败'
    ElMessage.error(message)
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

.page-title-sm {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 600;
  color: var(--text-main);
}

.glass-dialog :global(.el-dialog__header) {
  background: transparent !important;
  border-bottom: 1px solid var(--border-color) !important;
}
.glass-dialog :global(.el-dialog__body) { padding: 20px !important; }
</style>
