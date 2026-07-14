package formatter

import (
	"strings"
	"testing"
	"time"

	"github.com/kush/ocr-mcp/internal/ocr"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input string
		want  OutputFormat
	}{
		{"markdown", FormatMarkdown},
		{"json", FormatJSON},
		{"", FormatMarkdown},
		{"MARKDOWN", FormatMarkdown},
		{"JSON", FormatJSON},
		{"unknown", FormatMarkdown},
	}
	for _, tt := range tests {
		got := ParseFormat(tt.input)
		if got != tt.want {
			t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestFormatMarkdown(t *testing.T) {
	result := &ocr.OCRResult{
		Text:           "Hello World",
		Confidence:     0.95,
		OCRProvider:    "TestProvider",
		ProcessingTime: 42 * time.Millisecond,
		Blocks: []ocr.TextBlock{
			{Text: "Hello", Confidence: 0.98, BoundingBox: [4][2]int{{0, 0}, {50, 0}, {50, 20}, {0, 20}}},
			{Text: "World", Confidence: 0.92, BoundingBox: [4][2]int{{0, 25}, {60, 25}, {60, 45}, {0, 45}}},
		},
	}

	output, err := FormatString(result, "markdown")
	if err != nil {
		t.Fatalf("FormatString() error = %v", err)
	}

	checks := []string{
		"# OCR Result",
		"TestProvider",
		"42ms",
		"## Extracted Text",
		"Hello World",
		"## Text Blocks",
		"## Tables",
		"## Confidence",
		"95.0%",
		"Very High",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Markdown output missing: %q", check)
		}
	}
}

func TestFormatJSON(t *testing.T) {
	result := &ocr.OCRResult{
		Text:       "Hello",
		Confidence: 0.95,
		OCRProvider: "Test",
	}

	output, err := FormatString(result, "json")
	if err != nil {
		t.Fatalf("FormatString() error = %v", err)
	}

	checks := []string{
		`"text": "Hello"`,
		`"confidence": 0.95`,
		`"confidence_pct": 95`,
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("JSON output missing: %q\nOutput: %s", check, output)
		}
	}
}

func TestFormatMarkdownEmpty(t *testing.T) {
	result := &ocr.OCRResult{
		Text:       "",
		Confidence: 0.0,
		Blocks:     []ocr.TextBlock{},
		Tables:     []ocr.Table{},
	}

	output, err := FormatString(result, "markdown")
	if err != nil {
		t.Fatalf("FormatString() error = %v", err)
	}

	if !strings.Contains(output, "No text detected") {
		t.Error("expected 'No text detected' for empty result")
	}
	if !strings.Contains(output, "No tables detected") {
		t.Error("expected 'No tables detected' for empty tables")
	}
}

func TestFormatTable(t *testing.T) {
	result := &ocr.OCRResult{
		Text: "Data",
		Tables: []ocr.Table{
			{Rows: [][]string{{"Name", "Value"}, {"A", "1"}, {"B", "2"}}},
		},
	}

	output, _ := FormatString(result, "markdown")
	if !strings.Contains(output, "| Name | Value |") {
		t.Errorf("Table header missing: %s", output)
	}
}

func TestFormatNil(t *testing.T) {
	_, err := Format(nil, FormatMarkdown)
	if err == nil {
		t.Error("expected error for nil result")
	}
}

func TestConfidenceLabel(t *testing.T) {
	tests := []struct {
		confidence float64
		want       string
	}{
		{0.98, "Very High"},
		{0.90, "High"},
		{0.75, "Medium"},
		{0.60, "Low"},
		{0.30, "Very Low"},
	}
	for _, tt := range tests {
		got := confidenceLabel(tt.confidence)
		if got != tt.want {
			t.Errorf("confidenceLabel(%v) = %q, want %q", tt.confidence, got, tt.want)
		}
	}
}
