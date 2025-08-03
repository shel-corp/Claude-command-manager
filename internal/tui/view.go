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
	case StateRemoteImport:
		stateStr = "RemoteImport"
		return m.remoteImportView()
	case StateRemoteResults:
		stateStr = "RemoteResults"
		return m.remoteResultsView()
	}

	// Fallback with debug info
	return "DEBUG: Unknown state (" + stateStr + "), falling back to main menu\n\n" + m.mainMenuView()
}

// mainMenuView renders the main menu
func (m Model) mainMenuView() string {
	// Remove debug info now that TUI is working
	// debugInfo := fmt.Sprintf("DEBUG: MainMenu - Width: %d, Height: %d, ListItems: %d\n", 
	//	m.width, m.height, len(m.list.Items()))
	
	// Elegant header for Claude Command Manager with margins
	asciiHeader := `







    ╭──────────────────────────────────────────────────────╮
    │                                                      │
    │   ██████╗██╗      █████╗ ██╗   ██╗██████╗ ███████╗   │
    │  ██╔════╝██║     ██╔══██╗██║   ██║██╔══██╗██╔════╝   │
    │  ██║     ██║     ███████║██║   ██║██║  ██║█████╗     │
    │  ██║     ██║     ██╔══██║██║   ██║██║  ██║██╔══╝     │
    │  ╚██████╗███████╗██║  ██║╚██████╔╝██████╔╝███████╗   │
    │   ╚═════╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝   │
    │                                                      │
    │                  	Command Manager                    │
    │                                                      │
    ╰──────────────────────────────────────────────────────╯`

	// Always show ASCII header for now to test display
	finalHeader := asciiHeader
	
	content := m.list.View()
	footer := "↑/↓: Navigate • Enter: Select • q: Quit • h: Help"
	
	// Debug info to see what's happening
	debugInfo := fmt.Sprintf("DEBUG: Terminal size: %dx%d, Header length: %d chars", 
		m.width, m.height, len(finalHeader))
	
	// Clean layout with debug info for testing
	result := debugInfo + "\n" + finalHeader + "\n\n" + content + "\n\n" + footer
	
	return result
}

// libraryView renders the main application view
func (m Model) libraryView() string {
	// Header with library type
	libraryType := m.GetLibraryModeString()
	var icon string
	if m.libraryMode == LibraryModeUser {
		icon = "👤"
	} else {
		icon = "📁"
	}
	header := fmt.Sprintf("%s Command Library (%s)", icon, libraryType)
	content := m.list.View()
	footer := m.renderFooter()
	
	return centerView(header, content, footer, m.width)
}

// renameView renders the rename input view
func (m Model) renameView() string {
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

	footer := "Enter: Confirm • Esc: Cancel • Ctrl+C: Quit"
	
	return centerView(header, content.String(), footer, m.width)
}

// helpView renders the help screen
func (m Model) helpView() string {
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
func (m Model) renderFooter() string {
	if m.state == StateLibrary {
		return "Enter/t: Toggle • r: Rename • l: Location • s: Switch Library • i: Import • Esc: Main Menu • q: Quit • h: Help"
	}
	return "Enter/t: Toggle • r: Rename • l: Location • i: Browse/Import • q: Quit • h: Help"
}

// Remote import view functions

// remoteBrowseView renders the repository browsing interface
func (m Model) remoteBrowseView() string {
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
func (m Model) registryErrorView() string {
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
func (m Model) categoryBrowseView() string {
	header := "📋 Browse Command Repositories"
	
	var content strings.Builder
	content.WriteString(subtleStyle.Render("Select a category to explore available repositories:"))
	content.WriteString("\n\n")
	content.WriteString(m.list.View())
	
	footer := "Enter: Browse Category • /: Search • c: Custom URL • Esc: Cancel"
	
	return centerView(header, content.String(), footer, m.width)
}

// repositoryBrowseView renders the repository browsing interface
func (m Model) repositoryBrowseView() string {
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
func (m Model) searchBrowseView() string {
	header := "🔍 Search Repositories"
	
	var content strings.Builder
	// Search input
	content.WriteString("Search: ")
	content.WriteString(m.textInput.View())
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
	if m.textInput.Focused() {
		footer = "Tab: Switch to Results • Esc: Clear/Exit • Enter: Search"
	} else {
		footer = "Tab: Search Input • Enter: Browse Commands • c: Custom URL • Esc: Exit"
	}
	
	return centerView(header, content.String(), footer, m.width)
}

// remoteURLView renders the GitHub URL input view
func (m Model) remoteURLView() string {
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

	// Show error if present
	if m.remoteError != "" {
		content.WriteString("\n")
		content.WriteString(dangerStyle.Render("Error: " + m.remoteError))
	}

	footer := "Enter: Continue • Esc: Cancel • Ctrl+C: Quit"
	
	return centerView(header, content.String(), footer, m.width)
}

// remoteLoadingView renders the loading view
func (m Model) remoteLoadingView() string {
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

	// Loading animation
	content.WriteString("🔍 Connecting to repository...\n")
	content.WriteString("📦 Scanning for commands...\n")
	content.WriteString("🔄 Loading command details...\n")
	content.WriteString("⚠️  Checking for conflicts...\n\n")

	content.WriteString(subtleStyle.Render("Please wait..."))

	footer := ""
	
	return centerView(header, content.String(), footer, m.width)
}

// remoteSelectView renders the command selection view
func (m Model) remoteSelectView() string {
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

	footer := "Enter: Toggle • a: Select All • n: Select None • i: Import • Esc: Cancel"
	
	return centerView(header, content.String(), footer, m.width)
}

// remoteImportView renders the import progress view
func (m Model) remoteImportView() string {
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

// remoteResultsView renders the import results view
func (m Model) remoteResultsView() string {
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
func (m Model) remoteRepoDetailsView() string {
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

	footer := "Tab: Select Category • Enter: Continue • Esc: Cancel"
	
	return centerView(header, content.String(), footer, m.width)
}

// remoteCategoryView renders the category selection view
func (m Model) remoteCategoryView() string {
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
func (m Model) newCategoryCreationView() string {
	header := "Create New Category"
	
	var content strings.Builder
	content.WriteString(subtleStyle.Render("Create a new category for your repositories:"))
	content.WriteString("\n\n")
	
	content.WriteString("Category Name:\n")
	content.WriteString(m.textInput.View())
	content.WriteString("\n\n")
	
	content.WriteString(subtleStyle.Render("The category will be created with a default icon and description."))
	content.WriteString("\n")
	content.WriteString(subtleStyle.Render("You can customize these later."))

	footer := "Enter: Create Category • Esc: Back to Category List"
	
	return centerView(header, content.String(), footer, m.width)
}
