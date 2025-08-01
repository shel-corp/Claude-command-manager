package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindClaudeDirectory traverses up the directory tree to find the nearest .claude directory
func FindClaudeDirectory() (string, error) {
	// Start from current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	return findClaudeDirectoryFrom(currentDir)
}

// findClaudeDirectoryFrom searches for .claude directory starting from the given path
func findClaudeDirectoryFrom(startDir string) (string, error) {
	currentDir := startDir
	
	for {
		// Check if .claude directory exists in current directory
		claudeDir := filepath.Join(currentDir, ".claude")
		if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
			return claudeDir, nil
		}
		
		// Get parent directory
		parentDir := filepath.Dir(currentDir)
		
		// If we've reached the root directory, stop searching
		if parentDir == currentDir {
			break
		}
		
		currentDir = parentDir
	}
	
	return "", fmt.Errorf("no .claude directory found in current directory or any parent directories")
}

// GetCommandLibraryPaths returns all the paths needed for the command library
// based on the discovered .claude directory
func GetCommandLibraryPaths() (commandsDir, configPath, claudeDir string, err error) {
	claudeDir, err = FindClaudeDirectory()
	if err != nil {
		return "", "", "", err
	}
	
	commandLibraryDir := filepath.Join(claudeDir, "command_library")
	
	// Ensure command_library directory exists
	if err := os.MkdirAll(commandLibraryDir, 0755); err != nil {
		return "", "", "", fmt.Errorf("failed to create command_library directory: %w", err)
	}
	
	commandsDir = filepath.Join(commandLibraryDir, "commands")
	configPath = filepath.Join(commandLibraryDir, ".config.json")
	
	// Ensure commands directory exists
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return "", "", "", fmt.Errorf("failed to create commands directory: %w", err)
	}
	
	return commandsDir, configPath, claudeDir, nil
}