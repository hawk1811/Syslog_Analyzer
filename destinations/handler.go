package destinations

import (
	"fmt"
	"log"
	"sync"

	"syslog-analyzer/models"
)

// Handler manages destination processing for multiple destinations
type Handler struct {
	destinations map[string]DestinationProcessor
	mutex        sync.RWMutex
}

// DestinationProcessor interface for different destination types
type DestinationProcessor interface {
	ProcessBatch(batch *models.LogBatch, sourceName string) error
	Close() error
}

// NewHandler creates a new destination handler
func NewHandler() *Handler {
	return &Handler{
		destinations: make(map[string]DestinationProcessor),
	}
}

// AddDestination adds a new destination for processing
func (h *Handler) AddDestination(dest models.Destination, sourceName string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if !dest.Enabled {
		log.Printf("⚠ Destination '%s' is disabled, skipping", dest.Name)
		return nil
	}
	
	var processor DestinationProcessor
	var err error
	
	switch dest.Type {
	case "storage":
		processor, err = h.createStorageProcessor(dest)
	case "hec":
		processor, err = h.createHECProcessor(dest)
	default:
		return fmt.Errorf("unknown destination type: %s", dest.Type)
	}
	
	if err != nil {
		return fmt.Errorf("failed to create %s processor: %v", dest.Type, err)
	}
	
	key := fmt.Sprintf("%s_%s", sourceName, dest.ID)
	h.destinations[key] = processor
	
	log.Printf("✓ Added %s destination '%s' for source '%s'", dest.Type, dest.Name, sourceName)
	return nil
}

// RemoveDestination removes a destination from processing
func (h *Handler) RemoveDestination(destID, sourceName string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	key := fmt.Sprintf("%s_%s", sourceName, destID)
	
	if processor, exists := h.destinations[key]; exists {
		if err := processor.Close(); err != nil {
			log.Printf("⚠ Error closing destination processor: %v", err)
		}
		delete(h.destinations, key)
		log.Printf("✓ Removed destination '%s' for source '%s'", destID, sourceName)
	}
	
	return nil
}

// ProcessBatch processes a batch of events through all destinations
func (h *Handler) ProcessBatch(batch *models.LogBatch, sourceName string) error {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	var errors []error
	
	for key, processor := range h.destinations {
		// Check if this destination belongs to the source
		expectedPrefix := sourceName + "_"
		if len(key) < len(expectedPrefix) || key[:len(expectedPrefix)] != expectedPrefix {
			continue
		}
		
		if err := processor.ProcessBatch(batch, sourceName); err != nil {
			log.Printf("⚠ Error processing batch for destination %s: %v", key, err)
			errors = append(errors, err)
		}
	}
	
	// Return first error if any occurred
	if len(errors) > 0 {
		return errors[0]
	}
	
	return nil
}

// Close closes all destination processors
func (h *Handler) Close() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	var errors []error
	
	for key, processor := range h.destinations {
		if err := processor.Close(); err != nil {
			log.Printf("⚠ Error closing destination processor %s: %v", key, err)
			errors = append(errors, err)
		}
	}
	
	// Clear all destinations
	h.destinations = make(map[string]DestinationProcessor)
	
	// Return first error if any occurred
	if len(errors) > 0 {
		return errors[0]
	}
	
	return nil
}

// createStorageProcessor creates a storage destination processor
func (h *Handler) createStorageProcessor(dest models.Destination) (DestinationProcessor, error) {
	configMap, ok := dest.Config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid storage configuration type")
	}
	
	pathInterface, exists := configMap["path"]
	if !exists {
		return nil, fmt.Errorf("storage path not specified")
	}
	
	path, ok := pathInterface.(string)
	if !ok {
		return nil, fmt.Errorf("invalid storage path format")
	}
	
	if path == "" {
		return nil, fmt.Errorf("storage path is empty")
	}
	
	// Get max events per file (optional)
	maxEventsPerFile := 50000 // default
	if maxEventsInterface, exists := configMap["max_events_per_file"]; exists {
		if maxEvents, ok := maxEventsInterface.(float64); ok {
			maxEventsPerFile = int(maxEvents)
		}
	}
	
	config := models.StorageConfig{
		Path:             path,
		MaxEventsPerFile: maxEventsPerFile,
	}
	
	return NewStorageHandler(config), nil
}

// createHECProcessor creates a HEC destination processor
func (h *Handler) createHECProcessor(dest models.Destination) (DestinationProcessor, error) {
	configMap, ok := dest.Config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid HEC configuration type")
	}
	
	urlInterface, exists := configMap["url"]
	if !exists {
		return nil, fmt.Errorf("HEC URL not specified")
	}
	
	url, ok := urlInterface.(string)
	if !ok {
		return nil, fmt.Errorf("invalid HEC URL format")
	}
	
	apiKeyInterface, exists := configMap["api_key"]
	if !exists {
		return nil, fmt.Errorf("HEC API key not specified")
	}
	
	apiKey, ok := apiKeyInterface.(string)
	if !ok {
		return nil, fmt.Errorf("invalid HEC API key format")
	}
	
	if url == "" {
		return nil, fmt.Errorf("HEC URL is empty")
	}
	
	if apiKey == "" {
		return nil, fmt.Errorf("HEC API key is empty")
	}
	
	// Get verify SSL setting (optional, defaults to true)
	verifySSL := true
	if verifySSLInterface, exists := configMap["verify_ssl"]; exists {
		if verify, ok := verifySSLInterface.(bool); ok {
			verifySSL = verify
		}
	}
	
	config := models.HECConfig{
		URL:       url,
		APIKey:    apiKey,
		VerifySSL: verifySSL,
	}
	
	return NewHECHandler(config), nil
}

// GetDestinationCount returns the number of active destinations
func (h *Handler) GetDestinationCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.destinations)
}

// GetDestinationKeys returns all destination keys
func (h *Handler) GetDestinationKeys() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	keys := make([]string, 0, len(h.destinations))
	for key := range h.destinations {
		keys = append(keys, key)
	}
	
	return keys
} 
