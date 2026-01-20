# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Legal Document Extractor is a cross-platform desktop application that extracts structured data from Chinese legal documents (.docx and .pdf). Built with Wails 2 (Go backend + Vue 3 frontend).

## Development Commands

```bash
# Development mode with hot-reload
wails dev

# Install frontend dependencies
cd frontend && npm install

# Build for macOS (Universal Binary)
make mac-universal

# Build for Windows (requires mingw-w64)
make windows

# Build Python PDF bridge binary (required for PDF extraction)
make build-bridge

# Build all targets
make all

# Clean build artifacts
make clean
```

## Running Tests

```bash
go test ./internal/extractor/...
```

## Architecture

### Backend (Go)

- **main.go**: Wails application entry point, binds `App` struct to frontend
- **internal/app/app.go**: Frontend-facing API methods exposed via Wails bindings:
  - `SelectFile()`, `SelectOutputPath()` - native file dialogs
  - `ScanFields()`, `PreviewData()` - document analysis
  - `ExtractToPath()`, `ExportData()` - extraction and export
- **internal/extractor/extractor.go**: Core extraction logic
  - DOCX: Direct XML parsing from `word/document.xml`
  - PDF: Delegates to compiled Python binary (`pdf_extractor_core`) or falls back to Python script
- **internal/extractor/patterns.go**: Regex patterns for Chinese legal document fields (被告, 身份证号码, 诉讼请求, 事实与理由)
- **internal/extractor/export.go**: Export to CSV, JSON, Excel (.xlsx)
- **internal/mcp/client.go**: Optional MCP OCR client for scanned documents

### Frontend (Vue 3 + TypeScript)

- **frontend/src/App.vue**: Main application component
- **frontend/src/components/**: UI components (MainDropZone, PreviewTable, ResultCard, ConfigPanel)
- **frontend/wailsjs/**: Auto-generated Wails bindings for calling Go methods

### Python Bridge (PDF Extraction)

- **internal/extractor/bridge_bin/pdf_bridge.py**: Python script for PDF text extraction
- Uses pdfplumber, rapidocr_onnxruntime, pymupdf
- Compiled to `pdf_extractor_core` binary via PyInstaller for distribution

### Key Data Flow

1. Frontend calls Go methods via Wails bindings
2. `App` struct delegates to `Extractor` for document processing
3. For PDF: Go spawns Python binary/script, parses JSON response
4. For DOCX: Go parses XML directly using regex patterns
5. Results returned to frontend as `ExtractResult` struct with `Record` maps

## Extraction Fields

Fields are defined in `PatternRegistry` (patterns.go):
- `defendant` (被告)
- `idNumber` (身份证号码)
- `request` (诉讼请求)
- `factsReason` (事实与理由)
