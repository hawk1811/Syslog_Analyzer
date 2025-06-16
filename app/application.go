package app

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"syslog-analyzer/config"
	"syslog-analyzer/models"
	"syslog-analyzer/syslog"
	"syslog-analyzer/web"
)

// Application represents the main syslog analyzer application
type Application struct {
	configManager    *config.Manager
	sources          map[string]*syslog.SyslogSource
	sourceMutex      sync.RWMutex
	webServer        *web.Server
	sharedListeners  map[string]*syslog.SharedListener // map[port:protocol] -> SharedListener
	listenerMutex    sync.RWMutex
}

// NewApplication creates a new application instance
func NewApplication(configFile string) *Application {
	app := &Application{
		configManager:   config.NewManager(configFile),
		sources:         make(map[string]*syslog.SyslogSource),
		webServer:       web.NewServer(),
		sharedListeners: make(map[string]*syslog.SharedListener),
	}
	
	// Set up web server handlers
	app.webServer.SetHandlers(
		app.getMetrics,
		app.getSources,
		app.addSource,
		app.updateSource,
		app.deleteSource,
		app.validateSource,
	)
	
	return app
}

// LoadConfig loads configuration from file
func (app *Application) LoadConfig() error {
	_, err := app.configManager.LoadConfig()
	return err
}

// SaveConfig saves current configuration to file
func (app *Application) SaveConfig() error {
	return app.configManager.SaveConfig()
}

// StartSources starts all configured syslog sources
func (app *Application) StartSources() error {
	app.sourceMutex.Lock()
	defer app.sourceMutex.Unlock()
	
	config := app.configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	
	for _, sourceConfig := range config.Sources {
		source := syslog.NewSyslogSource(sourceConfig)
		if err := source.Start(app); err != nil {
			log.Printf("✗ Failed to start source %s: %v", sourceConfig.Name, err)
			continue
		}
		
		app.sources[sourceConfig.Name] = source
	}
	
	log.Printf("✓ Started %d syslog sources", len(app.sources))
	return nil
}

// GetGlobalMetrics calculates and returns global metrics
func (app *Application) GetGlobalMetrics() models.GlobalMetrics {
	app.sourceMutex.RLock()
	defer app.sourceMutex.RUnlock()
	
	global := models.GlobalMetrics{
		TotalSources: len(app.sources),
	}
	
	var totalHourlyLogs, totalDailyLogs int64
	var totalHourlyGB, totalDailyGB float64
	
	for _, source := range app.sources {
		if source == nil {
			continue
		}
		
		metrics := source.GetMetrics()
		
		global.TotalRealTimeEPS += metrics.RealTimeEPS
		global.TotalRealTimeGBps += metrics.RealTimeGBps
		
		if metrics.IsActive {
			global.ActiveSources++
		}
		
		totalHourlyLogs += metrics.HourlyAvgLogs
		totalHourlyGB += metrics.HourlyAvgGB
		totalDailyLogs += metrics.DailyAvgLogs
		totalDailyGB += metrics.DailyAvgGB
	}
	
	global.TotalHourlyAvg = models.SourceMetrics{
		Name:          "Global Total",
		HourlyAvgLogs: totalHourlyLogs,
		HourlyAvgGB:   totalHourlyGB,
		DailyAvgLogs:  totalDailyLogs,
		DailyAvgGB:    totalDailyGB,
	}
	
	global.TotalDailyAvg = global.TotalHourlyAvg
	
	return global
}

// StartWebServer starts the web management interface
func (app *Application) StartWebServer() error {
	config := app.configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	
	return app.webServer.Start(config.GlobalSettings.WebPort)
}

// Stop gracefully stops the application
func (app *Application) Stop() {
	log.Println("✓ Application shutting down...")
	
	// Stop all sources
	app.sourceMutex.Lock()
	for _, source := range app.sources {
		source.Stop(app)
	}
	app.sourceMutex.Unlock()
	
	log.Println("✓ Application stopped")
}

// GetWebPort returns the web server port
func (app *Application) GetWebPort() int {
	config := app.configManager.GetConfig()
	if config == nil {
		return 8080
	}
	return config.GlobalSettings.WebPort
}

// GetSourceCount returns the number of configured sources
func (app *Application) GetSourceCount() int {
	config := app.configManager.GetConfig()
	if config == nil {
		return 0
	}
	return len(config.Sources)
}

// GetSharedListener gets or creates a shared listener for the given protocol and port
func (app *Application) GetSharedListener(protocol string, port int) (*syslog.SharedListener, error) {
	listenerKey := fmt.Sprintf("%d:%s", port, strings.ToUpper(protocol))
	
	app.listenerMutex.Lock()
	defer app.listenerMutex.Unlock()
	
	if sharedListener, exists := app.sharedListeners[listenerKey]; exists {
		return sharedListener, nil
	}
	
	// Create new shared listener
	sharedListener := syslog.NewSharedListener(strings.ToUpper(protocol), port)
	
	if err := sharedListener.Start(); err != nil {
		return nil, fmt.Errorf("failed to start shared listener on %s port %d: %v", protocol, port, err)
	}
	
	app.sharedListeners[listenerKey] = sharedListener
	log.Printf("✓ Started %s listener on port %d", protocol, port)
	
	return sharedListener, nil
}

// RemoveSharedListener removes a shared listener
func (app *Application) RemoveSharedListener(protocol string, port int) {
	listenerKey := fmt.Sprintf("%d:%s", port, strings.ToUpper(protocol))
	
	app.listenerMutex.Lock()
	defer app.listenerMutex.Unlock()
	
	if sharedListener, exists := app.sharedListeners[listenerKey]; exists {
		sharedListener.Stop()
		delete(app.sharedListeners, listenerKey)
	}
}

// Web server handler functions

// getMetrics returns current metrics for the web server
func (app *Application) getMetrics() ([]models.SourceMetrics, models.GlobalMetrics) {
	app.sourceMutex.RLock()
	defer app.sourceMutex.RUnlock()
	
	var sourceMetrics []models.SourceMetrics
	for _, source := range app.sources {
		if source != nil {
			sourceMetrics = append(sourceMetrics, source.GetMetrics())
		}
	}
	
	// Sort sources by name for consistent ordering
	for i := 0; i < len(sourceMetrics)-1; i++ {
		for j := i + 1; j < len(sourceMetrics); j++ {
			if sourceMetrics[i].Name > sourceMetrics[j].Name {
				sourceMetrics[i], sourceMetrics[j] = sourceMetrics[j], sourceMetrics[i]
			}
		}
	}
	
	return sourceMetrics, app.GetGlobalMetrics()
}

// getSources returns all configured sources
func (app *Application) getSources() []models.SourceConfig {
	config := app.configManager.GetConfig()
	if config == nil {
		return []models.SourceConfig{}
	}
	return config.Sources
}

// addSource adds a new source
func (app *Application) addSource(newSource models.SourceConfig) error {
	config := app.configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	
	// Add source to configuration
	config.Sources = append(config.Sources, newSource)
	app.configManager.UpdateConfig(config)
	
	// Start the source
	source := syslog.NewSyslogSource(newSource)
	if err := source.Start(app); err != nil {
		return err
	}
	
	app.sourceMutex.Lock()
	app.sources[newSource.Name] = source
	app.sourceMutex.Unlock()
	
	// Save configuration
	if err := app.SaveConfig(); err != nil {
		log.Printf("⚠ Warning: Failed to save config: %v", err)
	}
	
	return nil
}

// updateSource updates an existing source
func (app *Application) updateSource(oldName string, updatedSource models.SourceConfig) error {
	config := app.configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	
	// Stop existing source
	app.sourceMutex.Lock()
	if existingSource, exists := app.sources[oldName]; exists {
		existingSource.Stop(app)
		delete(app.sources, oldName)
	}
	app.sourceMutex.Unlock()
	
	// Remove from configuration
	var newSources []models.SourceConfig
	for _, source := range config.Sources {
		if source.Name != oldName {
			newSources = append(newSources, source)
		}
	}
	
	// Add updated source to configuration
	newSources = append(newSources, updatedSource)
	config.Sources = newSources
	app.configManager.UpdateConfig(config)
	
	// Start the updated source
	source := syslog.NewSyslogSource(updatedSource)
	if err := source.Start(app); err != nil {
		return err
	}
	
	app.sourceMutex.Lock()
	app.sources[updatedSource.Name] = source
	app.sourceMutex.Unlock()
	
	// Save configuration
	if err := app.SaveConfig(); err != nil {
		log.Printf("⚠ Warning: Failed to save config: %v", err)
	}
	
	return nil
}

// deleteSource deletes a source
func (app *Application) deleteSource(name string) error {
	config := app.configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	
	// Stop and remove source
	app.sourceMutex.Lock()
	if source, exists := app.sources[name]; exists {
		source.Stop(app)
		delete(app.sources, name)
	}
	app.sourceMutex.Unlock()
	
	// Remove from configuration
	var newSources []models.SourceConfig
	for _, source := range config.Sources {
		if source.Name != name {
			newSources = append(newSources, source)
		}
	}
	config.Sources = newSources
	app.configManager.UpdateConfig(config)
	
	// Save configuration
	if err := app.SaveConfig(); err != nil {
		log.Printf("⚠ Warning: Failed to save config: %v", err)
	}
	
	return nil
}

// validateSource validates a source configuration
func (app *Application) validateSource(source models.SourceConfig) error {
	if source.Name == "" {
		return fmt.Errorf("source name is required")
	}
	
	if source.IP == "" {
		return fmt.Errorf("source IP is required")
	}
	
	if source.Port <= 0 || source.Port > 65535 {
		return fmt.Errorf("invalid port number")
	}
	
	config := app.configManager.GetConfig()
	if config == nil {
		return fmt.Errorf("no configuration loaded")
	}
	
	// Check for duplicate names or IPs
	for _, existing := range config.Sources {
		if existing.Name == source.Name {
			return fmt.Errorf("source name already exists")
		}
		if existing.IP == source.IP && existing.Port == source.Port {
			return fmt.Errorf("source IP and port combination already exists")
		}
	}
	
	return nil
}