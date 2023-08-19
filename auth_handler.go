package final_socks

import (
	"bufio"
	"net"
)

type AuthHandler interface {
	Authenticate(conn net.Conn, bufConn *bufio.Reader, rw ResponseWriter) (interface{}, error)
}
