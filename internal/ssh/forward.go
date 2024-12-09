package ssh

import (
	"fmt"
	"io"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/ssh"
)

type Forward struct{}

func StartForwardSession(sshHost string, sshUser string, localHost string, localPort string, remoteHost string, remotePort string, agentConn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
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
