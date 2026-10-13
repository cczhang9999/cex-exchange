package main

import (
	"context"
	"fmt"
	"time"
)

// 示例 3: Context 和超时控制
// 在微服务中，context 用于传递请求作用域的值和取消信号

func main() {
	fmt.Println("=== Go 学习示例 3: Context 和超时控制 ===")

	// 1. 基本 context
	ctx := context.Background()
	fmt.Printf("基础 Context: %v\n", ctx)

	// 2. 带超时的 context（模拟网络请求超时）
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	done := make(chan string, 1)

	go func() {
		// 模拟数据库查询
		time.Sleep(1 * time.Second)
		done <- "查询完成"
	}()

	select {
	case result := <-done:
		fmt.Printf("成功: %s\n", result)
	case <-ctx.Done():
		fmt.Printf("超时或取消: %v\n", ctx.Err())
	}

	// 3. 带值的 context（传递 trace_id, user_id 等）
	ctx = context.WithValue(context.Background(), "trace_id", "abc123")
	ctx = context.WithValue(ctx, "user_id", uint64(1001))

	traceID := ctx.Value("trace_id")
	userID := ctx.Value("user_id")
	fmt.Printf("TraceID: %v, UserID: %v\n", traceID, userID)
}
