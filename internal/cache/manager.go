package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/shel-corp/Claude-command-manager/internal/remote"
)

// Manager handles all cache operations
type Manager struct {
	config      CacheConfig
	cacheDir    string
	mu          sync.RWMutex
	stats       CacheStats
	initialized bool
}

// NewManager creates a new cache manager
func NewManager(config CacheConfig) (*Manager, error) {
	// Set default cache directory if not specified
	if config.Directory == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		config.Directory = filepath.Join(homeDir, ".config", "claude_command_manager", "cache")
	}

	manager := &Manager{
		config:   config,
		cacheDir: config.Directory,
		stats:    CacheStats{},
	}

	if config.Enabled {
		if err := manager.initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize cache: %w", err)
		}
	}

	return manager, nil
}

// initialize sets up the cache directory structure
func (m *Manager) initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create cache directories
	dirs := []string{
		m.cacheDir,
		filepath.Join(m.cacheDir, "repositories"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory %s: %w", dir, err)
		}
	}

	// Load existing metadata or create new
	if err := m.loadMetadata(); err != nil {
		// If metadata doesn't exist, create new
		metadata := CacheMetadata{
			Version:     "1.0",
			LastRefresh: time.Now(),
			TotalSize:   0,
			ItemCount:   0,
		}
		if err := m.saveMetadata(metadata); err != nil {
			return fmt.Errorf("failed to create cache metadata: %w", err)
		}
	}

	m.initialized = true
	return nil
}

// IsEnabled returns whether caching is enabled
func (m *Manager) IsEnabled() bool {
	return m.config.Enabled && m.initialized
}

// GetRegistryCacheRaw retrieves cached registry data as raw JSON
func (m *Manager) GetRegistryCacheRaw() ([]byte, time.Time, bool, error) {
	if !m.IsEnabled() {
		return nil, time.Time{}, false, fmt.Errorf("cache is disabled")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	registryPath := filepath.Join(m.cacheDir, "registry.json")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.stats.RegistryMisses++
			return nil, time.Time{}, false, nil // Cache miss
		}
		return nil, time.Time{}, false, fmt.Errorf("failed to read registry cache: %w", err)
	}

	var registryCache RegistryCache
	if err := json.Unmarshal(data, &registryCache); err != nil {
		// Cache corrupted, treat as miss
		m.stats.RegistryMisses++
		return nil, time.Time{}, false, nil
	}

	m.stats.RegistryHits++
	
	// Check if expired
	isExpired := registryCache.IsExpired()
	
	// Return the registry data as JSON
	registryData, err := json.Marshal(registryCache.Registry)
	if err != nil {
		return nil, time.Time{}, false, fmt.Errorf("failed to marshal registry data: %w", err)
	}
	
	return registryData, registryCache.CachedAt, isExpired, nil
}

// GetRegistryCache retrieves cached registry data
func (m *Manager) GetRegistryCache() (*RegistryCache, error) {
	if !m.IsEnabled() {
		return nil, fmt.Errorf("cache is disabled")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	registryPath := filepath.Join(m.cacheDir, "registry.json")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.stats.RegistryMisses++
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to read registry cache: %w", err)
	}

	var registryCache RegistryCache
	if err := json.Unmarshal(data, &registryCache); err != nil {
		// Cache corrupted, treat as miss
		m.stats.RegistryMisses++
		return nil, nil
	}

	m.stats.RegistryHits++
	return &registryCache, nil
}

// SetRegistryCache stores registry data in cache
func (m *Manager) SetRegistryCache(registry interface{}, etag string) error {
	if !m.IsEnabled() {
		return nil // Silently skip if disabled
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	
	// Convert interface{} to our remote.RepositoryRegistry type
	var regData remote.RepositoryRegistry
	switch r := registry.(type) {
	case remote.RepositoryRegistry:
		regData = r
	default:
		// Try to marshal/unmarshal to convert
		data, err := json.Marshal(registry)
		if err != nil {
			return fmt.Errorf("failed to marshal registry for caching: %w", err)
		}
		if err := json.Unmarshal(data, &regData); err != nil {
			return fmt.Errorf("failed to unmarshal registry for caching: %w", err)
		}
	}
	
	registryCache := RegistryCache{
		Registry:    regData,
		CachedAt:    now,
		ExpiresAt:   now.Add(time.Duration(m.config.TTLHours) * time.Hour),
		ETag:        etag,
		LastChecked: now,
	}

	data, err := json.MarshalIndent(registryCache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry cache: %w", err)
	}

	registryPath := filepath.Join(m.cacheDir, "registry.json")
	if err := os.WriteFile(registryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry cache: %w", err)
	}

	return nil
}

// GetRepositoryCache retrieves cached repository data
func (m *Manager) GetRepositoryCache(repoKey string) (*RepositoryCache, error) {
	if !m.IsEnabled() {
		return nil, fmt.Errorf("cache is disabled")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	repoPath := filepath.Join(m.cacheDir, "repositories", m.sanitizeRepoKey(repoKey)+".json")
	data, err := os.ReadFile(repoPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.stats.RepoMisses++
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to read repository cache: %w", err)
	}

	var repoCache RepositoryCache
	if err := json.Unmarshal(data, &repoCache); err != nil {
		// Cache corrupted, treat as miss
		m.stats.RepoMisses++
		return nil, nil
	}

	m.stats.RepoHits++
	return &repoCache, nil
}

// GetRepositoryCacheRaw retrieves cached repository data as raw JSON
func (m *Manager) GetRepositoryCacheRaw(repoKey string) ([]byte, []byte, time.Time, bool, string, error) {
	if !m.IsEnabled() {
		return nil, nil, time.Time{}, false, "", fmt.Errorf("cache is disabled")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	repoPath := filepath.Join(m.cacheDir, "repositories", m.sanitizeRepoKey(repoKey)+".json")
	data, err := os.ReadFile(repoPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.stats.RepoMisses++
			return nil, nil, time.Time{}, false, "", nil // Cache miss
		}
		return nil, nil, time.Time{}, false, "", fmt.Errorf("failed to read repository cache: %w", err)
	}

	var repoCache RepositoryCache
	if err := json.Unmarshal(data, &repoCache); err != nil {
		// Cache corrupted, treat as miss
		m.stats.RepoMisses++
		return nil, nil, time.Time{}, false, "", nil
	}

	m.stats.RepoHits++
	
	// Check if expired
	isExpired := repoCache.IsExpired()
	
	// Return the repository and commands data as JSON
	repoData, err := json.Marshal(repoCache.Repository)
	if err != nil {
		return nil, nil, time.Time{}, false, "", fmt.Errorf("failed to marshal repository data: %w", err)
	}
	
	commandsData, err := json.Marshal(repoCache.Commands)
	if err != nil {
		return nil, nil, time.Time{}, false, "", fmt.Errorf("failed to marshal commands data: %w", err)
	}
	
	return repoData, commandsData, repoCache.CachedAt, isExpired, repoCache.ETag, nil
}

// SetRepositoryCache stores repository data in cache
func (m *Manager) SetRepositoryCache(repoKey string, repo remote.RemoteRepository, commands []remote.RemoteCommand, etag string) error {
	if !m.IsEnabled() {
		return nil // Silently skip if disabled
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	
	// Calculate size of cached data
	size := int64(0)
	for _, cmd := range commands {
		size += int64(len(cmd.Content))
	}

	repoCache := RepositoryCache{
		Repository:  repo,
		Commands:    commands,
		CachedAt:    now,
		ExpiresAt:   now.Add(time.Duration(m.config.TTLHours) * time.Hour),
		ETag:        etag,
		LastChecked: now,
		Size:        size,
	}

	data, err := json.MarshalIndent(repoCache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal repository cache: %w", err)
	}

	repoPath := filepath.Join(m.cacheDir, "repositories", m.sanitizeRepoKey(repoKey)+".json")
	if err := os.WriteFile(repoPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write repository cache: %w", err)
	}

	return nil
}

// GetRepositoryKey generates a cache key for a repository
func (m *Manager) GetRepositoryKey(owner, repo, branch, path string) string {
	return fmt.Sprintf("%s_%s_%s_%s", owner, repo, branch, strings.ReplaceAll(path, "/", "_"))
}

// sanitizeRepoKey ensures the key is safe for filesystem use
func (m *Manager) sanitizeRepoKey(key string) string {
	// Replace unsafe characters and limit length
	key = strings.ReplaceAll(key, "/", "_")
	key = strings.ReplaceAll(key, "\\", "_")
	key = strings.ReplaceAll(key, ":", "_")
	key = strings.ReplaceAll(key, "*", "_")
	key = strings.ReplaceAll(key, "?", "_")
	key = strings.ReplaceAll(key, "\"", "_")
	key = strings.ReplaceAll(key, "<", "_")
	key = strings.ReplaceAll(key, ">", "_")
	key = strings.ReplaceAll(key, "|", "_")
	
	// If key is too long, use MD5 hash
	if len(key) > 200 {
		hash := md5.Sum([]byte(key))
		key = fmt.Sprintf("%x", hash)
	}
	
	return key
}

// BackgroundRefresh starts a background routine to refresh cached data
func (m *Manager) BackgroundRefresh(ctx context.Context, registryManager *remote.RegistryManager, githubClient *remote.GitHubClient) {
	if !m.IsEnabled() || !m.config.BackgroundRefresh {
		return
	}

	go func() {
		// Initial refresh after a short delay
		time.Sleep(2 * time.Second)
		m.refreshAll(registryManager, githubClient)

		// Set up periodic refresh
		ticker := time.NewTicker(time.Duration(m.config.TTLHours/2) * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.refreshAll(registryManager, githubClient)
			}
		}
	}()
}

// refreshAll refreshes all cached data in background
func (m *Manager) refreshAll(registryManager *remote.RegistryManager, githubClient *remote.GitHubClient) {
	startTime := time.Now()

	// Refresh registry first
	if err := m.refreshRegistry(registryManager); err != nil {
		// Log error but continue
		fmt.Printf("Background cache refresh - registry error: %v\n", err)
	}

	// Refresh repositories concurrently
	if err := m.refreshRepositories(githubClient); err != nil {
		// Log error but continue
		fmt.Printf("Background cache refresh - repositories error: %v\n", err)
	}

	m.mu.Lock()
	m.stats.LastRefresh = time.Now()
	m.stats.RefreshDuration = time.Since(startTime)
	m.mu.Unlock()
}

// refreshRegistry refreshes the cached registry
func (m *Manager) refreshRegistry(registryManager *remote.RegistryManager) error {
	// Check if we have cached registry and if it needs refresh
	cachedRegistry, err := m.GetRegistryCache()
	if err != nil {
		return err
	}

	// If cache is fresh, skip refresh
	if cachedRegistry != nil && !cachedRegistry.ShouldRefresh() {
		return nil
	}

	// Load fresh registry
	if err := registryManager.LoadRegistry(); err != nil {
		return fmt.Errorf("failed to load fresh registry: %w", err)
	}

	registry := registryManager.GetRegistry()
	if registry != nil {
		// Cache the fresh registry
		if err := m.SetRegistryCache(*registry, ""); err != nil {
			return fmt.Errorf("failed to cache registry: %w", err)
		}
	}

	return nil
}

// refreshRepositories refreshes cached repositories
func (m *Manager) refreshRepositories(githubClient *remote.GitHubClient) error {
	// For now, we'll implement a simple refresh
	// In a full implementation, this would:
	// 1. List all cached repositories
	// 2. Check which ones need refresh based on TTL
	// 3. Refresh them concurrently with worker pool
	// 4. Use ETags to minimize API calls
	
	return nil // Placeholder
}

// loadMetadata loads cache metadata from disk
func (m *Manager) loadMetadata() error {
	metadataPath := filepath.Join(m.cacheDir, "metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return err
	}

	var metadata CacheMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return err
	}

	return nil
}

// saveMetadata saves cache metadata to disk
func (m *Manager) saveMetadata(metadata CacheMetadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	metadataPath := filepath.Join(m.cacheDir, "metadata.json")
	return os.WriteFile(metadataPath, data, 0644)
}

// GetStats returns current cache statistics
func (m *Manager) GetStats() CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats := m.stats
	if stats.RegistryHits+stats.RegistryMisses > 0 {
		stats.HitRate = float64(stats.RegistryHits+stats.RepoHits) / float64(stats.RegistryHits+stats.RegistryMisses+stats.RepoHits+stats.RepoMisses)
	}
	
	return stats
}

// Clear removes all cached data
func (m *Manager) Clear() error {
	if !m.IsEnabled() {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return os.RemoveAll(m.cacheDir)
}