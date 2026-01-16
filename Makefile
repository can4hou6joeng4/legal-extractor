APP_NAME := legal-extractor
BUILD_DIR := build/bin

# Default target: Build for macOS (Host) and Windows
all: mac-universal windows

# Build for macOS (Universal: amd64 + arm64)
mac: mac-universal

mac-universal:
	@echo "🍎 Building for macOS (Universal Binary)..."
	wails build -platform darwin/universal

mac-amd64:
	@echo "🍎 Building for macOS (Intel)..."
	wails build -platform darwin/amd64

mac-arm64:
	@echo "🍎 Building for macOS (Apple Silicon)..."
	wails build -platform darwin/arm64

# Build for Windows
# UIres MinGW-w64 to be installed: brew install mingw-w64
windows:
	@echo "🪟 Building for Windows (amd64)..."
	wails build -platform windows/amd64

# Build for Linux (Only works if you have gcc setup, usually fails on macOS)
linux:
	@echo "🐧 Building for Linux (amd64)..."
	@echo "⚠️  Note: Building Linux binaries on macOS is difficult. Use Docker or CI instead."
	wails build -platform linux/amd64

# Install dependencies (Homebrew required)
deps:
	@echo "🛠 Checking dependencies..."
	@if ! command -v brew >/dev/null 2>&1; then \
		echo "❌ Homebrew not found. Please install Homebrew first."; \
		exit 1; \
	fi
	@echo "📦 Installing mingw-w64 for Windows cross-compilation..."
	brew install mingw-w64
	@echo "✅ Done."

# Clean build directory
clean:
	@echo "🧹 Cleaning build directory..."
	rm -rf $(BUILD_DIR)/*
	@echo "✅ Done."

.PHONY: all mac mac-universal mac-amd64 mac-arm64 windows linux deps clean
