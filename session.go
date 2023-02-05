package final_socks

import (
	"bitbucket.org/lunelabs/final-socks/pool"
	"context"
	"net"
	"time"
)

type message struct {
	dst net.Addr
	msg []byte
}

// Session is a udp session
type Session struct {
	key   string
	src   net.Addr
	dst   net.Addr
	srcPC *PktConn
	msgCh chan message
	finCh chan struct{}
}

func newSession(key string, src, dst net.Addr, srcPC *PktConn) *Session {
	return &Session{key, src, dst, srcPC, make(chan message, 32), make(chan struct{})}
}

func serveSession(session *Session) {
	dstPC, err := DialUDP("udp", session.srcPC.GetTarget())

	if err != nil {
		//log.F("[socks5u] remote dial error: %v", err)
		//nm.Delete(session.key)
		return
	}
	defer dstPC.Close()

	go func() {
		CopyUDP(session.srcPC, nil, dstPC, 2*time.Minute, 5*time.Second)
		//nm.Delete(session.key)
		close(session.finCh)
	}()

	//log.F("[socks5u] %s <-> %s via %s", session.src, session.srcPC.target, dialer.Addr())

	for {
		select {
		case msg := <-session.msgCh:
			_, err = dstPC.WriteTo(msg.msg, msg.dst)
			if err != nil {
				//log.F("[socks5u] writeTo %s error: %v", msg.dst, err)
			}
			pool.PutBuffer(msg.msg)
			msg.msg = nil
		case <-session.finCh:
			return
		}
	}
}

func DialUDP(network, addr string) (pc net.PacketConn, err error) {
	var la string
	//if d.ip != nil {
	//	la = net.JoinHostPort(d.ip.String(), "0")
	//}

	lc := &net.ListenConfig{}
	//if d.iface != nil {
	//	lc.Control = sockopt.Control(sockopt.Bind(d.iface))
	//}

	return lc.ListenPacket(context.Background(), network, la)
}

// CopyUDP copys from src to dst at target with read timeout.
// if step sets to non-zero value,
// the read timeout will be increased from 0 to timeout by step in every read operation.
func CopyUDP(dst net.PacketConn, writeTo net.Addr, src net.PacketConn, timeout time.Duration, step time.Duration) error {
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
