package final_socks

import (
	"net"
)

type NoAuthHandler struct {
}

func NewNoAuthHandler() *NoAuthHandler {
	return &NoAuthHandler{}
}

func (h *NoAuthHandler) Authenticate(conn net.Conn, rw ResponseWriter) (interface{}, error) {
	if err := rw.SendNoAuth(); err != nil {
		return nil, err
	}

	return nil, nil
}
