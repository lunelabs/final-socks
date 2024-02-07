package final_socks

type Option func(*Server) error

func NoAuthOption() Option {
	return func(s *Server) error {
		s.AuthHandlers[AuthNoAuth] = NewNoAuthHandler()

		return nil
	}
}

func UserPassAuth(user, pass string) Option {
	return func(s *Server) error {
		s.AuthHandlers[AuthUserPass] = NewUserPassAuthHandler(user, pass)

		return nil
	}
}

func DynamicUserPassAuth(authFunction AuthFunction) Option {
	return func(s *Server) error {
		s.AuthHandlers[AuthUserPass] = NewDynamicUserPassAuthHandler(authFunction)

		return nil
	}
}
