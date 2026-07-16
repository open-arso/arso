package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/openarso/arso/apps/internal/satellite"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestPrintApparentPositions(t *testing.T) {
	tests := []struct {
		name         string
		positions    []satellite.ApparentPosition
		outputFormat string
		expectError  bool
		expectedText string
	}{
		{
			name: "text format - single position",
			positions: []satellite.ApparentPosition{
				{
					Name:                 "ISS",
					Kind:                 "satellite",
					NoradID:              25544,
					ObjectID:             "1998-067A",
					ObserverName:         "London",
					TimeUTC:              "2026-07-16T12:00:00Z",
					AzimuthDeg:           45.5,
					ElevationDeg:         30.2,
					RangeKm:              450.75,
					RangeRateKms:         3.4567,
					AboveHorizon:         true,
					SatelliteLatitudeDeg: 51.5,
					SatelliteLongitudeDeg: -0.1,
					SatelliteAltitudeKm:  408.0,
				},
			},
			outputFormat: Text,
			expectError:  false,
			expectedText: `Name:          ISS
Kind:          satellite
NORAD ID:      25544
Object ID:     1998-067A
Observer:      London
Time UTC:      2026-07-16T12:00:00Z
Azimuth:       45.50°
Elevation:     30.20°
Range:         450.75 km
Range rate:    3.4567 km/s
Above horizon: true
Subpoint:      51.5000°, -0.1000°
Altitude:      408.00 km
`,
		},
		{
			name: "text format - multiple positions",
			positions: []satellite.ApparentPosition{
				{
					Name:                 "ISS",
					Kind:                 "satellite",
					NoradID:              25544,
					ObjectID:             "1998-067A",
					ObserverName:         "London",
					TimeUTC:              "2026-07-16T12:00:00Z",
					AzimuthDeg:           45.5,
					ElevationDeg:         30.2,
					RangeKm:              450.75,
					RangeRateKms:         3.4567,
					AboveHorizon:         true,
					SatelliteLatitudeDeg: 51.5,
					SatelliteLongitudeDeg: -0.1,
					SatelliteAltitudeKm:  408.0,
				},
				{
					Name:                 "Hubble",
					Kind:                 "satellite",
					NoradID:              20580,
					ObjectID:             "1990-037B",
					ObserverName:         "London",
					TimeUTC:              "2026-07-16T12:05:00Z",
					AzimuthDeg:           120.3,
					ElevationDeg:         15.8,
					RangeKm:              520.25,
					RangeRateKms:         2.1234,
					AboveHorizon:         false,
					SatelliteLatitudeDeg: 48.2,
					SatelliteLongitudeDeg: -2.5,
					SatelliteAltitudeKm:  540.0,
				},
			},
			outputFormat: Text,
			expectError:  false,
			expectedText: `Name:          ISS
Kind:          satellite
NORAD ID:      25544
Object ID:     1998-067A
Observer:      London
Time UTC:      2026-07-16T12:00:00Z
Azimuth:       45.50°
Elevation:     30.20°
Range:         450.75 km
Range rate:    3.4567 km/s
Above horizon: true
Subpoint:      51.5000°, -0.1000°
Altitude:      408.00 km

Name:          Hubble
Kind:          satellite
NORAD ID:      20580
Object ID:     1990-037B
Observer:      London
Time UTC:      2026-07-16T12:05:00Z
Azimuth:       120.30°
Elevation:     15.80°
Range:         520.25 km
Range rate:    2.1234 km/s
Above horizon: false
Subpoint:      48.2000°, -2.5000°
Altitude:      540.00 km
`,
		},
		{
			name:         "json format",
			positions:    []satellite.ApparentPosition{{Name: "ISS", NoradID: 25544}},
			outputFormat: JSON,
			expectError:  false,
		},
		{
			name:         "ndjson format",
			positions:    []satellite.ApparentPosition{{Name: "ISS", NoradID: 25544}},
			outputFormat: NDJSON,
			expectError:  false,
		},
		{
			name:         "invalid format",
			positions:    []satellite.ApparentPosition{{Name: "ISS"}},
			outputFormat: "invalid",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)

			err := PrintApparentPositions(cmd, tt.positions, tt.outputFormat)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.outputFormat == Text && tt.expectedText != "" {
				assert.Equal(t, tt.expectedText, buf.String())
			}

			if tt.outputFormat == JSON {
				var result []satellite.ApparentPosition
				err := json.Unmarshal(buf.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, tt.positions, result)
			}

			if tt.outputFormat == NDJSON {
				lines := bytes.Split(buf.Bytes(), []byte("\n"))
				// Remove empty last line
				if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
					lines = lines[:len(lines)-1]
				}
				assert.Equal(t, len(tt.positions), len(lines))
				for i, pos := range tt.positions {
					var result satellite.ApparentPosition
					err := json.Unmarshal(lines[i], &result)
					assert.NoError(t, err)
					assert.Equal(t, pos, result)
				}
			}
		})
	}
}

func TestPrintApparentPositionText(t *testing.T) {
	position := satellite.ApparentPosition{
		Name:                 "TestSat",
		Kind:                 "satellite",
		NoradID:              12345,
		ObjectID:             "2026-001A",
		ObserverName:         "TestObserver",
		TimeUTC:              "2026-07-16T12:00:00Z",
		AzimuthDeg:           45.6789,
		ElevationDeg:         30.1234,
		RangeKm:              450.789,
		RangeRateKms:         3.4567,
		AboveHorizon:         true,
		SatelliteLatitudeDeg: 51.5074,
		SatelliteLongitudeDeg: -0.1278,
		SatelliteAltitudeKm:  408.123,
	}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	PrintApparentPositionText(cmd, position)

	expected := `Name:          TestSat
Kind:          satellite
NORAD ID:      12345
Object ID:     2026-001A
Observer:      TestObserver
Time UTC:      2026-07-16T12:00:00Z
Azimuth:       45.68°
Elevation:     30.12°
Range:         450.79 km
Range rate:    3.4567 km/s
Above horizon: true
Subpoint:      51.5074°, -0.1278°
Altitude:      408.12 km
`

	assert.Equal(t, expected, buf.String())
}

func TestPrintElements(t *testing.T) {
	elements := []satellite.GPElement{
		{
			ObjectName:      "ISS",
			NoradCatID:      25544,
			ObjectID:        "1998-067A",
			Epoch:           "2026-07-16T12:00:00",
			Inclination:     51.641,
			RAOfAscNode:     100.123,
			Eccentricity:    0.0006789,
			ArgOfPericenter: 90.456,
			MeanAnomaly:     180.789,
			MeanMotion:      15.49569645,
			BStar:           0.000123,
			MeanMotionDot:   0.000001,
			MeanMotionDDot:  0.000000,
			ElementSetNo:    999,
			RevAtEpoch:      12345,
		},
	}

	tests := []struct {
		name         string
		elements     []satellite.GPElement
		outputFormat string
		expectError  bool
		expectedText string
	}{
		{
			name:         "text format",
			elements:     elements,
			outputFormat: Text,
			expectError:  false,
			expectedText: `Name:              ISS
Kind:              satellite
Source:            celestrak
NORAD ID:          25544
Object ID:         1998-067A
Epoch UTC:         2026-07-16T12:00:00Z
Inclination:       51.6410°
RAAN:              100.1230°
Eccentricity:      0.0006789
Arg. pericenter:   90.4560°
Mean anomaly:      180.7890°
Mean motion:       15.49569645 rev/day
BSTAR:             0.000123
`,
		},
		{
			name:         "json format",
			elements:     elements,
			outputFormat: JSON,
			expectError:  false,
		},
		{
			name:         "ndjson format",
			elements:     elements,
			outputFormat: NDJSON,
			expectError:  false,
		},
		{
			name:         "invalid format",
			elements:     elements,
			outputFormat: "invalid",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)

			err := PrintElements(cmd, tt.elements, tt.outputFormat)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.outputFormat == Text && tt.expectedText != "" {
				assert.Equal(t, tt.expectedText, buf.String())
			}

			if tt.outputFormat == JSON {
				var result []elementOutput
				err := json.Unmarshal(buf.Bytes(), &result)
				assert.NoError(t, err)
				assert.Len(t, result, 1)
				assert.Equal(t, "satellite", result[0].Kind)
				assert.Equal(t, "celestrak", result[0].Source)
			}
		})
	}
}

func TestPrintElementText(t *testing.T) {
	el := elementOutput{
		Name:                "ISS",
		Kind:                "satellite",
		Source:              "celestrak",
		NoradID:             25544,
		ObjectID:            "1998-067A",
		EpochUTC:            "2026-07-16T12:00:00Z",
		InclinationDeg:      51.641,
		RAANDeg:             100.123,
		Eccentricity:        0.0006789,
		ArgOfPericenterDeg:  90.456,
		MeanAnomalyDeg:      180.789,
		MeanMotionRevPerDay: 15.49569645,
		BStar:               0.000123,
	}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	PrintElementText(cmd, el)

	expected := `Name:              ISS
Kind:              satellite
Source:            celestrak
NORAD ID:          25544
Object ID:         1998-067A
Epoch UTC:         2026-07-16T12:00:00Z
Inclination:       51.6410°
RAAN:              100.1230°
Eccentricity:      0.0006789
Arg. pericenter:   90.4560°
Mean anomaly:      180.7890°
Mean motion:       15.49569645 rev/day
BSTAR:             0.000123
`

	assert.Equal(t, expected, buf.String())
}

func TestParseFindAt(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
		checkTime   func(time.Time) bool
	}{
		{
			name:        "empty string returns current time",
			value:       "",
			expectError: false,
			checkTime: func(t time.Time) bool {
				// Allow small difference for test execution time
				diff := time.Now().UTC().Sub(t)
				return diff < time.Second
			},
		},
		{
			name:        "valid RFC3339 with Z",
			value:       "2026-07-16T12:00:00Z",
			expectError: false,
			checkTime: func(t time.Time) bool {
				expected, _ := time.Parse(time.RFC3339, "2026-07-16T12:00:00Z")
				return t.Equal(expected)
			},
		},
		{
			name:        "valid RFC3339 with timezone",
			value:       "2026-07-16T12:00:00+02:00",
			expectError: false,
			checkTime: func(t time.Time) bool {
				// Should be converted to UTC
				expected, _ := time.Parse(time.RFC3339, "2026-07-16T10:00:00Z")
				return t.Equal(expected)
			},
		},
		{
			name:        "invalid format",
			value:       "2026-07-16 12:00:00",
			expectError: true,
			checkTime:   nil,
		},
		{
			name:        "invalid format - wrong separator",
			value:       "2026/07/16T12:00:00Z",
			expectError: true,
			checkTime:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFindAt(tt.value)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid --at value")
				return
			}

			assert.NoError(t, err)
			if tt.checkTime != nil {
				assert.True(t, tt.checkTime(result), "Time check failed")
			}
		})
	}
}

func TestToElementOutput(t *testing.T) {
	el := satellite.GPElement{
		ObjectName:      "ISS",
		NoradCatID:      25544,
		ObjectID:        "1998-067A",
		Epoch:           "2026-07-16T12:00:00",
		Inclination:     51.641,
		RAOfAscNode:     100.123,
		Eccentricity:    0.0006789,
		ArgOfPericenter: 90.456,
		MeanAnomaly:     180.789,
		MeanMotion:      15.49569645,
		BStar:           0.000123,
		MeanMotionDot:   0.000001,
		MeanMotionDDot:  0.000000,
		ElementSetNo:    999,
		RevAtEpoch:      12345,
	}

	result := toElementOutput(el)

	assert.Equal(t, "ISS", result.Name)
	assert.Equal(t, "satellite", result.Kind)
	assert.Equal(t, "celestrak", result.Source)
	assert.Equal(t, 25544, result.NoradID)
	assert.Equal(t, "1998-067A", result.ObjectID)
	assert.Equal(t, "2026-07-16T12:00:00Z", result.EpochUTC)
	assert.Equal(t, 51.641, result.InclinationDeg)
	assert.Equal(t, 100.123, result.RAANDeg)
	assert.Equal(t, 0.0006789, result.Eccentricity)
	assert.Equal(t, 90.456, result.ArgOfPericenterDeg)
	assert.Equal(t, 180.789, result.MeanAnomalyDeg)
	assert.Equal(t, 15.49569645, result.MeanMotionRevPerDay)
	assert.Equal(t, 0.000123, result.BStar)
	assert.Equal(t, 0.000001, result.MeanMotionDot)
	assert.Equal(t, 0.000000, result.MeanMotionDDot)
	assert.Equal(t, 999, result.ElementSetNo)
	assert.Equal(t, 12345, result.RevAtEpoch)
}

func TestNormalizeCelesTrakEpoch(t *testing.T) {
	tests := []struct {
		name     string
		epoch    string
		expected string
	}{
		{
			name:     "empty string",
			epoch:    "",
			expected: "",
		},
		{
			name:     "already has Z suffix",
			epoch:    "2026-07-16T12:00:00Z",
			expected: "2026-07-16T12:00:00Z",
		},
		{
			name:     "needs Z suffix",
			epoch:    "2026-07-16T12:00:00",
			expected: "2026-07-16T12:00:00Z",
		},
		{
			name:     "with timezone - no change needed",
			epoch:    "2026-07-16T12:00:00+02:00",
			expected: "2026-07-16T10:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeCelesTrakEpoch(tt.epoch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create a test command
func createTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
	}
	return cmd
}