package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/shel-corp/Claude-command-manager/internal/remote"
)

// EnhancedRegistryManager manages both bundled and user registries with caching
type EnhancedRegistryManager struct {
	bundledManager *remote.RegistryManager
	userManager    *UserRegistryManager
	merger         *RegistryMerger
	merged         *MergedRegistry
	loadedAt       time.Time
	cacheManager   remote.CacheManager
}

// NewEnhancedRegistryManager creates a new enhanced registry manager
func NewEnhancedRegistryManager() (*EnhancedRegistryManager, error) {
	// Initialize bundled registry manager
	bundledManager := remote.NewRegistryManager()
	
	// Initialize user registry manager
	userManager, err := NewUserRegistryManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create user registry manager: %w", err)
	}

	return &EnhancedRegistryManager{
		bundledManager: bundledManager,
		userManager:    userManager,
	}, nil
}

// SetCacheManager sets the cache manager for the registry
func (erm *EnhancedRegistryManager) SetCacheManager(cacheManager remote.CacheManager) {
	erm.cacheManager = cacheManager
	erm.bundledManager.SetCacheManager(cacheManager)
}

// LoadRegistries loads both bundled and user registries and merges them
func (erm *EnhancedRegistryManager) LoadRegistries() error {
	// Load bundled registry
	if err := erm.bundledManager.LoadRegistry(); err != nil {
		// Log warning but continue - user registry might still work
		fmt.Printf("Warning: failed to load bundled registry: %v\n", err)
	}

	// Load user registry
	if err := erm.userManager.Load(); err != nil {
		// This is more serious - we need user registry for adding repos
		return fmt.Errorf("failed to load user registry: %w", err)
	}

	// Merge registries
	return erm.mergeRegistries()
}

// mergeRegistries combines bundled and user registries
func (erm *EnhancedRegistryManager) mergeRegistries() error {
	bundledRegistry := erm.bundledManager.GetRegistry()
	userRegistry := erm.userManager.GetRegistry()

	erm.merger = NewRegistryMerger(bundledRegistry, userRegistry)
	
	merged, err := erm.merger.Merge()
	if err != nil {
		return fmt.Errorf("failed to merge registries: %w", err)
	}

	erm.merged = merged
	erm.loadedAt = time.Now()

	// Validate merge and log warnings
	warnings := erm.merger.ValidateMerge()
	for _, warning := range warnings {
		fmt.Printf("Registry merge warning: %s\n", warning)
	}

	return nil
}

// LoadRegistriesWithCache loads registries with cache support (for background refresh)
func (erm *EnhancedRegistryManager) LoadRegistriesWithCache(ctx context.Context) error {
	return erm.LoadRegistries()
}

// IsLoaded returns true if registries have been loaded and merged
func (erm *EnhancedRegistryManager) IsLoaded() bool {
	return erm.merged != nil
}

// GetRegistry returns the merged registry (compatible with existing remote.RegistryManager interface)
func (erm *EnhancedRegistryManager) GetRegistry() *remote.RepositoryRegistry {
	if !erm.IsLoaded() {
		return nil
	}

	// Convert merged registry back to remote.RepositoryRegistry format for compatibility
	bundledRegistry := erm.bundledManager.GetRegistry()
	if bundledRegistry != nil {
		return bundledRegistry
	}

	return nil
}

// GetMergedRegistry returns the full merged registry
func (erm *EnhancedRegistryManager) GetMergedRegistry() *MergedRegistry {
	return erm.merged
}

// GetCategories returns all categories from merged registry
func (erm *EnhancedRegistryManager) GetCategories() map[string]remote.RepositoryCategory {
	if !erm.IsLoaded() {
		return make(map[string]remote.RepositoryCategory)
	}

	// Convert merged categories back to remote.RepositoryCategory format
	categories := make(map[string]remote.RepositoryCategory)
	for key, mergedCategory := range erm.merged.Categories {
		categories[key] = remote.RepositoryCategory{
			Name:         mergedCategory.Name,
			Description:  mergedCategory.Description,
			Icon:         mergedCategory.Icon,
			Repositories: mergedCategory.Repositories,
		}
	}

	return categories
}

// GetCategoryRepositories returns repositories from a specific category
func (erm *EnhancedRegistryManager) GetCategoryRepositories(categoryKey string) []remote.CuratedRepository {
	if !erm.IsLoaded() {
		return nil
	}

	return erm.merger.GetCategoryRepositories(categoryKey)
}

// GetAllRepositories returns all repositories from merged registry
func (erm *EnhancedRegistryManager) GetAllRepositories() []remote.CuratedRepository {
	if !erm.IsLoaded() {
		return nil
	}

	return erm.merger.GetAllRepositories()
}

// SearchRepositories searches repositories by query string
func (erm *EnhancedRegistryManager) SearchRepositories(query string) []remote.CuratedRepository {
	if !erm.IsLoaded() {
		return nil
	}

	return erm.merger.SearchRepositories(query)
}

// GetLoadTime returns when the registries were loaded
func (erm *EnhancedRegistryManager) GetLoadTime() time.Time {
	return erm.loadedAt
}

// User Repository Management Methods

// AddCustomRepository adds a custom repository to the user registry
func (erm *EnhancedRegistryManager) AddCustomRepository(input RepositoryInput) error {
	if !erm.userManager.IsLoaded() {
		return fmt.Errorf("user registry not loaded")
	}

	// Create or get category
	categoryKey := input.Category.CategoryKey
	if input.Category.IsNew {
		// Create new category
		categoryKey = input.Category.CategoryKey
		if err := erm.userManager.AddCategory(
			categoryKey,
			input.Category.Name,
			input.Category.Description,
			input.Category.Icon,
		); err != nil {
			return fmt.Errorf("failed to create category: %w", err)
		}
	}

	// Create user repository
	userRepo := UserRepository{
		Name:        input.Name,
		URL:         input.URL,
		Description: input.Description,
		Author:      input.Author,
		Tags:        input.Tags,
		Verified:    false, // User repositories are not pre-verified
		AddedAt:     time.Now(),
	}

	// Add repository to category
	if err := erm.userManager.AddRepository(categoryKey, userRepo); err != nil {
		return fmt.Errorf("failed to add repository: %w", err)
	}

	// Re-merge registries to update the unified view
	return erm.mergeRegistries()
}

// RemoveCustomRepository removes a custom repository from the user registry
func (erm *EnhancedRegistryManager) RemoveCustomRepository(repoURL string) error {
	if !erm.userManager.IsLoaded() {
		return fmt.Errorf("user registry not loaded")
	}

	// Find the repository
	_, categoryKey, err := erm.userManager.FindRepository(repoURL)
	if err != nil {
		return fmt.Errorf("repository not found: %w", err)
	}

	// Remove repository
	if err := erm.userManager.RemoveRepository(categoryKey, repoURL); err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}

	// Re-merge registries to update the unified view
	return erm.mergeRegistries()
}

// UpdateCustomRepository updates a custom repository in the user registry
func (erm *EnhancedRegistryManager) UpdateCustomRepository(repoURL string, input RepositoryInput) error {
	if !erm.userManager.IsLoaded() {
		return fmt.Errorf("user registry not loaded")
	}

	// Find the repository
	_, oldCategoryKey, err := erm.userManager.FindRepository(repoURL)
	if err != nil {
		return fmt.Errorf("repository not found: %w", err)
	}

	// Create updated repository
	updatedRepo := UserRepository{
		Name:        input.Name,
		URL:         input.URL,
		Description: input.Description,
		Author:      input.Author,
		Tags:        input.Tags,
		Verified:    false,
		LastChecked: time.Now(),
	}

	// Handle category changes
	newCategoryKey := input.Category.CategoryKey
	if input.Category.IsNew {
		// Create new category
		if err := erm.userManager.AddCategory(
			newCategoryKey,
			input.Category.Name,
			input.Category.Description,
			input.Category.Icon,
		); err != nil {
			return fmt.Errorf("failed to create category: %w", err)
		}
	}

	if oldCategoryKey != newCategoryKey {
		// Remove from old category and add to new category
		if err := erm.userManager.RemoveRepository(oldCategoryKey, repoURL); err != nil {
			return fmt.Errorf("failed to remove from old category: %w", err)
		}
		if err := erm.userManager.AddRepository(newCategoryKey, updatedRepo); err != nil {
			return fmt.Errorf("failed to add to new category: %w", err)
		}
	} else {
		// Update in same category
		if err := erm.userManager.UpdateRepository(oldCategoryKey, repoURL, updatedRepo); err != nil {
			return fmt.Errorf("failed to update repository: %w", err)
		}
	}

	// Re-merge registries to update the unified view
	return erm.mergeRegistries()
}

// IsCustomRepository checks if a repository is from the user registry
func (erm *EnhancedRegistryManager) IsCustomRepository(repoURL string) bool {
	if !erm.IsLoaded() {
		return false
	}

	return erm.merger.IsUserRepository(repoURL)
}

// GetCustomRepository gets a custom repository by URL
func (erm *EnhancedRegistryManager) GetCustomRepository(repoURL string) (*UserRepository, error) {
	if !erm.userManager.IsLoaded() {
		return nil, fmt.Errorf("user registry not loaded")
	}

	repo, _, err := erm.userManager.FindRepository(repoURL)
	return repo, err
}

// GetUserCategories returns categories that can be used for adding custom repositories
func (erm *EnhancedRegistryManager) GetUserCategories() map[string]UserCategory {
	if !erm.userManager.IsLoaded() {
		return make(map[string]UserCategory)
	}

	return erm.userManager.GetCategories()
}

// GetAvailableCategories returns all available categories (bundled + user) for selection
func (erm *EnhancedRegistryManager) GetAvailableCategories() map[string]string {
	categories := make(map[string]string)

	// Add bundled categories
	if bundledCategories := erm.GetCategories(); bundledCategories != nil {
		for key, category := range bundledCategories {
			categories[key] = category.Name
		}
	}

	// Add user categories
	if userCategories := erm.GetUserCategories(); userCategories != nil {
		for key, category := range userCategories {
			categories[key] = category.Name
		}
	}

	return categories
}

// GetUserRegistryManager returns the user registry manager for direct access
func (erm *EnhancedRegistryManager) GetUserRegistryManager() *UserRegistryManager {
	return erm.userManager
}