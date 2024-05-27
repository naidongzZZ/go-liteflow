# 使用官方的 Go 镜像作为构建环境
FROM golang:1.22 as builder

# 设置工作目录
WORKDIR /app
# 设置必要的环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
	GOPROXY="https://goproxy.cn,direct"

# 复制所有项目文件到 docekr 容器当中
COPY ../ ./

# 下载依赖项
RUN go mod download

# 构建可执行文件
# 请根据实际情况调整构建命令
RUN go build -a -installsuffix cgo -o liteflow ./cmd

# 使用 scratch 镜像作为运行环境压缩空间
# FROM scratch as prod

# 使用 Alpine 作为开发测试环境的基础镜像
FROM alpine:latest as dev

# 更新包索引并安装调试工具
RUN apk update && apk add --no-cache \
    curl 

WORKDIR /root/

# 从构建器镜像中复制可执行文件
COPY --from=builder /app/liteflow .

# 暴露端口
EXPOSE 20021

# 运行可执行文件
CMD ["./liteflow"]