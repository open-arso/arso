package satellite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"
)

func TestNormalizeTarget(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already uppercase",
			input:    "ISS",
			expected: "ISS",
		},
		{
			name:     "lowercase",
			input:    "iss",
			expected: "ISS",
		},
		{
			name:     "mixed case",
			input:    "IsS",
			expected: "ISS",
		},
		{
			name:     "with spaces",
			input:    "  ISS  ",
			expected: "ISS",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeTarget(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeTarget(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseNORADID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantID    int
		wantValid bool
	}{
		{
			name:      "valid NORAD ID",
			input:     "25544",
			wantID:    25544,
			wantValid: true,
		},
		{
			name:      "valid with spaces",
			input:     "  25544  ",
			wantID:    25544,
			wantValid: true,
		},
		{
			name:      "zero",
			input:     "0",
			wantID:    0,
			wantValid: false,
		},
		{
			name:      "negative",
			input:     "-1",
			wantID:    0,
			wantValid: false,
		},
		{
			name:      "non-numeric",
			input:     "ISS",
			wantID:    0,
			wantValid: false,
		},
		{
			name:      "empty",
			input:     "",
			wantID:    0,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, valid := parseNORADID(tt.input)
			if id != tt.wantID || valid != tt.wantValid {
				t.Errorf("parseNORADID(%q) = (%d, %v), want (%d, %v)",
					tt.input, id, valid, tt.wantID, tt.wantValid)
			}
		})
	}
}

func TestTargetCache(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache.json")

	t.Run("NewTargetCache", func(t *testing.T) {
		cache := NewTargetCache(cachePath)
		if cache.path != cachePath {
			t.Errorf("NewTargetCache path = %q, want %q", cache.path, cachePath)
		}
		if cache.Targets == nil {
			t.Error("NewTargetCache should initialize Targets map")
		}
	})

	t.Run("SetAndGetTarget", func(t *testing.T) {
		cache := NewTargetCache(cachePath)
		key := "ISS"
		target := ResolvedTarget{
			Name:    "ISS (ZARYA)",
			NoradID: 25544,
		}

		err := cache.SetTarget(key, target)
		if err != nil {
			t.Fatalf("SetTarget failed: %v", err)
		}

		got, ok := cache.GetTarget(key)
		if !ok {
			t.Error("GetTarget returned false for existing key")
		}

		if got.Name != target.Name {
			t.Errorf("GetTarget Name = %q, want %q", got.Name, target.Name)
		}
		if got.NoradID != target.NoradID {
			t.Errorf("GetTarget NoradID = %d, want %d", got.NoradID, target.NoradID)
		}
		if got.Query != key {
			t.Errorf("GetTarget Query = %q, want %q", got.Query, key)
		}
		if got.Source != "celestrak" {
			t.Errorf("GetTarget Source = %q, want %q", got.Source, "celestrak")
		}
		if got.ResolvedAt.IsZero() {
			t.Error("GetTarget ResolvedAt should be set")
		}
		if got.ExpiresAt.IsZero() {
			t.Error("GetTarget ExpiresAt should be set")
		}
	})

	t.Run("GetTargetExpired", func(t *testing.T) {
		cache := NewTargetCache(cachePath)
		key := "STARLINK"
		target := ResolvedTarget{
			Name:      "STARLINK-1000",
			NoradID:   12345,
			ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
		}
		cache.Targets[key] = target

		got, ok := cache.GetTarget(key)
		if ok {
			t.Error("GetTarget should return false for expired entry")
		}
		if got.Name != "" {
			t.Error("GetTarget should return empty ResolvedTarget for expired entry")
		}
	})

	t.Run("SaveAndLoad", func(t *testing.T) {
		cache := NewTargetCache(cachePath)
		key := "HUBBLE"
		target := ResolvedTarget{
			Name:    "HUBBLE SPACE TELESCOPE",
			NoradID: 20580,
		}

		err := cache.SetTarget(key, target)
		if err != nil {
			t.Fatalf("SetTarget failed: %v", err)
		}

		loaded, err := LoadTargetCache(cachePath)
		if err != nil {
			t.Fatalf("LoadTargetCache failed: %v", err)
		}

		got, ok := loaded.GetTarget(key)
		if !ok {
			t.Error("Loaded cache doesn't contain key")
		}
		if got.Name != target.Name {
			t.Errorf("Loaded Name = %q, want %q", got.Name, target.Name)
		}
		if got.NoradID != target.NoradID {
			t.Errorf("Loaded NoradID = %d, want %d", got.NoradID, target.NoradID)
		}
	})

	t.Run("LoadNonExistent", func(t *testing.T) {
		cache, err := LoadTargetCache("/non/existent/path.json")
		if err != nil {
			t.Errorf("LoadTargetCache should not error for non-existent file: %v", err)
		}
		if cache == nil {
			t.Error("LoadTargetCache should return cache for non-existent file")
		}
		if cache.Targets == nil {
			t.Error("LoadTargetCache should initialize Targets map")
		}
	})

	t.Run("SaveWithEmptyPath", func(t *testing.T) {
		cache := NewTargetCache("")
		err := cache.Save()
		if err != nil {
			t.Errorf("Save with empty path should not error: %v", err)
		}
	})
}

func TestDefaultTargetCachePath(t *testing.T) {
	path, err := DefaultTargetCachePath()
	if err != nil {
		t.Fatalf("DefaultTargetCachePath failed: %v", err)
	}
	if path == "" {
		t.Error("DefaultTargetCachePath returned empty path")
	}
	if !filepath.IsAbs(path) || filepath.Base(path) != "targets.json" {
		t.Errorf("DefaultTargetCachePath = %q, should be targets.json in cache dir", path)
	}
}

func TestLoadDefaultTargetCache(t *testing.T) {
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
}

func TestSplitLookahead(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantValue   int
		wantUnit    string
		expectError bool
	}{
		{
			name:      "hours",
			input:     "24h",
			wantValue: 24,
			wantUnit:  "h",
		},
		{
			name:      "days",
			input:     "7d",
			wantValue: 7,
			wantUnit:  "d",
		},
		{
			name:      "minutes",
			input:     "30m",
			wantValue: 30,
			wantUnit:  "m",
		},
		{
			name:      "seconds",
			input:     "45s",
			wantValue: 45,
			wantUnit:  "s",
		},
		{
			name:      "years",
			input:     "1Y",
			wantValue: 1,
			wantUnit:  "Y",
		},
		{
			name:      "months",
			input:     "6M",
			wantValue: 6,
			wantUnit:  "M",
		},
		{
			name:        "empty",
			input:       "",
			expectError: true,
		},
		{
			name:        "no number",
			input:       "h",
			expectError: true,
		},
		{
			name:        "no unit",
			input:       "24",
			expectError: true,
		},
		{
			name:        "negative",
			input:       "-24h",
			expectError: true,
		},
		{
			name:        "invalid unit",
			input:       "24x",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, unit, err := splitLookahead(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("splitLookahead(%q) should error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("splitLookahead(%q) unexpected error: %v", tt.input, err)
			}
			if value != tt.wantValue {
				t.Errorf("splitLookahead(%q) value = %d, want %d", tt.input, value, tt.wantValue)
			}
			if unit != tt.wantUnit {
				t.Errorf("splitLookahead(%q) unit = %q, want %q", tt.input, unit, tt.wantUnit)
			}
		})
	}
}

func TestComputeLookaheadTime(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		start       time.Time
		lookahead   string
		expected    time.Time
		expectError bool
	}{
		{
			name:      "add hours",
			start:     now,
			lookahead: "2h",
			expected:  now.Add(2 * time.Hour),
		},
		{
			name:      "add days",
			start:     now,
			lookahead: "3d",
			expected:  now.AddDate(0, 0, 3),
		},
		{
			name:      "add minutes",
			start:     now,
			lookahead: "15m",
			expected:  now.Add(15 * time.Minute),
		},
		{
			name:      "add seconds",
			start:     now,
			lookahead: "30s",
			expected:  now.Add(30 * time.Second),
		},
		{
			name:      "add years",
			start:     now,
			lookahead: "1Y",
			expected:  now.AddDate(1, 0, 0),
		},
		{
			name:      "add months",
			start:     now,
			lookahead: "2M",
			expected:  now.AddDate(0, 2, 0),
		},
		{
			name:        "invalid lookahead",
			start:       now,
			lookahead:   "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := computeLookaheadTime(tt.start, tt.lookahead)
			if tt.expectError {
				if err == nil {
					t.Errorf("computeLookaheadTime should error")
				}
				return
			}
			if err != nil {
				t.Fatalf("computeLookaheadTime unexpected error: %v", err)
			}
			if !result.Equal(tt.expected) {
				t.Errorf("computeLookaheadTime = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseTimeStr(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:  "empty returns now",
			input: "",
		},
		{
			name:  "valid RFC3339",
			input: "2026-07-15T12:00:00Z",
		},
		{
			name:        "invalid format",
			input:       "2026-07-15 12:00:00",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimeStr(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("parseTimeStr(%q) should error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseTimeStr(%q) unexpected error: %v", tt.input, err)
			}
			if tt.input == "" {
				// Just verify it returns a time within a reasonable range
				if result.Before(now.Add(-1*time.Second)) || result.After(now.Add(1*time.Second)) {
					t.Errorf("parseTimeStr empty should return current time")
				}
			} else {
				expected, _ := time.Parse(time.RFC3339, tt.input)
				if !result.Equal(expected) {
					t.Errorf("parseTimeStr(%q) = %v, want %v", tt.input, result, expected)
				}
			}
		})
	}
}

func TestFormatLookaheadDuration(t *testing.T) {
	tests := []struct {
		name        string
		duration    time.Duration
		expected    string
		expectError bool
	}{
		{
			name:     "hours",
			duration: 2 * time.Hour,
			expected: "2h",
		},
		{
			name:     "days",
			duration: 3 * 24 * time.Hour,
			expected: "3d",
		},
		{
			name:     "minutes",
			duration: 15 * time.Minute,
			expected: "15m",
		},
		{
			name:     "seconds",
			duration: 30 * time.Second,
			expected: "30s",
		},
		{
			name:        "zero",
			duration:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatLookaheadDuration(tt.duration)
			if tt.expectError {
				if err == nil {
					t.Errorf("formatLookaheadDuration should error for %v", tt.duration)
				}
				return
			}
			if err != nil {
				t.Fatalf("formatLookaheadDuration unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("formatLookaheadDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestClient_NewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.baseURL != DefaultBaseURL {
		t.Errorf("NewClient baseURL = %q, want %q", client.baseURL, DefaultBaseURL)
	}
	if client.httpClient == nil {
		t.Error("NewClient httpClient is nil")
	}
	if client.targetCache == nil {
		t.Error("NewClient targetCache is nil")
	}
}

func TestClient_buildURL(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name      string
		queryKey  string
		queryVal  string
		contains  string
		expectErr bool
	}{
		{
			name:     "valid URL",
			queryKey: "NAME",
			queryVal: "ISS",
			contains: "FORMAT=JSON",
		},
		{
			name:     "URL with special chars",
			queryKey: "SEARCH",
			queryVal: "space station",
			contains: "FORMAT=JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := client.buildURL(tt.queryKey, tt.queryVal)
			if tt.expectErr {
				if err == nil {
					t.Error("buildURL should error")
				}
				return
			}
			if err != nil {
				t.Fatalf("buildURL unexpected error: %v", err)
			}
			if url == "" {
				t.Error("buildURL returned empty URL")
			}
			if !filepath.IsAbs(url) {
				t.Logf("buildURL result: %s", url)
			}
		})
	}
}

func TestClient_CacheResolvedTarget(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name      string
		query     string
		resolved  ResolvedTarget
		expectErr bool
	}{
		{
			name:  "valid target",
			query: "ISS",
			resolved: ResolvedTarget{
				Name:    "ISS (ZARYA)",
				NoradID: 25544,
			},
		},
		{
			name:      "empty target",
			query:     "",
			resolved:  ResolvedTarget{},
			expectErr: true,
		},
		{
			name:  "target with spaces",
			query: "  ISS  ",
			resolved: ResolvedTarget{
				Name:    "ISS (ZARYA)",
				NoradID: 25544,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.CacheResolvedTarget(tt.query, tt.resolved)
			if tt.expectErr {
				if err == nil {
					t.Error("CacheResolvedTarget should error")
				}
				return
			}
			if err != nil {
				t.Errorf("CacheResolvedTarget unexpected error: %v", err)
			}

			// Verify it was cached
			normalized := normalizeTarget(tt.query)
			cached, ok := client.targetCache.GetTarget(normalized)
			if !ok {
				t.Error("Target not found in cache after CacheResolvedTarget")
			}
			if cached.Name != tt.resolved.Name {
				t.Errorf("Cached Name = %q, want %q", cached.Name, tt.resolved.Name)
			}
			if cached.NoradID != tt.resolved.NoradID {
				t.Errorf("Cached NoradID = %d, want %d", cached.NoradID, tt.resolved.NoradID)
			}
		})
	}
}

// MockClient for integration tests that require network access
type mockClient struct {
	*Client
	fetchRawFunc func(ctx context.Context, queryKey, queryValue string) ([]byte, error)
}

func (m *mockClient) fetchRaw(ctx context.Context, queryKey, queryValue string) ([]byte, error) {
	if m.fetchRawFunc != nil {
		return m.fetchRawFunc(ctx, queryKey, queryValue)
	}
	return m.Client.fetchRaw(ctx, queryKey, queryValue)
}

func TestClient_ResolveTarget(t *testing.T) {
	// This test demonstrates the logic but requires network access for full testing
	client := NewClient()
	ctx := context.Background()

	// Test with numeric NORAD ID
	t.Run("numeric NORAD ID", func(t *testing.T) {
		resolved, err := client.ResolveTarget(ctx, "25544")
		if err != nil {
			// Skip if network is not available
			t.Skip("Network not available or CelesTrak API error:", err)
		}
		if resolved.NoradID != 25544 {
			t.Errorf("ResolveTarget NoradID = %d, want 25544", resolved.NoradID)
		}
		if resolved.Kind != "satellite" {
			t.Errorf("ResolveTarget Kind = %q, want satellite", resolved.Kind)
		}
		if resolved.Source != "numeric" {
			t.Errorf("ResolveTarget Source = %q, want numeric", resolved.Source)
		}
	})

	// Test with empty target
	t.Run("empty target", func(t *testing.T) {
		_, err := client.ResolveTarget(ctx, "")
		if err == nil {
			t.Error("ResolveTarget with empty target should error")
		}
	})

	// Test with whitespace-only target
	t.Run("whitespace target", func(t *testing.T) {
		_, err := client.ResolveTarget(ctx, "   ")
		if err == nil {
			t.Error("ResolveTarget with whitespace target should error")
		}
	})
}

func TestAmbiguousTargetError(t *testing.T) {
	err := &AmbiguousTargetError{
		Target: "ISS",
		Candidates: []ResolvedTarget{
			{Name: "ISS (ZARYA)", NoradID: 25544},
			{Name: "ISS (DECO)", NoradID: 12345},
		},
	}

	expected := "target \"ISS\" is ambiguous: 2 satellites found"
	if err.Error() != expected {
		t.Errorf("AmbiguousTargetError.Error() = %q, want %q", err.Error(), expected)
	}
}

// Benchmark tests
func BenchmarkNormalizeTarget(b *testing.B) {
	for i := 0; i < b.N; i++ {
		normalizeTarget("  iSs  ")
	}
}

func BenchmarkParseNORADID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseNORADID("25544")
	}
}

func BenchmarkTargetCache_SetAndGet(b *testing.B) {
	cache := NewTargetCache("")
	target := ResolvedTarget{
		Name:    "ISS (ZARYA)",
		NoradID: 25544,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.SetTarget("ISS", target)
		_, _ = cache.GetTarget("ISS")
	}
}

func TestClient_ResolveTarget_FromCache(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Pre-populate cache
	normalized := "ISS"
	cachedTarget := ResolvedTarget{
		Query:   normalized,
		Name:    "ISS (ZARYA)",
		NoradID: 25544,
		Kind:    "satellite",
		Source:  "celestrak",
	}
	client.targetCache.SetTarget(normalized, cachedTarget)

	resolved, err := client.ResolveTarget(ctx, "ISS")
	if err != nil {
		t.Fatalf("ResolveTarget unexpected error: %v", err)
	}

	if resolved.Source != "cache" {
		t.Errorf("ResolveTarget Source = %q, want cache", resolved.Source)
	}
	if resolved.Name != cachedTarget.Name {
		t.Errorf("ResolveTarget Name = %q, want %q", resolved.Name, cachedTarget.Name)
	}
	if resolved.NoradID != cachedTarget.NoradID {
		t.Errorf("ResolveTarget NoradID = %d, want %d", resolved.NoradID, cachedTarget.NoradID)
	}
}

func TestClient_ResolveTarget_CacheMiss(t *testing.T) {
	// This requires mocking fetchRaw to avoid network calls
	mockElements := []GPElement{
		{
			ObjectName: "TEST SATELLITE UNIQUE",
			ObjectID:   "2026-001A",
			NoradCatID: 99999,
		},
	}
	mockData, _ := json.Marshal(mockElements)

	testClient := &Client{
		baseURL:     DefaultBaseURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		targetCache: NewTargetCache(""),
		fetchRaw: func(ctx context.Context, queryKey, queryValue string) ([]byte, error) {
			if queryValue == "TEST" {
				return mockData, nil
			}

			return nil, fmt.Errorf("not found")
		},
	}
	testClient.targetCache.Targets = make(map[string]ResolvedTarget)

	ctx := context.Background()
	_, err := testClient.ResolveTarget(ctx, "TEST")
	if err != nil {
		t.Fatalf("ResolveTarget unexpected error: %v", err)
	}

	// Test with a target that's not in cache
	// This will try to hit CelesTrak - we'll skip if network not available
	_, err = testClient.ResolveTarget(ctx, "NONEXISTENT_12345")
	if err == nil {
		t.Error("ResolveTarget with nonexistent target should error")
	}
}

func TestClient_ResolveTarget_CacheSetOnMiss(t *testing.T) {
	mockElements := []GPElement{
		{
			ObjectName: "TEST SATELLITE",
			ObjectID:   "2026-001A",
			NoradCatID: 99999,
		},
	}
	mockData, _ := json.Marshal(mockElements)

	// Create test client with mock
	testClient := &Client{
		baseURL:     DefaultBaseURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		targetCache: NewTargetCache(""),
		fetchRaw: func(ctx context.Context, queryKey, queryValue string) ([]byte, error) {
			return mockData, nil
		},
	}
	testClient.targetCache.Targets = make(map[string]ResolvedTarget)

	ctx := context.Background()
	resolved, err := testClient.ResolveTarget(ctx, "TEST")
	if err != nil {
		t.Fatalf("ResolveTarget unexpected error: %v", err)
	}

	// Verify it was cached
	cached, ok := testClient.targetCache.GetTarget("TEST")
	if !ok {
		t.Error("Target should be cached after resolve")
	}
	if cached.Name != resolved.Name {
		t.Errorf("Cached Name = %q, want %q", cached.Name, resolved.Name)
	}
	if cached.NoradID != resolved.NoradID {
		t.Errorf("Cached NoradID = %d, want %d", cached.NoradID, resolved.NoradID)
	}
}

func TestClient_ResolveTargetFromCelesTrak_EmptyResponse(t *testing.T) {
	testClient := &Client{
		baseURL:     DefaultBaseURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		targetCache: NewTargetCache(""),
		fetchRaw: func(ctx context.Context, queryKey, queryValue string) ([]byte, error) {
			return []byte("[]"), nil
		},
	}

	ctx := context.Background()
	_, err := testClient.resolveTargetFromCelesTrak(ctx, "NONEXISTENT_UNIQUE_12345")
	if err == nil {
		t.Error("resolveTargetFromCelesTrak should error for empty response")
	}
	expected := `no satellite found for "NONEXISTENT_UNIQUE_12345"`
	if err.Error() != expected {
		t.Errorf("Error = %q, want %q", err.Error(), expected)
	}
}

func TestClient_ResolveTargetFromCelesTrak_Ambiguous(t *testing.T) {
	mockElements := []GPElement{
		{
			ObjectName: "SATELLITE A",
			ObjectID:   "2026-001A",
			NoradCatID: 11111,
		},
		{
			ObjectName: "SATELLITE B",
			ObjectID:   "2026-001B",
			NoradCatID: 22222,
		},
	}
	mockData, _ := json.Marshal(mockElements)

	testClient := &Client{
		baseURL:     DefaultBaseURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		targetCache: NewTargetCache(""),
		fetchRaw: func(ctx context.Context, queryKey, queryValue string) ([]byte, error) {
			return mockData, nil
		},
	}

	ctx := context.Background()
	_, err := testClient.resolveTargetFromCelesTrak(ctx, "AMBIGUOUS")
	if err == nil {
		t.Error("resolveTargetFromCelesTrak should error for ambiguous targets")
	}

	ambiguousErr, ok := err.(*AmbiguousTargetError)
	if !ok {
		t.Fatalf("Error should be AmbiguousTargetError, got %T", err)
	}
	if len(ambiguousErr.Candidates) != 2 {
		t.Errorf("AmbiguousTargetError Candidates length = %d, want 2", len(ambiguousErr.Candidates))
	}
}

func TestClient_ResolveTargetFromCelesTrak_Success(t *testing.T) {
	mockElements := []GPElement{
		{
			ObjectName: "ISS (ZARYA)",
			ObjectID:   "1998-067A",
			NoradCatID: 25544,
		},
	}
	mockData, _ := json.Marshal(mockElements)

	testClient := &Client{
		baseURL:     DefaultBaseURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		targetCache: NewTargetCache(""),
		fetchRaw: func(ctx context.Context, queryKey, queryValue string) ([]byte, error) {
			return mockData, nil
		},
	}

	ctx := context.Background()
	resolved, err := testClient.resolveTargetFromCelesTrak(ctx, "ISS")
	if err != nil {
		t.Fatalf("resolveTargetFromCelesTrak unexpected error: %v", err)
	}

	if resolved.Name != "ISS (ZARYA)" {
		t.Errorf("Name = %q, want ISS (ZARYA)", resolved.Name)
	}
	if resolved.NoradID != 25544 {
		t.Errorf("NoradID = %d, want 25544", resolved.NoradID)
	}
	if resolved.Source != "celestrak" {
		t.Errorf("Source = %q, want celestrak", resolved.Source)
	}
	if resolved.ResolvedAt.IsZero() {
		t.Error("ResolvedAt should be set")
	}
	if resolved.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should be set")
	}
	expectedExpiry := resolved.ResolvedAt.Add(30 * 24 * time.Hour)
	if !resolved.ExpiresAt.Equal(expectedExpiry) {
		t.Errorf("ExpiresAt = %v, want %v", resolved.ExpiresAt, expectedExpiry)
	}
}
