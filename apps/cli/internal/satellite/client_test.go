package satellite

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestClientBuildURLIncludesQueryAndFormat(t *testing.T) {
	client := &Client{baseURL: "https://example.com/gp.php"}

	got, err := client.buildURL(QueryNAME, "ISS (ZARYA)")
	if err != nil {
		t.Fatalf("buildURL() unexpected error: %v", err)
	}

	for _, fragment := range []string{"NAME=ISS+%28ZARYA%29", "FORMAT=JSON"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("buildURL() = %q, want fragment %q", got, fragment)
		}
	}
}

func TestClientFetchDecodesElements(t *testing.T) {
	client := &Client{
		baseURL: "https://example.com/gp.php",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if got := r.URL.Query().Get(QueryNAME); got != "ISS" {
				t.Fatalf("query NAME = %q, want %q", got, "ISS")
			}
			if got := r.URL.Query().Get("FORMAT"); got != "JSON" {
				t.Fatalf("query FORMAT = %q, want %q", got, "JSON")
			}

			return jsonHTTPResponse(http.StatusOK, `[{"OBJECT_NAME":"ISS (ZARYA)","OBJECT_ID":"1998-067A","NORAD_CAT_ID":25544}]`), nil
		})},
		targetCache: NewTargetCache(""),
	}

	elements, err := client.Fetch(context.Background(), QueryNAME, "ISS")
	if err != nil {
		t.Fatalf("Fetch() unexpected error: %v", err)
	}
	if got, want := len(elements), 1; got != want {
		t.Fatalf("len(elements) = %d, want %d", got, want)
	}
	if got, want := elements[0].NoradCatID, 25544; got != want {
		t.Fatalf("NoradCatID = %d, want %d", got, want)
	}
}

func TestClientFetchRawReturnsStatusError(t *testing.T) {
	client := &Client{
		baseURL: "https://example.com/gp.php",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return jsonHTTPResponse(http.StatusBadGateway, "nope"), nil
		})},
		targetCache: NewTargetCache(""),
	}

	_, err := client.fetchRaw(context.Background(), QueryNAME, "ISS")
	if err == nil {
		t.Fatal("fetchRaw() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "CelesTrak returned status 502 Bad Gateway") {
		t.Fatalf("fetchRaw() error = %q, want status message", err)
	}
}

func TestComputeLookaheadTime(t *testing.T) {
	start := time.Date(2026, 6, 10, 22, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		lookahead string
		want      time.Time
	}{
		{name: "hours", lookahead: "48h", want: start.Add(48 * time.Hour)},
		{name: "days", lookahead: "7d", want: start.AddDate(0, 0, 7)},
		{name: "months", lookahead: "2M", want: start.AddDate(0, 2, 0)},
		{name: "years", lookahead: "1Y", want: start.AddDate(1, 0, 0)},
		{name: "minutes", lookahead: "30m", want: start.Add(30 * time.Minute)},
		{name: "seconds", lookahead: "45s", want: start.Add(45 * time.Second)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computeLookaheadTime(start, tt.lookahead)
			if err != nil {
				t.Fatalf("computeLookaheadTime() unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("computeLookaheadTime(%q) = %s, want %s", tt.lookahead, got, tt.want)
			}
		})
	}
}

func TestComputeLookaheadTimeRejectsInvalidInput(t *testing.T) {
	start := time.Date(2026, 6, 10, 22, 0, 0, 0, time.UTC)

	for _, lookahead := range []string{"", "abc", "10", "0h", "1w"} {
		t.Run(lookahead, func(t *testing.T) {
			if _, err := computeLookaheadTime(start, lookahead); err == nil {
				t.Fatal("computeLookaheadTime() expected error, got nil")
			}
		})
	}
}

func TestFormatLookaheadDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
		wantErr  bool
	}{
		{name: "days", duration: 48 * time.Hour, want: "2d"},
		{name: "hours", duration: 3 * time.Hour, want: "3h"},
		{name: "minutes", duration: 90 * time.Minute, want: "90m"},
		{name: "seconds", duration: 45 * time.Second, want: "45s"},
		{name: "rejects zero", duration: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatLookaheadDuration(tt.duration)
			if tt.wantErr {
				if err == nil {
					t.Fatal("formatLookaheadDuration() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("formatLookaheadDuration() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("formatLookaheadDuration() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTimeStrRejectsInvalidRFC3339(t *testing.T) {
	if _, err := parseTimeStr("tomorrow"); err == nil {
		t.Fatal("parseTimeStr() expected error, got nil")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func jsonHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
