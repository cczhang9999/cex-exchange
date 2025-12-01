<template>
  <el-header style="background: linear-gradient(135deg, #e0e7ff 0%, #f3e8ff 100%); color: #1e293b; display: flex; align-items: center; justify-content: space-between; height: 60px; flex-shrink: 0; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
    <div style="display: flex; align-items: center;">
      <h2 style="margin: 0; color: #1e293b; font-weight: 700;">CEX交易所</h2>
    </div>
    <div style="display: flex; align-items: center; gap: 20px;">
      <el-button link style="color: #1e293b; font-weight: 600;" @click="$emit('nav', 'assets')">资产管理</el-button>
      <el-button link style="color: #1e293b; font-weight: 600;" @click="$emit('nav', 'trade')">交易中心</el-button>
      <el-button link style="color: #1e293b; font-weight: 600;" @click="$emit('nav', 'orders')">订单管理</el-button>
      <el-button link style="color: #1e293b; font-weight: 600;" @click="$emit('nav', 'admin')">后台管理</el-button>
    </div>
  </el-header>
</template> 