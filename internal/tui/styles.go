package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	primaryColor   = lipgloss.Color("#0EA5E9")  // Blue
	successColor   = lipgloss.Color("#10B981")  // Green
	dangerColor    = lipgloss.Color("#EF4444")  // Red
	warningColor   = lipgloss.Color("#F59E0B")  // Yellow
	mutedColor     = lipgloss.Color("#6B7280")  // Gray
	backgroundColor = lipgloss.Color("#111827") // Dark gray
	textColor      = lipgloss.Color("#F9FAFB")  // Light gray
)

// Base styles
var (
	baseStyle = lipgloss.NewStyle().
		Foreground(textColor)

	headerStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Padding(1, 0, 0, 0)

	highlightStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

	successStyle = lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true)

	dangerStyle = lipgloss.NewStyle().
		Foreground(dangerColor).
		Bold(true)

	warningStyle = lipgloss.NewStyle().
		Foreground(warningColor).
		Bold(true)

	subtleStyle = lipgloss.NewStyle().
		Foreground(mutedColor)

	keyStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Width(12).
		Align(lipgloss.Right)

	sessionHeaderStyle = lipgloss.NewStyle().
		Foreground(warningColor).
		Bold(true).
		Padding(1, 0, 0, 0)

	sessionChangeStyle = lipgloss.NewStyle().
		Foreground(successColor).
		Padding(0, 0, 0, 2)
)