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
		log.Fatalf("Failed to get SSH config: %s", err)
	}

	conn, err := ssh.Dial("tcp", sshHost, config)
	if err != nil {
		log.Fatalf("Failed to dial: %s", err)
	}
	defer conn.Close()

	localListener, err := net.Listen("tcp", localHost+":"+localPort)
	if err != nil {
		log.Fatalf("Failed to listen on local port: %s", err)
	}
	defer localListener.Close()

	fmt.Printf("Listening on %s:%s and forwarding to %s:%s\n", localHost, localPort, remoteHost, remotePort)

	for {
		localConn, err := localListener.Accept()
		if err != nil {
			log.Printf("Failed to accept local connection: %s", err)
			continue
		}

		go handleConnection(conn, localConn, remoteHost, remotePort)
	}
}

func handleConnection(sshConn *ssh.Client, localConn net.Conn, remoteHost, remotePort string) {
	defer localConn.Close()

	remoteConn, err := sshConn.Dial("tcp", fmt.Sprintf("%s:%s", remoteHost, remotePort))
	if err != nil {
		log.Debugf("Failed to connect to remote host: %s", err)
		return
	}
	defer remoteConn.Close()

	go io.Copy(remoteConn, localConn)
	io.Copy(localConn, remoteConn)
}
