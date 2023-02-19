package final_socks

import (
	"fmt"
	"io"
)

type UserPassAuthHandler struct {
	username string
	password string
}

func NewUserPassAuthHandler(username, password string) *UserPassAuthHandler {
	return &UserPassAuthHandler{
		username: username,
		password: password,
	}
}

func (h *UserPassAuthHandler) Authenticate(reader io.Reader, rw ResponseWriter) error {
	if err := rw.SendUserPassAuth(); err != nil {
		return err
	}

	header := []byte{0, 0}

	if _, err := io.ReadAtLeast(reader, header, 2); err != nil {
		return err
	}

	if header[0] != AuthVersion {
		return fmt.Errorf("unsupported auth version: %v", header[0])
	}

	userLen := int(header[1])
	user := make([]byte, userLen)

	if _, err := io.ReadAtLeast(reader, user, userLen); err != nil {
		return err
	}

	if _, err := reader.Read(header[:1]); err != nil {
		return err
	}

	passLen := int(header[0])
	pass := make([]byte, passLen)

	if _, err := io.ReadAtLeast(reader, pass, passLen); err != nil {
		return err
	}

	if h.username == string(user) && h.password == string(pass) {
		if err := rw.SendAuthSuccess(); err != nil {
			return err
		}

		return nil
	}

	if err := rw.SendAuthFailure(); err != nil {
		return err
	}

	return nil
}
