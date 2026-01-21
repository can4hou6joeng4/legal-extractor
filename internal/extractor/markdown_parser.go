package extractor

import (
	"regexp"
	"strings"
)

// ParseMarkdown 针对 PaddleOCR-VL 优化的解析器
func ParseMarkdown(markdown string) []Record {
	if markdown == "" {
		return nil
	}

	// 1. 预处理：剔除所有 HTML 标签 (VLM 经常返回 div/img)
	cleanMd := stripHTML(markdown)
	record := make(Record)

	// 2. 按标题和常见关键词切分
	// 增加对常见法律文书关键词的切分支持，增加“此致”作为结束标志
	delimiters := []string{"#", "诉讼请求", "事实与理由", "事实和理由", "此致"}
	content := cleanMd
	for _, d := range delimiters {
		content = strings.ReplaceAll(content, d, "\n[SEP]"+d)
	}
	sections := strings.Split(content, "[SEP]")

	for _, section := range sections {
		trimmed := strings.TrimSpace(section)
		if trimmed == "" || strings.HasPrefix(trimmed, "此致") {
			continue
		}
		lowered := strings.ToLower(trimmed)

		if strings.Contains(lowered, "被告") || strings.Contains(lowered, "当事人") {
			if record["defendant"] == "" {
				record["defendant"] = extractField(trimmed, "被告")
			}
		}
		if strings.Contains(lowered, "诉讼请求") {
			record["request"] = cleanMarkdown(trimmed)
		}
		if strings.Contains(lowered, "事实") && (strings.Contains(lowered, "理由") || strings.Contains(lowered, "事实经过")) {
			record["factsReason"] = cleanMarkdown(trimmed)
		}
	}

	// 3. 兜底全局匹配
	if record["defendant"] == "" {
		record["defendant"] = extractField(cleanMd, "被告")
	}
	if record["idNumber"] == "" {
		// 使用 patterns.go 中定义的身份证号正则
		match := DefaultPatterns.ID.FindStringSubmatch(cleanMd)
		if len(match) > 1 {
			record["idNumber"] = strings.TrimSpace(match[1])
		}
	}

	// 只有当至少有一个字段有值时才返回记录
	hasData := false
	for _, v := range record {
		if v != "" {
			hasData = true
			break
		}
	}

	if hasData {
		return []Record{record}
	}

	return nil
}

// stripHTML 使用正则剥离所有 HTML 标签
func stripHTML(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

// cleanMarkdown 移除 Markdown 格式符号，保持纯文本整洁
func cleanMarkdown(s string) string {
	// 移除标题符
	s = strings.ReplaceAll(s, "#", "")
	// 移除粗体符
	s = strings.ReplaceAll(s, "**", "")
	// 移除表格线
	s = strings.ReplaceAll(s, "|", " ")
	s = strings.ReplaceAll(s, "---", "")
	s = strings.ReplaceAll(s, "&nbsp;", " ")

	// 移除关键词头部，防止内容中重复出现标题
	reHeader := regexp.MustCompile(`^(?i)(诉讼请求|事实与理由|事实和理由|事实经过)[:：\s]*`)
	s = reHeader.ReplaceAllString(s, "")

	// 规范化换行和空格
	return smartMerge(s)
}

// extractField 从行中提取关键字段
func extractField(text, keyword string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.Contains(line, keyword) {
			// 尝试分割冒号
			parts := regexp.MustCompile(`[:：]`).Split(line, 2)
			val := ""
			if len(parts) > 1 {
				val = strings.TrimSpace(parts[1])
			}

			// 如果本行没内容，尝试看下一行（处理换行排版）
			if val == "" && i+1 < len(lines) {
				val = strings.TrimSpace(lines[i+1])
			}

			if val != "" {
				// 再次利用 DefEnd 正则清理多余后缀
				locEnd := DefaultPatterns.DefEnd.FindStringIndex(val)
				if locEnd != nil {
					val = val[:locEnd[0]]
				}
				return strings.Trim(val, " ,，、;；")
			}
		}
	}
	return ""
}
