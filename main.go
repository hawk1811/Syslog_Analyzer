 
// Package main implements a high-performance syslog analysis tool
// that can handle up to 500,000 EPS from multiple simultaneous sources
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"syslog-analyzer/app"
)

func main() {
	// Get executable directory for config file
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	
	configFile := filepath.Join(filepath.Dir(execPath), "syslog_analyzer.json")
	
	// Create application
	application := app.NewApplication(configFile)
	
	// Load configuration
	if err := application.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		application.Stop()
		os.Exit(0)
	}()
	
	// Start syslog sources
	if err := application.StartSources(); err != nil {
		log.Fatalf("Failed to start syslog sources: %v", err)
	}
	
	// Display startup information
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("🚀 Professional Syslog Analyzer Started Successfully!\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n")
	fmt.Printf("📊 Service Name: Syslog Analyzer Service\n")
	fmt.Printf("🌐 Web Interface: http://localhost:%d\n", application.GetWebPort())
	fmt.Printf("📡 Sources Loaded: %d\n", application.GetSourceCount())
	fmt.Printf("⚡ Status: Ready to ingest data\n")
	fmt.Printf("🔧 Config File: %s\n", configFile)
	fmt.Printf(strings.Repeat("=", 60) + "\n")
	fmt.Printf("💡 Access the dashboard at: http://localhost:%d\n", application.GetWebPort())
	fmt.Printf("🛑 Press Ctrl+C to stop the service\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n\n")
	
	// Start web server (blocking)
	if err := application.StartWebServer(); err != nil {
		if err != http.ErrServerClosed {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}
}