#!/bin/bash

# 定义颜色
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 显示带颜色的信息函数
info() {
    echo -e "${BLUE}[INFO] $1${NC}"
}

success() {
    echo -e "${GREEN}[SUCCESS] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
}

# 脚本开始
info "开始构建和部署Scira2API应用"

# 检查是否安装了Docker
if ! command -v docker &> /dev/null; then
    error "Docker未安装，请先安装Docker"
    exit 1
fi

# 检查docker compose命令是否可用
if ! docker compose version &> /dev/null; then
    error "Docker Compose未安装或不可用，请确保Docker版本足够新（Docker Engine 20.10+）"
    exit 1
fi

# 获取当前时间戳作为版本号
TIMESTAMP=$(date +%Y%m%d%H%M%S)
TAG="scira2api:$TIMESTAMP"

# 构建新镜像
info "构建新的Docker镜像: $TAG"
if docker build -t $TAG .; then
    success "Docker镜像构建成功: $TAG"
else
    error "Docker镜像构建失败"
    exit 1
fi

# 添加latest标签
info "添加latest标签"
docker tag $TAG scira2api:latest

# 检查当前是否有正在运行的容器
if docker ps -q --filter name=scira2api | grep -q .; then
    info "检测到现有的Scira2API容器正在运行"
    
    # 停止并移除旧的应用容器
    info "停止并移除旧的应用容器"
    docker compose stop scira2api
    docker compose rm -f scira2api
fi

# 启动新的容器
info "使用新镜像启动应用容器"
if docker compose up -d; then
    success "应用容器启动成功"
else
    error "应用容器启动失败"
    exit 1
fi

# 显示运行中的容器
info "当前运行的容器状态:"
docker compose ps


success "部署完成! Scira2API应用现在已经更新并运行在最新版本" 