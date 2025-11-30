// src/api/api.js
import request from '../utils/request'

// 登录
export const login = (form) => request.post('/api/login', form)
// 注册
export const register = (form) => request.post('/api/register', form)
// 下单
export const placeOrder = (order) => request.post('/api/order', order)
// 撤单
export const cancelOrder = (id) => request.post(`/api/order/cancel/${id}`)
// 获取订单簿
export const getOrderbook = (symbol) => request.get('/api/orderbook', { params: { symbol } })
// 获取成交记录
export const getTrades = (symbol) => request.get('/api/trades', { params: { symbol } })
// 获取我的订单
export const getMyOrders = () => request.get('/api/my_orders')
// 获取我的成交
export const getMyTrades = () => request.get('/api/my_trades') 
// 获取用户列表
export const getUsers = (params) => request.get('/api/admin/users', params)
// 获取用户资产
export const getAccounts = () => request.get('/api/accounts')

// 充值
export const deposit = (form) => request.post('/api/deposit', form)

// 提现
export const withdraw = (form) => request.post('/api/withdraw', form)

// 添加资金
export const addFunds = (form) => request.post('/api/admin/accounts/add-funds', form)

export const blockUser = (id) => request.post(`/api/admin/users/${id}/block`)   

export const getOrders = (params) => request.get('/api/admin/orders', params)

export const adjustBalance = (form) => request.post('/api/admin/accounts/adjust', form)

export const cancelAadminOrder = (id) => request.post(`/api/admin/orders/${id}/cancel`)

export const getAdminAccounts = (params) => request.get('/api/admin/accounts', params)
