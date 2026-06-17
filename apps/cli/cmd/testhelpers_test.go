package cmd

import (
	"bytes"
	"time"

	"github.com/openarso/arso/apps/cli/internal/appconfig"
	"github.com/openarso/arso/apps/cli/internal/satellite"
	"github.com/spf13/cobra"
)

func float64Ptr(value float64) *float64 {
	return &value
}

func configuredTestConfig() appconfig.Config {
	cfg := appconfig.Default()
	cfg.Node.Name = "paris-node"
	cfg.Node.ID = "paris"
	cfg.Observatory.Latitude = float64Ptr(48.8566)
	cfg.Observatory.Longitude = float64Ptr(2.3522)
	cfg.Observatory.ElevationMeters = 35
	return cfg
}

func samplePassPredictionResult() satellite.PassPredictionResult {
	return satellite.PassPredictionResult{
		Name:         "ISS (ZARYA)",
		Kind:         "satellite",
		Source:       "celestrak_sgp4",
		NoradID:      25544,
		ObjectID:     "1998-067A",
		ObserverName: "paris-node",
		Passes: []satellite.PredictedPass{
			{
				AcquisitionOfSignal: time.Date(2026, 6, 10, 22, 15, 0, 0, time.UTC),
				LossOfSignal:        time.Date(2026, 6, 10, 22, 27, 30, 0, time.UTC),
				Duration:            12*time.Minute + 30*time.Second,
				MaxElevation:        47.2,
				MaxElevationTime:    time.Date(2026, 6, 10, 22, 21, 15, 0, time.UTC),
				AzimuthAtAOS:        130.4,
				AzimuthAtLOS:        44.8,
			},
			{
				AcquisitionOfSignal: time.Date(2026, 6, 11, 0, 3, 0, 0, time.UTC),
				LossOfSignal:        time.Date(2026, 6, 11, 0, 14, 0, 0, time.UTC),
				Duration:            11 * time.Minute,
				MaxElevation:        31.5,
				MaxElevationTime:    time.Date(2026, 6, 11, 0, 8, 45, 0, time.UTC),
				AzimuthAtAOS:        214.2,
				AzimuthAtLOS:        98.6,
			},
		},
	}
}

func sampleSinglePassPredictionResult() satellite.PassPredictionResult {
	result := samplePassPredictionResult()
	result.Passes = result.Passes[:1]
	return result
}

func sampleApparentPositions() []satellite.ApparentPosition {
	return []satellite.ApparentPosition{
		{
			Name:                  "ISS (ZARYA)",
			Kind:                  "satellite",
			Source:                "celestrak_sgp4",
			NoradID:               25544,
			ObjectID:              "1998-067A",
			ObserverName:          "paris-node",
			TimeUTC:               "2026-06-10T22:00:00Z",
			AzimuthDeg:            123.45,
			ElevationDeg:          67.89,
			RangeKm:               420.12,
			RangeRateKms:          -0.1234,
			AboveHorizon:          true,
			SatelliteLatitudeDeg:  10.1234,
			SatelliteLongitudeDeg: 20.5678,
			SatelliteAltitudeKm:   408.55,
		},
	}
}

func sampleElements() []satellite.GPElement {
	return []satellite.GPElement{
		{
			ObjectName:      "ISS (ZARYA)",
			ObjectID:        "1998-067A",
			NoradCatID:      25544,
			Epoch:           "2026-06-10T22:00:00",
			Inclination:     51.6432,
			RAOfAscNode:     120.1234,
			Eccentricity:    0.0006703,
			ArgOfPericenter: 65.4321,
			MeanAnomaly:     12.3456,
			MeanMotion:      15.48912345,
			BStar:           0.0000123,
		},
	}
}

func newTestCommandIO() (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	cmd := &cobra.Command{Use: "test"}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd, stdout, stderr
}
