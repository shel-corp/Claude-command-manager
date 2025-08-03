package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shel-corp/Claude-command-manager/internal/commands"
	"github.com/shel-corp/Claude-command-manager/internal/config"
	"github.com/shel-corp/Claude-command-manager/internal/remote"
	"github.com/shel-corp/Claude-command-manager/internal/tui"
)

func main() {
	// Get paths by traversing up to find .claude directory
	commandsDir, configPath, claudeDir, err := config.GetCommandLibraryPaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure you are running this command from within a directory that contains a .claude folder.\n")
		os.Exit(1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not get home directory: %v\n", err)
		os.Exit(1)
	}
	
	userCommandsDir := filepath.Join(homeDir, ".claude", "commands")
	projectCommandsDir := filepath.Join(claudeDir, "commands")

	// Handle CLI arguments for backward compatibility
	if len(os.Args) > 1 {
		if handleCLICommands(os.Args[1:], commandsDir, configPath, userCommandsDir, projectCommandsDir) {
			return
		}
	}

	// Initialize managers for project library
	configManager := config.NewManager(configPath)
	if err := configManager.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	commandManager := commands.NewManager(commandsDir, userCommandsDir, projectCommandsDir, configManager)

	// Initialize managers for user library
	userCommandLibraryDir := filepath.Join(homeDir, ".claude", "command_library")
	userConfigPath := filepath.Join(userCommandLibraryDir, ".config.json")
	
	// For user library, the commands are directly in the command_library directory
	// (this maintains compatibility with existing user setups)
	userCommandsLibraryDir := userCommandLibraryDir
	
	// Ensure user command library directory exists
	if err := os.MkdirAll(userCommandsLibraryDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating user command library: %v\n", err)
		os.Exit(1)
	}
	
	userConfigManager := config.NewManager(userConfigPath)
	if err := userConfigManager.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading user configuration: %v\n", err)
		os.Exit(1)
	}

	userCommandManager := commands.NewManager(userCommandsLibraryDir, userCommandsDir, projectCommandsDir, userConfigManager)

	// Clean up any broken symlinks
	if err := commandManager.CleanupBrokenSymlinks(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to cleanup broken symlinks: %v\n", err)
	}
	if err := userCommandManager.CleanupBrokenSymlinks(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to cleanup broken user symlinks: %v\n", err)
	}

	// Create TUI model
	model, err := tui.NewModel(commandManager, configManager, userCommandManager, userConfigManager)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating TUI model: %v\n", err)
		os.Exit(1)
	}
	
	// Use alt screen to ensure proper screen clearing
	p := tea.NewProgram(model, 
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

// handleCLICommands handles command-line interface commands for backward compatibility
func handleCLICommands(args []string, commandsDir, configPath, userCommandsDir, projectCommandsDir string) bool {
	if len(args) == 0 {
		return false
	}

	// Initialize managers
	configManager := config.NewManager(configPath)
	if err := configManager.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	commandManager := commands.NewManager(commandsDir, userCommandsDir, projectCommandsDir, configManager)

	switch args[0] {
	case "list":
		return handleListCommands(commandManager)
	case "status":
		return handleStatusCommands(commandManager)
	case "enable":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: command_library enable <command_name>\n")
			os.Exit(1)
		}
		return handleEnableCommand(commandManager, configManager, args[1])
	case "disable":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: command_library disable <command_name>\n")
			os.Exit(1)
		}
		return handleDisableCommand(commandManager, configManager, args[1])
	case "rename":
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: ccm rename <command_name> <new_name>\n")
			os.Exit(1)
		}
		return handleRenameCommand(commandManager, configManager, args[1], args[2])
	case "import":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: ccm import <github_url>\n")
			os.Exit(1)
		}
		return handleImportCommand(args[1])
	case "browse":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: ccm browse <github_url>\n")
			os.Exit(1)
		}
		return handleBrowseCommand(args[1])
	case "help", "-h", "--help":
		printUsage()
		return true
	case "debug":
		return handleDebugCommand(commandManager, configManager)
	case "test-header":
		return handleTestHeaderCommand()
	case "simple-tui":
		return handleSimpleTUICommand()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}

	return false
}

func handleListCommands(commandManager *commands.Manager) bool {
	cmds, err := commandManager.ScanCommands()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning commands: %v\n", err)
		os.Exit(1)
	}

	for _, cmd := range cmds {
		status := "[ ]"
		if cmd.Enabled {
			status = "[✓]"
		}
		
		// Add location decorator
		var locationIcon string
		switch cmd.SymlinkLocation {
		case config.SymlinkLocationUser:
			locationIcon = "👤"
		case config.SymlinkLocationProject:
			locationIcon = "📁"
		default:
			locationIcon = "👤"
		}
		
		fmt.Printf("%s %s %s: %s\n", status, locationIcon, cmd.DisplayName, cmd.Description)
	}
	return true
}

func handleStatusCommands(commandManager *commands.Manager) bool {
	cmds, err := commandManager.ScanCommands()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning commands: %v\n", err)
		os.Exit(1)
	}

	enabledCount := 0
	for _, cmd := range cmds {
		// Add location decorator
		var locationIcon string
		switch cmd.SymlinkLocation {
		case config.SymlinkLocationUser:
			locationIcon = "👤"
		case config.SymlinkLocationProject:
			locationIcon = "📁"
		default:
			locationIcon = "👤"
		}
		
		if cmd.Enabled {
			fmt.Printf("✓ %s %s (enabled)\n", locationIcon, cmd.DisplayName)
			enabledCount++
		} else {
			fmt.Printf("○ %s %s (disabled)\n", locationIcon, cmd.DisplayName)
		}
	}

	fmt.Printf("\nSummary: %d/%d commands enabled\n", enabledCount, len(cmds))
	return true
}

func handleEnableCommand(commandManager *commands.Manager, configManager *config.Manager, name string) bool {
	cmds, err := commandManager.ScanCommands()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning commands: %v\n", err)
		os.Exit(1)
	}

	for _, cmd := range cmds {
		if cmd.Name == name {
			if err := commandManager.EnableCommand(cmd); err != nil {
				fmt.Fprintf(os.Stderr, "Error enabling command: %v\n", err)
				os.Exit(1)
			}
			if err := configManager.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Enabled command: %s\n", cmd.DisplayName)
			return true
		}
	}

	fmt.Fprintf(os.Stderr, "Command not found: %s\n", name)
	os.Exit(1)
	return true
}

func handleDisableCommand(commandManager *commands.Manager, configManager *config.Manager, name string) bool {
	cmds, err := commandManager.ScanCommands()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning commands: %v\n", err)
		os.Exit(1)
	}

	for _, cmd := range cmds {
		if cmd.Name == name {
			if err := commandManager.DisableCommand(cmd); err != nil {
				fmt.Fprintf(os.Stderr, "Error disabling command: %v\n", err)
				os.Exit(1)
			}
			if err := configManager.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Disabled command: %s\n", cmd.DisplayName)
			return true
		}
	}

	fmt.Fprintf(os.Stderr, "Command not found: %s\n", name)
	os.Exit(1)
	return true
}

func handleRenameCommand(commandManager *commands.Manager, configManager *config.Manager, name, newName string) bool {
	cmds, err := commandManager.ScanCommands()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning commands: %v\n", err)
		os.Exit(1)
	}

	for _, cmd := range cmds {
		if cmd.Name == name {
			oldDisplayName := cmd.DisplayName
			if err := commandManager.RenameCommand(cmd, newName); err != nil {
				fmt.Fprintf(os.Stderr, "Error renaming command: %v\n", err)
				os.Exit(1)
			}
			if err := configManager.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Renamed command: %s → %s\n", oldDisplayName, newName)
			return true
		}
	}

	fmt.Fprintf(os.Stderr, "Command not found: %s\n", name)
	os.Exit(1)
	return true
}

// centerText centers text in the terminal or returns it as-is if centering fails
func centerText(text string) string {
	// Try to get terminal width using tput command
	if cmd := exec.Command("tput", "cols"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			if width, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil && width > len(text) {
				padding := (width - len(text)) / 2
				return strings.Repeat(" ", padding) + text
			}
		}
	}
	
	// Fallback if we can't get terminal size
	return text
}

func printUsage() {
	// Center the title
	fmt.Println(centerText("Claude Command Manager"))
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ccm                          Launch interactive TUI")
	fmt.Println("  ccm list                     List all available commands")
	fmt.Println("  ccm status                   Show current command status")
	fmt.Println("  ccm enable <command_name>    Enable a specific command")
	fmt.Println("  ccm disable <command_name>   Disable a specific command")
	fmt.Println("  ccm rename <cmd> <new_name>  Rename a command")
	fmt.Println("  ccm import <github_url>      Import commands from GitHub repository")
	fmt.Println("  ccm browse <github_url>      Browse available commands in repository")
	fmt.Println("  ccm help                     Show this help message")
	fmt.Println()
	
	// Center the copyright text
	copyrightText := fmt.Sprintf("© %d shelcorp. All rights reserved.", time.Now().Year())
	fmt.Println(centerText(copyrightText))
}

// handleBrowseCommand lists available commands in a remote repository
func handleBrowseCommand(url string) bool {
	// Parse the GitHub URL
	repo, err := remote.ParseGitHubURL(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Initialize GitHub client
	client := remote.NewGitHubClient()

	// Show loading and validate
	fmt.Printf("🔍 Connecting to %s/%s...", repo.Owner, repo.Repo)
	if err := client.ValidateRepository(repo); err != nil {
		fmt.Printf(" ❌\n")
		fmt.Fprintf(os.Stderr, "Repository not accessible: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(" ✅\n")

	// Fetch commands with loading indicator
	fmt.Printf("📦 Scanning for commands...")
	if err := client.FetchCommands(repo); err != nil {
		fmt.Printf(" ❌\n")
		fmt.Fprintf(os.Stderr, "Failed to fetch commands: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(" ✅\n")

	if len(repo.Commands) == 0 {
		fmt.Println("No commands found in repository.")
		return true
	}

	// Load command details
	fmt.Printf("🔄 Loading command details...")
	for i := range repo.Commands {
		if err := client.FetchCommandContent(repo, &repo.Commands[i]); err != nil {
			repo.Commands[i].Description = "Failed to load description"
		}
	}
	fmt.Printf(" ✅\n")

	// Display commands
	fmt.Printf("\n📋 Available commands in %s/%s:\n\n", repo.Owner, repo.Repo)
	for i, cmd := range repo.Commands {
		fmt.Printf("  %2d. %-20s %s\n", i+1, cmd.Name, 
			truncateDescription(cmd.Description, 60))
	}

	fmt.Printf("\n💡 To import commands: ccm import %s\n", url)
	return true
}

// handleImportCommand provides interactive import from a remote repository
func handleImportCommand(url string) bool {
	// Parse the GitHub URL
	repo, err := remote.ParseGitHubURL(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Initialize GitHub client
	client := remote.NewGitHubClient()

	// Show loading and validate
	fmt.Printf("🔍 Connecting to %s/%s...", repo.Owner, repo.Repo)
	if err := client.ValidateRepository(repo); err != nil {
		fmt.Printf(" ❌\n")
		fmt.Fprintf(os.Stderr, "Repository not accessible: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(" ✅\n")

	// Fetch commands with loading indicator
	fmt.Printf("📦 Scanning for commands...")
	if err := client.FetchCommands(repo); err != nil {
		fmt.Printf(" ❌\n")
		fmt.Fprintf(os.Stderr, "Failed to fetch commands: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(" ✅\n")

	if len(repo.Commands) == 0 {
		fmt.Println("No commands found in repository.")
		return true
	}

	// Get target directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not get home directory: %v\n", err)
		os.Exit(1)
	}
	targetDir := filepath.Join(homeDir, ".claude", "command_library")

	// Load command contents and check local conflicts
	fmt.Printf("🔄 Loading command details...")
	importer := remote.NewImporter(targetDir)
	
	for i := range repo.Commands {
		if err := client.FetchCommandContent(repo, &repo.Commands[i]); err != nil {
			// Skip commands that fail to load
			repo.Commands = append(repo.Commands[:i], repo.Commands[i+1:]...)
			i--
			continue
		}
	}
	
	if err := importer.CheckLocalExists(repo.Commands, targetDir); err != nil {
		fmt.Printf(" ❌\n")
		fmt.Fprintf(os.Stderr, "Error checking local commands: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(" ✅\n")

	// Display commands for selection
	fmt.Printf("\n📋 Found %d commands:\n\n", len(repo.Commands))
	
	for i, cmd := range repo.Commands {
		status := "NEW"
		statusIcon := "🆕"
		if cmd.LocalExists {
			status = "EXISTS"
			statusIcon = "⚠️"
		}
		
		fmt.Printf("  %2d. %-20s %s %s %s\n", 
			i+1, cmd.Name, statusIcon, status, 
			truncateDescription(cmd.Description, 50))
	}

	// Interactive selection
	fmt.Print("\n🎯 Select commands to import:\n")
	fmt.Print("   • Enter numbers (e.g., 1,3,5-8) or 'all' for all commands\n")
	fmt.Print("   • Commands marked ⚠️ already exist locally\n")
	fmt.Print("\nSelection: ")
	
	var input string
	fmt.Scanln(&input)
	
	if input == "" {
		fmt.Println("No commands selected.")
		return true
	}

	// Parse selection
	selectedIndices, err := parseSelection(input, len(repo.Commands))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid selection: %v\n", err)
		os.Exit(1)
	}

	// Mark selected commands
	for i := range repo.Commands {
		repo.Commands[i].Selected = false
	}
	for _, idx := range selectedIndices {
		repo.Commands[idx].Selected = true
	}

	// Check for conflicts and ask about overwriting
	hasConflicts := false
	for _, idx := range selectedIndices {
		if repo.Commands[idx].LocalExists {
			hasConflicts = true
			break
		}
	}

	options := remote.GetDefaultImportOptions(targetDir)
	if hasConflicts {
		fmt.Print("\n⚠️  Some selected commands already exist. Overwrite them? (y/N): ")
		var response string
		fmt.Scanln(&response)
		options.OverwriteExisting = strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
	}

	// Import selected commands
	fmt.Printf("\n📥 Importing %d commands...", len(selectedIndices))
	result, err := importer.ImportCommands(repo, repo.Commands, options)
	if err != nil {
		fmt.Printf(" ❌\n")
		fmt.Fprintf(os.Stderr, "Import failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(" ✅\n")

	// Show results
	fmt.Printf("\n🎉 Import Summary:\n")
	fmt.Printf("   ✅ Imported: %d\n", len(result.Imported))
	fmt.Printf("   ⏭️  Skipped:  %d\n", len(result.Skipped))
	fmt.Printf("   ❌ Failed:   %d\n", len(result.Failed))

	if len(result.Failed) > 0 {
		fmt.Printf("\n❌ Failed imports:\n")
		for i, name := range result.Failed {
			fmt.Printf("   • %s: %s\n", name, result.Errors[i])
		}
	}

	if len(result.Imported) > 0 {
		fmt.Printf("\n📁 Commands saved to: %s\n", targetDir)
	}

	return true
}

// truncateDescription truncates a description to fit display width
func truncateDescription(desc string, maxLen int) string {
	if len(desc) <= maxLen {
		return desc
	}
	return desc[:maxLen-3] + "..."
}

// parseSelection parses user input like "1,3,5-8" or "all" 
func parseSelection(input string, maxCount int) ([]int, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	
	if input == "all" {
		indices := make([]int, maxCount)
		for i := range indices {
			indices[i] = i
		}
		return indices, nil
	}

	var indices []int
	parts := strings.Split(input, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		
		if strings.Contains(part, "-") {
			// Handle ranges like "5-8"
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}
			
			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", rangeParts[0])
			}
			
			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", rangeParts[1])
			}
			
			if start < 1 || end < 1 || start > maxCount || end > maxCount {
				return nil, fmt.Errorf("numbers must be between 1 and %d", maxCount)
			}
			
			if start > end {
				start, end = end, start
			}
			
			for i := start; i <= end; i++ {
				indices = append(indices, i-1) // Convert to 0-based
			}
		} else {
			// Handle single numbers
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", part)
			}
			
			if num < 1 || num > maxCount {
				return nil, fmt.Errorf("number must be between 1 and %d", maxCount)
			}
			
			indices = append(indices, num-1) // Convert to 0-based
		}
	}

	// Remove duplicates
	seen := make(map[int]bool)
	uniqueIndices := []int{}
	for _, idx := range indices {
		if !seen[idx] {
			seen[idx] = true
			uniqueIndices = append(uniqueIndices, idx)
		}
	}

	return uniqueIndices, nil
}

// handleDebugCommand shows debug information and tests header rendering
func handleDebugCommand(commandManager *commands.Manager, configManager *config.Manager) bool {
	fmt.Println("=== Claude Command Manager Debug Information ===")
	fmt.Println()
	
	// Test ASCII header rendering
	asciiHeader := `
 ╭─────────────────────────────────────────────────────────────╮
 │                                                             │
 │   ██████╗██╗      █████╗ ██╗   ██╗██████╗ ███████╗         │
 │  ██╔════╝██║     ██╔══██╗██║   ██║██╔══██╗██╔════╝         │
 │  ██║     ██║     ███████║██║   ██║██║  ██║█████╗           │
 │  ██║     ██║     ██╔══██║██║   ██║██║  ██║██╔══╝           │
 │  ╚██████╗███████╗██║  ██║╚██████╔╝██████╔╝███████╗         │
 │   ╚═════╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝         │
 │                                                             │
 │           Command Manager • Interactive TUI                 │
 │                                                             │
 ╰─────────────────────────────────────────────────────────────╯`
 
	fmt.Println("ASCII Header Test:")
	fmt.Println(asciiHeader)
	fmt.Println()
	
	// Get terminal width
	cmd := exec.Command("tput", "cols")
	if output, err := cmd.Output(); err == nil {
		if width, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
			fmt.Printf("Terminal width: %d columns\n", width)
			if width < 70 {
				fmt.Println("⚠️  Terminal is narrow - ASCII header may not display properly")
				fmt.Println("    Recommended: 70+ columns for full header display")
			} else {
				fmt.Println("✅ Terminal width is sufficient for ASCII header")
			}
		}
	} else {
		fmt.Println("Could not detect terminal width")
	}
	
	// Check command loading
	cmds, err := commandManager.ScanCommands()
	if err != nil {
		fmt.Printf("❌ Error loading commands: %v\n", err)
	} else {
		fmt.Printf("✅ Commands loaded: %d found\n", len(cmds))
	}
	
	fmt.Println()
	fmt.Println("To test TUI, run: ccm")
	fmt.Println("If TUI doesn't show header, it may be a terminal compatibility issue.")
	
	return true
}

// handleTestHeaderCommand tests header display without TUI framework
func handleTestHeaderCommand() bool {
	fmt.Println("=== Direct Header Test (No TUI Framework) ===")
	fmt.Println()
	
	// Clear screen
	fmt.Print("\033[2J\033[H")
	
	// Display header exactly as it would appear in TUI with margins
	asciiHeader := `


    ╭─────────────────────────────────────────────────────────────╮
    │                                                             │
    │   ██████╗██╗      █████╗ ██╗   ██╗██████╗ ███████╗         │
    │  ██╔════╝██║     ██╔══██╗██║   ██║██╔══██╗██╔════╝         │
    │  ██║     ██║     ███████║██║   ██║██║  ██║█████╗           │
    │  ██║     ██║     ██╔══██║██║   ██║██║  ██║██╔══╝           │
    │  ╚██████╗███████╗██║  ██║╚██████╔╝██████╔╝███████╗         │
    │   ╚═════╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝         │
    │                                                             │
    │           Command Manager • Interactive TUI                 │
    │                                                             │
    ╰─────────────────────────────────────────────────────────────╯`
 
	fmt.Println(asciiHeader)
	fmt.Println()
	fmt.Println("Manage your Claude AI command library with ease")
	fmt.Println()
	fmt.Println("📚 Library")
	fmt.Println("   Manage your command library")
	fmt.Println()
	fmt.Println("📦 Import") 
	fmt.Println("   Browse and import repository commands")
	fmt.Println()
	fmt.Println("↑/↓: Navigate • Enter: Select • q: Quit • h: Help")
	fmt.Println()
	fmt.Println("This is how the header SHOULD look in the TUI.")
	fmt.Println("Press any key to continue...")
	
	// Wait for user input
	var input string
	fmt.Scanln(&input)
	
	return true
}

// Simple model for testing
type simpleModel struct {
	content string
}

func (m simpleModel) Init() tea.Cmd {
	return nil
}

func (m simpleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m simpleModel) View() string {
	header := `


    ╭─────────────────────────────────────────────────────────────╮
    │                                                             │
    │   ██████╗██╗      █████╗ ██╗   ██╗██████╗ ███████╗         │
    │  ██╔════╝██║     ██╔══██╗██║   ██║██╔══██╗██╔════╝         │
    │  ██║     ██║     ███████║██║   ██║██║  ██║█████╗           │
    │  ██║     ██║     ██╔══██║██║   ██║██║  ██║██╔══╝           │
    │  ╚██████╗███████╗██║  ██║╚██████╔╝██████╔╝███████╗         │
    │   ╚═════╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝         │
    │                                                             │
    │           Command Manager • Interactive TUI                 │
    │                                                             │
    ╰─────────────────────────────────────────────────────────────╯`
	
	return header + "\n\nSimple TUI Test - ASCII Header Above\n\nPress 'q' to quit"
}

// handleSimpleTUICommand creates a minimal TUI test
func handleSimpleTUICommand() bool {
	fmt.Println("=== Simple TUI Test ===")
	fmt.Println("Creating minimal Bubble Tea program...")
	
	model := simpleModel{content: "test"}
	p := tea.NewProgram(model)
	
	fmt.Println("Starting simple TUI... (press 'q' to quit)")
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}
	
	return true
}
