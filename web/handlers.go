package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"syslog-analyzer/destinations"
	"syslog-analyzer/models"
	"syslog-analyzer/pdf"
)

// handleDashboard serves the main dashboard HTML
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(HTMLContent))
}

// NOTE: handleWebSocket method removed from here - it's now in server.go

// handleGetMetrics returns current metrics as JSON
func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	if s.getMetricsFunc == nil {
		http.Error(w, "Metrics function not available", http.StatusInternalServerError)
		return
	}
	
	sources, global := s.getMetricsFunc()
	
	response := map[string]interface{}{
		"sources": sources,
		"global":  global,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetSources returns all configured sources
func (s *Server) handleGetSources(w http.ResponseWriter, r *http.Request) {
	if s.getSourcesFunc == nil {
		http.Error(w, "Sources function not available", http.StatusInternalServerError)
		return
	}
	
	sources := s.getSourcesFunc()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sources)
}

// handleAddSource adds a new syslog source
func (s *Server) handleAddSource(w http.ResponseWriter, r *http.Request) {
	if s.addSourceFunc == nil || s.validateSourceFunc == nil {
		http.Error(w, "Source functions not available", http.StatusInternalServerError)
		return
	}
	
	var source models.SourceConfig
	if err := json.NewDecoder(r.Body).Decode(&source); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	
	// Set creation time
	source.CreatedAt = time.Now()
	
	// Validate the source
	if err := s.validateSourceFunc(source); err != nil {
		s.sendErrorResponse(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}
	
	// Add the source
	if err := s.addSourceFunc(source); err != nil {
		s.sendErrorResponse(w, fmt.Sprintf("Failed to add source: %v", err), http.StatusInternalServerError)
		return
	}
	
	s.sendSuccessResponse(w, "Source added successfully")
}

// handleUpdateSource updates an existing syslog source
func (s *Server) handleUpdateSource(w http.ResponseWriter, r *http.Request) {
	if s.updateSourceFunc == nil || s.validateSourceFunc == nil {
		http.Error(w, "Source functions not available", http.StatusInternalServerError)
		return
	}
	
	vars := mux.Vars(r)
	oldName := vars["name"]
	if oldName == "" {
		s.sendErrorResponse(w, "Source name is required", http.StatusBadRequest)
		return
	}
	
	var source models.SourceConfig
	if err := json.NewDecoder(r.Body).Decode(&source); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	
	// Preserve creation time if updating
	if source.CreatedAt.IsZero() {
		source.CreatedAt = time.Now()
	}
	
	// Validate the updated source (skip duplicate name check if name hasn't changed)
	if err := s.validateSourceFunc(source); err != nil && source.Name != oldName {
		s.sendErrorResponse(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}
	
	// Update the source
	if err := s.updateSourceFunc(oldName, source); err != nil {
		s.sendErrorResponse(w, fmt.Sprintf("Failed to update source: %v", err), http.StatusInternalServerError)
		return
	}
	
	s.sendSuccessResponse(w, "Source updated successfully")
}

// handleDeleteSource deletes a syslog source
func (s *Server) handleDeleteSource(w http.ResponseWriter, r *http.Request) {
	if s.deleteSourceFunc == nil {
		http.Error(w, "Delete function not available", http.StatusInternalServerError)
		return
	}
	
	vars := mux.Vars(r)
	name := vars["name"]
	if name == "" {
		s.sendErrorResponse(w, "Source name is required", http.StatusBadRequest)
		return
	}
	
	// Delete the source
	if err := s.deleteSourceFunc(name); err != nil {
		s.sendErrorResponse(w, fmt.Sprintf("Failed to delete source: %v", err), http.StatusInternalServerError)
		return
	}
	
	s.sendSuccessResponse(w, "Source deleted successfully")
}

// handleTestDestination tests a destination configuration
func (s *Server) handleTestDestination(w http.ResponseWriter, r *http.Request) {
	var request models.TestDestinationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if request.SourceName == "" {
		s.sendTestResponse(w, false, "Source name is required")
		return
	}
	
	if request.SourceIP == "" {
		s.sendTestResponse(w, false, "Source IP is required")
		return
	}
	
	// Test the destination
	tester := destinations.NewTester()
	success, message := tester.TestDestination(&request.Destination, request.SourceName, request.SourceIP)
	
	s.sendTestResponse(w, success, message)
}

// handleGenerateReport generates and returns a PDF report
func (s *Server) handleGenerateReport(w http.ResponseWriter, r *http.Request) {
	if s.getMetricsFunc == nil {
		http.Error(w, "Metrics function not available", http.StatusInternalServerError)
		return
	}
	
	// Get current metrics
	sources, global := s.getMetricsFunc()
	
	// Generate PDF report
	generator := pdf.NewGenerator()
	pdfData, err := generator.GenerateReport(sources, global)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate report: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Set headers for PDF download
	filename := fmt.Sprintf("syslog_analyzer_report_%s.pdf", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfData)))
	
	// Write PDF data
	w.Write(pdfData)
}

// sendSuccessResponse sends a JSON success response
func (s *Server) sendSuccessResponse(w http.ResponseWriter, message string) {
	response := map[string]interface{}{
		"success": true,
		"message": message,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// sendErrorResponse sends a JSON error response
func (s *Server) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"success": false,
		"error":   message,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// sendTestResponse sends a test destination response
func (s *Server) sendTestResponse(w http.ResponseWriter, success bool, message string) {
	response := models.TestDestinationResponse{
		Success: success,
		Message: message,
	}
	
	w.Header().Set("Content-Type", "application/json")
	if success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(response)
}