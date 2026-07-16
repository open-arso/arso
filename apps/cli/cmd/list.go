/*
Copyright © 2026 acortino <arso@acortino.me>
*/
package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/openarso/arso/apps/cli/cmd/output"
	"github.com/openarso/arso/apps/internal/config"
	"github.com/openarso/arso/apps/internal/satellite"
	"github.com/spf13/cobra"
)

var fromTimeList string
var lookaheadList string
var minElevationList int
var outputList string

var listCmd = &cobra.Command{
	Use:   "list TARGET",
	Args:  cobra.ExactArgs(1),
	Short: "List satellite passes over your observatory",
	Long: `List predicted satellite passes for a target above your configured observatory location.

	The command uses the observatory latitude, longitude, and elevation from your ARSO config,
	then searches from the given start time until the end of the lookahead window.

	By default, it returns all passes whose maximum elevation is at least 10 degrees.

	Examples:
	  arso pass list ISS
	  arso pass list ISS --lookahead 48h
	  arso pass list ISS --lookahead 7d
	  arso pass list ISS --min-elevation 20
	  arso pass list ISS --from 2026-06-10T22:00:00Z
	  arso pass list ISS --output json`,

	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		findOutput = output.Normalize(outputList)

		if err := output.Validate(findOutput, output.Text, output.JSON, output.NDJSON); err != nil {
			return err
		}

		client := satellite.NewClient()

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if err := cfg.Observatory.RequireConfigured(); err != nil {
			return err
		}

		observer := config.Observer{
			Name:            cfg.Node.Name,
			LatitudeDeg:     *cfg.Observatory.Latitude,
			LongitudeDeg:    *cfg.Observatory.Longitude,
			ElevationMeters: cfg.Observatory.ElevationMeters,
		}

		nextPassPredictions, err := client.PassPredictions(cmd.Context(), target, observer, fromTimeList, lookaheadList, minElevationList)
		if err != nil {
			var ambiguousErr *satellite.AmbiguousTargetError

			if errors.As(err, &ambiguousErr) {
				selected, selectErr := SelectResolvedTarget(
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

				target = strconv.Itoa(selected.NoradID)
				client.PassPredictions(cmd.Context(), target, observer, fromTimeList, lookaheadList, minElevationList)

				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		return output.PrintPassPredictions(cmd, nextPassPredictions, outputList)
	},
}

func init() {
	passCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(
		&outputList,
		"output",
		"o",
		"text",
		"Output format: text, json or ndjson",
	)

	listCmd.Flags().StringVarP(
		&fromTimeList,
		"from",
		"f",
		"",
		"Start search time in RFC3339 format, for example 2026-06-09T17:00:00Z. Default to now.",
	)

	listCmd.Flags().StringVarP(
		&lookaheadList,
		"lookahead",
		"l",
		"48h",
		"Maximum search window. Default to 48h",
	)

	listCmd.Flags().IntVarP(
		&minElevationList,
		"min-elevation",
		"e",
		10,
		"Minimum maximum elevation required for a pass, in degrees. Default: 10.",
	)

}
