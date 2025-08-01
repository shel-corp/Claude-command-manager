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

	"github.com/sheltontolbert/claude_command_manager/internal/commands"
	"github.com/sheltontolbert/claude_command_manager/internal/config"
	"github.com/sheltontolbert/claude_command_manager/internal/tui"
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

	// Initialize managers
	configManager := config.NewManager(configPath)
	if err := configManager.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	commandManager := commands.NewManager(commandsDir, userCommandsDir, projectCommandsDir, configManager)

	// Clean up any broken symlinks
	if err := commandManager.CleanupBrokenSymlinks(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to cleanup broken symlinks: %v\n", err)
	}

	// Create TUI model
	model, err := tui.NewModel(commandManager, configManager)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating TUI model: %v\n", err)
		os.Exit(1)
	}

	// Start the TUI application
	p := tea.NewProgram(*model, tea.WithAltScreen())
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
			fmt.Fprintf(os.Stderr, "Usage: command_library rename <command_name> <new_name>\n")
			os.Exit(1)
		}
		return handleRenameCommand(commandManager, configManager, args[1], args[2])
	case "help", "-h", "--help":
		printUsage()
		return true
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
	fmt.Println("  ccm help                     Show this help message")
	fmt.Println()
	
	// Center the copyright text
	copyrightText := fmt.Sprintf("© %d shelcorp. All rights reserved.", time.Now().Year())
	fmt.Println(centerText(copyrightText))
}
