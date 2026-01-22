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
		return "", fmt.Errorf("文件数据为空，无法提交到百度 API")
	}

	// 1. 转 Base64
	base64Data := base64.StdEncoding.EncodeToString(fileData)

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
		return "", fmt.Errorf("提交任务 HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var taskResp TaskResponse
	if err := json.Unmarshal(body, &taskResp); err != nil {
		return "", fmt.Errorf("解析任务响应失败: %w", err)
	}

	if taskResp.ErrorCode != 0 {
		return "", fmt.Errorf("提交任务业务错误: [%d] %s", taskResp.ErrorCode, taskResp.ErrorMsg)
	}

	taskID := taskResp.Result.TaskID

	// 3. 轮询结果
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
			return "", fmt.Errorf("百度解析任务执行失败: %s", queryResp.Result.TaskError)
		}

		time.Sleep(2 * time.Second)
	}

	if markdownURL == "" {
		return "", fmt.Errorf("解析任务超时")
	}

	// 4. 下载 Markdown 内容
	mResp, err := c.httpClient.Get(markdownURL)
	if err != nil {
		return "", fmt.Errorf("下载 Markdown 失败: %w", err)
	}
	defer mResp.Body.Close()

	mBody, err := io.ReadAll(mResp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 Markdown 内容失败: %w", err)
	}

	return string(mBody), nil
}
