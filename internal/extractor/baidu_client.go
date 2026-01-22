package extractor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"legal-extractor/internal/config"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// BaiduAPIError 百度 API 错误类型，提供友好的用户提示
type BaiduAPIError struct {
	Code    int
	Message string
	Hint    string // 用户友好提示
}

func (e *BaiduAPIError) Error() string {
	if e.Hint != "" {
		return e.Hint
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// 错误码映射表：将百度 API 错误码转换为用户友好提示
var errorCodeHints = map[int]string{
	17:     "API 调用次数已达今日上限，请明日再试或升级套餐",
	18:     "API 调用频率过快，请稍后重试",
	19:     "API 调用总量已达上限，请升级套餐",
	100:    "API 参数无效，请检查请求格式",
	110:    "Access Token 已过期，正在自动刷新...",
	111:    "Access Token 无效，请检查 API 密钥配置",
	216200: "图片内容为空，请检查文件是否损坏或为空白页",
	216201: "图片格式不支持，请使用 PDF/JPG/PNG 格式",
	216202: "图片大小超限，请压缩后重试（建议 < 10MB）",
	216630: "文档识别失败，请确保文档内容清晰可读",
	216631: "文档页数超限，请分割后重试",
	282000: "服务器内部错误，请稍后重试",
}

// translateError 将百度 API 错误转换为用户友好错误
func translateError(code int, msg string) error {
	hint, ok := errorCodeHints[code]
	if !ok {
		hint = fmt.Sprintf("百度 API 错误: %s", msg)
	}
	return &BaiduAPIError{Code: code, Message: msg, Hint: hint}
}

// isRetryableError 判断是否为可重试错误
func isRetryableError(code int) bool {
	retryableCodes := map[int]bool{
		18:     true, // QPS 限制
		110:    true, // Token 过期
		282000: true, // 服务器内部错误
	}
	return retryableCodes[code]
}

// BaiduClient 百度 AI 客户端
type BaiduClient struct {
	config     config.BaiduConfig
	httpClient *http.Client

	// Token 缓存相关
	accessToken string
	expireTime  time.Time
	mu          sync.RWMutex
}

// TokenResponse 百度鉴权响应结构
type TokenResponse struct {
	AccessToken   string `json:"access_token"`
	ExpiresIn     int64  `json:"expires_in"`
	Error         string `json:"error"`
	ErrorDesc     string `json:"error_description"`
}

// NewBaiduClient 创建一个新的百度 AI 客户端
func NewBaiduClient() *BaiduClient {
	return &BaiduClient{
		config: config.GetBaidu(),
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // 增加超时时间以应对大文件上传
		},
	}
}

// GetAccessToken 获取有效的 Access Token (带缓存机制)
func (c *BaiduClient) GetAccessToken() (string, error) {
	c.mu.RLock()
	if c.accessToken != "" && time.Now().Before(c.expireTime.Add(-5*time.Minute)) {
		token := c.accessToken
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expireTime.Add(-5*time.Minute)) {
		return c.accessToken, nil
	}

	token, expiresIn, err := c.fetchTokenFromAPI()
	if err != nil {
		return "", fmt.Errorf("获取百度 AccessToken 失败: %w", err)
	}

	c.accessToken = token
	c.expireTime = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return c.accessToken, nil
}

func (c *BaiduClient) fetchTokenFromAPI() (string, int64, error) {
	if c.config.APIKey == "" || c.config.SecretKey == "" {
		return "", 0, fmt.Errorf("百度 API Key 或 Secret Key 未配置，请检查 config/conf.yaml")
	}

	apiURL := fmt.Sprintf("https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%s&client_secret=%s",
		url.QueryEscape(c.config.APIKey), url.QueryEscape(c.config.SecretKey))

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return "", 0, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("请求百度鉴权接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("读取响应体失败: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", 0, fmt.Errorf("解析 Token JSON 失败: %w", err)
	}

	if tokenResp.Error != "" {
		return "", 0, fmt.Errorf("百度 API 错误: %s (%s)", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}

// TaskResponse 提交任务响应
type TaskResponse struct {
	LogID     string `json:"log_id"`
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	Result    struct {
		TaskID string `json:"task_id"`
	} `json:"result"`
}

// QueryResponse 查询任务响应
type QueryResponse struct {
	LogID     string `json:"log_id"`
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	Result    struct {
		Status      string `json:"status"`      // success, running, failed
		TaskError   string `json:"task_error"`  // 任务失败时的具体错误信息
		MarkdownURL string `json:"markdown_url"`
	} `json:"result"`
}

// ParseDocument 调用百度 PaddleOCR-VL 异步解析文档并返回 Markdown 结果
func (c *BaiduClient) ParseDocument(fileData []byte, fileName string) (string, error) {
	if len(fileData) == 0 {
		return "", &BaiduAPIError{Code: 0, Hint: "文件内容为空，请检查文件是否损坏"}
	}

	// 1. 转 Base64
	base64Data := base64.StdEncoding.EncodeToString(fileData)

	// 带重试的任务提交
	var taskID string
	maxSubmitRetries := 3

	for attempt := 1; attempt <= maxSubmitRetries; attempt++ {
		token, err := c.GetAccessToken()
		if err != nil {
			return "", err
		}

		// 2. 提交任务
		taskURL := "https://aip.baidubce.com/rest/2.0/brain/online/v2/paddle-vl-parser/task?access_token=" + token

		// 只传文件名，不传完整路径
		baseName := filepath.Base(fileName)

		// 构造 URL 编码后的 Payload 字符串
		payloadString := fmt.Sprintf("file_data=%s&file_url=&file_name=%s&analysis_chart=false",
			url.QueryEscape(base64Data),
			url.QueryEscape(baseName),
		)
		payload := strings.NewReader(payloadString)

		req, err := http.NewRequest("POST", taskURL, payload)
		if err != nil {
			return "", fmt.Errorf("创建提交任务请求失败: %w", err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt < maxSubmitRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return "", fmt.Errorf("网络连接失败，请检查网络后重试: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var taskResp TaskResponse
		if err := json.Unmarshal(body, &taskResp); err != nil {
			return "", fmt.Errorf("解析任务响应失败: %w", err)
		}

		if taskResp.ErrorCode != 0 {
			// Token 过期时清除缓存并重试
			if taskResp.ErrorCode == 110 || taskResp.ErrorCode == 111 {
				c.mu.Lock()
				c.accessToken = ""
				c.mu.Unlock()
				if attempt < maxSubmitRetries {
					continue
				}
			}
			// 可重试错误
			if isRetryableError(taskResp.ErrorCode) && attempt < maxSubmitRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return "", translateError(taskResp.ErrorCode, taskResp.ErrorMsg)
		}

		taskID = taskResp.Result.TaskID
		break
	}

	if taskID == "" {
		return "", &BaiduAPIError{Code: 0, Hint: "任务提交失败，请稍后重试"}
	}

	// 3. 轮询结果
	token, _ := c.GetAccessToken()
	queryURL := "https://aip.baidubce.com/rest/2.0/brain/online/v2/paddle-vl-parser/task/query?access_token=" + token
	maxRetries := 60 // 最多等待 120 秒
	var markdownURL string

	for range maxRetries {
		queryPayload := url.Values{}
		queryPayload.Set("task_id", taskID)

		qResp, err := c.httpClient.PostForm(queryURL, queryPayload)
		if err != nil {
			return "", fmt.Errorf("查询任务失败: %w", err)
		}

		qBody, _ := io.ReadAll(qResp.Body)
		qResp.Body.Close()

		var queryResp QueryResponse
		json.Unmarshal(qBody, &queryResp)

		if queryResp.Result.Status == "success" {
			markdownURL = queryResp.Result.MarkdownURL
			break
		} else if queryResp.Result.Status == "failed" {
			taskErr := queryResp.Result.TaskError
			if taskErr == "" {
				taskErr = "文档解析失败"
			}
			return "", &BaiduAPIError{Code: 0, Hint: fmt.Sprintf("文档解析失败: %s，请确保文档内容清晰可读", taskErr)}
		}

		time.Sleep(2 * time.Second)
	}

	if markdownURL == "" {
		return "", &BaiduAPIError{Code: 0, Hint: "文档解析超时，请尝试上传更小的文件或稍后重试"}
	}

	// 4. 下载 Markdown 内容
	mResp, err := c.httpClient.Get(markdownURL)
	if err != nil {
		return "", fmt.Errorf("下载解析结果失败: %w", err)
	}
	defer mResp.Body.Close()

	mBody, err := io.ReadAll(mResp.Body)
	if err != nil {
		return "", fmt.Errorf("读取解析结果失败: %w", err)
	}

	return string(mBody), nil
}
