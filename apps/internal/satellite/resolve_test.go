package satellite

import (
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
			name:           "numeric NORAD ID",
			target:         "25544",
			wantQueryKey:   QueryCATNR,
			wantQueryValue: "25544",
			wantErr:        false,
		},
		{
			name:           "numeric with spaces",
			target:         "  25544  ",
			wantQueryKey:   QueryCATNR,
			wantQueryValue: "25544",
			wantErr:        false,
		},
		{
			name:           "satellite name",
			target:         "ISS",
			wantQueryKey:   QueryNAME,
			wantQueryValue: "ISS",
			wantErr:        false,
		},
		{
			name:           "satellite name with spaces",
			target:         "  ISS  ",
			wantQueryKey:   QueryNAME,
			wantQueryValue: "ISS",
			wantErr:        false,
		},
		{
			name:           "satellite name with letters and numbers",
			target:         "STARLINK-1000",
			wantQueryKey:   QueryNAME,
			wantQueryValue: "STARLINK-1000",
			wantErr:        false,
		},
		{
			name:           "satellite name with spaces internally",
			target:         "HUBBLE SPACE",
			wantQueryKey:   QueryNAME,
			wantQueryValue: "HUBBLE SPACE",
			wantErr:        false,
		},
		{
			name:           "empty target",
			target:         "",
			wantQueryKey:   "",
			wantQueryValue: "",
			wantErr:        true,
		},
		{
			name:           "whitespace only",
			target:         "   ",
			wantQueryKey:   "",
			wantQueryValue: "",
			wantErr:        true,
		},
		{
			name:           "zero as string",
			target:         "0",
			wantQueryKey:   QueryCATNR,
			wantQueryValue: "0",
			wantErr:        false,
		},
		{
			name:           "negative number",
			target:         "-1",
			wantQueryKey:   QueryCATNR,
			wantQueryValue: "-1",
			wantErr:        false,
		},
		{
			name:           "numeric with leading zeros",
			target:         "001234",
			wantQueryKey:   QueryCATNR,
			wantQueryValue: "001234",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryKey, queryValue, err := BuildCelesTrakQuery(tt.target)

			if tt.wantErr {
				if err == nil {
					t.Errorf("BuildCelesTrakQuery(%q) should error", tt.target)
				}
				return
			}

			if err != nil {
				t.Fatalf("BuildCelesTrakQuery(%q) unexpected error: %v", tt.target, err)
			}

			if queryKey != tt.wantQueryKey {
				t.Errorf("BuildCelesTrakQuery(%q) queryKey = %q, want %q", tt.target, queryKey, tt.wantQueryKey)
			}

			if queryValue != tt.wantQueryValue {
				t.Errorf("BuildCelesTrakQuery(%q) queryValue = %q, want %q", tt.target, queryValue, tt.wantQueryValue)
			}
		})
	}
}

func TestBuildCelesTrakQuery_Constants(t *testing.T) {
	// Verify constants are correctly defined
	if QueryCATNR != "CATNR" {
		t.Errorf("QueryCATNR = %q, want CATNR", QueryCATNR)
	}
	if QueryNAME != "NAME" {
		t.Errorf("QueryNAME = %q, want NAME", QueryNAME)
	}
}

func TestBuildCelesTrakQuery_NumericDetection(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		expected string
	}{
		{
			name:     "simple number",
			target:   "123",
			expected: QueryCATNR,
		},
		{
			name:     "number with plus sign",
			target:   "+123",
			expected: QueryCATNR,
		},
		{
			name:     "number with leading zeros",
			target:   "00123",
			expected: QueryCATNR,
		},
		{
			name:     "starts with number but has letters",
			target:   "123ABC",
			expected: QueryNAME,
		},
		{
			name:     "starts with number but has spaces",
			target:   "123 ABC",
			expected: QueryNAME,
		},
		{
			name:     "contains decimal",
			target:   "123.45",
			expected: QueryNAME,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryKey, _, err := BuildCelesTrakQuery(tt.target)
			if err != nil {
				t.Fatalf("BuildCelesTrakQuery(%q) unexpected error: %v", tt.target, err)
			}
			if queryKey != tt.expected {
				t.Errorf("BuildCelesTrakQuery(%q) queryKey = %q, want %q", tt.target, queryKey, tt.expected)
			}
		})
	}
}

// Test that demonstrates the function's behavior with edge cases
func TestBuildCelesTrakQuery_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		target        string
		shouldSucceed bool
		queryKey      string
	}{
		{
			name:          "very long number",
			target:        "12345678901234567890",
			shouldSucceed: true,
			queryKey:      QueryNAME,
		},
		{
			name:          "very long name",
			target:        "THIS_IS_A_VERY_LONG_SATELLITE_NAME_THAT_EXCEEDS_NORMAL_LENGTH",
			shouldSucceed: true,
			queryKey:      QueryNAME,
		},
		{
			name:          "special characters in name",
			target:        "SATELLITE-123_ABC",
			shouldSucceed: true,
			queryKey:      QueryNAME,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryKey, queryValue, err := BuildCelesTrakQuery(tt.target)
			if tt.shouldSucceed {
				if err != nil {
					t.Fatalf("BuildCelesTrakQuery(%q) unexpected error: %v", tt.target, err)
				}
				if queryKey != tt.queryKey {
					t.Errorf("queryKey = %q, want %q", queryKey, tt.queryKey)
				}
				if queryValue == "" {
					t.Errorf("queryValue should not be empty for %q", tt.target)
				}
			}
		})
	}
}
