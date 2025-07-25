<template>
  <el-table :data="assets" style="width: 100%">
    <el-table-column prop="asset" label="币种" />
    <el-table-column prop="balance" label="可用余额" />
    <el-table-column prop="frozen" label="冻结余额" />
    <el-table-column label="操作">
      <template #default="{ row }">
        <el-button size="small" @click="openDeposit(row)">充值</el-button>
        <el-button size="small" @click="openWithdraw(row)">提现</el-button>
      </template>
    </el-table-column>
  </el-table>
  <el-dialog v-model="showDeposit" title="充值">
    <el-form :model="depositForm">
      <el-form-item label="币种">
        <el-input v-model="depositForm.asset" disabled />
      </el-form-item>
      <el-form-item label="金额">
        <el-input v-model="depositForm.amount" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="doDeposit">确认充值</el-button>
      </el-form-item>
    </el-form>
  </el-dialog>
  <el-dialog v-model="showWithdraw" title="提现">
    <el-form :model="withdrawForm">
      <el-form-item label="币种">
        <el-input v-model="withdrawForm.asset" disabled />
      </el-form-item>
      <el-form-item label="金额">
        <el-input v-model="withdrawForm.amount" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="doWithdraw">确认提现</el-button>
      </el-form-item>
    </el-form>
  </el-dialog>
</template>
<script setup>
import { ref } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const assets = ref([])
const showDeposit = ref(false)
const showWithdraw = ref(false)
const depositForm = ref({ asset: '', amount: '' })
const withdrawForm = ref({ asset: '', amount: '' })

const fetchAssets = async () => {
  const { data } = await axios.get('/api/accounts')
  assets.value = data
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
    ElMessage.error('请输入有效的充值金额');
    return;
  }
  try {
    const response = await axios.post('/api/deposit', {
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
    ElMessage.error('请输入有效的提现金额');
    return;
  }
  try {
    const response = await axios.post('/api/withdraw', {
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
