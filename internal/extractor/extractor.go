package extractor

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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

// ExtractData 根据文件类型选择提取策略
// 参数说明：
//   - fileData: 文件的二进制内容（支持内存数据，便于 Web 版集成）
//   - fileName: 文件名（用于判断文件类型和缓存键）
//   - fields: 需要提取的字段列表
func (e *Extractor) ExtractData(fileData []byte, fileName string, fields []string) ([]Record, error) {
	ext := strings.ToLower(filepath.Ext(fileName))

	// 1. 检查缓存（使用 fileName 作为缓存键）
	e.cacheMu.RLock()
	if cached, ok := e.cache[fileName]; ok {
		e.logger.Info("命中缓存结果，跳过 API 调用", "file", fileName)
		e.cacheMu.RUnlock()
		return cached, nil
	}
	e.cacheMu.RUnlock()

	var records []Record
	var err error

	switch ext {
	case ".pdf", ".jpg", ".png", ".jpeg":
		e.logger.Info("使用腾讯云 OCR 提取数据", "file", fileName)
		records, err = e.extractViaCloud(fileData)
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

// extractViaCloud 调用腾讯云 OCR API 进行提取
func (e *Extractor) extractViaCloud(fileData []byte) ([]Record, error) {
	record, err := e.tencentClient.ParseDocument(fileData)
	if err != nil {
		e.logger.Error("腾讯云 OCR 提取失败", "error", err)
		return nil, err
	}

	if len(record) == 0 {
		return nil, fmt.Errorf("未能从文档中提取到有效字段")
	}

	return []Record{record}, nil
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
