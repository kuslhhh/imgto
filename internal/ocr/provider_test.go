package ocr

import (
	"context"
	"errors"
	"testing"
)

// mockProvider implements OCRProvider for testing.
type mockProvider struct {
	name      string
	result    *OCRResult
	err       error
}

func (m *mockProvider) ExtractText(_ context.Context, _ []byte) (*OCRResult, error) {
	return m.result, m.err
}

func (m *mockProvider) Name() string { return m.name }

func TestMockProvider(t *testing.T) {
	mock := &mockProvider{
		name: "TestOCR",
		result: &OCRResult{
			Text:       "Hello World",
			Confidence: 0.95,
		},
	}

	t.Run("Name", func(t *testing.T) {
		if got := mock.Name(); got != "TestOCR" {
			t.Errorf("Name() = %q, want %q", got, "TestOCR")
		}
	})

	t.Run("ExtractText", func(t *testing.T) {
		result, err := mock.ExtractText(context.Background(), []byte("fake-image"))
		if err != nil {
			t.Fatalf("ExtractText() error = %v", err)
		}
		if result.Text != "Hello World" {
			t.Errorf("ExtractText().Text = %q, want %q", result.Text, "Hello World")
		}
		if result.Confidence != 0.95 {
			t.Errorf("ExtractText().Confidence = %v, want %v", result.Confidence, 0.95)
		}
	})
}

func TestMockProviderError(t *testing.T) {
	mock := &mockProvider{
		name: "ErrorOCR",
		err:  errors.New("connection refused"),
	}

	_, err := mock.ExtractText(context.Background(), []byte("test"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "connection refused" {
		t.Errorf("error = %q, want %q", err.Error(), "connection refused")
	}
}

func TestOCRResultDefaults(t *testing.T) {
	result := &OCRResult{}
	if result.Text != "" {
		t.Errorf("expected empty text, got %q", result.Text)
	}
	if result.Confidence != 0 {
		t.Errorf("expected 0 confidence, got %v", result.Confidence)
	}
	if result.Blocks != nil {
		t.Error("expected nil blocks")
	}
}

func TestTextBlock(t *testing.T) {
	block := TextBlock{
		Text:       "test",
		Confidence: 0.9,
		BoundingBox: [4][2]int{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
	}
	if block.Text != "test" {
		t.Errorf("Text = %q", block.Text)
	}
	if block.Confidence != 0.9 {
		t.Errorf("Confidence = %v", block.Confidence)
	}
}

func TestTable(t *testing.T) {
	table := Table{
		Rows: [][]string{{"A", "B"}, {"1", "2"}},
	}
	if len(table.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(table.Rows))
	}
	if table.Rows[0][0] != "A" {
		t.Errorf("expected A, got %s", table.Rows[0][0])
	}
}

func TestProviderConfig(t *testing.T) {
	cfg := ProviderConfig{
		Type: "paddleocr",
	}
	if cfg.Type != "paddleocr" {
		t.Errorf("Type = %q", cfg.Type)
	}
}
