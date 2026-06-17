package satellite

import "testing"

func TestResolveTarget(t *testing.T) {
	tests := []struct {
		name           string
		target         string
		wantQueryKey   string
		wantQueryValue string
		wantErr        bool
	}{
		{
			name:           "resolves ISS alias to NORAD catalog ID",
			target:         "ISS",
			wantQueryKey:   QueryCATNR,
			wantQueryValue: ISSNoradCatalogID,
			wantErr:        false,
		},
		{
			name:           "resolves ISS alias case insensitively",
			target:         "iss",
			wantQueryKey:   QueryCATNR,
			wantQueryValue: ISSNoradCatalogID,
			wantErr:        false,
		},
		{
			name:           "trims ISS alias",
			target:         "  ISS  ",
			wantQueryKey:   QueryCATNR,
			wantQueryValue: ISSNoradCatalogID,
			wantErr:        false,
		},
		{
			name:           "resolves numeric target as catalog ID",
			target:         "25544",
			wantQueryKey:   QueryCATNR,
			wantQueryValue: "25544",
			wantErr:        false,
		},
		{
			name:           "trims numeric target",
			target:         "  25544  ",
			wantQueryKey:   QueryCATNR,
			wantQueryValue: "25544",
			wantErr:        false,
		},
		{
			name:           "resolves object name as NAME query",
			target:         "HUBBLE",
			wantQueryKey:   QueryNAME,
			wantQueryValue: "HUBBLE",
			wantErr:        false,
		},
		{
			name:           "preserves object name spacing after trimming",
			target:         "  STARLINK 1234  ",
			wantQueryKey:   QueryNAME,
			wantQueryValue: "STARLINK 1234",
			wantErr:        false,
		},
		{
			name:           "returns error for empty target",
			target:         "",
			wantQueryKey:   "",
			wantQueryValue: "",
			wantErr:        true,
		},
		{
			name:           "returns error for whitespace-only target",
			target:         "   ",
			wantQueryKey:   "",
			wantQueryValue: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQueryKey, gotQueryValue, err := ResolveTarget(tt.target)

			if tt.wantErr {
				if err == nil {
					t.Fatal("ResolveTarget() expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("ResolveTarget() unexpected error: %v", err)
			}

			if gotQueryKey != tt.wantQueryKey {
				t.Fatalf("ResolveTarget() queryKey = %q, want %q", gotQueryKey, tt.wantQueryKey)
			}

			if gotQueryValue != tt.wantQueryValue {
				t.Fatalf("ResolveTarget() queryValue = %q, want %q", gotQueryValue, tt.wantQueryValue)
			}
		})
	}
}

func TestAmbiguousTargetErrorMessage(t *testing.T) {
	err := (&AmbiguousTargetError{
		Target: "ISS",
		Candidates: []ResolvedTarget{
			{Name: "ISS (ZARYA)"},
			{Name: "ISS-DEB"},
		},
	}).Error()

	for _, fragment := range []string{`target "ISS" is ambiguous`, "2 satellites found"} {
		if !strings.Contains(err, fragment) {
			t.Fatalf("error message %q missing fragment %q", err, fragment)
		}
	}
}
