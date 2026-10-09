-- 2026-10-09 修复 SQL 报错：补齐后端 models 中 gorm.Model 所需的列
-- 背景：
--   backend-v2/internal/data 下的 Trade / Account / AccountFlow / Kline 均内嵌 gorm.Model，
--   GORM 会自动在查询里追加 `deleted_at IS NULL` 并在写操作里使用 created_at/updated_at/deleted_at。
--   但线上 test 库是早年 AutoMigrate 生成的旧表结构，缺少这些列，导致：
--     Error 1054: Unknown column 'trades.deleted_at'    (GET /api/my_trades)
--     Error 1054: Unknown column 'accounts.deleted_at'  (GET /api/accounts)
--   以及 orders 表缺少 user_id 索引导致的慢查询（SLOW SQL >= 1s）。
-- 本脚本幂等：重复执行时已存在的列/索引会被忽略（执行端对 1060/1061 错误做忽略处理）。
-- 目标库：78.154.103.23:9257/test

-- 1) trades：补 updated_at / deleted_at
ALTER TABLE trades
    ADD COLUMN updated_at datetime(3) NULL DEFAULT NULL AFTER created_at,
    ADD COLUMN deleted_at datetime(3) NULL DEFAULT NULL AFTER updated_at;

-- 2) accounts：补 deleted_at
ALTER TABLE accounts
    ADD COLUMN deleted_at datetime(3) NULL DEFAULT NULL AFTER updated_at;

-- 3) account_flows：补 updated_at / deleted_at（充值/提现/调账写流水时会用到）
ALTER TABLE account_flows
    ADD COLUMN updated_at datetime(3) NULL DEFAULT NULL AFTER created_at,
    ADD COLUMN deleted_at datetime(3) NULL DEFAULT NULL AFTER updated_at;

-- 4) klines：补 created_at / updated_at / deleted_at
ALTER TABLE klines
    ADD COLUMN created_at datetime(3) NULL DEFAULT NULL AFTER close_time,
    ADD COLUMN updated_at datetime(3) NULL DEFAULT NULL AFTER created_at,
    ADD COLUMN deleted_at datetime(3) NULL DEFAULT NULL AFTER updated_at;

-- 5) 索引：消除 orders 慢查询，并补充常用的用户维度索引
-- 注意：线上 orders.status 为 longtext，无法直接建 (user_id, status) 复合索引，
--       故复合索引改为 (user_id, created_at)，贴合“按用户按时间排序”的查询。
ALTER TABLE orders
    ADD INDEX idx_orders_user_id (user_id),
    ADD INDEX idx_orders_user_created (user_id, created_at);

ALTER TABLE trades
    ADD INDEX idx_trades_buy_user_id (buy_user_id),
    ADD INDEX idx_trades_sell_user_id (sell_user_id);

ALTER TABLE accounts
    ADD INDEX idx_accounts_user_id (user_id);

ALTER TABLE account_flows
    ADD INDEX idx_account_flows_user_id (user_id);
