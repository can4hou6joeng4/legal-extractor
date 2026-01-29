package extractor

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/dslipak/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Extractor 处理器，负责协调不同格式的提取策略
type Extractor struct {
	logger        *slog.Logger
	tencentClient *TencentClient
	cache         map[string][]Record
	cacheMu       sync.RWMutex
}

// NewExtractor 创建一个新的提取器实例
func NewExtractor(logger *slog.Logger) *Extractor {
	if logger == nil {
		logger = slog.Default()
	}
	return &Extractor{
		logger:        logger,
		tencentClient: NewTencentClient(),
		cache:         make(map[string][]Record),
	}
}

// Record 代表一条提取的记录
type Record map[string]string

// ProgressCallback 进度回调函数
type ProgressCallback func(current, total int)

// ExtractData 根据文件类型选择提取策略
func (e *Extractor) ExtractData(fileData []byte, fileName string, fields []string, onProgress ProgressCallback) ([]Record, error) {
	ext := strings.ToLower(filepath.Ext(fileName))

	// 1. 检查缓存
	e.cacheMu.RLock()
	if cached, ok := e.cache[fileName]; ok {
		e.logger.Info("命中缓存结果，跳过提取", "file", fileName)
		e.cacheMu.RUnlock()
		return cached, nil
	}
	e.cacheMu.RUnlock()

	var records []Record
	var err error

	switch ext {
	case ".pdf":
		records, err = e.extractPdf(fileData, fields, onProgress)
	case ".jpg", ".png", ".jpeg":
		return nil, fmt.Errorf("图片识别功能已暂时禁用（仅支持PDF）")
	case ".docx":
		e.logger.Info("使用本地原生逻辑提取 DOCX", "file", fileName)
		records, err = e.extractFromDocx(fileData, fields)
	default:
		return nil, fmt.Errorf("不支持的文件格式: %s", ext)
	}

	if err != nil {
		return nil, err
	}

	// 2. 写入缓存
	e.cacheMu.Lock()
	e.cache[fileName] = records
	e.cacheMu.Unlock()

	return records, nil
}

// extractPdf 处理 PDF 提取（优先本地提取文本层）
func (e *Extractor) extractPdf(fileData []byte, fields []string, onProgress ProgressCallback) ([]Record, error) {
	// 1. 获取总页数
	totalPages := 1
	pageCount, err := api.PageCount(bytes.NewReader(fileData), nil)
	if err == nil {
		totalPages = pageCount
	}

	// 2. 探测第一页文本层
	firstPageText, _ := e.extractPageTextLocally(fileData, 1)
	if len(strings.TrimSpace(firstPageText)) > 20 {
		e.logger.Info("检测到 PDF 文本层，切换至 [本地高速解析] 模式", "totalPages", totalPages)
		return e.batchExtractLocalPdf(fileData, fields, totalPages, onProgress)
	}

	e.logger.Info("未检测到 PDF 文本层或文本过少，切换至 [本地系统 OCR] 模式", "totalPages", totalPages)
	return e.extractViaWinOcr(fileData, totalPages, onProgress)
}

// extractPageTextLocally 本地提取指定页码的文本
func (e *Extractor) extractPageTextLocally(fileData []byte, pageNum int) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
	if err != nil {
		return "", err
	}

	if pageNum > r.NumPage() {
		return "", fmt.Errorf("页码超出范围")
	}

	p := r.Page(pageNum)
	if p.V.IsNull() {
		return "", fmt.Errorf("页面内容为空")
	}

	text, _ := p.GetPlainText(nil)
	return text, nil
}

// batchExtractLocalPdf 批量本地提取 PDF 文本层
func (e *Extractor) batchExtractLocalPdf(fileData []byte, fields []string, totalPages int, onProgress ProgressCallback) ([]Record, error) {
	r, err := pdf.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
	if err != nil {
		return nil, err
	}

	var allRecords []Record
	for i := 1; i <= totalPages; i++ {
		if onProgress != nil {
			onProgress(i, totalPages)
		}

		text, _ := r.Page(i).GetPlainText(nil)
		if strings.TrimSpace(text) == "" {
			continue
		}

		// 复用 parseCases 逻辑
		pageRecords := e.parseCases(text, fields)
		if len(pageRecords) > 0 {
			for _, rec := range pageRecords {
				rec["page"] = fmt.Sprintf("%d", i)
				allRecords = append(allRecords, rec)
			}
		}
	}

	return allRecords, nil
}

// extractViaWinOcr 调用 Windows 系统原生 OCR 桥接工具
func (e *Extractor) extractViaWinOcr(fileData []byte, totalPages int, onProgress ProgressCallback) ([]Record, error) {
	// 1. 创建临时文件存储 PDF 内容
	tempFile, err := os.CreateTemp("", "legal_ocr_*.pdf")
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := tempFile.Write(fileData); err != nil {
		return nil, fmt.Errorf("写入临时文件失败: %w", err)
	}

	// 2. 定位桥接工具路径
	exePath, _ := os.Executable()
	baseDir := filepath.Dir(exePath)
	bridgePath := filepath.Join(baseDir, "bridge_bin", "WinOcrBridge.exe")

	// 开发环境适配
	if _, err := os.Stat(bridgePath); os.IsNotExist(err) {
		// 尝试查找源码同级目录 (wails dev 模式)
		bridgePath = filepath.Join("internal", "extractor", "bridge_bin", "WinOcrBridge.exe")
		if _, err := os.Stat(bridgePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("找不到 Windows OCR 桥接工具 (WinOcrBridge.exe)，请确保它位于 bridge_bin 目录下")
		}
	}

	var allRecords []Record
	for i := 1; i <= totalPages; i++ {
		if onProgress != nil {
			onProgress(i, totalPages)
		}

		// 3. 调用命令行工具
		cmd := exec.Command(bridgePath, tempFile.Name(), fmt.Sprintf("%d", i))
		output, err := cmd.CombinedOutput()
		if err != nil {
			e.logger.Warn("本地 OCR 识别页面失败", "page", i, "error", err, "output", string(output))
			continue
		}

		text := strings.TrimSpace(string(output))
		if text == "" {
			continue
		}

		// 4. 解析识别出的文字
		pageRecords := e.parseCases(text, nil) // 使用所有已注册字段
		if len(pageRecords) > 0 {
			for _, rec := range pageRecords {
				rec["page"] = fmt.Sprintf("%d", i)
				allRecords = append(allRecords, rec)
			}
		}
	}

	return allRecords, nil
}

// extractFromDocx 保留原有的本地 DOCX 提取逻辑
func (e *Extractor) extractFromDocx(fileData []byte, fields []string) ([]Record, error) {
	text, err := extractTextFromDocx(fileData)
	if err != nil {
		return nil, err
	}

	if len(fields) == 0 {
		for k := range PatternRegistry {
			fields = append(fields, k)
		}
	}

	return e.parseCases(text, fields), nil
}

// extractTextFromDocx 核心 DOCX 文本提取逻辑
func extractTextFromDocx(fileData []byte) (string, error) {
	r, err := zip.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
	if err != nil {
		return "", err
	}

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

// parseCases 现有的本地正则解析逻辑 (用于 DOCX)
func (e *Extractor) parseCases(text string, fields []string) []Record {
	parts := DefaultPatterns.Split.Split(text, -1)
	var data []Record

	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}

		record := make(Record)
		fieldSet := make(map[string]bool)
		for _, f := range fields {
			fieldSet[f] = true
		}

		// 1. 提取被告
		if fieldSet["defendant"] {
			loc := DefaultPatterns.DefStart.FindStringIndex(part)
			if loc != nil {
				startIdx := loc[1]
				remaining := part[startIdx:]
				cleanRemaining := strings.ReplaceAll(remaining, "\n", "")
				locEnd := DefaultPatterns.DefEnd.FindStringIndex(cleanRemaining)

				var name string
				if locEnd != nil {
					name = cleanRemaining[:locEnd[0]]
				} else {
					if len(cleanRemaining) > 50 {
						name = cleanRemaining[:50]
					} else {
						name = cleanRemaining
					}
				}
				record["defendant"] = strings.TrimSpace(name)
			}
		}

		// 2. 提取身份证
		if fieldSet["idNumber"] {
			matchID := DefaultPatterns.ID.FindStringSubmatch(part)
			if len(matchID) > 1 {
				record["idNumber"] = strings.TrimSpace(matchID[1])
			}
		}

		// 3. 提取请求
		if fieldSet["request"] {
			matchReq := DefaultPatterns.Request.FindStringSubmatch(part)
			if len(matchReq) > 1 {
				record["request"] = smartMerge(matchReq[1])
			}
		}

		// 4. 提取事实
		if fieldSet["factsReason"] {
			matchFact := DefaultPatterns.Facts.FindStringSubmatch(part)
			if len(matchFact) > 1 {
				record["factsReason"] = smartMerge(matchFact[1])
			}
		}

		if len(record) > 0 {
			data = append(data, record)
		}
	}
	return data
}

// smartMerge 智能合并换行符
// 逻辑：保留句号、分号、冒号后的换行，或者新条目序号（如"二、"）之前的换行，其他的换行符视作布局造成的干扰并予以合并。
func smartMerge(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// 1. 标准化换行符
	s = strings.ReplaceAll(s, "\r\n", "\n")
	reMultipleNL := regexp.MustCompile(`\n+`)
	s = reMultipleNL.ReplaceAllString(s, "\n")

	// 2. 标记需要保留的"逻辑断点"
	// A. 句末标点后：。；？！
	rePreserveAfter := regexp.MustCompile(`([。；？！])\n`)
	s = rePreserveAfter.ReplaceAllString(s, "$1[LOGICAL_NL]")

	// B. 条目序号前：\n一、 \n(1) 等
	rePreserveBefore := regexp.MustCompile(`\n(\s*(?:[一二三四五六七八九十\d]+[、．]|[(（][一二三四五六七八九十\d]+[)）]))`)
	s = rePreserveBefore.ReplaceAllString(s, "[LOGICAL_NL]$1")

	// 3. 合并 OCR 碎行：将剩余的非逻辑换行符替换为一个小空格，防止文字粘连
	s = strings.ReplaceAll(s, "\n", " ")

	// 4. 将占位符还原为真正的换行符
	s = strings.ReplaceAll(s, "[LOGICAL_NL]", "\n")

	// 5. 深度清理：合并每行内部的多余空格
	lines := strings.Split(s, "\n")
	var resultLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// 压缩行内连续空格，但保留单个空格
		fields := strings.Fields(trimmed)
		resultLines = append(resultLines, strings.Join(fields, " "))
	}

	return strings.Join(resultLines, "\n")
}
