# CEX交易所部署文档

## 1. 环境准备
- 需安装 Docker、docker-compose
- MySQL 8.0+

## 2. Dockerfile（后端示例）
```dockerfile
FROM golang:1.21-alpine as builder
WORKDIR /app
COPY . .
RUN go build -o cex-exchange main.go models.go auth.go funds.go match.go kline.go admin.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/cex-exchange .
COPY config.yaml .
CMD ["./cex-exchange"]
```

## 3. Dockerfile（前端示例）
```dockerfile
FROM node:20-alpine as builder
WORKDIR /app
COPY frontend/ .
RUN npm install && npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
```

## 4. docker-compose.yml
```yaml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: yourpassword
      MYSQL_DATABASE: cex_exchange
    ports:
      - "3306:3306"
    volumes:
      - ./mysql_data:/var/lib/mysql
  backend:
    build: .
    depends_on:
      - mysql
    environment:
      - CONFIG=config.yaml
    ports:
      - "8080:8080"
  frontend:
    build:
      context: .
      dockerfile: frontend/Dockerfile
    ports:
      - "80:80"
```

## 5. 环境变量说明
- `config.yaml`：后端配置文件，包含数据库、JWT等信息
- `MYSQL_ROOT_PASSWORD`、`MYSQL_DATABASE`：MySQL初始化参数

## 6. 一键部署脚本（deploy.sh）
```sh
#!/bin/bash
docker-compose up --build -d
```

---

如需自定义端口、数据库等，请修改 `config.yaml` 和 `docker-compose.yml`。 

 go run 

go run main.go models.go auth.go funds.go match.go kline.go admin.go

go run main.go models.go auth.go funds.go match.go kline.go admin.go websocket.go


现在的脚本说明：cd fronend

npm run dev: 只启动前端Vite开发服务器
npm run backend: 只启动后端Go服务（使用nodemon热重载）
npm run start: 同时启动前后端服务
npm run build: 构建前端生产版本
npm run preview: 预览构建后的前端


cd /Users/zhanjun/IdeaProjects/cex-exchange/cex-exchange
go run cmd/grpc/main.go


go run cmd/server/main.go

make build
./bin/server