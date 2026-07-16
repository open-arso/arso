/*
Copyright © 2026 acortino <arso@acortino.me>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/openarso/arso/apps/internal/config"
	"github.com/spf13/cobra"
)

var configOverwrite bool

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
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default ARSO config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.Init(configOverwrite)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created config at %s\n", path)
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the ARSO config file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.Path()
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), path)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current ARSO configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		encoded, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a config value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value, err := config.Get(args[0])
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%v\n", value)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		path, err := config.Set(key, value)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Set %s=%s\n", key, value)
		fmt.Fprintf(cmd.OutOrStdout(), "Saved config at %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)

	configInitCmd.Flags().BoolVar(
		&configOverwrite,
		"overwrite",
		false,
		"Overwrite the config file if it already exists",
	)
}
