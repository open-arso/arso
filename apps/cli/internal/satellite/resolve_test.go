package satellite

import (
	"strings"
	"testing"
)

func TestBuildCelesTrakQuery(t *testing.T) {
	tests := []struct {
		name           string
		target         string
		wantQueryKey   string
		wantQueryValue string
		wantErr        bool
	}{
		{
			name:           "resolves numeric target as catalog ID",
			target:         "25544",
			wantQueryKey:   QueryCATNR,
			wantQueryValue: "25544",
		},
		{
			name:           "trims numeric target",
			target:         " 25544 ",
			wantQueryKey:   QueryCATNR,
			wantQueryValue: "25544",
		},
		{
			name:           "resolves object name as NAME query",
			target:         "HUBBLE",
			wantQueryKey:   QueryNAME,
			wantQueryValue: "HUBBLE",
		},
		{
			name:           "preserves name spacing after trimming",
			target:         "  STARLINK 1234  ",
			wantQueryKey:   QueryNAME,
			wantQueryValue: "STARLINK 1234",
		},
		{
			name:    "rejects empty target",
			target:  "",
			wantErr: true,
		},
		{
			name:    "rejects whitespace-only target",
			target:  "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQueryKey, gotQueryValue, err := BuildCelesTrakQuery(tt.target)

			if tt.wantErr {
				if err == nil {
					t.Fatal("BuildCelesTrakQuery() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("BuildCelesTrakQuery() unexpected error: %v", err)
			}
			if gotQueryKey != tt.wantQueryKey {
				t.Fatalf("queryKey = %q, want %q", gotQueryKey, tt.wantQueryKey)
			}
			if gotQueryValue != tt.wantQueryValue {
				t.Fatalf("queryValue = %q, want %q", gotQueryValue, tt.wantQueryValue)
			}
		})
	}
}
