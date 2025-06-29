package destinations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"syslog-analyzer/models"
)

// HECHandler handles HEC destination processing
type HECHandler struct {
	config models.HECConfig
	client *http.Client
}

// NewHECHandler creates a new HEC handler
func NewHECHandler(config models.HECConfig) *HECHandler {
	return &HECHandler{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessBatch processes a batch of events to HEC
func (h *HECHandler) ProcessBatch(batch *models.LogBatch, sourceName string) error {
	if len(batch.Events) == 0 {
		return nil
	}
	
	// Convert events to HEC format
	hecEvents := make([]map[string]interface{}, 0, len(batch.Events))
	
	for _, event := range batch.Events {
		hecEvent := map[string]interface{}{
			"time":   event.Time.Unix(),
			"event":  event.Event,
			"source": sourceName,
		}
		hecEvents = append(hecEvents, hecEvent)
	}
	
	// Send to HEC
	return h.sendToHEC(hecEvents)
}

// Close closes the HEC handler (implements DestinationProcessor interface)
func (h *HECHandler) Close() error {
	// For HTTP client, we don't need to do anything special to close
	// The client will be garbage collected
	log.Printf("✓ HEC handler closed")
	return nil
}

// sendToHEC sends events to the HEC endpoint
func (h *HECHandler) sendToHEC(events []map[string]interface{}) error {
	// Convert to JSON
	var payload bytes.Buffer
	encoder := json.NewEncoder(&payload)
	
	for _, event := range events {
		if err := encoder.Encode(event); err != nil {
			return fmt.Errorf("failed to encode event: %v", err)
		}
	}
	
	// Create request
	req, err := http.NewRequest("POST", h.config.URL, &payload)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Splunk "+h.config.APIKey)
	
	// Send request
	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("HEC returned HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	log.Printf("✓ Sent %d events to HEC", len(events))
	return nil
}