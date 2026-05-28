/*
Copyright © 2026 acortino <arso@acortino.me>

*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"

	versionOutput string
)

type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the ARSO CLI version",
	Long: `Print the current ARSO CLI version and build information.

This command is useful for checking which version of the CLI is installed,
debugging local environments, and reporting issues with reproducible context.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		info := versionInfo{
			Version: Version,
			Commit:  Commit,
			Date:    Date,
		}

		switch versionOutput {
		case "text":
			fmt.Fprintf(cmd.OutOrStdout(), "ARSO %s\n", info.Version)
			fmt.Fprintf(cmd.OutOrStdout(), "Commit: %s\n", info.Commit)
			fmt.Fprintf(cmd.OutOrStdout(), "Built:  %s\n", info.Date)
			return nil

		case "json":
			encoded, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
			return nil

		default:
			return fmt.Errorf("unsupported output format %q, expected one of: text, json", versionOutput)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().StringVarP(
		&versionOutput,
		"output",
		"o",
		"text",
		"Output format: text or json",
	)
}
