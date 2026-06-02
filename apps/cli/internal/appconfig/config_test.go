package appconfig

import "testing"

func float64Ptr(value float64) *float64 {
	return &value
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
