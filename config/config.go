package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"syslog-analyzer/models"
)

// Manager handles configuration loading and saving
type Manager struct {
	configFile string
	config     *models.Config
}

// NewManager creates a new configuration manager
func NewManager(configFile string) *Manager {
	return &Manager{
		configFile: configFile,
	}
}

// LoadConfig loads configuration from file
func (m *Manager) LoadConfig() (*models.Config, error) {
	if _, err := os.Stat(m.configFile); os.IsNotExist(err) {
		// Create default configuration
		m.config = &models.Config{
			Sources: []models.SourceConfig{},
			GlobalSettings: models.GlobalSettings{
				WebPort:               8080,
				MaxMemoryPerSource:    "100MB",
				MetricsRetentionHours: 24,
			},
		}
		if err := m.SaveConfig(); err != nil {
			return nil, err
		}
		return m.config, nil
	}
	
	data, err := ioutil.ReadFile(m.configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	
	m.config = &models.Config{}
	if err := json.Unmarshal(data, m.config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}
	
	return m.config, nil
}

// SaveConfig saves current configuration to file
func (m *Manager) SaveConfig() error {
	if m.config == nil {
		return fmt.Errorf("no configuration to save")
	}
	
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	
	// Create backup
	if _, err := os.Stat(m.configFile); !os.IsNotExist(err) {
		backupFile := m.configFile + ".backup"
		if err := os.Rename(m.configFile, backupFile); err != nil {
			log.Printf("Warning: failed to create backup: %v", err)
		}
	}
	
	if err := ioutil.WriteFile(m.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	
	return nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *models.Config {
	return m.config
}

// UpdateConfig updates the current configuration
func (m *Manager) UpdateConfig(config *models.Config) {
	m.config = config
}