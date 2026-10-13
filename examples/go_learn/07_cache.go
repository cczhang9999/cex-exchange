package main

import (
	"context"
	"fmt"
	"time"
)

// 示例 7: Redis 缓存模式
// 在高并发系统中，缓存是提升性能的关键

// 模拟 Redis 客户端
type RedisMock struct {
	data map[string]string
}

func NewRedisMock() *RedisMock {
	return &RedisMock{data: make(map[string]string)}
}

func (r *RedisMock) Get(key string) (string, error) {
	return r.data[key], nil
}

func (r *RedisMock) Set(key, value string) error {
	r.data[key] = value
	return nil
}

func (r *RedisMock) SetEX(key, value string, expiration int) error {
	r.data[key] = value
	return nil
}

func (r *RedisMock) Del(key string) error {
	delete(r.data, key)
	return nil
}

// UserCache 用户缓存服务
type UserCache struct {
	redis *RedisMock
}

func NewUserCache(redis *RedisMock) *UserCache {
	return &UserCache{redis: redis}
}

// GetUser 获取用户（先查缓存，再查数据库）
func (c *UserCache) GetUser(ctx context.Context, userID uint64) (string, error) {
	cacheKey := fmt.Sprintf("user:%d", userID)

	// 1. 先查缓存
	userJSON, err := c.redis.Get(cacheKey)
	if err == nil && userJSON != "" {
		fmt.Printf("[缓存命中] 用户ID: %d\n", userID)
		return userJSON, nil
	}

	fmt.Printf("[缓存未命中] 从数据库查询用户ID: %d\n", userID)
	// 2. 模拟从数据库查询
	userData := fmt.Sprintf(`{"id": %d, "username": "user%d", "balance": 10000}`, userID, userID)

	// 3. 写入缓存（5分钟过期）
	c.redis.SetEX(cacheKey, userData, 300)

	return userData, nil
}

// InvalidateUserCache 清除用户缓存
func (c *UserCache) InvalidateUserCache(userID uint64) {
	cacheKey := fmt.Sprintf("user:%d", userID)
	c.redis.Del(cacheKey)
	fmt.Printf("[缓存清除] 用户ID: %d\n", userID)
}

// OrderCache 订单缓存服务
type OrderCache struct {
	redis *RedisMock
}

func NewOrderCache(redis *RedisMock) *OrderCache {
	return &OrderCache{redis: redis}
}

// GetOrders 获取订单列表（带缓存）
func (c *OrderCache) GetOrders(ctx context.Context, userID uint64, symbol string) ([]string, error) {
	cacheKey := fmt.Sprintf("orders:%d:%s", userID, symbol)

	// 缓存有效期 20 秒
	ordersJSON, err := c.redis.Get(cacheKey)
	if err == nil && ordersJSON != "" {
		fmt.Printf("[订单缓存命中] user=%d, symbol=%s\n", userID, symbol)
		// 实际需要解析 JSON，这里简化
		return []string{ordersJSON}, nil
	}

	fmt.Printf("[订单缓存未命中] user=%d, symbol=%s\n", userID, symbol)
	// 模拟数据库查询
	orders := []string{
		`{"id": 1, "symbol": "BTC/USDT", "side": "buy", "price": 50000}`,
		`{"id": 2, "symbol": "BTC/USDT", "side": "sell", "price": 51000}`,
	}

	c.redis.SetEX(cacheKey, fmt.Sprintf("%v", orders), 20)

	return orders, nil
}

func main() {
	fmt.Println("=== Go 学习示例 7: Redis 缓存模式 ===")

	redis := NewRedisMock()
	userCache := NewUserCache(redis)

	ctx := context.Background()

	// 第一次查询 - 缓存未命中
	fmt.Println("\n--- 第一次查询用户 ---")
	userCache.GetUser(ctx, 12345)

	// 第二次查询 - 缓存命中
	fmt.Println("\n--- 第二次查询用户 ---")
	userCache.GetUser(ctx, 12345)

	// 清除缓存后再次查询
	fmt.Println("\n--- 清除缓存后查询 ---")
	userCache.InvalidateUserCache(12345)
	userCache.GetUser(ctx, 12345)

	// 订单缓存示例
	fmt.Println("\n--- 订单缓存示例 ---")
	orderCache := NewOrderCache(redis)

	// 第一次查询
	orderCache.GetOrders(ctx, 1001, "BTC/USDT")
	time.Sleep(100 * time.Millisecond)

	// 20秒内再次查询（缓存命中）
	orderCache.GetOrders(ctx, 1001, "BTC/USDT")
}
