<p align="center">
  <img src="build/appicon.png" alt="Legal Extractor Logo" width="120" height="120">
</p>

<h1 align="center">Legal Document Extractor</h1>

<p align="center">
  <strong>Intelligent information extraction from legal documents with one-click structured export</strong>
</p>

<p align="center">
  English | <a href="README.zh-CN.md">简体中文</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat-square&logo=vue.js" alt="Vue Version">
  <img src="https://img.shields.io/badge/Wails-2.x-DF0000?style=flat-square" alt="Wails Version">
  <img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Windows-blue?style=flat-square" alt="Platform">
</p>

---

## ✨ Features

- 📄 **Smart Parsing** - Auto-detect structure of `.docx` and `.pdf` legal documents
- 🎯 **Precise Extraction** - Extract key fields like defendant, ID, requests, and facts
- 👁️ **Live Preview** - Preview data before extraction to ensure accuracy
- 💾 **Multi-format Export** - Support Excel (.xlsx), CSV, and JSON
- 🖥️ **Cross-platform** - Native support for macOS and Windows
- 🎨 **Modern UI** - Dark theme with Glassmorphism design
- 🔧 **OCR Support** - Optional MCP OCR for scanned documents

---

## 🚀 Quick Start

### 🅰️ Desktop Version (Recommended)

1. Download the installer for your platform from [Releases](https://github.com/can4hou6joeng4/legal-extractor/releases)
2. **macOS**: Download `.dmg`, open and drag to Applications
3. **Windows**: Run `legal-extractor_setup.exe` installer

### 🅱️ Web Version (Docker)

Run the following command to start the Web version instantly:

```bash
# 1. Set your Baidu API Key (Required for PDF/Image OCR)
export BAIDU_API_KEY="your_api_key"
export BAIDU_SECRET_KEY="your_secret_key"

# 2. Start with Docker Compose
docker-compose up -d

# 3. Access in Browser
# http://localhost:8080
```

### Usage

1. Click **"Select Files"** to choose legal documents
2. Click **"Preview"** to verify extracted data (Optional)
3. Click **"Extract & Save"** to export structured data

---

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### Setup

```bash
# Clone project
git clone https://github.com/can4hou6joeng4/legal-extractor.git
cd legal-extractor

# Install dependencies
cd frontend && npm install && cd ..
```

### Dev Mode

```bash
wails dev
```

---

## ⚙️ OCR Configuration (Optional)

This project supports OCR via [Model Context Protocol (MCP)](https://modelcontextprotocol.io/).

Create `config/conf.yaml` in the root directory:

```yaml
mcp:
  bin: "npx"
  args:
    - "-y"
    - "@modelcontextprotocol/server-ocr"
```

---

## 📁 Project Structure

```
legal-extractor/
├── internal/            # Core logic
│   ├── app/             # Backend API bindings
│   ├── config/          # Configuration management
│   ├── extractor/       # Extraction & Export engines
│   └── mcp/             # OCR Client
├── frontend/            # Vue 3 Frontend
├── build/               # Build assets & installers
└── RELEASE_NOTES_v1.0.0.md
```

---

## 📝 Extraction Fields

| Field         | Rule                                           |
| :------------ | :--------------------------------------------- |
| **Defendant** | Extracted from text after "被告:" (Defendant:) |
| **ID Number** | 18-digit ID number patterns                    |
| **Requests**  | Content between "诉讼请求" and "事实与理由"    |
| **Facts**     | Content between "事实与理由" and "此致"        |

---

## 📄 License

[MIT License](LICENSE) © 2026

---

<p align="center">
  <sub>Made with ❤️ using <a href="https://wails.io">Wails</a></sub>
</p>
