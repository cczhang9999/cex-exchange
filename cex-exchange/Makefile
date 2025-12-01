# Makefile for gRPC Exchange

.PHONY: help proto server client clean install test

# 默认目标
help:
	@echo "CEX Exchange - gRPC 后台"
	@echo ""
	@echo "可用命令:"
	@echo "  make install    - 安装依赖"
	@echo "  make proto      - 生成 protobuf 代码"
	@echo "  make server     - 启动 gRPC 服务器"
	@echo "  make client     - 运行客户端示例"
	@echo "  make test       - 运行测试"
	@echo "  make clean      - 清理生成的文件"
	@echo "  make build      - 编译服务器"
	@echo "  make docker     - 构建 Docker 镜像"
	@echo ""

# 安装依赖
install:
	@echo "📦 安装 Go 依赖..."
	go mod tidy
	@echo "📦 安装 protoc 插件..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "✅ 依赖安装完成"

# 生成 protobuf 代码
proto:
	@echo "🔨 生成 protobuf 代码..."
	./generate_proto.sh
	@echo "✅ protobuf 代码生成完成"

# 启动 gRPC 服务器
server:
	@echo "🚀 启动 gRPC 服务器..."
	go run cmd/grpc/main.go

# 运行客户端示例
client:
	@echo "💻 运行客户端示例..."
	go run examples/grpc_client.go

# 编译服务器
build:
	@echo "🔨 编译 gRPC 服务器..."
	go build -o bin/grpc-server cmd/grpc/main.go
	@echo "✅ 编译完成: bin/grpc-server"

# 运行测试
test:
	@echo "🧪 运行测试..."
	go test -v ./...

# 清理生成的文件
clean:
	@echo "🧹 清理生成的文件..."
	rm -f api/proto/*.pb.go
	rm -rf bin/
	@echo "✅ 清理完成"

# 构建 Docker 镜像
docker:
	@echo "🐳 构建 Docker 镜像..."
	docker build -t cex-exchange-grpc:latest -f Dockerfile.grpc .
	@echo "✅ Docker 镜像构建完成"

# 检查 protoc 是否安装
check-protoc:
	@which protoc > /dev/null || (echo "❌ protoc 未安装，请运行: brew install protobuf" && exit 1)
	@echo "✅ protoc 已安装"

# 检查 grpcurl 是否安装
check-grpcurl:
	@which grpcurl > /dev/null || (echo "❌ grpcurl 未安装，请运行: brew install grpcurl" && exit 1)
	@echo "✅ grpcurl 已安装"

# 列出所有 gRPC 服务
list-services: check-grpcurl
	@echo "📋 列出所有 gRPC 服务..."
	grpcurl -plaintext localhost:50051 list

# 测试登录接口
test-login: check-grpcurl
	@echo "🧪 测试登录接口..."
	grpcurl -plaintext -d '{"username":"testuser","password":"password123"}' \
		localhost:50051 exchange.ExchangeService/Login

# 格式化代码
fmt:
	@echo "✨ 格式化代码..."
	go fmt ./...
	@echo "✅ 代码格式化完成"

# 代码检查
lint:
	@echo "🔍 代码检查..."
	golangci-lint run
	@echo "✅ 代码检查完成"

# 完整的开发流程
dev: install proto server

# 生产环境构建
prod: clean install proto build
	@echo "🎉 生产环境构建完成"
