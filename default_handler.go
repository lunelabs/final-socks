package final_socks

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/lunelabs/final-socks/pool"
)

var DefaultHandler Handler = func(w ResponseWriter, r *Request) {
	fmt.Println(time.Now().Unix(), "start", "command", r.Command)

	defer func() {
		fmt.Println(time.Now().Unix(), "end", "command", r.Command)
	}()

	if r.Command != CommandConnect && r.Command != CommandAssociate {
		_ = w.SendNotSupportedCommand()

		return
	}

	if r.Command == CommandConnect {
		target, err := net.Dial("tcp", r.DestAddr.Address())

		if err != nil {
			_ = w.SendNetworkError(err.Error())

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

	udpListener, err := net.ListenPacket("udp", net.JoinHostPort(r.LocalAddr.IP.String(), "0"))

	if err != nil {
		_ = w.SendGeneralServerFailure()

		return
	}

	defer udpListener.Close()

	udpAddr, ok := udpListener.LocalAddr().(*net.UDPAddr)

	if !ok {
		_ = w.SendGeneralServerFailure()

		return
	}

	if err := w.SendSucceeded(&AddrSpec{IP: udpAddr.IP, Port: udpAddr.Port}); err != nil {
		return
	}

	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := NewPktConn(udpListener, nil, nil, nil)
		buf := pool.GetBuffer(UDPBufSize)

		n, srcAddr, dstAddr, err := c.ReadFrom2(buf)

		if err != nil {
			errChan <- err

			return
		}

		sessionKey := srcAddr.String()
		session := NewSession(sessionKey, srcAddr, dstAddr, c, nil)

		go session.Serve(ctx, errChan)

		session.ProcessMessage(Message{dstAddr, buf[:n]})
	}()

	go func() {
		_, _ = io.Copy(io.Discard, r.BufConn)

		errChan <- nil
	}()

	<-errChan

	cancel()
}
