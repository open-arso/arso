package output

import (
	"fmt"

	"github.com/openarso/arso/apps/internal/satellite"
	"github.com/spf13/cobra"
)

func PrintPassPredictions(cmd *cobra.Command, result satellite.PassPredictionResult, outputFormat string) error {
	switch outputFormat {
	case Text:
		printPassPredictionResultText(cmd, result)
		for i, pass := range result.Passes {
			if i > 0 {
				fmt.Fprintln(cmd.OutOrStdout())
			}

			printPredictionText(cmd, pass)
		}
		return nil

	case JSON:
		return PrintJSON(cmd, result)

	case NDJSON:
		return PrintNDJSON(cmd, result.Passes)

	default:
		return fmt.Errorf("unhandled output format %q", outputFormat)
	}
}

func printPassPredictionResultText(cmd *cobra.Command, prediction satellite.PassPredictionResult) {
	out := cmd.OutOrStdout()

	const labelWidth = 30

	fmt.Fprintf(out, "%-*s %s\n", labelWidth, "Name:", prediction.Name)
	fmt.Fprintf(out, "%-*s %s\n", labelWidth, "Kind:", prediction.Kind)
	fmt.Fprintf(out, "%-*s %d\n", labelWidth, "NORAD ID:", prediction.NoradID)
	fmt.Fprintf(out, "%-*s %s\n", labelWidth, "Object ID:", prediction.ObjectID)
	fmt.Fprintf(out, "%-*s %s\n", labelWidth, "Observer:", prediction.ObserverName)
}

func printPredictionText(cmd *cobra.Command, prediction satellite.PredictedPass) {
	out := cmd.OutOrStdout()

	const labelWidth = 30
	const timeFormat = "2006-01-02 15:04:05 UTC"

	fmt.Fprintf(out, "%-*s %s\n", labelWidth, "Acquisition of signal (AOS):", prediction.AcquisitionOfSignal.UTC().Format(timeFormat))
	fmt.Fprintf(out, "%-*s %s\n", labelWidth, "Loss of signal (LOS):", prediction.LossOfSignal.UTC().Format(timeFormat))
	fmt.Fprintf(out, "%-*s %v\n", labelWidth, "Duration:", prediction.Duration)
	fmt.Fprintf(out, "%-*s %.1f°\n", labelWidth, "Maximum elevation:", prediction.MaxElevation)
	fmt.Fprintf(out, "%-*s %s\n", labelWidth, "Maximum elevation time:", prediction.MaxElevationTime.UTC().Format(timeFormat))
	fmt.Fprintf(out, "%-*s %.1f°\n", labelWidth, "Azimuth at AOS:", prediction.AzimuthAtAOS)
	fmt.Fprintf(out, "%-*s %.1f°\n", labelWidth, "Azimuth at LOS:", prediction.AzimuthAtLOS)
}
