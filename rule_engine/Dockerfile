# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o rule-engine ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装基础工具
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建工作目录
WORKDIR /app

# 创建配置目录
RUN mkdir -p configs logs

# 从构建阶段复制二进制文件
COPY --from=builder /app/rule-engine .
COPY --from=builder /app/configs/config.yaml ./configs/

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./rule-engine", "-config", "configs/config.yaml"] 