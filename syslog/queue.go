package syslog

import (
	"sync"
	"sync/atomic"
	"time"

	"syslog-analyzer/models"
)

// LogQueue implements a thread-safe queue for log batches
type LogQueue struct {
	batches     chan *models.LogBatch
	processed   int64
	sent        int64
	depth       int64
	batchPool   sync.Pool
	eventPool   sync.Pool
	maxCapacity int
}

// NewLogQueue creates a new log queue
func NewLogQueue(capacity int) *LogQueue {
	queue := &LogQueue{
		batches:     make(chan *models.LogBatch, capacity),
		maxCapacity: capacity,
	}
	
	// Initialize object pools for memory efficiency
	queue.batchPool = sync.Pool{
		New: func() interface{} {
			return &models.LogBatch{
				Events: make([]models.LogEvent, 0, 1000),
			}
		},
	}
	
	queue.eventPool = sync.Pool{
		New: func() interface{} {
			return &models.LogEvent{}
		},
	}
	
	return queue
}

// Enqueue adds a batch to the queue
func (q *LogQueue) Enqueue(batch *models.LogBatch) bool {
	select {
	case q.batches <- batch:
		atomic.AddInt64(&q.depth, 1)
		return true
	default:
		// Queue is full
		return false
	}
}

// Dequeue removes and returns a batch from the queue
func (q *LogQueue) Dequeue() *models.LogBatch {
	select {
	case batch := <-q.batches:
		atomic.AddInt64(&q.depth, -1)
		return batch
	default:
		return nil
	}
}

// GetBatch gets a batch from the pool
func (q *LogQueue) GetBatch() *models.LogBatch {
	batch := q.batchPool.Get().(*models.LogBatch)
	batch.Events = batch.Events[:0] // Reset slice
	return batch
}

// ReturnBatch returns a batch to the pool
func (q *LogQueue) ReturnBatch(batch *models.LogBatch) {
	// Return events to pool
	for i := range batch.Events {
		q.eventPool.Put(&batch.Events[i])
	}
	q.batchPool.Put(batch)
}

// GetEvent gets an event from the pool
func (q *LogQueue) GetEvent() *models.LogEvent {
	return q.eventPool.Get().(*models.LogEvent)
}

// ReturnEvent returns an event to the pool
func (q *LogQueue) ReturnEvent(event *models.LogEvent) {
	*event = models.LogEvent{} // Reset
	q.eventPool.Put(event)
}

// IncrementProcessed increments the processed counter
func (q *LogQueue) IncrementProcessed(count int64) {
	atomic.AddInt64(&q.processed, count)
}

// IncrementSent increments the sent counter
func (q *LogQueue) IncrementSent(count int64) {
	atomic.AddInt64(&q.sent, count)
}

// GetStats returns current queue statistics
func (q *LogQueue) GetStats() models.QueueStats {
	return models.QueueStats{
		Depth:      atomic.LoadInt64(&q.depth),
		Processed:  atomic.LoadInt64(&q.processed),
		Sent:       atomic.LoadInt64(&q.sent),
		LastUpdate: time.Now(),
	}
}

// GetCapacity returns the queue capacity
func (q *LogQueue) GetCapacity() int {
	return q.maxCapacity
}

// IsFull returns true if the queue is at capacity
func (q *LogQueue) IsFull() bool {
	return atomic.LoadInt64(&q.depth) >= int64(q.maxCapacity)
}