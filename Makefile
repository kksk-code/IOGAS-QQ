# Makefile 变量定义
BINARY_NAME=synctoqq     # 输出的可执行文件名
LINUX_BINARY=synctoqq_linux   # Linux 平台的可执行文件名
GOOS=linux                 # 目标操作系统
GOARCH=amd64               # 目标架构
REMOTE_HOST=user@your_linux_server # 远程 Linux 服务器
REMOTE_DIR=/path/to/deploy # 远程部署目录

# 默认目标：编译项目
all: build

# 编译项目为本地可执行文件
build:
	go build -o $(BINARY_NAME)

# 交叉编译项目为 Linux 可执行文件
build-linux:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(LINUX_BINARY)

# 将 Linux 可执行文件上传到远程服务器
deploy: build-linux
	scp $(LINUX_BINARY) $(REMOTE_HOST):$(REMOTE_DIR)

# 在远程服务器上启动程序
start:
	ssh $(REMOTE_HOST) 'cd $(REMOTE_DIR) && ./$(LINUX_BINARY)'

# 清理构建的二进制文件
clean:
	rm -f $(BINARY_NAME) $(LINUX_BINARY)

# 帮助命令，显示所有可用的 make 命令
help:
	@echo "Makefile commands:"
	@echo "  make build         - 编译项目为本地可执行文件"
	@echo "  make build-linux   - 交叉编译为 Linux 可执行文件"
	@echo "  make deploy        - 将可执行文件上传到远程服务器"
	@echo "  make start         - 在远程服务器上启动程序"
	@echo "  make clean         - 清理生成的二进制文件"
	@echo "  make help          - 显示帮助信息"
