package workers

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kush/ocr-mcp/internal/ocr"
)

// mockOCRProvider implements ocr.OCRProvider for testing.
type mockOCRProvider struct {
	name       string
	delay      time.Duration
	shouldFail bool
	callCount  atomic.Int64
}

func (m *mockOCRProvider) ExtractText(_ context.Context, _ []byte) (*ocr.OCRResult, error) {
	m.callCount.Add(1)
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.shouldFail {
		return nil, errors.New("ocr failed")
	}
	return &ocr.OCRResult{
		Text:       "test result",
		Confidence: 0.95,
	}, nil
}

func (m *mockOCRProvider) Name() string { return m.name }

func TestPoolSubmitAndResult(t *testing.T) {
	mock := &mockOCRProvider{name: "test"}
	pool, err := NewPool(Config{
		WorkerCount: 2,
		QueueSize:   10,
		JobTimeout:  5 * time.Second,
		OCRProvider: mock,
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	pool.Start()
	defer pool.Shutdown(2 * time.Second)

	resultCh := make(chan *JobResult, 1)
	job := Job{
		ID:      "test-1",
		Image:   []byte("test-image"),
		Timeout: 5 * time.Second,
		Result:  resultCh,
		Ctx:     context.Background(),
	}

	if err := pool.Submit(job); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	select {
	case result := <-resultCh:
		if result.Err != nil {
			t.Fatalf("job error = %v", result.Err)
		}
		if result.Result.Text != "test result" {
			t.Errorf("Text = %q, want %q", result.Result.Text, "test result")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for job result")
	}
}

func TestPoolError(t *testing.T) {
	mock := &mockOCRProvider{name: "err", shouldFail: true}
	pool, _ := NewPool(Config{
		WorkerCount: 2,
		QueueSize:   10,
		JobTimeout:  5 * time.Second,
		OCRProvider: mock,
	})
	pool.Start()
	defer pool.Shutdown(2 * time.Second)

	resultCh := make(chan *JobResult, 1)
	pool.Submit(Job{
		ID: "err-1", Image: []byte("img"),
		Result: resultCh, Ctx: context.Background(),
	})

	result := <-resultCh
	if result.Err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPoolNoProvider(t *testing.T) {
	_, err := NewPool(Config{})
	if err == nil {
		t.Error("expected error for nil provider")
	}
}

func TestPoolMultipleJobs(t *testing.T) {
	mock := &mockOCRProvider{name: "multi"}
	pool, _ := NewPool(Config{
		WorkerCount: 4,
		QueueSize:   20,
		JobTimeout:  5 * time.Second,
		OCRProvider: mock,
	})
	pool.Start()
	defer pool.Shutdown(2 * time.Second)

	for i := 0; i < 5; i++ {
		resultCh := make(chan *JobResult, 1)
		pool.Submit(Job{
			ID:     "job-1",
			Image:  []byte("img"),
			Result: resultCh,
			Ctx:    context.Background(),
		})
		<-resultCh
	}

	if count := mock.callCount.Load(); count != 5 {
		t.Errorf("expected 5 calls, got %d", count)
	}
}
