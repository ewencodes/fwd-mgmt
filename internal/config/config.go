package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	SSH SSH `mapstructure:"ssh"`
}

type SSH struct {
	Tunnels []SSHTunnel `mapstructure:"tunnels"`
	Key     string      `mapstructure:"private_key"`
}

type SSHTunnel struct {
	RemoteHost string `mapstructure:"remote_host"`
	RemotePort string `mapstructure:"remote_port"`
	LocalPort  string `mapstructure:"local_port"`
	LocalHost  string `mapstructure:"local_host"`
	SSHHost    string `mapstructure:"ssh_host"`
	SSHUser    string `mapstructure:"ssh_user"`
}

func NewConfig() (*Config, error) {
	var config Config

	err := viper.Unmarshal(&config)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	return &config, nil
}
