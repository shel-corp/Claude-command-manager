package tui

import (
	"fmt"
	"strings"
)

// View renders the application UI
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	switch m.state {
	case StateMain:
		return m.mainView()
	case StateRename:
		return m.renameView()
	case StateHelp:
		return m.helpView()
	}

	return m.mainView()
}

// mainView renders the main application view
func (m Model) mainView() string {
	var content strings.Builder

	// Header
	content.WriteString(headerStyle.Render("Claude Command Manager (ccm)"))
	content.WriteString("\n\n")

	// Main list
	content.WriteString(m.list.View())
	content.WriteString("\n")

	// Footer with controls
	footer := m.renderFooter()
	content.WriteString(footer)

	return content.String()
}

// renameView renders the rename input view
func (m Model) renameView() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Rename Command"))
	content.WriteString("\n\n")

	if len(m.commands) > m.renameIndex {
		cmd := m.commands[m.renameIndex]
		content.WriteString(fmt.Sprintf("Current name: %s\n", 
			highlightStyle.Render(cmd.DisplayName)))
		content.WriteString(fmt.Sprintf("Description: %s\n\n", 
			subtleStyle.Render(cmd.Description)))
	}

	content.WriteString("New name:\n")
	content.WriteString(m.textInput.View())
	content.WriteString("\n\n")

	footer := footerStyle.Render("Enter: Confirm • Esc: Cancel • Ctrl+C: Quit")
	content.WriteString(footer)

	return content.String()
}

// helpView renders the help screen
func (m Model) helpView() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Help"))
	content.WriteString("\n\n")

	helpItems := []struct {
		key  string
		desc string
	}{
		{"↑/↓, j/k", "Navigate up/down"},
		{"Enter, t", "Toggle command enabled/disabled"},
		{"r", "Rename selected command"},
		{"l", "Toggle symlink location (👤 user / 📁 project)"},
		{"q", "Quit"},
		{"h, ?", "Show this help screen"},
		{"Ctrl+C", "Force quit"},
	}

	for _, item := range helpItems {
		content.WriteString(fmt.Sprintf("  %s  %s\n", 
			keyStyle.Render(item.key),
			item.desc))
	}

	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("Commands are stored as .md files in the commands/ directory."))
	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("Enabled commands are symlinked to ~/.claude/commands/"))
	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("All changes are saved immediately."))
	content.WriteString("\n\n")

	footer := footerStyle.Render("Press any key to return")
	content.WriteString(footer)

	return content.String()
}

// Note: Confirm quit view removed since changes are saved immediately

// renderFooter renders the footer with key bindings
func (m Model) renderFooter() string {
	return footerStyle.Render("Enter/t: Toggle • r: Rename • l: Location • q: Quit • h: Help")
}