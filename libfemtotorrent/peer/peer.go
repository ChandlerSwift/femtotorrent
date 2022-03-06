package peer

import (
	"net"
)

type Peer struct {
	IPAddress  net.IP
	Port       uint16
	Choked     bool
	Interested bool
	conn       net.Conn
}
