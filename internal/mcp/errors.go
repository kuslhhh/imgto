package mcp

import (
	"context"
	"errors"
	"fmt"
)

// ErrorCode represents a machine-readable error category for MCP responses.
type ErrorCode string

const (
	// ErrCodeValidation indicates invalid input from the caller.
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"
	// ErrCodeOCRFailed indicates the OCR service returned an error.
	ErrCodeOCRFailed ErrorCode = "OCR_FAILED"
	// ErrCodeCache indicates a cache operation failed.
	ErrCodeCache ErrorCode = "CACHE_ERROR"
	// ErrCodeTimeout indicates the operation exceeded its deadline.
	ErrCodeTimeout ErrorCode = "TIMEOUT"
	// ErrCodeInternal indicates an unexpected server error.
	ErrCodeInternal ErrorCode = "INTERNAL_ERROR"
	// ErrCodeImageTooLarge indicates the image exceeds size limits.
	ErrCodeImageTooLarge ErrorCode = "IMAGE_TOO_LARGE"
	// ErrCodeUnsupportedFormat indicates the image format is not supported.
	ErrCodeUnsupportedFormat ErrorCode = "UNSUPPORTED_FORMAT"
	// ErrCodeProviderNotFound indicates no OCR provider is configured.
	ErrCodeProviderNotFound ErrorCode = "PROVIDER_NOT_FOUND"
	// ErrCodeVisionFailed indicates the vision service returned an error.
	ErrCodeVisionFailed ErrorCode = "VISION_FAILED"
)

// ToolError is a structured error that maps to a user-facing MCP error result.
type ToolError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

func (e *ToolError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap implements errors.Unwrap for potential future wrapped errors.
func (e *ToolError) Unwrap() error {
	return nil
}

// NewToolError creates a new ToolError.
func NewToolError(code ErrorCode, message string) *ToolError {
	return &ToolError{Code: code, Message: message}
}

// NewToolErrorf creates a new ToolError with a formatted message.
func NewToolErrorf(code ErrorCode, format string, args ...interface{}) *ToolError {
	return &ToolError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// WithDetails adds contextual details to a ToolError.
func (e *ToolError) WithDetails(details string) *ToolError {
	e.Details = details
	return e
}

// Common tool errors as pre-built sentinels.
var (
	ErrImageRequired     = NewToolError(ErrCodeValidation, "image_data is required")
	ErrImageTooLarge     = NewToolError(ErrCodeImageTooLarge, "image exceeds maximum allowed size")
	ErrUnsupportedFormat = NewToolError(ErrCodeUnsupportedFormat, "unsupported image format")
	ErrOCRFailed         = NewToolError(ErrCodeOCRFailed, "OCR processing failed")
	ErrCacheFailed       = NewToolError(ErrCodeCache, "cache operation failed")
	ErrProviderNotFound  = NewToolError(ErrCodeProviderNotFound, "no OCR provider configured")
	ErrTimeout           = NewToolError(ErrCodeTimeout, "operation timed out")
	ErrInternal          = NewToolError(ErrCodeInternal, "internal server error")
)

// IsCancelledOrTimeout checks if a context error is cancellation or deadline exceeded.
func IsCancelledOrTimeout(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if ctx.Err() != nil {
		return true
	}
	return false
}
