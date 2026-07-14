// Package configs provides configuration loading and management for the OCR MCP server.
package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the OCR MCP server.
type Config struct {
	// Server settings
	ServerPort int           `yaml:"server_port" env:"SERVER_PORT"`
	ServerHost string        `yaml:"server_host" env:"SERVER_HOST"`
	Shutdown   time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT"`

	// OCR Service
	OCRServiceURL  string        `yaml:"ocr_service_url" env:"OCR_SERVICE_URL"`
	OCRServicePort int           `yaml:"ocr_service_port" env:"OCR_SERVICE_PORT"`
	OCRTimeout     time.Duration `yaml:"ocr_timeout" env:"OCR_TIMEOUT"`
	OCRMaxRetries  int           `yaml:"ocr_max_retries" env:"OCR_MAX_RETRIES"`

	// Cache settings
	CacheType     string        `yaml:"cache_type" env:"CACHE_TYPE"`
	CacheTTL      time.Duration `yaml:"cache_ttl" env:"CACHE_TTL"`
	CacheMaxSize  int           `yaml:"cache_max_size" env:"CACHE_MAX_SIZE"`
	RedisURL      string        `yaml:"redis_url" env:"REDIS_URL"`
	RedisPassword string        `yaml:"redis_password" env:"REDIS_PASSWORD"`
	RedisDB       int           `yaml:"redis_db" env:"REDIS_DB"`

	// Worker pool
	WorkerCount int           `yaml:"worker_count" env:"WORKER_COUNT"`
	QueueSize   int           `yaml:"queue_size" env:"QUEUE_SIZE"`
	JobTimeout  time.Duration `yaml:"job_timeout" env:"JOB_TIMEOUT"`

	// Image preprocessing
	MaxImageWidth  int  `yaml:"max_image_width" env:"MAX_IMAGE_WIDTH"`
	MaxImageHeight int  `yaml:"max_image_height" env:"MAX_IMAGE_HEIGHT"`
	MaxImageSizeMB int  `yaml:"max_image_size_mb" env:"MAX_IMAGE_SIZE_MB"`
	AutoPreprocess bool `yaml:"auto_preprocess" env:"AUTO_PREPROCESS"`

	// Output
	OutputFormat string `yaml:"output_format" env:"OUTPUT_FORMAT"`

	// Auth
	APIKey string `yaml:"api_key" env:"API_KEY"`

	// Rate limiting
	RateLimitPerMin int `yaml:"rate_limit_per_min" env:"RATE_LIMIT_PER_MIN"`

	// Logging
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL"`

	// Vision service (optional)
	VisionServiceURL string `yaml:"vision_service_url" env:"VISION_SERVICE_URL"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		ServerPort: 7070,
		ServerHost: "0.0.0.0",
		Shutdown:   10 * time.Second,

		OCRServiceURL:  "http://localhost",
		OCRServicePort: 9090,
		OCRTimeout:     30 * time.Second,
		OCRMaxRetries:  3,

		CacheType:    "memory",
		CacheTTL:     1 * time.Hour,
		CacheMaxSize: 1000,
		RedisDB:      0,

		WorkerCount: 4,
		QueueSize:   100,
		JobTimeout:  60 * time.Second,

		MaxImageWidth:  4096,
		MaxImageHeight: 4096,
		MaxImageSizeMB: 20,
		AutoPreprocess: true,

		OutputFormat: "markdown",

		RateLimitPerMin: 60,

		LogLevel: "info",
	}
}

// LoadConfig loads configuration from environment variables,
// falling back to defaults where not set.
func LoadConfig() *Config {
	cfg := DefaultConfig()

	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.ServerPort = port
		}
	}
	if v := os.Getenv("SERVER_HOST"); v != "" {
		cfg.ServerHost = v
	}
	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Shutdown = d
		}
	}

	// OCR Service
	if v := os.Getenv("OCR_SERVICE_URL"); v != "" {
		cfg.OCRServiceURL = v
	}
	if v := os.Getenv("OCR_SERVICE_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.OCRServicePort = port
		}
	}
	if v := os.Getenv("OCR_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.OCRTimeout = d
		}
	}
	if v := os.Getenv("OCR_MAX_RETRIES"); v != "" {
		if retries, err := strconv.Atoi(v); err == nil {
			cfg.OCRMaxRetries = retries
		}
	}

	// Cache
	if v := os.Getenv("CACHE_TYPE"); v != "" {
		cfg.CacheType = v
	}
	if v := os.Getenv("CACHE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.CacheTTL = d
		}
	}
	if v := os.Getenv("CACHE_MAX_SIZE"); v != "" {
		if size, err := strconv.Atoi(v); err == nil {
			cfg.CacheMaxSize = size
		}
	}
	if v := os.Getenv("REDIS_URL"); v != "" {
		cfg.RedisURL = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.RedisPassword = v
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			cfg.RedisDB = db
		}
	}

	// Worker pool
	if v := os.Getenv("WORKER_COUNT"); v != "" {
		if count, err := strconv.Atoi(v); err == nil {
			cfg.WorkerCount = count
		}
	}
	if v := os.Getenv("QUEUE_SIZE"); v != "" {
		if size, err := strconv.Atoi(v); err == nil {
			cfg.QueueSize = size
		}
	}
	if v := os.Getenv("JOB_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.JobTimeout = d
		}
	}

	// Image preprocessing
	if v := os.Getenv("MAX_IMAGE_WIDTH"); v != "" {
		if w, err := strconv.Atoi(v); err == nil {
			cfg.MaxImageWidth = w
		}
	}
	if v := os.Getenv("MAX_IMAGE_HEIGHT"); v != "" {
		if h, err := strconv.Atoi(v); err == nil {
			cfg.MaxImageHeight = h
		}
	}
	if v := os.Getenv("MAX_IMAGE_SIZE_MB"); v != "" {
		if mb, err := strconv.Atoi(v); err == nil {
			cfg.MaxImageSizeMB = mb
		}
	}
	if v := os.Getenv("AUTO_PREPROCESS"); v != "" {
		cfg.AutoPreprocess = v == "true" || v == "1" || v == "yes"
	}

	// Output
	if v := os.Getenv("OUTPUT_FORMAT"); v != "" {
		cfg.OutputFormat = v
	}

	// Auth
	if v := os.Getenv("API_KEY"); v != "" {
		cfg.APIKey = v
	}

	// Rate limiting
	if v := os.Getenv("RATE_LIMIT_PER_MIN"); v != "" {
		if limit, err := strconv.Atoi(v); err == nil {
			cfg.RateLimitPerMin = limit
		}
	}

	// Logging
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}

	// Vision service
	if v := os.Getenv("VISION_SERVICE_URL"); v != "" {
		cfg.VisionServiceURL = v
	}

	return cfg
}

// OCRServiceAddr returns the full address of the OCR service.
func (c *Config) OCRServiceAddr() string {
	return fmt.Sprintf("%s:%d", c.OCRServiceURL, c.OCRServicePort)
}

// ServerAddr returns the address the MCP server should listen on.
func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}
