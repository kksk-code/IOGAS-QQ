# 项目信息
APP_NAME := synctoqq
SRC_DIR := .
BUILD_DIR := ./build

# 编译器选项
GO := go
GO_FLAGS :=

# 目标平台和架构
OS_LIST := linux windows
ARCH_LIST := amd64 arm64

# 默认目标：编译所有
.PHONY: all
all: build

# 生成二进制文件
.PHONY: build
build: clean
	@mkdir -p $(BUILD_DIR)
	$(foreach os, $(OS_LIST), \
		$(foreach arch, $(ARCH_LIST), \
			GOOS=$(os) GOARCH=$(arch) $(GO) build -o $(BUILD_DIR)/$(APP_NAME)-$(os)-$(arch) $(SRC_DIR); \
		) \
	)

# 清理生成的二进制文件
.PHONY: clean
clean:
	@rm -rf $(BUILD_DIR)

# 针对特定平台进行编译
.PHONY: linux
linux:
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(SRC_DIR)

.PHONY: windows
windows:
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(SRC_DIR)

# 运行项目
.PHONY: run
run:
	$(GO) run $(SRC_DIR)

# 安装依赖项
.PHONY: deps
deps:
	$(GO) mod tidy
