package extractor

import (
	"testing"
)

func TestCleanFieldValue(t *testing.T) {
	tests := []struct {
		name     string
		fieldKey string
		input    string
		expected string
	}{
		{
			name:     "被告字段移除末尾顿号",
			fieldKey: "defendant",
			input:    "廖剑丰、",
			expected: "廖剑丰",
		},
		{
			name:     "被告字段移除末尾逗号",
			fieldKey: "defendant",
			input:    "张三，",
			expected: "张三",
		},
		{
			name:     "被告字段移除多个末尾标点",
			fieldKey: "defendant",
			input:    "李四、，",
			expected: "李四",
		},
		{
			name:     "被告字段移除开头标点",
			fieldKey: "defendant",
			input:    "、王五",
			expected: "王五",
		},
		{
			name:     "被告字段无需清理",
			fieldKey: "defendant",
			input:    "赵六",
			expected: "赵六",
		},
		{
			name:     "身份证号保持不变",
			fieldKey: "idNumber",
			input:    "44142419820208097X",
			expected: "44142419820208097X",
		},
		{
			name:     "身份证号小写x转大写",
			fieldKey: "idNumber",
			input:    "44142419820208097x",
			expected: "44142419820208097X",
		},
		{
			name:     "身份证号移除非法字符",
			fieldKey: "idNumber",
			input:    "441424 1982 0208 097X",
			expected: "44142419820208097X",
		},
		{
			name:     "其他字段不处理",
			fieldKey: "request",
			input:    "请求内容、",
			expected: "请求内容、",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanFieldValue(tt.fieldKey, tt.input)
			if result != tt.expected {
				t.Errorf("cleanFieldValue(%q, %q) = %q, want %q",
					tt.fieldKey, tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanIDNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"44142419820208097X", "44142419820208097X"},
		{"44142419820208097x", "44142419820208097X"},
		{"441424 1982 0208 097X", "44142419820208097X"},
		{"4414-2419-8202-0809-7X", "44142419820208097X"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := cleanIDNumber(tt.input)
			if result != tt.expected {
				t.Errorf("cleanIDNumber(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTencentFieldMapping(t *testing.T) {
	// 验证字段映射表完整性
	expectedMappings := map[string]string{
		"被告":       "defendant",
		"被告人":     "defendant",
		"身份证号码": "idNumber",
		"身份证":     "idNumber",
		"诉讼请求":   "request",
		"事实与理由": "factsReason",
		"事实和理由": "factsReason",
	}

	for key, expectedValue := range expectedMappings {
		if actualValue, ok := tencentFieldMapping[key]; !ok {
			t.Errorf("缺少字段映射: %q", key)
		} else if actualValue != expectedValue {
			t.Errorf("字段映射错误: %q -> %q, 期望 %q", key, actualValue, expectedValue)
		}
	}
}

func TestLegalDocItemNames(t *testing.T) {
	// 验证固定的 ItemNames 包含必要字段
	expectedItems := []string{"被告", "身份证号码", "诉讼请求", "事实与理由"}

	if len(LegalDocItemNames) != len(expectedItems) {
		t.Errorf("LegalDocItemNames 长度错误: got %d, want %d",
			len(LegalDocItemNames), len(expectedItems))
	}

	for i, expected := range expectedItems {
		if LegalDocItemNames[i] != expected {
			t.Errorf("LegalDocItemNames[%d] = %q, want %q", i, LegalDocItemNames[i], expected)
		}
	}
}
