package remote

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// GitHubContent represents a file or directory from GitHub API
type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "dir"
	Size        int64  `json:"size"`
	DownloadURL string `json:"download_url"`
	Content     string `json:"content,omitempty"` // Base64 encoded for files
}

// GitHubClient handles GitHub API interactions using gh command
type GitHubClient struct{}

// NewGitHubClient creates a new GitHub client
func NewGitHubClient() *GitHubClient {
	return &GitHubClient{}
}

// CheckGHInstalled verifies that gh command is available
func (c *GitHubClient) CheckGHInstalled() error {
	cmd := exec.Command("gh", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh command not found - please install GitHub CLI: https://cli.github.com/")
	}
	return nil
}

// FetchCommands recursively fetches all .md files from the repository's commands directory
func (c *GitHubClient) FetchCommands(repo *RemoteRepository) error {
	if err := c.CheckGHInstalled(); err != nil {
		return err
	}

	// Start recursive fetching from the root commands path
	commands, err := c.fetchCommandsRecursive(repo, "")
	if err != nil {
		return fmt.Errorf("failed to fetch commands: %w", err)
	}

	repo.Commands = commands
	return nil
}

// fetchCommandsRecursive recursively fetches commands from a directory
func (c *GitHubClient) fetchCommandsRecursive(repo *RemoteRepository, subPath string) ([]RemoteCommand, error) {
	// Build API URL for this directory
	apiURL := repo.BuildGitHubAPIURL(subPath)
	
	// Fetch directory contents
	cmd := exec.Command("gh", "api", apiURL)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("GitHub API error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute gh command: %w", err)
	}

	// Parse JSON response
	var contents []GitHubContent
	if err := json.Unmarshal(output, &contents); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	var commands []RemoteCommand

	// Process each item in the directory
	for _, item := range contents {
		if item.Type == "dir" {
			// Recursively fetch from subdirectory
			relativePath := item.Path
			if strings.HasPrefix(relativePath, repo.Path+"/") {
				relativePath = relativePath[len(repo.Path)+1:]
			}
			subCommands, err := c.fetchCommandsRecursive(repo, relativePath)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch from subdirectory %s: %w", item.Name, err)
			}
			commands = append(commands, subCommands...)
		} else if item.Type == "file" && strings.HasSuffix(item.Name, ".md") {
			// This is a command file
			cmd := RemoteCommand{
				Name: strings.TrimSuffix(item.Name, ".md"),
				Path: item.Path,
				Size: item.Size,
			}
			commands = append(commands, cmd)
		}
	}

	return commands, nil
}

// FetchCommandContent downloads the full content of a specific command
func (c *GitHubClient) FetchCommandContent(repo *RemoteRepository, command *RemoteCommand) error {
	// Build API URL for the specific file
	apiURL := fmt.Sprintf("repos/%s/%s/contents/%s?ref=%s", repo.Owner, repo.Repo, command.Path, repo.Branch)
	
	cmd := exec.Command("gh", "api", apiURL)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("GitHub API error: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("failed to execute gh command: %w", err)
	}

	// Parse JSON response
	var content GitHubContent
	if err := json.Unmarshal(output, &content); err != nil {
		return fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	// Decode base64 content
	if content.Content != "" {
		// GitHub API returns base64 encoded content
		decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(content.Content, "\n", ""))
		if err != nil {
			return fmt.Errorf("failed to decode base64 content: %w", err)
		}
		command.Content = string(decoded)
	} else if content.DownloadURL != "" {
		// Fallback to download URL
		downloadCmd := exec.Command("curl", "-s", content.DownloadURL)
		downloadOutput, err := downloadCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to download file content: %w", err)
		}
		command.Content = string(downloadOutput)
	} else {
		return fmt.Errorf("no content available for file: %s", command.Path)
	}

	// Extract description from YAML frontmatter
	command.Description = extractDescription(command.Content)

	return nil
}

// extractDescription extracts the description from YAML frontmatter or first paragraph
func extractDescription(content string) string {
	// First try YAML frontmatter
	yamlPattern := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---`)
	matches := yamlPattern.FindStringSubmatch(content)
	if len(matches) >= 2 {
		yamlContent := matches[1]
		
		// Extract description field
		descPattern := regexp.MustCompile(`(?m)^description:\s*(.+)$`)
		descMatches := descPattern.FindStringSubmatch(yamlContent)
		if len(descMatches) >= 2 {
			// Clean up the description (remove quotes if present)
			description := strings.TrimSpace(descMatches[1])
			description = strings.Trim(description, `"'`)
			return description
		}
	}

	// Fallback: use first line or paragraph
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and markdown headers
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") {
			// Take first meaningful line, truncate if too long
			if len(line) > 80 {
				line = line[:77] + "..."
			}
			return line
		}
	}
	
	return "No description available"
}

// ValidateRepository checks if the repository and commands path exist
func (c *GitHubClient) ValidateRepository(repo *RemoteRepository) error {
	if err := c.CheckGHInstalled(); err != nil {
		return err
	}

	// Try to fetch the repository info first
	repoURL := fmt.Sprintf("repos/%s/%s", repo.Owner, repo.Repo)
	cmd := exec.Command("gh", "api", repoURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("repository not found or not accessible: %s/%s", repo.Owner, repo.Repo)
	}

	// Check if the commands directory exists
	apiURL := repo.BuildGitHubAPIURL("")
	cmd = exec.Command("gh", "api", apiURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("commands directory not found at path: %s", repo.Path)
	}

	return nil
}