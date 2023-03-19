package final_socks

import (
	"context"
	"fmt"
	"github.com/lunelabs/final-socks/pool"
	"net"
)

var DefaultHandler Handler = func(w ResponseWriter, r *Request) {
	defer func() {
		fmt.Println("end", r.Command)
	}()

	if r.Command != CommandConnect && r.Command != CommandAssociate {
		w.SendNotSupportedCommand()

		return
	}

	if r.Command == CommandConnect {
		target, err := net.Dial("tcp", r.DestAddr.Address())

		if err != nil {
			w.SendNetworkError(err.Error())

			return
		}

		defer target.Close()

		addr := target.LocalAddr().(*net.TCPAddr)

		if err := w.SendSucceeded(&AddrSpec{IP: addr.IP, Port: addr.Port}); err != nil {
			return
		}

		if err := w.Proxy(target, r.BufConn); err != nil {
			fmt.Println(err)
		}

		return
	}

	if r.Command == CommandAssociate {
		udpListener, err := net.ListenPacket("udp", net.JoinHostPort(r.LocalAddr.IP.String(), "0"))

		if err != nil {
			w.SendGeneralServerFailure()

			return
		}

		defer udpListener.Close()

		udpAddr, ok := udpListener.LocalAddr().(*net.UDPAddr)

		if !ok {
			w.SendGeneralServerFailure()

			return
		}

		if err := w.SendSucceeded(&AddrSpec{IP: udpAddr.IP, Port: udpAddr.Port}); err != nil {
			return
		}

		errChan := make(chan error)
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			buf := pool.GetBuffer(1)
			defer pool.PutBuffer(buf)

			for {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := r.BufConn.Read(buf)

					if err, ok := err.(net.Error); ok && err.Timeout() {
						continue
					}

					errChan <- nil
				}
			}
		}()

		go func() {
			c := NewPktConn(udpListener, nil, nil, nil)
			buf := pool.GetBuffer(UDPBufSize)

			n, srcAddr, dstAddr, err := c.ReadFrom2(buf)

			if err != nil {
				errChan <- err

				return
			}

			sessionKey := srcAddr.String()
			session := NewSession(sessionKey, srcAddr, dstAddr, c)

			go session.Serve(ctx, errChan)

			session.ProcessMessage(Message{dstAddr, buf[:n]})
		}()

		err = <-errChan
		cancel()

		return
	}
}
