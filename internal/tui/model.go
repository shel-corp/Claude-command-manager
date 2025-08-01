package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sheltontolbert/claude_command_manager/internal/commands"
	"github.com/sheltontolbert/claude_command_manager/internal/config"
)

// State represents the current application state
type State int

const (
	StateMain State = iota
	StateRename
	StateHelp
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

// NewModel creates a new TUI model
func NewModel(commandManager *commands.Manager, configManager *config.Manager) (*Model, error) {
	// Initialize text input for rename functionality
	ti := textinput.New()
	ti.Placeholder = "Enter new name..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 30

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
