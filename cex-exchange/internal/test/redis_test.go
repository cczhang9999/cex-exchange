package test

import (
	"cex-exchange/internal/service"
	"context"
	"testing"

	_ "cex-exchange/internal/boot" // 导入 boot 包以加载配置

	"github.com/gogf/gf/v2/frame/g"
)

func TestRedisBasic(t *testing.T) {
	ctx := context.Background()

	// 1. 打印配置信息
	config, _ := g.Cfg().Get(ctx, "redis")
	t.Logf("Redis Config: %v", config)

	// 2. 测试 GoFrame Redis PING
	pong, err := service.Redis.Do(ctx, "PING")
	if err != nil {
		t.Fatalf("GoFrame Redis PING 失败: %v", err)
	}
	t.Logf("✅ GoFrame PING 成功: %v", pong)

	// 3. 测试设置值
	err = service.Redis.Set(ctx, "test_key", "test_value")
	if err != nil {
		t.Fatalf("Redis Set 失败: %v", err)
	}
	t.Log("✅ Set 成功")

	// 2. 测试获取值
	value, err := service.Redis.Get(ctx, "test_key")
	if err != nil {
		t.Fatalf("Redis Get 失败: %v", err)
	}
	if value != "test_value" {
		t.Fatalf("期望值 'test_value', 实际得到 '%s'", value)
	}
	t.Logf("✅ Get 成功, 值: %s", value)

	// 3. 测试设置带过期时间的值（10秒）
	userData := map[string]interface{}{
		"user_id": 123,
		"name":    "测试用户",
	}
	err = service.Redis.SetEX(ctx, "session:123", userData, 10)
	if err != nil {
		t.Fatalf("Redis SetEX 失败: %v", err)
	}
	t.Log("✅ SetEX 成功")

	// 4. 验证 key 存在
	exists, err := service.Redis.Exists(ctx, "session:123")
	if err != nil {
		t.Fatalf("Redis Exists 失败: %v", err)
	}
	if !exists {
		t.Fatal("期望 key 存在，但实际不存在")
	}
	t.Log("✅ Exists 检查成功")

	// 5. 清理测试数据
	err = service.Redis.Del(ctx, "test_key", "session:123")
	if err != nil {
		t.Logf("⚠️  清理失败: %v", err)
	} else {
		t.Log("✅ 清理成功")
	}
}

func TestRedisHash(t *testing.T) {
	ctx := context.Background()
	key := "test_user:1001"

	// 1. 设置 Hash 字段
	err := service.Redis.HSet(ctx, key, "name", "张三")
	if err != nil {
		t.Fatalf("HSet 失败: %v", err)
	}

	err = service.Redis.HSet(ctx, key, "age", "25")
	if err != nil {
		t.Fatalf("HSet 失败: %v", err)
	}
	t.Log("✅ HSet 成功")

	// 2. 获取单个字段
	name, err := service.Redis.HGet(ctx, key, "name")
	if err != nil {
		t.Fatalf("HGet 失败: %v", err)
	}
	if name != "张三" {
		t.Fatalf("期望 '张三', 实际得到 '%s'", name)
	}
	t.Logf("✅ HGet 成功, 姓名: %s", name)

	// 3. 获取所有字段
	fields, err := service.Redis.HGetAll(ctx, key)
	if err != nil {
		t.Fatalf("HGetAll 失败: %v", err)
	}
	t.Logf("✅ HGetAll 成功, 数据: %+v", fields)

	// 4. 清理
	service.Redis.Del(ctx, key)
	t.Log("✅ 清理成功")
}

func TestRedisCounter(t *testing.T) {
	ctx := context.Background()
	key := "test_counter"

	// 清理旧数据
	service.Redis.Del(ctx, key)

	// 1. 测试自增
	count, err := service.Redis.Incr(ctx, key)
	if err != nil {
		t.Fatalf("Incr 失败: %v", err)
	}
	if count != 1 {
		t.Fatalf("期望 1, 实际得到 %d", count)
	}
	t.Logf("✅ Incr 成功, 计数: %d", count)

	// 2. 测试增加指定值
	count, err = service.Redis.IncrBy(ctx, key, 5)
	if err != nil {
		t.Fatalf("IncrBy 失败: %v", err)
	}
	if count != 6 {
		t.Fatalf("期望 6, 实际得到 %d", count)
	}
	t.Logf("✅ IncrBy 成功, 计数: %d", count)

	// 3. 清理
	service.Redis.Del(ctx, key)
	t.Log("✅ 清理成功")
}