package remote

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Importer handles importing remote commands to local storage
type Importer struct {
	client          *GitHubClient
	targetDir       string
	shouldBackup    bool
}

// NewImporter creates a new command importer
func NewImporter(targetDir string) *Importer {
	return &Importer{
		client:          NewGitHubClient(),
		targetDir:       targetDir,
		shouldBackup:    true,
	}
}

// ImportCommands imports selected commands from a remote repository
func (i *Importer) ImportCommands(repo *RemoteRepository, selectedCommands []RemoteCommand, options ImportOptions) (*ImportResult, error) {
	result := &ImportResult{
		Imported: make([]string, 0),
		Skipped:  make([]string, 0),
		Failed:   make([]string, 0),
		Errors:   make([]string, 0),
	}

	// Ensure target directory exists
	if err := os.MkdirAll(options.TargetDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Process each selected command
	for _, command := range selectedCommands {
		if !command.Selected {
			continue
		}

		// Fetch command content if not already loaded
		if command.Content == "" {
			if err := i.client.FetchCommandContent(repo, &command); err != nil {
				result.Failed = append(result.Failed, command.Name)
				result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", command.Name, err.Error()))
				continue
			}
		}

		// Import the command
		if err := i.importSingleCommand(command, options, result); err != nil {
			result.Failed = append(result.Failed, command.Name)
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", command.Name, err.Error()))
		}
	}

	return result, nil
}

// importSingleCommand imports a single command with conflict resolution
func (i *Importer) importSingleCommand(command RemoteCommand, options ImportOptions, result *ImportResult) error {
	// Sanitize filename
	safeFilename := sanitizeFilename(command.Name) + ".md"
	targetPath := filepath.Join(options.TargetDirectory, safeFilename)

	// Check if file already exists
	if _, err := os.Stat(targetPath); err == nil {
		if !options.OverwriteExisting {
			result.Skipped = append(result.Skipped, command.Name)
			return nil
		}

		// Create backup if requested
		if options.CreateBackups {
			if err := i.createBackup(targetPath); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}
		}
	}

	// Validate content if requested
	if options.ValidateContent {
		if err := i.validateCommandContent(command.Content); err != nil {
			return fmt.Errorf("content validation failed: %w", err)
		}
	}

	// Write the command file
	if err := os.WriteFile(targetPath, []byte(command.Content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	result.Imported = append(result.Imported, command.Name)
	return nil
}

// createBackup creates a backup of an existing file
func (i *Importer) createBackup(filePath string) error {
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.backup_%s", filePath, timestamp)
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read original file: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// validateCommandContent performs basic validation on command content
func (i *Importer) validateCommandContent(content string) error {
	// Check for minimum content length
	if len(strings.TrimSpace(content)) < 10 {
		return fmt.Errorf("content too short (minimum 10 characters)")
	}

	// Check for potential security issues
	if err := i.checkForSuspiciousContent(content); err != nil {
		return err
	}

	// Validate YAML frontmatter format if present
	if strings.HasPrefix(strings.TrimSpace(content), "---") {
		if err := i.validateYAMLFrontmatter(content); err != nil {
			return fmt.Errorf("invalid YAML frontmatter: %w", err)
		}
	}

	return nil
}

// checkForSuspiciousContent scans for potentially malicious patterns
func (i *Importer) checkForSuspiciousContent(content string) error {
	// List of suspicious patterns to check for
	suspiciousPatterns := []struct {
		pattern string
		message string
	}{
		{`(?i)curl.*\|.*sh`, "potential remote code execution"},
		{`(?i)wget.*\|.*sh`, "potential remote code execution"},
		{`(?i)rm\s+-rf\s+/`, "dangerous file deletion"},
		{`(?i)sudo\s+rm`, "privileged file deletion"},
		{`(?i)format\s+c:`, "potential disk formatting"},
		{`(?i):\(\)\{.*\}`, "potential fork bomb"},
	}

	for _, pattern := range suspiciousPatterns {
		matched, err := regexp.MatchString(pattern.pattern, content)
		if err != nil {
			continue // Skip regex errors
		}
		if matched {
			return fmt.Errorf("suspicious content detected: %s", pattern.message)
		}
	}

	return nil
}

// validateYAMLFrontmatter performs basic YAML frontmatter validation
func (i *Importer) validateYAMLFrontmatter(content string) error {
	// Extract YAML frontmatter
	yamlPattern := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---`)
	matches := yamlPattern.FindStringSubmatch(content)
	if len(matches) < 2 {
		return fmt.Errorf("malformed YAML frontmatter")
	}

	yamlContent := matches[1]
	
	// Basic YAML structure validation
	lines := strings.Split(yamlContent, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Check for key-value format
		if !strings.Contains(line, ":") {
			return fmt.Errorf("invalid YAML syntax at line %d: %s", i+1, line)
		}
	}

	return nil
}

// sanitizeFilename removes dangerous characters from filenames
func sanitizeFilename(filename string) string {
	// Remove path separators and other dangerous characters
	dangerous := []string{"/", "\\", "..", ":", "*", "?", "\"", "<", ">", "|"}
	safe := filename
	
	for _, char := range dangerous {
		safe = strings.ReplaceAll(safe, char, "_")
	}

	// Remove leading/trailing whitespace and dots
	safe = strings.Trim(safe, " .")
	
	// Ensure filename isn't empty
	if safe == "" {
		safe = "unnamed_command"
	}

	// Limit length
	if len(safe) > 100 {
		safe = safe[:100]
	}

	return safe
}

// CheckLocalExists checks which remote commands already exist locally
func (i *Importer) CheckLocalExists(commands []RemoteCommand, localDir string) error {
	for idx := range commands {
		safeFilename := sanitizeFilename(commands[idx].Name) + ".md"
		localPath := filepath.Join(localDir, safeFilename)
		
		if _, err := os.Stat(localPath); err == nil {
			commands[idx].LocalExists = true
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("error checking file %s: %w", localPath, err)
		}
	}
	
	return nil
}

// GetDefaultImportOptions returns default import options
func GetDefaultImportOptions(targetDir string) ImportOptions {
	return ImportOptions{
		OverwriteExisting: false,
		TargetDirectory:   targetDir,
		CreateBackups:     true,
		ValidateContent:   true,
	}
}