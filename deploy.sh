#!/bin/bash

# CEX Exchange Docker 部署脚本
# 使用方法: ./deploy.sh [dev|prod] [start|stop|restart|logs|build]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
PROJECT_NAME="cex-exchange"
COMPOSE_FILE="docker-compose.yml"
PROD_COMPOSE_FILE="docker-compose.prod.yml"

# 打印带颜色的消息
print_message() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  CEX Exchange Docker 部署工具${NC}"
    echo -e "${BLUE}================================${NC}"
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi

    print_message "Docker 环境检查通过"
}

# 检查必要文件
check_files() {
    local missing_files=()
    
    if [[ ! -f "$COMPOSE_FILE" ]]; then
        missing_files+=("$COMPOSE_FILE")
    fi
    
    if [[ ! -f "Dockerfile" ]]; then
        missing_files+=("Dockerfile")
    fi
    
    if [[ ! -f "frontend/Dockerfile" ]]; then
        missing_files+=("frontend/Dockerfile")
    fi
    
    if [[ ! -f "cex-exchange/config.yaml" ]]; then
        missing_files+=("cex-exchange/config.yaml")
    fi
    
    if [[ ${#missing_files[@]} -gt 0 ]]; then
        print_error "缺少必要文件:"
        for file in "${missing_files[@]}"; do
            echo "  - $file"
        done
        exit 1
    fi
    
    print_message "文件检查通过"
}

# 创建必要的目录
create_directories() {
    mkdir -p logs/nginx
    mkdir -p ssl
    print_message "创建必要目录"
}

# 生成自签名SSL证书（开发环境）
generate_ssl_cert() {
    if [[ ! -f "ssl/cert.pem" ]] || [[ ! -f "ssl/key.pem" ]]; then
        print_message "生成自签名SSL证书..."
        mkdir -p ssl
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout ssl/key.pem -out ssl/cert.pem \
            -subj "/C=CN/ST=State/L=City/O=Organization/CN=localhost"
        print_message "SSL证书生成完成"
    fi
}

# 启动服务
start_services() {
    local env=$1
    local compose_files=("$COMPOSE_FILE")
    
    if [[ "$env" == "prod" ]]; then
        compose_files+=("$PROD_COMPOSE_FILE")
        generate_ssl_cert
    fi
    
    print_message "启动 $env 环境服务..."
    
    local compose_cmd="docker-compose"
    for file in "${compose_files[@]}"; do
        compose_cmd="$compose_cmd -f $file"
    done
    
    $compose_cmd up -d
    
    print_message "服务启动完成"
    print_message "前端访问地址: http://localhost"
    print_message "后端API地址: http://localhost:8080"
    if [[ "$env" == "prod" ]]; then
        print_message "HTTPS地址: https://localhost"
    fi
}

# 停止服务
stop_services() {
    local env=$1
    local compose_files=("$COMPOSE_FILE")
    
    if [[ "$env" == "prod" ]]; then
        compose_files+=("$PROD_COMPOSE_FILE")
    fi
    
    print_message "停止 $env 环境服务..."
    
    local compose_cmd="docker-compose"
    for file in "${compose_files[@]}"; do
        compose_cmd="$compose_cmd -f $file"
    done
    
    $compose_cmd down
    
    print_message "服务停止完成"
}

# 重启服务
restart_services() {
    local env=$1
    stop_services "$env"
    sleep 2
    start_services "$env"
}

# 查看日志
show_logs() {
    local env=$1
    local service=$2
    local compose_files=("$COMPOSE_FILE")
    
    if [[ "$env" == "prod" ]]; then
        compose_files+=("$PROD_COMPOSE_FILE")
    fi
    
    local compose_cmd="docker-compose"
    for file in "${compose_files[@]}"; do
        compose_cmd="$compose_cmd -f $file"
    done
    
    if [[ -n "$service" ]]; then
        print_message "查看 $service 服务日志..."
        $compose_cmd logs -f "$service"
    else
        print_message "查看所有服务日志..."
        $compose_cmd logs -f
    fi
}

# 构建镜像
build_images() {
    local env=$1
    local compose_files=("$COMPOSE_FILE")
    
    if [[ "$env" == "prod" ]]; then
        compose_files+=("$PROD_COMPOSE_FILE")
    fi
    
    print_message "构建 $env 环境镜像..."
    
    local compose_cmd="docker-compose"
    for file in "${compose_files[@]}"; do
        compose_cmd="$compose_cmd -f $file"
    done
    
    $compose_cmd build --no-cache
    
    print_message "镜像构建完成"
}

# 清理资源
cleanup() {
    print_warning "清理Docker资源..."
    
    # 停止并删除容器
    docker-compose down -v 2>/dev/null || true
    docker-compose -f "$COMPOSE_FILE" -f "$PROD_COMPOSE_FILE" down -v 2>/dev/null || true
    
    # 删除未使用的镜像
    docker image prune -f
    
    # 删除未使用的网络
    docker network prune -f
    
    # 删除未使用的卷
    docker volume prune -f
    
    print_message "清理完成"
}

# 显示状态
show_status() {
    local env=$1
    local compose_files=("$COMPOSE_FILE")
    
    if [[ "$env" == "prod" ]]; then
        compose_files+=("$PROD_COMPOSE_FILE")
    fi
    
    local compose_cmd="docker-compose"
    for file in "${compose_files[@]}"; do
        compose_cmd="$compose_cmd -f $file"
    done
    
    print_message "服务状态:"
    $compose_cmd ps
    
    print_message "资源使用情况:"
    docker stats --no-stream
}

# 显示帮助信息
show_help() {
    echo "使用方法: $0 [环境] [命令] [服务名]"
    echo ""
    echo "环境:"
    echo "  dev    - 开发环境"
    echo "  prod   - 生产环境"
    echo ""
    echo "命令:"
    echo "  start     - 启动服务"
    echo "  stop      - 停止服务"
    echo "  restart   - 重启服务"
    echo "  logs      - 查看日志"
    echo "  build     - 构建镜像"
    echo "  status    - 查看状态"
    echo "  cleanup   - 清理资源"
    echo "  help      - 显示帮助"
    echo ""
    echo "示例:"
    echo "  $0 dev start          # 启动开发环境"
    echo "  $0 prod start         # 启动生产环境"
    echo "  $0 dev logs backend   # 查看后端日志"
    echo "  $0 prod status        # 查看生产环境状态"
}

# 主函数
main() {
    print_header
    
    local env=${1:-dev}
    local command=${2:-start}
    local service=${3:-}
    
    # 如果是help命令，直接显示帮助
    if [[ "$env" == "help" || "$command" == "help" ]]; then
        show_help
        exit 0
    fi
    
    # 检查环境参数
    if [[ "$env" != "dev" && "$env" != "prod" ]]; then
        print_error "无效的环境参数: $env"
        show_help
        exit 1
    fi
    
    # 检查命令参数
    case "$command" in
        start|stop|restart|logs|build|status|cleanup)
            ;;
        *)
            print_error "无效的命令: $command"
            show_help
            exit 1
            ;;
    esac
    
    # 执行命令
    case "$command" in
        start)
            check_docker
            check_files
            create_directories
            start_services "$env"
            ;;
        stop)
            stop_services "$env"
            ;;
        restart)
            restart_services "$env"
            ;;
        logs)
            show_logs "$env" "$service"
            ;;
        build)
            check_docker
            check_files
            build_images "$env"
            ;;
        status)
            show_status "$env"
            ;;
        cleanup)
            cleanup
            ;;
        help)
            show_help
            ;;
    esac
}

# 如果脚本被直接执行
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 