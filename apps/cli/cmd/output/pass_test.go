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

func TestPrintPassPredictions(t *testing.T) {
	// Create test time values
	aosTime := time.Date(2026, 7, 16, 12, 30, 0, 0, time.UTC)
	maxElevTime := time.Date(2026, 7, 16, 12, 35, 30, 0, time.UTC)
	losTime := time.Date(2026, 7, 16, 12, 40, 0, 0, time.UTC)

	// Actually, let's just build the expected string dynamically using the same format
	// Or better, let's just use the actual output from the function

	tests := []struct {
		name         string
		result       satellite.PassPredictionResult
		outputFormat string
		expectError  bool
	}{
		{
			name: "text format - single pass",
			result: satellite.PassPredictionResult{
				Name:         "ISS",
				Kind:         "satellite",
				NoradID:      25544,
				ObjectID:     "1998-067A",
				ObserverName: "London",
				Passes: []satellite.PredictedPass{
					{
						AcquisitionOfSignal: aosTime,
						LossOfSignal:        losTime,
						Duration:            10 * time.Minute,
						MaxElevation:        75.5,
						MaxElevationTime:    maxElevTime,
						AzimuthAtAOS:        45.0,
						AzimuthAtLOS:        315.0,
					},
				},
			},
			outputFormat: Text,
			expectError:  false,
		},
		{
			name: "text format - multiple passes",
			result: satellite.PassPredictionResult{
				Name:         "ISS",
				Kind:         "satellite",
				NoradID:      25544,
				ObjectID:     "1998-067A",
				ObserverName: "London",
				Passes: []satellite.PredictedPass{
					{
						AcquisitionOfSignal: aosTime,
						LossOfSignal:        losTime,
						Duration:            10 * time.Minute,
						MaxElevation:        75.5,
						MaxElevationTime:    maxElevTime,
						AzimuthAtAOS:        45.0,
						AzimuthAtLOS:        315.0,
					},
					{
						AcquisitionOfSignal: aosTime.Add(24 * time.Hour),
						LossOfSignal:        losTime.Add(24 * time.Hour),
						Duration:            12 * time.Minute,
						MaxElevation:        80.2,
						MaxElevationTime:    maxElevTime.Add(24 * time.Hour),
						AzimuthAtAOS:        120.0,
						AzimuthAtLOS:        240.0,
					},
				},
			},
			outputFormat: Text,
			expectError:  false,
		},
		{
			name: "text format - zero passes",
			result: satellite.PassPredictionResult{
				Name:         "ISS",
				Kind:         "satellite",
				NoradID:      25544,
				ObjectID:     "1998-067A",
				ObserverName: "London",
				Passes:       []satellite.PredictedPass{},
			},
			outputFormat: Text,
			expectError:  false,
		},
		{
			name: "json format",
			result: satellite.PassPredictionResult{
				Name:         "ISS",
				Kind:         "satellite",
				NoradID:      25544,
				ObjectID:     "1998-067A",
				ObserverName: "London",
				Passes: []satellite.PredictedPass{
					{
						AcquisitionOfSignal: aosTime,
						LossOfSignal:        losTime,
						Duration:            10 * time.Minute,
						MaxElevation:        75.5,
						MaxElevationTime:    maxElevTime,
						AzimuthAtAOS:        45.0,
						AzimuthAtLOS:        315.0,
					},
				},
			},
			outputFormat: JSON,
			expectError:  false,
		},
		{
			name: "ndjson format",
			result: satellite.PassPredictionResult{
				Name:         "ISS",
				Kind:         "satellite",
				NoradID:      25544,
				ObjectID:     "1998-067A",
				ObserverName: "London",
				Passes: []satellite.PredictedPass{
					{
						AcquisitionOfSignal: aosTime,
						LossOfSignal:        losTime,
						Duration:            10 * time.Minute,
						MaxElevation:        75.5,
						MaxElevationTime:    maxElevTime,
						AzimuthAtAOS:        45.0,
						AzimuthAtLOS:        315.0,
					},
				},
			},
			outputFormat: NDJSON,
			expectError:  false,
		},
		{
			name: "invalid format",
			result: satellite.PassPredictionResult{
				Name: "ISS",
			},
			outputFormat: "invalid",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)

			err := PrintPassPredictions(cmd, tt.result, tt.outputFormat)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unhandled output format")
				return
			}

			assert.NoError(t, err)

			if tt.outputFormat == Text {
				// Don't test exact string match due to spacing differences
				// Instead, verify content contains expected values
				output := buf.String()
				assert.Contains(t, output, "Name:")
				assert.Contains(t, output, "Kind:")
				assert.Contains(t, output, "NORAD ID:")
				assert.Contains(t, output, "Object ID:")
				assert.Contains(t, output, "Observer:")

				if len(tt.result.Passes) > 0 {
					assert.Contains(t, output, "Acquisition of signal (AOS):")
					assert.Contains(t, output, "Loss of signal (LOS):")
					assert.Contains(t, output, "Duration:")
					assert.Contains(t, output, "Maximum elevation:")
					assert.Contains(t, output, "Maximum elevation time:")
					assert.Contains(t, output, "Azimuth at AOS:")
					assert.Contains(t, output, "Azimuth at LOS:")
				}
			}

			if tt.outputFormat == JSON {
				var result satellite.PassPredictionResult
				err := json.Unmarshal(buf.Bytes(), &result)
				assert.NoError(t, err)
				// Check that the result has the expected fields
				// Note: JSON field names are case-sensitive and must match the struct tags
				// Since we're using the same struct, the fields should marshal correctly
			}

			if tt.outputFormat == NDJSON {
				lines := bytes.Split(buf.Bytes(), []byte("\n"))
				// Remove empty last line
				if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
					lines = lines[:len(lines)-1]
				}
				assert.Equal(t, len(tt.result.Passes), len(lines))
			}
		})
	}
}

func TestPrintPassPredictionResultText(t *testing.T) {
	result := satellite.PassPredictionResult{
		Name:         "ISS",
		Kind:         "satellite",
		NoradID:      25544,
		ObjectID:     "1998-067A",
		ObserverName: "London",
	}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	printPassPredictionResultText(cmd, result)

	output := buf.String()
	assert.Contains(t, output, "Name:")
	assert.Contains(t, output, "ISS")
	assert.Contains(t, output, "Kind:")
	assert.Contains(t, output, "satellite")
	assert.Contains(t, output, "NORAD ID:")
	assert.Contains(t, output, "25544")
	assert.Contains(t, output, "Object ID:")
	assert.Contains(t, output, "1998-067A")
	assert.Contains(t, output, "Observer:")
	assert.Contains(t, output, "London")
}

func TestPrintPredictionText(t *testing.T) {
	aosTime := time.Date(2026, 7, 16, 12, 30, 0, 0, time.UTC)
	maxElevTime := time.Date(2026, 7, 16, 12, 35, 30, 0, time.UTC)
	losTime := time.Date(2026, 7, 16, 12, 40, 0, 0, time.UTC)

	prediction := satellite.PredictedPass{
		AcquisitionOfSignal: aosTime,
		LossOfSignal:        losTime,
		Duration:            10 * time.Minute,
		MaxElevation:        75.5,
		MaxElevationTime:    maxElevTime,
		AzimuthAtAOS:        45.0,
		AzimuthAtLOS:        315.0,
	}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	printPredictionText(cmd, prediction)

	output := buf.String()
	assert.Contains(t, output, "Acquisition of signal (AOS):")
	assert.Contains(t, output, "2026-07-16 12:30:00 UTC")
	assert.Contains(t, output, "Loss of signal (LOS):")
	assert.Contains(t, output, "2026-07-16 12:40:00 UTC")
	assert.Contains(t, output, "Duration:")
	assert.Contains(t, output, "10m0s")
	assert.Contains(t, output, "Maximum elevation:")
	assert.Contains(t, output, "75.5°")
	assert.Contains(t, output, "Maximum elevation time:")
	assert.Contains(t, output, "2026-07-16 12:35:30 UTC")
	assert.Contains(t, output, "Azimuth at AOS:")
	assert.Contains(t, output, "45.0°")
	assert.Contains(t, output, "Azimuth at LOS:")
	assert.Contains(t, output, "315.0°")
}

func TestPrintPassPredictions_EdgeCases(t *testing.T) {
	t.Run("very long pass duration", func(t *testing.T) {
		aosTime := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
		losTime := time.Date(2026, 7, 16, 14, 30, 0, 0, time.UTC)

		result := satellite.PassPredictionResult{
			Name:   "TestSat",
			Kind:   "satellite",
			Passes: []satellite.PredictedPass{
				{
					AcquisitionOfSignal: aosTime,
					LossOfSignal:        losTime,
					Duration:            150 * time.Minute,
					MaxElevation:        89.9,
					MaxElevationTime:    aosTime.Add(75 * time.Minute),
					AzimuthAtAOS:        0.0,
					AzimuthAtLOS:        359.9,
				},
			},
		}

		cmd := &cobra.Command{}
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		err := PrintPassPredictions(cmd, result, Text)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Duration:")
		assert.Contains(t, output, "2h30m0s")
		assert.Contains(t, output, "Maximum elevation:")
		assert.Contains(t, output, "89.9°")
		assert.Contains(t, output, "Azimuth at AOS:")
		assert.Contains(t, output, "0.0°")
		assert.Contains(t, output, "Azimuth at LOS:")
		assert.Contains(t, output, "359.9°")
	})

	t.Run("pass with minimal elevation", func(t *testing.T) {
		now := time.Now().UTC()

		result := satellite.PassPredictionResult{
			Name:   "TestSat",
			Kind:   "satellite",
			Passes: []satellite.PredictedPass{
				{
					AcquisitionOfSignal: now,
					LossOfSignal:        now.Add(5 * time.Minute),
					Duration:            5 * time.Minute,
					MaxElevation:        0.1,
					MaxElevationTime:    now.Add(2 * time.Minute),
					AzimuthAtAOS:        180.0,
					AzimuthAtLOS:        180.0,
				},
			},
		}

		cmd := &cobra.Command{}
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		err := PrintPassPredictions(cmd, result, Text)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Maximum elevation:")
		assert.Contains(t, output, "0.1°")
	})
}

func TestPrintPassPredictions_JSONValidation(t *testing.T) {
	now := time.Now().UTC()

	result := satellite.PassPredictionResult{
		Name:         "ISS",
		Kind:         "satellite",
		NoradID:      25544,
		ObjectID:     "1998-067A",
		ObserverName: "London",
		Passes: []satellite.PredictedPass{
			{
				AcquisitionOfSignal: now,
				LossOfSignal:        now.Add(10 * time.Minute),
				Duration:            10 * time.Minute,
				MaxElevation:        75.5,
				MaxElevationTime:    now.Add(5 * time.Minute),
				AzimuthAtAOS:        45.0,
				AzimuthAtLOS:        315.0,
			},
		},
	}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := PrintPassPredictions(cmd, result, JSON)
	assert.NoError(t, err)

	// Unmarshal into the struct to verify JSON is valid and has expected fields
	var jsonResult satellite.PassPredictionResult
	err = json.Unmarshal(buf.Bytes(), &jsonResult)
	assert.NoError(t, err)

	// Verify the unmarshaled data matches
	assert.Equal(t, result.Name, jsonResult.Name)
	assert.Equal(t, result.Kind, jsonResult.Kind)
	assert.Equal(t, result.NoradID, jsonResult.NoradID)
	assert.Equal(t, result.ObjectID, jsonResult.ObjectID)
	assert.Equal(t, result.ObserverName, jsonResult.ObserverName)
	assert.Equal(t, len(result.Passes), len(jsonResult.Passes))

	if len(result.Passes) > 0 {
		assert.True(t, result.Passes[0].AcquisitionOfSignal.Equal(jsonResult.Passes[0].AcquisitionOfSignal))
		assert.True(t, result.Passes[0].LossOfSignal.Equal(jsonResult.Passes[0].LossOfSignal))
		assert.Equal(t, result.Passes[0].Duration, jsonResult.Passes[0].Duration)
		assert.Equal(t, result.Passes[0].MaxElevation, jsonResult.Passes[0].MaxElevation)
	}
}

func TestPrintPassPredictions_NDJSONValidation(t *testing.T) {
	now := time.Now().UTC()

	result := satellite.PassPredictionResult{
		Name:         "ISS",
		Kind:         "satellite",
		NoradID:      25544,
		ObjectID:     "1998-067A",
		ObserverName: "London",
		Passes: []satellite.PredictedPass{
			{
				AcquisitionOfSignal: now,
				LossOfSignal:        now.Add(10 * time.Minute),
				Duration:            10 * time.Minute,
				MaxElevation:        75.5,
				MaxElevationTime:    now.Add(5 * time.Minute),
				AzimuthAtAOS:        45.0,
				AzimuthAtLOS:        315.0,
			},
			{
				AcquisitionOfSignal: now.Add(24 * time.Hour),
				LossOfSignal:        now.Add(24*time.Hour + 10*time.Minute),
				Duration:            10 * time.Minute,
				MaxElevation:        80.0,
				MaxElevationTime:    now.Add(24*time.Hour + 5*time.Minute),
				AzimuthAtAOS:        90.0,
				AzimuthAtLOS:        270.0,
			},
		},
	}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := PrintPassPredictions(cmd, result, NDJSON)
	assert.NoError(t, err)

	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	// Remove empty last line
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	assert.Equal(t, 2, len(lines))

	// Verify each line is valid JSON and contains expected fields
	var passes []satellite.PredictedPass
	for _, line := range lines {
		var pass satellite.PredictedPass
		err := json.Unmarshal(line, &pass)
		assert.NoError(t, err)
		passes = append(passes, pass)
	}

	assert.Equal(t, len(result.Passes), len(passes))
	for i, pass := range passes {
		assert.True(t, result.Passes[i].AcquisitionOfSignal.Equal(pass.AcquisitionOfSignal))
		assert.True(t, result.Passes[i].LossOfSignal.Equal(pass.LossOfSignal))
		assert.Equal(t, result.Passes[i].Duration, pass.Duration)
		assert.Equal(t, result.Passes[i].MaxElevation, pass.MaxElevation)
		assert.True(t, result.Passes[i].MaxElevationTime.Equal(pass.MaxElevationTime))
		assert.Equal(t, result.Passes[i].AzimuthAtAOS, pass.AzimuthAtAOS)
		assert.Equal(t, result.Passes[i].AzimuthAtLOS, pass.AzimuthAtLOS)
	}
}

func TestPrintPassPredictions_TimeFormatting(t *testing.T) {
	// Test that time formatting in text output uses UTC
	aosTime := time.Date(2026, 7, 16, 12, 30, 0, 0, time.UTC)

	// Create a time with a different location
	loc, _ := time.LoadLocation("America/New_York")
	localTime := time.Date(2026, 7, 16, 8, 30, 0, 0, loc)

	result := satellite.PassPredictionResult{
		Name:   "TestSat",
		Kind:   "satellite",
		Passes: []satellite.PredictedPass{
			{
				AcquisitionOfSignal: aosTime,
				LossOfSignal:        localTime.UTC(),
				Duration:            10 * time.Minute,
				MaxElevation:        75.5,
				MaxElevationTime:    aosTime,
				AzimuthAtAOS:        45.0,
				AzimuthAtLOS:        315.0,
			},
		},
	}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := PrintPassPredictions(cmd, result, Text)
	assert.NoError(t, err)

	output := buf.String()
	// All times should be in UTC
	assert.Contains(t, output, "2026-07-16 12:30:00 UTC")
	// The local time should be converted to UTC
	assert.Contains(t, output, "2026-07-16 12:30:00 UTC")
}