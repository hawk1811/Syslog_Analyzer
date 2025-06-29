 
package syslog

import (
	"fmt"
	"log"
	"sync"

	"syslog-analyzer/models"
)

// SyslogSource represents a single syslog source processor
type SyslogSource struct {
	config    models.SourceConfig
	processor *LogProcessor
	mutex     sync.RWMutex
}

// ApplicationInterface defines the interface for application methods needed by SyslogSource
type ApplicationInterface interface {
	GetSharedListener(protocol string, port int) (*SharedListener, error)
	RemoveSharedListener(protocol string, port int)
}

// NewSyslogSource creates a new syslog source processor
func NewSyslogSource(config models.SourceConfig, batchSize int) *SyslogSource {
	return &SyslogSource{
		config:    config,
		processor: NewLogProcessor(config, batchSize),
	}
}

// Start begins processing syslog messages for this source
func (s *SyslogSource) Start(app ApplicationInterface) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Get or create shared listener
	sharedListener, err := app.GetSharedListener(s.config.Protocol, s.config.Port)
	if err != nil {
		return fmt.Errorf("failed to get shared listener on %s port %d: %v", s.config.Protocol, s.config.Port, err)
	}
	
	// Register this source with the shared listener
	sharedListener.AddSource(s)
	
	// Start the log processor
	if err := s.processor.Start(); err != nil {
		return fmt.Errorf("failed to start log processor: %v", err)
	}
	
	log.Printf("✓ Source '%s' started on %s:%d (simulation: %v)", s.config.Name, s.config.Protocol, s.config.Port, s.config.SimulationMode)
	return nil
}

// Stop gracefully stops the syslog source
func (s *SyslogSource) Stop(app ApplicationInterface) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Stop the log processor
	s.processor.Stop()
	
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
	
	log.Printf("✓ Source '%s' stopped", s.config.Name)
}

// ProcessMessage processes a single syslog message
func (s *SyslogSource) ProcessMessage(data []byte, sourceIP string) {
	s.processor.ProcessRawMessage(data, sourceIP)
}

// GetMetrics returns current metrics for this source
func (s *SyslogSource) GetMetrics() models.SourceMetrics {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return s.processor.GetMetrics()
}

// IsRunning returns whether the source is currently running
func (s *SyslogSource) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.processor.IsRunning()
}

// GetConfig returns the source configuration
func (s *SyslogSource) GetConfig() models.SourceConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.config
}