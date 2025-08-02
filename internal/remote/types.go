package remote

import "time"

// RemoteRepository represents a GitHub repository containing Claude commands
type RemoteRepository struct {
	Owner       string           `json:"owner"`
	Repo        string           `json:"repo"`
	Branch      string           `json:"branch"`
	Path        string           `json:"path"`
	URL         string           `json:"url"`
	Commands    []RemoteCommand  `json:"commands"`
	LastFetched time.Time        `json:"last_fetched"`
}

// RemoteCommand represents a command found in a remote repository
type RemoteCommand struct {
	Name        string `json:"name"`         // Filename without .md extension
	Path        string `json:"path"`         // Full path in repository
	Description string `json:"description"`  // From YAML frontmatter
	Content     string `json:"content"`      // Full file content
	Size        int64  `json:"size"`         // File size in bytes
	LocalExists bool   `json:"local_exists"` // Whether command exists locally
	Selected    bool   `json:"selected"`     // For multi-select UI
}

// ImportOptions configures how commands are imported
type ImportOptions struct {
	OverwriteExisting bool   `json:"overwrite_existing"`
	TargetDirectory   string `json:"target_directory"`
	CreateBackups     bool   `json:"create_backups"`
	ValidateContent   bool   `json:"validate_content"`
}

// ImportResult contains the results of a command import operation
type ImportResult struct {
	Imported  []string `json:"imported"`   // Successfully imported commands
	Skipped   []string `json:"skipped"`    // Skipped due to conflicts
	Failed    []string `json:"failed"`     // Failed to import
	Errors    []string `json:"errors"`     // Error messages
}

// GitHubAPIError represents errors from GitHub API calls
type GitHubAPIError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (e *GitHubAPIError) Error() string {
	return e.Message
}