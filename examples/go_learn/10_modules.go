package main

import (
	"fmt"
)

// 示例 10: Go 模块化和最佳实践

// 1. 包的组织结构
// 标准的 Go 项目结构：
// .
// ├── cmd/
// │   └── myapp/
// │       └── main.go          # 应用入口
// ├── internal/
// │   ├── dao/                 # 数据访问层
// │   ├── service/             # 业务逻辑层
// │   ├── controller/          # 控制器层
// │   └── model/               # 数据模型
// ├── pkg/                     # 公共工具包
// │   ├── utils/
// │   └── errors/
// ├── api/                     # API 定义
// ├── configs/                 # 配置文件
// └── go.mod

func main() {
	fmt.Println("=== Go 学习示例 10: 模块化和项目结构 ===")

	// 2. 命名规范
	// - 包名：小写，无下划线，简洁有意义
	// - 变量名：驼峰命名，见名知意
	// - 常量名：全大写，下划线分隔，或驼峰

	// 好的命名示例
	type UserService struct{}        // ✓ 服务层
	type OrderDao interface{}        // ✓ 数据访问接口
	const MaxOrderAmount = 1000000.0 // ✓ 常量

	// 3. 函数设计原则
	// - 单一职责：每个函数只做一件事
	// - 短小精悍：理想 10-20 行
	// - 无副作用：避免修改全局状态

	// 4. 错误处理最佳实践
	// - 总是检查错误
	// - 错误向上传递
	// - 使用有意义的错误信息

	// 5. 接口设计
	// - 越小越好（接口越小越好）
	// - 依赖接口而非实现

	// 6. 并发安全
	// - 使用 chan 进行通信
	// - 使用 sync.Mutex 保护共享数据
	// - 避免死锁

	// 7. 依赖注入
	// 通过接口注入依赖，便于测试和替换实现

	fmt.Println("\n=== Go 项目结构示例 ===")
	fmt.Println(`
cex-exchange/                          # 项目根目录
├── cmd/                              # 应用程序入口
│   ├── server/                       # HTTP/gRPC 服务器
│   ├── worker/                       # 后台任务worker
│   └── cli/                          # 命令行工具
├── internal/                         # 私有应用程序代码
│   ├── api/                          # API 协议定义
│   ├── dao/                          # 数据访问层
│   │   ├── user.go
│   │   ├── order.go
│   │   └── trade.go
│   ├── service/                      # 业务逻辑层
│   │   ├── user.go
│   │   ├── order.go
│   │   └── trade.go
│   ├── controller/                   # 控制器层
│   ├── model/                        # 数据模型
│   │   ├── entity.go                 # 数据库实体
│   │   └── reqstruct.go              # 请求结构体
│   ├── engine/                       # 撮合引擎
│   │   ├── matching.go
│   │   └── orderbook.go
│   └── pkg/                          # 内部工具包
├── pkg/                              # 公共库
│   ├── utils/
│   └── errors/
├── api/                              # API 协议文件（protobuf 等）
├── configs/                          # 配置文件
├── scripts/                          # 脚本文件
├── go.mod                            # 依赖管理
└── README.md                         # 项目说明
`)
}
