<p align="center">
  <img src="build/appicon.png" alt="Legal Extractor Logo" width="120" height="120">
</p>

<h1 align="center">Legal Document Extractor / 法律文书提取器</h1>

<p align="center">
  <strong>Intelligent information extraction from legal documents with one-click structured export</strong><br>
  <strong>从法律文书中智能提取关键信息，一键导出为结构化数据</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat-square&logo=vue.js" alt="Vue Version">
  <img src="https://img.shields.io/badge/Wails-2.x-DF0000?style=flat-square" alt="Wails Version">
  <img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Windows-blue?style=flat-square" alt="Platform">
</p>

---

## ✨ Features / 功能特性

- 📄 **Smart Parsing / 智能解析** - Auto-detect structure of `.docx` and `.pdf` legal documents / 自动识别 `.docx` 和 `.pdf` 格式的法律文书结构
- 🎯 **Precise Extraction / 精准提取** - Extract key fields like defendant, ID, requests, and facts / 提取被告、身份证号码、诉讼请求、事实与理由等关键字段
- 👁️ **Live Preview / 实时预览** - Preview data before extraction to ensure accuracy / 提取前可预览数据，确保准确性
- 💾 **Multi-format Export / 多格式导出** - Support Excel (.xlsx), CSV, and JSON / 支持 Excel (.xlsx), CSV, JSON 格式导出
- 🖥️ **Cross-platform / 跨平台** - Native support for macOS and Windows / 支持 macOS 和 Windows 系统
- 🎨 **Modern UI / 现代界面** - Dark theme with Glassmorphism design / 暗色主题 + 玻璃拟态设计
- 🔧 **OCR Support / OCR 支持** - Optional MCP OCR for scanned documents / 支持通过 MCP 集成 OCR 处理扫描件

---

## 📸 界面预览

<p align="center">
  <em>现代化暗色主题界面，简洁直观的操作流程</em>
</p>

---

## 🚀 Quick Start / 快速开始

### Download / 下载运行

1. Download the installer for your platform from [Releases](https://github.com/can4hou6joeng4/legal-extractor/releases)
   从 [Releases](https://github.com/can4hou6joeng4/legal-extractor/releases) 下载对应平台的安装包
2. **macOS**: Drag `legal-extractor.app` to Applications / 将应用拖入应用程序文件夹
3. **Windows**: Run `legal-extractor_setup.exe` / 运行安装程序程序

### Usage / 使用步骤

1. Click **"Select Files"** to choose documents / 点击 **“选择文件”** 选择法律文书
2. Click **"Preview"** to verify data (Optional) / 点击 **“预览”** 查看提取结果（可选）
3. Click **"Extract & Save"** to export / 点击 **“提取并保存”** 导出文件

---

## 🛠️ Development / 开发指南

### Prerequisites / 环境要求

- Go 1.21+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### Setup / 安装依赖

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone project / 克隆项目
git clone https://github.com/can4hou6joeng4/legal-extractor.git
cd legal-extractor

# Install dependencies / 安装前端依赖
cd frontend && npm install && cd ..
```

### Dev Mode / 开发模式

```bash
wails dev
```

启动后会自动打开应用窗口，支持热重载。

### 构建发布

```bash
# 构建当前平台
wails build

# 构建 Windows 版本 (需要交叉编译环境)
wails build -platform windows/amd64

# 构建 macOS 版本
wails build -platform darwin/amd64
```

构建产物位于 `build/bin/` 目录。

---

## ⚙️ OCR Configuration / OCR 配置 (Optional)

This project supports OCR via [Model Context Protocol (MCP)](https://modelcontextprotocol.io/).
本项目支持通过 MCP 集成 OCR 能力。

Create `config/conf.yaml` in the root directory / 在根目录创建 `config/conf.yaml`：

```yaml
mcp:
  bin: "npx"
  args:
    - "-y"
    - "@modelcontextprotocol/server-ocr"
```

**说明**:

- 如果未配置或配置无效，将自动回退到原生文本提取模式。
- 确保运行环境已安装配置中指定的依赖（如 Node.js/npx）。
- 支持通过环境变量 `LEGAL_EXTRACTOR_CONFIG` 指定配置文件路径。

---

## 📁 Project Structure / 项目结构

```
legal-extractor/
├── main.go              # 应用入口
├── wails.json           # Wails 配置
│
├── internal/            # Core logic (重构后的核心逻辑)
│   ├── app/             # Backend API bindings
│   ├── config/          # Configuration management
│   ├── extractor/       # Extraction & Export engines
│   └── mcp/             # OCR Client
│
├── config/              # 配置文件
│   └── conf.yaml
│
├── frontend/            # Vue 3 Frontend (前端组件)
│   ├── src/
│   │   ├── App.vue      # 主界面组件
│   │   └── style.css    # 全局样式
│   └── wailsjs/         # Wails 自动生成的 TS 绑定
│
└── build/               # Build assets & installers
    ├── appicon.png      # 应用图标
    └── bin/             # 可执行文件
```

---

## 🔧 技术栈

| 层级         | 技术                                      |
| :----------- | :---------------------------------------- |
| **后端**     | Go 1.21+                                  |
| **前端**     | Vue 3 + TypeScript + Vite                 |
| **桌面框架** | Wails 2                                   |
| **文档解析** | Go 标准库 (`archive/zip`, `encoding/xml`) |
| **UI 风格**  | 暗色主题 + Glassmorphism                  |

---

## 📝 Extraction Fields / 提取字段

| Field / 字段            | Rule / 匹配规则                                            |
| :---------------------- | :--------------------------------------------------------- |
| **Defendant / 被告**    | Extracted from text after "被告:" / 从 "被告:" 后提取      |
| **ID / 身份证**         | 18-digit ID number patterns / 自动识别 18 位身份证号       |
| **Requests / 诉讼请求** | Content between "诉讼请求" and "事实与理由" / 诉讼请求段落 |
| **Facts / 事实与理由**  | Content between "事实与理由" and "此致" / 事实与理由段落   |

---

## 📄 License

MIT License © 2026

---

<p align="center">
  <sub>Made with ❤️ using <a href="https://wails.io">Wails</a></sub>
</p>
