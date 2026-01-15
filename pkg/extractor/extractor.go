package extractor

import (
	"archive/zip"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Record represents a single extracted case
type Record struct {
	Defendant   string `json:"defendant"`
	IDNumber    string `json:"idNumber"`
	Request     string `json:"request"`
	FactsReason string `json:"factsReason"`
}

// ProcessFile extracts text from a docx and writes it to a CSV
func ProcessFile(inputFile, outputFile string) (int, error) {
	// 1. Extract Text from Docx
	text, err := extractTextFromDocx(inputFile)
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
	text, err := extractTextFromDocx(inputFile)
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
		reDefStart := regexp.MustCompile(`被告\s*[:：]`)
		loc := reDefStart.FindStringIndex(part)

		if loc != nil {
			startIdx := loc[1]
			remaining := part[startIdx:]

			reKeywords := regexp.MustCompile(`[,，、\s]+(?:性别|生日|身份证|住址|联系电话)|\n|$`)
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
			record["被告"] = strings.Trim(name, " ,，、:：")
		} else {
			reDefFallback := regexp.MustCompile(`被告\s*[:：]\s*(.*?)\n`)
			match := reDefFallback.FindStringSubmatch(part)
			if len(match) > 1 {
				record["被告"] = strings.TrimSpace(match[1])
			}
		}

		// 2. Extract ID
		reID := regexp.MustCompile(`身份证号码\s*[:：]\s*([\dX]+)`)
		matchID := reID.FindStringSubmatch(part)
		if len(matchID) > 1 {
			record["身份证号码"] = strings.TrimSpace(matchID[1])
		}

		// 3. Extract Request
		reReq := regexp.MustCompile(`(?s)诉讼请求\s*[:：]\s*(.*?)\s*事实与理由`)
		matchReq := reReq.FindStringSubmatch(part)
		if len(matchReq) > 1 {
			record["诉讼请求"] = strings.TrimSpace(matchReq[1])
		}

		// 4. Extract Facts
		reFact := regexp.MustCompile(`(?s)事实与理由\s*[:：]\s*(.*?)\s*此致`)
		matchFact := reFact.FindStringSubmatch(part)
		if len(matchFact) > 1 {
			record["事实与理由"] = strings.TrimSpace(matchFact[1])
		}

		if record["被告"] != "" || record["身份证号码"] != "" {
			data = append(data, record)
		}
	}
	return data
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
