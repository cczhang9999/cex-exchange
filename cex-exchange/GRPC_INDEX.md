# gRPC 后台实现 - 文档索引

欢迎使用 CEX Exchange 的 gRPC 后台实现！

## 📚 文档列表

### 1. [快速开始指南](./GRPC_QUICKSTART.md) ⭐ 推荐新手阅读
简洁的入门指南，包含：
- 环境准备步骤
- 快速启动教程
- 常用 API 示例
- 常见问题解答

**适合**: 第一次使用 gRPC 的开发者

---

### 2. [完整技术文档](./GRPC_README.md)
详细的技术文档，包含：
- 完整的 API 文档
- 多语言客户端示例（Go/Python/JavaScript）
- 高级配置（TLS、拦截器）
- 性能优化建议
- 测试工具使用

**适合**: 需要深入了解的开发者

---

### 3. [实现总结](./GRPC_IMPLEMENTATION_SUMMARY.md)
项目实现概览，包含：
- 已创建文件列表
- 核心功能说明
- 架构设计图
- 扩展建议
- 学习路径

**适合**: 想要了解整体架构的开发者

---

## 🚀 快速导航

### 我想...

#### 🎯 快速开始
👉 阅读 [快速开始指南](./GRPC_QUICKSTART.md)

#### 📖 查看 API 文档
👉 阅读 [完整技术文档](./GRPC_README.md) 的 API 文档部分

#### 💻 编写客户端代码
👉 查看 [examples/grpc_client.go](./examples/grpc_client.go)  
👉 阅读 [完整技术文档](./GRPC_README.md) 的客户端示例部分

#### 🔧 修改 protobuf 定义
👉 编辑 [api/proto/exchange.proto](./api/proto/exchange.proto)  
👉 运行 `./generate_proto.sh` 重新生成代码

#### 🐛 排查问题
👉 查看 [快速开始指南](./GRPC_QUICKSTART.md) 的常见问题部分

#### 🏗️ 了解架构
👉 阅读 [实现总结](./GRPC_IMPLEMENTATION_SUMMARY.md) 的架构设计部分

---

## 📁 项目文件结构

```
cex-exchange/
├── 📄 文档
│   ├── GRPC_QUICKSTART.md              # 快速开始 ⭐
│   ├── GRPC_README.md                  # 完整文档
│   ├── GRPC_IMPLEMENTATION_SUMMARY.md  # 实现总结
│   └── GRPC_INDEX.md                   # 本文件
│
├── 🔧 工具
│   └── generate_proto.sh               # 代码生成脚本
│
├── 📦 API 定义
│   └── api/proto/
│       ├── exchange.proto              # protobuf 定义
│       ├── exchange.pb.go              # 生成的消息代码
│       └── exchange_grpc.pb.go         # 生成的服务代码
│
├── 💻 服务实现
│   ├── cmd/grpc/main.go                # 服务器启动程序
│   ├── internal/grpc/server.go         # gRPC 服务实现
│   └── internal/service/
│       ├── auth.go                     # 认证服务
│       ├── order.go                    # 订单服务
│       ├── funds.go                    # 资金服务
│       └── market.go                   # 市场数据服务
│
└── 📝 示例
    └── examples/grpc_client.go         # 客户端示例
```

---

## 🎓 学习路径

### 初学者路径
1. 📖 阅读 [快速开始指南](./GRPC_QUICKSTART.md)
2. 🛠️ 按步骤安装环境和启动服务
3. 🧪 运行客户端示例 `go run examples/grpc_client.go`
4. 🔍 查看 [api/proto/exchange.proto](./api/proto/exchange.proto) 了解接口定义
5. ✏️ 尝试修改客户端代码，调用不同的 API

### 进阶路径
1. 📚 阅读 [完整技术文档](./GRPC_README.md)
2. 🔧 学习如何启用 TLS 和添加拦截器
3. 🌐 尝试用其他语言（Python/JavaScript）编写客户端
4. 📊 使用 ghz 进行性能测试
5. 🚀 根据业务需求扩展新的 RPC 方法

### 高级路径
1. 🏗️ 阅读 [实现总结](./GRPC_IMPLEMENTATION_SUMMARY.md) 了解架构
2. 🔨 修改 [internal/grpc/server.go](./internal/grpc/server.go) 添加新功能
3. 📈 实现监控和日志
4. 🐳 Docker 容器化部署
5. ☸️ Kubernetes 集群部署

---

## 🔗 外部资源

- [gRPC 官方文档](https://grpc.io/docs/)
- [Protocol Buffers 文档](https://developers.google.com/protocol-buffers)
- [gRPC-Go 教程](https://grpc.io/docs/languages/go/quickstart/)
- [GoFrame 文档](https://goframe.org/)

---

## 💡 提示

- 🌟 **新手**: 从快速开始指南开始
- 📖 **查 API**: 使用完整技术文档
- 🔍 **找示例**: 查看 examples 目录
- 🐛 **遇到问题**: 先看常见问题，再查日志

---

## 📞 获取帮助

如果遇到问题：
1. 查看 [快速开始指南](./GRPC_QUICKSTART.md) 的常见问题部分
2. 检查服务器日志
3. 使用 `grpcurl` 测试接口
4. 查看 gRPC 官方文档

---

**祝你使用愉快！** 🎉
