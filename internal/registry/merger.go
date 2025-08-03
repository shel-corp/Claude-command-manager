package registry

import (
	"fmt"
	"strings"
	"time"

	"github.com/shel-corp/Claude-command-manager/internal/remote"
)

// RegistryMerger handles merging bundled and user registries
type RegistryMerger struct {
	bundledRegistry *remote.RepositoryRegistry
	userRegistry    *UserRegistry
	merged          *MergedRegistry
}

// NewRegistryMerger creates a new registry merger
func NewRegistryMerger(bundledRegistry *remote.RepositoryRegistry, userRegistry *UserRegistry) *RegistryMerger {
	return &RegistryMerger{
		bundledRegistry: bundledRegistry,
		userRegistry:    userRegistry,
	}
}

// Merge combines the bundled and user registries into a unified view
func (rm *RegistryMerger) Merge() (*MergedRegistry, error) {
	merged := &MergedRegistry{
		Version:         "1.0",
		LastUpdated:     time.Now().Format("2006-01-02"),
		Categories:      make(map[string]MergedCategory),
		HasUserRegistry: rm.userRegistry != nil,
	}

	// Start with bundled registry categories
	if rm.bundledRegistry != nil {
		for categoryKey, bundledCategory := range rm.bundledRegistry.Categories {
			mergedCategory := MergedCategory{
				Name:         bundledCategory.Name,
				Description:  bundledCategory.Description,
				Icon:         bundledCategory.Icon,
				UserCreated:  false,
				Repositories: make([]remote.CuratedRepository, 0),
			}

			// Add bundled repositories
			for _, repo := range bundledCategory.Repositories {
				repo.CategoryKey = categoryKey
				repo.CategoryName = bundledCategory.Name
				repo.CategoryIcon = bundledCategory.Icon
				mergedCategory.Repositories = append(mergedCategory.Repositories, repo)
			}

			merged.Categories[categoryKey] = mergedCategory
		}
	}

	// Add or merge user registry categories
	if rm.userRegistry != nil {
		for categoryKey, userCategory := range rm.userRegistry.Categories {
			if existingCategory, exists := merged.Categories[categoryKey]; exists {
				// Category exists in bundled registry - merge repositories
				for _, userRepo := range userCategory.Repositories {
					// Convert UserRepository to CuratedRepository
					curatedRepo := userRepo.ToRemoteCuratedRepository()
					
					// Check for duplicates (by URL)
					isDuplicate := false
					for _, existingRepo := range existingCategory.Repositories {
						if existingRepo.URL == curatedRepo.URL {
							isDuplicate = true
							break
						}
					}
					
					if !isDuplicate {
						existingCategory.Repositories = append(existingCategory.Repositories, curatedRepo)
					}
				}
				merged.Categories[categoryKey] = existingCategory
			} else {
				// New category from user registry
				mergedCategory := MergedCategory{
					Name:         userCategory.Name,
					Description:  userCategory.Description,
					Icon:         userCategory.Icon,
					UserCreated:  userCategory.UserCreated,
					Repositories: make([]remote.CuratedRepository, 0),
				}

				// Add user repositories
				for _, userRepo := range userCategory.Repositories {
					curatedRepo := userRepo.ToRemoteCuratedRepository()
					mergedCategory.Repositories = append(mergedCategory.Repositories, curatedRepo)
				}

				merged.Categories[categoryKey] = mergedCategory
			}
		}
	}

	rm.merged = merged
	return merged, nil
}

// GetMergedRegistry returns the merged registry
func (rm *RegistryMerger) GetMergedRegistry() *MergedRegistry {
	return rm.merged
}

// GetAllRepositories returns all repositories from the merged registry
func (rm *RegistryMerger) GetAllRepositories() []remote.CuratedRepository {
	if rm.merged == nil {
		return nil
	}

	var allRepos []remote.CuratedRepository
	for categoryKey, category := range rm.merged.Categories {
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

// GetCategoryRepositories returns repositories from a specific category
func (rm *RegistryMerger) GetCategoryRepositories(categoryKey string) []remote.CuratedRepository {
	if rm.merged == nil {
		return nil
	}

	category, exists := rm.merged.Categories[categoryKey]
	if !exists {
		return nil
	}

	// Ensure category metadata is set
	repos := make([]remote.CuratedRepository, len(category.Repositories))
	for i, repo := range category.Repositories {
		repo.CategoryKey = categoryKey
		repo.CategoryName = category.Name
		repo.CategoryIcon = category.Icon
		repos[i] = repo
	}

	return repos
}

// SearchRepositories searches repositories by query string across merged registry
func (rm *RegistryMerger) SearchRepositories(query string) []remote.CuratedRepository {
	if rm.merged == nil || query == "" {
		return rm.GetAllRepositories()
	}

	query = strings.ToLower(strings.TrimSpace(query))
	var results []remote.CuratedRepository

	for _, repo := range rm.GetAllRepositories() {
		if rm.matchesQuery(repo, query) {
			results = append(results, repo)
		}
	}

	return results
}

// matchesQuery checks if a repository matches the search query
func (rm *RegistryMerger) matchesQuery(repo remote.CuratedRepository, query string) bool {
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

// GetCategories returns all categories from the merged registry
func (rm *RegistryMerger) GetCategories() map[string]MergedCategory {
	if rm.merged == nil {
		return make(map[string]MergedCategory)
	}
	return rm.merged.Categories
}

// IsUserRepository checks if a repository URL is from the user registry
func (rm *RegistryMerger) IsUserRepository(repoURL string) bool {
	if rm.userRegistry == nil {
		return false
	}

	for _, category := range rm.userRegistry.Categories {
		for _, repo := range category.Repositories {
			if repo.URL == repoURL {
				return true
			}
		}
	}

	return false
}

// GetRepositorySource returns the source of a repository (bundled or user)
func (rm *RegistryMerger) GetRepositorySource(repoURL string) RepositorySource {
	if rm.IsUserRepository(repoURL) {
		return SourceUser
	}
	return SourceBundled
}

// GetRepositoryMetadata returns metadata about a repository
func (rm *RegistryMerger) GetRepositoryMetadata(repoURL string) (*RepositoryMetadata, error) {
	// Check user registry first
	if rm.userRegistry != nil {
		for _, category := range rm.userRegistry.Categories {
			for _, repo := range category.Repositories {
				if repo.URL == repoURL {
					return &RepositoryMetadata{
						Source:      SourceUser,
						UserCreated: true,
						AddedAt:     repo.AddedAt,
						LastChecked: repo.LastChecked,
					}, nil
				}
			}
		}
	}

	// Check bundled registry
	if rm.bundledRegistry != nil {
		for _, category := range rm.bundledRegistry.Categories {
			for _, repo := range category.Repositories {
				if repo.URL == repoURL {
					return &RepositoryMetadata{
						Source:      SourceBundled,
						UserCreated: false,
					}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("repository not found")
}

// ValidateMerge checks for potential conflicts or issues in the merge
func (rm *RegistryMerger) ValidateMerge() []string {
	var warnings []string

	if rm.bundledRegistry == nil && rm.userRegistry == nil {
		warnings = append(warnings, "No registries available to merge")
		return warnings
	}

	if rm.bundledRegistry == nil {
		warnings = append(warnings, "No bundled registry found - only user repositories will be available")
	}

	if rm.userRegistry == nil {
		warnings = append(warnings, "No user registry found - only bundled repositories will be available")
	}

	// Check for duplicate URLs across categories
	if rm.merged != nil {
		urlMap := make(map[string][]string) // URL -> category keys
		
		for categoryKey, category := range rm.merged.Categories {
			for _, repo := range category.Repositories {
				if existingCategories, exists := urlMap[repo.URL]; exists {
					existingCategories = append(existingCategories, categoryKey)
					urlMap[repo.URL] = existingCategories
				} else {
					urlMap[repo.URL] = []string{categoryKey}
				}
			}
		}
		
		for url, categories := range urlMap {
			if len(categories) > 1 {
				warnings = append(warnings, fmt.Sprintf("Repository URL '%s' appears in multiple categories: %v", url, categories))
			}
		}
	}

	return warnings
}