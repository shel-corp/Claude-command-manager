package tui

import (
	"fmt"
	"io"
	"strings"
	
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/shel-corp/Claude-command-manager/internal/cache"
	"github.com/shel-corp/Claude-command-manager/internal/commands"
	"github.com/shel-corp/Claude-command-manager/internal/config"
	"github.com/shel-corp/Claude-command-manager/internal/registry"
	"github.com/shel-corp/Claude-command-manager/internal/remote"
)

// State represents the current application state
type State int

const (
	StateMainMenu State = iota
	StateLibrary
	StateRename
	StateHelp
	StateRemoteBrowse
	StateRemoteURL
	StateRemoteRepoDetails  // Repository details input
	StateRemoteCategory     // Category selection
	StateRemoteLoading
	StateRemoteSelect
	StateRemotePreview      // Command preview
	StateRemoteImport
	StateRemoteResults
	StateReportIssue        // Report issue form
)

// BrowseMode represents the current browsing mode in the repository browser
type BrowseMode int

const (
	BrowseModeCategories BrowseMode = iota
	BrowseModeRepositories
	BrowseModeSearch
)

// LibraryMode represents which command library is currently being viewed
type LibraryMode int

const (
	LibraryModeProject LibraryMode = iota // Current project's command library
	LibraryModeUser                       // User's home command library
)

// StatusType represents the type of status message
type StatusType int

const (
	StatusInfo StatusType = iota
	StatusSuccess
	StatusError
	StatusWarning
)

// CustomDelegate is a custom list delegate that removes the active line indicator
type CustomDelegate struct {
	list.DefaultDelegate
}

// NewCustomDelegate creates a new custom delegate
func NewCustomDelegate() CustomDelegate {
	d := CustomDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}
	return d
}

// Render renders the list item with elegant styling and spacing
func (d CustomDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var str string
	
	// Check if this item is selected
	isSelected := index == m.Index()
	
	// Get the item content
	title := item.(interface{ Title() string }).Title()
	desc := item.(interface{ Description() string }).Description()
	
	// Calculate content width (leave margins for centering)
	contentWidth := m.Width() - 20 // Leave space for margins
	if contentWidth < 40 {
		contentWidth = 40
	}
	
	if isSelected {
		// Selected item with elegant card-like appearance
		cardStyle := lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Background(lipgloss.Color("#1E293B")).
			Padding(0, 2).  // Reduced from (1, 3) to save vertical space 
			Margin(0, 0)    // Reduced from (1, 0) to save vertical space
		
		titleStyle := lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Align(lipgloss.Center)
		
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CBD5E1")).
			Italic(true).
			Align(lipgloss.Center)
		
		content := titleStyle.Render(title)
		if desc != "" {
			content += "\n" + descStyle.Render(desc)
		}
		
		card := cardStyle.Render(content)
		
		// Center the entire card
		centerStyle := lipgloss.NewStyle().
			Width(m.Width()).
			Align(lipgloss.Center)
		
		str = centerStyle.Render(card) + "\n" // Add spacing after selected card
		
	} else {
		// Unselected item with subtle styling
		itemStyle := lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Padding(0, 2).  // Reduced from (1, 3) to save vertical space
			Margin(0, 0)    // Reduced from (1, 0) to save vertical space
		
		titleStyle := lipgloss.NewStyle().
			Foreground(textColor).
			Align(lipgloss.Center)
		
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Align(lipgloss.Center)
		
		content := titleStyle.Render(title)
		if desc != "" {
			content += "\n" + descStyle.Render(desc)
		}
		
		item := itemStyle.Render(content)
		
		// Center the entire item
		centerStyle := lipgloss.NewStyle().
			Width(m.Width()).
			Align(lipgloss.Center)
		
		str = centerStyle.Render(item) + "\n" // Add spacing after unselected card
	}
	
	fmt.Fprint(w, str)
}

// Model represents the application state for Bubble Tea
type Model struct {
	// Core components - separate instances for different contexts
	list           list.Model        // Main list for navigation
	textInput      textinput.Model   // Primary text input
	searchInput    textinput.Model   // Dedicated search input
	categoryInput  textinput.Model   // Category creation input
	issueTitleInput textinput.Model  // Issue title input
	issueBodyInput  textinput.Model  // Issue body input
	
	// Managers
	commandManager     *commands.Manager
	configManager      *config.Manager
	userCommandManager *commands.Manager
	userConfigManager  *config.Manager
	cacheManager       *cache.Manager
	
	// Application state
	state          State
	commands       []commands.Command
	libraryMode    LibraryMode
	
	// UI state
	width          int
	height         int
	quitting       bool
	
	// Rename state
	renameIndex    int
	renameOriginal string
	
	// Remote import state
	remoteURL       string
	remoteRepo      *remote.RemoteRepository
	remoteCommands  []remote.RemoteCommand
	remoteLoading   bool
	remoteError     string
	remoteSelected  map[int]bool
	remoteConflicts []remote.RemoteCommand
	remoteOptions   remote.ImportOptions
	remoteResult    *remote.ImportResult
	
	// Preview state
	previewCommand  *remote.RemoteCommand
	previousState   State  // State to return to after preview
	
	// Custom repository input state
	customRepoInput     registry.RepositoryInput
	availableCategories map[string]string  // key -> name mapping
	selectedCategoryKey string
	isNewCategory       bool
	
	// Input validation state
	validationErrors    map[string]string  // field -> error message
	
	// UI feedback state
	statusMessage       string             // Status message to display
	statusType          StatusType         // Type of status (info, success, error)
	showStatus          bool               // Whether to show status message
	
	// Report issue state
	issueCurrentField   int                // Current field in report issue form (0=title, 1=body)
	issueSubmitting     bool               // Whether currently submitting issue
	issueSubmitError    string             // Error from issue submission
	
	// Repository browsing state
	registryManager    *registry.EnhancedRegistryManager
	browseMode         BrowseMode
	currentCategory    string
	searchQuery        string
	filteredRepos      []remote.CuratedRepository
	browseSelected     map[int]bool
}

// commandItem implements list.Item for the Bubbles list component
type commandItem struct {
	command commands.Command
}

func (i commandItem) FilterValue() string {
	return i.command.DisplayName
}

func (i commandItem) Title() string {
	status := "[ ]"
	if i.command.Enabled {
		status = "[✓]"
	}
	
	// Add location decorator
	var locationIcon string
	switch i.command.SymlinkLocation {
	case config.SymlinkLocationUser:
		locationIcon = "👤" // User icon
	case config.SymlinkLocationProject:
		locationIcon = "📁" // Project folder icon
	default:
		locationIcon = "👤" // Default to user
	}
	
	return status + " " + locationIcon + " " + i.command.DisplayName
}

func (i commandItem) Description() string {
	return i.command.Description
}

// remoteCommandItem implements list.Item for remote commands with selection support
type remoteCommandItem struct {
	command  remote.RemoteCommand
	selected bool
	index    int
}

func (i remoteCommandItem) FilterValue() string {
	return i.command.Name
}

func (i remoteCommandItem) Title() string {
	// Selection checkbox
	checkbox := "[ ]"
	if i.selected {
		checkbox = "[✓]"
	}
	
	// Conflict indicator
	conflictIcon := ""
	if i.command.LocalExists {
		conflictIcon = " ⚠️"
	}
	
	return checkbox + " " + i.command.Name + conflictIcon
}

func (i remoteCommandItem) Description() string {
	status := ""
	if i.command.LocalExists {
		status = "(exists locally) "
	}
	return status + i.command.Description
}

// categoryItem implements list.Item for repository categories
type categoryItem struct {
	key      string
	category remote.RepositoryCategory
}

func (i categoryItem) FilterValue() string {
	return i.category.Name
}

func (i categoryItem) Title() string {
	return i.category.Icon + " " + i.category.Name
}

func (i categoryItem) Description() string {
	return i.category.Description
}

// repositoryItem implements list.Item for curated repositories
type repositoryItem struct {
	repository remote.CuratedRepository
	selected   bool
	index      int
}

// categorySelectionItem implements list.Item for category selection
type categorySelectionItem struct {
	key   string
	name  string
	isNew bool
}

func (i categorySelectionItem) FilterValue() string {
	return i.name
}

func (i categorySelectionItem) Title() string {
	if i.isNew {
		return "➕ " + i.name
	}
	return i.name
}

func (i categorySelectionItem) Description() string {
	if i.isNew {
		return "Create a new category for your repositories"
	}
	return "Existing category"
}

// menuItem implements list.Item for main menu options
type menuItem struct {
	title       string
	description string
	icon        string
	action      string
}

func (i menuItem) FilterValue() string {
	return i.title
}

func (i menuItem) Title() string {
	return i.icon + " " + i.title
}

func (i menuItem) Description() string {
	return i.description
}

func (i repositoryItem) FilterValue() string {
	return i.repository.Name
}

func (i repositoryItem) Title() string {
	// Verified badge
	verifiedBadge := ""
	if i.repository.Verified {
		verifiedBadge = " ✅"
	}
	
	return i.repository.Name + verifiedBadge
}

func (i repositoryItem) Description() string {
	return i.repository.Description + " • by " + i.repository.Author
}

// NewModel creates a new TUI model
func NewModel(commandManager *commands.Manager, configManager *config.Manager, userCommandManager *commands.Manager, userConfigManager *config.Manager) (*Model, error) {
	// Initialize cache manager
	cacheConfig := cache.DefaultCacheConfig()
	cacheManager, err := cache.NewManager(cacheConfig)
	if err != nil {
		// Log error but don't fail - caching is optional
		fmt.Printf("Warning: failed to initialize cache manager: %v\n", err)
		cacheManager = nil
	}
	// Initialize text inputs for different contexts
	ti := textinput.New()
	ti.Placeholder = "Enter new name..."
	ti.CharLimit = 300
	ti.Width = 60
	
	searchInput := textinput.New()
	searchInput.Placeholder = "Search repositories..."
	searchInput.CharLimit = 100
	searchInput.Width = 60
	
	categoryInput := textinput.New()
	categoryInput.Placeholder = "Enter category name..."
	categoryInput.CharLimit = 50
	categoryInput.Width = 60
	
	// Initialize report issue inputs
	issueTitleInput := textinput.New()
	issueTitleInput.Placeholder = "Enter issue title..."
	issueTitleInput.CharLimit = 100
	issueTitleInput.Width = 60
	
	issueBodyInput := textinput.New()
	issueBodyInput.Placeholder = "Describe the issue in detail..."
	issueBodyInput.CharLimit = 2000
	issueBodyInput.Width = 60

	// Initialize list with custom delegate to remove default styling
	delegate := NewCustomDelegate()
	delegate.SetHeight(3) // Account for card height (title + description + border)
	delegate.SetSpacing(1) // Add spacing between cards
	delegate.ShowDescription = true
	
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	
	// Remove default list styling
	l.SetShowTitle(false)
	l.SetShowPagination(true) // Enable pagination to handle overflow gracefully

	// Initialize enhanced registry manager with cache support
	registryManager, err := registry.NewEnhancedRegistryManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create registry manager: %w", err)
	}
	if cacheManager != nil {
		registryManager.SetCacheManager(cacheManager)
	}
	if err := registryManager.LoadRegistries(); err != nil {
		// Log error but don't fail - user can still use custom URLs
		fmt.Printf("Warning: failed to load registries: %v\n", err)
	}

	model := &Model{
		list:               l,
		textInput:          ti,
		searchInput:        searchInput,
		categoryInput:      categoryInput,
		issueTitleInput:    issueTitleInput,
		issueBodyInput:     issueBodyInput,
		commandManager:     commandManager,
		configManager:      configManager,
		userCommandManager: userCommandManager,
		userConfigManager:  userConfigManager,
		cacheManager:       cacheManager,
		state:              StateMainMenu,
		libraryMode:        LibraryModeProject, // Start with project library
		registryManager:    registryManager,
		browseSelected:      make(map[int]bool),
		availableCategories: make(map[string]string),
		customRepoInput:     registry.RepositoryInput{},
		validationErrors:    make(map[string]string),
	}

	// Load commands
	if err := model.RefreshCommands(); err != nil {
		return nil, err
	}
	
	// Initialize main menu since we start in StateMainMenu
	model.initMainMenu()

	// Start background cache refresh if caching is enabled
	if cacheManager != nil && cacheManager.IsEnabled() {
		// Note: We'll need to update the BackgroundRefresh method to work with enhanced registry manager
		// For now, skip this - the enhanced registry manager will handle caching internally
	}

	return model, nil
}

// initMainMenu initializes the main menu list
func (m *Model) initMainMenu() {
	items := []list.Item{
		menuItem{
			title:       "Library",
			description: "Manage your command library",
			icon:        "",
			action:      "library",
		},
		menuItem{
			title:       "Import",
			description: "Browse and import repository commands",
			icon:        "",
			action:      "import",
		},
		menuItem{
			title:       "Request feature or report issue",
			description: "Report a bug or request a feature",
			icon:        "",
			action:      "report_issue",
		},
	}
	
	m.list.SetItems(items)
}

// GetSelectedMenuItem returns the currently selected menu item
func (m *Model) GetSelectedMenuItem() *menuItem {
	index := m.list.Index()
	items := m.list.Items()
	
	if index < 0 || index >= len(items) {
		return nil
	}
	
	if item, ok := items[index].(menuItem); ok {
		return &item
	}
	
	return nil
}

// getCurrentCommandManager returns the current command manager based on library mode
func (m *Model) getCurrentCommandManager() *commands.Manager {
	if m.libraryMode == LibraryModeUser {
		return m.userCommandManager
	}
	return m.commandManager
}

// getCurrentConfigManager returns the current config manager based on library mode
func (m *Model) getCurrentConfigManager() *config.Manager {
	if m.libraryMode == LibraryModeUser {
		return m.userConfigManager
	}
	return m.configManager
}

// SwitchLibraryMode toggles between project and user library modes
func (m *Model) SwitchLibraryMode() tea.Cmd {
	if m.libraryMode == LibraryModeProject {
		m.libraryMode = LibraryModeUser
	} else {
		m.libraryMode = LibraryModeProject
	}
	
	// Refresh commands for the new library
	return func() tea.Msg {
		return RefreshMsg{}
	}
}

// GetLibraryModeString returns a human-readable string for the current library mode
func (m *Model) GetLibraryModeString() string {
	if m.libraryMode == LibraryModeUser {
		return "User"
	}
	return "Project"
}

// RefreshCommands reloads commands from disk and updates the list
func (m *Model) RefreshCommands() error {
	currentManager := m.getCurrentCommandManager()
	cmds, err := currentManager.ScanCommands()
	if err != nil {
		return err
	}

	m.commands = cmds

	// Convert to list items
	items := make([]list.Item, len(cmds))
	for i, cmd := range cmds {
		items[i] = commandItem{command: cmd}
	}

	m.list.SetItems(items)
	return nil
}

// GetSelectedCommand returns the currently selected command
func (m *Model) GetSelectedCommand() *commands.Command {
	if len(m.commands) == 0 {
		return nil
	}
	
	index := m.list.Index()
	if index < 0 || index >= len(m.commands) {
		return nil
	}
	
	return &m.commands[index]
}

// Note: Session change tracking removed - all changes now save immediately

// ToggleSelectedCommand toggles the enabled state of the selected command and saves immediately
func (m *Model) ToggleSelectedCommand() tea.Cmd {
	cmd := m.GetSelectedCommand()
	if cmd == nil {
		return nil
	}

	currentCommandManager := m.getCurrentCommandManager()
	currentConfigManager := m.getCurrentConfigManager()
	
	var err error
	wasEnabled := cmd.Enabled

	if cmd.Enabled {
		err = currentCommandManager.DisableCommand(*cmd)
	} else {
		err = currentCommandManager.EnableCommand(*cmd)
	}

	if err != nil {
		return func() tea.Msg {
			return ErrorMsg{Error: err}
		}
	}

	// Save configuration immediately
	if err := currentConfigManager.Save(); err != nil {
		return func() tea.Msg {
			return ErrorMsg{Error: err}
		}
	}
	
	// Set success status message  
	if wasEnabled {
		m.setStatus(fmt.Sprintf("Disabled command: %s", cmd.DisplayName), StatusSuccess)
	} else {
		m.setStatus(fmt.Sprintf("Enabled command: %s", cmd.DisplayName), StatusSuccess)
	}
	
	return func() tea.Msg {
		return RefreshMsg{}
	}
}

// StartRename initiates the rename process for the selected command
func (m *Model) StartRename() {
	cmd := m.GetSelectedCommand()
	if cmd == nil {
		return
	}

	m.state = StateRename
	m.renameIndex = m.list.Index()
	m.renameOriginal = cmd.DisplayName
	m.textInput.SetValue(cmd.DisplayName)
	m.textInput.Focus()
}

// ConfirmRename completes the rename process and saves immediately
func (m *Model) ConfirmRename() tea.Cmd {
	newName := m.textInput.Value()
	if newName == "" || newName == m.renameOriginal {
		m.state = StateLibrary
		return nil
	}

	currentCommandManager := m.getCurrentCommandManager()
	currentConfigManager := m.getCurrentConfigManager()
	
	cmd := &m.commands[m.renameIndex]
	err := currentCommandManager.RenameCommand(*cmd, newName)
	if err != nil {
		m.state = StateLibrary
		return func() tea.Msg {
			return ErrorMsg{Error: err}
		}
	}

	// Save configuration immediately
	if err := currentConfigManager.Save(); err != nil {
		m.state = StateLibrary
		return func() tea.Msg {
			return ErrorMsg{Error: err}
		}
	}

	m.state = StateLibrary

	return func() tea.Msg {
		return RefreshMsg{}
	}
}


// ToggleSelectedCommandLocation toggles the symlink location of the selected command and saves immediately
func (m *Model) ToggleSelectedCommandLocation() tea.Cmd {
	cmd := m.GetSelectedCommand()
	if cmd == nil {
		return nil
	}

	currentCommandManager := m.getCurrentCommandManager()
	currentConfigManager := m.getCurrentConfigManager()
	
	err := currentCommandManager.ToggleSymlinkLocation(*cmd)
	if err != nil {
		return func() tea.Msg {
			return ErrorMsg{Error: err}
		}
	}

	// Save configuration immediately
	if err := currentConfigManager.Save(); err != nil {
		return func() tea.Msg {
			return ErrorMsg{Error: err}
		}
	}
	
	return func() tea.Msg {
		return RefreshMsg{}
	}
}

// Quit exits the application immediately (no need for save confirmation since changes are saved immediately)
func (m *Model) Quit() tea.Cmd {
	m.quitting = true
	return tea.Sequence(
		tea.ExitAltScreen,
		tea.Quit,
	)
}

// Report issue methods

// StartReportIssue initiates the report issue flow
func (m *Model) StartReportIssue() {
	m.state = StateReportIssue
	m.issueCurrentField = 0 // Start with title field
	m.issueSubmitting = false
	m.issueSubmitError = ""
	
	// Clear and focus title input
	m.issueTitleInput.SetValue("")
	m.issueBodyInput.SetValue("")
	m.issueTitleInput.Focus()
	m.issueBodyInput.Blur()
	
	// Clear validation errors
	m.clearValidationErrors()
}

// Remote import methods

// StartRemoteImport initiates the remote import flow
func (m *Model) StartRemoteImport() {
	m.state = StateRemoteBrowse
	m.browseMode = BrowseModeCategories
	m.currentCategory = ""
	m.searchQuery = ""
	
	// Reset remote state
	m.remoteURL = ""
	m.remoteRepo = nil
	m.remoteCommands = nil
	m.remoteLoading = false
	m.remoteError = ""
	m.remoteSelected = make(map[int]bool)
	m.remoteConflicts = nil
	m.remoteResult = nil
	m.browseSelected = make(map[int]bool)
	
	// Initialize with categories if registry is loaded, otherwise try to load it
	if m.registryManager != nil {
		if !m.registryManager.IsLoaded() {
			// Try to load the registry again
			m.registryManager.LoadRegistries()
		}
		
		if m.registryManager.IsLoaded() {
			m.updateBrowseList()
			// Reset list selection
			m.list.Select(0)
		}
		// If registry still fails to load, the view will show the error screen
	}
}

// ProcessRemoteURL validates and processes the entered repository URL
func (m *Model) ProcessRemoteURL() tea.Cmd {
	url := strings.TrimSpace(m.textInput.Value())
	if url == "" {
		return nil
	}
	
	// Parse the GitHub URL to validate it
	_, err := remote.ParseGitHubURL(url)
	if err != nil {
		m.remoteError = err.Error()
		return nil
	}
	
	// Start the enhanced custom repository flow
	m.startCustomRepoFlow(url)
	return nil
}

// ToggleRemoteCommand toggles selection of a remote command
func (m *Model) ToggleRemoteCommand() {
	if m.state != StateRemoteSelect {
		return
	}
	
	index := m.list.Index()
	if index < 0 || index >= len(m.remoteCommands) {
		return
	}
	
	// Toggle selection state
	m.remoteSelected[index] = !m.remoteSelected[index]
	
	// Update list items
	m.updateRemoteCommandList()
}

// SelectAllRemoteCommands selects or deselects all remote commands
func (m *Model) SelectAllRemoteCommands(selectAll bool) {
	if m.state != StateRemoteSelect {
		return
	}
	
	for i := range m.remoteCommands {
		m.remoteSelected[i] = selectAll
	}
	
	m.updateRemoteCommandList()
}

// GetSelectedRemoteCommands returns the currently selected remote commands
func (m *Model) GetSelectedRemoteCommands() []remote.RemoteCommand {
	var selected []remote.RemoteCommand
	for i, command := range m.remoteCommands {
		if m.remoteSelected[i] {
			cmd := command
			cmd.Selected = true
			selected = append(selected, cmd)
		}
	}
	return selected
}

// updateRemoteCommandList refreshes the list with current selection state
func (m *Model) updateRemoteCommandList() {
	items := make([]list.Item, len(m.remoteCommands))
	for i, cmd := range m.remoteCommands {
		items[i] = remoteCommandItem{
			command:  cmd,
			selected: m.remoteSelected[i],
			index:    i,
		}
	}
	m.list.SetItems(items)
}

// StartRemoteImportProcess begins the actual import process
func (m *Model) StartRemoteImportProcess() tea.Cmd {
	selectedCommands := m.GetSelectedRemoteCommands()
	if len(selectedCommands) == 0 {
		return nil
	}
	
	m.state = StateRemoteImport
	
	// Return command to start async import
	return func() tea.Msg {
		return RemoteImportMsg{Commands: selectedCommands}
	}
}

// ReturnToMain returns to the main menu state and refreshes the command list
func (m *Model) ReturnToMain() tea.Cmd {
	m.state = StateMainMenu
	m.initMainMenu()
	return func() tea.Msg {
		return RefreshMsg{}
	}
}

// Repository browsing methods

// updateBrowseList updates the list based on current browse mode
func (m *Model) updateBrowseList() {
	if m.registryManager == nil || !m.registryManager.IsLoaded() {
		return
	}

	switch m.browseMode {
	case BrowseModeCategories:
		m.updateCategoryList()
	case BrowseModeRepositories:
		m.updateRepositoryList()
	case BrowseModeSearch:
		m.updateSearchResults()
	}
}

// updateCategoryList populates the list with categories
func (m *Model) updateCategoryList() {
	categories := m.registryManager.GetCategories()
	items := make([]list.Item, 0, len(categories))
	
	// Create a sorted list of category keys to ensure consistent ordering
	sortedKeys := []string{"development", "project_management", "performance", "testing", "security", "general"}
	
	for _, key := range sortedKeys {
		if category, exists := categories[key]; exists {
			items = append(items, categoryItem{
				key:      key,
				category: category,
			})
		}
	}
	
	// Add any categories that weren't in our predefined order
	for key, category := range categories {
		found := false
		for _, sortedKey := range sortedKeys {
			if key == sortedKey {
				found = true
				break
			}
		}
		if !found {
			items = append(items, categoryItem{
				key:      key,
				category: category,
			})
		}
	}
	
	m.list.SetItems(items)
}

// updateRepositoryList populates the list with repositories from current category
func (m *Model) updateRepositoryList() {
	var repositories []remote.CuratedRepository
	
	if m.currentCategory != "" {
		repositories = m.registryManager.GetCategoryRepositories(m.currentCategory)
	} else {
		repositories = m.registryManager.GetAllRepositories()
	}
	
	items := make([]list.Item, len(repositories))
	for i, repo := range repositories {
		items[i] = repositoryItem{
			repository: repo,
			selected:   m.browseSelected[i],
			index:      i,
		}
	}
	
	m.list.SetItems(items)
	m.filteredRepos = repositories
}

// updateSearchResults populates the list with search results
func (m *Model) updateSearchResults() {
	results := m.registryManager.SearchRepositories(m.searchQuery)
	items := make([]list.Item, len(results))
	
	for i, repo := range results {
		items[i] = repositoryItem{
			repository: repo,
			selected:   m.browseSelected[i],
			index:      i,
		}
	}
	
	m.list.SetItems(items)
	m.filteredRepos = results
}

// enterCategory enters a specific category
func (m *Model) enterCategory() {
	index := m.list.Index()
	if index < 0 || index >= len(m.list.Items()) {
		return
	}
	
	if item, ok := m.list.Items()[index].(categoryItem); ok {
		m.browseMode = BrowseModeRepositories
		m.currentCategory = item.key
		m.updateBrowseList()
	}
}

// toggleRepositorySelection toggles selection of a repository
func (m *Model) toggleRepositorySelection() {
	if m.browseMode != BrowseModeRepositories && m.browseMode != BrowseModeSearch {
		return
	}
	
	index := m.list.Index()
	if index < 0 || index >= len(m.filteredRepos) {
		return
	}
	
	m.browseSelected[index] = !m.browseSelected[index]
	m.updateBrowseList()
}

// selectAllRepositories selects or deselects all visible repositories
func (m *Model) selectAllRepositories(selectAll bool) {
	if m.browseMode != BrowseModeRepositories && m.browseMode != BrowseModeSearch {
		return
	}
	
	for i := range m.filteredRepos {
		m.browseSelected[i] = selectAll
	}
	
	m.updateBrowseList()
}

// getSelectedRepositories returns the currently selected repositories
func (m *Model) getSelectedRepositories() []remote.CuratedRepository {
	var selected []remote.CuratedRepository
	
	for i, repo := range m.filteredRepos {
		if m.browseSelected[i] {
			selected = append(selected, repo)
		}
	}
	
	return selected
}

// startSearch initiates search mode
func (m *Model) startSearch() {
	m.browseMode = BrowseModeSearch
	m.searchInput.SetValue("")
	m.searchInput.Focus()
}

// performSearch updates search results based on current query
func (m *Model) performSearch() {
	m.searchQuery = strings.TrimSpace(m.searchInput.Value())
	m.updateSearchResults()
}

// exitSearch exits search mode and returns to category browsing
func (m *Model) exitSearch() {
	m.browseMode = BrowseModeCategories
	m.searchQuery = ""
	m.searchInput.Blur()
	m.updateBrowseList()
}

// goToCustomURL switches to custom URL entry mode
func (m *Model) goToCustomURL() {
	m.state = StateRemoteURL
	m.textInput.SetValue("")
	m.textInput.Placeholder = "Enter GitHub repository URL..."
	m.textInput.Focus()
	
	// Reset custom repository input
	m.customRepoInput = registry.RepositoryInput{}
	m.remoteError = ""
}

// startCustomRepoFlow starts the enhanced custom repository flow
func (m *Model) startCustomRepoFlow(repoURL string) {
	// Parse the GitHub URL to get basic info
	repo, err := remote.ParseGitHubURL(repoURL)
	if err != nil {
		m.remoteError = err.Error()
		return
	}
	
	// Initialize custom repo input with parsed data
	m.customRepoInput = registry.RepositoryInput{
		URL:    repoURL,
		Name:   fmt.Sprintf("%s/%s", repo.Owner, repo.Repo),
		Author: repo.Owner,
		Tags:   []string{},
	}
	
	// Load available categories
	m.availableCategories = m.registryManager.GetAvailableCategories() 
	
	// Move to repository details input
	m.state = StateRemoteRepoDetails
	m.setupRepoDetailsInput()
}

// setupRepoDetailsInput sets up the text input for repository details
func (m *Model) setupRepoDetailsInput() {
	m.textInput.SetValue(m.customRepoInput.Description)
	m.textInput.Placeholder = "Enter repository description..."
	m.textInput.Focus()
}

// startCategorySelection starts the category selection flow
func (m *Model) startCategorySelection() {
	m.state = StateRemoteCategory
	m.setupCategorySelection()
}

// setupCategorySelection sets up the category selection list
func (m *Model) setupCategorySelection() {
	items := make([]list.Item, 0, len(m.availableCategories)+1)
	
	// Add existing categories
	for key, name := range m.availableCategories {
		items = append(items, categorySelectionItem{
			key:  key,
			name: name,
			isNew: false,
		})
	}
	
	// Add "Create New Category" option
	items = append(items, categorySelectionItem{
		key:  "new",
		name: "Create New Category...",
		isNew: true,
	})
	
	m.list.SetItems(items)
	m.list.Select(0)
}

// confirmCategorySelection confirms the category selection
func (m *Model) confirmCategorySelection() {
	selectedIndex := m.list.Index() 
	items := m.list.Items()
	
	if selectedIndex < 0 || selectedIndex >= len(items) {
		return
	}
	
	if item, ok := items[selectedIndex].(categorySelectionItem); ok {
		m.selectedCategoryKey = item.key
		m.isNewCategory = item.isNew
		
		if item.isNew {
			// Start new category creation
			m.startNewCategoryCreation()
		} else {
			// Use existing category
			m.customRepoInput.Category = registry.CategoryInput{
				CategoryKey: item.key,
				IsNew:      false,
			}
			m.finalizeCustomRepository()
		}
	}
}

// startNewCategoryCreation starts the new category creation flow
func (m *Model) startNewCategoryCreation() {
	m.categoryInput.SetValue("")
	m.categoryInput.Focus()
	// Stay in StateRemoteCategory but change the UI context
}

// finalizeCustomRepository completes the custom repository addition
func (m *Model) finalizeCustomRepository() {
	// Update description from text input
	m.customRepoInput.Description = strings.TrimSpace(m.textInput.Value())
	
	// Add the repository to the user registry
	if err := m.registryManager.AddCustomRepository(m.customRepoInput); err != nil {
		m.remoteError = fmt.Sprintf("Failed to add repository: %v", err)
		m.state = StateRemoteURL
		return
	}
	
	// Start the download process
	m.remoteURL = m.customRepoInput.URL
	repo, _ := remote.ParseGitHubURL(m.customRepoInput.URL)
	m.remoteRepo = repo
	m.state = StateRemoteLoading
	m.remoteLoading = true
}

// importSelectedRepositories starts the import process for selected repositories
func (m *Model) importSelectedRepositories() tea.Cmd {
	selected := m.getSelectedRepositories()
	
	// If no repositories are selected via checkboxes, import the currently focused repository
	if len(selected) == 0 {
		// Get the currently focused repository
		index := m.list.Index()
		if index < 0 || index >= len(m.filteredRepos) {
			return nil
		}
		
		// Import the focused repository directly
		focusedRepo := m.filteredRepos[index]
		return m.importSingleRepository(focusedRepo)
	}
	
	// For now, import the first selected repository
	// TODO: Support batch import of multiple repositories
	firstRepo := selected[0]
	
	// Parse the repository URL and start import
	repo, err := remote.ParseGitHubURL(firstRepo.URL)
	if err != nil {
		m.remoteError = err.Error()
		return nil
	}
	
	m.remoteURL = firstRepo.URL
	m.remoteRepo = repo
	m.remoteError = ""
	m.state = StateRemoteLoading
	m.remoteLoading = true
	
	// Return command to start async loading
	return func() tea.Msg {
		return RemoteLoadingMsg{}
	}
}

// importSingleRepository imports a single repository directly
func (m *Model) importSingleRepository(repository remote.CuratedRepository) tea.Cmd {
	// Parse the repository URL and start import
	repo, err := remote.ParseGitHubURL(repository.URL)
	if err != nil {
		// Show error to user instead of silent failure
		m.remoteError = fmt.Sprintf("Cannot load repository '%s': %s", repository.Name, err.Error())
		m.state = StateRemoteURL // Show error in URL input state
		return nil
	}
	
	m.remoteURL = repository.URL
	m.remoteRepo = repo
	m.remoteError = ""
	m.state = StateRemoteLoading
	m.remoteLoading = true
	
	// Return command to start async loading
	return func() tea.Msg {
		return RemoteLoadingMsg{}
	}
}

// calculateAvailableHeight dynamically calculates available height for lists based on current state
func (m *Model) calculateAvailableHeight() int {
	baseReserved := 4 // minimum space for footers and spacing
	
	switch m.state {
	case StateMainMenu:
		// Account for ASCII header and card styling overhead
		headerHeight := 15 // Full ASCII art header
		if m.width < 80 { // Updated threshold to match view.go
			headerHeight = 8 // Simple header is smaller
		}
		
		// Each card item takes approximately 2 lines (borders + content)
		// With 3 menu items, we need about 6 lines for content + some spacing
		cardOverhead := 8
		
		availableHeight := m.height - headerHeight - baseReserved - cardOverhead
		if availableHeight < 3 {
			availableHeight = 3 // Ensure minimum viable height
		}
		return availableHeight
		
	case StateLibrary, StateRemoteBrowse, StateRemoteSelect:
		return m.height - 6 - baseReserved // Header + footer space
		
	case StateRename, StateRemoteURL, StateRemoteRepoDetails, StateRemoteCategory:  
		return m.height - 10 - baseReserved // More space for input forms
		
	case StateHelp:
		return m.height - 4 - baseReserved // Minimal header for help
		
	default:
		return m.height - 8 - baseReserved // Default conservative estimate
	}
}

// validateReportIssueInput validates the report issue form inputs
func (m *Model) validateReportIssueInput() bool {
	m.validationErrors = make(map[string]string) // Clear previous errors
	isValid := true
	
	// Validate title
	title := strings.TrimSpace(m.issueTitleInput.Value())
	if title == "" {
		m.validationErrors["title"] = "Issue title is required"
		isValid = false
	} else if len(title) < 5 {
		m.validationErrors["title"] = "Title must be at least 5 characters"
		isValid = false
	} else if len(title) > 100 {
		m.validationErrors["title"] = "Title too long (max 100 characters)"
		isValid = false
	}
	
	// Validate body (optional but recommended)
	body := strings.TrimSpace(m.issueBodyInput.Value())
	if len(body) > 2000 {
		m.validationErrors["body"] = "Description too long (max 2000 characters)"
		isValid = false
	}
	
	return isValid
}

// SubmitIssue submits the issue to GitHub
func (m *Model) SubmitIssue() tea.Cmd {
	title := strings.TrimSpace(m.issueTitleInput.Value())
	body := strings.TrimSpace(m.issueBodyInput.Value())
	
	// Set submitting state
	m.issueSubmitting = true
	m.issueSubmitError = ""
	
	// Return async command to submit issue
	return func() tea.Msg {
		// Get repository information
		repoInfo, err := remote.GetRepositoryInfo()
		if err != nil {
			return IssueSubmissionCompleteMsg{
				Success: false,
				Error:   fmt.Sprintf("Failed to get repository info: %v", err),
			}
		}
		
		// Create the issue
		err = remote.CreateGitHubIssue(repoInfo, title, body)
		if err != nil {
			return IssueSubmissionCompleteMsg{
				Success: false,
				Error:   fmt.Sprintf("Failed to create issue: %v", err),
			}
		}
		
		return IssueSubmissionCompleteMsg{
			Success: true,
			IssueURL: fmt.Sprintf("https://github.com/%s/%s/issues", repoInfo.Owner, repoInfo.Repo),
		}
	}
}

// validateInput validates input fields and sets validation errors
func (m *Model) validateInput() bool {
	m.validationErrors = make(map[string]string) // Clear previous errors
	isValid := true
	
	switch m.state {
	case StateRename:
		newName := strings.TrimSpace(m.textInput.Value())
		if newName == "" {
			m.validationErrors["name"] = "Name cannot be empty"
			isValid = false
		} else if len(newName) > 100 {
			m.validationErrors["name"] = "Name too long (max 100 characters)"
			isValid = false
		}
		
	case StateRemoteURL:
		url := strings.TrimSpace(m.textInput.Value())
		if url == "" {
			m.validationErrors["url"] = "URL cannot be empty"
			isValid = false
		} else if !strings.Contains(url, "github.") {
			m.validationErrors["url"] = "Only GitHub URLs are supported"
			isValid = false
		}
		
	case StateRemoteRepoDetails:
		description := strings.TrimSpace(m.textInput.Value())
		if description == "" {
			m.validationErrors["description"] = "Description cannot be empty"
			isValid = false
		} else if len(description) > 500 {
			m.validationErrors["description"] = "Description too long (max 500 characters)"
			isValid = false
		}
		
	case StateRemoteCategory:
		if m.isNewCategory && m.selectedCategoryKey == "new" {
			categoryName := strings.TrimSpace(m.categoryInput.Value())
			if categoryName == "" {
				m.validationErrors["category"] = "Category name cannot be empty"
				isValid = false
			} else if len(categoryName) > 50 {
				m.validationErrors["category"] = "Category name too long (max 50 characters)"
				isValid = false
			}
		}
	}
	
	return isValid
}

// clearValidationErrors clears all validation errors
func (m *Model) clearValidationErrors() {
	m.validationErrors = make(map[string]string)
}

// setStatus sets a status message with the given type
func (m *Model) setStatus(message string, statusType StatusType) {
	m.statusMessage = message
	m.statusType = statusType
	m.showStatus = true
}

// clearStatus clears the current status message
func (m *Model) clearStatus() {
	m.statusMessage = ""
	m.showStatus = false
}

// getStatusStyle returns the appropriate style for the current status type
func (m *Model) getStatusStyle() lipgloss.Style {
	switch m.statusType {
	case StatusSuccess:
		return successStyle
	case StatusError:
		return dangerStyle
	case StatusWarning:
		return warningStyle
	default:
		return highlightStyle
	}
}

// StartPreview enters preview mode for the selected command
func (m *Model) StartPreview() {
	if m.state != StateRemoteSelect {
		return
	}
	
	index := m.list.Index()
	if index < 0 || index >= len(m.remoteCommands) {
		return
	}
	
	// Store the command to preview and the previous state
	m.previewCommand = &m.remoteCommands[index]
	m.previousState = m.state
	m.state = StateRemotePreview
}

// ExitPreview returns to the previous state from preview mode
func (m *Model) ExitPreview() {
	if m.state != StateRemotePreview {
		return
	}
	
	// Return to the previous state
	m.state = m.previousState
	m.previewCommand = nil
}

// SetRemoteCommands sets the remote commands for testing purposes
func (m *Model) SetRemoteCommands(commands []remote.RemoteCommand) {
	m.remoteCommands = commands
	m.remoteSelected = make(map[int]bool)
	m.state = StateRemoteSelect
	m.updateRemoteCommandList()
}
