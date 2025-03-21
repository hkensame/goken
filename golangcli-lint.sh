#!/bin/bash

# 定义项目根目录
PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# 确保 golangci-lint 已安装
if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint not found. Installing..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

# 在项目根目录运行 lint
cd "$PROJECT_ROOT"
echo "Running golangci-lint on all packages..."

# 运行 golangci-lint 并捕获错误，但不退出
golangci-lint run --disable errcheck ./...
if [ $? -ne 0 ]; then
    echo "Lint check completed with errors."
else
    echo "Lint check completed successfully."
fi