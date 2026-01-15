package main

import (
	"context"
	"fmt"
	"path/filepath"
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

// ExtractFromFile processes the docx and returns results
func (a *App) ExtractFromFile(inputPath string) ExtractResult {
	if inputPath == "" {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "No file selected",
		}
	}

	// Generate output path
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	outputPath := filepath.Join(dir, name+"_extracted.csv")

	// Process file
	count, err := extractor.ProcessFile(inputPath, outputPath)
	if err != nil {
		return ExtractResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Extraction failed: %v", err),
		}
	}

	return ExtractResult{
		Success:     true,
		RecordCount: count,
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
