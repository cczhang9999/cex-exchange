# Docker 部署指南

本项目支持使用 Docker 进行容器化部署，包含完整的后端、前端和数据库服务。

## 系统要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 2GB 可用内存
- 至少 10GB 可用磁盘空间

## 快速开始

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd cex-exchange
```

### 2. 构建并启动服务

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 3. 访问应用

- **前端应用**: http://localhost
- **后端API**: http://localhost:8080
- **数据库**: localhost:3306 (用户名: root, 密码: 123456)

## 服务说明

### 后端服务 (backend)
- **端口**: 8080
- **镜像**: 基于 golang:1.21-alpine 构建
- **功能**: 提供 REST API 和 WebSocket 服务
- **健康检查**: http://localhost:8080/api/klines

### 前端服务 (frontend)
- **端口**: 80
- **镜像**: 基于 nginx:alpine 构建
- **功能**: 提供 Vue.js 前端应用
- **特性**: 支持 Vue Router history 模式，API 代理到后端

### 数据库服务 (mysql)
- **端口**: 3306
- **镜像**: mysql:8.0
- **数据库**: cex_exchange
- **用户名**: root
- **密码**: 123456

### 缓存服务 (redis) - 可选
- **端口**: 6379
- **镜像**: redis:7-alpine
- **功能**: 会话存储和缓存

## 环境配置

### 修改数据库配置

编辑 `cex-exchange/config.yaml`:

```yaml
server:
  port: 8080

database:
  host: mysql  # Docker 服务名
  port: 3306
  user: root
  password: 123456
  name: cex_exchange

jwt:
  secret: your-secret-key
  expires: 86400
```

### 环境变量

可以通过环境变量覆盖配置：

```bash
# 后端环境变量
DB_HOST=mysql
DB_PORT=3306
DB_USER=root
DB_PASSWORD=123456
DB_NAME=cex_exchange
```

## 部署命令

### 开发环境

```bash
# 构建并启动（前台运行）
docker-compose up

# 后台运行
docker-compose up -d

# 重新构建
docker-compose up --build
```

### 生产环境

```bash
# 生产环境构建
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f [service-name]
```

## 数据持久化

### 数据库数据

数据库数据存储在 Docker volume 中：

```bash
# 查看 volumes
docker volume ls

# 备份数据库
docker exec cex-mysql mysqldump -u root -p123456 cex_exchange > backup.sql

# 恢复数据库
docker exec -i cex-mysql mysql -u root -p123456 cex_exchange < backup.sql
```

### 日志文件

```bash
# 查看后端日志
docker-compose logs backend

# 查看前端日志
docker-compose logs frontend

# 查看数据库日志
docker-compose logs mysql
```

## 故障排除

### 常见问题

1. **端口冲突**
   ```bash
   # 修改 docker-compose.yml 中的端口映射
   ports:
     - "8081:8080"  # 改为其他端口
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库服务状态
   docker-compose ps mysql
   
   # 查看数据库日志
   docker-compose logs mysql
   ```

3. **前端无法访问后端**
   ```bash
   # 检查网络连接
   docker network ls
   docker network inspect cex-exchange_cex-network
   ```

### 健康检查

```bash
# 检查所有服务健康状态
docker-compose ps

# 手动检查后端健康状态
curl http://localhost:8080/api/klines

# 手动检查前端健康状态
curl http://localhost/health
```

## 扩展部署

### 使用外部数据库

修改 `docker-compose.yml`，注释掉 mysql 服务，并更新后端配置：

```yaml
backend:
  environment:
    - DB_HOST=your-external-db-host
    - DB_PORT=3306
    - DB_USER=your-user
    - DB_PASSWORD=your-password
    - DB_NAME=cex_exchange
```

### 使用外部 Redis

```yaml
backend:
  environment:
    - REDIS_HOST=your-redis-host
    - REDIS_PORT=6379
    - REDIS_PASSWORD=your-redis-password
```

### 负载均衡

可以使用 nginx 或 traefik 进行负载均衡：

```yaml
# 示例：使用 nginx 负载均衡
nginx:
  image: nginx:alpine
  ports:
    - "80:80"
  volumes:
    - ./nginx.conf:/etc/nginx/nginx.conf
  depends_on:
    - frontend
```

## 监控和维护

### 资源监控

```bash
# 查看容器资源使用情况
docker stats

# 查看磁盘使用情况
docker system df
```

### 清理

```bash
# 停止所有服务
docker-compose down

# 停止并删除 volumes
docker-compose down -v

# 清理未使用的资源
docker system prune -a
```

## 安全建议

1. **修改默认密码**
   - 数据库密码
   - JWT 密钥
   - Redis 密码

2. **使用 HTTPS**
   - 配置 SSL 证书
   - 使用反向代理

3. **网络隔离**
   - 使用 Docker networks
   - 限制端口暴露

4. **定期更新**
   - 更新基础镜像
   - 应用安全补丁

## 备份和恢复

### 完整备份

```bash
# 备份数据库
docker exec cex-mysql mysqldump -u root -p123456 cex_exchange > backup_$(date +%Y%m%d_%H%M%S).sql

# 备份配置文件
tar -czf config_backup_$(date +%Y%m%d_%H%M%S).tar.gz cex-exchange/config.yaml
```

### 恢复

```bash
# 恢复数据库
docker exec -i cex-mysql mysql -u root -p123456 cex_exchange < backup.sql

# 恢复配置
tar -xzf config_backup.tar.gz
``` 