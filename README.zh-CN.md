<p align="center">
  <img src="build/appicon.png" alt="Legal Extractor Logo" width="120" height="120">
</p>

<h1 align="center">法律文书提取器 (Legal Document Extractor) v3.0.0</h1>

<p align="center">
  <strong>新一代法律文书智能提取工具，基于高性能 OCR 与并行解析引擎</strong>
</p>

<p align="center">
  <a href="README.md">English</a> | 简体中文
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Version-3.0.0-blue?style=flat-square" alt="Version">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat-square&logo=vue.js" alt="Vue Version">
  <img src="https://img.shields.io/badge/Wails-2.x-DF0000?style=flat-square" alt="Wails Version">
  <img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Windows%20%7C%20Docker-blue?style=flat-square" alt="Platform">
</p>

---

## ✨ 功能特性

- 🚀 **v3.0 新一代引擎** - 全面接入百度 AI Studio 高性能模型，识别精度大幅提升。
- 📄 **智能解析** - 自动识别 `.docx` 和 `.pdf` 格式的法律文书结构。
- ⚡ **并行提取** - 引入 Go 协程并发处理，本地文本解析速度提升 **300%**。
- 🎯 **精准提取** - 提取被告、身份证号码、诉讼请求、事实与理由等关键字段。
- 🧩 **物理切片** - 支持 50 页以上超长 PDF 文档的自动分片识别。
- 👁️ **实时预览** - 提取前可预览数据，确保准确性。
- 💾 **多格式导出** - 支持 Excel (.xlsx), CSV, JSON 格式导出。
- 🐳 **Docker 支持** - 内置 Docker 镜像，支持一键私有化部署。

---

## 🚀 快速开始

### 🅰️ 桌面版 (推荐个人用户)

1. 从 [Releases](https://github.com/can4hou6joeng4/legal-extractor/releases) 下载对应平台的安装包
2. **macOS**: 双击 `legal-extractor.dmg` 安装并拖入应用程序文件夹
3. **Windows**: 运行 `legal-extractor_setup.exe` 安装程序

### 🅱️ Web 版 (推荐团队/服务器)

使用 Docker 立即启动 Web 版本：

```bash
# 1. 设置百度 API 密钥 (处理 PDF/图片必须)
export BAIDU_API_KEY="您的API_KEY"
export BAIDU_SECRET_KEY="您的SECRET_KEY"

# 2. 使用 Docker Compose 启动
docker-compose up -d

# 3. 访问浏览器
# http://localhost:8080
```

### 使用步骤

1. 点击 **“选择文件”** 按钮，选择法律文书
2. 点击 **“预览”** 查看提取结果（可选）
3. 点击 **“提取并保存”** 导出结构化数据

---

## 🛠️ 开发指南

### 环境要求

- Go 1.24+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) (仅桌面版开发需要)
- Docker & Docker Compose (仅 Web 版开发需要)

### 安装依赖

```bash
# 克隆项目
git clone https://github.com/can4hou6joeng4/legal-extractor.git
cd legal-extractor

# 安装依赖
cd frontend && npm install && cd ..
```

### 开发模式

#### 桌面版 (Wails)
```bash
wails dev
```

#### Web 版 (前后端联调)
支持全栈热重载开发：

1. **启动后端 (Go)**
   ```bash
   # 安装 Air 热加载工具
   go install github.com/air-verse/air@latest

   # 启动服务
   air
   ```

2. **启动前端 (Vite)**
   ```bash
   cd frontend
   npm run dev
   ```
   打开 http://localhost:5173 (API 请求会自动代理到后端)

---

## ⚙️ 配置说明

### 百度 OCR (PDF/图片必须)

本项目接入了百度 AI Studio (PaddleOCR-VL) 大模型，用于处理复杂的 PDF 和扫描件。

📖 **[点击查看详细配置指南](docs/user/CONFIG_GUIDE.md)**

**方式 1: 环境变量**
- `BAIDU_TOKEN` (百度云访问令牌)

**方式 2: 配置文件**
在 `internal/config/baked_conf.yaml` 中配置：
```yaml
baidu:
  token: "您的百度Token"
```

---

## 📁 项目结构

```
legal-extractor/
├── cmd/
│   └── server/          # Web 服务入口 (REST API)
├── internal/            # 核心业务逻辑
│   ├── app/             # 桌面端逻辑 (Wails 绑定)
│   ├── config/          # 配置管理
│   ├── extractor/       # 提取引擎 (PDF/DOCX/OCR)
├── frontend/            # Vue 3 前端 (自适应 UI)
│   ├── src/services/    # API 适配层 (Web/Desktop)
├── build/               # 构建资源与安装程序配置
├── Dockerfile           # Web 版构建文件
├── docker-compose.yml   # Docker 编排配置
└── README.md
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

[MIT License](LICENSE) © 2026

---

<p align="center">
  <sub>Made with ❤️ using <a href="https://wails.io">Wails</a> & <a href="https://vuejs.org/">Vue 3</a></sub>
</p>
