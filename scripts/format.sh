#!/bin/bash

echo "Running code formatting and quality checks..."

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.20+ first."
    exit 1
fi

# 格式化代码
echo "Formatting code with go fmt..."
go fmt ./...

# 运行静态分析
echo "Running static analysis with go vet..."
go vet ./...

# 检查依赖
echo "Checking dependencies..."
go mod tidy

echo "Code formatting and quality checks completed successfully!"