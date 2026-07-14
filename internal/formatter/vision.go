package formatter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kush/ocr-mcp/internal/vision"
)

// FormatVision converts a VisionResult into a formatted Markdown string.
// The detail level controls how much information is included.
func FormatVision(result *vision.VisionResult, detailLevel string) string {
	var b strings.Builder

	// Title
	b.WriteString("# Image Description\n\n")

	// Provider metadata
	b.WriteString(fmt.Sprintf("> **Provider:** %s", result.VisionProvider))
	if result.ProcessingTime > 0 {
		b.WriteString(fmt.Sprintf(" | **Time:** %dms", result.ProcessingTime.Milliseconds()))
	}
	b.WriteString("\n\n")

	// --- Description ---
	b.WriteString("## Description\n\n")
	if result.Description != "" {
		b.WriteString(result.Description)
		b.WriteString("\n\n")
	} else {
		b.WriteString("_No description available._\n\n")
	}

	// --- Tags ---
	if len(result.Tags) > 0 {
		b.WriteString("## Tags\n\n")
		for _, tag := range result.Tags {
			b.WriteString(fmt.Sprintf("- `%s`\n", tag))
		}
		b.WriteString("\n")
	}

	// --- UI Components (shown for "detailed" and "ui" levels) ---
	if len(result.UIComponents) > 0 {
		b.WriteString("## UI Components\n\n")
		b.WriteString("| # | Type | Label | Description |\n")
		b.WriteString("|---|------|-------|-------------|\n")

		for i, comp := range result.UIComponents {
			label := truncate(comp.Label, 40)
			desc := truncate(comp.Description, 50)
			if desc == "" {
				desc = "—"
			}
			b.WriteString(fmt.Sprintf("| %d | `%s` | %s | %s |\n", i+1, comp.Type, label, desc))
		}
		b.WriteString("\n")
	}

	// --- Layout ---
	b.WriteString("## Layout\n\n")
	if result.Layout.Type != "" || result.Layout.Description != "" {
		if result.Layout.Type != "" {
			b.WriteString(fmt.Sprintf("- **Type:** `%s`\n", result.Layout.Type))
		}
		if result.Layout.Description != "" {
			b.WriteString(fmt.Sprintf("- **Structure:** %s\n", result.Layout.Description))
		}

		if len(result.Layout.Regions) > 0 {
			b.WriteString("\n### Regions\n\n")
			b.WriteString("| # | Name | Description |\n")
			b.WriteString("|---|------|-------------|\n")
			for i, region := range result.Layout.Regions {
				desc := truncate(region.Description, 60)
				b.WriteString(fmt.Sprintf("| %d | `%s` | %s |\n", i+1, region.Name, desc))
			}
		}
		b.WriteString("\n")
	} else {
		b.WriteString("_No layout data available._\n\n")
	}

	// --- Notes ---
	b.WriteString("---\n\n")
	b.WriteString(fmt.Sprintf("_Processed by OCR MCP Server using %s._\n",
		result.VisionProvider))

	return b.String()
}

// visionJSONOutput is the JSON output structure for vision results.
type visionJSONOutput struct {
	Description    string              `json:"description"`
	Provider       string              `json:"provider,omitempty"`
	ProcessingTime int64               `json:"processing_time_ms,omitempty"`
	Tags           []string            `json:"tags,omitempty"`
	UIComponents   []uiComponentJSONOut `json:"ui_components,omitempty"`
	Layout         layoutJSONOut       `json:"layout"`
}

type uiComponentJSONOut struct {
	Type        string    `json:"type"`
	Label       string    `json:"label,omitempty"`
	Description string    `json:"description,omitempty"`
	BoundingBox [4][2]int `json:"bounding_box,omitempty"`
}

type layoutJSONOut struct {
	Type        string          `json:"type,omitempty"`
	Description string          `json:"description,omitempty"`
	Regions     []regionJSONOut `json:"regions,omitempty"`
}

type regionJSONOut struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	BoundingBox [4][2]int `json:"bounding_box,omitempty"`
}

// FormatVisionJSON converts a VisionResult into a structured JSON string.
// Uses proper JSON serialization via encoding/json.
func FormatVisionJSON(result *vision.VisionResult) string {
	output := visionJSONOutput{
		Description:    result.Description,
		Provider:       result.VisionProvider,
		ProcessingTime: result.ProcessingTime.Milliseconds(),
		Tags:           result.Tags,
		UIComponents:   make([]uiComponentJSONOut, len(result.UIComponents)),
		Layout: layoutJSONOut{
			Type:        result.Layout.Type,
			Description: result.Layout.Description,
			Regions:     make([]regionJSONOut, len(result.Layout.Regions)),
		},
	}

	for i, comp := range result.UIComponents {
		output.UIComponents[i] = uiComponentJSONOut{
			Type:         comp.Type,
			Label:        comp.Label,
			Description:  comp.Description,
			BoundingBox:  comp.BoundingBox,
		}
	}

	for i, reg := range result.Layout.Regions {
		output.Layout.Regions[i] = regionJSONOut{
			Name:        reg.Name,
			Description: reg.Description,
			BoundingBox: reg.BoundingBox,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to serialize vision result: %s"}`, err.Error())
	}

	return string(data)
}


