/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/ewencodes/fwd-mgmt/internal/config"
	"github.com/ewencodes/fwd-mgmt/internal/hosts"
	"github.com/ewencodes/fwd-mgmt/internal/logs"
	"github.com/ewencodes/fwd-mgmt/internal/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var cfgFile string
var debug bool

var rootCmd = &cobra.Command{
	Use:    "fwd-mgmt",
	Short:  "A brief description of your application",
	Long:   ``,
	PreRun: logs.ToggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		parsedConfig, err := config.NewConfig()

		if err != nil {
			log.Fatalf("Failed to parse config file: %s", err)
		}

		err = hosts.UpdateHosts()

		if err != nil {
			log.Fatalf("Failed to update hosts file: %s", err)
		}

		// Start the SSH agent and add the key
		agent, err := ssh.NewSSHAgent(parsedConfig.SSH.Key)
		if err != nil {
			log.Fatalf("Failed to start SSH agent: %s", err)
		}
		defer agent.Conn.Close()

		fmt.Println("Press Ctrl+C to exit")

		for _, forward := range parsedConfig.SSH.Tunnels {
			go ssh.StartForwardSession(forward.SSHHost, forward.SSHUser, forward.LocalHost, forward.LocalPort, forward.RemoteHost, forward.RemotePort, agent.Conn)
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		ssh.KillSSHAgent(agent.Pid)
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
