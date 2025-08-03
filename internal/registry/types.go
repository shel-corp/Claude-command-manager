package registry

import (
	"time"

	"github.com/shel-corp/Claude-command-manager/internal/remote"
)

// UserRegistry represents the user's personal repository registry
type UserRegistry struct {
	Version     string                      `yaml:"version"`
	LastUpdated string                      `yaml:"last_updated"`
	Categories  map[string]UserCategory     `yaml:"categories"`
}

// UserCategory represents a user-defined category
type UserCategory struct {
	Name         string              `yaml:"name"`
	Description  string              `yaml:"description"`
	Icon         string              `yaml:"icon"`
	UserCreated  bool                `yaml:"user_created"`
	Repositories []UserRepository    `yaml:"repositories"`
}

// UserRepository represents a user-added repository
type UserRepository struct {
	Name        string    `yaml:"name"`
	URL         string    `yaml:"url"`
	Description string    `yaml:"description"`
	Author      string    `yaml:"author"`
	Tags        []string  `yaml:"tags"`
	Verified    bool      `yaml:"verified"`
	AddedAt     time.Time `yaml:"added_at"`
	LastChecked time.Time `yaml:"last_checked,omitempty"`
	
	// Runtime fields for UI (not saved to YAML)
	CategoryKey  string `yaml:"-"`
	CategoryName string `yaml:"-"`
	CategoryIcon string `yaml:"-"`
}

// MergedRegistry represents the combination of bundled and user registries
type MergedRegistry struct {
	Version           string                           `json:"version"`
	LastUpdated       string                           `json:"last_updated"`
	Categories        map[string]MergedCategory        `json:"categories"`
	UserRegistryPath  string                           `json:"user_registry_path"`
	HasUserRegistry   bool                             `json:"has_user_registry"`
}

// MergedCategory represents a category with both bundled and user repositories
type MergedCategory struct {
	Name         string                          `json:"name"`
	Description  string                          `json:"description"`
	Icon         string                          `json:"icon"`
	UserCreated  bool                            `json:"user_created"`
	Repositories []remote.CuratedRepository      `json:"repositories"`
}

// RepositorySource indicates where a repository came from
type RepositorySource int

const (
	SourceBundled RepositorySource = iota
	SourceUser
)

// RepositoryMetadata holds metadata about a repository's source
type RepositoryMetadata struct {
	Source      RepositorySource `json:"source"`
	UserCreated bool             `json:"user_created"`
	AddedAt     time.Time        `json:"added_at,omitempty"`
	LastChecked time.Time        `json:"last_checked,omitempty"`
}

// CategoryInput represents user input for category selection/creation
type CategoryInput struct {
	CategoryKey string `json:"category_key"`
	IsNew       bool   `json:"is_new"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// RepositoryInput represents user input for repository details
type RepositoryInput struct {
	URL         string   `json:"url"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	Category    CategoryInput `json:"category"`
}

// DefaultUserRegistry creates a new empty user registry
func DefaultUserRegistry() UserRegistry {
	return UserRegistry{
		Version:     "1.0",
		LastUpdated: time.Now().Format("2006-01-02"),
		Categories:  make(map[string]UserCategory),
	}
}

// ToRemoteCuratedRepository converts a UserRepository to remote.CuratedRepository
func (ur *UserRepository) ToRemoteCuratedRepository() remote.CuratedRepository {
	return remote.CuratedRepository{
		Name:         ur.Name,
		URL:          ur.URL,
		Description:  ur.Description,
		Author:       ur.Author,
		Tags:         ur.Tags,
		Verified:     ur.Verified,
		CategoryKey:  ur.CategoryKey,
		CategoryName: ur.CategoryName,
		CategoryIcon: ur.CategoryIcon,
	}
}

// FromRemoteCuratedRepository creates a UserRepository from remote.CuratedRepository
func FromRemoteCuratedRepository(repo remote.CuratedRepository) UserRepository {
	return UserRepository{
		Name:        repo.Name,
		URL:         repo.URL,
		Description: repo.Description,
		Author:      repo.Author,
		Tags:        repo.Tags,
		Verified:    false, // User repositories are not pre-verified
		AddedAt:     time.Now(),
		CategoryKey:  repo.CategoryKey,
		CategoryName: repo.CategoryName,
		CategoryIcon: repo.CategoryIcon,
	}
}