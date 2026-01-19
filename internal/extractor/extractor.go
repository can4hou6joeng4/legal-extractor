package extractor

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"legal-extractor/internal/mcp"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
)

// Extractor handles document extraction logic
type Extractor struct {
	mcpBin  string
	mcpArgs []string
	logger  *slog.Logger
}

// NewExtractor creates a new Extractor instance
func NewExtractor(mcpBin string, mcpArgs []string, logger *slog.Logger) *Extractor {
	if logger == nil {
		logger = slog.Default()
	}
	return &Extractor{
		mcpBin:  mcpBin,
		mcpArgs: mcpArgs,
		logger:  logger,
	}
}

// Record represents a single extracted case as a flexible map
type Record map[string]string

// ExtractData extracts records from a file with specific fields requested
func (e *Extractor) ExtractData(inputFile string, fields []string) ([]Record, error) {
	text, err := e.ExtractText(inputFile)
	if err != nil {
		return nil, fmt.Errorf("error extracting text: %w", err)
	}

	// If fields is nil or empty, use all registered fields
	if len(fields) == 0 {
		for k := range PatternRegistry {
			fields = append(fields, k)
		}
	}

	rawRecords := e.parseCases(text, fields)
	return rawRecords, nil
}

// ExtractText extracts text based on file extension
func (e *Extractor) ExtractText(inputFile string) (string, error) {
	ext := strings.ToLower(filepath.Ext(inputFile))
	switch ext {
	case ".docx":
		return extractTextFromDocx(inputFile)
	case ".pdf":
		return e.extractTextFromPDF(inputFile)
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
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

func (e *Extractor) extractTextFromPDF(path string) (string, error) {
	// 1. Try Native Text Extraction
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var collectedText strings.Builder
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err == nil {
			collectedText.WriteString(text)
		}
	}

	fullText := collectedText.String()

	// 2. Check if text is sufficient. If not, try OCR via MCP.
	// Threshold: < 50 chars suggests it might be a scanned image or empty
	if len(strings.TrimSpace(fullText)) < 50 {
		e.logger.Info("native text extraction yielded minimal content, attempting OCR", "path", path)

		// Initialize MCP Client
		// Use stored configuration
		if e.mcpBin == "" {
			e.logger.Warn("MCP OCR not configured", "reason", "mcpBin is empty")
			return fullText, nil
		}

		client, err := mcp.NewMCPClient(e.mcpBin, e.mcpArgs)
		if err != nil {
			e.logger.Error("Failed to create MCP client", "error", err)
			return fullText, nil
		}
		defer client.Close()

		ocrText, err := client.ExtractText(path)
		if err != nil {
			e.logger.Error("MCP OCR failed", "error", err)
			return fullText, nil // Return native text as fallback
		}

		if len(ocrText) > len(fullText) {
			return ocrText, nil
		}
	}

	return fullText, nil
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
				locEnd := DefaultPatterns.DefEnd.FindStringIndex(remaining)

				var name string
				if locEnd != nil {
					name = remaining[:locEnd[0]]
				} else {
					lines := strings.Split(remaining, "\n")
					if len(lines) > 0 {
						name = lines[0]
					}
				}
				record["defendant"] = strings.Trim(name, " ,，、:：\n\t")
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
