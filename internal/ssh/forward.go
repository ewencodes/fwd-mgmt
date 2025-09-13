package ssh

import (
	"fmt"
	"io"
	"net"

	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/ssh"
)

type Forward struct{}

func StartForwardSession(sshHost string, sshUser string, localHost string, localPort string, remoteHost string, remotePort string, privateKeyPath string) error {
	log.Debugf("using %s key", privateKeyPath)
	config, err := getSSHConfig(sshUser, privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to get SSH config: %s", err)
	}

	conn, err := ssh.Dial("tcp", sshHost, config)
	if err != nil {
		return fmt.Errorf("failed to dial: %s (%s for %s:%s)", err, sshHost, remoteHost, remotePort)
	}
	defer func(conn *ssh.Client) {
		err := conn.Close()
		if err != nil {
			log.Warnf("failed to close SSH connection: %s", err)
		}
	}(conn)

	localListener, err := net.Listen("tcp", "127.0.0.1:"+localPort)
	if err != nil {
		log.Debugf("failed to listen on local port: %s", err)
		return fmt.Errorf("failed to listen on %s:%s -> %s", localHost, localPort, err)
	}
	defer func(localListener net.Listener) {
		err := localListener.Close()
		if err != nil {
			log.Warnf("failed to close local listener: %s", err)
		}
	}(localListener)

	fmt.Printf("Listening on %s:%s and forwarding to %s:%s\n", localHost, localPort, remoteHost, remotePort)

	for {
		localConn, err := localListener.Accept()
		if err != nil {
			log.Debugf("failed to accept local connection: %s", err)
			return fmt.Errorf("failed to accept local connection: %s", err)
		}
		log.Debugf("new connection from %s", localConn.RemoteAddr())

		remoteConn, err := conn.Dial("tcp", fmt.Sprintf("%s:%s", remoteHost, remotePort))
		if err != nil {
			log.Debugf("failed to connect to remote host: %s", err)
			return fmt.Errorf("failed to connect to remote host: %s", err)
		}
		log.Debugf("forwarding connection from %s to %s", localConn.RemoteAddr(), remoteConn.RemoteAddr())

		go handleConnection(localConn, remoteConn)
	}
}

func handleConnection(localConn net.Conn, remoteConn net.Conn) {
	log.Debugf("handling connection from %s to %s", localConn.RemoteAddr(), remoteConn.RemoteAddr())

	defer func(localConn net.Conn) {
		err := localConn.Close()
		if err != nil {
			log.Warnf("failed to close local connection: %s", err)
		}
	}(localConn)
	defer func(remoteConn net.Conn) {
		err := remoteConn.Close()
		if err != nil {
			log.Warnf("failed to close remote connection: %s", err)
		}
	}(remoteConn)

	go func() {
		_, err := io.Copy(remoteConn, localConn)
		if err != nil {
			log.Warnf("failed to copy remote connection: %s", err)
		}
	}()
	_, err := io.Copy(localConn, remoteConn)
	if err != nil {
		log.Warnf("failed to copy local connection: %s", err)
	}
}
