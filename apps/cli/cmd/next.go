/*
Copyright © 2026 acortino <arso@acortino.me>
*/
package cmd

import (
	"errors"
	"fmt"

	"github.com/openarso/arso/apps/cli/internal/appconfig"
	"github.com/openarso/arso/apps/cli/internal/clioutput"
	"github.com/openarso/arso/apps/cli/internal/satellite"
	"github.com/spf13/cobra"
)

var fromTime string
var lookahead string
var minElevation int
var outputNext string

var nextCmd = &cobra.Command{
	Use:   "next TARGET",
	Args:  cobra.ExactArgs(1),
	Short: "Show the next satellite pass over your observatory",
	Long: `Show the next predicted satellite pass for a target above your configured observatory location.

	The command uses the observatory latitude, longitude, and elevation from your ARSO config,
	then searches from the given start time until the end of the lookahead window.

	By default, it returns the first pass whose maximum elevation is at least 10 degrees.

	Examples:
	  arso pass next ISS
	  arso pass next ISS --lookahead 48h
	  arso pass next ISS --min-elevation 20
	  arso pass next ISS --from 2026-06-10T22:00:00Z
	  arso pass next ISS --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		findOutput = clioutput.Normalize(outputNext)

		if err := clioutput.Validate(normalizedOutput, clioutput.Text, clioutput.JSON, clioutput.NDJSON); err != nil {
			return err
		}

		client := newSatelliteClient()

		cfg, err := loadConfig()
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

		nextPass, err := client.NextPass(cmd.Context(), target, observer, fromTime, lookahead, minElevation)
		if err != nil {
			var ambiguousErr *satellite.AmbiguousTargetError

			if errors.As(err, &ambiguousErr) {
				selected, selectErr := selectResolvedTarget(
					cmd.InOrStdin(),
					cmd.ErrOrStderr(),
					ambiguousErr.Candidates,
				)
				if selectErr != nil {
					return selectErr
				}

				if cacheErr := client.CacheResolvedTarget(target, selected); cacheErr != nil {
					fmt.Fprintf(
						cmd.ErrOrStderr(),
						"Warning: could not cache selected target: %v\n",
						cacheErr,
					)
				}

				nextPass, err = client.NextPass(cmd.Context(), target, observer, fromTime, lookahead, minElevation)

				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return printPassPredictions(cmd, nextPass, outputNext)
	},
}

func init() {
	passCmd.AddCommand(nextCmd)

	nextCmd.Flags().StringVarP(
		&outputNext,
		"output",
		"o",
		"text",
		"Output format: text, json or ndjson",
	)

	nextCmd.Flags().StringVarP(
		&fromTime,
		"from",
		"f",
		"",
		"Start search time in RFC3339 format, for example 2026-06-09T17:00:00Z. Default to now.",
	)

	nextCmd.Flags().StringVarP(
		&lookahead,
		"lookahead",
		"l",
		"48h",
		"Maximum search window, for example 1h, 48h, 3d. Default: 48h.",
	)

	nextCmd.Flags().IntVarP(
		&minElevation,
		"min-elevation",
		"e",
		10,
		"Minimum maximum elevation required for a pass, in degrees. Default: 10.",
	)

}
