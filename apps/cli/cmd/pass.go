/*
Copyright © 2026 acortino <arso@acortino.me>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var passCmd = &cobra.Command{
	Use:   "pass",
	Short: "Predict satellite passes for your observatory",
	Long: `Predict satellite passes for your configured observatory location.

Use the pass subcommands to list upcoming passes or to stop after the first pass
that meets your elevation threshold.

Examples:
  arso pass next ISS
  arso pass list HUBBLE --lookahead 7d`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pass called")
	},
}

func init() {
	rootCmd.AddCommand(passCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// passCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// passCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
