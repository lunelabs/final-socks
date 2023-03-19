package final_socks

import (
	"context"
	"fmt"
	"github.com/lunelabs/final-socks/pool"
	"net"
	"time"
)

type Message struct {
	Dst net.Addr
	Msg []byte
}

// Session is udp session
type Session struct {
	key    string
	src    net.Addr
	dst    net.Addr
	srcPC  *PktConn
	exitIP net.IP
	msgCh  chan Message
	finCh  chan struct{}
}

func NewSession(
	key string,
	src,
	dst net.Addr,
	srcPC *PktConn,
	exitIP net.IP,
) *Session {
	return &Session{
		key:    key,
		src:    src,
		dst:    dst,
		srcPC:  srcPC,
		exitIP: exitIP,
		msgCh:  make(chan Message, 32),
		finCh:  make(chan struct{}),
	}
}

func (s *Session) ProcessMessage(message Message) {
	s.msgCh <- message
}

func (s *Session) Serve(ctx context.Context, errChan chan error) {
	dstPC, err := s.dialUDP("udp", s.srcPC.GetTarget())

	if err != nil {
		errChan <- err

		return
	}

	defer dstPC.Close()

	go func() {
		s.copyUDP(s.srcPC, nil, dstPC, 2*time.Minute, 5*time.Second)

		close(s.finCh)
	}()

	for {
		select {
		case msg := <-s.msgCh:
			_, err = dstPC.WriteTo(msg.Msg, msg.Dst)

			if err != nil {
				fmt.Println(err)
			}

			pool.PutBuffer(msg.Msg)
			msg.Msg = nil
		case <-s.finCh:
			errChan <- nil

			return
		case <-ctx.Done():

			return
		}
	}
}

func (s *Session) dialUDP(network, addr string) (pc net.PacketConn, err error) {
	var la string

	if s.exitIP != nil {
		la = net.JoinHostPort(s.exitIP.String(), "0")
	}

	lc := &net.ListenConfig{}

	return lc.ListenPacket(context.Background(), network, la)
}

func (s *Session) copyUDP(
	dst net.PacketConn,
	writeTo net.Addr,
	src net.PacketConn,
	timeout time.Duration,
	step time.Duration,
) error {
	buf := pool.GetBuffer(UDPBufSize)
	defer pool.PutBuffer(buf)

	var t time.Duration

	for {
		if t += step; t == 0 || t > timeout {
			t = timeout
		}

		src.SetReadDeadline(time.Now().Add(t))
		n, addr, err := src.ReadFrom(buf)

		if err != nil {
			return err
		}

		if writeTo != nil {
			addr = writeTo
		}

		_, err = dst.WriteTo(buf[:n], addr)

		if err != nil {
			return err
		}
	}
}
