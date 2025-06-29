package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jung-kurt/gofpdf"
	"syslog-analyzer/destinations"
	"syslog-analyzer/models"
)

// Server handles HTTP requests and WebSocket connections
type Server struct {
	wsManager      *WSManager
	destTester     *destinations.Tester
	metricsHandler func() ([]models.SourceMetrics, models.GlobalMetrics)
	sourcesHandler func() []models.SourceConfig
	addSourceHandler func(models.SourceConfig) error
	updateSourceHandler func(string, models.SourceConfig) error
	deleteSourceHandler func(string) error
	validateSourceHandler func(models.SourceConfig) error
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
			"name":             source.Name,
			"source_ip":        source.SourceIP,
			"port":             source.Port,
			"protocol":         source.Protocol,
			"realtime_eps":     source.RealTimeEPS,
			"realtime_gbps":    source.RealTimeGBps,
			"hourly_avg_logs":  source.HourlyAvgLogs,
			"hourly_avg_gb":    source.HourlyAvgGB,
			"daily_avg_logs":   source.DailyAvgLogs,
			"daily_avg_gb":     source.DailyAvgGB,
			"last_updated":     source.LastUpdated,
			"is_active":        source.IsActive,
			"is_receiving":     source.IsReceiving,
			"last_message_at":  source.LastMessageAt,
		}
	}
	return result
}

// Helper function to convert GlobalMetrics to frontend-expected format
func convertGlobalMetricsToJSON(global models.GlobalMetrics) map[string]interface{} {
	return map[string]interface{}{
		"total_realtime_eps":  global.TotalRealTimeEPS,
		"total_realtime_gbps": global.TotalRealTimeGBps,
		"total_hourly_avg": map[string]interface{}{
			"hourly_avg_gb":   global.TotalHourlyAvg.HourlyAvgGB,
			"hourly_avg_logs": global.TotalHourlyAvg.HourlyAvgLogs,
		},
		"total_daily_avg": map[string]interface{}{
			"daily_avg_gb":   global.TotalDailyAvg.DailyAvgGB,
			"daily_avg_logs": global.TotalDailyAvg.DailyAvgLogs,
		},
		"active_sources": global.ActiveSources,
		"total_sources":  global.TotalSources,
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
	
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Syslog Analysis Report")
	pdf.Ln(12)
	
	// Timestamp
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
	pdf.Ln(12)
	
	// Global metrics
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Global Summary")
	pdf.Ln(8)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, fmt.Sprintf("Total Sources: %d", globalMetrics.TotalSources))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Active Sources: %d", globalMetrics.ActiveSources))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Total Real-time EPS: %.2f", globalMetrics.TotalRealTimeEPS))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Total Real-time GB/s: %.6f", globalMetrics.TotalRealTimeGBps))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Total Hourly Avg GB/s: %.5f", globalMetrics.TotalHourlyAvg.HourlyAvgGB))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Total Daily Avg GB/s: %.5f", globalMetrics.TotalDailyAvg.DailyAvgGB))
	pdf.Ln(12)
	
	// Source details
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Source Details")
	pdf.Ln(8)
	
	for _, metrics := range sourceMetrics {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(40, 8, metrics.Name)
		pdf.Ln(6)
		
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(40, 6, fmt.Sprintf("  IP: %s, Port: %d", metrics.SourceIP, metrics.Port))
		pdf.Ln(5)
		pdf.Cell(40, 6, fmt.Sprintf("  Real-time EPS: %.2f", metrics.RealTimeEPS))
		pdf.Ln(5)
		pdf.Cell(40, 6, fmt.Sprintf("  Real-time GB/s: %.6f", metrics.RealTimeGBps))
		pdf.Ln(5)
		pdf.Cell(40, 6, fmt.Sprintf("  Hourly Avg Logs: %d", metrics.HourlyAvgLogs))
		pdf.Ln(5)
		pdf.Cell(40, 6, fmt.Sprintf("  Daily Avg Logs: %d", metrics.DailyAvgLogs))
		pdf.Ln(8)
	}
	
	// Output PDF
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to generate PDF"})
		return
	}
	
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=syslog_report.pdf")
	w.Write(buf.Bytes())
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