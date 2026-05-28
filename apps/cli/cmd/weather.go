/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// weatherCmd represents the weather command
var weatherCmd = &cobra.Command{
	Use:   "weather",
	Short: "Inspect weather and observing conditions",
	Long: `Inspect weather data and observing conditions for an ARSO node or location.

The weather command provides environmental information useful for astronomy,
remote observatory safety, and observation planning. This may include cloud cover,
temperature, humidity, wind, pressure, precipitation, and sky visibility.

In future versions, this command can help decide whether the observatory should
start, pause, or cancel an observation session.

Examples:
  arso weather current
  arso weather forecast
  arso weather safety`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("weather called")
	},
}

func init() {
	rootCmd.AddCommand(weatherCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// weatherCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// weatherCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
