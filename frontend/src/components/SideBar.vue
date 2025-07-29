<template>
  <el-aside width="200px" style="background: #f5f7fa; height: 100%; overflow-y: auto;">
    <el-menu
      :default-active="activeMenu"
      @select="$emit('menu', $event)"
      style="border-right: none; height: 100%;"
    >
      <el-menu-item index="home">
        <el-icon><House /></el-icon>
        <span>首页概览</span>
      </el-menu-item>
      <el-menu-item index="login">
        <el-icon><User /></el-icon>
        <span>用户登录</span>
      </el-menu-item>
      <el-menu-item index="register">
        <el-icon><UserFilled /></el-icon>
        <span>用户注册</span>
      </el-menu-item>
      <el-divider />
      <el-menu-item index="assets">
        <el-icon><Wallet /></el-icon>
        <span>资产管理</span>
      </el-menu-item>
      <el-menu-item index="trade">
        <el-icon><TrendCharts /></el-icon>
        <span>交易中心</span>
      </el-menu-item>
      <el-menu-item index="realtime">
        <el-icon><TrendCharts /></el-icon>
        <span>实时行情</span>
      </el-menu-item>
      <el-menu-item index="orders">
        <el-icon><Document /></el-icon>
        <span>订单管理</span>
      </el-menu-item>
      <el-divider />
      <el-menu-item index="admin">
        <el-icon><Setting /></el-icon>
        <span>后台管理</span>
      </el-menu-item>
    </el-menu>
  </el-aside>
</template>
<script setup>
import { House, User, UserFilled, Wallet, TrendCharts, Document, Setting } from '@element-plus/icons-vue'
const props = defineProps({ activeMenu: String })
</script> 