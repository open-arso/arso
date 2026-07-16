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

var findOutput string
var findAt string
var findElements bool

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

		findOutput = output.Normalize(findOutput)

		if err := output.Validate(findOutput, output.Text, output.JSON, output.NDJSON); err != nil {
			return err
		}

		client := satellite.NewClient()

		if findElements {
			elements, err := client.Elements(cmd.Context(), target)
			if err != nil {
				return err
			}

			if len(elements) == 0 {
				return fmt.Errorf("no orbital elements found for %q", target)
			}

			return output.PrintElements(cmd, elements, findOutput)
		}

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

		findAtTime, err := output.ParseFindAt(findAt)
		if err != nil {
			return err
		}

		positions, err := client.Locate(cmd.Context(), target, observer, findAtTime)
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

				positions, err = client.Locate(
					cmd.Context(),
					strconv.Itoa(selected.NoradID),
					observer,
					findAtTime,
				)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		if len(positions) == 0 {
			return fmt.Errorf("no object found for %q", target)
		}

		return output.PrintApparentPositions(cmd, positions, findOutput)
	},
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

	findCmd.Flags().StringVarP(
		&findAt,
		"at",
		"a",
		"",
		"Observation time in RFC3339 format, for example 2026-06-03T22:00:00Z. Defaults to now.",
	)

	findCmd.Flags().BoolVarP(
		&findElements,
		"elements",
		"e",
		false,
		"Show orbital elements instead of apparent observer position",
	)

}
