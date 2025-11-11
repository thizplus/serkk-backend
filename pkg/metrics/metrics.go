package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds application metrics
type Metrics struct {
	// Request metrics
	totalRequests     uint64
	successfulReqs    uint64
	failedReqs        uint64

	// Response time metrics
	totalResponseTime uint64 // in milliseconds
	minResponseTime   uint64
	maxResponseTime   uint64

	// Status code counters
	status2xx uint64
	status3xx uint64
	status4xx uint64
	status5xx uint64

	// Error counters
	totalErrors uint64

	// Custom counters
	customCounters sync.Map

	// Start time
	startTime time.Time

	mu sync.RWMutex
}

// MetricsSnapshot represents a snapshot of metrics
type MetricsSnapshot struct {
	TotalRequests     uint64            `json:"total_requests"`
	SuccessfulReqs    uint64            `json:"successful_requests"`
	FailedReqs        uint64            `json:"failed_requests"`
	AvgResponseTime   float64           `json:"avg_response_time_ms"`
	MinResponseTime   uint64            `json:"min_response_time_ms"`
	MaxResponseTime   uint64            `json:"max_response_time_ms"`
	Status2xx         uint64            `json:"status_2xx"`
	Status3xx         uint64            `json:"status_3xx"`
	Status4xx         uint64            `json:"status_4xx"`
	Status5xx         uint64            `json:"status_5xx"`
	TotalErrors       uint64            `json:"total_errors"`
	Uptime            string            `json:"uptime"`
	CustomCounters    map[string]uint64 `json:"custom_counters,omitempty"`
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

// IncrementRequests increments the total request counter
func (m *Metrics) IncrementRequests() {
	atomic.AddUint64(&m.totalRequests, 1)
}

// IncrementSuccessful increments successful requests
func (m *Metrics) IncrementSuccessful() {
	atomic.AddUint64(&m.successfulReqs, 1)
}

// IncrementFailed increments failed requests
func (m *Metrics) IncrementFailed() {
	atomic.AddUint64(&m.failedReqs, 1)
}

// RecordResponseTime records a response time
func (m *Metrics) RecordResponseTime(durationMs uint64) {
	atomic.AddUint64(&m.totalResponseTime, durationMs)

	// Update min/max
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.minResponseTime == 0 || durationMs < m.minResponseTime {
		m.minResponseTime = durationMs
	}

	if durationMs > m.maxResponseTime {
		m.maxResponseTime = durationMs
	}
}

// RecordStatusCode records a status code
func (m *Metrics) RecordStatusCode(statusCode int) {
	switch {
	case statusCode >= 200 && statusCode < 300:
		atomic.AddUint64(&m.status2xx, 1)
	case statusCode >= 300 && statusCode < 400:
		atomic.AddUint64(&m.status3xx, 1)
	case statusCode >= 400 && statusCode < 500:
		atomic.AddUint64(&m.status4xx, 1)
	case statusCode >= 500:
		atomic.AddUint64(&m.status5xx, 1)
	}
}

// IncrementErrors increments error counter
func (m *Metrics) IncrementErrors() {
	atomic.AddUint64(&m.totalErrors, 1)
}

// IncrementCustomCounter increments a custom counter
func (m *Metrics) IncrementCustomCounter(name string) {
	value, _ := m.customCounters.LoadOrStore(name, new(uint64))
	counter := value.(*uint64)
	atomic.AddUint64(counter, 1)
}

// GetCustomCounter gets a custom counter value
func (m *Metrics) GetCustomCounter(name string) uint64 {
	if value, ok := m.customCounters.Load(name); ok {
		counter := value.(*uint64)
		return atomic.LoadUint64(counter)
	}
	return 0
}

// GetSnapshot returns a snapshot of current metrics
func (m *Metrics) GetSnapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalReqs := atomic.LoadUint64(&m.totalRequests)
	totalRespTime := atomic.LoadUint64(&m.totalResponseTime)

	var avgRespTime float64
	if totalReqs > 0 {
		avgRespTime = float64(totalRespTime) / float64(totalReqs)
	}

	// Collect custom counters
	customCounters := make(map[string]uint64)
	m.customCounters.Range(func(key, value interface{}) bool {
		name := key.(string)
		counter := value.(*uint64)
		customCounters[name] = atomic.LoadUint64(counter)
		return true
	})

	return MetricsSnapshot{
		TotalRequests:   totalReqs,
		SuccessfulReqs:  atomic.LoadUint64(&m.successfulReqs),
		FailedReqs:      atomic.LoadUint64(&m.failedReqs),
		AvgResponseTime: avgRespTime,
		MinResponseTime: m.minResponseTime,
		MaxResponseTime: m.maxResponseTime,
		Status2xx:       atomic.LoadUint64(&m.status2xx),
		Status3xx:       atomic.LoadUint64(&m.status3xx),
		Status4xx:       atomic.LoadUint64(&m.status4xx),
		Status5xx:       atomic.LoadUint64(&m.status5xx),
		TotalErrors:     atomic.LoadUint64(&m.totalErrors),
		Uptime:          time.Since(m.startTime).String(),
		CustomCounters:  customCounters,
	}
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreUint64(&m.totalRequests, 0)
	atomic.StoreUint64(&m.successfulReqs, 0)
	atomic.StoreUint64(&m.failedReqs, 0)
	atomic.StoreUint64(&m.totalResponseTime, 0)
	m.minResponseTime = 0
	m.maxResponseTime = 0
	atomic.StoreUint64(&m.status2xx, 0)
	atomic.StoreUint64(&m.status3xx, 0)
	atomic.StoreUint64(&m.status4xx, 0)
	atomic.StoreUint64(&m.status5xx, 0)
	atomic.StoreUint64(&m.totalErrors, 0)
	m.customCounters = sync.Map{}
	m.startTime = time.Now()
}

// Global metrics instance
var global *Metrics
var once sync.Once

// GetGlobalMetrics returns the global metrics instance
func GetGlobalMetrics() *Metrics {
	once.Do(func() {
		global = NewMetrics()
	})
	return global
}
