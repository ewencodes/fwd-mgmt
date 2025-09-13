package hosts

import (
	"fmt"

	"github.com/ewencodes/fwd-mgmt/internal/config"
	"github.com/txn2/txeh"
)

func UpdateHosts() error {
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		return fmt.Errorf("failed to read hosts file: %s", err)
	}

	parsedConfig, err := config.NewConfig()

	if err != nil {
		return fmt.Errorf("failed to parse config file: %s", err)
	}

	for _, forward := range parsedConfig.SSH.Tunnels {
		hosts.AddHost("127.0.0.1", forward.LocalHost)
	}

	err = hosts.Save()
	if err != nil {
		return fmt.Errorf("failed to save hosts file: %s", err)
	}

	return nil
}

func CleanHosts() error {
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		return fmt.Errorf("failed to read hosts file: %s", err)
	}

	parsedConfig, err := config.NewConfig()

	if err != nil {
		return fmt.Errorf("failed to parse config file: %s", err)
	}

	for _, forward := range parsedConfig.SSH.Tunnels {
		hosts.RemoveHost(forward.LocalHost)
	}

	err = hosts.Save()

	if err != nil {
		return fmt.Errorf("failed to save hosts file: %s", err)
	}

	return nil
}
