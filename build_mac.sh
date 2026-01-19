#!/bin/bash

# build_mac.sh
# 自动化构建 macOS 应用并捆绑 Python 环境

APP_NAME="legal-extractor"
APP_BUNDLE="build/bin/${APP_NAME}.app"

echo "🚀 开始构建 macOS 应用..."
wails build -platform darwin/arm64

if [ ! -d "$APP_BUNDLE" ]; then
    echo "❌ 构建失败：未找到应用程序包 $APP_BUNDLE"
    exit 1
fi

echo "📦 开始捆绑 Python 环境..."

# 目标资源目录
RESOURCES_DIR="${APP_BUNDLE}/Contents/Resources/bridge_bin"
mkdir -p "$RESOURCES_DIR"

# 源目录
SOURCE_DIR="internal/extractor/bridge_bin"

# 复制文件 (排除 __pycache__ 和测试文件)
# 注意：必须保留 .venv
echo "   正在复制 Python 脚本和虚拟环境..."
rsync -av --exclude='__pycache__' --exclude='tests' --exclude='*.spec' --exclude='build' --exclude='dist' "$SOURCE_DIR/" "$RESOURCES_DIR/"

echo "✅ 捆绑完成！"
echo "👉 您的应用已准备就绪：$APP_BUNDLE"
