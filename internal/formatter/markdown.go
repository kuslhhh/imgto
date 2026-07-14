package formatter

import (
	"fmt"
	"strings"

	"github.com/kush/ocr-mcp/internal/ocr"
)

// formatMarkdown converts an OCR result into a rich Markdown document.
// The format is designed to improve reasoning quality for text-only LLMs
// by providing structured sections with clear hierarchy.
func formatMarkdown(result *ocr.OCRResult) string {
	var b strings.Builder

	// Title
	b.WriteString("# OCR Result\n\n")

	// Provider metadata
	b.WriteString(fmt.Sprintf("> **Provider:** %s", result.OCRProvider))
	if result.ProcessingTime > 0 {
		b.WriteString(fmt.Sprintf(" | **Time:** %dms", result.ProcessingTime.Milliseconds()))
	}
	b.WriteString("\n\n")

	// --- Extracted Text ---
	b.WriteString("## Extracted Text\n\n")
	if result.Text != "" {
		b.WriteString(result.Text)
		b.WriteString("\n\n")
	} else {
		b.WriteString("_No text detected._\n\n")
	}

	// --- Text Blocks (if available) ---
	if len(result.Blocks) > 0 {
		b.WriteString("## Text Blocks\n\n")
		b.WriteString("| # | Text | Confidence | Position |\n")
		b.WriteString("|---|------|------------|----------|\n")

		for i, block := range result.Blocks {
			label := fmt.Sprintf("%d", i+1)
			text := truncate(block.Text, 60)
			conf := fmt.Sprintf("%.1f%%", block.Confidence*100)
			pos := formatBBox(block.BoundingBox)

			b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", label, text, conf, pos))
		}
		b.WriteString("\n")
	}

	// --- Tables ---
	b.WriteString("## Tables\n\n")
	if len(result.Tables) > 0 {
		for i, table := range result.Tables {
			if len(table.Rows) > 0 {
				b.WriteString(fmt.Sprintf("### Table %d\n\n", i+1))

				// Render as Markdown table
				for rowIdx, row := range table.Rows {
					if rowIdx == 0 {
						// Header row
						b.WriteString("| " + strings.Join(row, " | ") + " |\n")
						// Separator row
						separators := make([]string, len(row))
						for j := range row {
							separators[j] = "---"
						}
						b.WriteString("| " + strings.Join(separators, " | ") + " |\n")
					} else {
						b.WriteString("| " + strings.Join(row, " | ") + " |\n")
					}
				}
				b.WriteString("\n")
			}
		}
	} else {
		b.WriteString("_No tables detected._\n\n")
	}

	// --- Confidence ---
	b.WriteString("## Confidence\n\n")
	confidenceLevel := confidenceLabel(result.Confidence)
	b.WriteString(fmt.Sprintf("**Overall:** %.1f%% (%s)\n\n", result.Confidence*100, confidenceLevel))

	if len(result.Blocks) > 0 {
		// Show min/max confidence range
		minConf, maxConf := 1.0, 0.0
		for _, block := range result.Blocks {
			if block.Confidence < minConf {
				minConf = block.Confidence
			}
			if block.Confidence > maxConf {
				maxConf = block.Confidence
			}
		}
		b.WriteString(fmt.Sprintf("**Range:** %.1f%% – %.1f%%\n\n", minConf*100, maxConf*100))
	}

	// --- Layout Summary ---
	b.WriteString("## Layout\n\n")
	if len(result.Blocks) > 0 {
		blocks := result.Blocks
		// Sort blocks by vertical position (rough top-to-bottom)
		// Simple summary based on block positions
		totalHeight := 0
		minY := int(^uint(0) >> 1) // max int
		maxY := 0
		for _, block := range blocks {
			for _, pt := range block.BoundingBox {
				if pt[1] < minY {
					minY = pt[1]
				}
				if pt[1] > maxY {
					maxY = pt[1]
				}
			}
		}
		totalHeight = maxY - minY
		if totalHeight < 0 {
			totalHeight = 0
		}

		b.WriteString(fmt.Sprintf("- **Elements detected:** %d\n", len(blocks)))
		b.WriteString(fmt.Sprintf("- **Height span:** %dpx\n", totalHeight))
		if totalHeight > 0 {
			b.WriteString(fmt.Sprintf("- **Text density:** ~%.1f elements per 100px\n\n",
				float64(len(blocks))/float64(totalHeight)*100))
		}
	} else {
		b.WriteString("_No layout data available._\n\n")
	}

	// Notes
	b.WriteString("---\n\n")
	b.WriteString(fmt.Sprintf("_Processed by OCR MCP Server using %s._\n", result.OCRProvider))

	return b.String()
}

// formatBBox converts a bounding box to a compact string representation.
func formatBBox(bbox [4][2]int) string {
	if len(bbox) < 4 {
		return "N/A"
	}
	return fmt.Sprintf("(%d,%d)-(%d,%d)",
		bbox[0][0], bbox[0][1],
		bbox[2][0], bbox[2][1])
}

// truncate truncates a string to the given max length with ellipsis.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// confidenceLabel returns a human-readable label for a confidence score.
func confidenceLabel(confidence float64) string {
	switch {
	case confidence >= 0.95:
		return "Very High"
	case confidence >= 0.85:
		return "High"
	case confidence >= 0.70:
		return "Medium"
	case confidence >= 0.50:
		return "Low"
	default:
		return "Very Low"
	}
}

