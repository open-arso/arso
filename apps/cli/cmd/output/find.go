package output

import (
	"fmt"
	"time"

	"github.com/openarso/arso/apps/internal/satellite"
	"github.com/spf13/cobra"
)

type elementOutput struct {
	Name                string  `json:"name"`
	Kind                string  `json:"kind"`
	Source              string  `json:"source"`
	NoradID             int     `json:"norad_id"`
	ObjectID            string  `json:"object_id"`
	EpochUTC            string  `json:"epoch_utc"`
	InclinationDeg      float64 `json:"inclination_deg"`
	RAANDeg             float64 `json:"raan_deg"`
	Eccentricity        float64 `json:"eccentricity"`
	ArgOfPericenterDeg  float64 `json:"arg_of_pericenter_deg"`
	MeanAnomalyDeg      float64 `json:"mean_anomaly_deg"`
	MeanMotionRevPerDay float64 `json:"mean_motion_rev_per_day"`
	BStar               float64 `json:"bstar,omitempty"`
	MeanMotionDot       float64 `json:"mean_motion_dot,omitempty"`
	MeanMotionDDot      float64 `json:"mean_motion_ddot,omitempty"`
	ElementSetNo        int     `json:"element_set_no,omitempty"`
	RevAtEpoch          int     `json:"rev_at_epoch,omitempty"`
}

func PrintApparentPositions(cmd *cobra.Command, positions []satellite.ApparentPosition, outputFormat string) error {
	switch outputFormat {
	case Text:
		for i, position := range positions {
			if i > 0 {
				fmt.Fprintln(cmd.OutOrStdout())
			}

			PrintApparentPositionText(cmd, position)
		}
		return nil

	case JSON:
		return PrintJSON(cmd, positions)

	case NDJSON:
		return PrintNDJSON(cmd, positions)

	default:
		return fmt.Errorf("unhandled output format %q", outputFormat)
	}
}

func PrintApparentPositionText(cmd *cobra.Command, position satellite.ApparentPosition) {
	fmt.Fprintf(cmd.OutOrStdout(), "Name:          %s\n", position.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "Kind:          %s\n", position.Kind)
	fmt.Fprintf(cmd.OutOrStdout(), "NORAD ID:      %d\n", position.NoradID)
	fmt.Fprintf(cmd.OutOrStdout(), "Object ID:     %s\n", position.ObjectID)
	fmt.Fprintf(cmd.OutOrStdout(), "Observer:      %s\n", position.ObserverName)
	fmt.Fprintf(cmd.OutOrStdout(), "Time UTC:      %s\n", position.TimeUTC)
	fmt.Fprintf(cmd.OutOrStdout(), "Azimuth:       %.2f°\n", position.AzimuthDeg)
	fmt.Fprintf(cmd.OutOrStdout(), "Elevation:     %.2f°\n", position.ElevationDeg)
	fmt.Fprintf(cmd.OutOrStdout(), "Range:         %.2f km\n", position.RangeKm)
	fmt.Fprintf(cmd.OutOrStdout(), "Range rate:    %.4f km/s\n", position.RangeRateKms)
	fmt.Fprintf(cmd.OutOrStdout(), "Above horizon: %t\n", position.AboveHorizon)
	fmt.Fprintf(cmd.OutOrStdout(), "Subpoint:      %.4f°, %.4f°\n", position.SatelliteLatitudeDeg, position.SatelliteLongitudeDeg)
	fmt.Fprintf(cmd.OutOrStdout(), "Altitude:      %.2f km\n", position.SatelliteAltitudeKm)
}

func PrintElements(cmd *cobra.Command, elements []satellite.GPElement, outputFormat string) error {
	outputElements := make([]elementOutput, 0, len(elements))

	for _, el := range elements {
		outputElements = append(outputElements, toElementOutput(el))
	}

	switch outputFormat {
	case Text:
		for i, el := range outputElements {
			if i > 0 {
				fmt.Fprintln(cmd.OutOrStdout())
			}

			PrintElementText(cmd, el)
		}
		return nil

	case JSON:
		return PrintJSON(cmd, outputElements)

	case NDJSON:
		return PrintNDJSON(cmd, outputElements)

	default:
		return fmt.Errorf("unhandled output format %q", outputFormat)
	}
}

func PrintElementText(cmd *cobra.Command, el elementOutput) {
	fmt.Fprintf(cmd.OutOrStdout(), "Name:              %s\n", el.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "Kind:              %s\n", el.Kind)
	fmt.Fprintf(cmd.OutOrStdout(), "Source:            %s\n", el.Source)
	fmt.Fprintf(cmd.OutOrStdout(), "NORAD ID:          %d\n", el.NoradID)
	fmt.Fprintf(cmd.OutOrStdout(), "Object ID:         %s\n", el.ObjectID)
	fmt.Fprintf(cmd.OutOrStdout(), "Epoch UTC:         %s\n", el.EpochUTC)
	fmt.Fprintf(cmd.OutOrStdout(), "Inclination:       %.4f°\n", el.InclinationDeg)
	fmt.Fprintf(cmd.OutOrStdout(), "RAAN:              %.4f°\n", el.RAANDeg)
	fmt.Fprintf(cmd.OutOrStdout(), "Eccentricity:      %.7f\n", el.Eccentricity)
	fmt.Fprintf(cmd.OutOrStdout(), "Arg. pericenter:   %.4f°\n", el.ArgOfPericenterDeg)
	fmt.Fprintf(cmd.OutOrStdout(), "Mean anomaly:      %.4f°\n", el.MeanAnomalyDeg)
	fmt.Fprintf(cmd.OutOrStdout(), "Mean motion:       %.8f rev/day\n", el.MeanMotionRevPerDay)
	fmt.Fprintf(cmd.OutOrStdout(), "BSTAR:             %.8g\n", el.BStar)
}

func ParseFindAt(value string) (time.Time, error) {
	if value == "" {
		return time.Now().UTC(), nil
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"invalid --at value %q: expected RFC3339 format like 2026-06-03T22:00:00Z",
			value,
		)
	}

	return t.UTC(), nil
}

func toElementOutput(el satellite.GPElement) elementOutput {
	return elementOutput{
		Name:                el.ObjectName,
		Kind:                "satellite",
		Source:              "celestrak",
		NoradID:             el.NoradCatID,
		ObjectID:            el.ObjectID,
		EpochUTC:            normalizeCelesTrakEpoch(el.Epoch),
		InclinationDeg:      el.Inclination,
		RAANDeg:             el.RAOfAscNode,
		Eccentricity:        el.Eccentricity,
		ArgOfPericenterDeg:  el.ArgOfPericenter,
		MeanAnomalyDeg:      el.MeanAnomaly,
		MeanMotionRevPerDay: el.MeanMotion,
		BStar:               el.BStar,
		MeanMotionDot:       el.MeanMotionDot,
		MeanMotionDDot:      el.MeanMotionDDot,
		ElementSetNo:        el.ElementSetNo,
		RevAtEpoch:          el.RevAtEpoch,
	}
}

func normalizeCelesTrakEpoch(epoch string) string {
	if epoch == "" {
		return ""
	}

	// Try to parse as RFC3339
	t, err := time.Parse(time.RFC3339, epoch)
	if err != nil {
		// Try parsing without Z
		t, err = time.Parse("2006-01-02T15:04:05", epoch)
		if err != nil {
			// Return as-is if we can't parse
			return epoch
		}
	}

	// Always return in UTC with Z suffix
	return t.UTC().Format("2006-01-02T15:04:05Z")
}
