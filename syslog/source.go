package syslog

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"encoding/json"
	"syslog-analyzer/destinations"
	"syslog-analyzer/models"
)

// SyslogSource represents a single syslog source processor
type SyslogSource struct {
	config          models.SourceConfig
	metrics         *models.SourceMetrics
	buffer          *CircularBuffer
	connections     map[string]net.Conn
	connMutex       sync.RWMutex
	stopChan        chan bool
	eventCount      int64
	dataSize        int64
	lastUpdate      time.Time
	lastMessageTime time.Time
	msgMutex        sync.RWMutex
	isRunning       bool
	mutex           sync.RWMutex
	queue           chan []byte
	processedCount  int64
}

// ApplicationInterface defines the interface for application methods needed by SyslogSource
type ApplicationInterface interface {
	GetSharedListener(protocol string, port int) (*SharedListener, error)
	RemoveSharedListener(protocol string, port int)
}

// NewSyslogSource creates a new syslog source processor
func NewSyslogSource(config models.SourceConfig) *SyslogSource {
	return &SyslogSource{
		config:      config,
		buffer:      NewCircularBuffer(3600), // Store 1 hour of data points
		connections: make(map[string]net.Conn),
		stopChan:    make(chan bool),
		metrics: &models.SourceMetrics{
			Name:     config.Name,
			SourceIP: config.IP,
			Port:     config.Port,
			Protocol: config.Protocol,
		},
		queue: make(chan []byte),
	}
}

// Start begins processing syslog messages for this source
func (s *SyslogSource) Start(app ApplicationInterface) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isRunning {
		return fmt.Errorf("source %s is already running", s.config.Name)
	}

	// Get or create shared listener
	sharedListener, err := app.GetSharedListener(s.config.Protocol, s.config.Port)
	if err != nil {
		return fmt.Errorf("failed to get shared listener on %s port %d: %v", s.config.Protocol, s.config.Port, err)
	}

	// Register this source with the shared listener
	sharedListener.AddSource(s)

	s.isRunning = true
	go s.updateMetrics()
	s.startWorker()

	log.Printf("✓ Source '%s' started on %s:%d", s.config.Name, s.config.Protocol, s.config.Port)
	return nil
}

// Stop gracefully stops the syslog source
func (s *SyslogSource) Stop(app ApplicationInterface) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isRunning {
		return
	}

	s.isRunning = false

	// Remove from shared listener
	if sharedListener, err := app.GetSharedListener(s.config.Protocol, s.config.Port); err == nil {
		sharedListener.RemoveSource(s)

		// Check if this was the last source for this listener
		if sharedListener.GetSourceCount() == 0 {
			// No more sources, can stop the shared listener
			app.RemoveSharedListener(s.config.Protocol, s.config.Port)
			log.Printf("✓ Stopped %s listener on port %d", s.config.Protocol, s.config.Port)
		}
	}

	s.metrics.IsActive = false
}

// processMessage processes a single syslog message
func (s *SyslogSource) processMessage(data []byte) {
	if !s.config.SimulationMode {
		s.queue <- data
		atomic.AddInt64(&s.processedCount, 1)
	}
}

// updateMetrics continuously updates metrics for this source
func (s *SyslogSource) updateMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.calculateMetrics()
		}
	}
}

// calculateMetrics calculates current metrics
func (s *SyslogSource) calculateMetrics() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	currentEvents := atomic.LoadInt64(&s.eventCount)
	currentDataSize := atomic.LoadInt64(&s.dataSize)

	// Calculate real-time EPS and GB/s
	if !s.lastUpdate.IsZero() {
		duration := now.Sub(s.lastUpdate).Seconds()
		if duration > 0 {
			s.metrics.RealTimeEPS = float64(currentEvents) / duration
			s.metrics.RealTimeGBps = float64(currentDataSize) / (1024 * 1024 * 1024) / duration
		}
	}

	// Check if we're receiving messages
	s.msgMutex.RLock()
	lastMsgTime := s.lastMessageTime
	s.msgMutex.RUnlock()

	s.metrics.IsReceiving = !lastMsgTime.IsZero() && now.Sub(lastMsgTime) < 10*time.Second
	s.metrics.LastMessageAt = lastMsgTime

	// Add current data point to buffer if we have data
	if currentEvents > 0 || currentDataSize > 0 {
		s.buffer.Add(models.MetricDataPoint{
			Timestamp: now,
			LogCount:  currentEvents,
			DataSize:  currentDataSize,
		})
	}

	// Calculate averages
	s.metrics.HourlyAvgLogs, s.metrics.HourlyAvgGB = s.buffer.GetAverage(1 * time.Hour)
	s.metrics.DailyAvgLogs, s.metrics.DailyAvgGB = s.buffer.GetAverage(24 * time.Hour)

	s.metrics.LastUpdated = now
	s.metrics.IsActive = s.isRunning

	// Reset counters
	atomic.StoreInt64(&s.eventCount, 0)
	atomic.StoreInt64(&s.dataSize, 0)
	s.lastUpdate = now

	s.metrics.QueueLength = int64(len(s.queue))
	s.metrics.ProcessedCount = atomic.LoadInt64(&s.processedCount)
}

// GetMetrics returns current metrics for this source
func (s *SyslogSource) GetMetrics() models.SourceMetrics {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return *s.metrics
}

// IsRunning returns whether the source is currently running
func (s *SyslogSource) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isRunning
}

func (s *SyslogSource) startWorker() {
	go func() {
		batch := make([][]byte, 0, 1000)
		batchSize := 1000
		for {
			if s.config.SimulationMode {
				time.Sleep(1 * time.Second)
				continue
			}
			select {
			case logData := <-s.queue:
				batch = append(batch, logData)
				if len(batch) >= batchSize {
					s.processBatch(batch)
					batch = make([][]byte, 0, batchSize)
				}
			case <-time.After(2 * time.Second):
				if len(batch) > 0 {
					s.processBatch(batch)
					batch = make([][]byte, 0, batchSize)
				}
			}
		}
	}()
}

func (s *SyslogSource) processBatch(batch [][]byte) {
	events := make([]map[string]interface{}, 0, len(batch))
	sourceName := s.config.Name
	simMode := s.config.SimulationMode
	for _, logData := range batch {
		var event map[string]interface{}
		if json.Valid(logData) {
			// Parse JSON log
			if err := json.Unmarshal(logData, &event); err != nil {
				event = map[string]interface{}{"event": string(logData)}
			}
			event = map[string]interface{}{
				"time":            time.Now().UTC().Format("Jan 02 15:04:05"),
				"event":           event,
				"source":          sourceName,
				"simulation_mode": simMode,
			}
		} else {
			event = map[string]interface{}{
				"time":            time.Now().UTC().Format("Jan 02 15:04:05"),
				"event":           string(logData),
				"source":          sourceName,
				"simulation_mode": simMode,
			}
		}
		events = append(events, event)
	}
	for _, dest := range s.config.Destinations {
		if !dest.Enabled {
			continue
		}
		switch dest.Type {
		case "storage":
			cfg, ok := dest.Config.(map[string]interface{})
			if ok {
				storageCfg := models.StorageConfig{}
				if path, exists := cfg["path"].(string); exists {
					storageCfg.Path = path
				}
				_ = destinations.WriteBatchToStorage(sourceName, storageCfg, events)
			}
		case "hec":
			cfg, ok := dest.Config.(map[string]interface{})
			if ok {
				hecCfg := models.HECConfig{}
				if url, exists := cfg["url"].(string); exists {
					hecCfg.URL = url
				}
				if apiKey, exists := cfg["api_key"].(string); exists {
					hecCfg.APIKey = apiKey
				}
				_ = destinations.PostBatchToHEC(sourceName, hecCfg, events)
			}
		}
	}
}
