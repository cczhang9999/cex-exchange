//go:build wireinject
// +build wireinject

// 这两个是构建标签（Build Tags），告诉 Go 编译器在常规编译时忽略此文件。
// 只有在使用 `wire` 工具生成代码时，才会处理这个文件。

package main

import (
	"backend-v2/internal/biz"
	"backend-v2/internal/conf"
	"backend-v2/internal/data"
	"backend-v2/internal/server"
	"backend-v2/internal/service"

	"github.com/google/wire"
)

// initApp 是应用程序的初始化入口（Injector）。
// 它接收配置参数 bc (*conf.Bootstrap)，并负责构建并返回：
// 1. *server.Server: 已经装配好所有依赖的服务器实例
// 2. func(): 用于资源释放的清理函数（例如关闭数据库连接）
// 3. error: 初始化过程中可能产生的错误
func initApp(bc *conf.Bootstrap) (*server.Server, func(), error) {
	// wire.Build 会根据传入的 ProviderSet（供应者集合）和构造函数，
	// 自动分析依赖关系图，并生成对应的初始化代码（通常在 wire_gen.go 中）。
	panic(wire.Build(
		server.ProviderSet,    // 网络服务层的依赖集合（如 HTTP/gRPC server）
		data.ProviderSet,      // 数据访问层的通用依赖（如 DB、Redis 客户端）
		data.UserProviderSet,  // 用户模块特有的数据层依赖
		biz.ProviderSet,       // 业务逻辑层（Domain/UseCase）的依赖集合
		service.ProviderSet,   // 服务实现层（实现 Proto 定义的接口）的依赖集合
		server.NewServer,      // 最终创建 Server 实例的构造函数
	))
}

// 关键点说明：
// 自动化组装：你不需要在 main.go 里手动写 repo := data.NewUserRepo(...), uc := biz.NewUserUseCase(repo), svc := service.NewUserService(uc) 这样繁琐的链式初始化。Wire 会帮你搞定。
// wire.Build：你只需要把各个层级的 ProviderSet（在各自包的 pkg.go 或类似文件中定义）放进去，Wire 就会自动寻找如何从 bc *conf.Bootstrap 一步步构建出 *server.Server。
// 生成的代码：当你运行 wire 命令后，它会生成一个 wire_gen.go 文件，里面包含了真正可以运行的初始化逻辑。
// panic 的作用：在 
// wire.go
//  中使用 panic 是固定写法。因为这个函数体实际上不会被执行，Wire 工具只是读取参数来分析依赖，生成的 wire_gen.go 会包含正确的实现。