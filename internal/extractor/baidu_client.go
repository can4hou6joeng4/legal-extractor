package extractor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"legal-extractor/internal/config"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dslipak/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// BaiduClient 百度 AI Studio PaddleOCR 客户端
type BaiduClient struct {
	config     config.BaiduConfig
	httpClient *http.Client
	logger     *slog.Logger
}

// BaiduOCRResponse 百度 Layout Parsing 响应结构
type BaiduOCRResponse struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	Result    struct {
		LayoutParsingResults []struct {
			Markdown struct {
				Text string `json:"text"`
			} `json:"markdown"`
		} `json:"layoutParsingResults"`
	} `json:"result"`
}

// NewBaiduClient 创建百度 OCR 客户端
func NewBaiduClient(logger *slog.Logger) *BaiduClient {
	if logger == nil {
		logger = slog.Default()
	}
	return &BaiduClient{
		config: config.GetBaidu(),
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // VLM 响应可能较慢，给予充足时间
		},
		logger: logger,
	}
}

// ParseDocument 调用百度 Layout Parsing 接口解析文档
func (c *BaiduClient) ParseDocument(fileData []byte, isPdf bool, onProgress ProgressCallback) ([]Record, error) {
	c.logger.Info("开始调用百度 OCR 接口", "isPdf", isPdf, "dataSize", len(fileData))
	if len(fileData) == 0 {
		return nil, fmt.Errorf("文件内容为空")
	}

	if c.config.Token == "" {
		return nil, fmt.Errorf("百度 AI Studio Token 未配置，请检查 config/conf.yaml")
	}

	// 1. 处理超长文档 (百度 API 限制单次 100 页)
	var allPagesMarkdown []string
	const maxPagesPerChunk = 50 // 调小切片粒度以提升大文件稳定性

	if isPdf {
		// 获取总页数
		r, err := pdf.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
		if err == nil {
			totalPages := r.NumPage()
			c.logger.Info("PDF 页数检测完成", "totalPages", totalPages)

			if totalPages > maxPagesPerChunk {
				c.logger.Info("启用大文件物理分块处理模式", "chunkSize", maxPagesPerChunk)
				// 分块处理逻辑
				for start := 1; start <= totalPages; start += maxPagesPerChunk {
					end := start + maxPagesPerChunk - 1
					if end > totalPages {
						end = totalPages
					}

					c.logger.Info(fmt.Sprintf("正在处理分块: 第 %d-%d 页", start, end))
					if onProgress != nil {
						onProgress(start, totalPages)
					}

					// 使用 pdfcpu 进行物理切片
					trimStart := time.Now()
					var chunkBuffer bytes.Buffer
					pageSelection := []string{fmt.Sprintf("%d-%d", start, end)}
					err := api.Trim(bytes.NewReader(fileData), &chunkBuffer, pageSelection, nil)
					if err != nil {
						c.logger.Error("PDF 物理切片失败", "error", err, "pages", pageSelection)
						return nil, fmt.Errorf("PDF 切片失败 (页码 %d-%d): %w", start, end, err)
					}
					c.logger.Info("切片操作完成", "duration", time.Since(trimStart), "size", chunkBuffer.Len())

					// 将物理切片后的数据发送给百度
					pages, err := c.callBaiduAPI(chunkBuffer.Bytes(), true)
					if err != nil {
						c.logger.Error("百度 API 调用失败", "error", err, "pages", pageSelection)
						return nil, err
					}
					allPagesMarkdown = append(allPagesMarkdown, pages...)
				}
			} else {
				pages, err := c.callBaiduAPI(fileData, true)
				if err != nil {
					return nil, err
				}
				allPagesMarkdown = append(allPagesMarkdown, pages...)
			}
		}
	} else {
		pages, err := c.callBaiduAPI(fileData, false)
		if err != nil {
			return nil, err
		}
		allPagesMarkdown = append(allPagesMarkdown, pages...)
	}

	// 2. 按页解析汇总后的 Markdown
	c.logger.Info("所有页面识别完成，开始按页提取法律实体", "totalFetchedPages", len(allPagesMarkdown))
	var allRecords []Record
	for i, pageMd := range allPagesMarkdown {
		records := ParseMarkdown(pageMd)
		for _, rec := range records {
			// 标注准确的页码
			if rec["page"] == "" {
				rec["page"] = fmt.Sprintf("%d", i+1)
			}
			allRecords = append(allRecords, rec)
		}
	}

	c.logger.Info("数据提取完成", "recordCount", len(allRecords))
	return allRecords, nil
}

// callBaiduAPI 封装底层的 API 调用逻辑
func (c *BaiduClient) callBaiduAPI(fileData []byte, isPdf bool) ([]string, error) {
	c.logger.Info("正在向百度 AI Studio 发送 POST 请求...")
	fileBase64 := base64.StdEncoding.EncodeToString(fileData)
	fileType := 1
	if isPdf {
		fileType = 0
	}

	payload := map[string]any{
		"file":                      fileBase64,
		"fileType":                  fileType,
		"useDocOrientationClassify": false,
		"useDocUnwarping":           false,
		"useChartRecognition":       false,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.config.ApiUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.config.Token))

	apiStart := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	c.logger.Info("百度 API 响应接收成功", "status", resp.Status, "duration", time.Since(apiStart))

	var ocrResp BaiduOCRResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocrResp); err != nil {
		return nil, err
	}

	if ocrResp.ErrorCode != 0 {
		return nil, fmt.Errorf("百度 API 错误 (%d): %s", ocrResp.ErrorCode, ocrResp.ErrorMsg)
	}

	var pages []string
	for _, result := range ocrResp.Result.LayoutParsingResults {
		pages = append(pages, result.Markdown.Text)
	}
	return pages, nil
}
