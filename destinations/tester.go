 
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
	
	if sourceIP == "" {
		return false, "Source IP is required and was not provided"
	}
	
	// Create test payload with the updated format
	currentTime := time.Now().UTC().Format("Jan 02 15:04:05")
	testPayload := map[string]interface{}{
		"time": currentTime,
		"event": map[string]interface{}{
			"message":     "Source OK - Test message from Syslog Analyzer",
			"source_ip":   sourceIP,
			"source_name": sourceName,
			"test":        true,
		},
		"source": sourceName,
	}
	
	payloadBytes, err := json.Marshal(testPayload)
	if err != nil {
		return false, fmt.Sprintf("Failed to create test payload: %v", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false, fmt.Sprintf("Failed to create HTTP request: %v", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Splunk "+apiKey)
	
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