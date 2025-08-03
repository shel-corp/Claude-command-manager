package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// RepositoryRegistry represents the complete repository registry
type RepositoryRegistry struct {
	Version     string                       `yaml:"version"`
	LastUpdated string                       `yaml:"last_updated"`
	Categories  map[string]RepositoryCategory `yaml:"categories"`
}

// RepositoryCategory represents a category of repositories
type RepositoryCategory struct {
	Name         string                   `yaml:"name"`
	Description  string                   `yaml:"description"`
	Icon         string                   `yaml:"icon"`
	Repositories []CuratedRepository      `yaml:"repositories"`
}

// CuratedRepository represents a repository from the curated registry
type CuratedRepository struct {
	Name        string   `yaml:"name"`
	URL         string   `yaml:"url"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Tags        []string `yaml:"tags"`
	Verified    bool     `yaml:"verified"`
	Language    string   `yaml:"language,omitempty"`
	Difficulty  string   `yaml:"difficulty,omitempty"`
	LastChecked string   `yaml:"last_checked,omitempty"`
	
	// Runtime fields for UI
	CategoryKey  string `yaml:"-"`
	CategoryName string `yaml:"-"`
	CategoryIcon string `yaml:"-"`
}

// RegistryManager handles loading and searching the repository registry
type RegistryManager struct {
	registry     *RepositoryRegistry
	loadedAt     time.Time
	allRepos     []CuratedRepository // Flattened list for searching
	cacheManager CacheManager        // Interface for cache operations
}

// CacheManager interface for cache operations (simplified)
type CacheManager interface {
	GetRegistryCacheRaw() ([]byte, time.Time, bool, error) // data, cachedAt, isExpired, error
	SetRegistryCache(registry interface{}, etag string) error
	IsEnabled() bool
}

// NewRegistryManager creates a new registry manager
func NewRegistryManager() *RegistryManager {
	return &RegistryManager{}
}

// SetCacheManager sets the cache manager for the registry
func (rm *RegistryManager) SetCacheManager(cacheManager CacheManager) {
	rm.cacheManager = cacheManager
}

// LoadRegistry loads the repository registry from cache or YAML file
func (rm *RegistryManager) LoadRegistry() error {
	// Try to load from cache first
	if rm.cacheManager != nil && rm.cacheManager.IsEnabled() {
		if cachedData, cachedAt, isExpired, err := rm.cacheManager.GetRegistryCacheRaw(); err == nil && cachedData != nil && !isExpired {
			// Parse cached registry data
			registry := &RepositoryRegistry{}
			if err := json.Unmarshal(cachedData, registry); err == nil {
				// Use cached data
				rm.registry = registry
				rm.loadedAt = cachedAt
				rm.buildFlattenedList()
				return nil
			}
		}
	}

	// Cache miss or expired - load from file
	registryPath, err := rm.findRegistryFile()
	if err != nil {
		return fmt.Errorf("failed to find registry file: %w", err)
	}

	data, err := os.ReadFile(registryPath)
	if err != nil {
		return fmt.Errorf("failed to read registry file: %w", err)
	}

	registry := &RepositoryRegistry{}
	if err := yaml.Unmarshal(data, registry); err != nil {
		return fmt.Errorf("failed to parse registry YAML: %w", err)
	}

	rm.registry = registry
	rm.loadedAt = time.Now()
	rm.buildFlattenedList()

	// Cache the loaded registry
	if rm.cacheManager != nil && rm.cacheManager.IsEnabled() {
		if err := rm.cacheManager.SetRegistryCache(*registry, ""); err != nil {
			// Log error but don't fail - caching is optional
			// In a real implementation, we'd use proper logging
			fmt.Printf("Warning: failed to cache registry: %v\n", err)
		}
	}

	return nil
}

// buildFlattenedList builds the flattened repository list for searching
func (rm *RegistryManager) buildFlattenedList() {
	rm.allRepos = make([]CuratedRepository, 0)
	if rm.registry == nil {
		return
	}
	
	for categoryKey, category := range rm.registry.Categories {
		for _, repo := range category.Repositories {
			repo.CategoryKey = categoryKey
			repo.CategoryName = category.Name
			repo.CategoryIcon = category.Icon
			rm.allRepos = append(rm.allRepos, repo)
		}
	}
}

// LoadRegistryWithCache loads registry with cache support (for background refresh)
func (rm *RegistryManager) LoadRegistryWithCache(ctx context.Context) error {
	return rm.LoadRegistry()
}

// GetRegistry returns the loaded registry
func (rm *RegistryManager) GetRegistry() *RepositoryRegistry {
	return rm.registry
}

// GetCategories returns all categories
func (rm *RegistryManager) GetCategories() map[string]RepositoryCategory {
	if rm.registry == nil {
		return nil
	}
	return rm.registry.Categories
}

// GetCategoryRepositories returns all repositories in a specific category
func (rm *RegistryManager) GetCategoryRepositories(categoryKey string) []CuratedRepository {
	if rm.registry == nil {
		return nil
	}
	
	category, exists := rm.registry.Categories[categoryKey]
	if !exists {
		return nil
	}
	
	// Enrich with category info
	repos := make([]CuratedRepository, len(category.Repositories))
	for i, repo := range category.Repositories {
		repo.CategoryKey = categoryKey
		repo.CategoryName = category.Name
		repo.CategoryIcon = category.Icon
		repos[i] = repo
	}
	
	return repos
}

// SearchRepositories searches repositories by query string
func (rm *RegistryManager) SearchRepositories(query string) []CuratedRepository {
	if rm.registry == nil || query == "" {
		return rm.allRepos
	}

	query = strings.ToLower(strings.TrimSpace(query))
	var results []CuratedRepository

	for _, repo := range rm.allRepos {
		if rm.matchesQuery(repo, query) {
			results = append(results, repo)
		}
	}

	return results
}

// FilterByTags filters repositories by tags
func (rm *RegistryManager) FilterByTags(tags []string) []CuratedRepository {
	if rm.registry == nil || len(tags) == 0 {
		return rm.allRepos
	}

	var results []CuratedRepository
	for _, repo := range rm.allRepos {
		if rm.hasAnyTag(repo, tags) {
			results = append(results, repo)
		}
	}

	return results
}

// FilterByCategory filters repositories by category
func (rm *RegistryManager) FilterByCategory(categoryKey string) []CuratedRepository {
	if rm.registry == nil || categoryKey == "" {
		return rm.allRepos
	}

	var results []CuratedRepository
	for _, repo := range rm.allRepos {
		if repo.CategoryKey == categoryKey {
			results = append(results, repo)
		}
	}

	return results
}

// GetAllRepositories returns all repositories from the registry
func (rm *RegistryManager) GetAllRepositories() []CuratedRepository {
	return rm.allRepos
}

// matchesQuery checks if a repository matches the search query
func (rm *RegistryManager) matchesQuery(repo CuratedRepository, query string) bool {
	// Search in name
	if strings.Contains(strings.ToLower(repo.Name), query) {
		return true
	}
	
	// Search in description
	if strings.Contains(strings.ToLower(repo.Description), query) {
		return true
	}
	
	// Search in author
	if strings.Contains(strings.ToLower(repo.Author), query) {
		return true
	}
	
	// Search in tags
	for _, tag := range repo.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	
	// Search in category name
	if strings.Contains(strings.ToLower(repo.CategoryName), query) {
		return true
	}

	return false
}

// hasAnyTag checks if repository has any of the specified tags
func (rm *RegistryManager) hasAnyTag(repo CuratedRepository, tags []string) bool {
	for _, targetTag := range tags {
		for _, repoTag := range repo.Tags {
			if strings.EqualFold(repoTag, targetTag) {
				return true
			}
		}
	}
	return false
}

// IsLoaded returns true if the registry has been loaded
func (rm *RegistryManager) IsLoaded() bool {
	return rm.registry != nil
}

// GetLoadTime returns when the registry was loaded
func (rm *RegistryManager) GetLoadTime() time.Time {
	return rm.loadedAt
}

// findRegistryFile finds the registry YAML file by searching up the directory tree
func (rm *RegistryManager) findRegistryFile() (string, error) {
	// Try different possible locations for the registry file
	possiblePaths := []string{
		"internal/assets/slash_repos.yaml",                    // From project root
		"../assets/slash_repos.yaml",                         // From internal/remote
		"../../internal/assets/slash_repos.yaml",             // From bin or other subdirs
		"assets/slash_repos.yaml",                            // From internal
		"slash_repos.yaml",                                   // Current directory
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Search in current directory and parent directories
	currentDir := wd
	for {
		for _, relativePath := range possiblePaths {
			fullPath := filepath.Join(currentDir, relativePath)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath, nil
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root directory
			break
		}
		currentDir = parentDir
	}

	return "", fs.ErrNotExist
}