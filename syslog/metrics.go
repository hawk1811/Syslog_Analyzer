package syslog

import (
	"sync"
	"time"

	"syslog-analyzer/models"
)

// MetricsCalculator handles real-time metrics calculation
type MetricsCalculator struct {
	buffer            *models.CircularBuffer
	totalLogsIngested int64
	mutex             sync.RWMutex
}

// NewMetricsCalculator creates a new metrics calculator
func NewMetricsCalculator() *MetricsCalculator {
	return &MetricsCalculator{
		buffer: models.NewCircularBuffer(3600), // Store 1 hour of data points
	}
}

// RecordMetrics records new metrics data
func (mc *MetricsCalculator) RecordMetrics(logCount, dataSize, processed, sent int64) {
	mc.mutex.Lock()
	mc.totalLogsIngested += logCount
	mc.mutex.Unlock()
	
	mc.buffer.Add(models.MetricDataPoint{
		Timestamp: time.Now(),
		LogCount:  logCount,
		DataSize:  dataSize,
		Processed: processed,
		Sent:      sent,
	})
}

// GetTotalLogsIngested returns the total logs ingested
func (mc *MetricsCalculator) GetTotalLogsIngested() int64 {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	return mc.totalLogsIngested
}

// CalculateMetrics calculates current metrics
func (mc *MetricsCalculator) CalculateMetrics(name, sourceIP string, port int, protocol string, simulationMode bool, queueStats models.QueueStats, isActive, isReceiving bool, lastMessageAt time.Time) models.SourceMetrics {
	// Calculate hourly and daily averages
	hourlyLogs, hourlyGB, _, _ := mc.buffer.GetAverage(1 * time.Hour)
	dailyLogs, dailyGB, _, _ := mc.buffer.GetAverage(24 * time.Hour)
	
	// Calculate real-time EPS and GB/s from recent data
	recentLogs, recentGB, _, _ := mc.buffer.GetAverage(1 * time.Minute)
	realTimeEPS := float64(recentLogs) / 60.0 // Per second
	realTimeGBps := recentGB / 60.0           // Per second
	
	return models.SourceMetrics{
		Name:              name,
		SourceIP:          sourceIP,
		Port:              port,
		Protocol:          protocol,
		SimulationMode:    simulationMode,
		RealTimeEPS:       realTimeEPS,
		RealTimeGBps:      realTimeGBps,
		TotalLogsIngested: mc.GetTotalLogsIngested(),
		HourlyAvgLogs:     hourlyLogs,
		HourlyAvgGB:       hourlyGB,
		DailyAvgLogs:      dailyLogs,
		DailyAvgGB:        dailyGB,
		QueueDepth:        queueStats.Depth,
		ProcessedCount:    queueStats.Processed,
		SentCount:         queueStats.Sent,
		LastUpdated:       time.Now(),
		IsActive:          isActive,
		IsReceiving:       isReceiving,
		LastMessageAt:     lastMessageAt,
	}
}