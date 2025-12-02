#!/bin/bash

# 实时行情监控系统启动脚本

echo "========================================="
echo "  实时行情监控系统 - 启动脚本"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查后端是否已编译
if [ ! -f "cex-exchange/cex-exchange" ]; then
    echo -e "${YELLOW}后端未编译，正在编译...${NC}"
    cd cex-exchange
    go build -o cex-exchange main.go
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 后端编译成功${NC}"
    else
        echo -e "${RED}✗ 后端编译失败${NC}"
        exit 1
    fi
    cd ..
fi

# 启动后端服务
echo -e "${YELLOW}正在启动后端服务...${NC}"
cd cex-exchange
./cex-exchange &
BACKEND_PID=$!
echo -e "${GREEN}✓ 后端服务已启动 (PID: $BACKEND_PID)${NC}"
echo "  - HTTP服务: http://localhost:8080"
echo "  - WebSocket: ws://localhost:8080/ws"
echo "  - 状态查询: http://localhost:8080/ws/status"
cd ..

# 等待后端启动
echo ""
echo -e "${YELLOW}等待后端服务启动...${NC}"
sleep 3

# 检查前端依赖
if [ ! -d "frontend/node_modules" ]; then
    echo -e "${YELLOW}前端依赖未安装，正在安装...${NC}"
    cd frontend
    npm install
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 前端依赖安装成功${NC}"
    else
        echo -e "${RED}✗ 前端依赖安装失败${NC}"
        kill $BACKEND_PID
        exit 1
    fi
    cd ..
fi

# 启动前端服务
echo ""
echo -e "${YELLOW}正在启动前端服务...${NC}"
cd frontend
npm run dev &
FRONTEND_PID=$!
echo -e "${GREEN}✓ 前端服务已启动 (PID: $FRONTEND_PID)${NC}"
echo "  - 前端地址: http://localhost:5173"
cd ..

# 显示访问信息
echo ""
echo "========================================="
echo -e "${GREEN}系统启动成功！${NC}"
echo "========================================="
echo ""
echo "📊 访问地址："
echo "  - 首页: http://localhost:5173"
echo "  - 实时行情: http://localhost:5173/market"
echo "  - 交易页面: http://localhost:5173/trade"
echo ""
echo "🔧 后端服务："
echo "  - API地址: http://localhost:8080/api"
echo "  - WebSocket: ws://localhost:8080/ws"
echo "  - 状态查询: http://localhost:8080/ws/status"
echo ""
echo "📝 进程信息："
echo "  - 后端PID: $BACKEND_PID"
echo "  - 前端PID: $FRONTEND_PID"
echo ""
echo "⚠️  按 Ctrl+C 停止所有服务"
echo "========================================="

# 保存PID到文件
echo $BACKEND_PID > .backend.pid
echo $FRONTEND_PID > .frontend.pid

# 等待用户中断
trap "echo ''; echo -e '${YELLOW}正在停止服务...${NC}'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; rm -f .backend.pid .frontend.pid; echo -e '${GREEN}服务已停止${NC}'; exit 0" INT

# 保持脚本运行
wait
