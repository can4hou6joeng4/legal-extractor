## ✨ Key Features

- 📄 **Smart Parsing** - Auto-detect structure of `.docx` and `.pdf` legal documents
- 🎯 **Precise Extraction** - Extract key fields like defendant, ID number, requests, and facts
- 👁️ **Live Preview** - Preview data before extraction
- 💾 **Multi-format Export** - Support Excel (.xlsx), CSV, and JSON
- 🔧 **OCR Support** - Optional MCP OCR capability for scanned PDFs
- 🎨 **Modern UI** - Dark mode with glassmorphism design

## 📥 Downloads

### macOS

| Architecture | Installer (.dmg) | Archive (.tar.gz) |
|---|---|---|
| **Intel (x64)** | [Download (4.8 MB)](legal-extractor_1.0.0_darwin_amd64.dmg) | [Download (5.1 MB)](legal-extractor_1.0.0_darwin_amd64.app.tar.gz) |
| **Apple Silicon (ARM64)** | [Download (4.8 MB)](legal-extractor_1.0.0_darwin_arm64.dmg) | [Download (4.8 MB)](legal-extractor_1.0.0_darwin_arm64.app.tar.gz) |

### Windows

| Architecture | Installer (.exe) | Archive (.zip) |
|---|---|---|
| **x64** | [Download (7.3 MB)](legal-extractor_1.0.0_windows_amd64_setup.exe) | [Download (5.6 MB)](legal-extractor_1.0.0_windows_amd64.zip) |

## 🚀 Installation

### macOS (Recommended)
1. Download the `.dmg` file for your architecture.
2. Open it and drag the app to your **Applications** folder.
3. **First run:** Right-click the app and select **Open** to bypass Gatekeeper.

### Windows
1. Download the `_setup.exe` file.
2. Run the installer and follow the instructions.

## ⚙️ OCR Configuration (Optional)

To enable OCR, create `config/conf.yaml`:

```yaml
mcp:
  bin: "npx"
  args: ["-y", "@modelcontextprotocol/server-ocr"]
```

---
**Full Changelog**: https://github.com/can4hou6joeng4/legal-extractor/commits/v1.0.0
