package tui

import (
	"os"
	"path/filepath"
	
	tea "github.com/charmbracelet/bubbletea"
	
	"github.com/sheltontolbert/claude_command_manager/internal/remote"
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
	case StateMain:
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		
	case StateRename:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		
	case StateRemoteURL:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		
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
	case StateMain:
		return m.handleMainStateKeys(msg)
	case StateRename:
		return m.handleRenameStateKeys(msg)
	case StateHelp:
		return m.handleHelpStateKeys(msg)
	case StateRemoteURL:
		return m.handleRemoteURLStateKeys(msg)
	case StateRemoteSelect:
		return m.handleRemoteSelectStateKeys(msg)
	case StateRemoteResults:
		return m.handleRemoteResultsStateKeys(msg)
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
		
	case "l":
		return *m, m.ToggleSelectedCommandLocation()
		
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

// Remote import message handlers

func (m *Model) handleRemoteLoading() (tea.Model, tea.Cmd) {
	// Start async loading of remote repository data
	return *m, func() tea.Msg {
		client := remote.NewGitHubClient()
		
		// Validate repository
		if err := client.ValidateRepository(m.remoteRepo); err != nil {
			return RemoteLoadedMsg{Error: err.Error()}
		}
		
		// Fetch commands
		if err := client.FetchCommands(m.remoteRepo); err != nil {
			return RemoteLoadedMsg{Error: err.Error()}
		}
		
		// Load command details
		for i := range m.remoteRepo.Commands {
			if err := client.FetchCommandContent(m.remoteRepo, &m.remoteRepo.Commands[i]); err != nil {
				m.remoteRepo.Commands[i].Description = "Failed to load description"
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

func (m *Model) handleRemoteURLStateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return *m, m.ProcessRemoteURL()
		
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
		m.state = StateMain
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