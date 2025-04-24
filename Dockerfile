# 构建阶段
FROM golang:1.24-alpine AS builder

# 设置Go环境变量
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# 设置工作目录
WORKDIR /app

# 安装git和其他必要的工具
RUN apk add --no-cache git

# 复制go.mod和go.sum文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 编译应用
RUN GOOS=linux go build -a -installsuffix cgo -o scira .

# 运行阶段
FROM alpine:3.18

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN adduser -D -g '' appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的应用和必要的文件
COPY --from=builder /app/scira .
COPY --from=builder /app/config/ ./config/

# 切换到非root用户
USER appuser

# 声明应用使用的端口
EXPOSE 8080

# 应用启动命令
CMD ["./scira"] 