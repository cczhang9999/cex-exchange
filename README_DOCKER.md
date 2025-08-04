# CEX Exchange Docker 部署指南

## 🚀 快速开始

### 1. 克隆项目
```bash
git clone <your-repo-url>
cd cex-exchange
```

### 2. 一键部署
```bash
# 开发环境
./deploy.sh dev start

# 生产环境
./deploy.sh prod start
```

### 3. 访问应用
- **前端**: http://localhost
- **后端API**: http://localhost:8080
- **生产环境HTTPS**: https://localhost

## 📋 部署命令

```bash
# 查看帮助
./deploy.sh help

# 开发环境
./deploy.sh dev start      # 启动
./deploy.sh dev stop       # 停止
./deploy.sh dev restart    # 重启
./deploy.sh dev logs       # 查看日志
./deploy.sh dev status     # 查看状态

# 生产环境
./deploy.sh prod start     # 启动
./deploy.sh prod stop      # 停止
./deploy.sh prod restart   # 重启
./deploy.sh prod logs      # 查看日志
./deploy.sh prod status    # 查看状态

# 构建镜像
./deploy.sh dev build      # 构建开发环境镜像
./deploy.sh prod build     # 构建生产环境镜像

# 清理资源
./deploy.sh cleanup        # 清理Docker资源
```

## 🏗️ 项目结构

```
cex-exchange/
├── Dockerfile                 # 后端Dockerfile
├── docker-compose.yml         # 开发环境配置
├── docker-compose.prod.yml    # 生产环境配置
├── deploy.sh                  # 部署脚本
├── nginx.prod.conf           # 生产环境nginx配置
├── cex-exchange/             # 后端代码
│   ├── main.go
│   ├── config.yaml
│   └── ...
└── frontend/                 # 前端代码
    ├── Dockerfile
    ├── nginx.conf
    └── ...
```

## 🔧 服务说明

| 服务 | 端口 | 说明 |
|------|------|------|
| frontend | 80 | Vue.js前端应用 |
| backend | 8080 | Go后端API服务 |
| mysql | 3306 | MySQL数据库 |
| redis | 6379 | Redis缓存 |
| nginx | 80/443 | 反向代理(生产环境) |

## ⚙️ 配置说明

### 环境变量
复制 `env.example` 为 `.env` 并修改配置：

```bash
cp env.example .env
```

### 数据库配置
编辑 `cex-exchange/config.yaml`:

```yaml
database:
  host: mysql  # Docker服务名
  port: 3306
  user: root
  password: 123456
  name: cex_exchange
```

## 🔍 故障排除

### 常见问题

1. **端口冲突**
   ```bash
   # 修改docker-compose.yml中的端口映射
   ports:
     - "8081:8080"  # 改为其他端口
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   ./deploy.sh dev logs mysql
   ```

3. **前端无法访问**
   ```bash
   # 检查前端状态
   ./deploy.sh dev logs frontend
   ```

### 查看日志
```bash
# 查看所有服务日志
./deploy.sh dev logs

# 查看特定服务日志
./deploy.sh dev logs backend
./deploy.sh dev logs frontend
./deploy.sh dev logs mysql
```

### 健康检查
```bash
# 检查服务状态
./deploy.sh dev status

# 手动检查API
curl http://localhost:8080/api/klines
```

## 🔒 安全建议

1. **修改默认密码**
   - 数据库密码
   - JWT密钥
   - Redis密码

2. **使用HTTPS** (生产环境)
   - 配置SSL证书
   - 启用安全头

3. **网络隔离**
   - 使用Docker networks
   - 限制端口暴露

## 📊 监控和维护

### 资源监控
```bash
# 查看容器资源使用
docker stats

# 查看磁盘使用
docker system df
```

### 备份和恢复
```bash
# 备份数据库
docker exec cex-mysql mysqldump -u root -p123456 cex_exchange > backup.sql

# 恢复数据库
docker exec -i cex-mysql mysql -u root -p123456 cex_exchange < backup.sql
```

## 🆘 获取帮助

```bash
# 查看部署脚本帮助
./deploy.sh help

# 查看详细文档
cat DOCKER_DEPLOY.md
```

## 📝 更新日志

- **v1.0.0**: 初始Docker部署配置
- 支持开发和生产环境
- 包含完整的监控和日志
- 支持HTTPS和负载均衡 