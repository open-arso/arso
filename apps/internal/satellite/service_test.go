package satellite

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/openarso/arso/apps/internal/config"
)

func TestNewClient(t *testing.T) {
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
	if client.fetchRaw == nil {
		t.Error("NewClient fetchRaw is nil")
	}
}

func TestClient_buildURL2(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name      string
		queryKey  string
		queryVal  string
		expectErr bool
		contains  []string
	}{
		{
			name:     "CATNR query",
			queryKey: QueryCATNR,
			queryVal: "25544",
			contains: []string{"CATNR=25544", "FORMAT=JSON"},
		},
		{
			name:     "NAME query",
			queryKey: QueryNAME,
			queryVal: "ISS",
			contains: []string{"NAME=ISS", "FORMAT=JSON"},
		},
		{
			name:     "query with spaces",
			queryKey: QueryNAME,
			queryVal: "HUBBLE SPACE",
			contains: []string{"NAME=HUBBLE+SPACE", "FORMAT=JSON"},
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
			for _, expected := range tt.contains {
				if !strings.Contains(url, expected) {
					t.Errorf("URL %q should contain %q", url, expected)
				}
			}
		})
	}
}

func TestClient_fetchRawHTTP(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"test": "data"}`))
		}))
		defer server.Close()

		client := NewClient()
		client.baseURL = server.URL

		body, err := client.fetchRawHTTP(context.Background(), "TEST", "value")
		if err != nil {
			t.Fatalf("fetchRawHTTP unexpected error: %v", err)
		}
		expected := `{"test": "data"}`
		if string(body) != expected {
			t.Errorf("fetchRawHTTP body = %q, want %q", string(body), expected)
		}
	})

	t.Run("non-200 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := NewClient()
		client.baseURL = server.URL

		_, err := client.fetchRawHTTP(context.Background(), "TEST", "value")
		if err == nil {
			t.Error("fetchRawHTTP should error for non-200 status")
		}
		if !strings.Contains(err.Error(), "CelesTrak returned status") {
			t.Errorf("Error should mention status, got: %v", err)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()

		client := NewClient()
		client.baseURL = server.URL

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := client.fetchRawHTTP(ctx, "TEST", "value")
		if err == nil {
			t.Error("fetchRawHTTP should error for cancelled context")
		}
	})
}

func TestClient_Elements(t *testing.T) {
	mockElements := []GPElement{
		{
			ObjectName: "ISS (ZARYA)",
			ObjectID:   "1998-067A",
			NoradCatID: 25544,
		},
	}
	mockData, _ := json.Marshal(mockElements)

	client := NewClient()
	client.fetchRaw = func(ctx context.Context, queryKey, queryValue string) ([]byte, error) {
		return mockData, nil
	}

	ctx := context.Background()
	elements, err := client.Elements(ctx, "ISS")
	if err != nil {
		t.Fatalf("Elements unexpected error: %v", err)
	}
	if len(elements) != 1 {
		t.Errorf("Elements length = %d, want 1", len(elements))
	}
	if elements[0].ObjectName != "ISS (ZARYA)" {
		t.Errorf("ObjectName = %q, want ISS (ZARYA)", elements[0].ObjectName)
	}
	if elements[0].NoradCatID != 25544 {
		t.Errorf("NoradCatID = %d, want 25544", elements[0].NoradCatID)
	}
}

func TestClient_Fetch(t *testing.T) {
	t.Run("successful decode", func(t *testing.T) {
		mockElements := []GPElement{
			{ObjectName: "TEST", NoradCatID: 12345},
		}
		mockData, _ := json.Marshal(mockElements)

		client := NewClient()
		client.fetchRaw = func(ctx context.Context, queryKey, queryValue string) ([]byte, error) {
			return mockData, nil
		}

		elements, err := client.Fetch(context.Background(), "NAME", "TEST")
		if err != nil {
			t.Fatalf("Fetch unexpected error: %v", err)
		}
		if len(elements) != 1 {
			t.Errorf("Fetch length = %d, want 1", len(elements))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		client := NewClient()
		client.fetchRaw = func(ctx context.Context, queryKey, queryValue string) ([]byte, error) {
			return []byte("{invalid json}"), nil
		}

		_, err := client.Fetch(context.Background(), "NAME", "TEST")
		if err == nil {
			t.Error("Fetch should error for invalid JSON")
		}
	})
}

func TestClient_CacheResolvedTarget3(t *testing.T) {
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

			normalized := normalizeTarget(tt.query)
			cached, ok := client.targetCache.GetTarget(normalized)
			if !ok {
				t.Error("Target not found in cache after CacheResolvedTarget")
			}
			if cached.Name != tt.resolved.Name {
				t.Errorf("Cached Name = %q, want %q", cached.Name, tt.resolved.Name)
			}
		})
	}
}

func TestClient_PassPredictions(t *testing.T) {
	// This test requires mocking multiple dependencies
	// We'll mock fetchRaw and ResolveTarget
	t.Skip("Requires complex mocking of sgp4 and TLE generation")

	client := NewClient()
	ctx := context.Background()
	observer := config.Observer{
		Name:            "Test Observer",
		LatitudeDeg:     40.7128,
		LongitudeDeg:    -74.0060,
		ElevationMeters: 0,
	}

	// Mock ResolveTarget
	client.resolveTarget = func(ctx context.Context, target string) (ResolvedTarget, error) {
		return ResolvedTarget{
			Name:    "ISS",
			NoradID: 25544,
		}, nil
	}

	// Mock fetchRaw to return valid TLE data
	client.fetchRaw = func(ctx context.Context, queryKey, queryValue string) ([]byte, error) {
		// Return valid OMM data
		return []byte(`[{"OBJECT_NAME":"ISS","NORAD_CAT_ID":25544}]`), nil
	}

	result, err := client.PassPredictions(ctx, "ISS", observer, "", "24h", 10)
	if err != nil {
		t.Fatalf("PassPredictions unexpected error: %v", err)
	}
	if result.Name == "" {
		t.Error("PassPredictions result Name is empty")
	}
}

func TestClient_NextPass(t *testing.T) {
	// This test requires complex mocking
	t.Skip("Requires complex mocking of PassPredictions")
}

func TestClient_Locate(t *testing.T) {
	// This test requires complex mocking
	t.Skip("Requires complex mocking of sgp4")
}

func TestClient_ResolveTarget_WithMock(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		client := NewClient()
		normalized := "ISS"
		cachedTarget := ResolvedTarget{
			Query:     normalized,
			Name:      "ISS (ZARYA)",
			NoradID:   25544,
			Kind:      "satellite",
			Source:    "celestrak",
			ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour),
		}
		client.targetCache.SetTarget(normalized, cachedTarget)

		resolved, err := client.ResolveTarget(context.Background(), "ISS")
		if err != nil {
			t.Fatalf("ResolveTarget unexpected error: %v", err)
		}
		if resolved.Source != "cache" {
			t.Errorf("Source = %q, want cache", resolved.Source)
		}
	})

	t.Run("numeric NORAD ID", func(t *testing.T) {
		client := NewClient()
		resolved, err := client.ResolveTarget(context.Background(), "25544")
		if err != nil {
			// Skip if network not available
			t.Skip("Network not available or API error:", err)
		}
		if resolved.NoradID != 25544 {
			t.Errorf("NoradID = %d, want 25544", resolved.NoradID)
		}
		if resolved.Source != "numeric" {
			t.Errorf("Source = %q, want numeric", resolved.Source)
		}
	})
}

// Helper function for tests
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func testObserver() config.Observer {
	return config.Observer{
		Name:            "Test Observer",
		LatitudeDeg:     47.2378,
		LongitudeDeg:    6.0241,
		ElevationMeters: 350,
	}
}

func testClient() *Client {
	return &Client{
		targetCache: NewTargetCache(""),
	}
}

func TestBuildURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		baseURL    string
		queryKey   string
		queryValue string
		wantErr    bool
	}{
		{
			name:       "builds URL with query parameters",
			baseURL:    "https://example.com/gp.php",
			queryKey:   QueryCATNR,
			queryValue: "25544",
		},
		{
			name:       "rejects invalid base URL",
			baseURL:    "://invalid-url",
			queryKey:   QueryCATNR,
			queryValue: "25544",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &Client{
				baseURL: tt.baseURL,
			}

			got, err := client.buildURL(tt.queryKey, tt.queryValue)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(got, "FORMAT=JSON") {
				t.Fatalf("expected FORMAT=JSON in URL, got %q", got)
			}

			if !strings.Contains(got, "CATNR=25544") {
				t.Fatalf("expected CATNR=25544 in URL, got %q", got)
			}
		})
	}

}

func TestFetch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		fetchError error
		wantLen    int
		wantErr    bool
	}{
		{
			name: "decodes valid JSON",
			body: `[
			{
				"OBJECT_NAME": "ISS (ZARYA)",
				"OBJECT_ID": "1998-067A",
				"NORAD_CAT_ID": 25544
			}
		]`,
			wantLen: 1,
		},
		{
			name:    "rejects invalid JSON",
			body:    `{invalid json}`,
			wantErr: true,
		},
		{
			name:       "returns fetch error",
			fetchError: errors.New("network failure"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := testClient()

			client.fetchRaw = func(
				ctx context.Context,
				queryKey string,
				queryValue string,
			) ([]byte, error) {
				return []byte(tt.body), tt.fetchError
			}

			got, err := client.Fetch(
				context.Background(),
				QueryCATNR,
				"25544",
			)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != tt.wantLen {
				t.Fatalf("expected %d elements, got %d", tt.wantLen, len(got))
			}
		})
	}

}

func TestElements(t *testing.T) {
	t.Parallel()

	client := testClient()

	var gotQueryKey string
	var gotQueryValue string

	client.fetchRaw = func(
		ctx context.Context,
		queryKey string,
		queryValue string,
	) ([]byte, error) {
		gotQueryKey = queryKey
		gotQueryValue = queryValue

		return []byte(`[
		{
			"OBJECT_NAME": "ISS (ZARYA)",
			"OBJECT_ID": "1998-067A",
			"NORAD_CAT_ID": 25544
		}
	]`), nil
	}

	elements, err := client.Elements(
		context.Background(),
		"ISS",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("expected one element, got %d", len(elements))
	}

	if gotQueryKey == "" {
		t.Fatal("expected query key to be passed")
	}

	if gotQueryValue == "" {
		t.Fatal("expected query value to be passed")
	}

}

func TestFetchRawHTTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
	}{
		{
			name:       "successful response",
			statusCode: http.StatusOK,
			body:       `{"success":true}`,
		},
		{
			name:       "non-200 response",
			statusCode: http.StatusInternalServerError,
			body:       `server error`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(
				http.HandlerFunc(func(
					writer http.ResponseWriter,
					request *http.Request,
				) {
					if request.Method != http.MethodGet {
						t.Errorf(
							"expected GET, got %s",
							request.Method,
						)
					}

					if request.URL.Query().Get("FORMAT") != "JSON" {
						t.Errorf("expected FORMAT=JSON")
					}

					writer.WriteHeader(tt.statusCode)
					_, _ = writer.Write([]byte(tt.body))
				}),
			)
			defer server.Close()

			client := &Client{
				baseURL:    server.URL,
				httpClient: server.Client(),
			}

			got, err := client.fetchRawHTTP(
				context.Background(),
				QueryCATNR,
				"25544",
			)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(got) != tt.body {
				t.Fatalf(
					"expected body %q, got %q",
					tt.body,
					string(got),
				)
			}
		})
	}

}

func TestFetchRawHTTP_ContextCancellation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(func(
			writer http.ResponseWriter,
			request *http.Request,
		) {
			<-request.Context().Done()
		}),
	)
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.fetchRawHTTP(
		ctx,
		QueryCATNR,
		"25544",
	)

	if err == nil {
		t.Fatal("expected context cancellation error")
	}

}

func TestParseTimeStr2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "empty value uses current UTC time",
			input: "",
		},
		{
			name:  "valid RFC3339 timestamp",
			input: "2026-06-03T22:00:00Z",
			want:  time.Date(2026, 6, 3, 22, 0, 0, 0, time.UTC),
		},
		{
			name:    "invalid timestamp",
			input:   "2026-06-03",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			before := time.Now().UTC()

			got, err := parseTimeStr(tt.input)

			after := time.Now().UTC()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.input == "" {
				if got.Before(before) || got.After(after) {
					t.Fatalf(
						"expected current UTC time, got %s",
						got,
					)
				}
				return
			}

			if !got.Equal(tt.want) {
				t.Fatalf(
					"expected %s, got %s",
					tt.want,
					got,
				)
			}
		})
	}

}

func TestComputeLookaheadTime2(t *testing.T) {
	t.Parallel()

	start := time.Date(
		2026,
		1,
		31,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	tests := []struct {
		lookahead string
		want      time.Time
		wantErr   bool
	}{
		{
			lookahead: "1Y",
			want:      start.AddDate(1, 0, 0),
		},
		{
			lookahead: "2M",
			want:      start.AddDate(0, 2, 0),
		},
		{
			lookahead: "3d",
			want:      start.AddDate(0, 0, 3),
		},
		{
			lookahead: "4h",
			want:      start.Add(4 * time.Hour),
		},
		{
			lookahead: "5m",
			want:      start.Add(5 * time.Minute),
		},
		{
			lookahead: "6s",
			want:      start.Add(6 * time.Second),
		},
		{
			lookahead: "invalid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.lookahead, func(t *testing.T) {
			t.Parallel()

			got, err := computeLookaheadTime(
				start,
				tt.lookahead,
			)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !got.Equal(tt.want) {
				t.Fatalf(
					"expected %s, got %s",
					tt.want,
					got,
				)
			}
		})
	}

}

func TestSplitLookahead3(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input     string
		wantValue int
		wantUnit  string
		wantErr   bool
	}{
		{
			input:     "1h",
			wantValue: 1,
			wantUnit:  "h",
		},
		{
			input:     "30m",
			wantValue: 30,
			wantUnit:  "m",
		},
		{
			input:     "2d",
			wantValue: 2,
			wantUnit:  "d",
		},
		{
			input:   "",
			wantErr: true,
		},
		{
			input:   "h",
			wantErr: true,
		},
		{
			input:   "10",
			wantErr: true,
		},
		{
			input:   "0h",
			wantErr: true,
		},
		{
			input:   "-1h",
			wantErr: true,
		},
		{
			input:   "1x",
			wantErr: true,
		},
		{
			input:   "1hh",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			gotValue, gotUnit, err := splitLookahead(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotValue != tt.wantValue {
				t.Fatalf(
					"expected value %d, got %d",
					tt.wantValue,
					gotValue,
				)
			}

			if gotUnit != tt.wantUnit {
				t.Fatalf(
					"expected unit %q, got %q",
					tt.wantUnit,
					gotUnit,
				)
			}
		})
	}

}

func TestFormatLookaheadDuration2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		duration time.Duration
		want     string
		wantErr  bool
	}{
		{
			duration: 24 * time.Hour,
			want:     "1d",
		},
		{
			duration: 2 * time.Hour,
			want:     "2h",
		},
		{
			duration: 30 * time.Minute,
			want:     "30m",
		},
		{
			duration: 45 * time.Second,
			want:     "45s",
		},
		{
			duration: 0,
			wantErr:  true,
		},
		{
			duration: 1500 * time.Millisecond,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			got, err := formatLookaheadDuration(tt.duration)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf(
					"expected %q, got %q",
					tt.want,
					got,
				)
			}
		})
	}

}

func TestCacheResolvedTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		query    string
		resolved ResolvedTarget
		wantErr  bool
	}{
		{
			name:  "caches normalized target",
			query: "  ISS  ",
			resolved: ResolvedTarget{
				NoradID: 25544,
			},
		},
		{
			name:    "rejects empty target",
			query:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := testClient()

			err := client.CacheResolvedTarget(
				tt.query,
				tt.resolved,
			)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

}

func TestPassPredictions_ResolveError(t *testing.T) {
	t.Parallel()

	client := testClient()

	client.resolveTarget = func(
		ctx context.Context,
		target string,
	) (ResolvedTarget, error) {
		return ResolvedTarget{}, errors.New("target not found")
	}

	_, err := client.PassPredictions(
		context.Background(),
		"unknown",
		testObserver(),
		"2026-06-03T22:00:00Z",
		"1h",
		10,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "target not found") {
		t.Fatalf(
			"expected target resolution error, got %v",
			err,
		)
	}

}

func TestPassPredictions_FetchError(t *testing.T) {
	t.Parallel()

	client := testClient()

	client.resolveTarget = func(
		ctx context.Context,
		target string,
	) (ResolvedTarget, error) {
		return ResolvedTarget{
			NoradID: 25544,
		}, nil
	}

	client.fetchRaw = func(
		ctx context.Context,
		queryKey string,
		queryValue string,
	) ([]byte, error) {
		return nil, errors.New("CelesTrak unavailable")
	}

	_, err := client.PassPredictions(
		context.Background(),
		"ISS",
		testObserver(),
		"2026-06-03T22:00:00Z",
		"1h",
		10,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "CelesTrak unavailable") {
		t.Fatalf(
			"expected fetch error, got %v",
			err,
		)
	}

}

func TestPassPredictions_InvalidTime(t *testing.T) {
	t.Parallel()

	client := testClient()

	client.resolveTarget = func(
		ctx context.Context,
		target string,
	) (ResolvedTarget, error) {
		t.Fatal("resolveTarget should not be called")
		return ResolvedTarget{}, nil
	}

	client.fetchRaw = func(
		ctx context.Context,
		queryKey string,
		queryValue string,
	) ([]byte, error) {
		t.Fatal("fetchRaw should not be called")
		return nil, nil
	}

	_, err := client.PassPredictions(
		context.Background(),
		"ISS",
		testObserver(),
		"invalid-time",
		"1h",
		10,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "invalid datetime") {
		t.Fatalf(
			"expected invalid datetime error, got %v",
			err,
		)
	}
}

func TestNextPass_ResolveError(t *testing.T) {
	t.Parallel()

	client := testClient()

	client.resolveTarget = func(
		ctx context.Context,
		target string,
	) (ResolvedTarget, error) {
		return ResolvedTarget{}, errors.New("cannot resolve satellite")
	}

	_, err := client.NextPass(
		context.Background(),
		"ISS",
		testObserver(),
		"2026-06-03T22:00:00Z",
		"2h",
		10,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

}

func TestNextPass_InvalidLookahead(t *testing.T) {
	t.Parallel()

	client := testClient()

	_, err := client.NextPass(
		context.Background(),
		"ISS",
		testObserver(),
		"2026-06-03T22:00:00Z",
		"invalid",
		10,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "invalid lookahead") {
		t.Fatalf(
			"expected invalid lookahead error, got %v",
			err,
		)
	}

}

func TestLocate_ResolveError(t *testing.T) {
	t.Parallel()

	client := testClient()

	client.resolveTarget = func(
		ctx context.Context,
		target string,
	) (ResolvedTarget, error) {
		return ResolvedTarget{}, errors.New("target not found")
	}

	_, err := client.Locate(
		context.Background(),
		"unknown",
		testObserver(),
		time.Now(),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

}

func TestLocate_FetchError(t *testing.T) {
	t.Parallel()

	client := testClient()

	client.resolveTarget = func(
		ctx context.Context,
		target string,
	) (ResolvedTarget, error) {
		return ResolvedTarget{
			NoradID: 25544,
		}, nil
	}

	client.fetchRaw = func(
		ctx context.Context,
		queryKey string,
		queryValue string,
	) ([]byte, error) {
		return nil, errors.New("network failure")
	}

	_, err := client.Locate(
		context.Background(),
		"ISS",
		testObserver(),
		time.Now(),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

}

func TestLocate_InvalidOMM(t *testing.T) {
	t.Parallel()

	client := testClient()

	client.resolveTarget = func(
		ctx context.Context,
		target string,
	) (ResolvedTarget, error) {
		return ResolvedTarget{
			NoradID: 25544,
		}, nil
	}

	client.fetchRaw = func(
		ctx context.Context,
		queryKey string,
		queryValue string,
	) ([]byte, error) {
		return []byte(`invalid OMM data`), nil
	}

	_, err := client.Locate(
		context.Background(),
		"ISS",
		testObserver(),
		time.Now(),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "parse CelesTrak OMM data") {
		t.Fatalf(
			"expected OMM parsing error, got %v",
			err,
		)
	}

}

func TestContextCancellation(t *testing.T) {
	t.Parallel()

	client := testClient()

	client.fetchRaw = func(
		ctx context.Context,
		queryKey string,
		queryValue string,
	) ([]byte, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return nil, errors.New("unexpected execution")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Fetch(
		ctx,
		QueryCATNR,
		"25544",
	)

	if err == nil {
		t.Fatal("expected context cancellation error")
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}

}

// This fixture is intentionally kept in the test file.
// It avoids any network dependency.
func validISSOMM() []byte {
	return []byte(`
[
  {
    "OBJECT_NAME": "ISS (ZARYA)",
    "OBJECT_ID": "1998-067A",
    "CENTER_NAME": "EARTH",
    "REF_FRAME": "TEME",
    "TIME_SYSTEM": "UTC",
    "MEAN_ELEMENT_THEORY": "SGP4",
    "EPOCH": "2026-06-03T22:00:00.000000",
    "MEAN_MOTION": 15.49515369,
    "ECCENTRICITY": 0.00014270,
    "INCLINATION": 51.6345,
    "ARG_OF_PERICENTER": 263.9877,
    "MEAN_ANOMALY": 96.1141,
    "EPHEMERIS_TYPE": 0,
    "CLASSIFICATION_TYPE": "U",
    "NORAD_CAT_ID": 25544,
    "ELEMENT_SET_NO": 999,
    "REV_AT_EPOCH": 12345,
    "BSTAR": 0.00012345,
    "MEAN_MOTION_DOT": 0.00001234,
    "MEAN_MOTION_DDOT": 0.00000000
  }
]
`)
}
