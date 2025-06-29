 
package destinations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"syslog-analyzer/models"
)

// StorageHandler handles storage destination processing
type StorageHandler struct {
	config          models.StorageConfig
	currentFile     *os.File
	currentPath     string
	eventCount      int
	mutex           sync.Mutex
}

// NewStorageHandler creates a new storage handler
func NewStorageHandler(config models.StorageConfig) *StorageHandler {
	return &StorageHandler{
		config: config,
	}
}

// ProcessBatch processes a batch of events to storage
func (s *StorageHandler) ProcessBatch(batch *models.LogBatch, sourceName string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if len(batch.Events) == 0 {
		return nil
	}
	
	// Check if we need to rotate files
	if s.shouldRotate() {
		if err := s.rotateFile(sourceName); err != nil {
			return fmt.Errorf("failed to rotate file: %v", err)
		}
	}
	
	// Ensure we have an open file
	if s.currentFile == nil {
		if err := s.openNewFile(sourceName); err != nil {
			return fmt.Errorf("failed to open new file: %v", err)
		}
	}
	
	// Write events to file
	encoder := json.NewEncoder(s.currentFile)
	for _, event := range batch.Events {
		if err := encoder.Encode(event); err != nil {
			return fmt.Errorf("failed to write event: %v", err)
		}
		s.eventCount++
	}
	
	// Flush to ensure data is written
	if err := s.currentFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %v", err)
	}
	
	return nil
}

// shouldRotate determines if the current file should be rotated
func (s *StorageHandler) shouldRotate() bool {
	maxEvents := s.config.MaxEventsPerFile
	if maxEvents == 0 {
		maxEvents = 50000 // Default
	}
	
	return s.eventCount >= maxEvents
}

// rotateFile closes the current file and prepares for a new one
func (s *StorageHandler) rotateFile(sourceName string) error {
	if s.currentFile != nil {
		s.currentFile.Close()
		s.currentFile = nil
	}
	
	s.eventCount = 0
	return nil
}

// openNewFile opens a new file for writing
func (s *StorageHandler) openNewFile(sourceName string) error {
	// Create filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.json", sourceName, timestamp)
	filePath := filepath.Join(s.config.Path, filename)
	
	// Ensure directory exists
	if err := os.MkdirAll(s.config.Path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	
	// Open file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	
	s.currentFile = file
	s.currentPath = filePath
	s.eventCount = 0
	
	return nil
}

// Close closes the current file
func (s *StorageHandler) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.currentFile != nil {
		return s.currentFile.Close()
	}
	
	return nil
}