package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ConfigManager handles loading and saving cache configuration
type ConfigManager struct {
	configPath string
	config     CacheConfig
}

// NewConfigManager creates a new cache config manager
func NewConfigManager() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".config", "claude_command_manager", "cache_config.json")
	
	cm := &ConfigManager{
		configPath: configPath,
		config:     DefaultCacheConfig(),
	}

	// Try to load existing config
	if err := cm.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return cm, nil
}

// Load loads cache configuration from disk
func (cm *ConfigManager) Load() error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &cm.config)
}

// Save saves cache configuration to disk
func (cm *ConfigManager) Save() error {
	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(cm.configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

// GetConfig returns the current cache configuration
func (cm *ConfigManager) GetConfig() CacheConfig {
	return cm.config
}

// SetConfig updates the cache configuration
func (cm *ConfigManager) SetConfig(config CacheConfig) {
	cm.config = config
}

// SetEnabled enables or disables caching
func (cm *ConfigManager) SetEnabled(enabled bool) {
	cm.config.Enabled = enabled
}

// SetTTL sets the cache TTL in hours
func (cm *ConfigManager) SetTTL(hours int) {
	cm.config.TTLHours = hours
}

// SetMaxSize sets the maximum cache size in MB
func (cm *ConfigManager) SetMaxSize(sizeMB int) {
	cm.config.MaxSizeMB = sizeMB
}

// SetBackgroundRefresh enables or disables background refresh
func (cm *ConfigManager) SetBackgroundRefresh(enabled bool) {
	cm.config.BackgroundRefresh = enabled
}

// SetConcurrentWorkers sets the number of concurrent workers for cache refresh
func (cm *ConfigManager) SetConcurrentWorkers(workers int) {
	if workers < 1 {
		workers = 1
	}
	if workers > 10 {
		workers = 10
	}
	cm.config.ConcurrentWorkers = workers
}