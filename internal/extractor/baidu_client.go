package extractor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"legal-extractor/internal/config"
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
func NewBaiduClient() *BaiduClient {
	return &BaiduClient{
		config: config.GetBaidu(),
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // VLM 响应可能较慢，给予充足时间
		},
	}
}

// ParseDocument 调用百度 Layout Parsing 接口解析文档
func (c *BaiduClient) ParseDocument(fileData []byte, isPdf bool, onProgress ProgressCallback) ([]Record, error) {
	if len(fileData) == 0 {
		return nil, fmt.Errorf("文件内容为空")
	}

	if c.config.Token == "" {
		return nil, fmt.Errorf("百度 AI Studio Token 未配置，请检查 config/conf.yaml")
	}

	// 1. 处理超长文档 (百度 API 限制单次 100 页)
	var allMarkdown strings.Builder
	const maxPagesPerChunk = 100

	if isPdf {
		// 获取总页数
		r, err := pdf.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
		if err == nil {
			totalPages := r.NumPage()
			if totalPages > maxPagesPerChunk {
				// 分块处理逻辑
				for start := 1; start <= totalPages; start += maxPagesPerChunk {
					end := start + maxPagesPerChunk - 1
					if end > totalPages {
						end = totalPages
					}

					if onProgress != nil {
						onProgress(start, totalPages)
					}

					// 使用 pdfcpu 进行物理切片
					var chunkBuffer bytes.Buffer
					pageSelection := []string{fmt.Sprintf("%d-%d", start, end)}
					err := api.Trim(bytes.NewReader(fileData), &chunkBuffer, pageSelection, nil)
					if err != nil {
						return nil, fmt.Errorf("PDF 切片失败 (页码 %d-%d): %w", start, end, err)
					}

					// 将物理切片后的数据发送给百度
					markdown, err := c.callBaiduAPI(chunkBuffer.Bytes(), true)
					if err != nil {
						return nil, err
					}
					allMarkdown.WriteString(markdown)
					allMarkdown.WriteString("\n\n")
				}
			} else {
				markdown, err := c.callBaiduAPI(fileData, true)
				if err != nil {
					return nil, err
				}
				allMarkdown.WriteString(markdown)
			}
		}
	} else {
		markdown, err := c.callBaiduAPI(fileData, false)
		if err != nil {
			return nil, err
		}
		allMarkdown.WriteString(markdown)
	}

	// 2. 解析汇总后的 Markdown
	return ParseMarkdown(allMarkdown.String()), nil
}

// callBaiduAPI 封装底层的 API 调用逻辑
func (c *BaiduClient) callBaiduAPI(fileData []byte, isPdf bool) (string, error) {
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
		return "", err
	}

	req, err := http.NewRequest("POST", c.config.ApiUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.config.Token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var ocrResp BaiduOCRResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocrResp); err != nil {
		return "", err
	}

	if ocrResp.ErrorCode != 0 {
		return "", fmt.Errorf("百度 API 错误 (%d): %s", ocrResp.ErrorCode, ocrResp.ErrorMsg)
	}

	var sb strings.Builder
	for _, result := range ocrResp.Result.LayoutParsingResults {
		sb.WriteString(result.Markdown.Text)
		sb.WriteString("\n\n") // 页间分隔
	}
	return sb.String(), nil
}
