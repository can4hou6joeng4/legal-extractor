package main

import (
	"context"
	"fmt"
	"strings"

	"legal-extractor/pkg/extractor"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectFile opens a file dialog to select a .docx file
func (a *App) SelectFile() (string, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Legal Document (.docx)",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Word Documents (*.docx)",
				Pattern:     "*.docx",
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
}

// SelectOutputPath opens a save dialog for the user to choose destination
func (a *App) SelectOutputPath(defaultName string) (string, error) {
	if defaultName == "" {
		defaultName = "extracted_data.csv"
	}

	// Ensure default name has correct extension base logic if needed,
	// but mostly we trust the caller or just use a generic name.
	// Actually, best to let frontend pass the input filename so we can suggest input_extracted.csv

	outputFile, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Select Output Location",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
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
func (a *App) ExtractToPath(inputPath, outputPath string) ExtractResult {
	if inputPath == "" || outputPath == "" {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "Invalid input or output path",
		}
	}

	// 1. Extract Data
	records, err := extractor.ExtractData(inputPath)
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
func (a *App) PreviewData(inputPath string) ExtractResult {
	if inputPath == "" {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "No file selected",
		}
	}

	records, err := extractor.ExtractData(inputPath)
	if err != nil {
		return ExtractResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Preview failed: %v", err),
		}
	}

	return ExtractResult{
		Success:     true,
		RecordCount: len(records),
		Records:     records,
	}
}

// SaveToPath allows user to choose output location
func (a *App) SaveToPath(records []extractor.Record) ExtractResult {
	file, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Extraction Results",
		DefaultFilename: "extraction_result.csv",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "CSV Files (*.csv)",
				Pattern:     "*.csv",
			},
		},
	})
	if err != nil || file == "" {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "Save cancelled",
		}
	}

	err = extractor.ExportCSV(file, records)
	if err != nil {
		return ExtractResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Save failed: %v", err),
		}
	}

	return ExtractResult{
		Success:     true,
		RecordCount: len(records),
		OutputPath:  file,
	}
}
