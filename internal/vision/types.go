// Package vision provides image understanding capabilities via a vision service
// (Qwen2.5-VL / Florence-2). It detects UI layouts, identifies components,
// and generates semantic descriptions of images.
package vision

import "time"

// VisionResult represents the result of a vision service call.
type VisionResult struct {
	// Description is a semantic description of the entire image.
	Description string `json:"description"`
	// UIComponents are detected UI elements (buttons, inputs, etc.).
	UIComponents []UIComponent `json:"ui_components,omitempty"`
	// Layout describes the overall layout structure.
	Layout LayoutInfo `json:"layout,omitempty"`
	// Tags are relevant tags/categories for the image.
	Tags []string `json:"tags,omitempty"`
	// OCRProvider indicates which vision model processed this.
	VisionProvider string `json:"vision_provider,omitempty"`
	// ProcessingTime is how long the vision processing took.
	ProcessingTime time.Duration `json:"processing_time_ms,omitempty"`
	// Error contains any error message.
	Error string `json:"error,omitempty"`
}

// UIComponent represents a detected UI element in the image.
type UIComponent struct {
	// Type is the component type (e.g., "button", "input", "modal", "navbar").
	Type string `json:"type"`
	// Label is the visible text/label of the component.
	Label string `json:"label,omitempty"`
	// Description describes what this component does.
	Description string `json:"description,omitempty"`
	// BoundingBox is the position of the component in the image.
	BoundingBox [4][2]int `json:"bounding_box,omitempty"`
}

// LayoutInfo describes the overall layout structure of the image.
type LayoutInfo struct {
	// Type is the layout type (e.g., "form", "dashboard", "article", "screenshot").
	Type string `json:"type"`
	// Regions describe the main regions/sections of the layout.
	Regions []LayoutRegion `json:"regions,omitempty"`
	// Description is a human-readable description of the layout.
	Description string `json:"description,omitempty"`
}

// LayoutRegion describes a region/section of the layout.
type LayoutRegion struct {
	// Name is the region name (e.g., "header", "sidebar", "content").
	Name string `json:"name"`
	// Description describes what this region contains.
	Description string `json:"description"`
	// BoundingBox is the position of the region.
	BoundingBox [4][2]int `json:"bounding_box,omitempty"`
}

// DescribeRequest is sent to the vision service.
type DescribeRequest struct {
	// DetailLevel controls how detailed the description should be.
	// Options: "basic", "detailed", "ui".
	DetailLevel string `json:"detail_level,omitempty"`
}

// DescribeResponse mirrors the JSON response from the vision service.
type DescribeResponse struct {
	Description   string         `json:"description"`
	UIComponents  []uiComponentJSON `json:"ui_components,omitempty"`
	Layout        layoutInfoJSON `json:"layout,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
	Error         string         `json:"error,omitempty"`
}

type uiComponentJSON struct {
	Type        string    `json:"type"`
	Label       string    `json:"label,omitempty"`
	Description string    `json:"description,omitempty"`
	BoundingBox [4][2]int `json:"bounding_box,omitempty"`
}

type layoutInfoJSON struct {
	Type        string             `json:"type"`
	Regions     []layoutRegionJSON `json:"regions,omitempty"`
	Description string             `json:"description,omitempty"`
}

type layoutRegionJSON struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	BoundingBox [4][2]int `json:"bounding_box,omitempty"`
}
