package extractor

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

// TestFileNameExtraction 测试文件名提取逻辑
func TestFileNameExtraction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "完整 Unix 路径",
			input:    "/Users/fast/Documents/code/legal-extractor/test.pdf",
			expected: "test.pdf",
		},
		{
			name:     "包含中文的文件名",
			input:    "/Users/fast/Documents/刘玄一_民事起诉状.pdf",
			expected: "刘玄一_民事起诉状.pdf",
		},
		{
			name:     "仅文件名",
			input:    "document.pdf",
			expected: "document.pdf",
		},
		{
			name:     "带空格的路径",
			input:    "/Users/fast/My Documents/test file.pdf",
			expected: "test file.pdf",
		},
		{
			name:     "深层嵌套路径",
			input:    "/a/b/c/d/e/f/g/test.pdf",
			expected: "test.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filepath.Base(tt.input)
			if got != tt.expected {
				t.Errorf("filepath.Base(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestQueryResponseParsing 测试百度 API 响应解析
func TestQueryResponseParsing(t *testing.T) {
	tests := []struct {
		name           string
		jsonResponse   string
		expectedStatus string
		expectedError  string
	}{
		{
			name: "成功响应",
			jsonResponse: `{
				"log_id": "123",
				"error_code": 0,
				"error_msg": "",
				"result": {
					"status": "success",
					"task_error": "",
					"markdown_url": "https://example.com/result.md"
				}
			}`,
			expectedStatus: "success",
			expectedError:  "",
		},
		{
			name: "任务执行失败",
			jsonResponse: `{
				"log_id": "456",
				"error_code": 0,
				"error_msg": "",
				"result": {
					"status": "failed",
					"task_error": "parse document task failed",
					"markdown_url": ""
				}
			}`,
			expectedStatus: "failed",
			expectedError:  "parse document task failed",
		},
		{
			name: "任务进行中",
			jsonResponse: `{
				"log_id": "789",
				"error_code": 0,
				"error_msg": "",
				"result": {
					"status": "running",
					"task_error": "",
					"markdown_url": ""
				}
			}`,
			expectedStatus: "running",
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp QueryResponse
			err := json.Unmarshal([]byte(tt.jsonResponse), &resp)
			if err != nil {
				t.Fatalf("JSON 解析失败: %v", err)
			}

			if resp.Result.Status != tt.expectedStatus {
				t.Errorf("Status = %q, want %q", resp.Result.Status, tt.expectedStatus)
			}

			if resp.Result.TaskError != tt.expectedError {
				t.Errorf("TaskError = %q, want %q", resp.Result.TaskError, tt.expectedError)
			}
		})
	}
}

// TestTaskResponseParsing 测试任务提交响应解析
func TestTaskResponseParsing(t *testing.T) {
	tests := []struct {
		name          string
		jsonResponse  string
		expectedCode  int
		expectedID    string
		shouldHaveErr bool
	}{
		{
			name: "成功提交任务",
			jsonResponse: `{
				"log_id": "123",
				"error_code": 0,
				"error_msg": "",
				"result": {
					"task_id": "task-abc123"
				}
			}`,
			expectedCode:  0,
			expectedID:    "task-abc123",
			shouldHaveErr: false,
		},
		{
			name: "API 错误 - 图片为空",
			jsonResponse: `{
				"log_id": "456",
				"error_code": 216200,
				"error_msg": "image is empty",
				"result": {}
			}`,
			expectedCode:  216200,
			expectedID:    "",
			shouldHaveErr: true,
		},
		{
			name: "API 错误 - 配额不足",
			jsonResponse: `{
				"log_id": "789",
				"error_code": 17,
				"error_msg": "Open api daily request limit reached",
				"result": {}
			}`,
			expectedCode:  17,
			expectedID:    "",
			shouldHaveErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp TaskResponse
			err := json.Unmarshal([]byte(tt.jsonResponse), &resp)
			if err != nil {
				t.Fatalf("JSON 解析失败: %v", err)
			}

			if resp.ErrorCode != tt.expectedCode {
				t.Errorf("ErrorCode = %d, want %d", resp.ErrorCode, tt.expectedCode)
			}

			if resp.Result.TaskID != tt.expectedID {
				t.Errorf("TaskID = %q, want %q", resp.Result.TaskID, tt.expectedID)
			}

			hasErr := resp.ErrorCode != 0
			if hasErr != tt.shouldHaveErr {
				t.Errorf("hasError = %v, want %v", hasErr, tt.shouldHaveErr)
			}
		})
	}
}

// TestEmptyFileDataCheck 测试空文件数据检查
func TestEmptyFileDataCheck(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		shouldError bool
	}{
		{
			name:        "空数据",
			data:        []byte{},
			shouldError: true,
		},
		{
			name:        "nil 数据",
			data:        nil,
			shouldError: true,
		},
		{
			name:        "有效数据",
			data:        []byte{0x25, 0x50, 0x44, 0x46}, // PDF magic bytes
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := len(tt.data) == 0
			if isEmpty != tt.shouldError {
				t.Errorf("len(data)==0 = %v, want %v", isEmpty, tt.shouldError)
			}
		})
	}
}
