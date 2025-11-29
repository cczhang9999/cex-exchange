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