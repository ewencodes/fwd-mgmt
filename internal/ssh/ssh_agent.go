package ssh

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SSHAgent struct {
	Conn net.Conn
	Pid  string
}

func NewSSHAgent(privateKeyPath string) (*SSHAgent, error) {
	cmd := exec.Command("ssh-agent")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to start ssh-agent: %w", err)
	}

	sshAuthSock, sshAgentPid := parseOutput(string(output))

	if err := os.Setenv("SSH_AUTH_SOCK", sshAuthSock); err != nil {
		return nil, err
	}
	if err := os.Setenv("SSH_AGENT_PID", sshAgentPid); err != nil {
		return nil, err
	}

	// Connect to the SSH agent
	agentSock := os.Getenv("SSH_AUTH_SOCK")
	agentConn, err := net.Dial("unix", agentSock)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH agent: %w", err)
	}

	// Add the private key to the agent
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	signer, err := ssh.ParseRawPrivateKey(key)
	if err != nil {
		// If the key is encrypted, prompt for the passphrase
		if _, ok := err.(*ssh.PassphraseMissingError); ok {
			fmt.Print("Enter passphrase for key: ")
			passphrase, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			passphrase = strings.TrimSpace(passphrase)

			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key with passphrase: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	agentClient := agent.NewClient(agentConn)
	if err := agentClient.Add(agent.AddedKey{PrivateKey: signer}); err != nil {
		return nil, fmt.Errorf("failed to add key to agent: %w", err)
	}

	return &SSHAgent{
		Conn: agentConn,
		Pid:  sshAgentPid,
	}, nil
}

func parseOutput(output string) (string, string) {
	var sshAuthSock, sshAgentPid string

	// Split the output into lines
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Check for SSH_AUTH_SOCK
		if strings.HasPrefix(line, "SSH_AUTH_SOCK=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				sshAuthSock = strings.TrimSpace(strings.Split(parts[1], ";")[0])
			}
		}
		// Check for SSH_AGENT_PID
		if strings.HasPrefix(line, "SSH_AGENT_PID=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				sshAgentPid = strings.TrimSpace(strings.Split(parts[1], ";")[0])
			}
		}
	}

	return sshAuthSock, sshAgentPid
}

func KillSSHAgent(agentPIDStr string) error {
	// Get the SSH_AGENT_PID from the environment variable
	// agentPIDStr := os.Getenv("SSH_AGENT_PID")
	// if agentPIDStr == "" {
	// 	return fmt.Errorf("SSH_AGENT_PID environment variable not set")
	// }

	// Convert the PID to an integer
	agentPID, err := strconv.Atoi(agentPIDStr)
	if err != nil {
		return fmt.Errorf("invalid SSH_AGENT_PID: %s", err)
	}

	// Kill the SSH agent process
	cmd := exec.Command("kill", "-9", strconv.Itoa(agentPID))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill process with PID %d: %s", agentPID, err)
	}

	// // Optionally, unset the environment variable
	// os.Unsetenv("SSH_AGENT_PID")
	// os.Unsetenv("SSH_AUTH_SOCK")

	return nil
}
