package extractor

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"

	"encoding/json"
	"os"
	"os/exec"
	"runtime"
)

// Extractor handles document extraction logic
type Extractor struct {
	logger *slog.Logger
}

// NewExtractor creates a new Extractor instance
func NewExtractor(logger *slog.Logger) *Extractor {
	if logger == nil {
		logger = slog.Default()
	}
	return &Extractor{
		logger: logger,
	}
}

// Record represents a single extracted case as a flexible map
type Record map[string]string

// PythonBridgeResponse represents the JSON response from Python script
type PythonBridgeResponse struct {
	Path      string   `json:"path"`
	Records   []Record `json:"records"`
	Count     int      `json:"count"`
	Status    string   `json:"status"`
	Error     string   `json:"error,omitempty"`
	IsOCRUsed bool     `json:"is_ocr_used"`
}

// ExtractData extracts records from a file
func (e *Extractor) ExtractData(inputFile string, fields []string) ([]Record, error) {
	ext := strings.ToLower(filepath.Ext(inputFile))

	switch ext {
	case ".pdf":
		return e.extractFromPDF(inputFile)
	case ".docx":
		// 对于 DOCX，仍使用原有逻辑
		text, err := extractTextFromDocx(inputFile)
		if err != nil {
			return nil, fmt.Errorf("error extracting text from docx: %w", err)
		}
		if len(fields) == 0 {
			for k := range PatternRegistry {
				fields = append(fields, k)
			}
		}
		return e.parseCases(text, fields), nil
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func extractTextFromDocx(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var documentXML io.ReadCloser
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			documentXML, err = f.Open()
			if err != nil {
				return "", err
			}
			break
		}
	}

	if documentXML == nil {
		return "", fmt.Errorf("word/document.xml not found")
	}
	defer documentXML.Close()

	decoder := xml.NewDecoder(documentXML)
	var sb strings.Builder

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "t" {
				var s string
				if err := decoder.DecodeElement(&s, &se); err == nil {
					sb.WriteString(s)
				}
			}
		case xml.EndElement:
			if se.Name.Local == "p" {
				sb.WriteString("\n")
			}
		}
	}

	return sb.String(), nil
}

// getBinaryPath 查找编译好的 pdf_extractor_core 二进制
// 返回二进制路径，如果找不到则返回空字符串
func (e *Extractor) getBinaryPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	exeDir := filepath.Dir(exePath)
	cwd, _ := os.Getwd()

	// 根据操作系统确定二进制文件名
	binaryName := "pdf_extractor_core"
	if runtime.GOOS == "windows" {
		binaryName = "pdf_extractor_core.exe"
	}

	// 检查的路径列表（按优先级排序）
	searchPaths := []string{
		// macOS App Bundle: Contents/Resources/bridge_bin/
		filepath.Join(exeDir, "..", "Resources", "bridge_bin", binaryName),
		// Windows/Linux 扁平结构: bridge_bin/
		filepath.Join(exeDir, "bridge_bin", binaryName),
		// 开发模式: internal/extractor/bridge_bin/
		filepath.Join(cwd, "internal", "extractor", "bridge_bin", binaryName),
	}

	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			e.logger.Info("Found compiled binary", "path", p)
			return p
		}
	}

	return ""
}

// getBridgePaths returns the python executable and script paths depending on the environment
// This is used as a fallback when the compiled binary is not available (development mode)
func (e *Extractor) getBridgePaths() (string, string, error) {
	// 1. Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return "", "", fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	// Define possible locations for the bridge_bin directory
	// Priority 1: macOS App Bundle Resources (Production)
	// Structure: legal-extractor.app/Contents/MacOS/legal-extractor -> ../Resources/bridge_bin
	prodResourcePath := filepath.Join(exeDir, "..", "Resources", "bridge_bin")

	// Priority 2: Relative to CWD (Development / Source)
	// Structure: ./internal/extractor/bridge_bin
	cwd, _ := os.Getwd()
	devResourcePath := filepath.Join(cwd, "internal", "extractor", "bridge_bin")

	var bridgeDir string

	// Check if we are running inside an App Bundle with Resources
	if _, err := os.Stat(filepath.Join(prodResourcePath, "pdf_bridge.py")); err == nil {
		e.logger.Info("Using Production Bridge Path", "path", prodResourcePath)
		bridgeDir = prodResourcePath
	} else if _, err := os.Stat(filepath.Join(devResourcePath, "pdf_bridge.py")); err == nil {
		e.logger.Info("Using Development Bridge Path", "path", devResourcePath)
		bridgeDir = devResourcePath
	} else {
		// Fallback: Check relative to executable directly (Linux/Windows flat binary)
		flatPath := filepath.Join(exeDir, "bridge_bin")
		if _, err := os.Stat(filepath.Join(flatPath, "pdf_bridge.py")); err == nil {
			e.logger.Info("Using Flat Binary Bridge Path", "path", flatPath)
			bridgeDir = flatPath
		} else {
			return "", "", fmt.Errorf("bridge directory not found. Checked: %s, %s, %s", prodResourcePath, devResourcePath, flatPath)
		}
	}

	scriptPath := filepath.Join(bridgeDir, "pdf_bridge.py")
	pythonPath := filepath.Join(bridgeDir, ".venv", "bin", "python3")

	// On Windows, python executable might be in Scripts/python.exe
	if runtime.GOOS == "windows" {
		pythonPath = filepath.Join(bridgeDir, ".venv", "Scripts", "python.exe")
	}

	return pythonPath, scriptPath, nil
}

// extractFromPDF 使用 Python Bridge 提取 PDF 字段
// 优先使用编译好的 pdf_extractor_core 二进制，如果不存在则 fallback 到 Python 脚本
func (e *Extractor) extractFromPDF(path string) ([]Record, error) {
	var cmd *exec.Cmd

	// 优先尝试使用编译好的二进制
	binaryPath := e.getBinaryPath()
	if binaryPath != "" {
		e.logger.Info("Extracting PDF using compiled binary", "binary", binaryPath)
		cmd = exec.Command(binaryPath, path)
	} else {
		// Fallback: 使用 Python 脚本（开发模式或二进制缺失）
		e.logger.Info("Compiled binary not found, falling back to Python script")
		pythonPath, scriptPath, err := e.getBridgePaths()
		if err != nil {
			return nil, fmt.Errorf("no extraction method available: %w", err)
		}
		e.logger.Info("Extracting PDF using Python Bridge", "script", scriptPath, "interpreter", pythonPath)
		cmd = exec.Command(pythonPath, scriptPath, path)
	}

	output, err := cmd.Output()
	if err != nil {
		// 尝试获取标准错误输出
		if exitErr, ok := err.(*exec.ExitError); ok {
			e.logger.Error("Extraction execution failed", "stderr", string(exitErr.Stderr))
			return nil, fmt.Errorf("extraction failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute extraction: %w", err)
	}

	// 解析 JSON 响应
	var response PythonBridgeResponse
	if err := json.Unmarshal(output, &response); err != nil {
		e.logger.Error("Failed to parse response", "output", string(output))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查状态
	if response.Status != "success" {
		return nil, fmt.Errorf("extraction failed: %s", response.Error)
	}

	e.logger.Info("Successfully extracted PDF fields", "count", response.Count)
	return response.Records, nil
}

// ScanResult represents the response from quick scan
type ScanResult struct {
	Status string   `json:"status"`
	Keys   []string `json:"keys"`
	Error  string   `json:"error,omitempty"`
}

// ScanFields performs a lightweight scan of the file to determine available fields
// It avoids full extraction and OCR
func (e *Extractor) ScanFields(inputFile string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(inputFile))

	// For DOCX, scanning is fast enough with regular extraction (xml parsing is cheap)
	if ext == ".docx" {
		text, err := extractTextFromDocx(inputFile)
		if err != nil {
			return nil, err
		}
		var keys []string
		// Check for presence of keywords
		if strings.Contains(text, "被告") { keys = append(keys, "defendant") }
		if strings.Contains(text, "身份证") || regexp.MustCompile(`\d{18}`).MatchString(text) { keys = append(keys, "idNumber") }
		if strings.Contains(text, "诉讼请求") { keys = append(keys, "request") }
		if strings.Contains(text, "事实与理由") { keys = append(keys, "factsReason") }

		// If nothing found, default to all
		if len(keys) == 0 {
			return []string{"defendant", "idNumber", "request", "factsReason"}, nil
		}
		return keys, nil
	}

	// For PDF, use the new python quick scan mode
	if ext == ".pdf" {
		var cmd *exec.Cmd
		binaryPath := e.getBinaryPath()

		if binaryPath != "" {
			cmd = exec.Command(binaryPath, "--scan", inputFile)
		} else {
			pythonPath, scriptPath, err := e.getBridgePaths()
			if err != nil {
				return nil, err
			}
			cmd = exec.Command(pythonPath, scriptPath, "--scan", inputFile)
		}

		output, err := cmd.Output()
		if err != nil {
			e.logger.Error("Scan failed", "error", err)
			// Fallback to all fields on error
			return []string{"defendant", "idNumber", "request", "factsReason"}, nil
		}

		var res ScanResult
		if err := json.Unmarshal(output, &res); err != nil {
			return nil, fmt.Errorf("failed to parse scan result: %w", err)
		}

		return res.Keys, nil
	}

	return nil, fmt.Errorf("unsupported file type")
}

func (e *Extractor) parseCases(text string, fields []string) []Record {
	parts := DefaultPatterns.Split.Split(text, -1)

	var data []Record

	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}

		record := make(Record)

		// Create a quick lookup for selected fields
		fieldSet := make(map[string]bool)
		for _, f := range fields {
			fieldSet[f] = true
		}

		// 1. Extract Defendant (Commonly used as primary identifier)
		if fieldSet["defendant"] {
			loc := DefaultPatterns.DefStart.FindStringIndex(part)
			if loc != nil {
				startIdx := loc[1]
				remaining := part[startIdx:]

				// 先移除所有换行和多余空格，获取一个连续的文本段
				// 这可以处理PDF中每个字符间有换行的情况
				cleanRemaining := strings.ReplaceAll(remaining, "\n", "")
				cleanRemaining = strings.ReplaceAll(cleanRemaining, "\r", "")

				// 在清洗后的文本中查找结束位置
				locEnd := DefaultPatterns.DefEnd.FindStringIndex(cleanRemaining)

				var name string
				if locEnd != nil {
					name = cleanRemaining[:locEnd[0]]
				} else {
					// 如果没找到结束标记，尝试取前面一段（假设姓名不会超过50个字符）
					if len(cleanRemaining) > 50 {
						name = cleanRemaining[:50]
					} else {
						name = cleanRemaining
					}
					// 尝试在这段文本中找到第一个非姓名字符
					for i, r := range name {
						if r == '性' || r == '男' || r == '女' || r == '生' || r == '住' || r == '联' {
							name = name[:i]
							break
						}
					}
				}

				// 清洗提取的姓名
				name = strings.Trim(name, " ,，、:：；;\t")
				// 移除可能的干扰词（如"被告"重复）
				name = strings.TrimPrefix(name, "被告")
				name = strings.TrimSpace(name)
				record["defendant"] = name
			} else {
				match := DefaultPatterns.DefFallback.FindStringSubmatch(part)
				if len(match) > 1 {
					record["defendant"] = strings.TrimSpace(match[1])
				}
			}
		}

		// 2. Extract ID
		if fieldSet["idNumber"] {
			matchID := DefaultPatterns.ID.FindStringSubmatch(part)
			if len(matchID) > 1 {
				record["idNumber"] = strings.TrimSpace(matchID[1])
			}
		}

		// 3. Extract Request
		if fieldSet["request"] {
			matchReq := DefaultPatterns.Request.FindStringSubmatch(part)
			if len(matchReq) > 1 {
				record["request"] = smartMerge(matchReq[1])
			}
		}

		// 4. Extract Facts
		if fieldSet["factsReason"] {
			matchFact := DefaultPatterns.Facts.FindStringSubmatch(part)
			if len(matchFact) > 1 {
				record["factsReason"] = smartMerge(matchFact[1])
			}
		}

		// If at least one field is non-empty, add the record
		hasData := false
		for _, val := range record {
			if val != "" {
				hasData = true
				break
			}
		}

		if hasData {
			data = append(data, record)
		}
	}
	return data
}

// smartMerge 智能合并换行符
// 逻辑：保留句号、分号、冒号后的换行，或者新条目序号（如“二、”）之前的换行，其他的换行符视作布局造成的干扰并予以合并。
func smartMerge(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// 1. 标准化换行符
	s = strings.ReplaceAll(s, "\r\n", "\n")
	reMultipleNL := regexp.MustCompile(`\n+`)
	s = reMultipleNL.ReplaceAllString(s, "\n")

	// 2. 标记需要保留的“逻辑断点”
	// A. 句末标点后：。；？！
	rePreserveAfter := regexp.MustCompile(`([。；？！])\n`)
	s = rePreserveAfter.ReplaceAllString(s, "$1[LOGICAL_NL]")

	// B. 条目序号前：\n一、 \n(1) 等
	rePreserveBefore := regexp.MustCompile(`\n(\s*(?:[一二三四五六七八九十\d]+[、．]|[(（][一二三四五六七八九十\d]+[)）]))`)
	s = rePreserveBefore.ReplaceAllString(s, "[LOGICAL_NL]$1")

	// 3. 将剩余的所有普通换行符替换为空格（彻底合并）
	s = strings.ReplaceAll(s, "\n", "")

	// 4. 将占位符还原为真正的换行
	s = strings.ReplaceAll(s, "[LOGICAL_NL]", "\n")

	// 5. 深度清理：合并每行内部的多余空格，并去除字词间的冗余
	lines := strings.Split(s, "\n")
	var resultLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// 将行内多余空格合并并剔除
		fields := strings.Fields(trimmed)
		resultLines = append(resultLines, strings.Join(fields, ""))
	}

	return strings.Join(resultLines, "\n")
}
