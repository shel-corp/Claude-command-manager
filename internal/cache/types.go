package cache

import (
	"time"

	"github.com/shel-corp/Claude-command-manager/internal/remote"
)

// CacheConfig holds configuration for the cache system
type CacheConfig struct {
	Enabled           bool   `json:"enabled"`
	Directory         string `json:"directory"`
	TTLHours          int    `json:"ttl_hours"`
	MaxSizeMB         int    `json:"max_size_mb"`
	BackgroundRefresh bool   `json:"background_refresh"`
	ConcurrentWorkers int    `json:"concurrent_workers"`
}

// DefaultCacheConfig returns the default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:           true,
		Directory:         "", // Will be set to ~/.config/claude_command_manager/cache
		TTLHours:          24,
		MaxSizeMB:         100,
		BackgroundRefresh: true,
		ConcurrentWorkers: 3,
	}
}

// CacheMetadata holds global cache metadata
type CacheMetadata struct {
	Version     string    `json:"version"`
	LastRefresh time.Time `json:"last_refresh"`
	TotalSize   int64     `json:"total_size_bytes"`
	ItemCount   int       `json:"item_count"`
}

// RegistryCache holds cached registry data
type RegistryCache struct {
	Registry    remote.RepositoryRegistry `json:"registry"`
	CachedAt    time.Time                 `json:"cached_at"`
	ExpiresAt   time.Time                 `json:"expires_at"`
	ETag        string                    `json:"etag,omitempty"`
	LastChecked time.Time                 `json:"last_checked"`
}

// RepositoryCache holds cached repository data with all commands
type RepositoryCache struct {
	Repository  remote.RemoteRepository `json:"repository"`
	Commands    []remote.RemoteCommand  `json:"commands"`
	CachedAt    time.Time               `json:"cached_at"`
	ExpiresAt   time.Time               `json:"expires_at"`
	ETag        string                  `json:"etag,omitempty"`
	LastChecked time.Time               `json:"last_checked"`
	Size        int64                   `json:"size_bytes"`
}

// CacheEntry represents a generic cache entry with metadata
type CacheEntry struct {
	Key       string    `json:"key"`
	Data      []byte    `json:"data"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
	ETag      string    `json:"etag,omitempty"`
	Size      int64     `json:"size_bytes"`
}

// CacheStats provides statistics about cache usage
type CacheStats struct {
	TotalEntries   int           `json:"total_entries"`
	TotalSize      int64         `json:"total_size_bytes"`
	HitRate        float64       `json:"hit_rate"`
	RegistryHits   int           `json:"registry_hits"`
	RegistryMisses int           `json:"registry_misses"`
	RepoHits       int           `json:"repo_hits"`
	RepoMisses     int           `json:"repo_misses"`
	LastRefresh    time.Time     `json:"last_refresh"`
	RefreshDuration time.Duration `json:"refresh_duration"`
}

// IsExpired checks if a cache entry has expired
func (rc *RegistryCache) IsExpired() bool {
	return time.Now().After(rc.ExpiresAt)
}

// IsExpired checks if a repository cache entry has expired
func (rc *RepositoryCache) IsExpired() bool {
	return time.Now().After(rc.ExpiresAt)
}

// ShouldRefresh checks if a cache entry should be refreshed (but may still be usable)
func (rc *RegistryCache) ShouldRefresh() bool {
	// Refresh if expired or if it's been more than half the TTL since last check
	halfTTL := rc.ExpiresAt.Sub(rc.CachedAt) / 2
	return rc.IsExpired() || time.Since(rc.LastChecked) > halfTTL
}

// ShouldRefresh checks if a repository cache should be refreshed
func (rc *RepositoryCache) ShouldRefresh() bool {
	halfTTL := rc.ExpiresAt.Sub(rc.CachedAt) / 2
	return rc.IsExpired() || time.Since(rc.LastChecked) > halfTTL
}