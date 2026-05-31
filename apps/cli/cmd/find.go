/*
Copyright © 2026 acortino <arso@acortino.me>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/openarso/arso/apps/cli/internal/satellite"
	"github.com/spf13/cobra"
)

type findResult struct {
	Name            string  `json:"name"`
	NoradID         int     `json:"norad_id"`
	ObjectID        string  `json:"object_id"`
	Epoch           string  `json:"epoch"`
	Inclination     float64 `json:"inclination"`
	MeanMotion      float64 `json:"mean_motion"`
	Eccentricity    float64 `json:"eccentricity"`
	RAOfAscNode     float64 `json:"ra_of_asc_node"`
	ArgOfPericenter float64 `json:"arg_of_pericenter"`
	MeanAnomaly     float64 `json:"mean_anomaly"`
}

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

		queryKey, queryValue, err := satellite.ResolveTarget(target)
		if err != nil {
			return err
		}

		client := satellite.NewClient()

		elements, err := client.Fetch(cmd.Context(), queryKey, queryValue)
		if err != nil {
			return err
		}

		if len(elements) == 0 {
			return fmt.Errorf("no object found for %q", target)
		}

		switch findOutput {
		case "text":
			printElementText(cmd, elements[0])
			return nil

		case "json":
			results := toFindResults(elements)

			encoded, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
			return nil

		case "ndjson":
			for _, element := range elements {
				result := toFindResult(element)

				encoded, err := json.Marshal(result)
				if err != nil {
					return err
				}

				fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
			}

			return nil		
		default:
			return fmt.Errorf("unsupported output format %q, expected one of: text, json, ndjson", findOutput)
		}
	},
}

func printElementText(cmd *cobra.Command, element satellite.GPElement) {
	fmt.Fprintf(cmd.OutOrStdout(), "Name:         %s\n", element.ObjectName)
	fmt.Fprintf(cmd.OutOrStdout(), "NORAD ID:     %d\n", element.NoradCatID)
	fmt.Fprintf(cmd.OutOrStdout(), "Object ID:    %s\n", element.ObjectID)
	fmt.Fprintf(cmd.OutOrStdout(), "Epoch:        %s\n", element.Epoch)
	fmt.Fprintf(cmd.OutOrStdout(), "Inclination:  %.4f°\n", element.Inclination)
	fmt.Fprintf(cmd.OutOrStdout(), "Mean motion:  %.8f rev/day\n", element.MeanMotion)
	fmt.Fprintf(cmd.OutOrStdout(), "Eccentricity: %.8f\n", element.Eccentricity)
}

func toFindResult(element satellite.GPElement) findResult {
	return findResult{
		Name:            element.ObjectName,
		NoradID:         element.NoradCatID,
		ObjectID:        element.ObjectID,
		Epoch:           element.Epoch,
		Inclination:     element.Inclination,
		MeanMotion:      element.MeanMotion,
		Eccentricity:    element.Eccentricity,
		RAOfAscNode:     element.RAOfAscNode,
		ArgOfPericenter: element.ArgOfPericenter,
		MeanAnomaly:     element.MeanAnomaly,
	}
}

func toFindResults(elements []satellite.GPElement) []findResult {
	results := make([]findResult, 0, len(elements))

	for _, element := range elements {
		results = append(results, toFindResult(element))
	}

	return results
}

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().StringVarP(
		&findOutput,
		"output",
		"o",
		"text",
		"Output format: text or json or ndjson",
	)
}
