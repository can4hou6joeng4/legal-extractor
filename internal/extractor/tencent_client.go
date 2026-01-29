package extractor

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"legal-extractor/internal/config"
	"net/http"
	"strings"
	"time"
)

// TencentAPIError 腾讯云 API 错误类型
type TencentAPIError struct {
	Code    string
	Message string
	Hint    string
}

func (e *TencentAPIError) Error() string {
	if e.Hint != "" {
		return e.Hint
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 腾讯云错误码映射表
var tencentErrorHints = map[string]string{
	"AuthFailure.SignatureFailure":    "签名验证失败，请检查 SecretId 和 SecretKey 配置",
	"AuthFailure.SecretIdNotFound":    "SecretId 不存在，请检查配置",
	"AuthFailure.SignatureExpire":     "签名已过期，请检查系统时间是否正确",
	"LimitExceeded.TooLargeFileError": "文件过大，请压缩后重试（建议 < 7MB）",
	"FailedOperation.ImageDecodeFailed": "图片解码失败，请检查文件是否损坏",
	"FailedOperation.OcrFailed":       "OCR 识别失败，请确保文档内容清晰可读",
	"FailedOperation.UnKnowError":     "服务器内部错误，请稍后重试",
	"InvalidParameterValue.InvalidFileType": "不支持的文件格式，请使用 PDF/JPG/PNG",
	"ResourceUnavailable.InArrears":   "账户欠费，请充值后重试",
	"RequestLimitExceeded":            "请求频率超限，请稍后重试",
}

// translateTencentError 将腾讯云错误转换为用户友好错误
func translateTencentError(code, msg string) error {
	hint, ok := tencentErrorHints[code]
	if !ok {
		hint = fmt.Sprintf("腾讯云 API 错误: %s", msg)
	}
	return &TencentAPIError{Code: code, Message: msg, Hint: hint}
}

// TencentClient 腾讯云 OCR 客户端
type TencentClient struct {
	config     config.TencentConfig
	httpClient *http.Client
}

// 法律文书标准提取字段 - 固定 ItemNames 提升性能
var LegalDocItemNames = []string{
	"被告",
	"身份证号码",
	"诉讼请求",
	"事实与理由",
}

// tencentFieldMapping 腾讯云字段名 → 项目字段 Key
var tencentFieldMapping = map[string]string{
	"被告":       "defendant",
	"被告人":     "defendant",
	"身份证号码": "idNumber",
	"身份证":     "idNumber",
	"诉讼请求":   "request",
	"事实与理由": "factsReason",
	"事实和理由": "factsReason",
}

// ========== 腾讯云 API 响应结构 ==========

// TencentOCRResponse 腾讯云 SmartStructuralOCRV2 响应
type TencentOCRResponse struct {
	Response struct {
		Angle          float64          `json:"Angle"`
		StructuralList []StructuralItem `json:"StructuralList"`
		WordList       []WordItem       `json:"WordList"`
		SealInfos      []SealInfo       `json:"SealInfos"`
		RequestId      string           `json:"RequestId"`
		Error          *TencentRespError `json:"Error,omitempty"`
	} `json:"Response"`
}

type TencentRespError struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type StructuralItem struct {
	Groups []GroupItem `json:"Groups"`
}

type GroupItem struct {
	Lines []LineItem `json:"Lines"`
}

type LineItem struct {
	Key   KeyInfo   `json:"Key"`
	Value ValueInfo `json:"Value"`
}

type KeyInfo struct {
	AutoName string `json:"AutoName"`
}

type ValueInfo struct {
	AutoContent string `json:"AutoContent"`
}

type WordItem struct {
	DetectedText string `json:"DetectedText"`
}

type SealInfo struct {
	SealBody string `json:"SealBody"`
}

// ========== 客户端方法 ==========

// NewTencentClient 创建腾讯云 OCR 客户端
func NewTencentClient() *TencentClient {
	return &TencentClient{
		config: config.GetTencent(),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ParseDocument 调用腾讯云结构化 OCR 解析文档指定页码
func (c *TencentClient) ParseDocument(fileData []byte, pageNumber int) (Record, error) {
	if len(fileData) == 0 {
		return nil, &TencentAPIError{Code: "InvalidParameter", Hint: "文件内容为空，请检查文件是否损坏"}
	}

	if c.config.SecretId == "" || c.config.SecretKey == "" {
		return nil, &TencentAPIError{Code: "ConfigError", Hint: "腾讯云 SecretId 或 SecretKey 未配置，请检查 config/conf.yaml"}
	}

	// 1. 构建请求体
	pdfBase64 := base64.StdEncoding.EncodeToString(fileData)
	requestBody := map[string]interface{}{
		"ImageBase64":         pdfBase64,
		"ItemNames":           LegalDocItemNames,
		"IsPdf":               true,
		"PdfPageNumber":       pageNumber,
		"EnableSealRecognize": true,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 2. 生成签名并发送请求
	resp, err := c.doRequest(bodyBytes)
	if err != nil {
		return nil, err
	}

	// 3. 检查 API 错误
	if resp.Response.Error != nil {
		return nil, translateTencentError(resp.Response.Error.Code, resp.Response.Error.Message)
	}

	// 4. 解析结构化结果
	record := c.parseStructuralList(resp)

	if len(record) == 0 {
		return nil, &TencentAPIError{Code: "NoData", Hint: "未能从文档中提取到有效字段"}
	}

	return record, nil
}

// parseStructuralList 解析腾讯云返回的结构化数据
func (c *TencentClient) parseStructuralList(resp *TencentOCRResponse) Record {
	result := make(Record)

	// 解析 StructuralList
	for _, item := range resp.Response.StructuralList {
		for _, group := range item.Groups {
			for _, line := range group.Lines {
				key := strings.TrimSpace(line.Key.AutoName)
				value := strings.TrimSpace(line.Value.AutoContent)

				if key == "" || value == "" {
					continue
				}

				// 映射到项目字段
				if mappedKey, ok := tencentFieldMapping[key]; ok {
					// 清理字段值
					cleanedValue := cleanFieldValue(mappedKey, value)

					// 如果字段已存在，追加内容（处理多行情况）
					if existing, exists := result[mappedKey]; exists {
						result[mappedKey] = existing + "\n" + cleanedValue
					} else {
						result[mappedKey] = cleanedValue
					}
				}
			}
		}
	}

	// 解析印章信息（如果需要）
	if len(resp.Response.SealInfos) > 0 {
		var seals []string
		for _, seal := range resp.Response.SealInfos {
			if seal.SealBody != "" {
				seals = append(seals, seal.SealBody)
			}
		}
		if len(seals) > 0 {
			result["seals"] = strings.Join(seals, "; ")
		}
	}

	return result
}

// doRequest 执行 HTTP 请求（含 TC3 签名）
func (c *TencentClient) doRequest(body []byte) (*TencentOCRResponse, error) {
	const (
		host      = "ocr.tencentcloudapi.com"
		service   = "ocr"
		version   = "2018-11-19"
		action    = "SmartStructuralOCRV2"
		algorithm = "TC3-HMAC-SHA256"
	)

	// 时间戳
	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")

	// ========== 步骤 1: 拼接规范请求串 ==========
	httpRequestMethod := "POST"
	canonicalURI := "/"
	canonicalQueryString := ""
	contentType := "application/json; charset=utf-8"
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-tc-action:%s\n",
		contentType, host, strings.ToLower(action))
	signedHeaders := "content-type;host;x-tc-action"
	hashedRequestPayload := sha256Hex(body)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod, canonicalURI, canonicalQueryString,
		canonicalHeaders, signedHeaders, hashedRequestPayload)

	// ========== 步骤 2: 拼接待签名字符串 ==========
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	hashedCanonicalRequest := sha256Hex([]byte(canonicalRequest))
	stringToSign := fmt.Sprintf("%s\n%d\n%s\n%s",
		algorithm, timestamp, credentialScope, hashedCanonicalRequest)

	// ========== 步骤 3: 计算签名 ==========
	secretDate := hmacSHA256([]byte("TC3"+c.config.SecretKey), date)
	secretService := hmacSHA256(secretDate, service)
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(secretSigning, stringToSign))

	// ========== 步骤 4: 拼接 Authorization ==========
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm, c.config.SecretId, credentialScope, signedHeaders, signature)

	// ========== 发送请求 ==========
	req, err := http.NewRequest("POST", "https://"+host, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Host", host)
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Version", version)
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("Authorization", authorization)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("网络请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var ocrResp TencentOCRResponse
	if err := json.Unmarshal(respBody, &ocrResp); err != nil {
		return nil, fmt.Errorf("解析响应 JSON 失败: %w", err)
	}

	return &ocrResp, nil
}

// ========== 签名辅助函数 ==========

func sha256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// cleanFieldValue 根据字段类型清理值中的多余字符
func cleanFieldValue(fieldKey, value string) string {
	value = strings.TrimSpace(value)

	switch fieldKey {
	case "defendant":
		// 被告字段：移除末尾的标点符号（顿号、逗号、冒号等）
		value = strings.TrimRight(value, "、，,：:;；。.")
		// 移除开头的标点
		value = strings.TrimLeft(value, "、，,：:;；")
	case "idNumber":
		// 身份证号：只保留数字和X
		value = cleanIDNumber(value)
	}

	return value
}

// cleanIDNumber 清理身份证号中的非法字符
func cleanIDNumber(value string) string {
	var result strings.Builder
	for _, r := range value {
		if (r >= '0' && r <= '9') || r == 'X' || r == 'x' {
			if r == 'x' {
				result.WriteRune('X')
			} else {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}
