# CEX交易所后端API文档

## 用户注册
- POST `/api/register`
- 参数: `{ username, password, email, phone }`
- 返回: `{ token }`

## 用户登录
- POST `/api/login`
- 参数: `{ username, password }`
- 返回: `{ token }`

## 资产账户列表
- GET `/api/accounts`
- Header: `Authorization: Bearer <token>`
- 返回: `[ { asset, balance, frozen, ... } ]`

## 充值
- POST `/api/deposit`
- Header: `Authorization: Bearer <token>`
- 参数: `{ asset, amount }`
- 返回: `{ message }`

## 提现
- POST `/api/withdraw`
- Header: `Authorization: Bearer <token>`
- 参数: `{ asset, amount }`
- 返回: `{ message }`

## 资金流水
- GET `/api/account_flows`
- Header: `Authorization: Bearer <token>`
- 返回: `[ { asset, change_type, amount, balance, ... } ]`

## 下单
- POST `/api/order`
- Header: `Authorization: Bearer <token>`
- 参数: `{ symbol, side, type, price, amount }`
- 返回: `Order对象`

## 撤单
- POST `/api/order/cancel/:id`
- Header: `Authorization: Bearer <token>`
- 返回: `{ message }`

## 订单簿
- GET `/api/orderbook?symbol=BTC/USDT`
- 返回: `{ bids: [...], asks: [...] }`

## 成交记录
- GET `/api/trades?symbol=BTC/USDT`
- 返回: `[ { price, amount, ... } ]`

## K线行情
- GET `/api/klines?symbol=BTC/USDT&interval=1m&limit=100`
- 返回: `[ { open, high, low, close, volume, open_time, close_time } ]`

## 后台管理
- GET `/api/admin/users` 用户列表
- GET `/api/admin/orders` 订单列表
- GET `/api/admin/accounts` 资金账户列表
- Header: `Authorization: Bearer <token>`

---

所有受保护接口需在Header中携带 `Authorization: Bearer <token>`。
返回数据均为JSON格式。 