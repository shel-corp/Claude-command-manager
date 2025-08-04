package remote

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// isExcludedFile checks if a file should be excluded from command scanning
// Excludes files with all uppercase names (like README.md, CLAUDE.md, etc.)
func isExcludedFile(filename string) bool {
	nameWithoutExt := strings.TrimSuffix(filename, ".md")
	return strings.ToUpper(nameWithoutExt) == nameWithoutExt
}

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
type GitHubClient struct{
	cacheManager CacheManager // For repository caching
}

// RepositoryCacheManager interface for repository caching operations
type RepositoryCacheManager interface {
	GetRepositoryCacheRaw(repoKey string) ([]byte, []byte, time.Time, bool, string, error) // repoData, commandsData, cachedAt, isExpired, etag, error
	SetRepositoryCache(repoKey string, repo interface{}, commands interface{}, etag string) error
	GetRepositoryKey(owner, repo, branch, path string) string
	IsEnabled() bool
}

// NewGitHubClient creates a new GitHub client
func NewGitHubClient() *GitHubClient {
	return &GitHubClient{}
}

// SetCacheManager sets the cache manager for the GitHub client
func (c *GitHubClient) SetCacheManager(cacheManager CacheManager) {
	c.cacheManager = cacheManager
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
	return c.FetchCommandsWithCache(repo, false)
}

// FetchCommandsWithCache fetches commands with optional cache support
func (c *GitHubClient) FetchCommandsWithCache(repo *RemoteRepository, useCache bool) error {
	if err := c.CheckGHInstalled(); err != nil {
		return err
	}

	// Try cache first if enabled
	if useCache && c.cacheManager != nil && c.cacheManager.IsEnabled() {
		repoKey := c.generateRepoKey(repo)
		if cachedRepo, cachedCommands, cachedAt, isExpired, _, err := c.getCachedRepositoryData(repoKey); err == nil && cachedRepo != nil && !isExpired {
			// Use cached data
			repo.Commands = cachedCommands
			repo.LastFetched = cachedAt
			return nil
		}
	}

	// Cache miss or disabled - fetch from GitHub
	commands, err := c.fetchCommandsRecursive(repo, "")
	if err != nil {
		return fmt.Errorf("failed to fetch commands: %w", err)
	}

	repo.Commands = commands
	repo.LastFetched = time.Now()

	// Cache the fetched data
	if useCache && c.cacheManager != nil && c.cacheManager.IsEnabled() {
		repoKey := c.generateRepoKey(repo)
		if err := c.cacheRepositoryData(repoKey, repo, commands); err != nil {
			// Log error but don't fail
			fmt.Printf("Warning: failed to cache repository data: %v\n", err)
		}
	}

	return nil
}

// generateRepoKey generates a cache key for the repository
func (c *GitHubClient) generateRepoKey(repo *RemoteRepository) string {
	if c.cacheManager != nil {
		if rm, ok := c.cacheManager.(RepositoryCacheManager); ok {
			return rm.GetRepositoryKey(repo.Owner, repo.Repo, repo.Branch, repo.Path)
		}
	}
	// Fallback key generation
	return fmt.Sprintf("%s_%s_%s_%s", repo.Owner, repo.Repo, repo.Branch, strings.ReplaceAll(repo.Path, "/", "_"))
}

// getCachedRepositoryData retrieves cached repository data
func (c *GitHubClient) getCachedRepositoryData(repoKey string) (*RemoteRepository, []RemoteCommand, time.Time, bool, string, error) {
	if rm, ok := c.cacheManager.(RepositoryCacheManager); ok {
		repoData, commandsData, cachedAt, isExpired, etag, err := rm.GetRepositoryCacheRaw(repoKey)
		if err != nil || repoData == nil {
			return nil, nil, time.Time{}, false, "", err
		}

		var repo RemoteRepository
		if err := json.Unmarshal(repoData, &repo); err != nil {
			return nil, nil, time.Time{}, false, "", err
		}

		var commands []RemoteCommand
		if err := json.Unmarshal(commandsData, &commands); err != nil {
			return nil, nil, time.Time{}, false, "", err
		}

		return &repo, commands, cachedAt, isExpired, etag, nil
	}
	return nil, nil, time.Time{}, false, "", fmt.Errorf("cache manager does not support repository caching")
}

// cacheRepositoryData stores repository data in cache
func (c *GitHubClient) cacheRepositoryData(repoKey string, repo *RemoteRepository, commands []RemoteCommand) error {
	if rm, ok := c.cacheManager.(RepositoryCacheManager); ok {
		return rm.SetRepositoryCache(repoKey, *repo, commands, "")
	}
	return fmt.Errorf("cache manager does not support repository caching")
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
		} else if item.Type == "file" && strings.HasSuffix(item.Name, ".md") && !isExcludedFile(item.Name) {
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

// GetRepositoryInfo detects the current Git repository information
func GetRepositoryInfo() (*RemoteRepository, error) {
	// Get the remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git remote URL (make sure you're in a git repository): %w", err)
	}
	
	remoteURL := strings.TrimSpace(string(output))
	if remoteURL == "" {
		return nil, fmt.Errorf("no git remote URL found")
	}
	
	// Parse the GitHub URL to extract owner and repo
	repo, err := ParseGitHubURL(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse git remote URL '%s': %w", remoteURL, err)
	}
	
	// Validate that this is the expected repository
	expectedOwner := "shel-corp"
	expectedRepo := "Claude-command-manager"
	
	if repo.Owner != expectedOwner || repo.Repo != expectedRepo {
		return nil, fmt.Errorf("unexpected repository: %s/%s (expected: %s/%s)", 
			repo.Owner, repo.Repo, expectedOwner, expectedRepo)
	}
	
	return repo, nil
}

// CreateGitHubIssue creates a GitHub issue using the gh CLI
func CreateGitHubIssue(repo *RemoteRepository, title, body string) error {
	// Check if gh CLI is available
	if err := exec.Command("gh", "--version").Run(); err != nil {
		return fmt.Errorf("GitHub CLI (gh) is required but not installed. Please install it from https://cli.github.com/")
	}
	
	// Prepare the issue body with additional context
	enhancedBody := body + "\n\n---\n\n**Submitted via ccm** 🤖\n\n" +
		"This issue was reported through the Claude Command Manager (ccm) application."
	
	repoSpec := fmt.Sprintf("%s/%s", repo.Owner, repo.Repo)
	
	// First, try to create the labels if they don't exist
	createLabelsIfNeeded(repoSpec)
	
	// Create the issue using gh CLI
	cmd := exec.Command("gh", "issue", "create", 
		"--repo", repoSpec,
		"--title", title,
		"--body", enhancedBody,
		"--label", "user-report,ccm-generated")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If it failed due to labels, try again without labels
		if strings.Contains(string(output), "not found") && strings.Contains(string(output), "label") {
			fmt.Printf("Warning: Could not add labels, creating issue without labels...\n")
			cmd = exec.Command("gh", "issue", "create", 
				"--repo", repoSpec,
				"--title", title,
				"--body", enhancedBody)
			
			output, err = cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to create GitHub issue: %w\n\nOutput: %s", err, string(output))
			}
		} else {
			return fmt.Errorf("failed to create GitHub issue: %w\n\nOutput: %s", err, string(output))
		}
	}
	
	return nil
}

// createLabelsIfNeeded creates the required labels if they don't exist
func createLabelsIfNeeded(repoSpec string) {
	labels := []struct {
		name        string
		color       string
		description string
	}{
		{"user-report", "0052cc", "Issue reported by a user through ccm"},
		{"ccm-generated", "5319e7", "Automatically generated by Claude Command Manager"},
	}
	
	for _, label := range labels {
		// Check if label exists (ignore errors - we'll handle missing labels gracefully)
		checkCmd := exec.Command("gh", "label", "list", "--repo", repoSpec, "--search", label.name)
		output, err := checkCmd.Output()
		if err != nil || !strings.Contains(string(output), label.name) {
			// Try to create the label (ignore errors - non-critical)
			createCmd := exec.Command("gh", "label", "create", label.name, 
				"--repo", repoSpec,
				"--color", label.color,
				"--description", label.description)
			createCmd.Run() // Ignore errors - labels are optional
		}
	}
}