#!/bin/bash

echo "🚀 Starting CEX Exchange with GoFrame..."

# 检查配置文件
if [ ! -f "manifest/config/config.yaml" ]; then
    echo "❌ Config file not found!"
    exit 1
fi

# 运行项目
go run main.go
