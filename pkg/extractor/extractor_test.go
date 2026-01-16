package extractor

import (
	"testing"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestExtractData_PDF_Text(t *testing.T) {
	// Hardcoded absolute paths for simplicity based on user provided paths
	pdfPath := "/Users/fast/Documents/code/legal-extractor/1胡修富.pdf"

	records, err := ExtractData(pdfPath)
	if err != nil {
		t.Fatalf("ExtractData failed: %v", err)
	}

	if len(records) == 0 {
		t.Errorf("Expected records, got 0")
		// Debug: print text
		text, _ := extractTextFromPDF(pdfPath)
		t.Logf("Extracted Text Start: %q", text[:min(len(text), 500)])
	} else {
		t.Logf("Extracted %d records", len(records))
		if len(records) > 0 {
			t.Logf("First record: %+v", records[0])
		}
	}
}

func TestExtractData_PDF_Scanned_NoMCP(t *testing.T) {
	// This assumes MCP is NOT configured, so it should return empty/native text (which is empty)
	pdfPath := "/Users/fast/Documents/code/legal-extractor/廖剑丰_1.pdf"

	// We expect this to execute without error, but return possibly empty content or fallback to raw
	// The function returns []Record. If text is empty, parseCases returns empty list.
	records, err := ExtractData(pdfPath)
	if err != nil {
		t.Fatalf("ExtractData failed: %v", err)
	}

	// Since it's scanned and no OCR, we expect 0 records (or very few chars extracted so regex fails)
	if len(records) > 0 {
		t.Logf("Unexpectedly extracted records from scanned PDF without OCR: %d", len(records))
	} else {
		t.Log("Correctly obtained 0 records for scanned PDF without OCR")
	}
}
