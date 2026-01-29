package extractor

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

func writeCSV(path string, records []Record) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("\xEF\xBB\xBF") // BOM for Excel

	w := csv.NewWriter(file)
	defer w.Flush()

	if len(records) == 0 {
		return nil
	}

	// 1. Determine Headers from the first record and PatternRegistry
	// We want to keep a consistent order if possible
	var keys []string
	var headers []string

	// Order based on PatternRegistry for consistency
	orderedKeys := []string{"defendant", "idNumber", "request", "factsReason"}
	for _, k := range orderedKeys {
		if _, ok := records[0][k]; ok {
			keys = append(keys, k)
			headers = append(headers, PatternRegistry[k].Label)
		}
	}

	if err := w.Write(headers); err != nil {
		return err
	}

	// 2. Write Data
	for _, r := range records {
		row := make([]string, len(keys))
		for i, k := range keys {
			row[i] = r[k]
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// ExportCSV exports records to a CSV file
func ExportCSV(path string, records []Record) error {
	return writeCSV(path, records)
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

	if len(records) == 0 {
		return nil
	}

	// 1. Determine Headers
	var keys []string
	var headers []string
	orderedKeys := []string{"page", "defendant", "idNumber", "request", "factsReason"}
	for _, k := range orderedKeys {
		if _, ok := records[0][k]; ok {
			keys = append(keys, k)
			headers = append(headers, PatternRegistry[k].Label)
		}
	}

	// Set headers
	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	// 2. Set values
	wrapStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			WrapText: true,
			Vertical: "top",
		},
	})

	for i, r := range records {
		row := i + 2
		for j, k := range keys {
			cell, err := excelize.CoordinatesToCellName(j+1, row)
			if err != nil {
				return err
			}
			if err := f.SetCellValue(sheetName, cell, r[k]); err != nil {
				return err
			}
			// Apply wrap text style
			f.SetCellStyle(sheetName, cell, cell, wrapStyle)
		}
	}

	// Set column widths for better readability
	f.SetColWidth(sheetName, "A", "B", 20)
	f.SetColWidth(sheetName, "C", "D", 50)

	if err := f.SaveAs(path); err != nil {
		return err
	}
	return nil
}
