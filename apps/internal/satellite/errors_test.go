package satellite

import (
	"testing"
)

func TestAmbiguousTargetError2(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		candidates []ResolvedTarget
		expected   string
	}{
		{
			name:   "two candidates",
			target: "ISS",
			candidates: []ResolvedTarget{
				{
					Name:    "ISS (ZARYA)",
					NoradID: 25544,
					Kind:    "satellite",
				},
				{
					Name:    "ISS (DECO)",
					NoradID: 12345,
					Kind:    "satellite",
				},
			},
			expected: `target "ISS" is ambiguous: 2 satellites found`,
		},
		{
			name:   "single candidate",
			target: "STARLINK",
			candidates: []ResolvedTarget{
				{
					Name:    "STARLINK-1000",
					NoradID: 99999,
					Kind:    "satellite",
				},
			},
			expected: `target "STARLINK" is ambiguous: 1 satellites found`,
		},
		{
			name:       "no candidates",
			target:     "UNKNOWN",
			candidates: []ResolvedTarget{},
			expected:   `target "UNKNOWN" is ambiguous: 0 satellites found`,
		},
		{
			name:   "many candidates",
			target: "SATELLITE",
			candidates: []ResolvedTarget{
				{Name: "SATELLITE A", NoradID: 11111},
				{Name: "SATELLITE B", NoradID: 22222},
				{Name: "SATELLITE C", NoradID: 33333},
				{Name: "SATELLITE D", NoradID: 44444},
				{Name: "SATELLITE E", NoradID: 55555},
			},
			expected: `target "SATELLITE" is ambiguous: 5 satellites found`,
		},
		{
			name:   "candidate with empty fields",
			target: "TEST",
			candidates: []ResolvedTarget{
				{},
				{Name: "TEST SATELLITE", NoradID: 99999},
			},
			expected: `target "TEST" is ambiguous: 2 satellites found`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &AmbiguousTargetError{
				Target:     tt.target,
				Candidates: tt.candidates,
			}

			if err.Error() != tt.expected {
				t.Errorf("AmbiguousTargetError.Error() = %q, want %q", err.Error(), tt.expected)
			}
		})
	}
}

func TestAmbiguousTargetError_ImplementsError(t *testing.T) {
	// Verify that AmbiguousTargetError implements the error interface
	var err error = &AmbiguousTargetError{
		Target:     "TEST",
		Candidates: []ResolvedTarget{},
	}
	if err == nil {
		t.Error("AmbiguousTargetError should implement error interface")
	}
}

func TestAmbiguousTargetError_Fields(t *testing.T) {
	target := "ISS"
	candidates := []ResolvedTarget{
		{Name: "Candidate 1", NoradID: 11111},
		{Name: "Candidate 2", NoradID: 22222},
	}

	err := &AmbiguousTargetError{
		Target:     target,
		Candidates: candidates,
	}

	if err.Target != target {
		t.Errorf("Target field = %q, want %q", err.Target, target)
	}
	if len(err.Candidates) != len(candidates) {
		t.Errorf("Candidates length = %d, want %d", len(err.Candidates), len(candidates))
	}
	for i, candidate := range err.Candidates {
		if candidate.Name != candidates[i].Name {
			t.Errorf("Candidate[%d] Name = %q, want %q", i, candidate.Name, candidates[i].Name)
		}
		if candidate.NoradID != candidates[i].NoradID {
			t.Errorf("Candidate[%d] NoradID = %d, want %d", i, candidate.NoradID, candidates[i].NoradID)
		}
	}
}

// Benchmark tests
func BenchmarkAmbiguousTargetError_Error(b *testing.B) {
	err := &AmbiguousTargetError{
		Target: "ISS",
		Candidates: []ResolvedTarget{
			{Name: "ISS (ZARYA)", NoradID: 25544},
			{Name: "ISS (DECO)", NoradID: 12345},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
