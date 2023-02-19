package final_socks

import "io"

type AuthHandler interface {
	Authenticate(reader io.Reader, rw ResponseWriter) error
}
