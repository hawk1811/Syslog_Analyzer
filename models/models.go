package models

import (
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
	Path string `json:"path"`
}

// HECConfig represents HEC destination configuration
type HECConfig struct {
	URL    string `json:"url"`
	APIKey string `json:"api_key"`
}

// SourceConfig represents a syslog source configuration
type SourceConfig struct {
	Name           string        `json:"name"`
	IP             string        `json:"ip"`
	Port           int           `json:"port"`
	Protocol       string        `json:"protocol"`
	Destinations   []Destination `json:"destinations"`
	SimulationMode bool          `json:"simulation_mode"`
	CreatedAt      time.Time     `json:"created_at"`
}

// GlobalSettings contains application-wide configuration
type GlobalSettings struct {
	WebPort               int    `json:"web_port"`
	MaxMemoryPerSource    string `json:"max_memory_per_source"`
	MetricsRetentionHours int    `json:"metrics_retention_hours"`
}

// Config represents the complete application configuration
type Config struct {
	Sources        []SourceConfig `json:"sources"`
	GlobalSettings GlobalSettings `json:"global_settings"`
}

// DestinationMetrics holds metrics for a specific destination
type DestinationMetrics struct {
	QueueLength    int64     `json:"queue_length"`
	ProcessedCount int64     `json:"processed_count"`
	LastUpdated    time.Time `json:"last_updated"`
}

// SourceMetrics holds real-time metrics for a syslog source
type SourceMetrics struct {
	Name           string                        `json:"name"`
	SourceIP       string                        `json:"source_ip"`
	Port           int                           `json:"port"`
	Protocol       string                        `json:"protocol"`
	RealTimeEPS    float64                       `json:"realtime_eps"`
	RealTimeGBps   float64                       `json:"realtime_gbps"`
	HourlyAvgLogs  int64                         `json:"hourly_avg_logs"`
	HourlyAvgGB    float64                       `json:"hourly_avg_gb"`
	DailyAvgLogs   int64                         `json:"daily_avg_logs"`
	DailyAvgGB     float64                       `json:"daily_avg_gb"`
	LastUpdated    time.Time                     `json:"last_updated"`
	IsActive       bool                          `json:"is_active"`
	IsReceiving    bool                          `json:"is_receiving"`
	LastMessageAt  time.Time                     `json:"last_message_at"`
	QueueLength    int64                         `json:"queue_length"`
	ProcessedCount int64                         `json:"processed_count"`
	DestMetrics    map[string]DestinationMetrics `json:"dest_metrics"`
}

// GlobalMetrics represents aggregated metrics across all sources
type GlobalMetrics struct {
	TotalRealTimeEPS  float64       `json:"total_realtime_eps"`
	TotalRealTimeGBps float64       `json:"total_realtime_gbps"`
	TotalHourlyAvg    SourceMetrics `json:"total_hourly_avg"`
	TotalDailyAvg     SourceMetrics `json:"total_daily_avg"`
	ActiveSources     int           `json:"active_sources"`
	TotalSources      int           `json:"total_sources"`
}

// MetricDataPoint represents a single metric measurement
type MetricDataPoint struct {
	Timestamp time.Time
	LogCount  int64
	DataSize  int64 // in bytes
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
