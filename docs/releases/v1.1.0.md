## ✨ Key Features / 主要特性

- 📄 **Smart Parsing / 智能解析** - Auto-detect structure of `.docx` and `.pdf` legal documents / 自动识别 `.docx` 和 `.pdf` 法律文书
- 🎯 **Precise Extraction / 精准提取** - Extract key fields like defendant, ID, requests, and facts / 提取被告、身份证、诉求和事实
- 🐍 **Python Bridge / Python 桥接** - Advanced PDF processing with electronic seal cleaning / 高级 PDF 处理，支持电子章清洗
- 👁️ **Live Preview / 实时预览** - Preview data before extraction / 提取前预览数据
- 💾 **Multi-format Export / 多格式导出** - Support Excel (.xlsx), CSV, and JSON / 支持 Excel, CSV 和 JSON 导出
- 🔧 **OCR Support / OCR 支持** - Optional MCP OCR capability for scanned PDFs / 支持 MCP OCR 处理扫描件
- 🎨 **Modern UI / 现代界面** - Dark mode with glassmorphism design / 暗色玻璃拟态设计
- 🤖 **CI/CD Automation / 自动化流水线** - Automated multi-platform builds / 自动化多平台构建

## 📥 Downloads / 下载

### macOS

| Architecture / 架构       | Installer / 安装包 (.dmg)                                                                                                     | Archive / 压缩包 (.tar.gz)                                                                                                           |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| **Intel (x64)**           | [Download](https://github.com/can4hou6joeng4/legal-extractor/releases/download/v1.1.0/legal-extractor_1.1.0_darwin_amd64.dmg) | [Download](https://github.com/can4hou6joeng4/legal-extractor/releases/download/v1.1.0/legal-extractor_1.1.0_darwin_amd64.app.tar.gz) |
| **Apple Silicon (ARM64)** | [Download](https://github.com/can4hou6joeng4/legal-extractor/releases/download/v1.1.0/legal-extractor_1.1.0_darwin_arm64.dmg) | [Download](https://github.com/can4hou6joeng4/legal-extractor/releases/download/v1.1.0/legal-extractor_1.1.0_darwin_arm64.app.tar.gz) |

### Windows

| Architecture / 架构 | Installer / 安装程序 (.exe)                                                                                                          | Archive / 压缩包 (.zip)                                                                                                        |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
| **x64**             | [Download](https://github.com/can4hou6joeng4/legal-extractor/releases/download/v1.1.0/legal-extractor_1.1.0_windows_amd64_setup.exe) | [Download](https://github.com/can4hou6joeng4/legal-extractor/releases/download/v1.1.0/legal-extractor_1.1.0_windows_amd64.zip) |

## 🚀 Installation / 安装说明

### macOS

1. Download the `.dmg` file for your architecture / 下载对应架构的 `.dmg`。
2. Drag the app to **Applications** / 拖动到 **应用程序**。
3. **First run**: Right-click the app and select **Open** / **首次运行**: 右键点击并选择 **打开** 以跳过安全检查。

### Windows

1. Run the `_setup.exe` installer / 运行 `_setup.exe` 安装程序。

## ⚙️ OCR Configuration / OCR 配置 (Optional)

To enable OCR, create `config/conf.yaml` / 启用 OCR 请创建 `config/conf.yaml`:

```yaml
mcp:
  bin: "npx"
  args: ["-y", "@modelcontextprotocol/server-ocr"]
```

## 🆕 What's New in v1.1.0 / v1.1.0 新特性

- ✨ **Python Bridge Engine** - Intelligent PDF processing with electronic seal interference removal
- 🔄 **Smart Line Merging** - Improved text extraction with `smartMerge` algorithm
- 🏗️ **Automated Build Pipeline** - Full CI/CD integration with GitHub Actions
- 📦 **Cross-platform Packaging** - Automated builds for macOS (Intel + ARM) and Windows

---

**Full Changelog**: https://github.com/can4hou6joeng4/legal-extractor/compare/v1.0.0...v1.1.0
