package ssh

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// Alternative SSH config that doesn't rely on SSH agent
// This is useful when SSH agent is not available on Windows
func getDirectSSHConfig(user string, privateKeyPath string) (*ssh.ClientConfig, error) {
	// Read the private key directly
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	// Parse the private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		// If the key is encrypted, try to get passphrase
		if _, ok := err.(*ssh.PassphraseMissingError); ok {
			fmt.Print("Enter passphrase for key: ")
			var passphrase string
			fmt.Scanln(&passphrase)

			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key with passphrase: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}, nil
}

// Alternative function to create SSH config without agent
// Use this as fallback when SSH agent is not available
func getSSHConfigWithoutAgent(user string, privateKeyPath string) (*ssh.ClientConfig, error) {
	return getDirectSSHConfig(user, privateKeyPath)
}
