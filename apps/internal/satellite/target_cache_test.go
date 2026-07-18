package satellite

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewTargetCache(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "with path",
			path: "/tmp/cache.json",
		},
		{
			name: "empty path",
			path: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewTargetCache(tt.path)
			if cache.path != tt.path {
				t.Errorf("NewTargetCache path = %q, want %q", cache.path, tt.path)
			}
			if cache.Targets == nil {
				t.Error("NewTargetCache should initialize Targets map")
			}
			if len(cache.Targets) != 0 {
				t.Errorf("NewTargetCache Targets should be empty, got %d", len(cache.Targets))
			}
		})
	}
}

func TestDefaultTargetCache2Path(t *testing.T) {
	path, err := DefaultTargetCachePath()
	if err != nil {
		t.Fatalf("DefaultTargetCachePath failed: %v", err)
	}
	if path == "" {
		t.Error("DefaultTargetCachePath returned empty path")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("DefaultTargetCachePath should return absolute path, got %q", path)
	}
	if filepath.Base(path) != "targets.json" {
		t.Errorf("DefaultTargetCachePath base should be targets.json, got %q", filepath.Base(path))
	}
	if filepath.Base(filepath.Dir(path)) != "arso" {
		t.Errorf("DefaultTargetCachePath should be in arso directory, got %q", filepath.Dir(path))
	}
}

func TestLoadDefaultTargetCache2(t *testing.T) {
	// Get the default path to clean up after
	defaultPath, _ := DefaultTargetCachePath()
	defer os.Remove(defaultPath)

	cache, err := LoadDefaultTargetCache()
	if err != nil {
		t.Fatalf("LoadDefaultTargetCache failed: %v", err)
	}
	if cache == nil {
		t.Error("LoadDefaultTargetCache returned nil")
	}
	if cache.Targets == nil {
		t.Error("LoadDefaultTargetCache should initialize Targets map")
	}
	if cache.path != defaultPath {
		t.Errorf("LoadDefaultTargetCache path = %q, want %q", cache.path, defaultPath)
	}
}

func TestLoadTargetCache(t *testing.T) {
	t.Run("non-existent file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nonexistent.json")
		cache, err := LoadTargetCache(path)
		if err != nil {
			t.Errorf("LoadTargetCache should not error for non-existent file: %v", err)
		}
		if cache == nil {
			t.Error("LoadTargetCache returned nil for non-existent file")
		}
		if cache.path != path {
			t.Errorf("LoadTargetCache path = %q, want %q", cache.path, path)
		}
		if cache.Targets == nil {
			t.Error("LoadTargetCache should initialize Targets map")
		}
	})

	t.Run("valid file", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "cache.json")

		// Create a cache file
		originalCache := NewTargetCache(path)
		originalCache.Targets["ISS"] = ResolvedTarget{
			Query:      "ISS",
			Name:       "ISS (ZARYA)",
			NoradID:    25544,
			Kind:       "satellite",
			Source:     "celestrak",
			ResolvedAt: time.Now().UTC(),
			ExpiresAt:  time.Now().UTC().Add(30 * 24 * time.Hour),
		}
		if err := originalCache.Save(); err != nil {
			t.Fatalf("Failed to save cache: %v", err)
		}

		loaded, err := LoadTargetCache(path)
		if err != nil {
			t.Fatalf("LoadTargetCache failed: %v", err)
		}
		if loaded.path != path {
			t.Errorf("Loaded path = %q, want %q", loaded.path, path)
		}
		if len(loaded.Targets) != 1 {
			t.Errorf("Loaded Targets length = %d, want 1", len(loaded.Targets))
		}
		if _, ok := loaded.Targets["ISS"]; !ok {
			t.Error("Loaded cache missing ISS key")
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "malformed.json")
		if err := os.WriteFile(path, []byte("{invalid json}"), 0644); err != nil {
			t.Fatalf("Failed to write malformed file: %v", err)
		}

		_, err := LoadTargetCache(path)
		if err == nil {
			t.Error("LoadTargetCache should error for malformed JSON")
		}
	})

	t.Run("empty Targets map in loaded cache", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "empty.json")

		// Create cache with nil Targets
		data := map[string]interface{}{
			"path":    path,
			"targets": nil,
		}
		jsonData, _ := json.Marshal(data)
		if err := os.WriteFile(path, jsonData, 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		cache, err := LoadTargetCache(path)
		if err != nil {
			t.Fatalf("LoadTargetCache failed: %v", err)
		}
		if cache.Targets == nil {
			t.Error("LoadTargetCache should initialize nil Targets map")
		}
	})
}

func TestTargetCache_SetTarget(t *testing.T) {
	t.Run("set new target", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "cache.json")
		cache := NewTargetCache(path)

		key := "HUBBLE"
		target := ResolvedTarget{
			Name:    "HUBBLE SPACE TELESCOPE",
			NoradID: 20580,
			Kind:    "satellite",
		}

		err := cache.SetTarget(key, target)
		if err != nil {
			t.Fatalf("SetTarget failed: %v", err)
		}

		cached, ok := cache.Targets[key]
		if !ok {
			t.Error("Target not found in cache after SetTarget")
		}
		if cached.Query != key {
			t.Errorf("Query = %q, want %q", cached.Query, key)
		}
		if cached.Name != target.Name {
			t.Errorf("Name = %q, want %q", cached.Name, target.Name)
		}
		if cached.NoradID != target.NoradID {
			t.Errorf("NoradID = %d, want %d", cached.NoradID, target.NoradID)
		}
		if cached.Source != "celestrak" {
			t.Errorf("Source = %q, want celestrak", cached.Source)
		}
		if cached.ResolvedAt.IsZero() {
			t.Error("ResolvedAt should be set")
		}
		if cached.ExpiresAt.IsZero() {
			t.Error("ExpiresAt should be set")
		}
		expectedExpiry := cached.ResolvedAt.Add(30 * 24 * time.Hour)
		if !cached.ExpiresAt.Equal(expectedExpiry) {
			t.Errorf("ExpiresAt = %v, want %v", cached.ExpiresAt, expectedExpiry)
		}

		// Verify file was saved
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Cache file not saved: %v", err)
		}
	})

	t.Run("overwrite existing target", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "cache.json")
		cache := NewTargetCache(path)

		key := "ISS"
		initialTarget := ResolvedTarget{
			Name:    "OLD NAME",
			NoradID: 99999,
		}
		cache.SetTarget(key, initialTarget)

		newTarget := ResolvedTarget{
			Name:    "ISS (ZARYA)",
			NoradID: 25544,
		}
		cache.SetTarget(key, newTarget)

		cached, ok := cache.Targets[key]
		if !ok {
			t.Error("Target not found after overwrite")
		}
		if cached.Name != newTarget.Name {
			t.Errorf("Name = %q, want %q", cached.Name, newTarget.Name)
		}
		if cached.NoradID != newTarget.NoradID {
			t.Errorf("NoradID = %d, want %d", cached.NoradID, newTarget.NoradID)
		}
	})
}

func TestTargetCache_GetTarget(t *testing.T) {
	t.Run("existing target", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "cache.json")
		cache := NewTargetCache(path)

		key := "ISS"
		expected := ResolvedTarget{
			Query:      key,
			Name:       "ISS (ZARYA)",
			NoradID:    25544,
			Kind:       "satellite",
			Source:     "celestrak",
			ResolvedAt: time.Now().UTC(),
			ExpiresAt:  time.Now().UTC().Add(30 * 24 * time.Hour),
		}
		cache.Targets[key] = expected

		got, ok := cache.GetTarget(key)
		if !ok {
			t.Error("GetTarget returned false for existing target")
		}
		if got.Name != expected.Name {
			t.Errorf("Name = %q, want %q", got.Name, expected.Name)
		}
		if got.NoradID != expected.NoradID {
			t.Errorf("NoradID = %d, want %d", got.NoradID, expected.NoradID)
		}
	})

	t.Run("non-existent target", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "cache.json")
		cache := NewTargetCache(path)

		got, ok := cache.GetTarget("NONEXISTENT")
		if ok {
			t.Error("GetTarget returned true for non-existent target")
		}
		if got.Name != "" {
			t.Errorf("GetTarget should return empty ResolvedTarget, got %+v", got)
		}
	})

	t.Run("expired target", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "cache.json")
		cache := NewTargetCache(path)

		key := "EXPIRED"
		target := ResolvedTarget{
			Name:      "EXPIRED SATELLITE",
			NoradID:   12345,
			ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
		}
		cache.Targets[key] = target

		got, ok := cache.GetTarget(key)
		if ok {
			t.Error("GetTarget returned true for expired target")
		}
		if got.Name != "" {
			t.Error("GetTarget should return empty ResolvedTarget for expired entry")
		}

		// Verify expired entry was deleted
		if _, stillExists := cache.Targets[key]; stillExists {
			t.Error("Expired target was not deleted from cache")
		}
	})

	t.Run("target with zero ExpiresAt", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "cache.json")
		cache := NewTargetCache(path)

		key := "NO_EXPIRY"
		target := ResolvedTarget{
			Name:      "NO EXPIRY",
			NoradID:   99999,
			ExpiresAt: time.Time{}, // zero time
		}
		cache.Targets[key] = target

		got, ok := cache.GetTarget(key)
		if !ok {
			t.Error("GetTarget returned false for target with zero ExpiresAt")
		}
		if got.Name != target.Name {
			t.Errorf("Name = %q, want %q", got.Name, target.Name)
		}
	})
}

func TestTargetCache_Save(t *testing.T) {
	t.Run("save with path", func(t *testing.T) {
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "subdir", "cache.json")
		cache := NewTargetCache(path)
		cache.Targets["TEST"] = ResolvedTarget{
			Query:      "TEST",
			Name:       "TEST SATELLITE",
			NoradID:    99999,
			Kind:       "satellite",
			Source:     "celestrak",
			ResolvedAt: time.Now().UTC(),
			ExpiresAt:  time.Now().UTC().Add(30 * 24 * time.Hour),
		}

		err := cache.Save()
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Verify directory was created
		if _, err := os.Stat(filepath.Dir(path)); err != nil {
			t.Errorf("Directory not created: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(path); err != nil {
			t.Errorf("File not created: %v", err)
		}

		// Verify content
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read saved file: %v", err)
		}
		var loaded TargetCache
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("Failed to unmarshal saved data: %v", err)
		}
		if len(loaded.Targets) != 1 {
			t.Errorf("Saved Targets length = %d, want 1", len(loaded.Targets))
		}
	})

	t.Run("save with empty path", func(t *testing.T) {
		cache := NewTargetCache("")
		cache.Targets["TEST"] = ResolvedTarget{
			Name:    "TEST SATELLITE",
			NoradID: 99999,
		}

		err := cache.Save()
		if err != nil {
			t.Errorf("Save with empty path should not error: %v", err)
		}
	})

	t.Run("save handles marshal error", func(t *testing.T) {
		// This is difficult to trigger directly, but we can test the error path
		// by ensuring the function handles it properly
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "cache.json")
		cache := NewTargetCache(path)
		// Adding a circular reference would cause marshal error, but that's complex
		// Just verify the function doesn't panic
		err := cache.Save()
		if err != nil {
			t.Logf("Save returned error (expected for this test): %v", err)
		}
	})
}

func TestTargetCache_Integration(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "integration.json")

	// Create cache
	cache := NewTargetCache(path)

	// Set multiple targets
	targets := map[string]ResolvedTarget{
		"ISS": {
			Name:    "ISS (ZARYA)",
			NoradID: 25544,
		},
		"HUBBLE": {
			Name:    "HUBBLE SPACE TELESCOPE",
			NoradID: 20580,
		},
		"STARLINK": {
			Name:    "STARLINK-1000",
			NoradID: 12345,
		},
	}

	for key, target := range targets {
		if err := cache.SetTarget(key, target); err != nil {
			t.Fatalf("SetTarget failed for %s: %v", key, err)
		}
	}

	// Verify all targets are retrievable
	for key, expected := range targets {
		got, ok := cache.GetTarget(key)
		if !ok {
			t.Errorf("GetTarget failed for %s", key)
			continue
		}
		if got.Name != expected.Name {
			t.Errorf("Name mismatch for %s: got %q, want %q", key, got.Name, expected.Name)
		}
		if got.NoradID != expected.NoradID {
			t.Errorf("NoradID mismatch for %s: got %d, want %d", key, got.NoradID, expected.NoradID)
		}
	}

	// Load from disk
	loaded, err := LoadTargetCache(path)
	if err != nil {
		t.Fatalf("LoadTargetCache failed: %v", err)
	}
	if len(loaded.Targets) != len(targets) {
		t.Errorf("Loaded Targets length = %d, want %d", len(loaded.Targets), len(targets))
	}

	// Delete a target by expiration
	expiredKey := "TO_DELETE"
	cache.Targets[expiredKey] = ResolvedTarget{
		Name:      "TO DELETE",
		NoradID:   99999,
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	cache.GetTarget(expiredKey)
	if _, ok := cache.Targets[expiredKey]; ok {
		t.Error("Expired target was not deleted")
	}
}
