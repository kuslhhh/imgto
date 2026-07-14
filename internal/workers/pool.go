package workers

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/kush/ocr-mcp/internal/ocr"
)

// Pool is a bounded worker pool for processing OCR jobs concurrently.
// It limits the number of simultaneous OCR calls to avoid overwhelming
// the OCR service and provides per-job timeouts.
type Pool struct {
	workerCount int
	jobQueue    chan Job
	provider    ocr.OCRProvider
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	stopped     bool
	mu          sync.Mutex
}

// Config holds configuration for the worker pool.
type Config struct {
	// WorkerCount is the number of worker goroutines (default: runtime.NumCPU()).
	WorkerCount int

	// QueueSize is the maximum number of jobs that can be queued.
	QueueSize int

	// JobTimeout is the default timeout for each job.
	JobTimeout time.Duration

	// OCRProvider is the OCR provider to use for processing.
	OCRProvider ocr.OCRProvider
}

// NewPool creates a new worker pool but does not start it.
// Call Start() to begin processing jobs.
func NewPool(cfg Config) (*Pool, error) {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = runtime.NumCPU()
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 100
	}
	if cfg.JobTimeout <= 0 {
		cfg.JobTimeout = 60 * time.Second
	}
	if cfg.OCRProvider == nil {
		return nil, fmt.Errorf("workers: OCR provider is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		workerCount: cfg.WorkerCount,
		jobQueue:    make(chan Job, cfg.QueueSize),
		provider:    cfg.OCRProvider,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Start launches the worker goroutines. Each worker reads from the
// shared job queue and processes jobs independently.
func (p *Pool) Start() {
	slog.Info("starting worker pool",
		slog.Int("workers", p.workerCount),
		slog.Int("queue_capacity", cap(p.jobQueue)),
	)

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// Submit sends a job to the pool for processing.
// It blocks if the queue is full until the job is accepted or the context
// is cancelled. Returns an error if the pool is stopped.
func (p *Pool) Submit(job Job) error {
	p.mu.Lock()
	if p.stopped {
		p.mu.Unlock()
		return fmt.Errorf("workers: pool is stopped")
	}
	p.mu.Unlock()

	select {
	case p.jobQueue <- job:
		return nil
	case <-job.Ctx.Done():
		return job.Ctx.Err()
	case <-p.ctx.Done():
		return fmt.Errorf("workers: pool is shutting down")
	}
}

// Shutdown gracefully stops the pool. It stops accepting new jobs and
// waits for all in-flight jobs to complete. If a timeout is provided,
// it will force-stop after the timeout.
func (p *Pool) Shutdown(timeout time.Duration) error {
	p.mu.Lock()
	p.stopped = true
	p.mu.Unlock()

	// Stop accepting new jobs
	p.cancel()

	if timeout > 0 {
		// Wait with timeout
		done := make(chan struct{})
		go func() {
			p.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			slog.Info("worker pool shut down gracefully")
			return nil
		case <-time.After(timeout):
			return fmt.Errorf("workers: forced shutdown after %v", timeout)
		}
	}

	// Wait indefinitely
	p.wg.Wait()
	slog.Info("worker pool shut down gracefully")
	return nil
}

// Running returns the number of active workers.
func (p *Pool) Running() int {
	return p.workerCount
}

// QueueLength returns the current number of queued jobs.
func (p *Pool) QueueLength() int {
	return len(p.jobQueue)
}

// worker is a single worker goroutine that processes jobs from the queue.
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	slog.Debug("worker started", slog.Int("id", id))

	for {
		select {
		case <-p.ctx.Done():
			slog.Debug("worker stopped", slog.Int("id", id))
			return

		case job, ok := <-p.jobQueue:
			if !ok {
				// Queue closed
				return
			}
			p.processJob(id, job)
		}
	}
}

// processJob handles a single job with timeout and error handling.
func (p *Pool) processJob(workerID int, job Job) {
	timeout := job.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	// Create a context with timeout
	jobCtx, cancel := context.WithTimeout(job.Ctx, timeout)
	defer cancel()

	slog.Debug("worker processing job",
		slog.Int("worker_id", workerID),
		slog.String("job_id", job.ID),
		slog.Duration("timeout", timeout),
	)

	// Run OCR
	result, err := p.provider.ExtractText(jobCtx, job.Image)

	// Send result back (non-blocking send with context check)
	resultCh := job.Result
	if resultCh != nil {
		select {
		case resultCh <- &JobResult{Result: result, Err: err}:
		case <-job.Ctx.Done():
			// Caller already cancelled, result not needed
			slog.Debug("job cancelled before result delivered",
				slog.String("job_id", job.ID),
			)
		default:
			// Channel full or closed — log warning
			slog.Warn("job result channel full/closed, dropping result",
				slog.String("job_id", job.ID),
			)
		}
		close(resultCh)
	}

	slog.Debug("worker completed job",
		slog.Int("worker_id", workerID),
		slog.String("job_id", job.ID),
		slog.Bool("success", err == nil),
	)
}
