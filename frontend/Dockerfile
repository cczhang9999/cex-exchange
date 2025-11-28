# 多阶段构建 - 前端Vue应用
FROM node:18-alpine AS frontend-builder

# 设置工作目录
WORKDIR /app

# 复制package文件
COPY package*.json ./

# 安装依赖（包括开发依赖，因为构建需要vite）
RUN npm install

# 复制源代码
COPY . .

# 构建应用
RUN npm run build

# 运行阶段 - 使用nginx服务静态文件
FROM nginx:alpine

# 从构建阶段复制构建产物
COPY --from=frontend-builder /app/dist /usr/share/nginx/html

# 复制nginx配置
COPY nginx.conf /etc/nginx/nginx.conf

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:80 || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"] 