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

// ========== Phase 1 新增 API ==========

// 获取充币地址
export const getDepositAddress = (asset) => request.get('/api/deposit/address', { params: { asset } })

// 获取K线数据（支持多周期）
export const getKline = (params) => request.get('/api/kline', { params: {
  symbol: params.symbol,
  period: params.period || '1m',
  limit: params.limit || 100
}})

// 地址簿管理
export const getAddressBook = () => request.get('/api/address-book')
export const addAddressBook = (data) => request.post('/api/address-book', data)
export const deleteAddressBook = (id) => request.delete(`/api/address-book/${id}`)

// 止损止盈订单列表（可选，用于展示触发条件）
export const getStopOrders = () => request.get('/api/my_orders', { params: { type: 'stop' } })
