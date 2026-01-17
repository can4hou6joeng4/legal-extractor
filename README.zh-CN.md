<p align="center">
  <img src="build/appicon.png" alt="Legal Extractor Logo" width="120" height="120">
</p>

<h1 align="center">法律文书提取器 (Legal Document Extractor)</h1>

<p align="center">
  <strong>从法律文书中智能提取关键信息，一键导出为结构化数据</strong>
</p>

<p align="center">
  <a href="README.md">English</a> | 简体中文
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat-square&logo=vue.js" alt="Vue Version">
  <img src="https://img.shields.io/badge/Wails-2.x-DF0000?style=flat-square" alt="Wails Version">
  <img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Windows-blue?style=flat-square" alt="Platform">
</p>

---

## ✨ 功能特性

- 📄 **智能解析** - 自动识别 `.docx` 和 `.pdf` 格式的法律文书结构
- 🎯 **精准提取** - 提取被告、身份证号码、诉讼请求、事实与理由等关键字段
- 👁️ **实时预览** - 提取前可预览数据，确保准确性
- 💾 **多格式导出** - 支持 Excel (.xlsx), CSV, JSON 格式导出
- 🖥️ **跨平台** - 原生支持 macOS 和 Windows 系统
- 🎨 **现代界面** - 暗色主题 + 玻璃拟态设计
- 🔧 **OCR 支持** - 支持通过 MCP 集成 OCR 处理扫描件

---

## 🚀 快速开始

### 下载运行

1. 从 [Releases](https://github.com/can4hou6joeng4/legal-extractor/releases) 下载对应平台的安装包
2. **macOS**: 双击 `legal-extractor.dmg` 安装并拖入应用程序文件夹
3. **Windows**: 运行 `legal-extractor_setup.exe` 安装程序

### 使用步骤

1. 点击 **“选择文件”** 按钮，选择法律文书
2. 点击 **“预览”** 查看提取结果（可选）
3. 点击 **“提取并保存”** 导出结构化数据

---

## 🛠️ 开发指南

### 环境要求

- Go 1.21+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### 安装依赖

```bash
# 克隆项目
git clone https://github.com/can4hou6joeng4/legal-extractor.git
cd legal-extractor

# 安装前端依赖
cd frontend && npm install && cd ..
```

### 开发模式

```bash
wails dev
```

---

## ⚙️ OCR 配置 (可选)

本项目支持通过 [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) 集成 OCR 能力。

在项目根目录下创建 `config/conf.yaml` 文件：

```yaml
mcp:
  bin: "npx"
  args:
    - "-y"
    - "@modelcontextprotocol/server-ocr"
```

---

## 📁 项目结构

```
legal-extractor/
├── internal/            # 后端核心逻辑
│   ├── app/             # API 绑定层
│   ├── config/          # 配置管理
│   ├── extractor/       # 提取与导出引擎
│   └── mcp/             # OCR 客户端
├── frontend/            # Vue 3 前端代码
├── build/               # 构建资源与安装程序配置
└── RELEASE_NOTES_v1.0.0.md
```

---

## 📝 提取字段说明

| 字段           | 匹配规则                                   |
| :------------- | :----------------------------------------- |
| **被告**       | 从 "被告:" 关键词后提取姓名                |
| **身份证号码** | 自动识别 18 位身份证号码模式               |
| **诉讼请求**   | 提取 "诉讼请求" 与 "事实与理由" 之间的内容 |
| **事实与理由** | 提取 "事实与理由" 与 "此致" 之间的内容     |

---

## 📄 开源协议

MIT License © 2026

---

<p align="center">
  <sub>Made with ❤️ using <a href="https://wails.io">Wails</a></sub>
</p>
