/*
Copyright © 2026 acortino <arso@acortino.me>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage and inspect ARSO observatory nodes",
	Long: `Manage and inspect ARSO observatory nodes.

A node represents an observatory device or station capable of collecting data,
running sensors, controlling hardware, or communicating with the ARSO network.

This command can be used to check node status, list known nodes, inspect hardware
capabilities, and later connect multiple observatories into a distributed mesh.

Examples:
  arso node status
  arso node list
  arso node info`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("node called")
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
