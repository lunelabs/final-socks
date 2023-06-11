package final_socks

import (
	"net"
)

type AuthHandler interface {
	Authenticate(conn net.Conn, rw ResponseWriter) (interface{}, error)
}
