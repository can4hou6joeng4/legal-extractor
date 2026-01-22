package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"legal-extractor/internal/config"
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

// GetTrialStatus 返回试用期状态
func (a *App) GetTrialStatus() config.TrialStatus {
	return config.GetTrialStatus()
}

// GetMachineID 返回当前设备的唯一机器码
func (a *App) GetMachineID() string {
	return config.GetMachineID()
}

// Activate 验证并激活授权码
func (a *App) Activate(licenseKey string) (bool, error) {
	machineID := config.GetMachineID()
	if config.VerifyLicense(machineID, licenseKey) {
		err := config.SaveLicense(licenseKey)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, fmt.Errorf("授权码无效，请检查后重试")
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

	// 适配器层：负责读取本地文件
	fileData, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// 提取数据并检测字段（PDF 和 DOCX 都使用统一接口）
	records, err := a.extractor.ExtractData(fileData, inputFile, []string{"defendant", "idNumber", "request", "factsReason"})
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
	// 检查试用期状态
	status := config.GetTrialStatus()
	if status.IsExpired {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "试用期已结束（限 7 天），功能已锁定。请联系开发者获取正式版。",
		}
	}

	if inputPath == "" || outputPath == "" {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "Invalid input or output path",
		}
	}

	// 适配器层：负责读取本地文件
	fileData, err := os.ReadFile(inputPath)
	if err != nil {
		return ExtractResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to read file: %v", err),
		}
	}

	// 1. Extract Data
	records, err := a.extractor.ExtractData(fileData, inputPath, fields)
	if err != nil {
		// 转换特定错误码
		errMsg := err.Error()
		if strings.Contains(errMsg, "PDF_ENCRYPTED_OR_LOCKED") {
			errMsg = "PDF_ENCRYPTED_OR_LOCKED"
		}
		return ExtractResult{
			Success:      false,
			ErrorMessage: errMsg,
		}
	}

	if len(records) == 0 {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "No records found in document",
		}
	}

	// 2. Save based on extension
	return a.ExportData(records, outputPath)
}

// ExportData 接收用户编辑后的数据并直接保存到指定路径
func (a *App) ExportData(records []extractor.Record, outputPath string) ExtractResult {
	if len(records) == 0 || outputPath == "" {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "无有效数据或未指定输出路径",
		}
	}

	var err error
	lowerPath := strings.ToLower(outputPath)
	if strings.HasSuffix(lowerPath, ".json") {
		err = extractor.ExportJSON(outputPath, records)
	} else if strings.HasSuffix(lowerPath, ".xlsx") {
		err = extractor.ExportExcel(outputPath, records)
	} else {
		err = extractor.ExportCSV(outputPath, records)
	}

	if err != nil {
		return ExtractResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("导出失败: %v", err),
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
	// 检查试用期状态
	status := config.GetTrialStatus()
	if status.IsExpired {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "试用期已结束（限 7 天），预览功能已锁定。请联系开发者获取正式版。",
		}
	}

	if inputPath == "" {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "No file selected",
		}
	}

	// 适配器层：负责读取本地文件
	fileData, err := os.ReadFile(inputPath)
	if err != nil {
		return ExtractResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to read file: %v", err),
		}
	}

	// 检查文件内容是否为空
	if len(fileData) == 0 {
		return ExtractResult{
			Success:      false,
			ErrorMessage: "文件内容为空，请检查文件是否损坏",
		}
	}

	records, err := a.extractor.ExtractData(fileData, inputPath, fields)
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
