package final_socks

import "io"

type NoAuthHandler struct {
}

func NewNoAuthHandler() *NoAuthHandler {
	return &NoAuthHandler{}
}

func (h *NoAuthHandler) Authenticate(reader io.Reader, rw ResponseWriter) error {
	if err := rw.SendNoAuth(); err != nil {
		return err
	}

	return nil
}
