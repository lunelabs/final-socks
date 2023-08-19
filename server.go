package final_socks

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"net"
)

type Handler func(ResponseWriter, *Request)

type Server struct {
	addr         string
	handler      Handler
	AuthHandlers map[uint8]AuthHandler
}

func NewServer(addr string, handler Handler) *Server {
	if addr == "" {
		addr = ":1080"
	}

	return &Server{
		addr:         addr,
		handler:      handler,
		AuthHandlers: map[uint8]AuthHandler{},
	}
}

func (s *Server) SetOption(option Option) error {
	return option(s)
}

func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.addr)

	if err != nil {
		return err
	}

	return s.Serve(listener)
}

func (s *Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()

		if err != nil {
			return err
		}

		go s.ServeConn(conn)
	}
}

func (s *Server) ServeConn(conn net.Conn) error {
	defer conn.Close()

	bufConn := bufio.NewReader(conn)
	socksVersion, err := ReadSocksVersion(bufConn)

	if err != nil {
		return err
	}

	if socksVersion != VersionSocks5 {
		return fmt.Errorf("unsupported socks version: %v", socksVersion)
	}

	rw := NewResponseWriter(conn)
	user, err := s.authenticate(bufConn, rw)

	if err != nil {
		return errors.Wrap(err, "failed to authenticate")
	}

	req, err := ReadRequest(bufConn)

	if err != nil {
		return errors.Wrap(err, "failed to read request")
	}

	if req.Version != VersionSocks5 {
		return fmt.Errorf("unsupported socks version: %v", socksVersion)
	}

	req.User = user
	req = s.decorateRequestWithConnectionInfo(req, conn)

	s.handler(rw, req)

	return nil
}

func (s *Server) decorateRequestWithConnectionInfo(req *Request, conn net.Conn) *Request {
	if val, ok := conn.LocalAddr().(*net.TCPAddr); ok {
		req.LocalAddr = *val
	}

	if val, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		req.RemoteAddr = *val
	}

	return req
}

func (s *Server) authenticate(bufConn *bufio.Reader, rw ResponseWriter) (interface{}, error) {
	authMethods, err := ReadAuthenticateMethods(bufConn)

	if err != nil {
		return nil, err
	}

	for _, authMethod := range authMethods {
		if handler, ok := s.AuthHandlers[authMethod]; ok {
			return handler.Authenticate(bufConn, rw)
		}
	}

	if err = rw.SendNoAcceptableAuth(); err != nil {
		return nil, err
	}

	return nil, errors.New("authentication failed")
}

func Handle(handler Handler) {
	DefaultHandler = handler
}

func ListenAndServe(addr string, handler Handler, options ...Option) error {
	serverHandler := DefaultHandler

	if handler != nil {
		serverHandler = handler
	}

	s := NewServer(addr, serverHandler)

	for _, option := range options {
		if err := s.SetOption(option); err != nil {
			return err
		}
	}

	return s.ListenAndServe()
}
