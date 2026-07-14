// Package formatter provides output formatting for OCR results.
// It supports Markdown (for human/LLM consumption) and JSON (for programmatic use).
package formatter

import (
	"fmt"
	"strings"

	"github.com/kush/ocr-mcp/internal/ocr"
)

// OutputFormat represents the type of output formatting to apply.
type OutputFormat string

const (
	// FormatMarkdown produces human-readable Markdown output.
	FormatMarkdown OutputFormat = "markdown"
	// FormatJSON produces structured JSON output.
	FormatJSON OutputFormat = "json"
)

// ParseFormat parses an output format string, defaulting to Markdown.
func ParseFormat(s string) OutputFormat {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "json":
		return FormatJSON
	case "markdown", "":
		return FormatMarkdown
	default:
		return FormatMarkdown
	}
}

// Format formats an OCR result into the specified output format.
// Returns the formatted string and any error encountered.
func Format(result *ocr.OCRResult, format OutputFormat) (string, error) {
	if result == nil {
		return "", fmt.Errorf("cannot format nil OCR result")
	}

	switch format {
	case FormatJSON:
		return formatJSON(result), nil
	case FormatMarkdown:
		return formatMarkdown(result), nil
	default:
		return formatMarkdown(result), nil
	}
}

// FormatString formats an OCR result based on a string format type.
func FormatString(result *ocr.OCRResult, formatType string) (string, error) {
	return Format(result, ParseFormat(formatType))
}
