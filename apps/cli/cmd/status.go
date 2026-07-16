/*
Copyright © 2026 acortino <arso@acortino.me>
*/
package cmd

import (
	"github.com/openarso/arso/apps/cli/cmd/output"
	"github.com/openarso/arso/apps/internal/node"
	"github.com/spf13/cobra"
)

var outputStatus string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		findOutput = output.Normalize(outputStatus)

		if err := output.Validate(findOutput, output.Text, output.JSON, output.NDJSON); err != nil {
			return err
		}

		nodeService := node.NewService()
		status, err := nodeService.Status(cmd.Context())
		if err != nil {
			return err
		}

		return output.PrintNodeStatus(cmd, status, outputStatus)
	},
}

func init() {
	nodeCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVarP(
		&outputStatus,
		"output",
		"o",
		"text",
		"Output format: text, json or ndjson",
	)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
