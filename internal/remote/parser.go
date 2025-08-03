package remote

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ParseGitHubURL parses various GitHub URL formats and extracts repository information
func ParseGitHubURL(rawURL string) (*RemoteRepository, error) {
	// Normalize URL - add https:// if missing
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	// Validate it's a GitHub URL
	if parsedURL.Host != "github.com" && parsedURL.Host != "www.github.com" {
		return nil, fmt.Errorf("only GitHub URLs are supported, got: %s", parsedURL.Host)
	}

	// Extract path components
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid GitHub URL: missing owner/repo")
	}

	owner := pathParts[0]
	repo := pathParts[1]
	branch := "main" // default branch
	commandPath := ""

	// Handle different URL formats:
	// 1. https://github.com/owner/repo/tree/branch/path/to/commands
	// 2. https://github.com/owner/repo/path/to/commands (assume main branch)
	// 3. https://github.com/owner/repo (assume main/.claude/commands)

	if len(pathParts) > 2 {
		if pathParts[2] == "tree" && len(pathParts) > 3 {
			// Format 1: explicit branch with tree
			branch = pathParts[3]
			if len(pathParts) > 4 {
				commandPath = strings.Join(pathParts[4:], "/")
			}
		} else {
			// Format 2: assume main branch, path starts at index 2
			commandPath = strings.Join(pathParts[2:], "/")
		}
	}

	// If no path specified, assume .claude/commands as default
	if commandPath == "" {
		commandPath = ".claude/commands"
	}

	// No validation on directory name - allow any directory structure
	// The GitHub client will validate if the directory actually contains command files

	// Validate owner and repo names (GitHub naming rules)
	if err := validateGitHubName(owner); err != nil {
		return nil, fmt.Errorf("invalid owner name '%s': %w", owner, err)
	}
	if err := validateGitHubName(repo); err != nil {
		return nil, fmt.Errorf("invalid repository name '%s': %w", repo, err)
	}

	return &RemoteRepository{
		Owner:  owner,
		Repo:   repo,
		Branch: branch,
		Path:   commandPath,
		URL:    rawURL,
	}, nil
}

// validateGitHubName validates GitHub username/repository name format
func validateGitHubName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// GitHub naming rules: alphanumeric, hyphens, underscores, dots
	// Cannot start or end with hyphen, cannot contain consecutive hyphens
	pattern := `^[a-zA-Z0-9]([a-zA-Z0-9._-]*[a-zA-Z0-9])?$`
	matched, err := regexp.MatchString(pattern, name)
	if err != nil {
		return fmt.Errorf("regex error: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid format (must be alphanumeric with . _ - allowed)")
	}

	// Additional GitHub-specific rules
	if strings.Contains(name, "--") {
		return fmt.Errorf("consecutive hyphens not allowed")
	}
	if len(name) > 100 {
		return fmt.Errorf("name too long (max 100 characters)")
	}

	return nil
}

// BuildGitHubAPIURL creates the GitHub API URL for accessing repository contents
func (r *RemoteRepository) BuildGitHubAPIURL(subPath string) string {
	path := r.Path
	if subPath != "" {
		path = path + "/" + strings.TrimPrefix(subPath, "/")
	}
	return fmt.Sprintf("repos/%s/%s/contents/%s?ref=%s", r.Owner, r.Repo, path, r.Branch)
}

// BuildWebURL creates the web URL for viewing the repository in browser
func (r *RemoteRepository) BuildWebURL() string {
	return fmt.Sprintf("https://github.com/%s/%s/tree/%s/%s", r.Owner, r.Repo, r.Branch, r.Path)
}