package models

import (
	"sync"
	"time"
)

// Destination represents a single destination configuration
type Destination struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"` // "storage" or "hec"
	Name        string      `json:"name"`
	Config      interface{} `json:"config"` // StorageConfig or HECConfig
	Enabled     bool        `json:"enabled"`
	Tested      bool        `json:"tested"`
	TestStatus  string      `json:"test_status"`  // "idle", "testing", "success", "failed"
	TestMessage string      `json:"test_message"` // Error message or success message
}

// StorageConfig represents storage destination configuration
type StorageConfig struct {
	Path              string `json:"path"`
	MaxEventsPerFile  int    `json:"max_events_per_file"`
}

// HECConfig represents HEC destination configuration
type HECConfig struct {
	URL       string `json:"url"`
	APIKey    string `json:"api_key"`
	VerifySSL bool   `json:"verify_ssl"`
}

// FilterRule represents a filtering rule
type FilterRule struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // "contains", "equals", "regex"
	Value    string `json:"value"`
	Action   string `json:"action"` // "include", "exclude"
}

// AggregationRule represents an aggregation rule
type AggregationRule struct {
	GroupBy    []string      `json:"group_by"`
	TimeWindow time.Duration `json:"time_window"`
}

// SourceConfig represents a syslog source configuration
type SourceConfig struct {
	Name            string            `json:"name"`
	IP              string            `json:"ip"`
	Port            int               `json:"port"`
	Protocol        string            `json:"protocol"`
	Destinations    []Destination     `json:"destinations"`
	SimulationMode  bool              `json:"simulation_mode"`
	Filters         []FilterRule      `json:"filters"`
	Aggregations    []AggregationRule `json:"aggregations"`
	CreatedAt       time.Time         `json:"created_at"`
}

// GlobalSettings contains application-wide configuration
type GlobalSettings struct {
	WebPort               int    `json:"web_port"`
	MaxMemoryPerSource    string `json:"max_memory_per_source"`
	MetricsRetentionHours int    `json:"metrics_retention_hours"`
	BatchSize             int    `json:"batch_size"`
	MaxEPSPerSource       int    `json:"max_eps_per_source"`
}

// Config represents the complete application configuration
type Config struct {
	Sources        []SourceConfig `json:"sources"`
	GlobalSettings GlobalSettings `json:"global_settings"`
}

// LogEvent represents a processed log event
type LogEvent struct {
	Time   time.Time   `json:"time"`
	Event  interface{} `json:"event"`
	Source string      `json:"source"`
	Size   int64       `json:"-"` // Internal use for metrics
}

// SourceMetrics holds real-time metrics for a syslog source
type SourceMetrics struct {
	Name              string    `json:"name"`
	SourceIP          string    `json:"source_ip"`
	Port              int       `json:"port"`
	Protocol          string    `json:"protocol"`
	SimulationMode    bool      `json:"simulation_mode"`
	RealTimeEPS       float64   `json:"realtime_eps"`
	RealTimeGBps      float64   `json:"realtime_gbps"`
	TotalLogsIngested int64     `json:"total_logs_ingested"`
	HourlyAvgLogs     int64     `json:"hourly_avg_logs"`
	HourlyAvgGB       float64   `json:"hourly_avg_gb"`
	DailyAvgLogs      int64     `json:"daily_avg_logs"`
	DailyAvgGB        float64   `json:"daily_avg_gb"`
	QueueDepth        int64     `json:"queue_depth"`
	ProcessedCount    int64     `json:"processed_count"`
	SentCount         int64     `json:"sent_count"`
	LastUpdated       time.Time `json:"last_updated"`
	IsActive          bool      `json:"is_active"`
	IsReceiving       bool      `json:"is_receiving"`
	LastMessageAt     time.Time `json:"last_message_at"`
}

// GlobalMetrics represents aggregated metrics across all sources
type GlobalMetrics struct {
	TotalRealTimeEPS      float64 `json:"total_realtime_eps"`
	TotalRealTimeGBps     float64 `json:"total_realtime_gbps"`
	TotalLogsIngested     int64   `json:"total_logs_ingested"`
	TotalHourlyAvgLogs    int64   `json:"total_hourly_avg_logs"`
	TotalHourlyAvgGB      float64 `json:"total_hourly_avg_gb"`
	TotalDailyAvgLogs     int64   `json:"total_daily_avg_logs"`
	TotalDailyAvgGB       float64 `json:"total_daily_avg_gb"`
	TotalQueueDepth       int64   `json:"total_queue_depth"`
	TotalProcessedCount   int64   `json:"total_processed_count"`
	TotalSentCount        int64   `json:"total_sent_count"`
	ActiveSources         int     `json:"active_sources"`
	TotalSources          int     `json:"total_sources"`
}

// MetricDataPoint represents a single metric measurement
type MetricDataPoint struct {
	Timestamp   time.Time
	LogCount    int64
	DataSize    int64 // in bytes
	Processed   int64
	Sent        int64
}

// CircularBuffer implements a thread-safe circular buffer for metrics
type CircularBuffer struct {
	Buffer []MetricDataPoint
	Size   int
	Head   int
	Count  int
	Mutex  sync.RWMutex
}

// NewCircularBuffer creates a new circular buffer with specified size
func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		Buffer: make([]MetricDataPoint, size),
		Size:   size,
	}
}

// Add inserts a new data point into the circular buffer
func (cb *CircularBuffer) Add(point MetricDataPoint) {
	cb.Mutex.Lock()
	defer cb.Mutex.Unlock()
	
	cb.Buffer[cb.Head] = point
	cb.Head = (cb.Head + 1) % cb.Size
	if cb.Count < cb.Size {
		cb.Count++
	}
}

// GetAverage calculates average for the specified duration
func (cb *CircularBuffer) GetAverage(duration time.Duration) (int64, float64, int64, int64) {
	cb.Mutex.RLock()
	defer cb.Mutex.RUnlock()
	
	if cb.Count == 0 {
		return 0, 0, 0, 0
	}
	
	cutoff := time.Now().Add(-duration)
	var totalLogs, totalBytes, totalProcessed, totalSent int64
	var validPoints int
	
	for i := 0; i < cb.Count; i++ {
		idx := (cb.Head - 1 - i + cb.Size) % cb.Size
		if cb.Buffer[idx].Timestamp.After(cutoff) {
			totalLogs += cb.Buffer[idx].LogCount
			totalBytes += cb.Buffer[idx].DataSize
			totalProcessed += cb.Buffer[idx].Processed
			totalSent += cb.Buffer[idx].Sent
			validPoints++
		}
	}
	
	if validPoints == 0 {
		return 0, 0, 0, 0
	}
	
	avgLogs := totalLogs / int64(validPoints)
	avgGB := float64(totalBytes) / (1024 * 1024 * 1024) / float64(validPoints)
	avgProcessed := totalProcessed / int64(validPoints)
	avgSent := totalSent / int64(validPoints)
	
	return avgLogs, avgGB, avgProcessed, avgSent
}

// LogBatch represents a batch of log events for processing
type LogBatch struct {
	Events    []LogEvent
	SourceIP  string
	Timestamp time.Time
}

// QueueStats represents queue statistics
type QueueStats struct {
	Depth      int64
	Processed  int64
	Sent       int64
	LastUpdate time.Time
}

// TestDestinationRequest represents the test request payload
type TestDestinationRequest struct {
	SourceName  string      `json:"source_name"`
	SourceIP    string      `json:"source_ip"`
	Destination Destination `json:"destination"`
}

// TestDestinationResponse represents the test response
type TestDestinationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}