package config

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func float64Ptr(value float64) *float64 {
	return &value
}

func TestPathUsesEnvVar(t *testing.T) {
	customPath := filepath.Join(t.TempDir(), "custom-config.yaml")
	t.Setenv(ConfigEnvVar, customPath)

	got, err := Path()
	if err != nil {
		t.Fatalf("Path() unexpected error: %v", err)
	}

	if got != customPath {
		t.Fatalf("Path() = %q, want %q", got, customPath)
	}
}

func TestLoadReturnsDefaultsWhenConfigMissing(t *testing.T) {
	t.Setenv(ConfigEnvVar, filepath.Join(t.TempDir(), "config.yaml"))

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if want := Default(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestInitCreatesConfigAndExists(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv(ConfigEnvVar, configPath)

	gotPath, err := Init(false)
	if err != nil {
		t.Fatalf("Init() unexpected error: %v", err)
	}

	if gotPath != configPath {
		t.Fatalf("Init() path = %q, want %q", gotPath, configPath)
	}

	exists, err := Exists()
	if err != nil {
		t.Fatalf("Exists() unexpected error: %v", err)
	}
	if !exists {
		t.Fatal("Exists() = false, want true")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if want := Default(); !reflect.DeepEqual(cfg, want) {
		t.Fatalf("Load() = %#v, want %#v", cfg, want)
	}
}

func TestInitRejectsExistingWithoutOverwrite(t *testing.T) {
	t.Setenv(ConfigEnvVar, filepath.Join(t.TempDir(), "config.yaml"))

	if _, err := Init(false); err != nil {
		t.Fatalf("first Init() unexpected error: %v", err)
	}

	if _, err := Init(false); err == nil {
		t.Fatal("second Init() expected error, got nil")
	}
}

func TestSetAndGetRoundTrip(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv(ConfigEnvVar, configPath)

	updates := []struct {
		key   string
		value string
	}{
		{key: "node.name", value: "remote-1"},
		{key: "node.id", value: "remote-1"},
		{key: "api.url", value: "https://example.com"},
		{key: "observatory.latitude", value: "48.8566"},
		{key: "observatory.longitude", value: "2.3522"},
		{key: "observatory.elevation_meters", value: "35"},
		{key: "output.format", value: "json"},
	}

	for _, update := range updates {
		gotPath, err := Set(update.key, update.value)
		if err != nil {
			t.Fatalf("Set(%q, %q) unexpected error: %v", update.key, update.value, err)
		}
		if gotPath != configPath {
			t.Fatalf("Set(%q, %q) path = %q, want %q", update.key, update.value, gotPath, configPath)
		}
	}

	gotValue, err := Get("node.name")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if gotValue != "remote-1" {
		t.Fatalf("Get(node.name) = %#v, want %#v", gotValue, "remote-1")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Node.Name != "remote-1" || cfg.Node.ID != "remote-1" {
		t.Fatalf("node config = %#v, want updated values", cfg.Node)
	}
	if cfg.API.URL != "https://example.com" {
		t.Fatalf("API URL = %q, want %q", cfg.API.URL, "https://example.com")
	}
	if cfg.Observatory.Latitude == nil || *cfg.Observatory.Latitude != 48.8566 {
		t.Fatalf("latitude = %#v, want 48.8566", cfg.Observatory.Latitude)
	}
	if cfg.Observatory.Longitude == nil || *cfg.Observatory.Longitude != 2.3522 {
		t.Fatalf("longitude = %#v, want 2.3522", cfg.Observatory.Longitude)
	}
	if cfg.Observatory.ElevationMeters != 35 {
		t.Fatalf("elevation = %v, want 35", cfg.Observatory.ElevationMeters)
	}
	if cfg.Output.Format != "json" {
		t.Fatalf("output format = %q, want %q", cfg.Output.Format, "json")
	}
}

func TestSetRejectsInvalidValues(t *testing.T) {
	t.Setenv(ConfigEnvVar, filepath.Join(t.TempDir(), "config.yaml"))

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr string
	}{
		{name: "unknown key", key: "unknown.key", value: "value", wantErr: "unsupported config key"},
		{name: "node id whitespace", key: "node.id", value: "bad id", wantErr: "cannot contain whitespace"},
		{name: "invalid api url", key: "api.url", value: "localhost:8080", wantErr: "must include scheme and host"},
		{name: "latitude out of range", key: "observatory.latitude", value: "100", wantErr: "must be between -90.00 and 90.00"},
		{name: "longitude out of range", key: "observatory.longitude", value: "200", wantErr: "must be between -180.00 and 180.00"},
		{name: "invalid output format", key: "output.format", value: "ndjson", wantErr: "must be one of: text, json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Set(tt.key, tt.value)
			if err == nil {
				t.Fatal("Set() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Set() error = %q, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestGetUnknownKeyReturnsError(t *testing.T) {
	t.Setenv(ConfigEnvVar, filepath.Join(t.TempDir(), "config.yaml"))

	if _, err := Get("unknown.key"); err == nil {
		t.Fatal("Get() expected error, got nil")
	}
}

func TestObservatoryConfigIsConfigured(t *testing.T) {
	tests := []struct {
		name string
		cfg  ObservatoryConfig
		want bool
	}{
		{
			name: "not configured when latitude and longitude are nil",
			cfg: ObservatoryConfig{
				Latitude:        nil,
				Longitude:       nil,
				ElevationMeters: 0,
			},
			want: false,
		},
		{
			name: "not configured when latitude is set but longitude is nil",
			cfg: ObservatoryConfig{
				Latitude:        float64Ptr(48.8566),
				Longitude:       nil,
				ElevationMeters: 35,
			},
			want: false,
		},
		{
			name: "not configured when longitude is set but latitude is nil",
			cfg: ObservatoryConfig{
				Latitude:        nil,
				Longitude:       float64Ptr(2.3522),
				ElevationMeters: 35,
			},
			want: false,
		},
		{
			name: "configured when latitude is zero and longitude is set",
			cfg: ObservatoryConfig{
				Latitude:        float64Ptr(0),
				Longitude:       float64Ptr(43),
				ElevationMeters: 0,
			},
			want: true,
		},
		{
			name: "configured when latitude and longitude are both set",
			cfg: ObservatoryConfig{
				Latitude:        float64Ptr(48.8566),
				Longitude:       float64Ptr(2.3522),
				ElevationMeters: 35,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.IsConfigured()

			if got != tt.want {
				t.Fatalf("IsConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObservatoryConfigRequireConfigured(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ObservatoryConfig
		wantErr bool
	}{
		{
			name: "returns error when latitude and longitude are nil",
			cfg: ObservatoryConfig{
				Latitude:        nil,
				Longitude:       nil,
				ElevationMeters: 0,
			},
			wantErr: true,
		},
		{
			name: "returns error when latitude is nil",
			cfg: ObservatoryConfig{
				Latitude:        nil,
				Longitude:       float64Ptr(2.3522),
				ElevationMeters: 35,
			},
			wantErr: true,
		},
		{
			name: "returns error when longitude is nil",
			cfg: ObservatoryConfig{
				Latitude:        float64Ptr(48.8566),
				Longitude:       nil,
				ElevationMeters: 35,
			},
			wantErr: true,
		},
		{
			name: "does not return error when latitude is zero and longitude is set",
			cfg: ObservatoryConfig{
				Latitude:        float64Ptr(0),
				Longitude:       float64Ptr(43),
				ElevationMeters: 0,
			},
			wantErr: false,
		},
		{
			name: "does not return error when latitude and longitude are set",
			cfg: ObservatoryConfig{
				Latitude:        float64Ptr(48.8566),
				Longitude:       float64Ptr(2.3522),
				ElevationMeters: 35,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.RequireConfigured()

			if tt.wantErr && err == nil {
				t.Fatal("RequireConfigured() expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("RequireConfigured() unexpected error: %v", err)
			}
		})
	}
}
