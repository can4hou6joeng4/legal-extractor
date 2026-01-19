package app

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"legal-extractor/internal/extractor"

	wr "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx       context.Context
	extractor *extractor.Extractor
}

// NewApp creates a new App application struct
func NewApp(e *extractor.Extractor) *App {
	return &App{
		extractor: e,
	}
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectFile opens a file dialog to select a .docx file
func (a *App) SelectFile() (string, error) {
	file, err := wr.OpenFileDialog(a.ctx, wr.OpenDialogOptions{
		Title: "Select Legal Document (.docx)",
		Filters: []wr.FileFilter{
			{
				DisplayName: "Legal Documents (*.docx;*.pdf)",
				Pattern:     "*.docx;*.pdf",
			},
		},
	})
	if err != nil {
		return "", err
	}
	return file, nil
}

// ExtractResult holds the extraction result
type ExtractResult struct {
	Success      bool               `json:"success"`
	RecordCount  int                `json:"recordCount"`
	OutputPath   string             `json:"outputPath"`
	ErrorMessage string             `json:"errorMessage,omitempty"`
	Records      []extractor.Record `json:"records,omitempty"`
	FieldLabels  map[string]string  `json:"fieldLabels,omitempty"` // Map of key -> Chinese label
}

// FieldOption represents a selectable extraction field
type FieldOption struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// ScanFields analyzes the file and returns fields that exist in the content
func (a *App) ScanFields(inputFile string) ([]FieldOption, error) {
	if inputFile == "" {
		return nil, fmt.Errorf("no file selected")
	}

	// 提取数据并检测字段（PDF 和 DOCX 都使用统一接口）
	records, err := a.extractor.ExtractData(inputFile, []string{"defendant", "idNumber", "request", "factsReason"})
	if err != nil {
		return nil, fmt.Errorf("failed to extract data: %v", err)
	}

	var options []FieldOption
	orderedKeys := []string{"defendant", "idNumber", "request", "factsReason"}

	// 检查哪些字段在提取的记录中有值
	fieldExists := make(map[string]bool)
	for _, record := range records {
		for k, v := range record {
			if v != "" {
				fieldExists[k] = true
			}
		}
	}

	// 按顺序返回存在的字段
	for _, k := range orderedKeys {
		if fieldExists[k] {
			if p, ok := extractor.PatternRegistry[k]; ok {
				options = append(options, FieldOption{
					Key:   k,
					Label: p.Label,
				})
			}
		}
	}

	return options, nil
}

// SelectOutputPath opens a save dialog for the user to choose destination
func (a *App) SelectOutputPath(defaultName string) (string, error) {
	if defaultName == "" {
		defaultName = "extracted_data.csv"
	}

	// Ensure default name has correct extension base logic if needed,
	// but mostly we trust the caller or just use a generic name.
	// Actually, best to let frontend pass the input filename so we can suggest input_extracted.csv

	outputFile, err := wr.SaveFileDialog(a.ctx, wr.SaveDialogOptions{
		Title:           "Select Output Location",
		DefaultFilename: defaultName,
		Filters: []wr.FileFilter{
			{
				DisplayName: "Excel Files (*.xlsx)",
				Pattern:     "*.xlsx",
			},
			{
				DisplayName: "CSV Files (*.csv)",
				Pattern:     "*.csv",
			},
			{
				DisplayName: "JSON Files (*.json)",
				Pattern:     "*.json",
			},
		},
	})

	if err != nil {
		return "", err
	}
	return outputFile, nil
}

// ExtractToPath processes the input file and saves to the specific output path
func (a *App) ExtractToPath(inputPath, outputPath string, fields []string) ExtractResult {
	if inputPath == "" || outputPath == "" {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "Invalid input or output path",
		}
	}

	// 1. Extract Data
	records, err := a.extractor.ExtractData(inputPath, fields)
	if err != nil {
		return ExtractResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Extraction failed: %v", err),
		}
	}

	if len(records) == 0 {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "No records found in document",
		}
	}

	// 2. Save based on extension
	if strings.HasSuffix(strings.ToLower(outputPath), ".json") {
		err = extractor.ExportJSON(outputPath, records)
	} else if strings.HasSuffix(strings.ToLower(outputPath), ".xlsx") {
		err = extractor.ExportExcel(outputPath, records)
	} else {
		err = extractor.ExportCSV(outputPath, records)
	}

	if err != nil {
		return ExtractResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Save failed: %v", err),
		}
	}

	return ExtractResult{
		Success:     true,
		RecordCount: len(records),
		OutputPath:  outputPath,
	}
}

// PreviewData extracts and returns records for preview (without saving)
func (a *App) PreviewData(inputPath string, fields []string) ExtractResult {
	if inputPath == "" {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "No file selected",
		}
	}

	records, err := a.extractor.ExtractData(inputPath, fields)
	if err != nil {
		return ExtractResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Preview failed: %v", err),
		}
	}

	// Get labels for UI
	labels := make(map[string]string)
	for k, p := range extractor.PatternRegistry {
		labels[k] = p.Label
	}

	return ExtractResult{
		Success:     true,
		RecordCount: len(records),
		Records:     records,
		FieldLabels: labels,
	}
}

// OpenFile opens the file at the given path using the system's default application
func (a *App) OpenFile(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}
