package tui

import (
	"strings"
	
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/shel-corp/Claude-command-manager/internal/commands"
	"github.com/shel-corp/Claude-command-manager/internal/config"
	"github.com/shel-corp/Claude-command-manager/internal/remote"
)

// State represents the current application state
type State int

const (
	StateMain State = iota
	StateRename
	StateHelp
	StateRemoteURL
	StateRemoteLoading
	StateRemoteSelect
	StateRemoteConflict
	StateRemoteImport
	StateRemoteResults
)

// Model represents the application state for Bubble Tea
type Model struct {
	// Core components
	list         list.Model
	textInput    textinput.Model
	
	// Managers
	commandManager *commands.Manager
	configManager  *config.Manager
	
	// Application state
	state          State
	commands       []commands.Command
	
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

// NewModel creates a new TUI model
func NewModel(commandManager *commands.Manager, configManager *config.Manager) (*Model, error) {
	// Initialize text input for rename functionality
	ti := textinput.New()
	ti.Placeholder = "Enter new name..."
	ti.Focus()
	ti.CharLimit = 300
	ti.Width = 60

	// Initialize list
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	model := &Model{
		list:           l,
		textInput:      ti,
		commandManager: commandManager,
		configManager:  configManager,
		state:          StateMain,
	}

	// Load commands
	if err := model.RefreshCommands(); err != nil {
		return nil, err
	}

	return model, nil
}

// RefreshCommands reloads commands from disk and updates the list
func (m *Model) RefreshCommands() error {
	cmds, err := m.commandManager.ScanCommands()
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

	var err error

	if cmd.Enabled {
		err = m.commandManager.DisableCommand(*cmd)
	} else {
		err = m.commandManager.EnableCommand(*cmd)
	}

	if err != nil {
		return func() tea.Msg {
			return ErrorMsg{Error: err}
		}
	}

	// Save configuration immediately
	if err := m.configManager.Save(); err != nil {
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
		m.state = StateMain
		return nil
	}

	cmd := &m.commands[m.renameIndex]
	err := m.commandManager.RenameCommand(*cmd, newName)
	if err != nil {
		m.state = StateMain
		return func() tea.Msg {
			return ErrorMsg{Error: err}
		}
	}

	// Save configuration immediately
	if err := m.configManager.Save(); err != nil {
		m.state = StateMain
		return func() tea.Msg {
			return ErrorMsg{Error: err}
		}
	}

	m.state = StateMain

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

	err := m.commandManager.ToggleSymlinkLocation(*cmd)
	if err != nil {
		return func() tea.Msg {
			return ErrorMsg{Error: err}
		}
	}

	// Save configuration immediately
	if err := m.configManager.Save(); err != nil {
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
	return tea.Quit
}

// Remote import methods

// StartRemoteImport initiates the remote import flow
func (m *Model) StartRemoteImport() {
	m.state = StateRemoteURL
	m.textInput.SetValue("")
	m.textInput.Placeholder = "Enter GitHub repository URL..."
	m.textInput.Focus()
	
	// Reset remote state
	m.remoteURL = ""
	m.remoteRepo = nil
	m.remoteCommands = nil
	m.remoteLoading = false
	m.remoteError = ""
	m.remoteSelected = make(map[int]bool)
	m.remoteConflicts = nil
	m.remoteResult = nil
}

// ProcessRemoteURL validates and processes the entered repository URL
func (m *Model) ProcessRemoteURL() tea.Cmd {
	url := strings.TrimSpace(m.textInput.Value())
	if url == "" {
		return nil
	}
	
	// Parse the GitHub URL
	repo, err := remote.ParseGitHubURL(url)
	if err != nil {
		m.remoteError = err.Error()
		return nil
	}
	
	m.remoteURL = url
	m.remoteRepo = repo
	m.remoteError = ""
	m.state = StateRemoteLoading
	m.remoteLoading = true
	
	// Return command to start async loading
	return func() tea.Msg {
		return RemoteLoadingMsg{}
	}
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

// ReturnToMain returns to the main state and refreshes the command list
func (m *Model) ReturnToMain() tea.Cmd {
	m.state = StateMain
	return func() tea.Msg {
		return RefreshMsg{}
	}
}
