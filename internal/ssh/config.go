package ssh

import (
	"net"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func getSSHConfig(user string, agentConn net.Conn) (*ssh.ClientConfig, error) {
	sshAgent := agent.NewClient(agentConn)

	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(sshAgent.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}
