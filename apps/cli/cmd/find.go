/*
Copyright © 2026 acortino <arso@acortino.me>
*/
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/openarso/arso/apps/cli/internal/appconfig"
	"github.com/openarso/arso/apps/cli/internal/clioutput"
	"github.com/openarso/arso/apps/cli/internal/satellite"
	"github.com/spf13/cobra"
)

var findOutput string
var findAt string
var findElements bool

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

		normalizedOutput := clioutput.Normalize(findOutput)

		if err := clioutput.Validate(normalizedOutput, clioutput.Text, clioutput.JSON, clioutput.NDJSON); err != nil {
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

			return printElements(cmd, elements, normalizedOutput)
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

		findAtTime, err := parseFindAt(findAt)
		if err != nil {
			return err
		}

		positions, err := client.Locate(cmd.Context(), target, observer, findAtTime)
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

		return printApparentPositions(cmd, positions, normalizedOutput)
	},
}

func selectResolvedTarget(
	in io.Reader,
	out io.Writer,
	candidates []satellite.ResolvedTarget,
) (satellite.ResolvedTarget, error) {
	if len(candidates) == 0 {
		return satellite.ResolvedTarget{}, fmt.Errorf("no candidates to select")
	}

	fmt.Fprintln(out, "Several satellites match your target:")
	fmt.Fprintln(out)

	for i, candidate := range candidates {
		fmt.Fprintf(
			out,
			"%d) %s — NORAD %d — Object ID %s\n",
			i+1,
			candidate.Name,
			candidate.NoradID,
			candidate.ObjectID,
		)
	}

	fmt.Fprintln(out)

	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, "Select satellite [1-%d]: ", len(candidates))

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return satellite.ResolvedTarget{}, fmt.Errorf("read selection: %w", err)
			}

			return satellite.ResolvedTarget{}, io.EOF
		}

		rawInput := strings.TrimSpace(scanner.Text())

		selectedIndex, err := strconv.Atoi(rawInput)
		if err != nil {
			fmt.Fprintln(out, "Please enter a valid number.")
			continue
		}

		if selectedIndex < 1 || selectedIndex > len(candidates) {
			fmt.Fprintf(out, "Please enter a number between 1 and %d.\n", len(candidates))
			continue
		}

		return candidates[selectedIndex-1], nil
	}
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
