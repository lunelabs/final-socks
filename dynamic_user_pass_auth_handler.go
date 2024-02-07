package final_socks

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

type AuthFunction func(username, password string) (interface{}, error)

type DynamicUserPassAuthHandler struct {
	authFunction AuthFunction
}

func NewDynamicUserPassAuthHandler(authFunction AuthFunction) *DynamicUserPassAuthHandler {
	return &DynamicUserPassAuthHandler{
		authFunction: authFunction,
	}
}

func (h *DynamicUserPassAuthHandler) Authenticate(conn net.Conn, bufConn *bufio.Reader, rw ResponseWriter) (interface{}, error) {
	if err := rw.SendUserPassAuth(); err != nil {
		return nil, err
	}

	header := []byte{0, 0}

	if _, err := io.ReadAtLeast(bufConn, header, 2); err != nil {
		return nil, err
	}

	if header[0] != AuthVersion {
		return nil, fmt.Errorf("unsupported auth version: %v", header[0])
	}

	userLen := int(header[1])
	user := make([]byte, userLen)

	if _, err := io.ReadAtLeast(bufConn, user, userLen); err != nil {
		return nil, err
	}

	if _, err := bufConn.Read(header[:1]); err != nil {
		return nil, err
	}

	passLen := int(header[0])
	pass := make([]byte, passLen)

	if _, err := io.ReadAtLeast(bufConn, pass, passLen); err != nil {
		return nil, err
	}

	data, err := h.authFunction(string(user), string(pass))

	if err == nil {
		if err := rw.SendAuthSuccess(); err != nil {
			return nil, err
		}

		return data, nil
	}

	if err := rw.SendAuthFailure(); err != nil {
		return nil, err
	}

	return nil, nil
}
