<p align="center">
  <img src="build/appicon.png" alt="Legal Extractor Logo" width="120" height="120">
</p>

<h1 align="center">Legal Document Extractor</h1>

<p align="center">
  <strong>从法律文书中智能提取关键信息，一键导出为结构化数据</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat-square&logo=vue.js" alt="Vue Version">
  <img src="https://img.shields.io/badge/Wails-2.x-DF0000?style=flat-square" alt="Wails Version">
  <img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Windows-blue?style=flat-square" alt="Platform">
</p>

---

## ✨ 功能特性

- 📄 **智能解析** - 自动识别 `.docx` 格式的法律文书结构
- 🎯 **精准提取** - 提取被告、身份证号码、诉讼请求、事实与理由等关键字段
- 👁️ **实时预览** - 提取前可预览数据，确保准确性
- 💾 **一键导出** - 生成标准 CSV 文件，可直接用 Excel 打开
- 🖥️ **跨平台** - 支持 macOS 和 Windows 系统
- 🎨 **现代界面** - 暗色主题 + 玻璃拟态设计

---

## 📸 界面预览

<p align="center">
  <em>现代化暗色主题界面，简洁直观的操作流程</em>
</p>

---

## 🚀 快速开始

### 下载运行

1. 从 [Releases](./build/bin) 下载对应平台的安装包
2. **macOS**: 双击 `legal-extractor.app` 运行
3. **Windows**: 双击 `legal-extractor.exe` 运行

### 使用步骤

1. 点击 **「选择 .docx 文件」** 按钮，选择法律文书
2. 点击 **「预览数据」** 查看提取结果（可选）
3. 点击 **「提取并保存」** 导出 CSV 文件

---

## 🛠️ 开发指南

### 环境要求

- Go 1.21+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### 安装依赖

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 克隆项目
cd /path/to/legal-extractor

# 安装前端依赖
cd frontend && npm install && cd ..
```

### 开发模式

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

## 📁 项目结构

```
legal-extractor/
├── main.go              # 应用入口
├── app.go               # 后端 API 绑定层
├── wails.json           # Wails 配置
│
├── pkg/extractor/       # 核心业务逻辑
│   └── extractor.go     # 文档解析 & CSV 导出
│
├── frontend/            # Vue 3 前端
│   ├── src/
│   │   ├── App.vue      # 主界面组件
│   │   └── style.css    # 全局样式
│   └── wailsjs/         # Wails 自动生成的 TS 绑定
│
└── build/               # 构建资源 & 产物
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

## 📝 提取字段说明

| 字段           | 来源                             | 匹配规则                          |
| :------------- | :------------------------------- | :-------------------------------- |
| **被告**       | `被告:` 后的姓名                 | 截取至 `性别`/`身份证` 等关键词前 |
| **身份证号码** | `身份证号码:` 后的数字串         | 18 位数字+X                       |
| **诉讼请求**   | `诉讼请求:` 至 `事实与理由` 之间 | 多行文本                          |
| **事实与理由** | `事实与理由:` 至 `此致` 之间     | 多行文本                          |

---

## 📄 License

MIT License © 2026

---

<p align="center">
  <sub>Made with ❤️ using <a href="https://wails.io">Wails</a></sub>
</p>
