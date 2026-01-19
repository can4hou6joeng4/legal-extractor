<#
.SYNOPSIS
    Automated build script for Legal Extractor on Windows.
    This script sets up the Python environment and bundles it with the Wails application.

.DESCRIPTION
    1. Checks for Python installation.
    2. Creates a Python virtual environment in internal/extractor/bridge_bin/.venv
    3. Installs dependencies from requirements.txt.
    4. Builds the Wails application.
    5. Copies the bridge_bin folder (including .venv) to the build output directory.

.NOTES
    Run this script from the project root directory in PowerShell.
#>

$ErrorActionPreference = "Stop"

Write-Host "🚀 Starting Legal Extractor Windows Build..." -ForegroundColor Cyan

# 1. Check Python
Write-Host "🔍 Checking for Python..."
try {
    $pythonVersion = python --version 2>&1
    Write-Host "   Found: $pythonVersion" -ForegroundColor Green
} catch {
    Write-Error "❌ Python not found. Please install Python 3.9+ and add it to your PATH."
}

# 2. Setup Python Environment
$bridgeDir = "internal\extractor\bridge_bin"
$venvDir = "$bridgeDir\.venv"
$reqFile = "$bridgeDir\requirements.txt"

Write-Host "🛠  Setting up Python environment in $bridgeDir..."

if (-not (Test-Path $venvDir)) {
    Write-Host "   Creating virtual environment..."
    python -m venv $venvDir
} else {
    Write-Host "   Virtual environment already exists."
}

# Install dependencies
Write-Host "📦 Installing dependencies..."
$pipPath = "$venvDir\Scripts\pip.exe"
& $pipPath install -r $reqFile
if ($LASTEXITCODE -ne 0) {
    Write-Error "❌ Failed to install Python dependencies."
}

# 3. Build Wails App
Write-Host "🔨 Building Wails application..."
wails build -platform windows/amd64

if ($LASTEXITCODE -ne 0) {
    Write-Error "❌ Wails build failed."
}

# 4. Bundle Assets
$buildBinDir = "build\bin"
$destBridgeDir = "$buildBinDir\bridge_bin"

Write-Host "📂 Bundling Python bridge to $destBridgeDir..."

# Clean previous bundle if exists
if (Test-Path $destBridgeDir) {
    Remove-Item -Path $destBridgeDir -Recurse -Force
}

# Create destination directory
New-Item -ItemType Directory -Force -Path $destBridgeDir | Out-Null

# Copy files
# We need to exclude __pycache__ and maybe other junk, but for now simple copy is fine
# PowerShell Copy-Item -Recurse is simple but doesn't have easy excludes. 
# We'll copy everything and then clean up.

Copy-Item -Path "$bridgeDir\*" -Destination $destBridgeDir -Recurse

# Cleanup unnecessary files in destination
Write-Host "🧹 Cleaning up bundled files..."
Get-ChildItem -Path $destBridgeDir -Recurse -Filter "__pycache__" | Remove-Item -Recurse -Force
Get-ChildItem -Path $destBridgeDir -Recurse -Filter "*.pyc" | Remove-Item -Recurse -Force

Write-Host "✅ Build Complete!" -ForegroundColor Green
Write-Host "👉 Executable: $buildBinDir\legal-extractor.exe" -ForegroundColor Green
Write-Host "👉 Bridge:     $destBridgeDir" -ForegroundColor Gray
