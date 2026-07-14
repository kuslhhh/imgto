package mcp

import (
	"expvar"
	"sync/atomic"
)

// Metrics holds atomic counters for server monitoring.
// These are exposed via the /metrics endpoint and expvar.
type Metrics struct {
	RequestsTotal    atomic.Int64
	RequestsActive   atomic.Int64
	CacheHits        atomic.Int64
	CacheMisses      atomic.Int64
	OCRCalls         atomic.Int64
	OCRErrors        atomic.Int64
	ProcessingErrors atomic.Int64
	RateLimited      atomic.Int64
	AuthFailures     atomic.Int64
}

// NewMetrics creates and registers metrics with expvar.
func NewMetrics() *Metrics {
	m := &Metrics{}
	m.register()
	return m
}

func (m *Metrics) register() {
	expvar.Publish("requests_total", expvar.Func(func() interface{} { return m.RequestsTotal.Load() }))
	expvar.Publish("requests_active", expvar.Func(func() interface{} { return m.RequestsActive.Load() }))
	expvar.Publish("cache_hits", expvar.Func(func() interface{} { return m.CacheHits.Load() }))
	expvar.Publish("cache_misses", expvar.Func(func() interface{} { return m.CacheMisses.Load() }))
	expvar.Publish("ocr_calls", expvar.Func(func() interface{} { return m.OCRCalls.Load() }))
	expvar.Publish("ocr_errors", expvar.Func(func() interface{} { return m.OCRErrors.Load() }))
	expvar.Publish("processing_errors", expvar.Func(func() interface{} { return m.ProcessingErrors.Load() }))
	expvar.Publish("rate_limited", expvar.Func(func() interface{} { return m.RateLimited.Load() }))
	expvar.Publish("auth_failures", expvar.Func(func() interface{} { return m.AuthFailures.Load() }))
}
