package final_socks

import "fmt"

type Option func(*Server) error

func NoAuthOption() Option {
	return func(s *Server) error {
		fmt.Println("hi from option")

		return nil
	}
}
