# CEX交易所数据库表结构说明

## 1. users（用户表）
- **id**：用户唯一ID，自增主键。
- **username**：用户名，唯一。
- **password_hash**：加密存储的密码哈希。
- **email/phone**：邮箱/手机号，唯一。
- **status**：用户状态（1-正常，0-禁用）。
- **created_at/updated_at**：创建和更新时间。

## 2. accounts（资金账户表）
- **id**：账户唯一ID，自增主键。
- **user_id**：所属用户ID。
- **asset**：币种（如BTC、USDT等）。
- **balance**：可用余额。
- **frozen**：冻结余额（如挂单时冻结）。
- **created_at/updated_at**：创建和更新时间。

## 3. account_flows（资金流水表）
- **id**：流水唯一ID，自增主键。
- **user_id**：用户ID。
- **account_id**：账户ID。
- **asset**：币种。
- **change_type**：变动类型（充值、提现、交易等）。
- **amount**：变动金额。
- **balance**：变动后余额。
- **ref_id**：关联业务ID（如订单ID、提现ID等）。
- **remark**：备注。
- **created_at**：创建时间。

## 4. orders（订单表）
- **id**：订单唯一ID，自增主键。
- **user_id**：下单用户ID。
- **symbol**：交易对（如BTC/USDT）。
- **side**：买卖方向（buy/sell）。
- **type**：订单类型（限价/市价）。
- **price**：委托价格（市价单可为空）。
- **amount**：委托数量。
- **filled**：已成交数量。
- **status**：订单状态（open/partially_filled/filled/cancelled）。
- **created_at/updated_at**：创建和更新时间。

## 5. trades（成交表）
- **id**：成交唯一ID，自增主键。
- **buy_order_id/sell_order_id**：买单/卖单ID。
- **symbol**：交易对。
- **price**：成交价格。
- **amount**：成交数量。
- **buy_user_id/sell_user_id**：买方/卖方用户ID。
- **created_at**：成交时间。

## 6. klines（K线表）
- **id**：K线唯一ID，自增主键。
- **symbol**：交易对。
- **interval**：K线周期（如1m,5m,1h等）。
- **open/high/low/close**：开盘/最高/最低/收盘价。
- **volume**：成交量。
- **open_time/close_time**：K线开始/结束时间。
- **唯一索引**：symbol+interval+open_time唯一。

---

如需扩展更多业务表，可在此基础上增加。 