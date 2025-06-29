package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"syslog-analyzer/models"
)

// Server represents the web server for the dashboard
type Server struct {
	server      *http.Server
	router      *mux.Router
	wsManager   *WebSocketManager
	
	// Handler functions
	getMetricsFunc    func() ([]models.SourceMetrics, models.GlobalMetrics)
	getSourcesFunc    func() []models.SourceConfig
	addSourceFunc     func(models.SourceConfig) error
	updateSourceFunc  func(string, models.SourceConfig) error
	deleteSourceFunc  func(string) error
	validateSourceFunc func(models.SourceConfig) error
}

// NewServer creates a new web server instance
func NewServer() *Server {
	server := &Server{
		router:    mux.NewRouter(),
		wsManager: NewWebSocketManager(),
	}
	
	server.setupRoutes()
	return server
}

// SetHandlers sets the handler functions for the server
func (s *Server) SetHandlers(
	getMetrics func() ([]models.SourceMetrics, models.GlobalMetrics),
	getSources func() []models.SourceConfig,
	addSource func(models.SourceConfig) error,
	updateSource func(string, models.SourceConfig) error,
	deleteSource func(string) error,
	validateSource func(models.SourceConfig) error,
) {
	s.getMetricsFunc = getMetrics
	s.getSourcesFunc = getSources
	s.addSourceFunc = addSource
	s.updateSourceFunc = updateSource
	s.deleteSourceFunc = deleteSource
	s.validateSourceFunc = validateSource
}

// setupRoutes configures all the HTTP routes
func (s *Server) setupRoutes() {
	// WebSocket endpoint - COMPLETELY SEPARATE, NO MIDDLEWARE
	wsRouter := mux.NewRouter()
	wsRouter.HandleFunc("/ws", s.handleWebSocketRaw).Methods("GET")
	
	// All other routes with middleware
	mainRouter := mux.NewRouter()
	
	// Static files and dashboard
	mainRouter.HandleFunc("/", s.handleDashboard).Methods("GET")
	mainRouter.HandleFunc("/dashboard", s.handleDashboard).Methods("GET")
	
	// API endpoints
	api := mainRouter.PathPrefix("/api").Subrouter()
	api.HandleFunc("/metrics", s.handleGetMetrics).Methods("GET")
	api.HandleFunc("/sources", s.handleGetSources).Methods("GET")
	api.HandleFunc("/sources", s.handleAddSource).Methods("POST")
	api.HandleFunc("/sources/{name}", s.handleUpdateSource).Methods("PUT")
	api.HandleFunc("/sources/{name}", s.handleDeleteSource).Methods("DELETE")
	api.HandleFunc("/destinations/test", s.handleTestDestination).Methods("POST")
	api.HandleFunc("/report", s.handleGenerateReport).Methods("GET")
	
	// Apply middleware to main router only
	mainRouter.Use(s.corsMiddleware)
	mainRouter.Use(s.loggingMiddleware)
	
	// Combine routers: WebSocket first (no middleware), then main router (with middleware)
	s.router = mux.NewRouter()
	s.router.PathPrefix("/ws").Handler(wsRouter)
	s.router.PathPrefix("/").Handler(mainRouter)
}

// handleWebSocketRaw handles WebSocket upgrade with RAW ResponseWriter
func (s *Server) handleWebSocketRaw(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔌 WebSocket connection from %s", r.RemoteAddr)
	
	// Ensure WebSocket manager is available
	if s.wsManager == nil {
		log.Printf("❌ WebSocket manager not initialized")
		http.Error(w, "WebSocket service unavailable", http.StatusServiceUnavailable)
		return
	}
	
	// Handle the WebSocket upgrade with completely raw ResponseWriter
	s.wsManager.HandleWebSocket(w, r)
}

// Start starts the web server on the specified port
func (s *Server) Start(port int) error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Start the WebSocket manager first
	log.Printf("🔌 Starting WebSocket manager...")
	go s.wsManager.Start()
	
	// Start metrics broadcasting
	log.Printf("📊 Starting metrics broadcast...")
	go s.startMetricsBroadcast()
	
	log.Printf("✓ Web server starting on port %d", port)
	log.Printf("🌐 Dashboard: http://localhost:%d", port)
	log.Printf("🔌 WebSocket: ws://localhost:%d/ws", port)
	
	return s.server.ListenAndServe()
}

// Stop gracefully stops the web server
func (s *Server) Stop() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := s.server.Shutdown(ctx); err != nil {
			log.Printf("⚠ Error stopping web server: %v", err)
		} else {
			log.Printf("✓ Web server stopped")
		}
	}
	
	// Stop WebSocket manager
	if s.wsManager != nil {
		s.wsManager.Stop()
	}
}

// startMetricsBroadcast starts broadcasting metrics to WebSocket clients
func (s *Server) startMetricsBroadcast() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	log.Printf("📡 Metrics broadcast started (every 2 seconds)")
	
	// Track last broadcast time for less spammy logging
	lastLogTime := time.Now()
	
	for range ticker.C {
		if s.getMetricsFunc != nil && s.wsManager != nil {
			sources, global := s.getMetricsFunc()
			
			data := map[string]interface{}{
				"sources": sources,
				"global":  global,
			}
			
			// Only broadcast if we have connected clients
			clientCount := s.wsManager.GetClientCount()
			if clientCount > 0 {
				s.wsManager.Broadcast(data)
				
				// Log only every 30 seconds instead of every 2 seconds to reduce spam
				if time.Since(lastLogTime) >= 30*time.Second {
					log.Printf("📊 Broadcasting metrics to %d clients (last 30 seconds)", clientCount)
					lastLogTime = time.Now()
				}
			}
		}
	}
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests - SIMPLE VERSION WITHOUT RESPONSE WRAPPING
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Call next handler WITHOUT wrapping ResponseWriter
		next.ServeHTTP(w, r)
		
		duration := time.Since(start)
		
		// Only log API calls, not static files
		if r.URL.Path != "/" && r.URL.Path != "/dashboard" {
			log.Printf("🌐 %s %s %v", r.Method, r.URL.Path, duration)
		}
	})
}