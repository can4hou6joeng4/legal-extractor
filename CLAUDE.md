# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🛠 开发与构建命令

### 开发环境
- **启动开发模式**: `wails dev` (启动后端并带有前端热重载)
- **安装前端依赖**: `cd frontend && npm install`
- **运行 Python 提取器测试**: `python3 -m pytest internal/extractor/bridge_bin/tests/`

### 构建命令
- **全平台本地构建**: `make all` (构建 Python 桥接程序以及 macOS/Windows 二进制文件)
- **构建 Python 桥接程序**: `make build-bridge` (生成生产环境所需的 `pdf_extractor_core` 二进制文件)
- **构建特定平台**: `make mac` 或 `make windows`
- **构建脚本**: `build_mac.sh` (处理 macOS DMG 打包) 和 `build_windows.ps1` (处理 Windows 安装包)

### 测试与检查
- **运行 Go 测试**: `go test ./internal/...`
- **前端检查**: `cd frontend && npm run build` (类型检查与构建验证)

## 🏗 分支策略与流程
- **`main`**: 稳定发布分支，包含生产环境代码。
- **`develop`**: 开发主分支，所有特性开发均在此分支合并。
- **`release/v*`**: 版本发布分支（如 `release/v1.1.0`），用于最后阶段的 CI 构建、DMG/EXE 打包及发布前测试。

## 🧩 高-层级架构说明

这是一个基于 **Wails (Go + Vue 3)** 开发的法律文书信息提取工具。

### 核心设计模式
- **环境适配逻辑**: `internal/extractor/extractor.go` 实现了动态路径解析，能自动识别是处于 `wails dev` 开发环境（调用 `.py` 脚本）还是生产环境（调用打包好的 Python 二进制文件）。
- **Python 桥接 (Bridge)**: 核心 PDF 解析由 `internal/extractor/bridge_bin/pdf_bridge.py` 驱动。在生产构建中，必须先运行 `make build-bridge` 确保 Go 后端能找到对应的二进制执行文件。
- **混合提取策略**: 系统支持针对 .docx 的纯文本提取和针对 .pdf（含扫描件）的 OCR 增强提取策略。OCR 服务通过 MCP 协议（`internal/mcp/`）外接。

### CI/CD 与交付
- **GitHub Actions**: 配置文件位于 `.github/workflows/`，实现了 macOS/Windows 的自动化构建流水线。
- **打包细节**: macOS 使用 `hdiutil` 生成标准的 DMG 镜像；Windows 通过 `NSIS` 或 wails 默认工具生成安装程序。
