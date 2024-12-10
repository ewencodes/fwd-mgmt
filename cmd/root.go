/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os/signal"
	"syscall"
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
			tunnels = append(tunnels, parsedConfig.SSH.GetTunnelsByTags(tags)...)
		} else {
			tunnels = append(tunnels, parsedConfig.SSH.Tunnels...)
		}

		if len(tunnels) == 0 {
			return fmt.Errorf("no tunnels found with tags: %s", tags)
		}

		for _, forward := range tunnels {

			// Start the SSH agent and add the key
			agent, err := ssh.NewSSHAgent(parsedConfig.SSH.Key)
			if err != nil {
				return fmt.Errorf("failed to start SSH agent: %s", err)
			}
			defer agent.Conn.Close()

			agentsPids = append(agentsPids, agent.Pid)

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

			go ssh.StartForwardSession(fmt.Sprintf("%s:%s", sshHost, sshPort), sshUser, forward.LocalHost, forward.LocalPort, forward.RemoteHost, forward.RemotePort, agent.Conn)
		}

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			for _, pid := range agentsPids {
				err := ssh.KillSSHAgent(pid)
				if err != nil {
					log.Infof("failed to kill SSH agent: %s", err)
				}
			}
			os.Exit(0)
		}()
		for {
			time.Sleep(1 * time.Second)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		for _, pid := range agentsPids {
			err := ssh.KillSSHAgent(pid)
			if err != nil {
				log.Errorf("Failed to kill SSH agent: %s", err)
			}
		}
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
