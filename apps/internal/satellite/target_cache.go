package satellite

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// TargetCache stores resolved target lookups so repeat CLI searches can avoid
// another CelesTrak name query.
type TargetCache struct {
	path    string
	Targets map[string]ResolvedTarget `json:"targets"`
}

// NewTargetCache returns an empty cache backed by path. When path is empty, the
// cache stays in memory only.
func NewTargetCache(path string) *TargetCache {
	return &TargetCache{
		path:    path,
		Targets: map[string]ResolvedTarget{},
	}
}

// DefaultTargetCachePath returns the standard on-disk cache file location.
func DefaultTargetCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, "arso", "targets.json"), nil
}

// LoadDefaultTargetCache loads the cache from the default cache path.
func LoadDefaultTargetCache() (*TargetCache, error) {
	path, err := DefaultTargetCachePath()
	if err != nil {
		return nil, err
	}

	return LoadTargetCache(path)
}

// LoadTargetCache loads target cache data from path. Missing files return an
// empty cache instead of an error.
func LoadTargetCache(path string) (*TargetCache, error) {
	targetCache := NewTargetCache(path)

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return targetCache, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, targetCache); err != nil {
		return nil, err
	}

	if targetCache.Targets == nil {
		targetCache.Targets = map[string]ResolvedTarget{}
	}

	return targetCache, nil
}

// GetTarget returns a cached target when it exists and has not expired.
func (c *TargetCache) GetTarget(key string) (ResolvedTarget, bool) {
	target, ok := c.Targets[key]
	if !ok {
		return ResolvedTarget{}, false
	}

	if !target.ExpiresAt.IsZero() && time.Now().UTC().After(target.ExpiresAt) {
		delete(c.Targets, key)
		_ = c.Save()
		return ResolvedTarget{}, false
	}

	return target, true
}

// SetTarget stores target under key and refreshes its resolution timestamps.
func (c *TargetCache) SetTarget(key string, target ResolvedTarget) error {
	now := time.Now().UTC()

	target.Query = key
	target.Source = "celestrak"
	target.ResolvedAt = now
	target.ExpiresAt = now.Add(30 * 24 * time.Hour)

	c.Targets[key] = target

	return c.Save()
}

// Save writes the cache to disk when it has a backing path.
func (c *TargetCache) Save() error {
	if c.path == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(c.path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.path, data, 0644)
}
