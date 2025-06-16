package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"syslog-analyzer/destinations"
	"syslog-analyzer/models"

	"github.com/gorilla/mux"
	"github.com/jung-kurt/gofpdf"
)

// Server handles HTTP requests and WebSocket connections
type Server struct {
	wsManager             *WSManager
	destTester            *destinations.Tester
	metricsHandler        func() ([]models.SourceMetrics, models.GlobalMetrics)
	sourcesHandler        func() []models.SourceConfig
	addSourceHandler      func(models.SourceConfig) error
	updateSourceHandler   func(string, models.SourceConfig) error
	deleteSourceHandler   func(string) error
	validateSourceHandler func(models.SourceConfig) error
	sourceMetrics         []models.SourceMetrics
	globalMetrics         models.GlobalMetrics
	metricsMutex          sync.RWMutex
}

// NewServer creates a new web server
func NewServer() *Server {
	return &Server{
		wsManager:  NewWSManager(),
		destTester: destinations.NewTester(),
	}
}

// SetHandlers sets the callback handlers for the server
func (s *Server) SetHandlers(
	metricsHandler func() ([]models.SourceMetrics, models.GlobalMetrics),
	sourcesHandler func() []models.SourceConfig,
	addSourceHandler func(models.SourceConfig) error,
	updateSourceHandler func(string, models.SourceConfig) error,
	deleteSourceHandler func(string) error,
	validateSourceHandler func(models.SourceConfig) error,
) {
	s.metricsHandler = metricsHandler
	s.sourcesHandler = sourcesHandler
	s.addSourceHandler = addSourceHandler
	s.updateSourceHandler = updateSourceHandler
	s.deleteSourceHandler = deleteSourceHandler
	s.validateSourceHandler = validateSourceHandler
}

// Start starts the web server
func (s *Server) Start(port int) error {
	s.wsManager.Start()

	router := mux.NewRouter()

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.HandlerFunc(s.serveStatic)))

	// API endpoints
	router.HandleFunc("/api/metrics", s.handleAPIMetrics).Methods("GET")
	router.HandleFunc("/api/sources", s.handleAPISources).Methods("GET", "POST")
	router.HandleFunc("/api/sources/{name}", s.handleAPISource).Methods("PUT", "DELETE")
	router.HandleFunc("/api/destinations/test", s.handleTestDestination).Methods("POST")
	router.HandleFunc("/api/report", s.handleAPIReport).Methods("GET")

	// WebSocket endpoint
	router.HandleFunc("/ws", s.wsManager.HandleConnection)

	// Main dashboard
	router.HandleFunc("/", s.handleDashboard).Methods("GET")

	// Start metrics broadcaster
	go s.broadcastMetrics()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	log.Printf("✓ Web server starting on port %d", port)
	return server.ListenAndServe()
}

// serveStatic serves static files
func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/static/")

	switch path {
	case "style.css":
		w.Header().Set("Content-Type", "text/css")
		w.Write([]byte(CSSContent))
	case "app.js":
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(JSContent))
	default:
		http.NotFound(w, r)
	}
}

// handleDashboard serves the main dashboard
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(HTMLContent))
}

// Helper function to convert SourceMetrics to frontend-expected format
func convertSourceMetricsToJSON(sources []models.SourceMetrics) []map[string]interface{} {
	result := make([]map[string]interface{}, len(sources))
	for i, source := range sources {
		result[i] = map[string]interface{}{
			"name":            source.Name,
			"source_ip":       source.SourceIP,
			"port":            source.Port,
			"protocol":        source.Protocol,
			"realtime_eps":    source.RealTimeEPS,
			"realtime_gbps":   source.RealTimeGBps,
			"hourly_avg_logs": source.HourlyAvgLogs,
			"hourly_avg_gb":   source.HourlyAvgGB,
			"daily_avg_logs":  source.DailyAvgLogs,
			"daily_avg_gb":    source.DailyAvgGB,
			"last_updated":    source.LastUpdated,
			"is_active":       source.IsActive,
			"is_receiving":    source.IsReceiving,
			"last_message_at": source.LastMessageAt,
		}
	}
	return result
}

// Helper function to convert GlobalMetrics to frontend-expected format
func convertGlobalMetricsToJSON(global models.GlobalMetrics) map[string]interface{} {
	return map[string]interface{}{
		"total_realtime_eps":  global.TotalRealTimeEPS,
		"total_realtime_gbps": global.TotalRealTimeGBps,
		"total_hourly_avg_gb": global.TotalHourlyAvgGB,
		"total_daily_avg_gb":  global.TotalDailyAvgGB,
		"active_sources":      global.ActiveSources,
		"total_sources":       global.TotalSources,
	}
}

// handleAPIMetrics returns current metrics as JSON
func (s *Server) handleAPIMetrics(w http.ResponseWriter, r *http.Request) {
	if s.metricsHandler == nil {
		http.Error(w, "Metrics handler not set", http.StatusInternalServerError)
		return
	}

	sourceMetrics, globalMetrics := s.metricsHandler()

	response := map[string]interface{}{
		"sources": convertSourceMetricsToJSON(sourceMetrics),
		"global":  convertGlobalMetricsToJSON(globalMetrics),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("⚠ Error encoding metrics response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleAPISources handles source management
func (s *Server) handleAPISources(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if s.sourcesHandler == nil {
			http.Error(w, "Sources handler not set", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.sourcesHandler())

	case "POST":
		if s.addSourceHandler == nil || s.validateSourceHandler == nil {
			http.Error(w, "Add source handler not set", http.StatusInternalServerError)
			return
		}

		var newSource models.SourceConfig
		if err := json.NewDecoder(r.Body).Decode(&newSource); err != nil {
			log.Printf("⚠ Error decoding source JSON: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
			return
		}

		// Set default protocol if not specified
		if newSource.Protocol == "" {
			newSource.Protocol = "UDP"
		}

		// Validate source
		if err := s.validateSourceHandler(newSource); err != nil {
			log.Printf("⚠ Source validation failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Add source
		newSource.CreatedAt = time.Now()
		if err := s.addSourceHandler(newSource); err != nil {
			log.Printf("✗ Failed to add source: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to add source: %v", err)})
			return
		}

		log.Printf("✓ Source added: %s", newSource.Name)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newSource)
	}
}

// handleAPISource handles individual source operations
func (s *Server) handleAPISource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	switch r.Method {
	case "PUT":
		if s.updateSourceHandler == nil || s.validateSourceHandler == nil {
			http.Error(w, "Update source handler not set", http.StatusInternalServerError)
			return
		}

		var updatedSource models.SourceConfig
		if err := json.NewDecoder(r.Body).Decode(&updatedSource); err != nil {
			log.Printf("⚠ Error decoding updated source JSON: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
			return
		}

		// Set default protocol if not specified
		if updatedSource.Protocol == "" {
			updatedSource.Protocol = "UDP"
		}

		// Validate updated source (skip name check if same name)
		if updatedSource.Name != name {
			if err := s.validateSourceHandler(updatedSource); err != nil {
				log.Printf("⚠ Updated source validation failed: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
		}

		// Update source
		updatedSource.CreatedAt = time.Now()
		if err := s.updateSourceHandler(name, updatedSource); err != nil {
			log.Printf("✗ Failed to update source: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to update source: %v", err)})
			return
		}

		log.Printf("✓ Source updated: %s", updatedSource.Name)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updatedSource)

	case "DELETE":
		if s.deleteSourceHandler == nil {
			http.Error(w, "Delete source handler not set", http.StatusInternalServerError)
			return
		}

		if err := s.deleteSourceHandler(name); err != nil {
			log.Printf("✗ Failed to delete source: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to delete source: %v", err)})
			return
		}

		log.Printf("✓ Source deleted: %s", name)
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleTestDestination handles destination testing
func (s *Server) handleTestDestination(w http.ResponseWriter, r *http.Request) {
	var testRequest models.TestDestinationRequest
	if err := json.NewDecoder(r.Body).Decode(&testRequest); err != nil {
		log.Printf("⚠ Error decoding test request JSON: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.TestDestinationResponse{
			Success: false,
			Message: "Invalid JSON request",
		})
		return
	}

	// Debug logging
	log.Printf("🔍 Debug - Received test request:")
	log.Printf("  Source Name: %s", testRequest.SourceName)
	log.Printf("  Source IP: %s", testRequest.SourceIP)
	log.Printf("  Destination Type: %s", testRequest.Destination.Type)

	// Validate required fields
	if testRequest.SourceName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.TestDestinationResponse{
			Success: false,
			Message: "Source name is required",
		})
		return
	}

	if testRequest.Destination.Type == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.TestDestinationResponse{
			Success: false,
			Message: "Destination type is required",
		})
		return
	}

	// Perform the test
	success, message := s.destTester.TestDestination(&testRequest.Destination, testRequest.SourceName, testRequest.SourceIP)

	// Return the test result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.TestDestinationResponse{
		Success: success,
		Message: message,
	})
}

// handleAPIReport generates and returns a PDF report
func (s *Server) handleAPIReport(w http.ResponseWriter, r *http.Request) {
	if s.metricsHandler == nil {
		http.Error(w, "Metrics handler not set", http.StatusInternalServerError)
		return
	}

	sourceMetrics, globalMetrics := s.metricsHandler()

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Syslog Analyzer Report")
	pdf.Ln(20)

	// Global Metrics
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Global Metrics")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(95, 10, fmt.Sprintf("Total Real-time EPS: %.2f", globalMetrics.TotalRealTimeEPS))
	pdf.Ln(5)
	pdf.Cell(95, 10, fmt.Sprintf("Total Real-time GB/s: %.2f", globalMetrics.TotalRealTimeGBps))
	pdf.Ln(5)
	pdf.Cell(95, 10, fmt.Sprintf("Total Hourly Average GB: %.2f", globalMetrics.TotalHourlyAvgGB))
	pdf.Ln(5)
	pdf.Cell(95, 10, fmt.Sprintf("Total Daily Average GB: %.2f", globalMetrics.TotalDailyAvgGB))
	pdf.Ln(5)
	pdf.Cell(95, 10, fmt.Sprintf("Active Sources: %d", globalMetrics.ActiveSources))
	pdf.Ln(5)
	pdf.Cell(95, 10, fmt.Sprintf("Total Sources: %d", globalMetrics.TotalSources))
	pdf.Ln(15)

	// Source Metrics
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Source Metrics")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	for _, source := range sourceMetrics {
		pdf.Cell(190, 10, fmt.Sprintf("Source: %s (%s:%d)", source.Name, source.SourceIP, source.Port))
		pdf.Ln(5)
		pdf.Cell(95, 10, fmt.Sprintf("Real-time EPS: %.2f", source.RealTimeEPS))
		pdf.Ln(5)
		pdf.Cell(95, 10, fmt.Sprintf("Real-time GB/s: %.2f", source.RealTimeGBps))
		pdf.Ln(5)
		pdf.Cell(95, 10, fmt.Sprintf("Hourly Average GB: %.2f", source.HourlyAvgGB))
		pdf.Ln(5)
		pdf.Cell(95, 10, fmt.Sprintf("Daily Average GB: %.2f", source.DailyAvgGB))
		pdf.Ln(5)
		pdf.Cell(95, 10, fmt.Sprintf("Status: %s", getSourceStatus(source)))
		pdf.Ln(10)
	}

	// Output PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=syslog-report.pdf")
	if err := pdf.Output(w); err != nil {
		log.Printf("⚠ Error generating PDF: %v", err)
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}
}

// broadcastMetrics periodically broadcasts metrics to WebSocket clients
func (s *Server) broadcastMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if s.metricsHandler != nil {
			sourceMetrics, globalMetrics := s.metricsHandler()

			// Convert to proper JSON format for frontend
			response := map[string]interface{}{
				"sources": convertSourceMetricsToJSON(sourceMetrics),
				"global":  convertGlobalMetricsToJSON(globalMetrics),
			}

			data, err := json.Marshal(response)
			if err != nil {
				log.Printf("Error marshaling WebSocket data: %v", err)
				continue
			}

			s.wsManager.BroadcastRaw(data)
		}
	}
}

// GetWSManager returns the WebSocket manager
func (s *Server) GetWSManager() *WSManager {
	return s.wsManager
}

// Update global metrics
func (s *Server) updateGlobalMetrics() {
	s.metricsMutex.Lock()
	defer s.metricsMutex.Unlock()

	var totalRealTimeEPS float64
	var totalRealTimeGBps float64
	var totalHourlyAvgGB float64
	var totalDailyAvgGB float64
	var activeSources int

	for _, source := range s.sourceMetrics {
		if source.IsActive {
			totalRealTimeEPS += source.RealTimeEPS
			totalRealTimeGBps += source.RealTimeGBps
			totalHourlyAvgGB += source.HourlyAvgGB
			totalDailyAvgGB += source.DailyAvgGB
			activeSources++
		}
	}

	s.globalMetrics = models.GlobalMetrics{
		TotalRealTimeEPS:  totalRealTimeEPS,
		TotalRealTimeGBps: totalRealTimeGBps,
		TotalHourlyAvgGB:  totalHourlyAvgGB,
		TotalDailyAvgGB:   totalDailyAvgGB,
		ActiveSources:     activeSources,
		TotalSources:      len(s.sourceMetrics),
	}
}

// handleGetMetrics handles metrics retrieval
func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	s.metricsMutex.RLock()
	defer s.metricsMutex.RUnlock()

	response := map[string]interface{}{
		"sources": s.sourceMetrics,
		"global":  s.globalMetrics,
	}

	json.NewEncoder(w).Encode(response)
}

// handleGetGlobalMetrics handles global metrics retrieval
func (s *Server) handleGetGlobalMetrics(w http.ResponseWriter, r *http.Request) {
	s.metricsMutex.RLock()
	defer s.metricsMutex.RUnlock()

	json.NewEncoder(w).Encode(s.globalMetrics)
}

// getSourceStatus returns a human-readable status for a source
func getSourceStatus(source models.SourceMetrics) string {
	if !source.IsActive {
		return "Inactive"
	}
	if !source.IsReceiving {
		return "No Data"
	}
	return "Active"
}
