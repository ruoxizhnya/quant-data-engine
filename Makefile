# Makefile for Quant Data Engine

# 项目名称
PROJECT_NAME := quant-data-engine

# 可执行文件
BINARY := data-engine

# 构建目录
BUILD_DIR := ./bin

# Go命令
GO := go

# 测试命令
TEST := $(GO) test

# 构建命令
BUILD := $(GO) build

# 安装依赖命令
TIDY := $(GO) mod tidy

# 格式化命令
FMT := $(GO) fmt

# 运行命令
RUN := $(GO) run

# 目标：默认构建
.PHONY: all
all: build

# 目标：构建项目
.PHONY: build
build:
	@echo "Building $(PROJECT_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(BUILD) -o $(BUILD_DIR)/$(BINARY) ./cmd/data-engine

# 目标：运行项目
.PHONY: run
run:
	@echo "Running $(PROJECT_NAME)..."
	@$(RUN) ./cmd/data-engine

# 目标：安装依赖
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@$(TIDY)

# 目标：运行测试
.PHONY: test
test:
	@echo "Running tests..."
	@$(TEST) ./...

# 目标：运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@$(TEST) -coverprofile=coverage.out ./...
	@$(GO) tool cover -func=coverage.out
	@$(GO) tool cover -html=coverage.out -o coverage.html

# 目标：格式化代码
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@$(FMT) ./...

# 目标：代码质量检查
.PHONY: quality
quality:
	@echo "Running code quality checks..."
	@chmod +x ./scripts/format.sh
	@./scripts/format.sh

# 目标：清理构建产物
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# 目标：显示帮助
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all             Build the project"
	@echo "  build           Build the project"
	@echo "  run             Run the project"
	@echo "  deps            Install dependencies"
	@echo "  test            Run tests"
	@echo "  test-coverage   Run tests with coverage"
	@echo "  fmt             Format code"
	@echo "  clean           Clean build artifacts"
	@echo "  help            Show this help message"