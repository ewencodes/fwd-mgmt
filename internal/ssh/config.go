package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
)

func getSSHConfig(user string, privateKeyPath string) (*ssh.ClientConfig, error) {
	file, err := os.ReadFile(privateKeyPath)

	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %s", err)
	}

	privateKey, err := ssh.ParsePrivateKey(file)

	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %s", err)
	}

	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}
