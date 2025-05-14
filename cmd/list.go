/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/ewencodes/fwd-mgmt/internal/config"
	"github.com/ewencodes/fwd-mgmt/internal/logs"
	"github.com/spf13/cobra"

	"fmt"
)

var listCmd = &cobra.Command{
	Use:    "list",
	Short:  "A brief description of your application",
	Long:   ``,
	PreRun: logs.ToggleDebug,
	RunE: func(cmd *cobra.Command, args []string) error {
		parsedConfig, err := config.NewConfig()

		if err != nil {
			return fmt.Errorf("failed to parse config file: %s", err)
		}

		tunnels := parsedConfig.SSH.GetTunnelsByTags(tags)

		if len(tunnels) == 0 {
			return fmt.Errorf("no tunnels found with tags: %s", tags)
		}

		for _, forward := range tunnels {
			fmt.Printf("- %s:%s to %s:%s\n", forward.LocalHost, forward.LocalPort, forward.RemoteHost, forward.RemotePort)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
