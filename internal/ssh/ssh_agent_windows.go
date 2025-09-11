//go:build windows

package ssh

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Windows-specific SSH agent implementation
func NewSSHAgent(privateKeyPath string) (*SSHAgent, error) {
	// Try to use existing SSH agent first
	agentSock := os.Getenv("SSH_AUTH_SOCK")
	if agentSock != "" {
		// Try to connect to existing agent
		if conn, err := connectToExistingAgent(agentSock, privateKeyPath); err == nil {
			return conn, nil
		}
	}

	// Try to start SSH agent on Windows
	return startWindowsSSHAgent(privateKeyPath)
}

func connectToExistingAgent(agentSock string, privateKeyPath string) (*SSHAgent, error) {
	// On Windows, SSH_AUTH_SOCK might be a named pipe or TCP connection
	var conn net.Conn
	var err error

	// Try different connection methods for Windows
	if strings.HasPrefix(agentSock, "\\\\") {
		// Named pipe connection (not directly supported by net.Dial)
		return nil, fmt.Errorf("named pipe connections not yet supported")
	} else if strings.Contains(agentSock, ":") {
		// TCP connection (some Windows SSH agents use this)
		conn, err = net.Dial("tcp", agentSock)
	} else {
		// Try as Unix socket (works in WSL or with newer Windows versions)
		conn, err = net.Dial("unix", agentSock)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to existing SSH agent: %w", err)
	}

	// Add the private key to the existing agent
	if err := addKeyToAgent(conn, privateKeyPath); err != nil {
		conn.Close()
		return nil, err
	}

	return &SSHAgent{
		Conn: conn,
		Pid:  "", // We don't own this agent
	}, nil
}

func startWindowsSSHAgent(privateKeyPath string) (*SSHAgent, error) {
	// Try different SSH agent implementations available on Windows

	// 1. Try OpenSSH for Windows
	if agent, err := tryOpenSSHAgent(privateKeyPath); err == nil {
		return agent, nil
	}

	// 2. Try Git Bash SSH agent
	if agent, err := tryGitBashSSHAgent(privateKeyPath); err == nil {
		return agent, nil
	}

	// 3. Fallback: return error with helpful message
	return nil, fmt.Errorf("no SSH agent available on Windows. Please install OpenSSH for Windows or Git for Windows")
}

func tryOpenSSHAgent(privateKeyPath string) (*SSHAgent, error) {
	// Check if OpenSSH ssh-agent is available
	cmd := exec.Command("ssh-agent", "-s")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("OpenSSH ssh-agent not available: %w", err)
	}

	sshAuthSock, sshAgentPid := parseOutput(string(output))

	if err := os.Setenv("SSH_AUTH_SOCK", sshAuthSock); err != nil {
		return nil, err
	}
	if err := os.Setenv("SSH_AGENT_PID", sshAgentPid); err != nil {
		return nil, err
	}

	// Connect to the agent - on Windows this might be different
	conn, err := net.Dial("unix", sshAuthSock)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH agent: %w", err)
	}

	// Add the private key to the agent
	if err := addKeyToAgent(conn, privateKeyPath); err != nil {
		conn.Close()
		return nil, err
	}

	return &SSHAgent{
		Conn: conn,
		Pid:  sshAgentPid,
	}, nil
}

func tryGitBashSSHAgent(privateKeyPath string) (*SSHAgent, error) {
	// Try to find Git Bash ssh-agent
	gitBashPaths := []string{
		`C:\Program Files\Git\usr\bin\ssh-agent.exe`,
		`C:\Program Files (x86)\Git\usr\bin\ssh-agent.exe`,
	}

	for _, path := range gitBashPaths {
		if _, err := os.Stat(path); err == nil {
			cmd := exec.Command(path)
			output, err := cmd.Output()
			if err != nil {
				continue
			}

			sshAuthSock, sshAgentPid := parseOutput(string(output))

			if err := os.Setenv("SSH_AUTH_SOCK", sshAuthSock); err != nil {
				continue
			}
			if err := os.Setenv("SSH_AGENT_PID", sshAgentPid); err != nil {
				continue
			}

			// Try to connect
			conn, err := net.Dial("unix", sshAuthSock)
			if err != nil {
				continue
			}

			// Add the private key to the agent
			if err := addKeyToAgent(conn, privateKeyPath); err != nil {
				conn.Close()
				continue
			}

			return &SSHAgent{
				Conn: conn,
				Pid:  sshAgentPid,
			}, nil
		}
	}

	return nil, fmt.Errorf("Git Bash ssh-agent not found")
}

func addKeyToAgent(conn net.Conn, privateKeyPath string) error {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	signer, err := ssh.ParseRawPrivateKey(key)
	if err != nil {
		// If the key is encrypted, prompt for the passphrase
		if _, ok := err.(*ssh.PassphraseMissingError); ok {
			fmt.Print("Enter passphrase for key: ")
			var passphrase string
			fmt.Scanln(&passphrase)

			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
			if err != nil {
				return fmt.Errorf("failed to parse private key with passphrase: %w", err)
			}
		} else {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	agentClient := agent.NewClient(conn)
	if err := agentClient.Add(agent.AddedKey{PrivateKey: signer}); err != nil {
		return fmt.Errorf("failed to add key to agent: %w", err)
	}

	return nil
}

func parseOutput(output string) (string, string) {
	var sshAuthSock, sshAgentPid string

	// Windows SSH agent output might be different
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

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
	if agentPIDStr == "" {
		return nil // Nothing to kill
	}

	if runtime.GOOS == "windows" {
		// Use taskkill on Windows
		cmd := exec.Command("taskkill", "/F", "/PID", agentPIDStr)
		if err := cmd.Run(); err != nil {
			// Try alternative method using Windows API
			return killProcessWindows(agentPIDStr)
		}
		return nil
	}

	// Fallback to Unix-style kill (shouldn't happen in this file)
	cmd := exec.Command("kill", "-9", agentPIDStr)
	return cmd.Run()
}

func killProcessWindows(pidStr string) error {
	// Convert PID to integer for Windows API call
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid PID: %s", err)
	}

	// Open process handle
	handle, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return fmt.Errorf("failed to open process: %w", err)
	}
	defer syscall.CloseHandle(handle)

	// Terminate the process
	err = syscall.TerminateProcess(handle, 0)
	if err != nil {
		return fmt.Errorf("failed to terminate process: %w", err)
	}

	return nil
}
