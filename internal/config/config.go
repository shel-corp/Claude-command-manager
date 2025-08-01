package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SymlinkLocation represents where a command should be symlinked
type SymlinkLocation string

const (
	SymlinkLocationUser    SymlinkLocation = "user"    // ~/.claude/commands/
	SymlinkLocationProject SymlinkLocation = "project" // <project>/.claude/commands/
)

// CommandConfig represents the configuration for a single command
type CommandConfig struct {
	Enabled         bool            `json:"enabled"`
	OriginalName    string          `json:"original_name"`
	DisplayName     string          `json:"display_name"`
	SourcePath      string          `json:"source_path"`
	SymlinkLocation SymlinkLocation `json:"symlink_location"`
}

// Config represents the entire configuration file structure
type Config struct {
	Commands map[string]CommandConfig `json:"commands"`
}

// Manager handles configuration file operations
type Manager struct {
	configPath string
	config     *Config
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
		config:     &Config{Commands: make(map[string]CommandConfig)},
	}
}

// Load reads the configuration from disk
func (m *Manager) Load() error {
	// Create config file with default content if it doesn't exist
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return m.initializeConfig()
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Validate JSON before unmarshaling
	if !json.Valid(data) {
		// Backup corrupt file and reinitialize
		if err := m.backupAndReinitialize(); err != nil {
			return fmt.Errorf("failed to recover from corrupt config: %w", err)
		}
		return nil
	}

	if err := json.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// Save writes the configuration to disk
func (m *Manager) Save() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetCommand returns the configuration for a specific command
func (m *Manager) GetCommand(name string) (CommandConfig, bool) {
	cmd, exists := m.config.Commands[name]
	return cmd, exists
}

// SetCommand updates the configuration for a specific command
func (m *Manager) SetCommand(name string, config CommandConfig) {
	m.config.Commands[name] = config
}

// DeleteCommand removes a command from the configuration
func (m *Manager) DeleteCommand(name string) {
	delete(m.config.Commands, name)
}

// GetAllCommands returns all command configurations
func (m *Manager) GetAllCommands() map[string]CommandConfig {
	return m.config.Commands
}

// initializeConfig creates a new configuration file with default content
func (m *Manager) initializeConfig() error {
	m.config = &Config{Commands: make(map[string]CommandConfig)}
	return m.Save()
}

// backupAndReinitialize creates a backup of the corrupt config and initializes a new one
func (m *Manager) backupAndReinitialize() error {
	backupPath := fmt.Sprintf("%s.backup.%d", m.configPath, os.Getuid())
	
	// Attempt to backup the corrupt file
	if err := copyFile(m.configPath, backupPath); err != nil {
		// If backup fails, just log and continue
		fmt.Fprintf(os.Stderr, "Warning: failed to backup corrupt config: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "Corrupt config backed up to: %s\n", backupPath)
	}

	return m.initializeConfig()
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}