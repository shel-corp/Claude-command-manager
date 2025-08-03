package tui

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

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
	StateRemoteRepoDetails  // New: Repository details input
	StateRemoteCategory     // New: Category selection
	StateRemoteLoading
	StateRemoteSelect
	StateRemoteConflict
	StateRemoteImport
	StateRemoteResults
)

// BrowseMode represents the current browsing mode in the repository browser
type BrowseMode int

const (
	BrowseModeCategories BrowseMode = iota
	BrowseModeRepositories
	BrowseModeSearch
	BrowseModeCustomURL
)

// LibraryMode represents which command library is currently being viewed
type LibraryMode int

const (
	LibraryModeProject LibraryMode = iota // Current project's command library
	LibraryModeUser                       // User's home command library
)

// Model represents the application state for Bubble Tea
type Model struct {
	// Core components
	list         list.Model
	textInput    textinput.Model
	
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
	
	// Custom repository input state
	customRepoInput   registry.RepositoryInput
	availableCategories map[string]string  // key -> name mapping
	selectedCategoryKey string
	isNewCategory       bool
	
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
	// Initialize text input for rename functionality
	ti := textinput.New()
	ti.Placeholder = "Enter new name..."
	ti.Focus()
	ti.CharLimit = 300
	ti.Width = 60

	// Initialize list with custom delegate to remove default styling
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)
	delegate.SetSpacing(0)
	delegate.ShowDescription = true
	
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	
	// Remove default list styling
	l.SetShowTitle(false)
	l.SetShowPagination(false)

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
		commandManager:     commandManager,
		configManager:      configManager,
		userCommandManager: userCommandManager,
		userConfigManager:  userConfigManager,
		cacheManager:       cacheManager,
		state:              StateMainMenu,
		libraryMode:        LibraryModeProject, // Start with project library
		registryManager:    registryManager,
		browseSelected:     make(map[int]bool),
		availableCategories: make(map[string]string),
		customRepoInput:    registry.RepositoryInput{},
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
	m.textInput.SetValue("")
	m.textInput.Placeholder = "Search repositories..."
	m.textInput.Focus()
}

// performSearch updates search results based on current query
func (m *Model) performSearch() {
	m.searchQuery = strings.TrimSpace(m.textInput.Value())
	m.updateSearchResults()
}

// exitSearch exits search mode and returns to category browsing
func (m *Model) exitSearch() {
	m.browseMode = BrowseModeCategories
	m.searchQuery = ""
	m.textInput.Blur()
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
	m.textInput.SetValue("")
	m.textInput.Placeholder = "Enter new category name..."
	m.textInput.Focus()
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
