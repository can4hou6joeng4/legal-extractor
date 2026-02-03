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
			Timeout: 180 * time.Second, // 增加到 180 秒，为复杂长文档预留充足处理时间
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
	const maxPagesPerChunk = 20 // 调小切片粒度（从50改为20）以显著提升云端解析的稳定性

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
						onProgress(start, totalPages, "正在分析文档结构并准备解析...")
					}

					// 1. 物理切片操作
					var chunkBuffer bytes.Buffer
					pageSelection := []string{fmt.Sprintf("%d-%d", start, end)}
					err := api.Trim(bytes.NewReader(fileData), &chunkBuffer, pageSelection, nil)
					if err != nil {
						return nil, fmt.Errorf("PDF 切片失败: %w", err)
					}

					// 2. 实施“避让重试”策略处理云端 500 错误
					var pages []string
					maxRetries := 2
					for retry := 0; retry <= maxRetries; retry++ {
						if retry > 0 {
							c.logger.Warn(fmt.Sprintf("分块 %d-%d 尝试第 %d 次重试...", start, end, retry))
							time.Sleep(20 * time.Second) // 收到 500 后重试需等待更久，给服务器释放资源
						}

						if onProgress != nil {
							onProgress(start, totalPages, fmt.Sprintf("正在对第 %d-%d 页进行深度识别...", start, end))
						}

						pages, err = c.callBaiduAPI(chunkBuffer.Bytes(), true, onProgress)
						if err == nil {
							break
						}

						// 如果是 500 错误且还有重试机会
						if strings.Contains(err.Error(), "500") && retry < maxRetries {
							continue
						}
						return nil, err // 其他严重错误或重试耗尽则退出
					}

					allPagesMarkdown = append(allPagesMarkdown, pages...)

					// 3. 强制冷却，防止连续高压导致百度后端崩溃
					if end < totalPages {
						c.logger.Info("分块处理完成，进入 10 秒冷却期以释放云端算力...")
						time.Sleep(10 * time.Second)
					}
				}
			} else {
				if onProgress != nil {
					onProgress(1, totalPages, "正在进行深度识别与内容校对，请稍候...")
				}
				pages, err := c.callBaiduAPI(fileData, true, onProgress)
				if err != nil {
					return nil, err
				}
				allPagesMarkdown = append(allPagesMarkdown, pages...)
			}
		}
	} else {
		if onProgress != nil {
			onProgress(1, 1, "正在对文档进行语义化识别...")
		}
		pages, err := c.callBaiduAPI(fileData, false, onProgress)
		if err != nil {
			return nil, err
		}
		allPagesMarkdown = append(allPagesMarkdown, pages...)
	}

	// 2. 按页解析汇总后的 Markdown
	c.logger.Info("所有页面识别完成，开始按页提取法律实体", "totalFetchedPages", len(allPagesMarkdown))
	var allRecords []Record
	totalPages := len(allPagesMarkdown)
	for i, pageMd := range allPagesMarkdown {
		if onProgress != nil {
			// 增加微小延迟 (50ms)，让前端有足够时间渲染进度条的跳动，避免瞬间完成
			time.Sleep(50 * time.Millisecond)
			onProgress(i+1, totalPages, fmt.Sprintf("正在结构化提取第 %d/%d 页的法律信息...", i+1, totalPages))
		}
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
func (c *BaiduClient) callBaiduAPI(fileData []byte, isPdf bool, onProgress ProgressCallback) ([]string, error) {
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

	// 开启心跳协程，在长耗时请求期间持续反馈进度，防止 UI “假死”
	done := make(chan bool)
	go func(cb ProgressCallback) {
		ticker := time.NewTicker(7 * time.Second)
		defer ticker.Stop()
		messages := []string{
			"正在进行深度语义识别...",
			"正在校对文档布局信息...",
			"正在提取法律实体关联...",
			"云端处理中，请稍候...",
		}
		i := 0
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if cb != nil {
					// 保持进度在 10% 左右摆动描述，直到 API 返回
					cb(1, 10, messages[i%len(messages)])
					i++
				}
			}
		}
	}(onProgress)

	apiStart := time.Now()
	resp, err := c.httpClient.Do(req)
	close(done) // 停止心跳
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	c.logger.Info("百度 API 响应接收成功", "status", resp.Status, "duration", time.Since(apiStart))

	// 增加状态码校验：非 200 状态码一律视为失败，触发重试
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("百度 API 响应异常 (HTTP %d)", resp.StatusCode)
	}

	var ocrResp BaiduOCRResponse
	// 使用流式解析，处理超大响应，减少内存压力
	if err := json.NewDecoder(resp.Body).Decode(&ocrResp); err != nil {
		c.logger.Error("JSON 解析失败", "error", err)
		return nil, fmt.Errorf("解析云端数据失败: %w", err)
	}

	if ocrResp.ErrorCode != 0 {
		return nil, fmt.Errorf("百度 API 错误 (%d): %s", ocrResp.ErrorCode, ocrResp.ErrorMsg)
	}

	var pages []string
	if len(ocrResp.Result.LayoutParsingResults) == 0 {
		c.logger.Warn("百度 API 返回结果为空")
	}
	for _, result := range ocrResp.Result.LayoutParsingResults {
		pages = append(pages, result.Markdown.Text)
	}
	return pages, nil
}
