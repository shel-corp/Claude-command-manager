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
	case StateRemoteURL:
		return m.remoteURLView()
	case StateRemoteLoading:
		return m.remoteLoadingView()
	case StateRemoteSelect:
		return m.remoteSelectView()
	case StateRemoteImport:
		return m.remoteImportView()
	case StateRemoteResults:
		return m.remoteResultsView()
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
		{"i", "Import commands from GitHub repository"},
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
	return footerStyle.Render("Enter/t: Toggle • r: Rename • l: Location • i: Import • q: Quit • h: Help")
}

// Remote import view functions

// remoteURLView renders the GitHub URL input view
func (m Model) remoteURLView() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Import Commands from GitHub"))
	content.WriteString("\n\n")

	content.WriteString(subtleStyle.Render("Enter a GitHub repository URL containing Claude commands:"))
	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("Examples:"))
	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("  • https://github.com/user/repo/.claude/commands"))
	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("  • github.com/user/repo/commands"))
	content.WriteString("\n\n")

	content.WriteString("Repository URL:\n")
	content.WriteString(m.textInput.View())
	content.WriteString("\n\n")

	// Show error if present
	if m.remoteError != "" {
		content.WriteString(dangerStyle.Render("Error: " + m.remoteError))
		content.WriteString("\n\n")
	}

	footer := footerStyle.Render("Enter: Continue • Esc: Cancel • Ctrl+C: Quit")
	content.WriteString(footer)

	return content.String()
}

// remoteLoadingView renders the loading view
func (m Model) remoteLoadingView() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Loading Repository..."))
	content.WriteString("\n\n")

	if m.remoteRepo != nil {
		content.WriteString(fmt.Sprintf("Repository: %s\n", 
			highlightStyle.Render(fmt.Sprintf("%s/%s", m.remoteRepo.Owner, m.remoteRepo.Repo))))
		content.WriteString(fmt.Sprintf("Branch: %s\n", 
			subtleStyle.Render(m.remoteRepo.Branch)))
		content.WriteString(fmt.Sprintf("Path: %s\n\n", 
			subtleStyle.Render(m.remoteRepo.Path)))
	}

	// Loading animation
	content.WriteString("🔍 Connecting to repository...\n")
	content.WriteString("📦 Scanning for commands...\n")
	content.WriteString("🔄 Loading command details...\n")
	content.WriteString("⚠️  Checking for conflicts...\n\n")

	content.WriteString(subtleStyle.Render("Please wait..."))

	return content.String()
}

// remoteSelectView renders the command selection view
func (m Model) remoteSelectView() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Select Commands to Import"))
	content.WriteString("\n\n")

	if m.remoteRepo != nil {
		content.WriteString(fmt.Sprintf("From: %s\n\n", 
			highlightStyle.Render(fmt.Sprintf("%s/%s", m.remoteRepo.Owner, m.remoteRepo.Repo))))
	}

	// Show selection summary
	selectedCount := 0
	conflictCount := 0
	for i, cmd := range m.remoteCommands {
		if m.remoteSelected[i] {
			selectedCount++
		}
		if cmd.LocalExists {
			conflictCount++
		}
	}

	content.WriteString(fmt.Sprintf("Commands: %d total, %d selected", 
		len(m.remoteCommands), selectedCount))
	if conflictCount > 0 {
		content.WriteString(fmt.Sprintf(", %d conflicts", conflictCount))
	}
	content.WriteString("\n\n")

	// Command list
	content.WriteString(m.list.View())
	content.WriteString("\n")

	// Instructions
	footer := footerStyle.Render("Enter: Toggle • a: Select All • n: Select None • i: Import • Esc: Cancel")
	content.WriteString(footer)

	return content.String()
}

// remoteImportView renders the import progress view
func (m Model) remoteImportView() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Importing Commands..."))
	content.WriteString("\n\n")

	selectedCount := 0
	for i := range m.remoteCommands {
		if m.remoteSelected[i] {
			selectedCount++
		}
	}

	content.WriteString(fmt.Sprintf("Importing %d commands...\n\n", selectedCount))

	// Import animation
	content.WriteString("📥 Processing selected commands...\n")
	content.WriteString("💾 Writing files to disk...\n")
	content.WriteString("🔄 Updating configuration...\n\n")

	content.WriteString(subtleStyle.Render("Please wait..."))

	return content.String()
}

// remoteResultsView renders the import results view
func (m Model) remoteResultsView() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Import Complete"))
	content.WriteString("\n\n")

	if m.remoteResult != nil {
		// Success summary
		if len(m.remoteResult.Imported) > 0 {
			content.WriteString("🎉 " + successStyle.Render(fmt.Sprintf("Successfully imported %d commands:", len(m.remoteResult.Imported))))
			content.WriteString("\n")
			for _, name := range m.remoteResult.Imported {
				content.WriteString(fmt.Sprintf("  ✅ %s\n", name))
			}
			content.WriteString("\n")
		}

		// Skipped summary
		if len(m.remoteResult.Skipped) > 0 {
			content.WriteString(fmt.Sprintf("⏭️  Skipped %d commands (already exist):\n", len(m.remoteResult.Skipped)))
			for _, name := range m.remoteResult.Skipped {
				content.WriteString(fmt.Sprintf("  ⚠️ %s\n", name))
			}
			content.WriteString("\n")
		}

		// Failed summary
		if len(m.remoteResult.Failed) > 0 {
			content.WriteString("❌ " + dangerStyle.Render(fmt.Sprintf("Failed to import %d commands:", len(m.remoteResult.Failed))))
			content.WriteString("\n")
			for i, name := range m.remoteResult.Failed {
				content.WriteString(fmt.Sprintf("  ❌ %s: %s\n", name, m.remoteResult.Errors[i]))
			}
			content.WriteString("\n")
		}

		if len(m.remoteResult.Imported) > 0 {
			content.WriteString(subtleStyle.Render("💡 Imported commands are now available in your command library."))
			content.WriteString("\n")
			content.WriteString(subtleStyle.Render("Use 'ccm' to manage them or enable/disable as needed."))
			content.WriteString("\n\n")
		}
	}

	footer := footerStyle.Render("Press any key to return to main menu")
	content.WriteString(footer)

	return content.String()
}