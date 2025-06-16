package syslog

import (
	"sync"
	"time"

	"syslog-analyzer/models"
)

// CircularBuffer implements a thread-safe circular buffer for metrics
type CircularBuffer struct {
	buffer []models.MetricDataPoint
	size   int
	head   int
	count  int
	mutex  sync.RWMutex
}

// NewCircularBuffer creates a new circular buffer with specified size
func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		buffer: make([]models.MetricDataPoint, size),
		size:   size,
	}
}

// Add inserts a new data point into the circular buffer
func (cb *CircularBuffer) Add(point models.MetricDataPoint) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.buffer[cb.head] = point
	cb.head = (cb.head + 1) % cb.size
	if cb.count < cb.size {
		cb.count++
	}
}

// GetAverage calculates average for the specified duration
func (cb *CircularBuffer) GetAverage(duration time.Duration) (int64, float64) {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	if cb.count == 0 {
		return 0, 0
	}
	
	cutoff := time.Now().Add(-duration)
	var totalLogs int64
	var totalBytes int64
	var validPoints int
	
	for i := 0; i < cb.count; i++ {
		idx := (cb.head - 1 - i + cb.size) % cb.size
		if cb.buffer[idx].Timestamp.After(cutoff) {
			totalLogs += cb.buffer[idx].LogCount
			totalBytes += cb.buffer[idx].DataSize
			validPoints++
		}
	}
	
	if validPoints == 0 {
		return 0, 0
	}
	
	avgLogs := totalLogs / int64(validPoints)
	avgGB := float64(totalBytes) / (1024 * 1024 * 1024) / float64(validPoints)
	
	return avgLogs, avgGB
}