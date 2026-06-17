package satellite

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultTargetCachePathHasExpectedSuffix(t *testing.T) {
	path, err := DefaultTargetCachePath()
	if err != nil {
		t.Fatalf("DefaultTargetCachePath() unexpected error: %v", err)
	}

	if !strings.HasSuffix(path, filepath.Join("arso", "targets.json")) {
		t.Fatalf("DefaultTargetCachePath() = %q, want suffix %q", path, filepath.Join("arso", "targets.json"))
	}
}

func TestLoadTargetCacheMissingReturnsEmptyCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), "targets.json")

	cache, err := LoadTargetCache(path)
	if err != nil {
		t.Fatalf("LoadTargetCache() unexpected error: %v", err)
	}

	if got, want := len(cache.Targets), 0; got != want {
		t.Fatalf("len(cache.Targets) = %d, want %d", got, want)
	}
}

func TestTargetCacheSetAndGetRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "targets.json")
	cache := NewTargetCache(path)

	if err := cache.SetTarget("ISS", ResolvedTarget{Name: "ISS (ZARYA)", NoradID: 25544}); err != nil {
		t.Fatalf("SetTarget() unexpected error: %v", err)
	}

	reloaded, err := LoadTargetCache(path)
	if err != nil {
		t.Fatalf("LoadTargetCache() unexpected error: %v", err)
	}

	got, ok := reloaded.GetTarget("ISS")
	if !ok {
		t.Fatal("GetTarget() = false, want true")
	}
	if got.Query != "ISS" || got.Source != "celestrak" {
		t.Fatalf("cached target = %#v, want query and source metadata", got)
	}
	if got.ResolvedAt.IsZero() || got.ExpiresAt.IsZero() || !got.ExpiresAt.After(got.ResolvedAt) {
		t.Fatalf("cached target timestamps = %#v, want non-zero expiry after resolution", got)
	}
}

func TestExpiredTargetIsEvicted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "targets.json")
	cache := NewTargetCache(path)
	cache.Targets["ISS"] = ResolvedTarget{
		Name:      "ISS (ZARYA)",
		NoradID:   25544,
		ExpiresAt: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	if err := cache.Save(); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	if _, ok := cache.GetTarget("ISS"); ok {
		t.Fatal("GetTarget() = true, want false for expired entry")
	}

	reloaded, err := LoadTargetCache(path)
	if err != nil {
		t.Fatalf("LoadTargetCache() unexpected error: %v", err)
	}
	if _, ok := reloaded.Targets["ISS"]; ok {
		t.Fatal("expired cache entry still present after eviction")
	}
}
