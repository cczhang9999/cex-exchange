package main

import (
	"backend-v2/internal/conf"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 1. 初始化配置 (Bootstrap)
	// 在实际生产环境中，这些配置通常从 config.yaml 文件中读取，
	// 这里为了演示或简化，直接在代码中硬编码了配置信息。
	bc := &conf.Bootstrap{
		Server: &conf.Server{
			Http: &conf.ServerHTTP{Addr: ":8080", Timeout: "1s"}, // HTTP 服务配置
			Grpc: &conf.ServerGRPC{Addr: ":9000", Timeout: "1s"}, // gRPC 服务配置
		},
		Data: &conf.Data{
			// 数据库连接配置（MySQL）
			Database: &conf.Database{
				Driver: "mysql", 
				Source: "hobart:123456@tcp(212.227.166.131:9257)/cex_exchange?charset=utf8mb4&parseTime=True&loc=Local",
			},
			// Redis 缓存配置
			Redis: &conf.Redis{
				Addr: "194.164.194.118:9502", 
				Password: "pass123editmelol", 
				ReadTimeout: "5s", 
				WriteTimeout: "5s",
			},
		},
		Auth: &conf.Auth{
			JwtSecret: "super-secret-key-change-me", // 用于 JWT 签名的密钥
			JwtExpiry: "24h",                         // Token 过期时间
		},
	}

	// 2. 依赖注入与应用初始化
	// initApp 函数定义在 wire.go 中，由 wire 工具生成的 wire_gen.go 实现。
	// 它会自动根据 bc 配置创建出所有的 Repository, UseCase, Service 和 Server。
	app, cleanup, err := initApp(bc)
	if err != nil {
		panic(err) // 如果初始化失败，直接宕机
	}
	
	// 3. 注册资源释放回调
	// 使用 defer 确保在程序退出前执行 cleanup 函数。
	// cleanup 通常包含：关闭数据库连接、关闭 Redis 连接、刷新日志缓冲区等。
	defer cleanup()

	// 4. 运行服务器
	// 启动 HTTP 和 gRPC 服务。由于 app.Run() 通常是非阻塞或内部处理了运行逻辑，
	// 这里通过错误捕获来确保启动成功。
	if err := app.Run(); err != nil {
		panic(err)
	}

	// 5. 优雅关机 (Graceful Shutdown)
	// 创建一个监听系统信号的通道（Channel）。
	quit := make(chan os.Signal, 1)
	
	// 监听中断信号（Ctrl+C）和终止信号（如 kill 命令）。
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// 阻塞在这里，直到接收到上述信号。
	<-quit
	
	// 收到信号后打印日志并开始清理流程。
	fmt.Println("Shutting down server...")
}

// 核心流程总结：
// 配置准备：定义了服务器地址、数据库、Redis 以及安全认证相关的参数。
// initApp(bc)
// ：这是最关键的一步，它通过 Wire 将整个项目的所有层级串联起来。
// app.Run()：正式启动监听端口，开始处理客户端请求。
// 优雅退出：通过拦截操作系统信号，确保在关机前能够正常关闭数据库等外部连接，避免数据损坏或资源泄露。