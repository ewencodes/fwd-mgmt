package config

import (
	"fmt"

	"slices"

	"github.com/spf13/viper"
)

type Config struct {
	SSH SSH `mapstructure:"ssh"`
}

type SSH struct {
	Tunnels        []SSHTunnel `mapstructure:"tunnels"`
	PrivayeKeyPath string      `mapstructure:"private_key"`
	DefaultUser    string      `mapstructure:"default_ssh_user"`
	DefaultPort    string      `mapstructure:"default_ssh_port"`
	DefaultHost    string      `mapstructure:"default_ssh_host"`
}

type SSHTunnel struct {
	RemoteHost string   `mapstructure:"remote_host"`
	RemotePort string   `mapstructure:"remote_port"`
	LocalPort  string   `mapstructure:"local_port"`
	LocalHost  string   `mapstructure:"local_host"`
	SSHHost    string   `mapstructure:"ssh_host"`
	SSHUser    string   `mapstructure:"ssh_user"`
	SSHPort    string   `mapstructure:"ssh_port"`
	Tags       []string `mapstructure:"tags"`
}

func NewConfig() (*Config, error) {
	var config Config

	err := viper.Unmarshal(&config)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	if config.SSH.DefaultPort == "" {
		config.SSH.DefaultPort = "22"
	}

	return &config, nil
}

func (s *SSH) GetTunnelsByTags(tags []string) []SSHTunnel {
	var tunnels []SSHTunnel

	for _, tunnel := range s.Tunnels {
		if containsAll(tunnel.Tags, tags) {
			tunnels = append(tunnels, tunnel)
		}
	}

	return tunnels
}

func containsAll(slice, tags []string) bool {
	for _, tag := range tags {
		if !slices.Contains(slice, tag) {
			return false
		}
	}

	return true
}
