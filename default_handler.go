package final_socks

import (
	"bitbucket.org/lunelabs/final-socks/pool"
	"fmt"
	"net"
	"sync"
)

var DefaultHandler Handler = func(w ResponseWriter, r *Request) {
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

		var nm sync.Map

		go func() {
			for {
				c := NewPktConn(udpListener, nil, nil, nil)
				buf := pool.GetBuffer(UDPBufSize)

				n, srcAddr, dstAddr, err := c.ReadFrom2(buf)
				if err != nil {
					continue
				}

				var session *Session
				sessionKey := srcAddr.String()

				v, ok := nm.Load(sessionKey)
				if !ok || v == nil {
					session = newSession(sessionKey, srcAddr, dstAddr, c)
					nm.Store(sessionKey, session)
					go serveSession(session)
				} else {
					session = v.(*Session)
				}

				session.msgCh <- message{dstAddr, buf[:n]}
			}
		}()

		buf := pool.GetBuffer(1)
		defer pool.PutBuffer(buf)

		for {
			_, err := r.BufConn.Read(buf)

			if err, ok := err.(net.Error); ok && err.Timeout() {
				continue
			}

			return
		}

		return
	}
}
