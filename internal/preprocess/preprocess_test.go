package preprocess

import (
	"testing"
)

func TestParseMode(t *testing.T) {
	tests := []struct {
		input string
		want  PreprocessMode
	}{
		{"auto", ModeAuto},
		{"none", ModeNone},
		{"grayscale", ModeGrayscale},
		{"threshold", ModeThreshold},
		{"denoise", ModeDenoise},
		{"AUTO", ModeAuto},
		{"GRAYSCALE", ModeGrayscale},
		{"", ModeAuto},
		{"unknown", ModeAuto},
	}
	for _, tt := range tests {
		got := ParseMode(tt.input)
		if got != tt.want {
			t.Errorf("ParseMode(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestOptionsDefaults(t *testing.T) {
	opts := DefaultOptions()
	if opts.Mode != ModeAuto {
		t.Errorf("default Mode = %v, want auto", opts.Mode)
	}
	if opts.MaxWidth != 4096 {
		t.Errorf("default MaxWidth = %d, want 4096", opts.MaxWidth)
	}
	if opts.MaxHeight != 4096 {
		t.Errorf("default MaxHeight = %d, want 4096", opts.MaxHeight)
	}
	if !opts.AutoPreprocess {
		t.Error("expected AutoPreprocess to be true")
	}
}

func TestOptionsValidate(t *testing.T) {
	opts := DefaultOptions()
	if err := opts.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	opts.MaxWidth = -1
	if err := opts.Validate(); err == nil {
		t.Error("expected error for negative MaxWidth")
	}
}

func TestSupportedImageFormats(t *testing.T) {
	formats := SupportedImageFormats()
	expected := []string{"jpg", "jpeg", "png", "gif", "bmp", "tiff", "tif", "webp"}
	if len(formats) != len(expected) {
		t.Errorf("got %d formats, want %d", len(formats), len(expected))
	}
	for _, e := range expected {
		found := false
		for _, f := range formats {
			if f == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing format: %s", e)
		}
	}
}

func TestIsSupportedFormat(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{".jpg", true},
		{".jpeg", true},
		{".png", true},
		{".gif", true},
		{".bmp", true},
		{".tiff", true},
		{".tif", true},
		{".webp", true},
		{".pdf", false},
		{".txt", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsSupportedFormat(tt.ext)
		if got != tt.want {
			t.Errorf("IsSupportedFormat(%q) = %v, want %v", tt.ext, got, tt.want)
		}
	}
}
