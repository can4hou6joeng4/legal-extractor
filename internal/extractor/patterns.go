package extractor

import "regexp"

// ExtractionPatterns holds the regex patterns used for parsing
type ExtractionPatterns struct {
	Split       *regexp.Regexp
	DefStart    *regexp.Regexp
	DefEnd      *regexp.Regexp
	DefFallback *regexp.Regexp
	ID          *regexp.Regexp
	Request     *regexp.Regexp
	Facts       *regexp.Regexp
}

// DefaultPatterns defines the standard patterns for legal documents
var DefaultPatterns = ExtractionPatterns{
	Split:       regexp.MustCompile(`民\s*事\s*起\s*诉\s*状`),
	DefStart:    regexp.MustCompile(`被\s*告\s*[:：]`),
	DefEnd:      regexp.MustCompile(`[,，、\s]+(?:性\s*别|生\s*日|身\s*份\s*证|住\s*址|联\s*系\s*电\s*话)|\n|$`),
	DefFallback: regexp.MustCompile(`被\s*告\s*[:：]\s*(.*?)\n`),
	ID:          regexp.MustCompile(`身\s*份\s*证\s*号\s*码\s*[:：]\s*([\dX]+)`),
	Request:     regexp.MustCompile(`(?s)诉\s*讼\s*请\s*求\s*[:：]\s*(.*?)\s*事\s*实\s*与\s*理\s*由`),
	Facts:       regexp.MustCompile(`(?s)事\s*实\s*与\s*理\s*由\s*[:：]\s*(.*?)\s*此\s*致`),
}

// PatternRegistry maps field names to their respective patterns
var PatternRegistry = map[string]struct {
	Label   string
	Pattern *regexp.Regexp
}{
	"defendant":   {Label: "被告", Pattern: DefaultPatterns.DefStart},
	"idNumber":    {Label: "身份证号码", Pattern: DefaultPatterns.ID},
	"request":     {Label: "诉讼请求", Pattern: DefaultPatterns.Request},
	"factsReason": {Label: "事实与理由", Pattern: DefaultPatterns.Facts},
}
