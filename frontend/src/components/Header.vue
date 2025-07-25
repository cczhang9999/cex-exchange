<template>
  <el-header style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; display: flex; align-items: center; justify-content: space-between; height: 60px; flex-shrink: 0;">
    <div style="display: flex; align-items: center;">
      <h2 style="margin: 0; color: white;">CEX交易所</h2>
    </div>
    <div style="display: flex; align-items: center; gap: 20px;">
      <el-button link style="color: white;" @click="$emit('nav', 'assets')">资产管理</el-button>
      <el-button link style="color: white;" @click="$emit('nav', 'trade')">交易中心</el-button>
      <el-button link style="color: white;" @click="$emit('nav', 'orders')">订单管理</el-button>
      <el-button link style="color: white;" @click="$emit('nav', 'admin')">后台管理</el-button>
    </div>
  </el-header>
</template> 