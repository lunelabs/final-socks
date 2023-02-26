package final_socks

import (
	"bufio"
	"errors"
	"github.com/lunelabs/final-socks/pool"
	"net"
	"net/netip"
	"strconv"
)

var UDPBufSize = 2 << 10

// PktConn .
type PktConn struct {
	net.PacketConn
	ctrlConn *bufio.Reader // tcp control conn
	writeTo  net.Addr      // write to and read from addr
	target   Addr
}

// NewPktConn returns a PktConn, the writeAddr must be *net.UDPAddr or *net.UnixAddr.
func NewPktConn(c net.PacketConn, writeAddr net.Addr, targetAddr Addr, ctrlConn *bufio.Reader) *PktConn {
	pc := &PktConn{
		PacketConn: c,
		writeTo:    writeAddr,
		target:     targetAddr,
		ctrlConn:   ctrlConn,
	}

	if ctrlConn != nil {
		go func() {
			buf := pool.GetBuffer(1)
			defer pool.PutBuffer(buf)
			for {
				_, err := ctrlConn.Read(buf)
				if err, ok := err.(net.Error); ok && err.Timeout() {
					continue
				}
				// log.F("[socks5] dialudp udp associate end")
				return
			}
		}()
	}

	return pc
}

func (pc *PktConn) GetTarget() string {
	return pc.target.String()
}

// ReadFrom overrides the original function from net.PacketConn.
func (pc *PktConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, _, target, err := pc.readFrom(b)
	return n, target, err
}

// ReadFrom overrides the original function from net.PacketConn.
func (pc *PktConn) ReadFrom2(b []byte) (int, net.Addr, net.Addr, error) {
	return pc.readFrom(b)
}

func (pc *PktConn) readFrom(b []byte) (int, net.Addr, net.Addr, error) {
	buf := pool.GetBuffer(len(b))
	defer pool.PutBuffer(buf)

	n, raddr, err := pc.PacketConn.ReadFrom(buf)
	if err != nil {
		return n, raddr, nil, err
	}

	if n < 3 {
		return n, raddr, nil, errors.New("not enough size to get addr")
	}

	// https://www.rfc-editor.org/rfc/rfc1928#section-7
	// +----+------+------+----------+----------+----------+
	// |RSV | FRAG | ATYP | DST.ADDR | DST.PORT |   DATA   |
	// +----+------+------+----------+----------+----------+
	// | 2  |  1   |  1   | Variable |    2     | Variable |
	// +----+------+------+----------+----------+----------+
	tgtAddr := SplitAddr(buf[3:n])
	if tgtAddr == nil {
		return n, raddr, nil, errors.New("can not get target addr")
	}

	target, err := net.ResolveUDPAddr("udp", tgtAddr.String())
	if err != nil {
		return n, raddr, nil, errors.New("wrong target addr")
	}

	if pc.writeTo == nil {
		pc.writeTo = raddr
	}

	if pc.target == nil {
		pc.target = make([]byte, len(tgtAddr))
		copy(pc.target, tgtAddr)
	}

	n = copy(b, buf[3+len(tgtAddr):n])
	return n, raddr, target, err
}

// WriteTo overrides the original function from net.PacketConn.
func (pc *PktConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	target := pc.target
	if addr != nil {
		target = ParseAddr(addr.String())
	}

	if target == nil {
		return 0, errors.New("invalid addr")
	}

	buf := pool.GetBytesBuffer()
	defer pool.PutBytesBuffer(buf)

	buf.Write([]byte{0, 0, 0})
	tgtLen, _ := buf.Write(target)
	buf.Write(b)

	n, err := pc.PacketConn.WriteTo(buf.Bytes(), pc.writeTo)
	if n > tgtLen+3 {
		return n - tgtLen - 3, err
	}

	return 0, err
}

// Close .
func (pc *PktConn) Close() error {
	//if pc.ctrlConn != nil {
	//	pc.ctrlConn.Close()
	//}

	return pc.PacketConn.Close()
}

// Addr represents a SOCKS address as defined in RFC 1928 section 5.
type Addr []byte

// String serializes SOCKS address a to string form.
func (a Addr) String() string {
	var host, port string

	switch a[0] { // address type
	case AddressFqdn:
		host = string(a[2 : 2+int(a[1])])
		port = strconv.Itoa((int(a[2+int(a[1])]) << 8) | int(a[2+int(a[1])+1]))
	case AddressIpv4:
		host = net.IP(a[1 : 1+net.IPv4len]).String()
		port = strconv.Itoa((int(a[1+net.IPv4len]) << 8) | int(a[1+net.IPv4len+1]))
	case AddressIpv6:
		host = net.IP(a[1 : 1+net.IPv6len]).String()
		port = strconv.Itoa((int(a[1+net.IPv6len]) << 8) | int(a[1+net.IPv6len+1]))
	}

	return net.JoinHostPort(host, port)
}

// SplitAddr slices a SOCKS address from beginning of b. Returns nil if failed.
func SplitAddr(b []byte) Addr {
	addrLen := 1
	if len(b) < addrLen {
		return nil
	}

	switch b[0] {
	case AddressFqdn:
		if len(b) < 2 {
			return nil
		}
		addrLen = 1 + 1 + int(b[1]) + 2
	case AddressIpv4:
		addrLen = 1 + net.IPv4len + 2
	case AddressIpv6:
		addrLen = 1 + net.IPv6len + 2
	default:
		return nil
	}

	if len(b) < addrLen {
		return nil
	}

	return b[:addrLen]
}

// ParseAddr parses the address in string s. Returns nil if failed.
func ParseAddr(s string) Addr {
	var addr Addr
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return nil
	}

	if ip, err := netip.ParseAddr(host); err == nil {
		if ip.Is4() {
			addr = make([]byte, 1+net.IPv4len+2)
			addr[0] = AddressIpv4
		} else {
			addr = make([]byte, 1+net.IPv6len+2)
			addr[0] = AddressIpv6
		}
		copy(addr[1:], ip.AsSlice())
	} else {
		if len(host) > 255 {
			return nil
		}
		addr = make([]byte, 1+1+len(host)+2)
		addr[0] = AddressFqdn
		addr[1] = byte(len(host))
		copy(addr[2:], host)
	}

	portnum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil
	}

	addr[len(addr)-2], addr[len(addr)-1] = byte(portnum>>8), byte(portnum)

	return addr
}
