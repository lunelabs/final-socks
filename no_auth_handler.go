package final_socks

import (
	"net"
)

type NoAuthHandler struct {
}

func NewNoAuthHandler() *NoAuthHandler {
	return &NoAuthHandler{}
}

func (h *NoAuthHandler) Authenticate(conn net.Conn, rw ResponseWriter) error {
	if err := rw.SendNoAuth(); err != nil {
		return err
	}

	return nil
}
