// Package workers provides a configurable worker pool for concurrent
// image processing. It limits the number of simultaneous OCR calls
// and provides timeout and graceful shutdown support.
package workers

import (
	"context"
	"time"

	"github.com/kush/ocr-mcp/internal/ocr"
)

// Job represents a unit of work to be processed by the worker pool.
type Job struct {
	// ID is a unique identifier for the job (for logging).
	ID string

	// Image is the preprocessed image bytes to run OCR on.
	Image []byte

	// Timeout is the maximum duration allowed for this job.
	Timeout time.Duration

	// Result is a channel where the result will be sent.
	// The channel will be closed by the worker after sending the result.
	Result chan<- *JobResult

	// Ctx is the parent context for cancellation propagation.
	Ctx context.Context
}

// JobResult holds the outcome of a worker pool job.
type JobResult struct {
	// Result is the OCR result (nil on error).
	Result *ocr.OCRResult

	// Err is any error that occurred during processing.
	Err error
}

