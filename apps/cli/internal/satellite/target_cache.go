package satellite

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type TargetCache struct {
	path    string
	Targets map[string]ResolvedTarget `json:"targets"`
}

func NewTargetCache(path string) *TargetCache {
	return &TargetCache{
		path:    path,
		Targets: map[string]ResolvedTarget{},
	}
}

func DefaultTargetCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, "arso", "targets.json"), nil
}

func LoadDefaultTargetCache() (*TargetCache, error) {
	path, err := DefaultTargetCachePath()
	if err != nil {
		return nil, err
	}

	return LoadTargetCache(path)
}

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

func (c *TargetCache) SetTarget(key string, target ResolvedTarget) error {
	now := time.Now().UTC()

	target.Query = key
	target.Source = "celestrak"
	target.ResolvedAt = now
	target.ExpiresAt = now.Add(30 * 24 * time.Hour)

	c.Targets[key] = target

	return c.Save()
}

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
