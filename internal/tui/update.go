package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	tea "github.com/charmbracelet/bubbletea"
	
	"github.com/shel-corp/Claude-command-manager/internal/registry"
	"github.com/shel-corp/Claude-command-manager/internal/remote"
)

// Message types for Bubble Tea
type (
	// RefreshMsg signals that the command list should be refreshed
	RefreshMsg struct{}
	
	// ErrorMsg carries error information
	ErrorMsg struct {
		Error error
	}
	
	// Remote import message types
	
	// RemoteLoadingMsg signals to start loading remote repository data
	RemoteLoadingMsg struct{}
	
	// RemoteLoadedMsg contains loaded remote repository data
	RemoteLoadedMsg struct {
		Commands []remote.RemoteCommand
		Error    string
	}
	
	// RemoteImportMsg signals to start importing selected commands
	RemoteImportMsg struct {
		Commands []remote.RemoteCommand
	}
	
	// RemoteImportCompleteMsg contains import results
	RemoteImportCompleteMsg struct {
		Result *remote.ImportResult
		Error  string
	}
)

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust list size to account for header and footer space
		headerLines := 20 // More lines needed for header with margins 
		footerLines := 5  // Lines for footer and spacing
		availableHeight := msg.Height - headerLines - footerLines
		if availableHeight < 3 {
			availableHeight = 3 // Minimum height for list
		}
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(availableHeight)
		return m, nil

	case RefreshMsg:
		if err := m.RefreshCommands(); err != nil {
			return m, func() tea.Msg {
				return ErrorMsg{Error: err}
			}
		}
		return m, nil

	case ErrorMsg:
		// For now, just ignore errors. In a more complete implementation,
		// we might show them in a status bar or popup
		return m, nil

	case RemoteLoadingMsg:
		return m.handleRemoteLoading()

	case RemoteLoadedMsg:
		return m.handleRemoteLoaded(msg)

	case RemoteImportMsg:
		return m.handleRemoteImport(msg)

	case RemoteImportCompleteMsg:
		return m.handleRemoteImportComplete(msg)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	// Handle updates based on current state
	switch m.state {
	case StateMainMenu:
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		
	case StateLibrary:
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		
	case StateRename:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		
	case StateRemoteBrowse:
		// Handle both list and text input based on browse mode
		if m.browseMode == BrowseModeSearch {
			m.textInput, cmd = m.textInput.Update(msg)
			cmds = append(cmds, cmd)
			// Update search results in real-time
			m.performSearch()
		} else {
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		}
		
	case StateRemoteURL:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		
	case StateRemoteRepoDetails:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		
	case StateRemoteCategory:
		if m.isNewCategory && m.selectedCategoryKey == "new" {
			// Handle text input for new category creation
			m.textInput, cmd = m.textInput.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			// Handle list navigation for category selection
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		}
		
	case StateRemoteSelect:
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		
	case StateRemoteLoading, StateRemoteImport:
		// No input handling during loading/import states
		
	case StateRemoteResults:
		// No input needed, just wait for user to exit
	}

	return m, tea.Batch(cmds...)
}

// handleKeyMsg handles keyboard input based on current state
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMainMenu:
		return m.handleMainMenuStateKeys(msg)
	case StateLibrary:
		return m.handleLibraryStateKeys(msg)
	case StateRename:
		return m.handleRenameStateKeys(msg)
	case StateHelp:
		return m.handleHelpStateKeys(msg)
	case StateRemoteBrowse:
		return m.handleRemoteBrowseStateKeys(msg)
	case StateRemoteURL:
		return m.handleRemoteURLStateKeys(msg)
	case StateRemoteRepoDetails:
		return m.handleRemoteRepoDetailsStateKeys(msg)
	case StateRemoteCategory:
		return m.handleRemoteCategoryStateKeys(msg)
	case StateRemoteSelect:
		return m.handleRemoteSelectStateKeys(msg)
	case StateRemoteResults:
		return m.handleRemoteResultsStateKeys(msg)
	}
	
	return *m, nil
}

// handleMainMenuStateKeys handles keys in the main menu state
func (m *Model) handleMainMenuStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return *m, m.Quit()
		
	case "enter":
		return m.executeSelectedMenuItem()
		
	case "1":
		m.state = StateLibrary
		return *m, nil
		
	case "2", "i":
		m.StartRemoteImport()
		return *m, nil
		
	case "h", "?":
		m.state = StateHelp
		return *m, nil
	}
	
	// Let the list handle other keys (navigation)
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return *m, cmd
}

// executeSelectedMenuItem executes the action for the selected menu item
func (m *Model) executeSelectedMenuItem() (tea.Model, tea.Cmd) {
	selectedItem := m.GetSelectedMenuItem()
	if selectedItem == nil {
		return *m, nil
	}
	
	switch selectedItem.action {
	case "library":
		// Switch to library view and refresh command list
		m.state = StateLibrary
		return *m, func() tea.Msg {
			return RefreshMsg{}
		}
	case "import":
		m.StartRemoteImport()
		return *m, nil
	}
	
	return *m, nil
}

// handleLibraryStateKeys handles keys in the library state
func (m *Model) handleLibraryStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return *m, m.Quit()
		
	case "esc":
		m.state = StateMainMenu
		m.initMainMenu()
		return *m, nil
		
	case "enter", "t":
		return *m, m.ToggleSelectedCommand()
		
	case "r":
		m.StartRename()
		return *m, nil
		
	case "l":
		return *m, m.ToggleSelectedCommandLocation()
		
	case "s":
		return *m, m.SwitchLibraryMode()
		
	case "i":
		m.StartRemoteImport()
		return *m, nil
		
	case "h", "?":
		m.state = StateHelp
		return *m, nil
	}
	
	// Let the list handle other keys (navigation)
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return *m, cmd
}

// handleRenameStateKeys handles keys in the rename state
func (m *Model) handleRenameStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return *m, m.ConfirmRename()
		
	case "esc":
		m.state = StateMainMenu
		m.initMainMenu()
		return *m, nil
		
	case "ctrl+c":
		return *m, m.Quit()
	}
	
	// Let text input handle other keys
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return *m, cmd
}

// handleHelpStateKeys handles keys in the help state
func (m *Model) handleHelpStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "h", "?", "q", "enter":
		m.state = StateMainMenu
		m.initMainMenu()
		return *m, nil
		
	case "ctrl+c":
		return *m, m.Quit()
	}
	
	return *m, nil
}

// Note: Confirm quit state removed since changes are saved immediately

// Remote import message handlers

func (m *Model) handleRemoteLoading() (tea.Model, tea.Cmd) {
	// Start async loading of remote repository data with caching
	return *m, func() tea.Msg {
		client := remote.NewGitHubClient()
		
		// Set cache manager if available
		if m.cacheManager != nil {
			client.SetCacheManager(m.cacheManager)
		}
		
		// Validate repository
		if err := client.ValidateRepository(m.remoteRepo); err != nil {
			return RemoteLoadedMsg{Error: err.Error()}
		}
		
		// Fetch commands with caching enabled
		if err := client.FetchCommandsWithCache(m.remoteRepo, true); err != nil {
			return RemoteLoadedMsg{Error: err.Error()}
		}
		
		// Load command details for commands that don't have content yet
		for i := range m.remoteRepo.Commands {
			if m.remoteRepo.Commands[i].Content == "" {
				if err := client.FetchCommandContent(m.remoteRepo, &m.remoteRepo.Commands[i]); err != nil {
					m.remoteRepo.Commands[i].Description = "Failed to load description"
				}
			}
		}
		
		// Check for local conflicts
		importer := remote.NewImporter("")
		homeDir, _ := os.UserHomeDir()
		targetDir := filepath.Join(homeDir, ".claude", "command_library")
		if err := importer.CheckLocalExists(m.remoteRepo.Commands, targetDir); err != nil {
			return RemoteLoadedMsg{Error: err.Error()}
		}
		
		return RemoteLoadedMsg{Commands: m.remoteRepo.Commands}
	}
}

func (m *Model) handleRemoteLoaded(msg RemoteLoadedMsg) (tea.Model, tea.Cmd) {
	m.remoteLoading = false
	
	if msg.Error != "" {
		m.remoteError = msg.Error
		m.state = StateRemoteURL
		return *m, nil
	}
	
	// Store commands and initialize selection state
	m.remoteCommands = msg.Commands
	m.remoteSelected = make(map[int]bool)
	
	// Transition to selection state
	m.state = StateRemoteSelect
	m.updateRemoteCommandList()
	
	return *m, nil
}

func (m *Model) handleRemoteImport(msg RemoteImportMsg) (tea.Model, tea.Cmd) {
	// Start async import process
	return *m, func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return RemoteImportCompleteMsg{Error: err.Error()}
		}
		
		targetDir := filepath.Join(homeDir, ".claude", "command_library")
		options := remote.GetDefaultImportOptions(targetDir)
		
		// Set overwrite based on conflicts - for now, default to overwrite
		options.OverwriteExisting = true
		
		importer := remote.NewImporter(targetDir)
		result, err := importer.ImportCommands(m.remoteRepo, msg.Commands, options)
		if err != nil {
			return RemoteImportCompleteMsg{Error: err.Error()}
		}
		
		return RemoteImportCompleteMsg{Result: result}
	}
}

func (m *Model) handleRemoteImportComplete(msg RemoteImportCompleteMsg) (tea.Model, tea.Cmd) {
	if msg.Error != "" {
		m.remoteError = msg.Error
		m.state = StateRemoteSelect
		return *m, nil
	}
	
	m.remoteResult = msg.Result
	m.state = StateRemoteResults
	
	return *m, nil
}

// Remote state key handlers

func (m *Model) handleRemoteBrowseStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle registry error case
	if m.registryManager == nil || !m.registryManager.IsLoaded() {
		return m.handleRegistryErrorKeys(msg)
	}
	
	switch m.browseMode {
	case BrowseModeCategories:
		return m.handleCategoryBrowseKeys(msg)
	case BrowseModeRepositories:
		return m.handleRepositoryBrowseKeys(msg)
	case BrowseModeSearch:
		return m.handleSearchKeys(msg)
	default:
		return *m, nil
	}
}

func (m *Model) handleRegistryErrorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		m.goToCustomURL()
		return *m, nil
		
	case "esc":
		m.state = StateMainMenu
		m.initMainMenu()
		return *m, nil
		
	case "ctrl+c":
		return *m, m.Quit()
	}
	
	return *m, nil
}

func (m *Model) handleCategoryBrowseKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.enterCategory()
		return *m, nil
		
	case "/", "s":
		m.startSearch()
		return *m, nil
		
	case "c":
		m.goToCustomURL()
		return *m, nil
		
	case "esc":
		m.state = StateMainMenu
		m.initMainMenu()
		return *m, nil
		
	case "ctrl+c":
		return *m, m.Quit()
	}
	
	// Let the list handle other keys (navigation)
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return *m, cmd
}

func (m *Model) handleRepositoryBrowseKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Load commands from the focused repository
		index := m.list.Index()
		if index < 0 || index >= len(m.filteredRepos) {
			// Show bounds error to user
			m.remoteError = "No repository selected. Please select a repository from the list."
			m.state = StateRemoteURL
			return *m, nil
		}
		focusedRepo := m.filteredRepos[index]
		return *m, m.importSingleRepository(focusedRepo)
		
	case " ":
		// Space also loads repository commands (alternative to Enter)
		index := m.list.Index()
		if index < 0 || index >= len(m.filteredRepos) {
			// Show bounds error to user
			m.remoteError = "No repository selected. Please select a repository from the list."
			m.state = StateRemoteURL
			return *m, nil
		}
		focusedRepo := m.filteredRepos[index]
		return *m, m.importSingleRepository(focusedRepo)
		
	case "/", "s":
		m.startSearch()
		return *m, nil
		
	case "c":
		m.goToCustomURL()
		return *m, nil
		
	case "esc":
		// Go back to categories
		m.browseMode = BrowseModeCategories
		m.currentCategory = ""
		m.updateBrowseList()
		return *m, nil
		
	case "ctrl+c":
		return *m, m.Quit()
	}
	
	// Let the list handle other keys (navigation)
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return *m, cmd
}

func (m *Model) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// If search results are showing and text input is not focused, load repository commands
		if len(m.filteredRepos) > 0 && !m.textInput.Focused() {
			index := m.list.Index()
			if index >= 0 && index < len(m.filteredRepos) {
				focusedRepo := m.filteredRepos[index]
				return *m, m.importSingleRepository(focusedRepo)
			} else {
				// Show bounds error to user
				m.remoteError = "No repository selected. Please select a repository from the search results."
				m.state = StateRemoteURL
			}
		}
		return *m, nil
		
	case "tab":
		// Switch focus between search input and results
		if m.textInput.Focused() {
			m.textInput.Blur()
		} else {
			m.textInput.Focus()
		}
		return *m, nil
		
	case "esc":
		if m.textInput.Value() != "" {
			// Clear search first
			m.textInput.SetValue("")
			m.performSearch()
		} else {
			// Exit search mode
			m.exitSearch()
		}
		return *m, nil
		
	// Removed multi-select functionality - repositories are now single-select
		
	case "c":
		m.goToCustomURL()
		return *m, nil
		
	case "ctrl+c":
		return *m, m.Quit()
	}
	
	// Let text input or list handle other keys based on focus
	if m.textInput.Focused() {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return *m, cmd
	} else {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return *m, cmd
	}
}

func (m *Model) handleRemoteURLStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return *m, m.ProcessRemoteURL()
		
	case "esc":
		m.state = StateMainMenu
		m.initMainMenu()
		return *m, nil
		
	case "ctrl+c":
		return *m, m.Quit()
	}
	
	// Let text input handle other keys
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return *m, cmd
}

func (m *Model) handleRemoteSelectStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.ToggleRemoteCommand()
		return *m, nil
		
	case "a":
		m.SelectAllRemoteCommands(true)
		return *m, nil
		
	case "n":
		m.SelectAllRemoteCommands(false)
		return *m, nil
		
	case "i":
		return *m, m.StartRemoteImportProcess()
		
	case "esc":
		m.state = StateMainMenu
		m.initMainMenu()
		return *m, nil
		
	case "ctrl+c":
		return *m, m.Quit()
	}
	
	// Let the list handle other keys (navigation)
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return *m, cmd
}

func (m *Model) handleRemoteResultsStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc", "q":
		return *m, m.ReturnToMain()
		
	case "ctrl+c":
		return *m, m.Quit()
	}
	
	return *m, nil
}

// handleRemoteRepoDetailsStateKeys handles keys in the repository details input state
func (m *Model) handleRemoteRepoDetailsStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Update description from input
		m.customRepoInput.Description = strings.TrimSpace(m.textInput.Value())
		
		// Check if category is already selected
		if m.customRepoInput.Category.CategoryKey != "" {
			// Category already selected, finalize the repository
			m.finalizeCustomRepository()
			return *m, func() tea.Msg {
				return RemoteLoadingMsg{}
			}
		} else {
			// Need to select category first
			m.startCategorySelection()
			return *m, nil
		}
		
	case "tab":
		// Update description and move to category selection
		m.customRepoInput.Description = strings.TrimSpace(m.textInput.Value())
		m.startCategorySelection()
		return *m, nil
		
	case "esc":
		m.state = StateRemoteURL
		m.goToCustomURL()
		return *m, nil
		
	case "ctrl+c":
		return *m, m.Quit()
	}
	
	// Let text input handle other keys
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return *m, cmd
}

// handleRemoteCategoryStateKeys handles keys in the category selection state
func (m *Model) handleRemoteCategoryStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.isNewCategory && m.selectedCategoryKey == "new" {
			// Creating new category - get the name from text input
			newCategoryName := strings.TrimSpace(m.textInput.Value())
			if newCategoryName == "" {
				return *m, nil // Don't proceed with empty name
			}
			
			// Create category key from name (lowercase, replace spaces with underscores)
			categoryKey := strings.ToLower(strings.ReplaceAll(newCategoryName, " ", "_"))
			
			// Set up the category input
			m.customRepoInput.Category = registry.CategoryInput{
				CategoryKey: categoryKey,
				IsNew:       true,
				Name:        newCategoryName,
				Description: fmt.Sprintf("Custom category: %s", newCategoryName),
				Icon:        "📦", // Default icon
			}
			
			// Finalize the repository
			m.finalizeCustomRepository()
			return *m, func() tea.Msg {
				return RemoteLoadingMsg{}
			}
		} else {
			// Selecting existing category
			m.confirmCategorySelection()
			m.finalizeCustomRepository()
			return *m, func() tea.Msg {
				return RemoteLoadingMsg{}
			}
		}
		
	case "esc":
		if m.isNewCategory && m.selectedCategoryKey == "new" {
			// Go back to category list from new category creation
			m.isNewCategory = false
			m.selectedCategoryKey = ""
			m.setupCategorySelection()
			return *m, nil
		} else {
			// Go back to repository details
			m.state = StateRemoteRepoDetails
			m.setupRepoDetailsInput()
			return *m, nil
		}
		
	case "ctrl+c":
		return *m, m.Quit()
	}
	
	// Handle different input contexts
	if m.isNewCategory && m.selectedCategoryKey == "new" {
		// Handle text input for new category creation
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return *m, cmd
	} else {
		// Handle list navigation for category selection
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return *m, cmd
	}
}