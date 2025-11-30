package admin

import (
	"cex-exchange/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// TestRedis 测试 Redis 连接和基本操作
func (c *ControllerV1) TestRedis(r *ghttp.Request) {
	ctx := r.Context()

	// 1. 测试 Set
	err := service.Redis.Set(ctx, "test:hello", "Hello Redis!")
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code": 500,
			"msg":  "Redis Set 失败: " + err.Error(),
		})
		return
	}

	// 2. 测试 Get
	value, err := service.Redis.Get(ctx, "test:hello")
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code": 500,
			"msg":  "Redis Get 失败: " + err.Error(),
		})
		return
	}

	// 3. 测试 SetEX（带过期时间）
	err = service.Redis.SetEX(ctx, "test:expired", "将在10秒后过期", 10)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code": 500,
			"msg":  "Redis SetEX 失败: " + err.Error(),
		})
		return
	}

	// 4. 测试 Hash 操作
	err = service.Redis.HSet(ctx, "test:user:1001", "name", "张三")
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code": 500,
			"msg":  "Redis HSet 失败: " + err.Error(),
		})
		return
	}

	userName, err := service.Redis.HGet(ctx, "test:user:1001", "name")
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code": 500,
			"msg":  "Redis HGet 失败: " + err.Error(),
		})
		return
	}

	// 5. 测试计数器
	count, err := service.Redis.Incr(ctx, "test:counter")
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code": 500,
			"msg":  "Redis Incr 失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code": 200,
		"msg":  "Redis 测试成功！",
		"data": g.Map{
			"basic_get":     value,
			"hash_get":      userName,
			"counter_value": count,
		},
	})
}
