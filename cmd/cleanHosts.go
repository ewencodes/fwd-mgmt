/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/ewencodes/fwd-mgmt/internal/hosts"
	"github.com/spf13/cobra"
)

// cleanHostsCmd represents the cleanHosts command
var cleanHostsCmd = &cobra.Command{
	Use:   "clean",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := hosts.CleanHosts()

		if err != nil {
			return err
		}

		fmt.Println("Hosts file cleaned successfully")
		return nil
	},
}

func init() {
	hostsCmd.AddCommand(cleanHostsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cleanHostsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cleanHostsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
