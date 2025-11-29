#!/bin/bash

echo "🚧 Building for Linux (amd64)..."

# Cross-compile for Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cex-exchange-linux main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful: cex-exchange-linux"
    echo "📋 Next steps:"
    echo "1. Upload 'cex-exchange-linux' to your server."
    echo "2. Upload 'manifest/config/config.yaml' to your server."
    echo "3. Run it!"
else
    echo "❌ Build failed"
    exit 1
fi
