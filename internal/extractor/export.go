package extractor

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

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
