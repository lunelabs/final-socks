package final_socks

import (
	"bufio"
)

type AuthHandler interface {
	Authenticate(bufConn *bufio.Reader, rw ResponseWriter) (interface{}, error)
}
