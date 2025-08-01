package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Message types for Bubble Tea
type (
	// RefreshMsg signals that the command list should be refreshed
	RefreshMsg struct{}
	
	// ErrorMsg carries error information
	ErrorMsg struct {
		Error error
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
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4) // Leave space for footer
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

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	// Handle updates based on current state
	switch m.state {
	case StateMain:
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		
	case StateRename:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyMsg handles keyboard input based on current state
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMain:
		return m.handleMainStateKeys(msg)
	case StateRename:
		return m.handleRenameStateKeys(msg)
	case StateHelp:
		return m.handleHelpStateKeys(msg)
	}
	
	return *m, nil
}

// handleMainStateKeys handles keys in the main state
func (m *Model) handleMainStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return *m, m.Quit()
		
	case "enter", "t":
		return *m, m.ToggleSelectedCommand()
		
	case "r":
		m.StartRename()
		return *m, nil
		
	case "d":
		return *m, m.DisableSelectedCommand()
		
	case "l":
		return *m, m.ToggleSelectedCommandLocation()
		
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
		m.state = StateMain
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
		m.state = StateMain
		return *m, nil
		
	case "ctrl+c":
		return *m, m.Quit()
	}
	
	return *m, nil
}

// Note: Confirm quit state removed since changes are saved immediately