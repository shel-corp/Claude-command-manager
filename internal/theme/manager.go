package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// Settings represents theme-related configuration
type Settings struct {
	CurrentTheme string `json:"current_theme"`
	AutoDetect   bool   `json:"auto_detect"` // Auto-detect light/dark based on terminal
}

// Manager handles theme state and persistence
type Manager struct {
	mu           sync.RWMutex
	currentTheme Theme
	settings     Settings
	configPath   string
	styles       *Styles // Cached theme-aware styles
}

// Styles holds all theme-aware style functions
type Styles struct {
	// Base styles (functions match lipgloss.Style.Render signature)
	Base        func(...string) string
	Header      func(...string) string
	Footer      func(...string) string
	Highlight   func(...string) string
	Success     func(...string) string
	Danger      func(...string) string
	Warning     func(...string) string
	Subtle      func(...string) string
	Key         func(...string) string

	// UI component styles (adaptive colors)
	Primary     lipgloss.AdaptiveColor
	SuccessCol  lipgloss.AdaptiveColor
	DangerCol   lipgloss.AdaptiveColor
	WarningCol  lipgloss.AdaptiveColor
	MutedCol    lipgloss.AdaptiveColor
	BackgroundCol lipgloss.AdaptiveColor
	TextCol     lipgloss.AdaptiveColor
	BorderCol   lipgloss.AdaptiveColor

	// Lipgloss styles (for direct use)
	BaseStyle        lipgloss.Style
	HeaderStyle      lipgloss.Style
	FooterStyle      lipgloss.Style
	HighlightStyle   lipgloss.Style
	SuccessStyle     lipgloss.Style
	DangerStyle      lipgloss.Style
	WarningStyle     lipgloss.Style
	SubtleStyle      lipgloss.Style
	KeyStyle         lipgloss.Style
}

// NewManager creates a new theme manager
func NewManager(configPath string) *Manager {
	// Default to DefaultTheme if no config exists
	settings := Settings{
		CurrentTheme: DefaultTheme.ID,
		AutoDetect:   true,
	}

	manager := &Manager{
		currentTheme: DefaultTheme,
		settings:     settings,
		configPath:   configPath,
	}

	// Generate initial styles
	manager.generateStyles()

	return manager
}

// Load reads theme settings from disk
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create config file with defaults if it doesn't exist
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return m.save() // Save default settings
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read theme config: %w", err)
	}

	// Validate JSON
	if !json.Valid(data) {
		// Reset to defaults if corrupt
		return m.save()
	}

	if err := json.Unmarshal(data, &m.settings); err != nil {
		return fmt.Errorf("failed to parse theme config: %w", err)
	}

	// Apply the loaded theme
	return m.applyTheme(m.settings.CurrentTheme)
}

// Save writes theme settings to disk
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.save()
}

// save is the internal save method (caller must hold lock)
func (m *Manager) save() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create theme config directory: %w", err)
	}

	data, err := json.MarshalIndent(m.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal theme config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme config: %w", err)
	}

	return nil
}

// SetTheme changes the current theme and persists the change
func (m *Manager) SetTheme(themeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.applyTheme(themeID); err != nil {
		return err
	}

	m.settings.CurrentTheme = themeID
	return m.save()
}

// GetCurrentTheme returns the currently active theme
func (m *Manager) GetCurrentTheme() Theme {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTheme
}

// GetStyles returns the current theme-aware styles
func (m *Manager) GetStyles() *Styles {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.styles
}

// GetSettings returns the current theme settings
func (m *Manager) GetSettings() Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings
}

// applyTheme applies a theme by ID (caller must hold lock)
func (m *Manager) applyTheme(themeID string) error {
	theme := GetThemeByID(themeID)
	m.currentTheme = theme
	m.generateStyles()
	return nil
}

// generateStyles creates theme-aware style functions and colors
func (m *Manager) generateStyles() {
	theme := m.currentTheme

	// Extract adaptive colors for direct use
	primary := theme.Primary
	success := theme.Success
	danger := theme.Danger
	warning := theme.Warning
	muted := theme.Muted
	background := theme.Background
	text := theme.Text
	border := theme.Border

	// Create lipgloss styles
	baseStyle := lipgloss.NewStyle().Foreground(text)
	headerStyle := lipgloss.NewStyle().Foreground(primary).Bold(true).Padding(0, 1)
	footerStyle := lipgloss.NewStyle().Foreground(muted).Italic(true).Padding(1, 0, 0, 0)
	highlightStyle := lipgloss.NewStyle().Foreground(primary).Bold(true)
	successStyle := lipgloss.NewStyle().Foreground(success).Bold(true)
	dangerStyle := lipgloss.NewStyle().Foreground(danger).Bold(true)
	warningStyle := lipgloss.NewStyle().Foreground(warning).Bold(true)
	subtleStyle := lipgloss.NewStyle().Foreground(muted)
	keyStyle := lipgloss.NewStyle().Foreground(primary).Bold(true).Width(12).Align(lipgloss.Right)

	// Create style functions
	m.styles = &Styles{
		// Function-based styles
		Base:      baseStyle.Render,
		Header:    headerStyle.Render,
		Footer:    footerStyle.Render,
		Highlight: highlightStyle.Render,
		Success:   successStyle.Render,
		Danger:    dangerStyle.Render,
		Warning:   warningStyle.Render,
		Subtle:    subtleStyle.Render,
		Key:       keyStyle.Render,

		// Direct color access
		Primary:       primary,
		SuccessCol:    success,
		DangerCol:     danger,
		WarningCol:    warning,
		MutedCol:      muted,
		BackgroundCol: background,
		TextCol:       text,
		BorderCol:     border,

		// Lipgloss styles for direct use
		BaseStyle:      baseStyle,
		HeaderStyle:    headerStyle,
		FooterStyle:    footerStyle,
		HighlightStyle: highlightStyle,
		SuccessStyle:   successStyle,
		DangerStyle:    dangerStyle,
		WarningStyle:   warningStyle,
		SubtleStyle:    subtleStyle,
		KeyStyle:       keyStyle,
	}
}

// GetAvailableThemes returns all available themes for UI display
func (m *Manager) GetAvailableThemes() []Theme {
	return GetAllThemes()
}

// IsThemeActive checks if a theme is currently active
func (m *Manager) IsThemeActive(themeID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTheme.ID == themeID
}

// ResetToDefault resets theme to default settings
func (m *Manager) ResetToDefault() error {
	return m.SetTheme(DefaultTheme.ID)
}

// GetThemePreview returns a preview of a specific theme
func (m *Manager) GetThemePreview(themeID string) ThemePreview {
	theme := GetThemeByID(themeID)
	return theme.GeneratePreview()
}

// GetCurrentPreview returns a preview of the current theme
func (m *Manager) GetCurrentPreview() ThemePreview {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTheme.GeneratePreview()
}