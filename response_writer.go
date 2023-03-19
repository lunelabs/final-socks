package final_socks

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net"
	"strings"
)

type ResponseWriter struct {
	conn io.Writer
}

func NewResponseWriter(conn io.Writer) ResponseWriter {
	return ResponseWriter{
		conn: conn,
	}
}

func (rw ResponseWriter) SendNoAuth() error {
	_, err := rw.conn.Write([]byte{VersionSocks5, AuthNoAuth})

	if err != nil {
		return errors.Wrap(err, "sending no auth failed")
	}

	return nil
}

func (rw ResponseWriter) SendUserPassAuth() error {
	_, err := rw.conn.Write([]byte{VersionSocks5, AuthUserPass})

	if err != nil {
		return errors.Wrap(err, "sending no auth failed")
	}

	return nil
}

func (rw ResponseWriter) SendNoAcceptableAuth() error {
	_, err := rw.conn.Write([]byte{VersionSocks5, AuthNoAcceptable})

	if err != nil {
		return errors.Wrap(err, "sending no acceptable auth failed")
	}

	return nil
}

func (rw ResponseWriter) SendAuthSuccess() error {
	_, err := rw.conn.Write([]byte{UserAuthVersion, AuthSuccess})

	if err != nil {
		return errors.Wrap(err, "sending no acceptable auth failed")
	}

	return nil
}

func (rw ResponseWriter) SendAuthFailure() error {
	_, err := rw.conn.Write([]byte{UserAuthVersion, AuthFailure})

	if err != nil {
		return errors.Wrap(err, "sending no acceptable auth failed")
	}

	return nil
}

func (rw ResponseWriter) SendNotSupportedCommand() error {
	if err := rw.SendReply(ReplyCommandNotSupported, nil); err != nil {
		return errors.Wrap(err, "sending not supported command failed")
	}

	return nil
}

func (rw ResponseWriter) SendSucceeded(addr *AddrSpec) error {
	if err := rw.SendReply(ReplySucceeded, addr); err != nil {
		return errors.Wrap(err, "sending succeeded failed")
	}

	return nil
}

func (rw ResponseWriter) SendGeneralServerFailure() error {
	if err := rw.SendReply(ReplyGeneralServerFailure, nil); err != nil {
		return errors.Wrap(err, "sending general server failure failed")
	}

	return nil
}

func (rw ResponseWriter) SendNetworkError(msg string) error {
	resp := ReplyHostUnreachable

	if strings.Contains(msg, "refused") {
		resp = ReplyConnectionRefused
	} else if strings.Contains(msg, "network is unreachable") {
		resp = ReplyNetworkUnreachable
	}

	if err := rw.SendReply(resp, nil); err != nil {
		return errors.Wrap(err, "sending network error failed")
	}

	return nil
}

func (rw ResponseWriter) SendReply(resp uint8, addr *AddrSpec) error {
	var addrType uint8
	var addrBody []byte
	var addrPort uint16

	switch {
	case addr == nil:
		addrType = AddressIpv4
		addrBody = []byte{0, 0, 0, 0}
		addrPort = 0

	case addr.FQDN != "":
		addrType = AddressFqdn
		addrBody = append([]byte{byte(len(addr.FQDN))}, addr.FQDN...)
		addrPort = uint16(addr.Port)

	case addr.IP.To4() != nil:
		addrType = AddressIpv4
		addrBody = addr.IP.To4()
		addrPort = uint16(addr.Port)

	case addr.IP.To16() != nil:
		addrType = AddressIpv6
		addrBody = addr.IP.To16()
		addrPort = uint16(addr.Port)

	default:
		return fmt.Errorf("failed to format address: %v", addr)
	}

	// Format the Message
	msg := make([]byte, 6+len(addrBody))
	msg[0] = VersionSocks5
	msg[1] = resp
	msg[2] = 0 // Reserved
	msg[3] = addrType
	copy(msg[4:], addrBody)
	msg[4+len(addrBody)] = byte(addrPort >> 8)
	msg[4+len(addrBody)+1] = byte(addrPort & 0xff)

	// Send the Message
	_, err := rw.conn.Write(msg)

	return err
}

func (rw ResponseWriter) Proxy(target net.Conn, bufConn io.Reader) error {
	errCh := make(chan error, 2)

	go rw.proxy(target, bufConn, errCh)
	go rw.proxy(rw.conn, target, errCh)

	for i := 0; i < 2; i++ {
		err := <-errCh

		if err != nil {
			return err
		}
	}

	return nil
}

func (rw ResponseWriter) proxy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)

	if tcpConn, ok := dst.(closeWriter); ok {
		tcpConn.CloseWrite()
	}

	errCh <- err
}
