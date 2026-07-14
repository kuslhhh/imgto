package formatter

import (
	"encoding/json"
	"fmt"

	"github.com/kush/ocr-mcp/internal/ocr"
)

// formatJSON converts an OCR result into a structured JSON string.
// This is suitable for programmatic consumption and API responses.
func formatJSON(result *ocr.OCRResult) string {
	// Build a clean output structure with all relevant fields
	output := jsonOutput{
		Text:           result.Text,
		Confidence:     result.Confidence,
		ConfidencePct:  result.Confidence * 100,
		Blocks:         make([]jsonBlock, len(result.Blocks)),
		Tables:         make([]jsonTable, len(result.Tables)),
		Provider:       result.OCRProvider,
		ProcessingTime: result.ProcessingTime.Milliseconds(),
	}

	for i, block := range result.Blocks {
		output.Blocks[i] = jsonBlock{
			Text:        block.Text,
			Confidence:  block.Confidence,
			BoundingBox: block.BoundingBox,
		}
	}

	for i, table := range result.Tables {
		output.Tables[i] = jsonTable{Rows: table.Rows}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to serialize OCR result: %s"}`, err.Error())
	}

	return string(data)
}

// jsonOutput is the structured JSON output format.
type jsonOutput struct {
	Text           string      `json:"text"`
	Confidence     float64     `json:"confidence"`
	ConfidencePct  float64     `json:"confidence_pct"`
	Blocks         []jsonBlock `json:"blocks,omitempty"`
	Tables         []jsonTable `json:"tables,omitempty"`
	Provider       string      `json:"provider,omitempty"`
	ProcessingTime int64       `json:"processing_time_ms,omitempty"`
}

type jsonBlock struct {
	Text        string    `json:"text"`
	Confidence  float64   `json:"confidence"`
	BoundingBox [4][2]int `json:"bounding_box"`
}

type jsonTable struct {
	Rows [][]string `json:"rows"`
}
