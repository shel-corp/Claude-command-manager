package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shel-corp/Claude-command-manager/internal/config"
)

// Command represents a single command with its metadata
type Command struct {
	Name            string                 // Original filename without .md
	DisplayName     string                 // Display name (can be renamed)
	Description     string                 // From YAML frontmatter
	Enabled         bool                   // Whether it's currently enabled
	FilePath        string                 // Full path to the .md file
	SymlinkLocation config.SymlinkLocation // Where the command should be symlinked
}

// Manager handles command operations
type Manager struct {
	commandsDir          string
	userCommandsDir      string // ~/.claude/commands/
	projectCommandsDir   string // <project>/.claude/commands/
	configManager        *config.Manager
}

// NewManager creates a new command manager
func NewManager(commandsDir, userCommandsDir, projectCommandsDir string, configManager *config.Manager) *Manager {
	return &Manager{
		commandsDir:        commandsDir,
		userCommandsDir:    userCommandsDir,
		projectCommandsDir: projectCommandsDir,
		configManager:      configManager,
	}
}

// ScanCommands discovers all .md files in the commands directory
func (m *Manager) ScanCommands() ([]Command, error) {
	if _, err := os.Stat(m.commandsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("commands directory not found: %s", m.commandsDir)
	}

	var commands []Command
	
	err := filepath.Walk(m.commandsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			name := strings.TrimSuffix(info.Name(), ".md")
			
			// Get configuration
			cmdConfig, exists := m.configManager.GetCommand(name)
			displayName := name
			enabled := false
			symlinkLocation := config.SymlinkLocationUser // Default to user
			
			if exists {
				displayName = cmdConfig.DisplayName
				enabled = cmdConfig.Enabled
				symlinkLocation = cmdConfig.SymlinkLocation
				// Handle legacy configs without symlink_location field
				if symlinkLocation == "" {
					symlinkLocation = config.SymlinkLocationUser
				}
			}

			// Parse description from file
			description := m.parseDescription(path)

			commands = append(commands, Command{
				Name:            name,
				DisplayName:     displayName,
				Description:     description,
				Enabled:         enabled,
				FilePath:        path,
				SymlinkLocation: symlinkLocation,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan commands directory: %w", err)
	}

	// Sort commands by name for consistent ordering
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	return commands, nil
}

// EnableCommand enables a command by creating a symlink and updating config
func (m *Manager) EnableCommand(cmd Command) error {
	// Ensure symlink directory exists
	symlinkDir := m.getSymlinkDir(cmd.SymlinkLocation)
	if err := os.MkdirAll(symlinkDir, 0755); err != nil {
		return fmt.Errorf("failed to create symlink directory: %w", err)
	}

	// Create symlink
	if err := m.createSymlink(cmd); err != nil {
		return err
	}

	// Update configuration
	m.configManager.SetCommand(cmd.Name, config.CommandConfig{
		Enabled:         true,
		OriginalName:    cmd.Name,
		DisplayName:     cmd.DisplayName,
		SourcePath:      cmd.FilePath,
		SymlinkLocation: cmd.SymlinkLocation,
	})

	return nil
}

// DisableCommand disables a command by removing symlink and updating config
func (m *Manager) DisableCommand(cmd Command) error {
	// Remove symlink
	if err := m.removeSymlink(cmd); err != nil {
		return err
	}

	// Update configuration
	m.configManager.SetCommand(cmd.Name, config.CommandConfig{
		Enabled:         false,
		OriginalName:    cmd.Name,
		DisplayName:     cmd.DisplayName,
		SourcePath:      cmd.FilePath,
		SymlinkLocation: cmd.SymlinkLocation,
	})

	return nil
}

// RenameCommand renames a command's display name
func (m *Manager) RenameCommand(cmd Command, newDisplayName string) error {
	if cmd.DisplayName == newDisplayName {
		return nil // No change needed
	}

	// If command is enabled, update the symlink
	if cmd.Enabled {
		// Remove old symlink
		if err := m.removeSymlink(cmd); err != nil {
			return fmt.Errorf("failed to remove old symlink: %w", err)
		}

		// Create new symlink with new name
		newCmd := cmd
		newCmd.DisplayName = newDisplayName
		if err := m.createSymlink(newCmd); err != nil {
			// Try to restore old symlink on failure
			m.createSymlink(cmd)
			return fmt.Errorf("failed to create new symlink: %w", err)
		}
	}

	// Update configuration
	m.configManager.SetCommand(cmd.Name, config.CommandConfig{
		Enabled:         cmd.Enabled,
		OriginalName:    cmd.Name,
		DisplayName:     newDisplayName,
		SourcePath:      cmd.FilePath,
		SymlinkLocation: cmd.SymlinkLocation,
	})

	return nil
}

// getSymlinkDir returns the appropriate symlink directory based on location
func (m *Manager) getSymlinkDir(location config.SymlinkLocation) string {
	switch location {
	case config.SymlinkLocationProject:
		return m.projectCommandsDir
	default: // config.SymlinkLocationUser or empty
		return m.userCommandsDir
	}
}

// createSymlink creates a symlink for the command
func (m *Manager) createSymlink(cmd Command) error {
	sourcePath, err := filepath.Abs(cmd.FilePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	symlinkDir := m.getSymlinkDir(cmd.SymlinkLocation)
	targetPath := filepath.Join(symlinkDir, cmd.DisplayName+".md")
	
	// Ensure symlink directory exists
	if err := os.MkdirAll(symlinkDir, 0755); err != nil {
		return fmt.Errorf("failed to create symlink directory: %w", err)
	}

	// Check if target already exists
	if info, err := os.Lstat(targetPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			// It's a symlink, check if it points to our file
			link, err := os.Readlink(targetPath)
			if err == nil {
				// Convert link to absolute path for comparison
				linkAbs, err := filepath.Abs(link)
				if err == nil && linkAbs == sourcePath {
					return nil // Already correct
				}
			}
		}
		return fmt.Errorf("target file already exists: %s", targetPath)
	}

	if err := os.Symlink(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

// removeSymlink removes a symlink for the command
func (m *Manager) removeSymlink(cmd Command) error {
	symlinkDir := m.getSymlinkDir(cmd.SymlinkLocation)
	targetPath := filepath.Join(symlinkDir, cmd.DisplayName+".md")

	// Check if it exists and is a symlink
	if info, err := os.Lstat(targetPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(targetPath); err != nil {
				return fmt.Errorf("failed to remove symlink: %w", err)
			}
		} else {
			return fmt.Errorf("target is not a symlink: %s", targetPath)
		}
	}
	// If file doesn't exist, that's fine

	return nil
}

// ToggleSymlinkLocation toggles the symlink location between user and project
func (m *Manager) ToggleSymlinkLocation(cmd Command) error {
	// Determine new location
	var newLocation config.SymlinkLocation
	if cmd.SymlinkLocation == config.SymlinkLocationUser {
		newLocation = config.SymlinkLocationProject
	} else {
		newLocation = config.SymlinkLocationUser
	}

	// If command is enabled, move the symlink
	if cmd.Enabled {
		// Remove old symlink
		if err := m.removeSymlink(cmd); err != nil {
			return fmt.Errorf("failed to remove old symlink: %w", err)
		}

		// Create new symlink with new location
		newCmd := cmd
		newCmd.SymlinkLocation = newLocation
		if err := m.createSymlink(newCmd); err != nil {
			// Try to restore old symlink on failure
			m.createSymlink(cmd)
			return fmt.Errorf("failed to create new symlink: %w", err)
		}
	}

	// Update configuration
	m.configManager.SetCommand(cmd.Name, config.CommandConfig{
		Enabled:         cmd.Enabled,
		OriginalName:    cmd.Name,
		DisplayName:     cmd.DisplayName,
		SourcePath:      cmd.FilePath,
		SymlinkLocation: newLocation,
	})

	return nil
}

// parseDescription extracts the description from YAML frontmatter
func (m *Manager) parseDescription(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return "No description available"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	
	// Look for YAML frontmatter
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				// End of frontmatter
				break
			}
		}
		
		if inFrontmatter {
			// Look for description field
			if strings.HasPrefix(line, "description:") {
				// Extract description value
				description := strings.TrimPrefix(line, "description:")
				description = strings.TrimSpace(description)
				// Remove quotes if present
				description = strings.Trim(description, `"'`)
				if description != "" {
					return description
				}
			}
		}
	}

	return "No description available"
}

// CleanupBrokenSymlinks removes any broken symlinks in both user and project command directories
func (m *Manager) CleanupBrokenSymlinks() error {
	var totalRemoved []string
	
	// Clean up both directories
	dirs := []string{m.userCommandsDir, m.projectCommandsDir}
	
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue // Directory doesn't exist, nothing to clean
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to read commands directory %s: %v\n", dir, err)
			continue
		}

		for _, entry := range entries {
			if entry.Type()&os.ModeSymlink != 0 {
				fullPath := filepath.Join(dir, entry.Name())
				
				// Check if symlink target exists
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					// Broken symlink, remove it
					if err := os.Remove(fullPath); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to remove broken symlink %s: %v\n", fullPath, err)
					} else {
						totalRemoved = append(totalRemoved, entry.Name())
					}
				}
			}
		}
	}

	if len(totalRemoved) > 0 {
		fmt.Fprintf(os.Stderr, "Cleaned up %d broken symlinks: %v\n", len(totalRemoved), totalRemoved)
	}

	return nil
}