/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage ARSO CLI configuration",
	Long: `Manage local ARSO CLI configuration.

The config command controls settings such as API endpoints, default node IDs,
observer location, output format, authentication tokens, and development profiles.

Configuration values are used by other ARSO commands to know which observatory
node, backend service, or environment they should communicate with.

Examples:
  arso config get
  arso config set api.url http://localhost:8080
  arso config set node.default local`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("config called")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
