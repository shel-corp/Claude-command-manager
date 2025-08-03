package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// UserRegistryManager handles the user's personal repository registry
type UserRegistryManager struct {
	registryPath string
	registry     *UserRegistry
	loaded       bool
}

// NewUserRegistryManager creates a new user registry manager
func NewUserRegistryManager() (*UserRegistryManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "claude_command_manager")
	registryPath := filepath.Join(configDir, "slash_repos.yaml")

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &UserRegistryManager{
		registryPath: registryPath,
	}, nil
}

// Load loads the user registry from disk
func (urm *UserRegistryManager) Load() error {
	data, err := os.ReadFile(urm.registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default registry if it doesn't exist
			urm.registry = &UserRegistry{}
			*urm.registry = DefaultUserRegistry()
			urm.loaded = true
			return urm.Save() // Save the default registry
		}
		return fmt.Errorf("failed to read user registry: %w", err)
	}

	registry := &UserRegistry{}
	if err := yaml.Unmarshal(data, registry); err != nil {
		return fmt.Errorf("failed to parse user registry YAML: %w", err)
	}

	urm.registry = registry
	urm.loaded = true
	return nil
}

// Save saves the user registry to disk
func (urm *UserRegistryManager) Save() error {
	if urm.registry == nil {
		return fmt.Errorf("no registry loaded")
	}

	// Update last modified time
	urm.registry.LastUpdated = time.Now().Format("2006-01-02")

	data, err := yaml.Marshal(urm.registry)
	if err != nil {
		return fmt.Errorf("failed to marshal user registry: %w", err)
	}

	if err := os.WriteFile(urm.registryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write user registry: %w", err)
	}

	return nil
}

// IsLoaded returns true if the registry has been loaded
func (urm *UserRegistryManager) IsLoaded() bool {
	return urm.loaded && urm.registry != nil
}

// GetRegistry returns the loaded user registry
func (urm *UserRegistryManager) GetRegistry() *UserRegistry {
	return urm.registry
}

// GetCategories returns all categories from the user registry
func (urm *UserRegistryManager) GetCategories() map[string]UserCategory {
	if !urm.IsLoaded() {
		return make(map[string]UserCategory)
	}
	return urm.registry.Categories
}

// GetCategoryKeys returns all category keys sorted
func (urm *UserRegistryManager) GetCategoryKeys() []string {
	categories := urm.GetCategories()
	keys := make([]string, 0, len(categories))
	for key := range categories {
		keys = append(keys, key)
	}
	return keys
}

// AddCategory adds a new category to the user registry
func (urm *UserRegistryManager) AddCategory(key, name, description, icon string) error {
	if !urm.IsLoaded() {
		return fmt.Errorf("registry not loaded")
	}

	// Validate key
	key = strings.ToLower(strings.ReplaceAll(key, " ", "_"))
	if key == "" {
		return fmt.Errorf("category key cannot be empty")
	}

	// Check if category already exists
	if _, exists := urm.registry.Categories[key]; exists {
		return fmt.Errorf("category '%s' already exists", key)
	}

	// Add the category
	urm.registry.Categories[key] = UserCategory{
		Name:         name,
		Description:  description,
		Icon:         icon,
		UserCreated:  true,
		Repositories: make([]UserRepository, 0),
	}

	return urm.Save()
}

// AddRepository adds a repository to a category
func (urm *UserRegistryManager) AddRepository(categoryKey string, repo UserRepository) error {
	if !urm.IsLoaded() {
		return fmt.Errorf("registry not loaded")
	}

	// Check if category exists
	category, exists := urm.registry.Categories[categoryKey]
	if !exists {
		return fmt.Errorf("category '%s' does not exist", categoryKey)
	}

	// Check if repository already exists in this category
	for _, existingRepo := range category.Repositories {
		if existingRepo.URL == repo.URL {
			return fmt.Errorf("repository with URL '%s' already exists in category '%s'", repo.URL, categoryKey)
		}
	}

	// Set metadata
	repo.AddedAt = time.Now()
	repo.CategoryKey = categoryKey
	repo.CategoryName = category.Name
	repo.CategoryIcon = category.Icon

	// Add repository to category
	category.Repositories = append(category.Repositories, repo)
	urm.registry.Categories[categoryKey] = category

	return urm.Save()
}

// RemoveRepository removes a repository from a category
func (urm *UserRegistryManager) RemoveRepository(categoryKey, repoURL string) error {
	if !urm.IsLoaded() {
		return fmt.Errorf("registry not loaded")
	}

	category, exists := urm.registry.Categories[categoryKey]
	if !exists {
		return fmt.Errorf("category '%s' does not exist", categoryKey)
	}

	// Find and remove the repository
	for i, repo := range category.Repositories {
		if repo.URL == repoURL {
			// Remove repository from slice
			category.Repositories = append(category.Repositories[:i], category.Repositories[i+1:]...)
			urm.registry.Categories[categoryKey] = category
			return urm.Save()
		}
	}

	return fmt.Errorf("repository with URL '%s' not found in category '%s'", repoURL, categoryKey)
}

// UpdateRepository updates an existing repository
func (urm *UserRegistryManager) UpdateRepository(categoryKey, repoURL string, updatedRepo UserRepository) error {
	if !urm.IsLoaded() {
		return fmt.Errorf("registry not loaded")
	}

	category, exists := urm.registry.Categories[categoryKey]
	if !exists {
		return fmt.Errorf("category '%s' does not exist", categoryKey)
	}

	// Find and update the repository
	for i, repo := range category.Repositories {
		if repo.URL == repoURL {
			// Preserve original added time
			updatedRepo.AddedAt = repo.AddedAt
			updatedRepo.CategoryKey = categoryKey
			updatedRepo.CategoryName = category.Name
			updatedRepo.CategoryIcon = category.Icon
			
			category.Repositories[i] = updatedRepo
			urm.registry.Categories[categoryKey] = category
			return urm.Save()
		}
	}

	return fmt.Errorf("repository with URL '%s' not found in category '%s'", repoURL, categoryKey)
}

// GetAllRepositories returns all repositories from all categories
func (urm *UserRegistryManager) GetAllRepositories() []UserRepository {
	if !urm.IsLoaded() {
		return nil
	}

	var allRepos []UserRepository
	for categoryKey, category := range urm.registry.Categories {
		for _, repo := range category.Repositories {
			// Ensure category metadata is set
			repo.CategoryKey = categoryKey
			repo.CategoryName = category.Name
			repo.CategoryIcon = category.Icon
			allRepos = append(allRepos, repo)
		}
	}

	return allRepos
}

// GetCategoryRepositories returns all repositories in a specific category
func (urm *UserRegistryManager) GetCategoryRepositories(categoryKey string) []UserRepository {
	if !urm.IsLoaded() {
		return nil
	}

	category, exists := urm.registry.Categories[categoryKey]
	if !exists {
		return nil
	}

	// Ensure category metadata is set
	repos := make([]UserRepository, len(category.Repositories))
	for i, repo := range category.Repositories {
		repo.CategoryKey = categoryKey
		repo.CategoryName = category.Name
		repo.CategoryIcon = category.Icon
		repos[i] = repo
	}

	return repos
}

// FindRepository finds a repository by URL across all categories
func (urm *UserRegistryManager) FindRepository(repoURL string) (*UserRepository, string, error) {
	if !urm.IsLoaded() {
		return nil, "", fmt.Errorf("registry not loaded")
	}

	for categoryKey, category := range urm.registry.Categories {
		for _, repo := range category.Repositories {
			if repo.URL == repoURL {
				repo.CategoryKey = categoryKey
				repo.CategoryName = category.Name
				repo.CategoryIcon = category.Icon
				return &repo, categoryKey, nil
			}
		}
	}

	return nil, "", fmt.Errorf("repository not found")
}

// HasRepository checks if a repository URL exists in the user registry
func (urm *UserRegistryManager) HasRepository(repoURL string) bool {
	_, _, err := urm.FindRepository(repoURL)
	return err == nil
}

// GetRegistryPath returns the path to the user registry file
func (urm *UserRegistryManager) GetRegistryPath() string {
	return urm.registryPath
}