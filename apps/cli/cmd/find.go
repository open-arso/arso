/*
Copyright © 2026 acortino <arso@acortino.me>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/openarso/arso/apps/cli/internal/clioutput"
	"github.com/openarso/arso/apps/cli/internal/satellite"
	"github.com/openarso/arso/apps/cli/internal/appconfig"
	"github.com/spf13/cobra"
)

var findOutput string

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find <target>",
	Short: "Find orbital objects",
	Long: `Find orbital objects from the command line.

The find command searches for satellites and other Earth-orbiting objects using
external orbital data.

For now, this command supports simple target resolution such as ISS aliases,
NORAD catalog IDs, and object name searches.

Examples:
  arso find ISS
  arso find 25544
  arso find HUBBLE
  arso find ISS --output json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
    	target := args[0]

		findOutput = clioutput.Normalize(findOutput)

		if err := clioutput.Validate(findOutput, clioutput.Text, clioutput.JSON, clioutput.NDJSON); err != nil {
			return err
		}

    	cfg, err := appconfig.Load()
    	if err != nil {
    		return err
    	}

		if err := cfg.Observatory.RequireConfigured(); err != nil {
			return err
		}

    	observer := satellite.Observer{
    		Name:            cfg.Node.Name,
    		LatitudeDeg:     *cfg.Observatory.Latitude,
    		LongitudeDeg:    *cfg.Observatory.Longitude,
    		ElevationMeters: cfg.Observatory.ElevationMeters,
    	}

    	client := satellite.NewClient()

    	positions, err := client.Locate(cmd.Context(), target, observer, time.Now().UTC())
    	if err != nil {
    		return err
    	}

    	if len(positions) == 0 {
    		return fmt.Errorf("no object found for %q", target)
    	}

    	switch findOutput {
		case clioutput.Text:
    		printApparentPositionText(cmd, positions[0])
    		return nil

    	case clioutput.JSON:
    		encoded, err := json.MarshalIndent(positions, "", "  ")
    		if err != nil {
    			return err
    		}

    		fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
    		return nil

    	case clioutput.NDJSON:
    		for _, position := range positions {
    			encoded, err := json.Marshal(position)
    			if err != nil {
    				return err
    			}

    			fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
    		}

    		return nil

    	default:
			return fmt.Errorf("unhandled output format %q", findOutput)
    	}
    },
}

func printApparentPositionText(cmd *cobra.Command, position satellite.ApparentPosition) {
	fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", position.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "Kind:        %s\n", position.Kind)
	fmt.Fprintf(cmd.OutOrStdout(), "NORAD ID:    %d\n", position.NoradID)
	fmt.Fprintf(cmd.OutOrStdout(), "Object ID:   %s\n", position.ObjectID)
	fmt.Fprintf(cmd.OutOrStdout(), "Observer:    %s\n", position.ObserverName)
	fmt.Fprintf(cmd.OutOrStdout(), "Time UTC:    %s\n", position.TimeUTC)
	fmt.Fprintf(cmd.OutOrStdout(), "Azimuth:     %.2f°\n", position.AzimuthDeg)
	fmt.Fprintf(cmd.OutOrStdout(), "Elevation:   %.2f°\n", position.ElevationDeg)
	fmt.Fprintf(cmd.OutOrStdout(), "Range:       %.2f km\n", position.RangeKm)
	fmt.Fprintf(cmd.OutOrStdout(), "Range rate:  %.4f km/s\n", position.RangeRateKms)
	fmt.Fprintf(cmd.OutOrStdout(), "Above horizon: %t\n", position.AboveHorizon)
	fmt.Fprintf(cmd.OutOrStdout(), "Subpoint:    %.4f°, %.4f°\n", position.SatelliteLatitudeDeg, position.SatelliteLongitudeDeg)
	fmt.Fprintf(cmd.OutOrStdout(), "Altitude:    %.2f km\n", position.SatelliteAltitudeKm)
}

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().StringVarP(
		&findOutput,
		"output",
		"o",
		"text",
		"Output format: text, json or ndjson",
	)
}
