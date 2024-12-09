/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/ewencodes/fwd-mgmt/internal/hosts"
	"github.com/spf13/cobra"
)

// updateHostsCmd represents the updateHosts command
var updateHostsCmd = &cobra.Command{
	Use:   "update",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := hosts.UpdateHosts()

		if err != nil {
			return fmt.Errorf("Failed to update hosts file: %s", err)
		}

		fmt.Println("Hosts file updated.")
		return nil
	},
}

func init() {
	hostsCmd.AddCommand(updateHostsCmd)
}
