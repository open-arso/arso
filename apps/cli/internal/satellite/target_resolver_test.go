package satellite

import (
	"context"
	"net/http"
	"testing"
)

func TestResolveTargetNumericTarget(t *testing.T) {
	client := &Client{targetCache: NewTargetCache("")}

	got, err := client.ResolveTarget(context.Background(), "25544")
	if err != nil {
		t.Fatalf("ResolveTarget() unexpected error: %v", err)
	}

	if got.NoradID != 25544 || got.Source != "numeric" || got.Name != "25544" {
		t.Fatalf("ResolveTarget() = %#v, want numeric target metadata", got)
	}
}

func TestResolveTargetUsesCache(t *testing.T) {
	cache := NewTargetCache("")
	cache.Targets["ISS"] = ResolvedTarget{
		Query:    "ISS",
		Name:     "ISS (ZARYA)",
		NoradID:  25544,
		ObjectID: "1998-067A",
		Source:   "celestrak",
	}

	client := &Client{targetCache: cache}

	got, err := client.ResolveTarget(context.Background(), " iss ")
	if err != nil {
		t.Fatalf("ResolveTarget() unexpected error: %v", err)
	}

	if got.Source != "cache" || got.NoradID != 25544 {
		t.Fatalf("ResolveTarget() = %#v, want cached result", got)
	}
}

func TestResolveTargetReturnsAmbiguousError(t *testing.T) {
	client := &Client{
		baseURL: "https://example.com/gp.php",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return jsonHTTPResponse(http.StatusOK, `[
			{"OBJECT_NAME":"ISS (ZARYA)","OBJECT_ID":"1998-067A","NORAD_CAT_ID":25544},
			{"OBJECT_NAME":"ISS-DEB","OBJECT_ID":"1998-067B","NORAD_CAT_ID":25545}
		]`), nil
		})},
		targetCache: NewTargetCache(""),
	}

	_, err := client.ResolveTarget(context.Background(), "ISS")
	if err == nil {
		t.Fatal("ResolveTarget() expected error, got nil")
	}

	ambiguousErr, ok := err.(*AmbiguousTargetError)
	if !ok {
		t.Fatalf("ResolveTarget() error = %T, want *AmbiguousTargetError", err)
	}
	if got, want := len(ambiguousErr.Candidates), 2; got != want {
		t.Fatalf("len(Candidates) = %d, want %d", got, want)
	}
}

func TestResolveTargetFetchesSingleResultAndCachesIt(t *testing.T) {
	cache := NewTargetCache("")
	client := &Client{
		baseURL: "https://example.com/gp.php",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return jsonHTTPResponse(http.StatusOK, `[{"OBJECT_NAME":"ISS (ZARYA)","OBJECT_ID":"1998-067A","NORAD_CAT_ID":25544}]`), nil
		})},
		targetCache: cache,
	}

	got, err := client.ResolveTarget(context.Background(), "ISS")
	if err != nil {
		t.Fatalf("ResolveTarget() unexpected error: %v", err)
	}

	if got.Source != "celestrak" || got.NoradID != 25544 || got.Name != "ISS (ZARYA)" {
		t.Fatalf("ResolveTarget() = %#v, want fetched target metadata", got)
	}
	if got.ResolvedAt.IsZero() || got.ExpiresAt.IsZero() {
		t.Fatalf("ResolveTarget() timestamps = %#v, want non-zero timestamps", got)
	}

	cached, ok := cache.GetTarget("ISS")
	if !ok {
		t.Fatal("cache.GetTarget() = false, want true")
	}
	if cached.NoradID != 25544 {
		t.Fatalf("cached target = %#v, want NORAD 25544", cached)
	}
}
