package ssh

import "net"

type SSHAgent struct {
	Conn net.Conn
	Pid  string
}
