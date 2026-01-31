package extractor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"legal-extractor/internal/config"
	"net/http"
	"strings"
	"time"
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
func (c *BaiduClient) ParseDocument(fileData []byte, isPdf bool) ([]Record, error) {
	if len(fileData) == 0 {
		return nil, fmt.Errorf("文件内容为空")
	}

	if c.config.Token == "" {
		return nil, fmt.Errorf("百度 AI Studio Token 未配置，请检查 config/conf.yaml")
	}

	// 1. 构造请求 Payload
	fileBase64 := base64.StdEncoding.EncodeToString(fileData)
	fileType := 1 // 默认图像
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
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 2. 创建并发送请求
	req, err := http.NewRequest("POST", c.config.ApiUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.config.Token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("网络请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("接口请求失败 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 3. 解析响应
	var ocrResp BaiduOCRResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocrResp); err != nil {
		return nil, fmt.Errorf("解析响应 JSON 失败: %w", err)
	}

	if ocrResp.ErrorCode != 0 {
		return nil, fmt.Errorf("百度 API 错误 (%d): %s", ocrResp.ErrorCode, ocrResp.ErrorMsg)
	}

	// 4. 处理 Markdown 结果
	var allRecords []Record
	for i, result := range ocrResp.Result.LayoutParsingResults {
		markdownText := result.Markdown.Text
		if strings.TrimSpace(markdownText) == "" {
			continue
		}

		// 调用现有的 Markdown 解析逻辑
		pageRecords := ParseMarkdown(markdownText)
		for _, rec := range pageRecords {
			// 如果没抓取到页码，则标注当前的索引
			if rec["page"] == "" {
				rec["page"] = fmt.Sprintf("%d", i+1)
			}
			allRecords = append(allRecords, rec)
		}
	}

	return allRecords, nil
}
