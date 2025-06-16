package destinations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"syslog-analyzer/models"
)

// Tester handles destination testing functionality
type Tester struct{}

// NewTester creates a new destination tester
func NewTester() *Tester {
	return &Tester{}
}

// TestDestination tests a destination based on its type
func (t *Tester) TestDestination(dest *models.Destination, sourceName, sourceIP string) (bool, string) {
	if dest.Type == "storage" {
		return t.testStorageDestination(dest)
	} else if dest.Type == "hec" {
		return t.testHECDestination(dest, sourceName, sourceIP)
	}

	return false, "Unknown destination type: " + dest.Type
}

// testStorageDestination tests if storage destination is accessible
func (t *Tester) testStorageDestination(dest *models.Destination) (bool, string) {
	configMap, ok := dest.Config.(map[string]interface{})
	if !ok {
		return false, "Invalid storage configuration"
	}

	pathInterface, exists := configMap["path"]
	if !exists {
		return false, "Storage path not specified"
	}

	path, ok := pathInterface.(string)
	if !ok {
		return false, "Invalid storage path format"
	}

	if path == "" {
		return false, "Storage path is empty"
	}

	// Normalize path for different OS types
	path = filepath.FromSlash(path)

	// Handle UNC paths on Windows
	if strings.HasPrefix(path, "\\\\") {
		// UNC path - test by trying to access the network share
		if _, err := os.Stat(path); err != nil {
			// Try to create the directory if it doesn't exist
			if os.IsNotExist(err) {
				if err := os.MkdirAll(path, 0755); err != nil {
					return false, fmt.Sprintf("Cannot access UNC path '%s': %v", path, err)
				}
			} else {
				return false, fmt.Sprintf("Cannot access UNC path '%s': %v", path, err)
			}
		}
	} else {
		// Regular path - create directory if it doesn't exist
		if err := os.MkdirAll(path, 0755); err != nil {
			return false, fmt.Sprintf("Cannot create directory '%s': %v", path, err)
		}
	}

	// Test write permissions by creating a temporary file
	testFile := filepath.Join(path, "syslog_analyzer_test_"+fmt.Sprintf("%d", time.Now().Unix())+".tmp")
	file, err := os.Create(testFile)
	if err != nil {
		return false, fmt.Sprintf("Cannot create test file in '%s': %v", path, err)
	}

	// Write test content
	testContent := fmt.Sprintf("Syslog Analyzer Test - %s\nWritten at: %s\n",
		dest.Name, time.Now().Format(time.RFC3339))
	if _, err := file.WriteString(testContent); err != nil {
		file.Close()
		os.Remove(testFile)
		return false, fmt.Sprintf("Cannot write to test file in '%s': %v", path, err)
	}

	// Test read permissions by reading back the content
	if _, err := file.Seek(0, 0); err != nil {
		file.Close()
		os.Remove(testFile)
		return false, fmt.Sprintf("Cannot seek in test file in '%s': %v", path, err)
	}

	readBuffer := make([]byte, len(testContent))
	if _, err := file.Read(readBuffer); err != nil {
		file.Close()
		os.Remove(testFile)
		return false, fmt.Sprintf("Cannot read from test file in '%s': %v", path, err)
	}

	file.Close()

	// Clean up test file
	if err := os.Remove(testFile); err != nil {
		log.Printf("Warning: Could not remove test file %s: %v", testFile, err)
		return true, fmt.Sprintf("Storage destination '%s' is accessible and writable (test file cleanup warning)", path)
	}

	return true, fmt.Sprintf("Storage destination '%s' is accessible with full read/write permissions", path)
}

// testHECDestination tests if HEC destination is accessible
func (t *Tester) testHECDestination(dest *models.Destination, sourceName, sourceIP string) (bool, string) {
	// Debug logging
	log.Printf("🔍 Debug - HEC Test Parameters:")
	log.Printf("  Source Name: '%s'", sourceName)
	log.Printf("  Source IP: '%s'", sourceIP)

	configMap, ok := dest.Config.(map[string]interface{})
	if !ok {
		return false, "Invalid HEC configuration"
	}

	urlInterface, exists := configMap["url"]
	if !exists {
		return false, "HEC URL not specified"
	}

	url, ok := urlInterface.(string)
	if !ok {
		return false, "Invalid HEC URL format"
	}

	apiKeyInterface, exists := configMap["api_key"]
	if !exists {
		return false, "HEC API key not specified"
	}

	apiKey, ok := apiKeyInterface.(string)
	if !ok {
		return false, "Invalid HEC API key format"
	}

	if url == "" {
		return false, "HEC URL is empty"
	}

	if apiKey == "" {
		return false, "HEC API key is empty"
	}

	// Use the actual source IP from the form - no fallback
	if sourceIP == "" {
		return false, "Source IP is required and was not provided"
	}

	log.Printf("🔍 Debug - Using Source IP: '%s'", sourceIP)

	// Create test payload with the updated format
	currentTime := time.Now().UTC().Format("Jan 02 15:04:05")
	testPayload := map[string]interface{}{
		"time": currentTime,
		"event": map[string]interface{}{
			"message":     "Source OK",
			"source_ip":   sourceIP,   // Use actual IP from form
			"source_name": sourceName, // Use source name from form
		},
		"source": sourceName, // This should match source_name in event
	}

	payloadBytes, err := json.Marshal(testPayload)
	if err != nil {
		return false, fmt.Sprintf("Failed to create test payload: %v", err)
	}

	log.Printf("🔍 Debug - HEC Test Payload: %s", string(payloadBytes))

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false, fmt.Sprintf("Failed to create HTTP request: %v", err)
	}

	// Set headers as specified - using Bearer instead of Splunk prefix
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check response status - looking for 200 or 202 as specified
	if resp.StatusCode == 200 || resp.StatusCode == 202 {
		return true, fmt.Sprintf("HEC endpoint is accessible and accepting data (HTTP %d)", resp.StatusCode)
	}

	// Read response body for error details
	body, _ := ioutil.ReadAll(resp.Body)
	errorMessage := string(body)
	if len(errorMessage) > 200 {
		errorMessage = errorMessage[:200] + "..."
	}

	return false, fmt.Sprintf("HEC endpoint returned HTTP %d: %s", resp.StatusCode, errorMessage)
}

var storageState = struct {
	sync.Mutex
	files map[string]*storageFileState
}{files: make(map[string]*storageFileState)}

type storageFileState struct {
	file       *os.File
	createdAt  time.Time
	eventCount int
	filePath   string
}

func WriteBatchToStorage(sourceName string, destConfig models.StorageConfig, events []map[string]interface{}) error {
	if destConfig.Path == "" {
		return nil
	}
	key := sourceName + "::" + destConfig.Path
	storageState.Lock()
	state, ok := storageState.files[key]
	if !ok || state.eventCount+len(events) > 50000 {
		// Need to create/rotate file
		if ok && state.file != nil {
			state.file.Close()
		}
		timestamp := time.Now().UTC().Format("20060102_150405")
		fileName := sourceName + "_" + timestamp + ".json"
		filePath := filepath.Join(destConfig.Path, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			storageState.Unlock()
			return err
		}
		state = &storageFileState{
			file:       file,
			createdAt:  time.Now().UTC(),
			eventCount: 0,
			filePath:   filePath,
		}
		storageState.files[key] = state
	}
	storageState.Unlock()
	// Write events
	enc := json.NewEncoder(state.file)
	for _, event := range events {
		if err := enc.Encode(event); err != nil {
			return err
		}
		state.eventCount++
	}
	return nil
}

func PostBatchToHEC(sourceName string, destConfig models.HECConfig, events []map[string]interface{}) error {
	if destConfig.URL == "" || destConfig.APIKey == "" {
		return nil
	}
	payload, err := json.Marshal(events)
	if err != nil {
		return err
	}
	for {
		req, err := http.NewRequest("POST", destConfig.URL, bytes.NewBuffer(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+destConfig.APIKey)
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err == nil && (resp.StatusCode == 200 || resp.StatusCode == 202) {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
}
