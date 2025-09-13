/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"time"

	"github.com/ewencodes/fwd-mgmt/internal/config"
	"github.com/ewencodes/fwd-mgmt/internal/hosts"
	"github.com/ewencodes/fwd-mgmt/internal/logs"
	"github.com/ewencodes/fwd-mgmt/internal/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

var cfgFile string
var debug bool
var agentsPids []string = make([]string, 0)
var tags []string

var rootCmd = &cobra.Command{
	Use:    "fwd-mgmt",
	Short:  "A brief description of your application",
	Long:   ``,
	PreRun: logs.ToggleDebug,
	RunE: func(cmd *cobra.Command, args []string) error {
		parsedConfig, err := config.NewConfig()

		if err != nil {

			return fmt.Errorf("failed to parse config file: %s", err)
		}

		err = hosts.UpdateHosts()

		if err != nil {
			return fmt.Errorf("failed to update hosts file: %s", err)
		}
		fmt.Println("Press Ctrl+C to exit")

		tunnels := make([]config.SSHTunnel, 0)

		if len(tags) > 0 {
			log.Debugf("tags: %v", tags)
			tunnels = append(tunnels, parsedConfig.SSH.GetTunnelsByTags(tags)...)
		} else {
			log.Debugf("no tags: %v", tags)
			tunnels = append(tunnels, parsedConfig.SSH.Tunnels...)
		}

		if len(tunnels) == 0 {
			return fmt.Errorf("no tunnels found with tags: %s", tags)
		}

		for _, forward := range tunnels {
			if parsedConfig.SSH.DefaultUser == "" && forward.SSHUser == "" {
				return fmt.Errorf("no default user set in config or tunnel for %s", forward.RemoteHost)
			}

			sshUser := parsedConfig.SSH.DefaultUser

			if forward.SSHUser != "" {
				sshUser = forward.SSHUser
			}

			sshHost := parsedConfig.SSH.DefaultHost

			if forward.SSHHost != "" {
				sshHost = forward.SSHHost
			}

			sshPort := forward.SSHPort

			if sshPort == "" {
				sshPort = parsedConfig.SSH.DefaultPort
			}

			log.Debugf("ssh user: %s, ssh port: %s", sshUser, sshPort)

			go func() {
				err := ssh.StartForwardSession(fmt.Sprintf("%s:%s", sshHost, sshPort), sshUser, forward.LocalHost, forward.LocalPort, forward.RemoteHost, forward.RemotePort, parsedConfig.SSH.PrivayeKeyPath)
				if err != nil {
					log.Fatalf("failed to start ssh session: %s", err)
				}
			}()
		}

		for {
			time.Sleep(1 * time.Second)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fwd-mgmt.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "verbose logging")
	rootCmd.PersistentFlags().StringArrayVarP(&tags, "tag", "t", make([]string, 0), "tags to filter")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".fwd-mgmt" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".fwd-mgmt")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
