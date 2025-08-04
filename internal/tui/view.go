package tui

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
)

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// View renders the application UI
func (m *Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// Debug: Log current state and dimensions
	stateStr := "Unknown"
	switch m.state {
	case StateMainMenu:
		stateStr = "MainMenu"
		return m.mainMenuView()
	case StateLibrary:
		stateStr = "Library"
		return m.libraryView()
	case StateRename:
		stateStr = "Rename"
		return m.renameView()
	case StateHelp:
		stateStr = "Help"
		return m.helpView()
	case StateRemoteBrowse:
		stateStr = "RemoteBrowse"
		return m.remoteBrowseView()
	case StateRemoteURL:
		stateStr = "RemoteURL"
		return m.remoteURLView()
	case StateRemoteRepoDetails:
		stateStr = "RemoteRepoDetails"
		return m.remoteRepoDetailsView()
	case StateRemoteCategory:
		stateStr = "RemoteCategory"
		return m.remoteCategoryView()
	case StateRemoteLoading:
		stateStr = "RemoteLoading"
		return m.remoteLoadingView()
	case StateRemoteSelect:
		stateStr = "RemoteSelect"
		return m.remoteSelectView()
	case StateRemotePreview:
		stateStr = "RemotePreview"
		return m.remotePreviewView()
	case StateRemoteImport:
		stateStr = "RemoteImport"
		return m.remoteImportView()
	case StateRemoteResults:
		stateStr = "RemoteResults"
		return m.remoteResultsView()
	case StateReportIssue:
		stateStr = "ReportIssue"
		return m.reportIssueView()
	case StateSettings:
		stateStr = "Settings"
		return m.settingsView()
	case StateThemeSettings:
		stateStr = "ThemeSettings"
		return m.themeSettingsView()
	}

	// Fallback with debug info
	return "DEBUG: Unknown state (" + stateStr + "), falling back to main menu\n\n" + m.mainMenuView()
}

// mainMenuView renders the main menu
func (m *Model) mainMenuView() string {
	// Remove debug info now that TUI is working
	// debugInfo := fmt.Sprintf("DEBUG: MainMenu - Width: %d, Height: %d, ListItems: %d\n", 
	//	m.width, m.height, len(m.list.Items()))
	
	// Create styled header with consistent design language
	asciiHeader := `







 ██████╗██╗      █████╗ ██╗   ██╗██████╗ ███████╗
██╔════╝██║     ██╔══██╗██║   ██║██╔══██╗██╔════╝
██║     ██║     ███████║██║   ██║██║  ██║█████╗  
██║     ██║     ██╔══██║██║   ██║██║  ██║██╔══╝  
╚██████╗███████╗██║  ██║╚██████╔╝██████╔╝███████╗
 ╚═════╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝

Command Manager`

	var headerContent string
	if m.width < 80 { // Increased threshold since ASCII art is wide
		// Compact header for narrow terminals
		headerContent = "CLAUDE COMMANDS\nCommand Manager"
	} else {
		// Full ASCII art header for wider terminals  
		headerContent = asciiHeader
	}
	
	// Style the header with clean, borderless design
	headerStyle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Padding(2, 3).
		Margin(1, 0).
		Align(lipgloss.Center).
		Width(m.width - 10)
	
	// Apply styling and center the header
	finalHeader := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(headerStyle.Render(headerContent))
	
	// Get the menu content
	content := m.list.View()
	
	// Create an elegant footer with better styling
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		Background(lipgloss.Color("#0F172A")).
		Padding(1, 2).
		Margin(1, 0).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#334155")).
		Align(lipgloss.Center).
		Width(m.width - 10)
	
	footerText := "↑/↓ Navigate  •  Enter Select  •  q Quit  •  h Help"
	footer := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(footerStyle.Render(footerText))
	
	// Add extra spacing for better visual breathing room
	spacer := "\n\n"
	
	// Combine all elements with proper spacing
	result := finalHeader + spacer + content + spacer + footer
	
	return result
}

// libraryView renders the main application view
func (m *Model) libraryView() string {
	// Header with library type
	libraryType := m.GetLibraryModeString()
	var icon string
	if m.libraryMode == LibraryModeUser {
		icon = "👤"
	} else {
		icon = "📁"
	}
	header := fmt.Sprintf("%s Command Library (%s)", icon, libraryType)
	
	// Include status message and main content
	content := m.renderStatusMessage() + m.list.View()
	footer := m.renderFooter()
	
	return centerView(header, content, footer, m.width)
}

// renameView renders the rename input view
func (m *Model) renameView() string {
	header := "Rename Command"
	
	var content strings.Builder
	if len(m.commands) > m.renameIndex {
		cmd := m.commands[m.renameIndex]
		content.WriteString(fmt.Sprintf("Current name: %s\n", 
			highlightStyle.Render(cmd.DisplayName)))
		content.WriteString(fmt.Sprintf("Description: %s\n\n", 
			subtleStyle.Render(cmd.Description)))
	}

	content.WriteString("New name:\n")
	content.WriteString(m.textInput.View())
	
	// Show validation errors
	if errorMsg, hasError := m.validationErrors["name"]; hasError {
		content.WriteString("\n")
		content.WriteString(dangerStyle.Render("⚠️ " + errorMsg))
	}

	footer := "Enter: Confirm • Esc: Back to Library • Ctrl+C: Quit"
	
	return centerView(header, content.String(), footer, m.width)
}

// helpView renders the help screen
func (m *Model) helpView() string {
	header := "Help"
	
	var content strings.Builder
	helpItems := []struct {
		key  string
		desc string
	}{
		{"↑/↓, j/k", "Navigate up/down"},
		{"Enter, t", "Toggle command enabled/disabled"},
		{"r", "Rename selected command"},
		{"l", "Toggle symlink location (👤 user / 📁 project)"},
		{"s", "Switch library (👤 user / 📁 project)"},
		{"i", "Browse and import repository commands"},
		{"q", "Quit"},
		{"h, ?", "Show this help screen"},
		{"Ctrl+C", "Force quit"},
		{"", ""},
		{"Repository Browser:", ""},
		{"i", "Import focused repository (or selected repositories)"},
		{"Enter", "Select category or toggle repository selection"},
		{"/", "Search repositories"},
		{"c", "Enter custom GitHub URL"},
		{"a", "Select all repositories"},
		{"n", "Select none"},
		{"p", "Preview selected command"},
		{"Space", "Toggle repository selection"},
		{"Tab", "Switch between search and results"},
		{"Esc", "Go back or cancel"},
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

	footer := "Press any key to return"
	
	return centerView(header, content.String(), footer, m.width)
}

// Note: Confirm quit view removed since changes are saved immediately

// renderFooter renders the footer with key bindings
func (m *Model) renderFooter() string {
	if m.state == StateLibrary {
		return "Enter/t: Toggle • r: Rename • l: Location • s: Switch Library • i: Import • Esc: Main Menu • q: Quit • h: Help"
	}
	return "Enter/t: Toggle • r: Rename • l: Location • i: Browse/Import • q: Quit • h: Help"
}

// Remote import view functions

// remoteBrowseView renders the repository browsing interface
func (m *Model) remoteBrowseView() string {
	// Check if registry failed to load
	if m.registryManager == nil || !m.registryManager.IsLoaded() {
		return m.registryErrorView()
	}

	switch m.browseMode {
	case BrowseModeCategories:
		return m.categoryBrowseView()
	case BrowseModeRepositories:
		return m.repositoryBrowseView()
	case BrowseModeSearch:
		return m.searchBrowseView()
	default:
		return m.categoryBrowseView()
	}
}

// registryErrorView renders error when registry fails to load
func (m *Model) registryErrorView() string {
	header := "Repository Browser"
	
	var content strings.Builder
	content.WriteString(dangerStyle.Render("⚠️  Failed to load repository registry"))
	content.WriteString("\n\n")
	content.WriteString(subtleStyle.Render("The curated repository list could not be loaded."))
	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("You can still import from custom GitHub URLs."))

	footer := "c: Custom URL • Esc: Cancel • Ctrl+C: Quit"
	
	return centerView(header, content.String(), footer, m.width)
}

// categoryBrowseView renders the category browsing interface
func (m *Model) categoryBrowseView() string {
	header := "📋 Browse Command Repositories"
	
	var content strings.Builder
	content.WriteString(subtleStyle.Render("Select a category to explore available repositories:"))
	content.WriteString("\n\n")
	content.WriteString(m.list.View())
	
	footer := "Enter: Browse Category • /: Search • c: Custom URL • Esc: Cancel"
	
	return centerView(header, content.String(), footer, m.width)
}

// repositoryBrowseView renders the repository browsing interface
func (m *Model) repositoryBrowseView() string {
	// Header with category info
	categoryName := "All Repositories"
	categoryIcon := "📦"
	if m.currentCategory != "" {
		if categories := m.registryManager.GetCategories(); categories != nil {
			if cat, exists := categories[m.currentCategory]; exists {
				categoryName = cat.Name
				categoryIcon = cat.Icon
			}
		}
	}
	header := categoryIcon + " " + categoryName
	
	var content strings.Builder
	content.WriteString(subtleStyle.Render("Select a repository to browse its available commands:"))
	content.WriteString("\n\n")
	content.WriteString(m.list.View())

	footer := "Enter: Browse Commands • /: Search • c: Custom URL • Esc: Back"
	
	return centerView(header, content.String(), footer, m.width)
}

// searchBrowseView renders the search interface
func (m *Model) searchBrowseView() string {
	header := "🔍 Search Repositories"
	
	var content strings.Builder
	// Search input
	content.WriteString("Search: ")
	content.WriteString(m.searchInput.View())
	content.WriteString("\n\n")

	if m.searchQuery == "" {
		content.WriteString(subtleStyle.Render("Enter search terms to find repositories..."))
	} else {
		content.WriteString(fmt.Sprintf("Found %d repositories matching \"%s\"", 
			len(m.filteredRepos), m.searchQuery))
	}
	content.WriteString("\n\n")

	// Results list (if any)
	if len(m.filteredRepos) > 0 {
		content.WriteString(m.list.View())
	}

	// Instructions
	var footer string
	if m.searchInput.Focused() {
		footer = "Tab: Switch to Results • Esc: Clear/Exit • Enter: Search"
	} else {
		footer = "Tab: Search Input • Enter: Browse Commands • c: Custom URL • Esc: Exit"
	}
	
	return centerView(header, content.String(), footer, m.width)
}

// remoteURLView renders the GitHub URL input view
func (m *Model) remoteURLView() string {
	header := "Import Commands from GitHub"
	
	var content strings.Builder
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
	content.WriteString("\n")

	// Show validation errors
	if errorMsg, hasError := m.validationErrors["url"]; hasError {
		content.WriteString("\n")
		content.WriteString(dangerStyle.Render("⚠️ " + errorMsg))
	}

	// Show other errors if present
	if m.remoteError != "" {
		content.WriteString("\n")
		content.WriteString(dangerStyle.Render("Error: " + m.remoteError))
	}

	footer := "Enter: Continue • Esc: Back to Browse • Ctrl+C: Quit"
	
	return centerView(header, content.String(), footer, m.width)
}

// remoteLoadingView renders the loading view
func (m *Model) remoteLoadingView() string {
	header := "Loading Repository..."
	
	var content strings.Builder
	if m.remoteRepo != nil {
		content.WriteString(fmt.Sprintf("Repository: %s\n", 
			highlightStyle.Render(fmt.Sprintf("%s/%s", m.remoteRepo.Owner, m.remoteRepo.Repo))))
		content.WriteString(fmt.Sprintf("Branch: %s\n", 
			subtleStyle.Render(m.remoteRepo.Branch)))
		content.WriteString(fmt.Sprintf("Path: %s\n\n", 
			subtleStyle.Render(m.remoteRepo.Path)))
	}

	// Simple loading spinner
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	// Use a simple time-based animation (this is a simplified approach)
	spinnerChar := spinner[0] // In a real implementation, this would cycle
	
	content.WriteString(fmt.Sprintf("%s Connecting to repository...\n", spinnerChar))
	content.WriteString("📦 Scanning for commands...\n")
	content.WriteString("🔄 Loading command details...\n")
	content.WriteString("⚠️  Checking for conflicts...\n\n")

	content.WriteString(subtleStyle.Render("Please wait..."))

	footer := "Loading... Please wait"
	
	return centerView(header, content.String(), footer, m.width)
}

// remoteSelectView renders the command selection view
func (m *Model) remoteSelectView() string {
	header := "Select Commands to Import"
	
	var content strings.Builder
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

	footer := "Enter: Toggle • p: Preview • a: Select All • n: Select None • i: Import • Esc: Cancel"
	
	return centerView(header, content.String(), footer, m.width)
}

// remoteImportView renders the import progress view
func (m *Model) remoteImportView() string {
	header := "Importing Commands..."
	
	selectedCount := 0
	for i := range m.remoteCommands {
		if m.remoteSelected[i] {
			selectedCount++
		}
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("Importing %d commands...\n\n", selectedCount))

	// Import animation
	content.WriteString("📥 Processing selected commands...\n")
	content.WriteString("💾 Writing files to disk...\n")
	content.WriteString("🔄 Updating configuration...\n\n")

	content.WriteString(subtleStyle.Render("Please wait..."))

	footer := ""
	
	return centerView(header, content.String(), footer, m.width)
}

// remotePreviewView renders the command preview view
func (m *Model) remotePreviewView() string {
	if m.previewCommand == nil {
		return "No command to preview"
	}
	
	header := fmt.Sprintf("📄 Preview: %s", m.previewCommand.Name)
	
	var content strings.Builder
	
	// Command metadata
	content.WriteString(fmt.Sprintf("Name: %s\n", highlightStyle.Render(m.previewCommand.Name)))
	content.WriteString(fmt.Sprintf("Path: %s\n", subtleStyle.Render(m.previewCommand.Path)))
	content.WriteString(fmt.Sprintf("Description: %s\n", m.previewCommand.Description))
	if m.previewCommand.LocalExists {
		content.WriteString(warningStyle.Render("⚠️  A local command with this name already exists"))
		content.WriteString("\n")
	}
	content.WriteString("\n")
	
	// Content divider
	content.WriteString(strings.Repeat("─", min(m.width-4, 80)))
	content.WriteString("\n\n")
	
	// Command content
	if m.previewCommand.Content != "" {
		// Split content into lines and limit display height
		lines := strings.Split(m.previewCommand.Content, "\n")
		maxLines := m.height - 12 // Reserve space for header, metadata, and footer
		if maxLines < 5 {
			maxLines = 5
		}
		
		displayLines := lines
		if len(lines) > maxLines {
			displayLines = lines[:maxLines]
			// Add truncation indicator
			displayLines = append(displayLines, subtleStyle.Render("... (content truncated)"))
		}
		
		for _, line := range displayLines {
			content.WriteString(line)
			content.WriteString("\n")
		}
	} else {
		content.WriteString(subtleStyle.Render("Content not loaded"))
		content.WriteString("\n")
	}
	
	footer := "p/Esc: Back • Ctrl+C: Quit"
	
	return centerView(header, content.String(), footer, m.width)
}

// remoteResultsView renders the import results view
func (m *Model) remoteResultsView() string {
	header := "Import Complete"
	
	var content strings.Builder
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
		}
	}

	footer := "Press any key to return to main menu"
	
	return centerView(header, content.String(), footer, m.width)
}

// remoteRepoDetailsView renders the repository details input view
func (m *Model) remoteRepoDetailsView() string {
	header := "Repository Details"
	
	var content strings.Builder
	
	// Show repository URL and auto-detected info
	content.WriteString(fmt.Sprintf("URL: %s\n", 
		highlightStyle.Render(m.customRepoInput.URL)))
	content.WriteString(fmt.Sprintf("Name: %s\n", 
		highlightStyle.Render(m.customRepoInput.Name)))
	content.WriteString(fmt.Sprintf("Author: %s\n\n", 
		subtleStyle.Render(m.customRepoInput.Author)))
	
	// Description input
	content.WriteString("Description:\n")
	content.WriteString(m.textInput.View())
	
	// Show validation errors
	if errorMsg, hasError := m.validationErrors["description"]; hasError {
		content.WriteString("\n")
		content.WriteString(dangerStyle.Render("⚠️ " + errorMsg))
	}
	content.WriteString("\n\n")
	
	// Show current category selection status
	if m.customRepoInput.Category.CategoryKey != "" {
		if m.customRepoInput.Category.IsNew {
			content.WriteString(fmt.Sprintf("Category: %s (new)\n", 
				highlightStyle.Render(m.customRepoInput.Category.Name)))
		} else {
			categoryName := m.availableCategories[m.customRepoInput.Category.CategoryKey]
			content.WriteString(fmt.Sprintf("Category: %s\n", 
				highlightStyle.Render(categoryName)))
		}
	} else {
		content.WriteString(subtleStyle.Render("Press Tab to select category"))
		content.WriteString("\n")
	}

	// Show error if present
	if m.remoteError != "" {
		content.WriteString("\n")
		content.WriteString(dangerStyle.Render("Error: " + m.remoteError))
	}

	footer := "Tab: Select Category • Enter: Continue • Esc: Back to URL"
	
	return centerView(header, content.String(), footer, m.width)
}

// remoteCategoryView renders the category selection view
func (m *Model) remoteCategoryView() string {
	if m.isNewCategory && m.selectedCategoryKey == "new" {
		// Show new category creation
		return m.newCategoryCreationView()
	}
	
	// Show category selection list
	header := "Select Category"
	
	var content strings.Builder
	content.WriteString(subtleStyle.Render("Choose a category for your repository:"))
	content.WriteString("\n\n")
	content.WriteString(m.list.View())

	footer := "Enter: Select • Esc: Back"
	
	return centerView(header, content.String(), footer, m.width)
}

// newCategoryCreationView renders the new category creation view
func (m *Model) newCategoryCreationView() string {
	header := "Create New Category"
	
	var content strings.Builder
	content.WriteString(subtleStyle.Render("Create a new category for your repositories:"))
	content.WriteString("\n\n")
	
	content.WriteString("Category Name:\n")
	content.WriteString(m.categoryInput.View())
	
	// Show validation errors
	if errorMsg, hasError := m.validationErrors["category"]; hasError {
		content.WriteString("\n")
		content.WriteString(dangerStyle.Render("⚠️ " + errorMsg))
	}
	content.WriteString("\n\n")
	
	content.WriteString(subtleStyle.Render("The category will be created with a default icon and description."))
	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("You can customize these later."))

	footer := "Enter: Create Category • Esc: Back to Category List"
	
	return centerView(header, content.String(), footer, m.width)
}

// reportIssueView renders the report issue form
func (m *Model) reportIssueView() string {
	header := "Request Feature or Report Issue"
	
	var content strings.Builder
	content.WriteString(subtleStyle.Render("Help us improve ccm by reporting bugs or requesting features:"))
	content.WriteString("\n\n")
	
	// Field 1: Issue Title
	titleStyle := subtleStyle
	if m.issueCurrentField == 0 {
		titleStyle = highlightStyle
	}
	content.WriteString(titleStyle.Render("Issue Title:"))
	content.WriteString("\n")
	content.WriteString(m.issueTitleInput.View())
	content.WriteString("\n")
	
	// Show validation errors for title
	if errorMsg, hasError := m.validationErrors["title"]; hasError {
		content.WriteString(dangerStyle.Render("⚠️ " + errorMsg))
		content.WriteString("\n")
	}
	content.WriteString("\n")
	
	// Field 2: Issue Body
	bodyStyle := subtleStyle
	if m.issueCurrentField == 1 {
		bodyStyle = highlightStyle
	}
	content.WriteString(bodyStyle.Render("Issue Description:"))
	content.WriteString("\n")
	content.WriteString(m.issueBodyInput.View())
	content.WriteString("\n")
	
	// Show validation errors for body
	if errorMsg, hasError := m.validationErrors["body"]; hasError {
		content.WriteString(dangerStyle.Render("⚠️ " + errorMsg))
		content.WriteString("\n")
	}
	content.WriteString("\n")
	
	// Show submit error if present
	if m.issueSubmitError != "" {
		content.WriteString(dangerStyle.Render("Error: " + m.issueSubmitError))
		content.WriteString("\n\n")
	}
	
	// Show submission status
	if m.issueSubmitting {
		content.WriteString("📤 Submitting issue...")
		content.WriteString("\n")
	}
	
	footer := "Tab: Switch Field • Enter: Submit • Esc: Cancel • Ctrl+C: Quit"
	
	return centerView(header, content.String(), footer, m.width)
}

// settingsView renders the main settings menu
func (m *Model) settingsView() string {
	header := "⚙️ Settings"
	
	var content strings.Builder
	content.WriteString(subtleStyle.Render("Configure themes and preferences:"))
	content.WriteString("\n\n")
	content.WriteString(m.list.View())
	
	footer := "Enter: Select • Esc: Back to Main Menu • q: Quit • h: Help"
	
	return centerView(header, content.String(), footer, m.width)
}

// themeSettingsView renders the theme picker
func (m *Model) themeSettingsView() string {
	header := "🎨 Choose Theme"
	
	var content strings.Builder
	content.WriteString(subtleStyle.Render("Select a theme to customize your experience:"))
	content.WriteString("\n\n")
	
	// Show current theme info
	themeManager := GetThemeManager()
	currentTheme := themeManager.GetCurrentTheme()
	content.WriteString(fmt.Sprintf("Current: %s\n", highlightStyle.Render(currentTheme.Name)))
	content.WriteString(fmt.Sprintf("%s\n\n", subtleStyle.Render(currentTheme.Description)))
	
	// Theme list
	content.WriteString(m.list.View())
	
	// Show theme preview if available
	if len(themeManager.GetAvailableThemes()) > 0 {
		themes := themeManager.GetAvailableThemes()
		selectedIndex := m.list.Index()
		if selectedIndex >= 0 && selectedIndex < len(themes) {
			selectedTheme := themes[selectedIndex]
			preview := selectedTheme.GeneratePreview()
			content.WriteString("\n")
			content.WriteString("Preview: " + preview.ColorBar)
		}
	}
	
	footer := "Enter: Apply Theme • p: Preview • Esc: Back to Settings • q: Quit"
	
	return centerView(header, content.String(), footer, m.width)
}

// renderStatusMessage renders a status message if one is set
func (m *Model) renderStatusMessage() string {
	if !m.showStatus || m.statusMessage == "" {
		return ""
	}
	
	style := m.getStatusStyle()
	return "\n" + style.Render("● " + m.statusMessage) + "\n"
}
