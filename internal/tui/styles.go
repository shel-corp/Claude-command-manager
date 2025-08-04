package tui

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/shel-corp/Claude-command-manager/internal/theme"
)

// Global theme manager instance
var themeManager *theme.Manager

// InitializeThemeManager initializes the theme manager (call this before using styles)
func InitializeThemeManager() {
	if themeManager != nil {
		return // Already initialized
	}
	
	// Get config path for theme settings
	homeDir, _ := os.UserHomeDir()
	themeConfigPath := filepath.Join(homeDir, ".claude", "theme.json")
	
	themeManager = theme.NewManager(themeConfigPath)
	
	// Load theme settings
	if err := themeManager.Load(); err != nil {
		// Continue with defaults if loading fails
		themeManager.ResetToDefault()
	}
	
	// Refresh styles after initialization
	RefreshStyles()
}

// GetThemeManager returns the global theme manager
func GetThemeManager() *theme.Manager {
	return themeManager
}

// Theme-aware color getters with fallback to default colors
func getPrimaryColor() lipgloss.AdaptiveColor {
	if themeManager == nil {
		return lipgloss.AdaptiveColor{Light: "#0EA5E9", Dark: "#0EA5E9"} // Default blue
	}
	return themeManager.GetStyles().Primary
}

func getSuccessColor() lipgloss.AdaptiveColor {
	if themeManager == nil {
		return lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#10B981"} // Default green
	}
	return themeManager.GetStyles().SuccessCol
}

func getDangerColor() lipgloss.AdaptiveColor {
	if themeManager == nil {
		return lipgloss.AdaptiveColor{Light: "#EF4444", Dark: "#EF4444"} // Default red
	}
	return themeManager.GetStyles().DangerCol
}

func getWarningColor() lipgloss.AdaptiveColor {
	if themeManager == nil {
		return lipgloss.AdaptiveColor{Light: "#F59E0B", Dark: "#F59E0B"} // Default yellow
	}
	return themeManager.GetStyles().WarningCol
}

func getMutedColor() lipgloss.AdaptiveColor {
	if themeManager == nil {
		return lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#6B7280"} // Default gray
	}
	return themeManager.GetStyles().MutedCol
}

func getBackgroundColor() lipgloss.AdaptiveColor {
	if themeManager == nil {
		return lipgloss.AdaptiveColor{Light: "#F9FAFB", Dark: "#111827"} // Default adaptive
	}
	return themeManager.GetStyles().BackgroundCol
}

func getTextColor() lipgloss.AdaptiveColor {
	if themeManager == nil {
		return lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F9FAFB"} // Default adaptive
	}
	return themeManager.GetStyles().TextCol
}

func getBorderColor() lipgloss.AdaptiveColor {
	if themeManager == nil {
		return lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#334155"} // Default adaptive border
	}
	return themeManager.GetStyles().BorderCol
}

// Dynamic style getters that update when theme changes

// Color accessors (backward compatibility) - now adaptive
var primaryColor = getPrimaryColor()
var successColor = getSuccessColor() 
var dangerColor = getDangerColor()
var warningColor = getWarningColor()
var mutedColor = getMutedColor()
var backgroundColor = getBackgroundColor()
var textColor = getTextColor()

// Dynamic style functions that get fresh styles from theme manager with fallbacks
func getBaseStyle() lipgloss.Style {
	if themeManager == nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F9FAFB"))
	}
	return themeManager.GetStyles().BaseStyle
}

func getHeaderStyle() lipgloss.Style {
	if themeManager == nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#0EA5E9")).Bold(true).Padding(0, 1)
	}
	return themeManager.GetStyles().HeaderStyle
}

func getFooterStyle() lipgloss.Style {
	if themeManager == nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Italic(true).Padding(1, 0, 0, 0)
	}
	return themeManager.GetStyles().FooterStyle
}

func getHighlightStyle() lipgloss.Style {
	if themeManager == nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#0EA5E9")).Bold(true)
	}
	return themeManager.GetStyles().HighlightStyle
}

func getSuccessStyle() lipgloss.Style {
	if themeManager == nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
	}
	return themeManager.GetStyles().SuccessStyle
}

func getDangerStyle() lipgloss.Style {
	if themeManager == nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true)
	}
	return themeManager.GetStyles().DangerStyle
}

func getWarningStyle() lipgloss.Style {
	if themeManager == nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true)
	}
	return themeManager.GetStyles().WarningStyle
}

func getSubtleStyle() lipgloss.Style {
	if themeManager == nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	}
	return themeManager.GetStyles().SubtleStyle
}

func getKeyStyle() lipgloss.Style {
	if themeManager == nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#0EA5E9")).Bold(true).Width(12).Align(lipgloss.Right)
	}
	return themeManager.GetStyles().KeyStyle
}

// Backward compatibility - style variables (these get refreshed on theme change)
var (
	baseStyle      = getBaseStyle()
	headerStyle    = getHeaderStyle()
	footerStyle    = getFooterStyle()
	highlightStyle = getHighlightStyle()
	successStyle   = getSuccessStyle()
	dangerStyle    = getDangerStyle()
	warningStyle   = getWarningStyle()
	subtleStyle    = getSubtleStyle()
	keyStyle       = getKeyStyle()

	// Additional styles that don't change frequently
	sessionHeaderStyle = lipgloss.NewStyle().
		Foreground(getWarningColor()).
		Bold(true).
		Padding(1, 0, 0, 0)

	sessionChangeStyle = lipgloss.NewStyle().
		Foreground(getSuccessColor()).
		Padding(0, 0, 0, 2)

	// Left-aligned container styles with margin
	leftMarginContainerStyle = lipgloss.NewStyle().
		Align(lipgloss.Left).
		Padding(1, 2).
		MarginLeft(4)

	leftMarginHeaderStyle = lipgloss.NewStyle().
		Foreground(getPrimaryColor()).
		Bold(true).
		Align(lipgloss.Left).
		Padding(0, 1).
		MarginLeft(4)

	leftMarginFooterStyle = lipgloss.NewStyle().
		Foreground(getMutedColor()).
		Italic(true).
		Align(lipgloss.Left).
		Padding(1, 0, 0, 0).
		MarginLeft(4)
)

// RefreshStyles updates all cached styles when theme changes
func RefreshStyles() {
	// Update color variables
	primaryColor = getPrimaryColor()
	successColor = getSuccessColor()
	dangerColor = getDangerColor()
	warningColor = getWarningColor()
	mutedColor = getMutedColor()
	backgroundColor = getBackgroundColor()
	textColor = getTextColor()

	// Update style variables
	baseStyle = getBaseStyle()
	headerStyle = getHeaderStyle()
	footerStyle = getFooterStyle()
	highlightStyle = getHighlightStyle()
	successStyle = getSuccessStyle()
	dangerStyle = getDangerStyle()
	warningStyle = getWarningStyle()
	subtleStyle = getSubtleStyle()
	keyStyle = getKeyStyle()

	// Update other styles
	sessionHeaderStyle = lipgloss.NewStyle().
		Foreground(warningColor).
		Bold(true).
		Padding(1, 0, 0, 0)

	sessionChangeStyle = lipgloss.NewStyle().
		Foreground(successColor).
		Padding(0, 0, 0, 2)

	leftMarginHeaderStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Align(lipgloss.Left).
		Padding(0, 1).
		MarginLeft(4)

	leftMarginFooterStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Align(lipgloss.Left).
		Padding(1, 0, 0, 0).
		MarginLeft(4)
}

// Utility functions for layout

// leftMarginContent applies left margin to content
func leftMarginContent(content string, width int) string {
	if width <= 0 {
		return content
	}
	
	containerStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Left).
		MarginLeft(4)
	
	return containerStyle.Render(content)
}

// leftMarginView creates a left-aligned view with margin for header, content, and footer
func leftMarginView(header, content, footer string, width int) string {
	if width <= 0 {
		// Fallback for invalid width - add simple margin
		return "    " + header + "\n\n    " + content + "\n    " + footer
	}
	
	styledHeader := leftMarginHeaderStyle.Width(width).Render(header)
	styledFooter := leftMarginFooterStyle.Width(width).Render(footer)
	styledContent := leftMarginContainerStyle.Width(width).Render(content)
	
	return styledHeader + "\n\n" + styledContent + "\n" + styledFooter
}

// Deprecated: centerView is kept for backward compatibility
func centerView(header, content, footer string, width int) string {
	return leftMarginView(header, content, footer, width)
}