package clioutput

import "testing"

func TestNormalize(t *testing.T) {
	got := Normalize(" JSON ")

	if got != JSON {
		t.Fatalf("Normalize() = %q, want %q", got, JSON)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		allowed []string
		wantErr bool
	}{
		{
			name:    "accepts text",
			format:  Text,
			allowed: []string{Text, JSON},
			wantErr: false,
		},
		{
			name:    "accepts json",
			format:  JSON,
			allowed: []string{Text, JSON},
			wantErr: false,
		},
		{
			name:    "accepts ndjson when allowed",
			format:  NDJSON,
			allowed: []string{Text, JSON, NDJSON},
			wantErr: false,
		},
		{
			name:    "rejects ndjson when not allowed",
			format:  NDJSON,
			allowed: []string{Text, JSON},
			wantErr: true,
		},
		{
			name:    "rejects yaml",
			format:  "yaml",
			allowed: []string{Text, JSON, NDJSON},
			wantErr: true,
		},
		{
			name:    "uses text and json as default allowed formats",
			format:  JSON,
			allowed: nil,
			wantErr: false,
		},
		{
			name:    "rejects ndjson with default allowed formats",
			format:  NDJSON,
			allowed: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.format, tt.allowed...)

			if tt.wantErr && err == nil {
				t.Fatal("Validate() expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("Validate() unexpected error: %v", err)
			}
		})
	}
}
