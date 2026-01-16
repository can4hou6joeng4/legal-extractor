package extractor

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"legal-extractor/pkg/mcp"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
)

// MCP Config storage
var (
	mcpBin  string
	mcpArgs []string
)

// SetMCPConfig sets the configuration for the MCP client
func SetMCPConfig(bin string, args []string) {
	mcpBin = bin
	mcpArgs = args
}

// Record represents a single extracted case
type Record struct {
	Defendant   string `json:"defendant"`
	IDNumber    string `json:"idNumber"`
	Request     string `json:"request"`
	FactsReason string `json:"factsReason"`
}

// ProcessFile extracts text from a docx and writes it to a CSV
func ProcessFile(inputFile, outputFile string) (int, error) {
	// 1. Extract Text
	var text string
	var err error
	ext := strings.ToLower(filepath.Ext(inputFile))
	if ext == ".docx" {
		text, err = extractTextFromDocx(inputFile)
	} else if ext == ".pdf" {
		text, err = extractTextFromPDF(inputFile)
	} else {
		return 0, fmt.Errorf("unsupported file extension: %s", ext)
	}

	if err != nil {
		return 0, fmt.Errorf("error extracting text: %w", err)
	}

	// 2. Parse Data using Regex
	records := parseCases(text)

	// 3. Write to CSV
	err = writeCSV(outputFile, records)
	if err != nil {
		return 0, fmt.Errorf("error writing CSV: %w", err)
	}

	return len(records), nil
}

// ExtractData extracts records from a docx file and returns them
func ExtractData(inputFile string) ([]Record, error) {
	var text string
	var err error

	ext := strings.ToLower(filepath.Ext(inputFile))
	if ext == ".docx" {
		text, err = extractTextFromDocx(inputFile)
	} else if ext == ".pdf" {
		text, err = extractTextFromPDF(inputFile)
	} else {
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("error extracting text: %w", err)
	}
	rawRecords := parseCases(text)

	// Convert to typed struct
	records := make([]Record, len(rawRecords))
	for i, r := range rawRecords {
		records[i] = Record{
			Defendant:   r["被告"],
			IDNumber:    r["身份证号码"],
			Request:     r["诉讼请求"],
			FactsReason: r["事实与理由"],
		}
	}
	return records, nil
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

func extractTextFromPDF(path string) (string, error) {
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
		fmt.Println("Text extraction yielded minimal content. Attempting OCR via MCP...")

		// Initialize MCP Client
		// Use stored configuration
		if mcpBin == "" {
			fmt.Printf("MCP OCR not configured (mcpBin is empty)\n")
			return fullText, nil
		}

		client, err := mcp.NewMCPClient(mcpBin, mcpArgs)
		if err != nil {
			fmt.Printf("Failed to create MCP client: %v\n", err)
			return fullText, nil
		}
		defer client.Close()

		ocrText, err := client.ExtractText(path)
		if err != nil {
			fmt.Printf("MCP OCR failed: %v\n", err)
			return fullText, nil // Return native text as fallback
		}

		if len(ocrText) > len(fullText) {
			return ocrText, nil
		}
	}

	return fullText, nil
}

func parseCases(text string) []map[string]string {
	reSplit := regexp.MustCompile(`民\s*事\s*起\s*诉\s*状`)
	parts := reSplit.Split(text, -1)

	var data []map[string]string

	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}

		record := make(map[string]string)

		// 1. Extract Defendant
		reDefStart := regexp.MustCompile(`被\s*告\s*[:：]`)
		loc := reDefStart.FindStringIndex(part)

		if loc != nil {
			startIdx := loc[1]
			remaining := part[startIdx:]

			reKeywords := regexp.MustCompile(`[,，、\s]+(?:性\s*别|生\s*日|身\s*份\s*证|住\s*址|联\s*系\s*电\s*话)|\n|$`)
			locEnd := reKeywords.FindStringIndex(remaining)

			var name string
			if locEnd != nil {
				name = remaining[:locEnd[0]]
			} else {
				lines := strings.Split(remaining, "\n")
				if len(lines) > 0 {
					name = lines[0]
				}
			}
			record["被告"] = strings.Trim(name, " ,，、:：\n\t")
		} else {
			reDefFallback := regexp.MustCompile(`被\s*告\s*[:：]\s*(.*?)\n`)
			match := reDefFallback.FindStringSubmatch(part)
			if len(match) > 1 {
				record["被告"] = strings.TrimSpace(match[1])
			}
		}

		// 2. Extract ID
		reID := regexp.MustCompile(`身\s*份\s*证\s*号\s*码\s*[:：]\s*([\dX]+)`)
		matchID := reID.FindStringSubmatch(part)
		if len(matchID) > 1 {
			record["身份证号码"] = strings.TrimSpace(matchID[1])
		}

		// 3. Extract Request
		reReq := regexp.MustCompile(`(?s)诉\s*讼\s*请\s*求\s*[:：]\s*(.*?)\s*事\s*实\s*与\s*理\s*由`)
		matchReq := reReq.FindStringSubmatch(part)
		if len(matchReq) > 1 {
			record["诉讼请求"] = smartMerge(matchReq[1])
		}

		// 4. Extract Facts
		reFact := regexp.MustCompile(`(?s)事\s*实\s*与\s*理\s*由\s*[:：]\s*(.*?)\s*此\s*致`)
		matchFact := reFact.FindStringSubmatch(part)
		if len(matchFact) > 1 {
			record["事实与理由"] = smartMerge(matchFact[1])
		}

		if record["被告"] != "" || record["身份证号码"] != "" {
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

func writeCSV(path string, data []map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("\xEF\xBB\xBF") // BOM for Excel

	w := csv.NewWriter(file)
	defer w.Flush()

	header := []string{"被告", "身份证号码", "诉讼请求", "事实与理由"}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, row := range data {
		record := []string{
			row["被告"],
			row["身份证号码"],
			row["诉讼请求"],
			row["事实与理由"],
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// ExportCSV exports records to a CSV file
func ExportCSV(path string, records []Record) error {
	data := make([]map[string]string, len(records))
	for i, r := range records {
		data[i] = map[string]string{
			"被告":    r.Defendant,
			"身份证号码": r.IDNumber,
			"诉讼请求":  r.Request,
			"事实与理由": r.FactsReason,
		}
	}
	return writeCSV(path, data)
}

// ExportJSON exports records to a JSON file
func ExportJSON(path string, records []Record) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(records)
}

// ExportExcel exports records to an Excel file
func ExportExcel(path string, records []Record) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Create a new sheet.
	sheetName := "Sheet1"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// Set active sheet of the workbook.
	f.SetActiveSheet(index)

	// Set headers
	headers := []string{"被告", "身份证号码", "诉讼请求", "事实与理由"}
	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	// Set values
	for i, r := range records {
		row := i + 2
		values := []string{r.Defendant, r.IDNumber, r.Request, r.FactsReason}
		for j, v := range values {
			cell, err := excelize.CoordinatesToCellName(j+1, row)
			if err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, cell, v); err != nil {
				return err
			}
		}
	}

	if err := f.SaveAs(path); err != nil {
		return err
	}
	return nil
}
