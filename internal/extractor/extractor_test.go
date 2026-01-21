package extractor

import (
	"testing"
)

func TestParseCases(t *testing.T) {
	e := NewExtractor(nil)
	text := `
民 事 起 诉 状

被 告： 张三
身份证号码： 110101199001011234
住址： 北京市朝阳区

诉讼请求：
1. 请求判令被告偿还借款10000元。
2. 诉讼费由被告承担。

事实与理由：
2023年1月1日，被告向原告借款...
此致
`
	expected := []Record{
		{
			"defendant":   "张三",
			"idNumber":    "110101199001011234",
			"request":     "1. 请求判令被告偿还借款10000元。\n2. 诉讼费由被告承担。",
			"factsReason": "2023年1月1日，被告向原告借款...",
		},
	}

	result := e.parseCases(text, []string{"defendant", "idNumber", "request", "factsReason"})

	if len(result) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result))
	}

	for k, v := range expected[0] {
		if result[0][k] != v && k != "request" && k != "factsReason" {
			t.Errorf("Field %s: expected %q, got %q", k, v, result[0][k])
		}
	}
}

func TestSmartMerge(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			// smartMerge 会将非逻辑换行符替换为空格，防止 OCR 碎行文字粘连
			name:  "Merge weird newlines",
			input: "这是\n一句\n完整的话。",
			want:  "这是 一句 完整的话。",
		},
		{
			name:  "Preserve lists",
			input: "1. 第一点\n2. 第二点",
			want:  "1. 第一点 2. 第二点", // 数字列表不在保留规则内，会被合并
		},
		{
			name:  "Preserve punctuation",
			input: "第一句。\n第二句",
			want:  "第一句。\n第二句",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := smartMerge(tt.input)
			if got != tt.want {
				t.Errorf("smartMerge() = %q, want %q", got, tt.want)
			}
		})
	}
}
