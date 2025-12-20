#!/bin/bash

# gRPC 代码生成脚本
export PATH=$PATH:$(go env GOPATH)/bin

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    echo "错误: protoc 未安装"
    echo "请安装 protoc: brew install protobuf"
    exit 1
fi

# 检查 protoc-gen-go 是否安装
if ! command -v protoc-gen-go &> /dev/null; then
    echo "安装 protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# 检查 protoc-gen-go-grpc 是否安装
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "安装 protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# 创建输出目录
mkdir -p api/proto/exchange

# 生成 Go 代码
echo "生成 gRPC 代码..."
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/exchange.proto

if [ $? -eq 0 ]; then
    echo "✅ gRPC 代码生成成功！"
    echo "生成的文件:"
    ls -lh api/proto/exchange.pb.go
    ls -lh api/proto/exchange_grpc.pb.go
else
    echo "❌ gRPC 代码生成失败"
    exit 1
fi
