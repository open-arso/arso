/*
Copyright © 2026 acortino <arso@acortino.me>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/openarso/arso/apps/cli/cmd/output"
	"github.com/spf13/cobra"
)

var (
	// Version is the CLI version injected at build time.
	Version = "dev"
	// Commit is the source control revision injected at build time.
	Commit = "none"
	// Date is the build timestamp injected at build time.
	Date = "unknown"

	versionOutput string
)

type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

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

		versionOutput = output.Normalize(versionOutput)

		if err := output.Validate(versionOutput, output.Text, output.JSON); err != nil {
			return err
		}

		switch versionOutput {
		case output.Text:
			fmt.Fprintf(cmd.OutOrStdout(), "ARSO %s\n", info.Version)
			fmt.Fprintf(cmd.OutOrStdout(), "Commit: %s\n", info.Commit)
			fmt.Fprintf(cmd.OutOrStdout(), "Built:  %s\n", info.Date)
			return nil

		case output.JSON:
			encoded, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
			return nil

		default:
			return fmt.Errorf("unhandled output format %q", versionOutput)
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
