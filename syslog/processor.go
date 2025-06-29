 
package syslog

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"syslog-analyzer/filtering"
	"syslog-analyzer/models"
)

// LogProcessor handles the processing pipeline for syslog messages
type LogProcessor struct {
	config         models.SourceConfig
	queue          *LogQueue
	filterEngine   *filtering.Engine
	aggregator     *filtering.Aggregator
	metrics        *MetricsCalculator
	stopChan       chan bool
	batchSize      int
	isRunning      bool
	mutex          sync.RWMutex
	lastMessageAt  time.Time
	msgMutex       sync.RWMutex
}

// NewLogProcessor creates a new log processor
func NewLogProcessor(config models.SourceConfig, batchSize int) *LogProcessor {
	processor := &LogProcessor{
		config:       config,
		queue:        NewLogQueue(1000), // Queue capacity
		filterEngine: filtering.NewEngine(config.Filters),
		aggregator:   filtering.NewAggregator(config.Aggregations),
		metrics:      NewMetricsCalculator(),
		stopChan:     make(chan bool),
		batchSize:    batchSize,
	}
	
	return processor
}

// Start begins the log processing pipeline
func (lp *LogProcessor) Start() error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	if lp.isRunning {
		return fmt.Errorf("processor already running")
	}
	
	lp.isRunning = true
	
	// Start processing threads
	if !lp.config.SimulationMode {
		// Full processing mode with filtering/aggregation
		go lp.runFilteringThread()
		
		// Start destination threads
		for _, dest := range lp.config.Destinations {
			if dest.Enabled {
				go lp.runDestinationThread(dest)
			}
		}
	} else {
		// Simulation mode - just process for metrics
		go lp.runSimulationThread()
	}
	
	log.Printf("✓ Log processor started for source '%s' (simulation: %v)", lp.config.Name, lp.config.SimulationMode)
	return nil
}

// Stop gracefully stops the log processor
func (lp *LogProcessor) Stop() {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	
	if !lp.isRunning {
		return
	}
	
	lp.isRunning = false
	close(lp.stopChan)
	
	log.Printf("✓ Log processor stopped for source '%s'", lp.config.Name)
}

// ProcessRawMessage processes a raw syslog message
func (lp *LogProcessor) ProcessRawMessage(data []byte, sourceIP string) {
	// Update last message time
	lp.msgMutex.Lock()
	lp.lastMessageAt = time.Now()
	lp.msgMutex.Unlock()
	
	// Parse the message into a LogEvent
	event := lp.parseMessage(data, sourceIP)
	if event == nil {
		return
	}
	
	// Add to current batch
	lp.addToBatch(event)
}

// parseMessage parses raw syslog data into a LogEvent
func (lp *LogProcessor) parseMessage(data []byte, sourceIP string) *models.LogEvent {
	event := &models.LogEvent{
		Time:   time.Now().UTC(),
		Source: lp.config.Name,
		Size:   int64(len(data)),
	}
	
	// Try to parse as JSON first
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err == nil {
		// Valid JSON
		event.Event = jsonData
	} else {
		// Treat as string
		message := strings.TrimSpace(string(data))
		if message == "" {
			return nil
		}
		event.Event = message
	}
	
	return event
}

// addToBatch adds an event to the current batch
func (lp *LogProcessor) addToBatch(event *models.LogEvent) {
	// For simplicity, we'll create batches immediately
	// In a production system, you'd accumulate events and batch them
	batch := lp.queue.GetBatch()
	batch.Events = append(batch.Events, *event)
	batch.SourceIP = lp.config.IP
	batch.Timestamp = time.Now()
	
	// If batch is full or we're batching, enqueue it
	if len(batch.Events) >= lp.batchSize {
		if !lp.queue.Enqueue(batch) {
			// Queue full, drop batch
			lp.queue.ReturnBatch(batch)
			log.Printf("⚠ Queue full for source '%s', dropping batch", lp.config.Name)
		}
	} else {
		// For now, enqueue single events immediately
		// In production, you'd buffer until batch size reached
		if !lp.queue.Enqueue(batch) {
			lp.queue.ReturnBatch(batch)
		}
	}
}

// runSimulationThread processes events in simulation mode (metrics only)
func (lp *LogProcessor) runSimulationThread() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-lp.stopChan:
			return
		case <-ticker.C:
			// Process batches for metrics only
			var totalLogs, totalSize int64
			
			for {
				batch := lp.queue.Dequeue()
				if batch == nil {
					break
				}
				
				batchLogs := int64(len(batch.Events))
				var batchSize int64
				for _, event := range batch.Events {
					batchSize += event.Size
				}
				
				totalLogs += batchLogs
				totalSize += batchSize
				
				lp.queue.IncrementProcessed(batchLogs)
				lp.queue.ReturnBatch(batch)
			}
			
			if totalLogs > 0 {
				lp.metrics.RecordMetrics(totalLogs, totalSize, totalLogs, 0)
			}
		}
	}
}

// runFilteringThread processes events with filtering and aggregation
func (lp *LogProcessor) runFilteringThread() {
	for {
		select {
		case <-lp.stopChan:
			return
		default:
			batch := lp.queue.Dequeue()
			if batch == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			
			// Apply filtering
			filteredEvents := lp.filterEngine.ProcessBatch(batch.Events)
			
			// Apply aggregation
			processedEvents := lp.aggregator.ProcessBatch(filteredEvents)
			
			// Record metrics
			batchLogs := int64(len(batch.Events))
			processedLogs := int64(len(processedEvents))
			var batchSize int64
			for _, event := range batch.Events {
				batchSize += event.Size
			}
			
			lp.metrics.RecordMetrics(batchLogs, batchSize, processedLogs, 0)
			lp.queue.IncrementProcessed(processedLogs)
			
			// Create processed batch for destinations
			if len(processedEvents) > 0 {
				processedBatch := lp.queue.GetBatch()
				processedBatch.Events = processedEvents
				processedBatch.SourceIP = batch.SourceIP
				processedBatch.Timestamp = batch.Timestamp
				
				// Send to destination queues (simplified - in production you'd have separate destination queues)
				// For now, we'll just mark as sent
				lp.queue.IncrementSent(int64(len(processedEvents)))
				lp.queue.ReturnBatch(processedBatch)
			}
			
			lp.queue.ReturnBatch(batch)
		}
	}
}

// runDestinationThread processes events for a specific destination
func (lp *LogProcessor) runDestinationThread(dest models.Destination) {
	log.Printf("✓ Started destination thread for '%s' (%s)", dest.Name, dest.Type)
	
	// This is a simplified implementation
	// In production, you'd have separate queues per destination
	for {
		select {
		case <-lp.stopChan:
			return
		default:
			// Destination-specific processing would go here
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// GetMetrics returns current metrics for this processor
func (lp *LogProcessor) GetMetrics() models.SourceMetrics {
	lp.msgMutex.RLock()
	lastMsgTime := lp.lastMessageAt
	lp.msgMutex.RUnlock()
	
	lp.mutex.RLock()
	isActive := lp.isRunning
	lp.mutex.RUnlock()
	
	isReceiving := !lastMsgTime.IsZero() && time.Since(lastMsgTime) < 10*time.Second
	queueStats := lp.queue.GetStats()
	
	return lp.metrics.CalculateMetrics(
		lp.config.Name,
		lp.config.IP,
		lp.config.Port,
		lp.config.Protocol,
		lp.config.SimulationMode,
		queueStats,
		isActive,
		isReceiving,
		lastMsgTime,
	)
}

// IsRunning returns whether the processor is currently running
func (lp *LogProcessor) IsRunning() bool {
	lp.mutex.RLock()
	defer lp.mutex.RUnlock()
	return lp.isRunning
}