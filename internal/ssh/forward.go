package ssh

import (
	"fmt"
	"io"
	"net"

	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/ssh"
)

type Forward struct{}

func StartForwardSession(sshHost string, sshUser string, localHost string, localPort string, remoteHost string, remotePort string, agentConn net.Conn) {
	config, err := getSSHConfig(sshUser, agentConn)
	if err != nil {
		log.Debugf("failed to get SSH config: %s", err)
		return
	}

	conn, err := ssh.Dial("tcp", sshHost, config)
	if err != nil {
		log.Debugf("failed to dial: %s (%s for %s:%s -> %s:%s)", err, sshHost, remoteHost, remotePort, localHost, localPort)
		return
	}
	defer conn.Close()

	localListener, err := net.Listen("tcp", localHost+":"+localPort)
	if err != nil {
		log.Debugf("failed to listen on local port: %s", err)
		fmt.Printf("Failed to listen on %s:%s -> %s\n", localHost, localPort, err)
		return
	}
	defer localListener.Close()

	fmt.Printf("Listening on %s:%s and forwarding to %s:%s\n", localHost, localPort, remoteHost, remotePort)

	for {
		localConn, err := localListener.Accept()
		if err != nil {
			log.Printf("Failed to accept local connection: %s", err)
			continue
		}

		remoteConn, err := conn.Dial("tcp", fmt.Sprintf("%s:%s", remoteHost, remotePort))
		if err != nil {
			log.Debugf("Failed to connect to remote host: %s", err)
			continue
		}

		go handleConnection(localConn, remoteConn)
	}
}

func handleConnection(localConn net.Conn, remoteConn net.Conn) {
	log.Debugf("Handling connection from %s to %s", localConn.RemoteAddr(), remoteConn.RemoteAddr())

	defer localConn.Close()
	defer remoteConn.Close()

	go io.Copy(remoteConn, localConn)
	io.Copy(localConn, remoteConn)
}
